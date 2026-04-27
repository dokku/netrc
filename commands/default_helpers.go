package commands

import (
	"strings"

	"github.com/jdxcode/netrc"
)

const defaultMachineName = "default"

func findDefault(n *netrc.Netrc) *netrc.Machine {
	for _, m := range n.Machines() {
		if m.IsDefault {
			return m
		}
	}
	return nil
}

// getOrCreateDefault returns the existing default block from n or creates one.
// AddMachine always emits "machine <name>" and Machine.tokens is unexported, so
// the only public-API path to a fresh default block is to re-parse rendered
// content with a "default" keyword appended. Callers must use the returned
// *netrc.Netrc going forward.
func getOrCreateDefault(n *netrc.Netrc) (*netrc.Netrc, *netrc.Machine, error) {
	if m := findDefault(n); m != nil {
		return n, m, nil
	}

	body := n.Render()
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += "default\n"

	fresh, err := netrc.ParseString(body)
	if err != nil {
		return nil, nil, err
	}
	fresh.Path = n.Path
	return fresh, findDefault(fresh), nil
}
