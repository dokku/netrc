package main

import (
	"context"
	"fmt"
	"netrc/commands"
	"os"

	"github.com/josegonzalez/cli-skeleton/command"
	"github.com/mitchellh/cli"
)

var AppName = "netrc"

var Version string

func main() {
	os.Exit(Run(os.Args[1:]))
}

// Executes the specified command
func Run(args []string) int {
	ctx := context.Background()
	commandMeta := command.SetupRun(ctx, AppName, Version, args)
	c := cli.NewCLI(AppName, Version)
	c.Args = os.Args[1:]
	c.Commands = command.Commands(ctx, commandMeta, Commands)
	exitCode, err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}

// Returns a list of implemented commands
func Commands(ctx context.Context, meta command.Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"set": func() (cli.Command, error) {
			return &commands.SetCommand{Meta: meta}, nil
		},
		"get": func() (cli.Command, error) {
			return &commands.GetCommand{Meta: meta}, nil
		},
		"unset": func() (cli.Command, error) {
			return &commands.UnsetCommand{Meta: meta}, nil
		},
		"version": func() (cli.Command, error) {
			return &command.VersionCommand{Meta: meta}, nil
		},
	}
}
