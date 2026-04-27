package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

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

func (c *SetCommand) stdinArguments() []command.Argument {
	return []command.Argument{
		{Name: "name", Optional: false, Type: command.ArgumentString},
		{Name: "login", Optional: false, Type: command.ArgumentString},
		{Name: "account", Optional: true, Type: command.ArgumentString},
	}
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
		"Set an entry in the .netrc file":          fmt.Sprintf("%s %s github.com username password", appName, c.Name()),
		"Set an entry, reading password from stdin": fmt.Sprintf("echo \"$PW\" | %s %s github.com username --stdin", appName, c.Name()),
	}
}

func (c *SetCommand) FlagSet() *pflag.FlagSet {
	fs := c.Meta.FlagSet(c.Name(), command.FlagSetClient)
	fs.String("netrc-file", "", "path to the netrc file (overrides $NETRC and ~/.netrc)")
	fs.Bool("stdin", false, "read password from stdin instead of as a positional argument")
	return fs
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

	useStdin, _ := flags.GetBool("stdin")

	argDefs := c.Arguments()
	if useStdin {
		argDefs = c.stdinArguments()
	}

	arguments, err := command.ParseArguments(flags.Args(), argDefs)
	if err != nil {
		c.Ui.Error(err.Error())
		c.Ui.Error(command.CommandErrorText(c))
		return 1
	}

	name := arguments["name"].StringValue()
	login := arguments["login"].StringValue()
	account := arguments["account"].StringValue()

	var password string
	if useStdin {
		password, err = readPasswordFromStdin(os.Stdin)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
	} else {
		password = arguments["password"].StringValue()
	}

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

	machine := n.Machine(name)
	if machine == nil {
		n.AddMachine(name, login, password)
		machine = n.Machine(name)
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

func readPasswordFromStdin(r io.Reader) (string, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read password from stdin: %w", err)
	}

	password := strings.TrimRight(string(buf), "\n")
	if password == "" {
		return "", fmt.Errorf("--stdin given but stdin was empty")
	}

	return password, nil
}
