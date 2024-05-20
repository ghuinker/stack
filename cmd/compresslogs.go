package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CompressLogs() {
	compressLogs("logs")
}

func compressLogs(directory string) {
	// Generate zip file name based on today's date
	today := time.Now().Format("2006-01-02")
	zipFileName := fmt.Sprintf("logs/logs_%s.zip", today)

	// Create a new zip file
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		log.Fatalf("Error creating zip file: %v", err)
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through files in the logs directory
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing file %s: %v", path, err)
			return nil
		}
		// Only include .log files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".log") {
			// Open the log file
			logFile, err := os.Open(path)
			if err != nil {
				log.Printf("Error opening log file %s: %v", path, err)
				return nil
			}
			defer logFile.Close()

			// Create a new file in the zip archive
			zipEntry, err := zipWriter.Create(info.Name())
			if err != nil {
				log.Printf("Error creating zip entry for %s: %v", info.Name(), err)
				return nil
			}

			// Copy the log file content to the zip file
			_, err = io.Copy(zipEntry, logFile)
			if err != nil {
				log.Printf("Error copying log file content for %s: %v", info.Name(), err)
				return nil
			}

			// Remove content from the log file
			if err := removeContentFromLogFile(path); err != nil {
				log.Printf("Error removing content from log file %s: %v", path, err)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error compressing logs: %v", err)
	}

	log.Printf("Logs compressed to %s", zipFileName)
}

func removeContentFromLogFile(path string) error {
	// Open the log file with write-only mode and truncate
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	return nil
}
