package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/caddyserver/certmagic"
)

//go:embed all:static
var staticFiles embed.FS

var gunicornURL = "127.0.0.1:8000"

func main() {
	// You can handle more graceful failures here
	err := runMigrations()
	if err != nil {
		fmt.Println("Error running migrations: ", err)
		return
	}

	gunicornCmd, err := startGunicorn()
	if err != nil {
		fmt.Println("Error starting gunicorn: ", err)
		return
	}

	router := http.NewServeMux()

	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/icons/favicon.ico", http.StatusMovedPermanently)
	})
	router.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.FileServer(http.FS(staticFiles)).ServeHTTP(w, r)
	})

	// Django
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reverseProxy(w, r, gunicornURL)
	})

	server := &http.Server{Handler: router}
	go func() {
		if os.Getenv("CERT_AUTO_TLS") == "true" {
			certmagic.DefaultACME.Agreed = true
			certmagic.DefaultACME.Email = os.Getenv("CERT_EMAIL")
			if strings.EqualFold(os.Getenv("CERT_STAGING"), "true") {
				certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA
			}
			println("Starting server at: " + os.Getenv("HOST_NAME"))
			if err := certmagic.HTTPS([]string{os.Getenv("HOST_NAME")}, router); err != nil {
				println("Error starting server: ", err)
			}
		} else {
			println("Starting server")
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

func reverseProxy(w http.ResponseWriter, r *http.Request, gunicornURL string) {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   gunicornURL,
	})

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
}

func startGunicorn() (*exec.Cmd, error) {
	cmdArgs := []string{"app.config.wsgi", "-b " + gunicornURL, "--max-requests", "1200", "--max-requests-jitter", "50"}
	cmd := exec.Command("gunicorn", cmdArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()

	if err != nil {
		fmt.Println("Error starting gunicorn process:", err)
		return nil, nil
	}
	return cmd, nil
}

func runMigrations() error {
	cmdArgs := []string{"manage.py", "migrate"}
	cmd := exec.Command("python3", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
