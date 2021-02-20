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

type GetCommand struct {
	command.Meta
}

func (c *GetCommand) Help() string {
	return command.CommandHelp(c)
}

func (c *GetCommand) Arguments() []command.Argument {
	args := []command.Argument{}
	args = append(args, command.Argument{
		Name:     "name",
		Optional: false,
		Type:     command.ArgumentString,
	})
	return args
}

func (c *GetCommand) AutocompleteFlags() complete.Flags {
	return command.MergeAutocompleteFlags(
		c.Meta.AutocompleteFlags(command.FlagSetClient),
		complete.Flags{},
	)
}

func (c *GetCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *GetCommand) Examples() map[string]string {
	appName := os.Getenv("CLI_APP_NAME")
	return map[string]string{
		"Get an entry from the .netrc file": fmt.Sprintf("%s %s github.com", appName, c.Name()),
	}
}

func (c *GetCommand) FlagSet() *pflag.FlagSet {
	return c.Meta.FlagSet(c.Name(), command.FlagSetClient)
}

func (c *GetCommand) Name() string {
	return "get"
}

func (c *GetCommand) ParsedArguments(args []string) (map[string]command.Argument, error) {
	return command.ParseArguments(args, c.Arguments())
}

func (c *GetCommand) Synopsis() string {
	return "Get an entry from the .netrc file"
}

func (c *GetCommand) Run(args []string) int {
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

	login := machine.Get("login")
	password := machine.Get("password")
	c.Ui.Output(fmt.Sprintf("%s:%s", login, password))

	return 0
}
