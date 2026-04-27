package commands

import (
	"os"
	"os/user"
	"path/filepath"
)

func resolveNetrcPath(flagValue string) (string, error) {
	if flagValue != "" {
		return filepath.Clean(flagValue), nil
	}

	if env := os.Getenv("NETRC"); env != "" {
		return filepath.Clean(env), nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(usr.HomeDir, ".netrc"), nil
}

func ensureNetrcExists(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return err
	}

	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	return file.Close()
}
