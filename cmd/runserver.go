package cmd

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func Runserver() {
	var devMode bool
	addrport := os.Getenv("ADDR")

	flag.StringVar(&addrport, "addr", addrport, "Optional address and port number, or ipaddr:port")
	flag.BoolVar(&devMode, "dev", false, "Run server in dev mode")

	flag.CommandLine.Parse(os.Args[2:])

	if !strings.Contains(addrport, ":") {
		addrport = "127.0.0.1:" + addrport
	}

	err := runManageCommand([]string{"migrate", "--check"})
	if err != nil {
		println("Migrations not applied, run: stack manage migrate")
	}

	gunicornURL, err := startGunicorn(devMode)
	if err != nil {
		return
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/icons/favicon.ico", http.StatusMovedPermanently)
	})
	http.Handle("/static/", http.FileServer(http.FS(GlobalContext.StaticFiles)))

	// Django
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reverseProxy(w, r, gunicornURL)
	})

	println("Starting server at: " + addrport)
	err = http.ListenAndServe(addrport, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func findAvailablePort(startPort int) (int, error) {
	for port := startPort; port <= 65535; port++ {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in the range")
}

func reverseProxy(w http.ResponseWriter, r *http.Request, gunicornURL string) {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   gunicornURL,
	})

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
}

func startGunicorn(devMode bool) (string, error) {
	tempDir := GlobalContext.TempDir
	gunicornPort, err := findAvailablePort(8100)
	if err != nil {
		fmt.Println("Error finding gunicorn port:", err)
		return "", err
	}
	gunicornURL := fmt.Sprintf("127.0.0.1:%d", gunicornPort)

	// 3. Start a Python process to run the Python files in the temporary directory
	cmdArgs := []string{filepath.Join(tempDir, "venv/bin/gunicorn"), "app.config.wsgi", "-b " + gunicornURL}
	if devMode {
		cmdArgs = append(cmdArgs, "--reload")
	}
	cmd := exec.Command("python3", cmdArgs...)
	cmd.Env = append(cmd.Env, "DIST_DIR="+tempDir)
	if devMode {
		setDevPythonEnv(cmd)
		cmd.Env = append(cmd.Env, "DEBUG=true")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		setPythonEnv(cmd)
	}

	err = cmd.Start()

	if err != nil {
		fmt.Println("Error starting gunicorn process:", err)
		return "", nil
	}
	return gunicornURL, nil
}
