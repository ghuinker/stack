package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"stack/cmd"
)

//go:embed app
//go:embed venv/bin/gunicorn
//go:embed manage.py
//go:embed all:venv/lib/python3.12/site-packages/asgiref all:venv/lib/python3.12/site-packages/Django all:venv/lib/python3.12/site-packages/environs all:venv/lib/python3.12/site-packages/gunicorn all:venv/lib/python3.12/site-packages/marshmallow all:venv/lib/python3.12/site-packages/packaging all:venv/lib/python3.12/site-packages/dotenv all:venv/lib/python3.12/site-packages/sqlparse
var embeddedFiles embed.FS

//go:embed static/*
var staticFiles embed.FS

/* TODO:
1. Setup yarn
2. Import DDVVT into this
3. So I can make setup and it is ready
4. Make a command for updating the embeds
*/

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run script.go <command>")
		os.Exit(1)
	}

	command := os.Args[1]

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
