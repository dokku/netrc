package commands

import (
	"os"

	"netrc/command"

	"github.com/mattn/go-colorable"
	"github.com/mitchellh/cli"
)

// Commands returns the mapping of CLI commands. The meta
// parameter lets you set meta options for all commands.
func Commands(metaPtr *command.Meta, agentUi cli.Ui) map[string]cli.CommandFactory {
	if metaPtr == nil {
		metaPtr = new(command.Meta)
	}

	meta := *metaPtr
	if meta.Ui == nil {
		meta.Ui = &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      colorable.NewColorableStdout(),
			ErrorWriter: colorable.NewColorableStderr(),
		}
	}

	all := map[string]cli.CommandFactory{}

	for k, v := range SubCommands(meta) {
		all[k] = v
	}

	return all
}

func SubCommands(meta command.Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"set": func() (cli.Command, error) {
			return &SetCommand{Meta: meta}, nil
		},
		"get": func() (cli.Command, error) {
			return &GetCommand{Meta: meta}, nil
		},
		"unset": func() (cli.Command, error) {
			return &UnsetCommand{Meta: meta}, nil
		},
		"version": func() (cli.Command, error) {
			return &VersionCommand{Meta: meta}, nil
		},
	}
}
