package commands

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/jdxcode/netrc"
	"github.com/josegonzalez/cli-skeleton/command"
	"github.com/posener/complete"
	"github.com/spf13/pflag"
)

type SetCommand struct {
	command.Meta
}

func (c *SetCommand) Help() string {
	return command.CommandHelp(c)
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
	return command.MergeAutocompleteFlags(
		c.Meta.AutocompleteFlags(command.FlagSetClient),
		complete.Flags{},
	)
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

func (c *SetCommand) FlagSet() *pflag.FlagSet {
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
	account := arguments["account"].StringValue()

	usr, err := user.Current()
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	netrcFile := filepath.Join(usr.HomeDir, ".netrc")
	if _, err := os.Stat(netrcFile); os.IsNotExist(err) {
		file, err := os.OpenFile(netrcFile, os.O_RDONLY|os.O_CREATE, 0600)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		if err := file.Close(); err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
	}

	n, err := netrc.Parse(netrcFile)
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

	if account != "" {
		machine.Set("account", account)
	}

	if err := n.Save(); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	return 0
}
