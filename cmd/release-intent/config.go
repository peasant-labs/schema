package main

import (
	"fmt"
	"os"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// DefaultConfigPath is the conventional location of the release config.
const DefaultConfigPath = ".github/release-intent.yaml"

// loadConfigBytes parses config bytes from git through a temp file so the
// command layer reads the base revision while the contract package owns the
// grammar. The file lives in t.TempDir-style storage and is removed after
// parsing.
func loadConfigBytes(data []byte) (intent.Config, error) {
	f, err := os.CreateTemp("", "release-intent-*.yaml")
	if err != nil {
		return intent.Config{}, fmt.Errorf("cannot stage release config for parsing: %w", err)
	}
	path := f.Name()
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return intent.Config{}, fmt.Errorf("cannot stage release config for parsing: %w", err)
	}
	f.Close()
	defer os.Remove(path)
	cfg, err := intent.LoadConfig(path)
	if err != nil {
		return intent.Config{}, err
	}
	return cfg, nil
}

// runCheckConfig implements: release-intent check-config --config F.
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
	cfg, err := intent.LoadConfig(path)
	if err != nil {
		return err
	}
	_ = cfg
	return emitLines([]string{"ok=true"})
}
