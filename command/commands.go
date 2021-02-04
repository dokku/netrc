package command

const (
	// EnvCLINoColor is an env var that toggles colored UI output.
	EnvCLINoColor = `CLI_NO_COLOR`
)

// NamedCommand is a interface to denote a commmand's name.
type NamedCommand interface {
	Name() string
}
