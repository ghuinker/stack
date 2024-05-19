package cmd

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

func Runserver() {
	var devMode bool
	addrport := os.Getenv("ADDR")

	flag.StringVar(&addrport, "addr", addrport, "Optional address and port number, or ipaddr:port")
	flag.BoolVar(&devMode, "dev", false, "Run server in dev mode")

	flag.CommandLine.Parse(os.Args[2:])

	err := runManageCommand([]string{"migrate", "--check"})
	if err != nil {
		println("Migrations not applied, run: stack manage migrate")
	}

	gunicornURL, gunicornCmd, err := startGunicorn(devMode)
	if err != nil {
		return
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/icons/favicon.ico", http.StatusMovedPermanently)
	})
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.FileServer(http.FS(GlobalContext.StaticFiles)).ServeHTTP(w, r)
	})

	// Django
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reverseProxy(w, r, gunicornURL)
	})

	server := &http.Server{Addr: addrport}
	go func() {
		println("Starting server at: " + addrport)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Capture interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Handle shutdown
	fmt.Println("Shutting down...")

	// Terminate Python subprocess
	if err := gunicornCmd.Process.Signal(os.Interrupt); err != nil {
		fmt.Printf("Error terminating Python subprocess: %v\n", err)
	}
	gunicornCmd.Process.Wait()

	// Shutdown HTTP server gracefully
	if err := server.Shutdown(context.Background()); err != nil {
		fmt.Printf("Error shutting down HTTP server: %v\n", err)
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

func startGunicorn(devMode bool) (string, *exec.Cmd, error) {
	tempDir := GlobalContext.TempDir
	gunicornPort, err := findAvailablePort(8100)
	if err != nil {
		fmt.Println("Error finding gunicorn port:", err)
		return "", nil, err
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
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err = cmd.Start()

	if err != nil {
		fmt.Println("Error starting gunicorn process:", err)
		return "", nil, nil
	}
	return gunicornURL, cmd, nil
}
