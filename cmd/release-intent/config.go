package main

import (
	"github.com/peasant-labs/schema/internal/release/intent"
)

// DefaultConfigPath is the conventional location of the release config.
const DefaultConfigPath = ".github/release-intent.yaml"

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
