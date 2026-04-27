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

type ListCommand struct {
	command.Meta
}

type machineEntry struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Account  string `json:"account"`
}

func (c *ListCommand) Help() string {
	return command.CommandHelp(c)
}

func (c *ListCommand) Arguments() []command.Argument {
	return []command.Argument{}
}

func (c *ListCommand) AutocompleteFlags() complete.Flags {
	return command.MergeAutocompleteFlags(
		c.Meta.AutocompleteFlags(command.FlagSetClient),
		complete.Flags{},
	)
}

func (c *ListCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *ListCommand) Examples() map[string]string {
	appName := os.Getenv("CLI_APP_NAME")
	return map[string]string{
		"List all machines in the .netrc file":              fmt.Sprintf("%s %s", appName, c.Name()),
		"List machines including the default block as JSON": fmt.Sprintf("%s %s --include-default --format json", appName, c.Name()),
	}
}

func (c *ListCommand) FlagSet() *pflag.FlagSet {
	fs := c.Meta.FlagSet(c.Name(), command.FlagSetClient)
	fs.String("netrc-file", "", "path to the netrc file (overrides $NETRC and ~/.netrc)")
	fs.String("format", "text", "output format: text or json")
	fs.Bool("with-fields", false, "include login/password/account for each machine")
	fs.Bool("include-default", false, "include the default block in the output")
	return fs
}

func (c *ListCommand) Name() string {
	return "list"
}

func (c *ListCommand) ParsedArguments(args []string) (map[string]command.Argument, error) {
	return command.ParseArguments(args, c.Arguments())
}

func (c *ListCommand) Synopsis() string {
	return "List machines in the .netrc file"
}

func (c *ListCommand) Run(args []string) int {
	flags := c.FlagSet()
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return 1
	}

	if _, err := c.ParsedArguments(flags.Args()); err != nil {
		c.Ui.Error(err.Error())
		c.Ui.Error(command.CommandErrorText(c))
		return 1
	}

	format, _ := flags.GetString("format")
	if format != "text" && format != "json" {
		c.Ui.Error(fmt.Sprintf("Invalid format '%s' specified, must be 'text' or 'json'", format))
		return 1
	}

	withFields, _ := flags.GetBool("with-fields")
	includeDefault, _ := flags.GetBool("include-default")

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

	machines := make([]*netrc.Machine, 0, len(n.Machines()))
	for _, m := range n.Machines() {
		if m.IsDefault && !includeDefault {
			continue
		}
		machines = append(machines, m)
	}

	if format == "json" {
		out, err := renderJSON(machines, withFields)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		c.Ui.Output(out)
		return 0
	}

	if out := renderText(machines, withFields); out != "" {
		c.Ui.Output(out)
	}
	return 0
}

func renderText(machines []*netrc.Machine, withFields bool) string {
	var lines []string
	for _, m := range machines {
		if !withFields {
			lines = append(lines, m.Name)
			continue
		}
		parts := []string{
			m.Name,
			fmt.Sprintf("login=%s", m.Get("login")),
			fmt.Sprintf("password=%s", m.Get("password")),
		}
		if account := m.Get("account"); account != "" {
			parts = append(parts, fmt.Sprintf("account=%s", account))
		}
		lines = append(lines, strings.Join(parts, "\t"))
	}
	return strings.Join(lines, "\n")
}

func renderJSON(machines []*netrc.Machine, withFields bool) (string, error) {
	if !withFields {
		names := make([]string, 0, len(machines))
		for _, m := range machines {
			names = append(names, m.Name)
		}
		b, err := json.MarshalIndent(names, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	entries := make([]machineEntry, 0, len(machines))
	for _, m := range machines {
		entries = append(entries, machineEntry{
			Name:     m.Name,
			Login:    m.Get("login"),
			Password: m.Get("password"),
			Account:  m.Get("account"),
		})
	}
	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
