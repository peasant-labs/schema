package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// DefaultConfigPath is the conventional location of the release config.
const DefaultConfigPath = ".github/release-intent.yaml"

// configReadError maps a config blob read failure to an exit-1 internal
// error: the file could not be read at all, so nothing was judged.
func configReadError(path, rev string, err error) error {
	if rev == "" {
		return fmt.Errorf("cannot read release config %q: %v. Create .github/release-intent.yaml or fix the path", path, err)
	}
	return fmt.Errorf("cannot read release config %q at %q: %v. Confirm the revision carries the config file", path, rev, err)
}

// configInvalid maps any config content failure to exit 2 with the closed
// config-invalid code. A typed contract error passes through unchanged; a
// plain decode failure is wrapped with the same code so every invalid config
// exits 2 with a typed line, while an unreadable file stays exit 1.
func configInvalid(subcommand, where string, err error) error {
	var ie *intent.Error
	if errors.As(err, &ie) {
		return rejectAs(subcommand, err)
	}
	return rejectAs(subcommand, &intent.Error{
		Code:  intent.ErrConfigInvalid,
		What:  "release config is invalid: " + firstLine(err.Error()),
		Why:   "the release config is strict YAML with a fixed key set and required fields",
		Where: where + ", checked by release-intent " + subcommand,
		When:  "loading the release config before any check",
		Means: "no check can run without the repo's contract",
		Fix:   "fix the file or pass a readable path",
	})
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}

// runCheckConfig implements: release-intent check-config --config F. An
// unreadable file exits 1; an invalid file exits 2 with error=config-invalid.
func runCheckConfig(args []string) error {
	path := DefaultConfigPath
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--config":
			val, next, err := requireFlagValue(args, i, "--config")
			if err != nil {
				return err
			}
			path = val
			i = next
		default:
			return usageErr("unknown flag %q for check-config: expected --config F (got %q; usage: release-intent check-config [--config F])", args[i], args[i])
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return configReadError(path, "", err)
	}
	if _, err := intent.ParseConfig(data); err != nil {
		return configInvalid("check-config", "release config "+strconvQuote(path), err)
	}
	return emitLines([]string{"ok=true"})
}
