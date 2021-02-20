package main

import (
	"fmt"
	"os"

	"netrc/command"
	"netrc/commands"
	"netrc/meta"

	"github.com/mitchellh/cli"
)

var Version string

func main() {
	os.Exit(Run(os.Args[1:]))
}

func Run(args []string) int {
	c := cli.NewCLI(meta.AppName, Version)
	c.Args = os.Args[1:]
	c.Commands = commands.Commands(command.SetupRun(meta.AppName, Version, args))
	exitCode, err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}
