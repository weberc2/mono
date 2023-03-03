package agent

import (
	"os"
	"os/exec"
)

func runCmd(caption string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = NewPrefixWriter(os.Stdout, caption)
	cmd.Stderr = NewPrefixWriter(os.Stderr, caption)
	return cmd.Run()
}
