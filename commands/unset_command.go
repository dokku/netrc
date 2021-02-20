package commands

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"netrc/command"

	"github.com/jdxcode/netrc"
	"github.com/posener/complete"
)

type UnsetCommand struct {
	command.Meta
}

func (c *UnsetCommand) Help() string {
	appName := os.Getenv("CLI_APP_NAME")
	helpText := `
Usage: ` + appName + ` ` + c.Name() + ` ` + command.FlagString(c.FlagSet()) + ` ` + command.ArgumentAsString(c.Arguments()) + `

  ` + c.Synopsis() + `

General Options:
  ` + command.GeneralOptionsUsage() + `

Example:

` + command.ExampleString(c.Examples())

	return strings.TrimSpace(helpText)
}

func (c *UnsetCommand) Arguments() []command.Argument {
	args := []command.Argument{}
	args = append(args, command.Argument{
		Name:     "name",
		Optional: false,
		Type:     command.ArgumentString,
	})
	return args
}

func (c *UnsetCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{}
}

func (c *UnsetCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *UnsetCommand) Examples() map[string]string {
	appName := os.Getenv("CLI_APP_NAME")
	return map[string]string{
		"Unset an entry in the .netrc file": fmt.Sprintf("%s %s github.com", appName, c.Name()),
	}
}

func (c *UnsetCommand) FlagSet() *flag.FlagSet {
	return c.Meta.FlagSet(c.Name(), command.FlagSetClient)
}

func (c *UnsetCommand) Name() string {
	return "unset"
}

func (c *UnsetCommand) ParsedArguments(args []string) (map[string]command.Argument, error) {
	return command.ParseArguments(args, c.Arguments())
}

func (c *UnsetCommand) Synopsis() string {
	return "Unset an entry from the .netrc file"
}

func (c *UnsetCommand) Run(args []string) int {
	flags := c.FlagSet()
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return 1
	}

	arguments, err := c.ParsedArguments(flags.Args())
	if err != nil {
		c.Ui.Error(err.Error())
		c.Ui.Error(command.CommandErrorText(c))
		return 1
	}

	name := arguments["name"].StringValue()

	usr, err := user.Current()
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	machine := n.Machine(name)
	if machine == nil {
		c.Ui.Error(fmt.Sprintf("Invalid machine '%v' specified", name))
		return 1
	}

	n.RemoveMachine(name)
	if err := n.Save(); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	return 0
}
