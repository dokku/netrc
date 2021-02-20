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

type SetCommand struct {
	command.Meta
}

func (c *SetCommand) Help() string {
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

func (c *SetCommand) Arguments() []command.Argument {
	args := []command.Argument{}
	args = append(args, command.Argument{
		Name:     "name",
		Optional: false,
		Type:     command.ArgumentString,
	})
	args = append(args, command.Argument{
		Name:     "login",
		Optional: false,
		Type:     command.ArgumentString,
	})
	args = append(args, command.Argument{
		Name:     "password",
		Optional: false,
		Type:     command.ArgumentString,
	})
	args = append(args, command.Argument{
		Name:     "account",
		Optional: true,
		Type:     command.ArgumentString,
	})
	return args
}

func (c *SetCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{}
}

func (c *SetCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *SetCommand) Examples() map[string]string {
	appName := os.Getenv("CLI_APP_NAME")
	return map[string]string{
		"Set an entry in the .netrc file": fmt.Sprintf("%s %s github.com username password", appName, c.Name()),
	}
}

func (c *SetCommand) FlagSet() *flag.FlagSet {
	return c.Meta.FlagSet(c.Name(), command.FlagSetClient)
}

func (c *SetCommand) Name() string {
	return "set"
}

func (c *SetCommand) ParsedArguments(args []string) (map[string]command.Argument, error) {
	return command.ParseArguments(args, c.Arguments())
}

func (c *SetCommand) Synopsis() string {
	return "Set an entry in the .netrc file"
}

func (c *SetCommand) Run(args []string) int {
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
	login := arguments["login"].StringValue()
	password := arguments["password"].StringValue()

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
		n.AddMachine(name, login, password)
	} else {
		machine.Set("login", login)
		machine.Set("password", password)
	}

	if err := n.Save(); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	return 0
}
