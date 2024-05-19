package cmd

import (
	"os"
)

func Shell() {
	runManageCommand(os.Args[1:])
}
