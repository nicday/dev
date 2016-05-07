package cmd

import (
	"os"
	"os/exec"
)

func RunWithAttachedOutput(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
