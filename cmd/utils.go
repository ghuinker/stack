package cmd

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
)

type ManageContext struct {
	TempDir     string
	StaticFiles fs.FS
}

var GlobalContext *ManageContext

func setPythonEnv(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONPATH=%s:%s", filepath.Join(GlobalContext.TempDir, "app"), filepath.Join(GlobalContext.TempDir, "venv/lib/python3.12/site-packages")))
}

func setDevPythonEnv(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONPATH=%s:%s", "app", filepath.Join(GlobalContext.TempDir, "venv/lib/python3.12/site-packages")))
}
