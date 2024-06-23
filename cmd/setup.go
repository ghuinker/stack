package cmd

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

func Setup() {
	fmt.Println("Creating .env")
	err := copyFile(filepath.Join(GlobalContext.OutDir, ".env.example"), ".env")
	if err == nil {
		fmt.Println("Updating Secret Key")
		updateSecretKey()
	}
}

func updateSecretKey() {
	secretKey := getRandomSecretKey()

	// Read the contents of the file
	filePath := ".env"
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Replace the secret key in the file content
	newFileContent := strings.Replace(string(fileContent), "SECRET_KEY=", fmt.Sprintf("SECRET_KEY=%s", secretKey), 1)

	// Write the modified content back to the file
	if err := os.WriteFile(filePath, []byte(newFileContent), 0644); err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	fmt.Println("Secret key updated successfully.")
}

func getRandomSecretKey() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*(-_=+)"
	return getRandomString(50, chars)
}

func getRandomString(length int, charset string) string {
	var result string
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		result += string(charset[n.Int64()])
	}
	return result
}

func copyFile(src, dst string) error {
	if _, err := os.Stat(dst); !errors.Is(err, os.ErrNotExist) {
		fmt.Println(dst + " already exists, skipping.")
		return os.ErrNotExist
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
