package main

import (
	"fmt"
	"os"

	"netrc/command"
	"netrc/commands"
	"netrc/meta"

	"github.com/mattn/go-colorable"
	"github.com/mitchellh/cli"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	os.Exit(Run(os.Args[1:]))
}

func Run(args []string) int {
	return RunCustom(args)
}

func RunCustom(args []string) int {
	// Parse flags into env vars for global use
	args = setupEnv(args)

	// Create the meta object
	metaPtr := new(command.Meta)

	// Don't use color if disabled
	color := true
	if os.Getenv(command.EnvCLINoColor) != "" {
		color = false
	}

	isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	metaPtr.Ui = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      colorable.NewColorableStdout(),
		ErrorWriter: colorable.NewColorableStderr(),
	}

	// The Dokku command never outputs color
	agentUi := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	// Only use colored UI if stdout is a tty, and not disabled
	if isTerminal && color {
		metaPtr.Ui = &cli.ColoredUi{
			ErrorColor: cli.UiColorRed,
			WarnColor:  cli.UiColorYellow,
			InfoColor:  cli.UiColorGreen,
			Ui:         metaPtr.Ui,
		}
	}

	os.Setenv("CLI_APP_NAME", meta.AppName)
	c := cli.NewCLI(meta.AppName, "0.0.1")
	c.Args = os.Args[1:]
	c.Commands = commands.Commands(metaPtr, agentUi)

	exitCode, err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}

// setupEnv parses args and may replace them and sets some env vars to known
// values based on format options
func setupEnv(args []string) []string {
	noColor := false
	for _, arg := range args {
		// Check if color is set
		if arg == "-no-color" || arg == "--no-color" {
			noColor = true
		}
	}

	// Put back into the env for later
	if noColor {
		os.Setenv(command.EnvCLINoColor, "true")
	}

	return args
}
