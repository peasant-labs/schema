package assert

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// RequireManifestNames checks that a strict YAML manifest and a fixture contain
// exactly the same non-empty, unique case names, independent of their order.
// A missing or extra case must be reconciled with the manifest deliberately.
func RequireManifestNames(t *testing.T, label string, manifestYAML []byte, names []string) {
	t.Helper()
	if err := checkManifestNames(manifestYAML, names); err != nil {
		t.Fatalf("%s: %v", label, err)
	}
}

func checkManifestNames(manifestYAML []byte, names []string) error {
	var manifest struct {
		RequiredNames []string `yaml:"requiredNames"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(manifestYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err != nil {
			return fmt.Errorf("decode trailing manifest document: %w", err)
		}
		return fmt.Errorf("manifest must contain exactly one YAML document")
	}
	if len(manifest.RequiredNames) == 0 {
		return fmt.Errorf("manifest names no cases; add the required case names")
	}

	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if name == "" || strings.TrimSpace(name) != name {
			return fmt.Errorf("invalid fixture case name %q; use a non-empty name without surrounding whitespace", name)
		}
		if seen[name] {
			return fmt.Errorf("fixture repeats case name %q", name)
		}
		seen[name] = true
	}
	required := make(map[string]bool, len(manifest.RequiredNames))
	for _, name := range manifest.RequiredNames {
		if name == "" || strings.TrimSpace(name) != name {
			return fmt.Errorf("invalid manifest name %q; use a non-empty name without surrounding whitespace", name)
		}
		if required[name] {
			return fmt.Errorf("manifest repeats name %q", name)
		}
		required[name] = true
	}
	for _, name := range manifest.RequiredNames {
		if !seen[name] {
			return fmt.Errorf("fixture is missing required case %q", name)
		}
	}
	for _, name := range names {
		if !required[name] {
			return fmt.Errorf("case %q is not in the manifest; add it there", name)
		}
	}
	return nil
}
