package main

import (
	"errors"
	"fmt"

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
	if _, err := loadConfig("check-config", path); err != nil {
		return err
	}
	return emitLines([]string{"ok=true"})
}

// loadConfig reads one working-tree release config through the contract
// package: a read failure stays a plain error for exit 1, a parse failure is
// the typed config-invalid rejection for exit 2.
func loadConfig(subcommand, path string) (intent.Config, error) {
	cfg, err := intent.LoadConfig(path)
	if err != nil {
		var ie *intent.Error
		if errors.As(err, &ie) {
			return intent.Config{}, rejectAs(subcommand, err)
		}
		return intent.Config{}, configReadError(path, "", err)
	}
	return cfg, nil
}
