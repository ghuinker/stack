package cmd

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
)

type ManageContext struct {
	OutDir string
	Dist   fs.FS
}

var GlobalContext *ManageContext

func setPythonEnv(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONPATH=%s:%s", GlobalContext.OutDir, filepath.Join(GlobalContext.OutDir, "venv/lib/python3.12/site-packages")))
	cmd.Env = append(cmd.Env, "DIST_DIR="+GlobalContext.OutDir)
}

func setDevPythonEnv(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONPATH=%s:%s", "app", filepath.Join(GlobalContext.OutDir, "venv/lib/python3.12/site-packages")))
	cmd.Env = append(cmd.Env, "DIST_DIR="+GlobalContext.OutDir)
}
