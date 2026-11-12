package main

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the per-repo release intent: .github/release-intent.yaml.
type Config struct {
	Releases         bool              `yaml:"releases"`
	TagPrefix        string            `yaml:"tag_prefix"`
	AutomationAuthor string            `yaml:"automation_author"`
	Changelog        string            `yaml:"changelog"`
	VersionFiles     []VersionFile     `yaml:"version_files"`
	ProtectedPaths   []string          `yaml:"protected_paths"`
	Surfaces         map[string]string `yaml:"surfaces"`
	BotBranches      map[string]string `yaml:"bot_branches"`
}

// VersionFile names a JSON file and the single top-level field the release PR
// owns. Ordinary PRs may change anything else in the file.
type VersionFile struct {
	Path  string `yaml:"path"`
	Field string `yaml:"field"`
}

// DefaultConfigPath is the conventional location of the release config.
const DefaultConfigPath = ".github/release-intent.yaml"

// LoadConfig reads and validates a release config file. Unknown keys are
// rejected so a typo'd key fails loudly instead of silently disabling a gate.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("load release config: cannot read %s: %w. Create .github/release-intent.yaml or fix the --config path", path, err)
	}
	return ParseConfig(data, path)
}

// ParseConfig decodes and validates config bytes. It is pure so fixtures can
// drive it without files.
func ParseConfig(data []byte, path string) (Config, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("load release config: cannot parse %s as YAML: %w. Expected keys: releases, tag_prefix, automation_author, changelog, version_files, protected_paths, surfaces, bot_branches", path, err)
	}
	if err := ValidateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("load release config: %s %w", path, err)
	}
	return cfg, nil
}

// ValidateConfig applies the structural rules: a non-empty tag prefix and
// changelog, named version-file fields, and no surfaces where releases are off.
func ValidateConfig(cfg Config) error {
	if cfg.TagPrefix == "" {
		return fmt.Errorf("is invalid: tag_prefix is empty (set tag_prefix to \"v\" or \"fairtrade-v\")")
	}
	if cfg.Changelog == "" {
		return fmt.Errorf("is invalid: changelog is empty (set changelog to the changelog path, e.g. CHANGELOG.md)")
	}
	for i, vf := range cfg.VersionFiles {
		if vf.Path == "" || vf.Field == "" {
			return fmt.Errorf("is invalid: version_files[%d] needs both path and field (got path=%q field=%q)", i, vf.Path, vf.Field)
		}
	}
	if !cfg.Releases && len(cfg.Surfaces) > 0 {
		return fmt.Errorf("is invalid: surfaces must be empty where releases is false (village carries no surfaces)")
	}
	return nil
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
	cfg, err := LoadConfig(path)
	if err != nil {
		return err
	}
	_ = cfg
	return emitLines([]string{"ok=true"})
}
