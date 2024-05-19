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
	var addrport string
	var devMode bool
	flag.StringVar(&addrport, "addr", "127.0.0.1:8000", "Optional address and port number, or ipaddr:port")
	flag.BoolVar(&devMode, "dev", false, "Run server in dev mode")

	flag.CommandLine.Parse(os.Args[2:])

	if !strings.Contains(addrport, ":") {
		// If it doesn't contain a colon, assume it's just a port number and prepend 127.0.0.1
		addrport = "127.0.0.1:" + addrport
	}

	fmt.Println("Starting server at:", addrport)
	// TODO: we may want to check for migrations
	gunicornURL, err := startGunicorn(devMode)
	if err != nil {
		return
	}

	// TODO: if devmode then figure out asset thing for vite
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
