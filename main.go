package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"stack/cmd"
	"strings"
)

//go:embed all:app
//go:embed venv/bin/gunicorn
//go:embed manage.py
//go:embed static/assets/manifest.json
//go:embed .env.example
//go:embed all:venv/lib/python3.12/site-packages/asgiref all:venv/lib/python3.12/site-packages/certifi all:venv/lib/python3.12/site-packages/charset_normalizer all:venv/lib/python3.12/site-packages/django all:venv/lib/python3.12/site-packages/django_filters all:venv/lib/python3.12/site-packages/rest_framework all:venv/lib/python3.12/site-packages/environs all:venv/lib/python3.12/site-packages/gunicorn all:venv/lib/python3.12/site-packages/idna all:venv/lib/python3.12/site-packages/marshmallow all:venv/lib/python3.12/site-packages/packaging all:venv/lib/python3.12/site-packages/dotenv all:venv/lib/python3.12/site-packages/requests all:venv/lib/python3.12/site-packages/sqlparse all:venv/lib/python3.12/site-packages/urllib3
var embeddedFiles embed.FS

//go:embed static/*
var staticFiles embed.FS

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run script.go <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	loadEnvFile(".env")

	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Println("Error creating temporary directory:", err)
		return
	}

	cmd.GlobalContext = &cmd.ManageContext{TempDir: tempDir, StaticFiles: staticFiles}

	err = extractFiles(embeddedFiles, tempDir)
	defer os.RemoveAll(tempDir)

	if err != nil {
		fmt.Println("Error extracting embedded files:", err)
		return
	}

	switch command {
	case "manage":
		cmd.Manage()
	case "runserver":
		cmd.Runserver()
	case "shell":
		cmd.Shell()
	case "createsuperuser":
		cmd.CreateSuperUser()
	case "updateembed":
		cmd.UpdateEmbed()
	case "setup":
		cmd.Setup()
	case "compresslogs":
		cmd.CompressLogs()
	default:
		fmt.Println("Unknown command:", command)
	}
}

func extractFiles(embeddedFS embed.FS, targetDir string) error {
	return fs.WalkDir(embeddedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && !strings.Contains(path, "__pycache__") {
			fileContent, err := embeddedFS.ReadFile(path)
			if err != nil {
				return err
			}
			filePath := filepath.Join(targetDir, path)

			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(filePath, fileContent, 0755); err != nil {
				return err
			}
		}
		return nil
	})
}

func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line in .env file: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("error setting environment variable: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %v", err)
	}

	return nil
}
