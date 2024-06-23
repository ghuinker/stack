package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
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
	"time"

	"github.com/caddyserver/certmagic"
)

func Runserver() {
	var devMode bool
	addrport := os.Getenv("ADDR")

	flag.StringVar(&addrport, "addr", addrport, "Optional address and port number, or ipaddr:port")
	flag.BoolVar(&devMode, "dev", false, "Run server in dev mode")

	flag.CommandLine.Parse(os.Args[2:])

	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
		return
	}

	checkFileSize("logs", 500*1024*1024) // 500MB

	logFilePath := filepath.Join("logs", "requests.log")

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
		return
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	err = runManageCommand([]string{"migrate", "--check"})
	if err != nil {
		println("Migrations not applied, run: stack manage migrate")
	}

	gunicornURL, gunicornCmd, err := startGunicorn(devMode)
	if err != nil {
		return
	}

	router := http.NewServeMux()

	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/icons/favicon.ico", http.StatusMovedPermanently)
	})
	router.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.FileServer(http.FS(GlobalContext.Dist)).ServeHTTP(w, r)
	})

	// Django
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reverseProxy(w, r, gunicornURL)
	})

	configuredRouter := loggingMiddleware(router, !devMode)

	server := &http.Server{Addr: addrport, Handler: configuredRouter}
	go func() {
		if os.Getenv("AUTO_TLS") == "true" {
			println("Starting server at: " + os.Getenv("HOST_NAME"))
			if err := certmagic.HTTPS([]string{os.Getenv("HOST_NAME")}, configuredRouter); err != nil {
				println("Error starting server: ", err)
			}
		} else {
			println("Starting server at: " + addrport)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTP server error: %v\n", err)
			}
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

func loggingMiddleware(next http.Handler, shouldLog bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		if shouldLog {
			log.Printf("%s %s [%s] - %v", r.Method, r.RequestURI, r.RemoteAddr, duration)
		}
	})
}

func checkFileSize(directory string, maxSize int64) error {
	maxSizeMB := maxSize / (1024 * 1024) // Convert maxSize to megabytes

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Size() > maxSize {
			log.Printf("Log %s is over %d MB. Run ./stack compresslogs to zip.", path, maxSizeMB)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
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
	gunicornPort, err := findAvailablePort(8100)
	if err != nil {
		fmt.Println("Error finding gunicorn port:", err)
		return "", nil, err
	}
	gunicornURL := fmt.Sprintf("127.0.0.1:%d", gunicornPort)

	cmdArgs := []string{filepath.Join(GlobalContext.OutDir, "gunicorn"), "app.config.wsgi", "-b " + gunicornURL}
	if devMode {
		// Using runserver for dev does better job at reloading
		cmdArgs = []string{"manage.py", "runserver", gunicornURL}
	}
	cmd := exec.Command("python3", cmdArgs...)
	if devMode {
		setDevPythonEnv(cmd)
		cmd.Env = append(cmd.Env, "DEBUG=true")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		setPythonEnv(cmd)
		logFile, err := os.OpenFile("logs/gunicorn.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Failed to open log", err)
		}
		defer logFile.Close()
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	err = cmd.Start()

	if err != nil {
		fmt.Println("Error starting gunicorn process:", err)
		return "", nil, nil
	}
	return gunicornURL, cmd, nil
}
