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

func (c *SetCommand) flagArguments() []command.Argument {
	return []command.Argument{
		{Name: "name", Optional: false, Type: command.ArgumentString},
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
		"Set an entry in the .netrc file":           fmt.Sprintf("%s %s github.com username password", appName, c.Name()),
		"Set an entry, reading password from stdin": fmt.Sprintf("echo \"$PW\" | %s %s github.com username --stdin", appName, c.Name()),
		"Rotate the password on an existing entry":  fmt.Sprintf("%s %s github.com --password newpassword", appName, c.Name()),
	}
}

func (c *SetCommand) FlagSet() *pflag.FlagSet {
	fs := c.Meta.FlagSet(c.Name(), command.FlagSetClient)
	fs.String("netrc-file", "", "path to the netrc file (overrides $NETRC and ~/.netrc)")
	fs.Bool("stdin", false, "read password from stdin instead of as a positional argument")
	fs.String("login", "", "set the login field; other fields preserved if the entry exists")
	fs.String("password", "", "set the password field; other fields preserved if the entry exists")
	fs.String("account", "", "set the account field; pass an empty string to clear it")
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
	loginChanged := flags.Changed("login")
	passwordChanged := flags.Changed("password")
	accountChanged := flags.Changed("account")
	fieldFlagSet := loginChanged || passwordChanged || accountChanged

	if useStdin && passwordChanged {
		c.Ui.Error("--stdin and --password are mutually exclusive")
		return 1
	}

	argDefs := c.Arguments()
	switch {
	case fieldFlagSet:
		argDefs = c.flagArguments()
	case useStdin:
		argDefs = c.stdinArguments()
	}

	arguments, err := command.ParseArguments(flags.Args(), argDefs)
	if err != nil {
		c.Ui.Error(err.Error())
		c.Ui.Error(command.CommandErrorText(c))
		return 1
	}

	positional := func(name string) (string, bool) {
		if arg, ok := arguments[name]; ok && arg.HasValue {
			return arg.StringValue(), true
		}
		return "", false
	}

	name, _ := positional("name")
	loginPos, loginPosSet := positional("login")
	passwordPos, passwordPosSet := positional("password")
	accountPos, accountPosSet := positional("account")

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

	var machine *netrc.Machine
	if name == defaultMachineName {
		machine = findDefault(n)
	} else {
		machine = n.Machine(name)
	}
	existingLogin, existingPassword := "", ""
	if machine != nil {
		existingLogin = machine.Get("login")
		existingPassword = machine.Get("password")
	}

	loginFlag, _ := flags.GetString("login")
	passwordFlag, _ := flags.GetString("password")
	accountFlag, _ := flags.GetString("account")

	login := existingLogin
	switch {
	case loginChanged:
		login = loginFlag
	case loginPosSet:
		login = loginPos
	}

	var password string
	switch {
	case passwordChanged:
		password = passwordFlag
	case useStdin:
		password, err = readPasswordFromStdin(os.Stdin)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
	case passwordPosSet:
		password = passwordPos
	default:
		password = existingPassword
	}

	account := ""
	accountTouched := false
	switch {
	case accountChanged:
		account = accountFlag
		accountTouched = true
	case accountPosSet && accountPos != "":
		account = accountPos
		accountTouched = true
	}

	if machine == nil {
		if login == "" || password == "" {
			c.Ui.Error(fmt.Sprintf("Cannot create new entry '%s' without login and password", name))
			return 1
		}
		if name == defaultMachineName {
			n, machine, err = getOrCreateDefault(n)
			if err != nil {
				c.Ui.Error(err.Error())
				return 1
			}
			machine.Set("login", login)
			machine.Set("password", password)
		} else {
			n.AddMachine(name, login, password)
			machine = n.Machine(name)
		}
	} else {
		machine.Set("login", login)
		machine.Set("password", password)
	}

	if accountTouched {
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
