package cmd

import (
	"fmt"
	"os/exec"
	"syscall"
)

func Installed(cmd string) bool {
	c := exec.Command("which", "-s", cmd)
	err := c.Start()
	if err != nil {
		fmt.Println(err)
	}

	if err := c.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() != 0 {
					return false
				}
			}
		}
	}

	return true
}
