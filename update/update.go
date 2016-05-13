package update

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/nicday/dev/cmd"
	"github.com/nicday/dev/templates"
)

var (
	AppRepo = "github.com/nicday/dev"
)

func ActionFn(c *cli.Context) error {
	err := updateApp()
	if err != nil {
		fmt.Println("Error updating app:", err)
		return err
	}

	if ok := cmd.Installed("brew"); !ok {
		err := installBrew()
		if err != nil {
			return err
		}
	}

	if ok := cmd.Installed("docker"); !ok {
		err := brewInstall("docker")
		if err != nil {
			return err
		}
	}

	if ok := cmd.Installed("docker-compose"); !ok {
		err := brewInstall("docker-compose")
		if err != nil {
			return err
		}
	}

	err = sudoInitDNS()
	if err != nil {
		return err
	}

	fmt.Println("All up to date.")

	return nil
}

func installBrew() error {
	c := exec.Command(
		"/usr/bin/ruby",
		"-e",
		`"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`,
	)

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

func brewInstall(pkg string) error {
	c := exec.Command(
		"brew",
		"install",
		pkg,
	)

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

func updateApp() error {
	c := exec.Command("go", "get", "-u", AppRepo)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

func sudoInitDNS() error {
	if !haveSudo() {
		fmt.Println("Please enter your password to initalize DNS:")
	}

	err := cmd.RunWithAttachedOutput("sudo", "dev", "init-dns")
	if err != nil {
		return err
	}

	return nil
}

func CatchInitDNS() bool {
	if len(os.Args) > 1 {
		if os.Args[1] == "init-dns" {
			err := initDNS()
			if err != nil {
				fmt.Println("Unable to initialize DNS:", err)
			}
			return true
		}
	}

	return false
}

func initDNS() error {
	err := createResolverDir()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("/etc/resolver/something", templates.ResolverDev, 0755)
	if err != nil {
		return err
	}

	return nil
}

func createResolverDir() error {
	_, err := os.Stat("/etc/resolver")
	if err != nil {
		// Check the error type here
		return os.Mkdir("/etc/resolver", 0755)
	}

	return nil
}

func haveSudo() bool {
	cmd := exec.Command("sudo", "-n", "true")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if _, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return false
			}
		}
	}

	return true
}
