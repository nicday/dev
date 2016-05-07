package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/nicday/dev/cmd"
	"github.com/nicday/dev/templates"
	"github.com/nicday/dev/update"
)

var (
	additionalArgs []string

	// Name is the app name and plugin prefix.
	Name = "dev"

	// PluginPrefix is the prefix used when calling a plugin binary.
	PluginPrefix = fmt.Sprintf("%s-", Name)
)

func init() {
	// Peel off any additional args so we only parse them when required or pass them directly to another command.
	if len(os.Args) > 2 {
		additionalArgs = os.Args[2:]
		os.Args = os.Args[:2]
	}
}

func assertInitialized() error {
	createConfigDir()
	writeSupportServicesComposeYML()

	return nil
}

func createConfigDir() error {
	_, err := os.Stat(configDir())
	if err != nil {
		// Check the error type here
		return os.Mkdir(configDir(), 0750)
	}

	return nil
}

func configDir() string {
	return filepath.Join(os.Getenv("HOME"), fmt.Sprintf(".%s", Name))
}

func writeSupportServicesComposeYML() error {
	return ioutil.WriteFile(supportServicesComposeYML(), templates.SupportServicesCompose, 0644)
}

func supportServicesComposeYML() string {
	return filepath.Join(configDir(), "support_services.yml")
}

func main() {
	if ok := update.CatchInitDNS(); ok {
		return
	}

	err := assertInitialized()
	if err != nil {
		fmt.Printf("Unable to initialize %s tool\n", Name)
		return
	}

	app := cli.NewApp()
	app.Name = Name
	app.Usage = "a self-contained, mostly zero-configuration environment"
	app.Version = "0.2"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Nic Day",
			Email: "nic.day@me.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		if c.Args().First() == "" {
			cli.ShowAppHelp(c)
			return nil
		}
		return callPlugin(c)
	}

	app.Commands = []cli.Command{
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "Build or rebuild services",
			Action:  proxyUnmodifiedToCompose,
		},
		{
			Name:    "kill",
			Aliases: []string{"k"},
			Usage:   "Kill containers",
			Action:  proxyUnmodifiedToCompose,
		},
		{
			Name:    "logs",
			Aliases: []string{"l"},
			Usage:   "View output from containers",
			Action:  proxyUnmodifiedToCompose,
		},
		{
			Name:   "ps",
			Usage:  "List containers",
			Action: proxyUnmodifiedToCompose,
		},
		{
			Name:    "pull",
			Aliases: []string{"p"},
			Usage:   "Pulls service images",
			Action:  proxyUnmodifiedToCompose,
		},
		{
			Name:    "restart",
			Aliases: []string{"r"},
			Usage:   "Restart services",
			Action:  proxyUnmodifiedToCompose,
		},
		{
			Name:   "rm",
			Usage:  "Remove stopped containers",
			Action: proxyUnmodifiedToCompose,
		},
		{
			Name:   "run",
			Usage:  "Run a one-off command",
			Action: run,
		},
		{
			Name:   "scale",
			Usage:  "Scales services",
			Action: proxyUnmodifiedToCompose,
		},
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "Start services",
			Action:  proxyUnmodifiedToCompose,
		},
		{
			Name:    "stop",
			Aliases: []string{"st"},
			Usage:   "Stops services",
			Action:  proxyUnmodifiedToCompose,
		},
		{
			Name:    "up",
			Aliases: []string{"u"},
			Usage:   "Create and start containers",
			Action:  up,
		},
		{
			Name:    "support-services",
			Aliases: []string{"ss"},
			Usage:   "Manage support services (nginx and dnsmasq)",
			Action:  supportServices,
		},
		{
			Name:    "update",
			Aliases: []string{"U"},
			Usage:   "Update Campground developer environment",
			Action:  update.ActionFn,
		},
	}

	app.Run(os.Args)
}

func proxyUnmodifiedToCompose(c *cli.Context) error {
	args := []string{c.Command.FullName()}
	if c.NArg() != 0 {
		args = append(args, c.Args()...)
	}
	args = append(args, additionalArgs...)

	return cmd.RunWithAttachedOutput("docker-compose", args...)
}

func run(c *cli.Context) error {
	fmt.Println("docker-compose", "--rm", c.Args())
	return nil
}

func up(c *cli.Context) error {
	supportServices(supportServicesContext(c))

	args := upArgs(c.Args())
	args = append(args, additionalArgs...)

	return cmd.RunWithAttachedOutput("docker-compose", args...)
}

func upArgs(contextArgs cli.Args) []string {
	args := []string{
		"up",
		"-d",
	}

	if len(contextArgs) != 0 {
		args = append(args, contextArgs...)
	}

	args = append(args, additionalArgs...)

	return args
}

func supportServicesContext(parentCtx *cli.Context) *cli.Context {
	flags := flag.NewFlagSet("campground", flag.ContinueOnError)
	flags.Parse([]string{"up"})
	return cli.NewContext(parentCtx.App, flags, parentCtx)
}

// supportServices starts the support service containers
func supportServices(c *cli.Context) error {
	args := supportServicesArgs(c.Args())

	return cmd.RunWithAttachedOutput("docker-compose", args...)
}

func supportServicesArgs(contextArgs cli.Args) []string {
	args := []string{
		"-f",
		supportServicesComposeYML(),
		"-p",
		"dev",
	}
	if contextArgs.First() != "" {
		args = append(args, contextArgs.First())
	}

	if contextArgs.First() == "up" {
		args = append(args, "-d")
	}

	if len(contextArgs.Tail()) != 0 {
		args = append(args, contextArgs.Tail()...)
	}
	args = append(args, additionalArgs...)

	return args
}

// callPlugin checks if there is a there is a binary prefixed with PluginPrefix installed, if so, it is called with the
// tail args.
func callPlugin(c *cli.Context) error {
	plugin := fmt.Sprintf("%s%s", PluginPrefix, c.Args().First())
	if !cmd.Installed(plugin) {
		fmt.Printf("dev: '%s' is not a dev command.\n", c.Args().First())
		return nil
	}

	var args []string
	if len(c.Args().Tail()) != 0 {
		args = append(args, c.Args().Tail()...)
	}
	args = append(args, additionalArgs...)

	fmt.Println("calling", c.Args().First(), args)

	cmd := exec.Command(plugin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Unable to run %s plugin: %s\n", plugin, err)
		return err
	}

	return nil
}
