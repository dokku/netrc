package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jdxcode/netrc"
	"github.com/josegonzalez/cli-skeleton/command"
	"github.com/posener/complete"
	"github.com/spf13/pflag"
)

var validGetFields = []string{"login", "password", "account"}
var defaultGetFields = []string{"login", "password"}

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
		"Get an entry from the .netrc file":             fmt.Sprintf("%s %s github.com", appName, c.Name()),
		"Get a single field":                            fmt.Sprintf("%s %s github.com --field password", appName, c.Name()),
		"Get specific fields as JSON":                   fmt.Sprintf("%s %s github.com --field login --field password --format json", appName, c.Name()),
		"Get fields as eval-safe shell variable assigns": fmt.Sprintf("%s %s github.com --format shell", appName, c.Name()),
	}
}

func (c *GetCommand) FlagSet() *pflag.FlagSet {
	fs := c.Meta.FlagSet(c.Name(), command.FlagSetClient)
	fs.String("netrc-file", "", "path to the netrc file (overrides $NETRC and ~/.netrc)")
	fs.StringArray("field", []string{}, "field to output (login, password, account); repeatable")
	fs.String("format", "text", "output format: text, json, or shell")
	return fs
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
		c.Ui.Error(fmt.Sprintf("Invalid machine '%v' specified", name))
		return 1
	}

	format, _ := flags.GetString("format")
	if format != "text" && format != "json" && format != "shell" {
		c.Ui.Error(fmt.Sprintf("Invalid format '%s' specified, must be 'text', 'json', or 'shell'", format))
		return 1
	}

	fields, _ := flags.GetStringArray("field")
	userSuppliedFields := len(fields) > 0
	if !userSuppliedFields {
		fields = defaultGetFields
	}
	for _, f := range fields {
		if !isValidGetField(f) {
			c.Ui.Error(fmt.Sprintf("Invalid field '%s' specified, must be one of 'login', 'password', 'account'", f))
			return 1
		}
	}

	if format == "text" && !userSuppliedFields {
		c.Ui.Output(fmt.Sprintf("%s:%s", machine.Get("login"), machine.Get("password")))
		return 0
	}

	switch format {
	case "json":
		out, err := renderGetJSON(machine, fields)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		c.Ui.Output(out)
	case "shell":
		c.Ui.Output(renderGetShell(machine, fields))
	default:
		c.Ui.Output(renderGetText(machine, fields))
	}

	return 0
}

func isValidGetField(f string) bool {
	for _, v := range validGetFields {
		if v == f {
			return true
		}
	}
	return false
}

func renderGetText(m *netrc.Machine, fields []string) string {
	if len(fields) == 1 {
		return m.Get(fields[0])
	}
	lines := make([]string, 0, len(fields))
	for _, f := range fields {
		lines = append(lines, fmt.Sprintf("%s=%s", f, m.Get(f)))
	}
	return strings.Join(lines, "\n")
}

func renderGetJSON(m *netrc.Machine, fields []string) (string, error) {
	out := make(map[string]string, len(fields))
	for _, f := range fields {
		out[f] = m.Get(f)
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func renderGetShell(m *netrc.Machine, fields []string) string {
	lines := make([]string, 0, len(fields))
	for _, f := range fields {
		lines = append(lines, fmt.Sprintf("%s=%s", f, shellEscape(m.Get(f))))
	}
	return strings.Join(lines, "\n")
}

func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
