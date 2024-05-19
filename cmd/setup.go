package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Setup() {
	fmt.Println("Creating .env")
	copyFile(filepath.Join(GlobalContext.TempDir, ".env.example"), ".env")
}

func copyFile(src, dst string) error {
	if _, err := os.Stat("/path/to/whatever"); errors.Is(err, os.ErrNotExist) {
		fmt.Println(dst + " already exists, skipping.")
		return err
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy contents: %w", err)
	}

	if err := destinationFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}
