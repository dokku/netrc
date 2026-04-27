package commands

import (
	"fmt"
	"os"

	"github.com/jdxcode/netrc"
	"github.com/josegonzalez/cli-skeleton/command"
	"github.com/posener/complete"
	"github.com/spf13/pflag"
)

type RenameCommand struct {
	command.Meta
}

func (c *RenameCommand) Help() string {
	return command.CommandHelp(c)
}

func (c *RenameCommand) Arguments() []command.Argument {
	args := []command.Argument{}
	args = append(args, command.Argument{
		Name:     "old-name",
		Optional: false,
		Type:     command.ArgumentString,
	})
	args = append(args, command.Argument{
		Name:     "new-name",
		Optional: false,
		Type:     command.ArgumentString,
	})
	return args
}

func (c *RenameCommand) AutocompleteFlags() complete.Flags {
	return command.MergeAutocompleteFlags(
		c.Meta.AutocompleteFlags(command.FlagSetClient),
		complete.Flags{},
	)
}

func (c *RenameCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *RenameCommand) Examples() map[string]string {
	appName := os.Getenv("CLI_APP_NAME")
	return map[string]string{
		"Rename an entry in the .netrc file":             fmt.Sprintf("%s %s old.example.com new.example.com", appName, c.Name()),
		"Overwrite an existing destination with --force": fmt.Sprintf("%s %s old.example.com new.example.com --force", appName, c.Name()),
	}
}

func (c *RenameCommand) FlagSet() *pflag.FlagSet {
	fs := c.Meta.FlagSet(c.Name(), command.FlagSetClient)
	fs.String("netrc-file", "", "path to the netrc file (overrides $NETRC and ~/.netrc)")
	fs.Bool("force", false, "overwrite the destination machine if it already exists")
	return fs
}

func (c *RenameCommand) Name() string {
	return "rename"
}

func (c *RenameCommand) ParsedArguments(args []string) (map[string]command.Argument, error) {
	return command.ParseArguments(args, c.Arguments())
}

func (c *RenameCommand) Synopsis() string {
	return "Rename an entry in the .netrc file"
}

func (c *RenameCommand) Run(args []string) int {
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

	oldName := arguments["old-name"].StringValue()
	newName := arguments["new-name"].StringValue()
	force, _ := flags.GetBool("force")

	netrcFlag, _ := flags.GetString("netrc-file")
	netrcFile, err := resolveNetrcPath(netrcFlag)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	if err := ensureNetrcExists(netrcFile); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	n, err := netrc.Parse(netrcFile)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	src := n.Machine(oldName)
	if src == nil {
		c.Ui.Error(fmt.Sprintf("Invalid machine '%s' specified", oldName))
		return 1
	}

	if oldName == newName {
		return 0
	}

	login := src.Get("login")
	password := src.Get("password")
	account := src.Get("account")

	if dst := n.Machine(newName); dst != nil {
		if !force {
			c.Ui.Error(fmt.Sprintf("Machine '%s' already exists, pass --force to overwrite", newName))
			return 1
		}
		c.Ui.Warn(fmt.Sprintf("Warning: overwriting existing machine '%s'", newName))
		n.RemoveMachine(newName)
	}

	if newName == defaultMachineName {
		var dst *netrc.Machine
		n, dst, err = getOrCreateDefault(n)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		dst.Set("login", login)
		dst.Set("password", password)
		if account != "" {
			dst.Set("account", account)
		}
	} else {
		n.AddMachine(newName, login, password)
		if account != "" {
			n.Machine(newName).Set("account", account)
		}
	}
	n.RemoveMachine(oldName)

	if err := n.Save(); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	return 0
}
