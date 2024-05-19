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

//go:embed app
//go:embed venv/bin/gunicorn
//go:embed manage.py
//go:embed static/assets/manifest.json
//go:embed all:venv/lib/python3.12/site-packages/asgiref all:venv/lib/python3.12/site-packages/Django all:venv/lib/python3.12/site-packages/environs all:venv/lib/python3.12/site-packages/gunicorn all:venv/lib/python3.12/site-packages/marshmallow all:venv/lib/python3.12/site-packages/packaging all:venv/lib/python3.12/site-packages/dotenv all:venv/lib/python3.12/site-packages/sqlparse
var embeddedFiles embed.FS

//go:embed static/*
var staticFiles embed.FS

/* TODO:
1. Logging
2. gunicorn retries?
*/

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run script.go <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	loadEnvFile(".env")

	// 1. Create a temporary directory
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
	default:
		fmt.Println("Unknown command:", command)
	}
}

func extractFiles(embeddedFS embed.FS, targetDir string) error {
	return fs.WalkDir(embeddedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
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

		// Ignore comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Split the line into key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line in .env file: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set the environment variable
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("error setting environment variable: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %v", err)
	}

	return nil
}
