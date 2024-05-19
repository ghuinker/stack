package cmd

import (
	"fmt"
	"os"
)

func CreateSuperUser() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: stack createsuperuser <username> <email>")
		os.Exit(1) // Exit the program with an error code
	}

	username := os.Args[2]
	email := os.Args[3]

	if username == "" || email == "" {
		fmt.Println("Error: username and email cannot be empty")
		os.Exit(1) // Exit the program with an error code
	}

	err := runManageCommand([]string{"createsuperuser", "--no-input", "--username", username, "--email", email})
	if err != nil {
		os.Exit(1) // Exit the program with an error code
	}
	runManageCommand([]string{"changepassword", username})
}
