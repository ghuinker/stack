package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Manage() {
	runManageCommand(os.Args[2:])
}

func runManageCommand(args []string) error {
	cmdArgs := append([]string{filepath.Join(GlobalContext.OutDir, "manage.py")}, args...)
	cmd := exec.Command("python3", cmdArgs...)
	setPythonEnv(cmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		fmt.Println("Error running manage command:", err)
	}

	return err
}
