package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var overrides = map[string]string{
	"python-dotenv": "dotenv",
}

func readPackages(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var packages []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		// Ignore comments and empty lines
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Split on the first occurrence of '=' or '>' or '<'
		parts := strings.FieldsFunc(line, func(r rune) bool {
			return r == '=' || r == '>' || r == '<'
		})
		// Add the package name (first part) to the list
		if len(parts) > 0 {
			packages = append(packages, strings.TrimSpace(parts[0]))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return packages, nil
}

func UpdateEmbed() {
	var builder strings.Builder
	builder.WriteString("//go:embed ")

	packages, err := readPackages("requirements.txt")
	if err != nil {
		fmt.Println("Error reading requirements.txt: ", err)
		return
	}
	for _, pkg := range packages {
		builder.WriteString("all:venv/lib/python3.12/site-packages/")
		override, ok := overrides[pkg]
		if ok {
			builder.WriteString(override)
		} else {
			builder.WriteString(pkg)
		}
		builder.WriteString(" ")
	}

	err = replaceLineInFile("main.go", "//go:embed all:venv", builder.String())
	if err != nil {
		fmt.Println("Error updating file: ", err)
	}
}

func replaceLineInFile(filename, search, replace string) error {
	// Read the original file content
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, search) {
			line = replace
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Write the modified content back to the file
	file, err = os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("could not open file for writing: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %v", err)
	}

	return nil
}
