package schema_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// allowedDirectRequires is the EXACT set of third-party modules this contract-only
// leaf is permitted to import directly (PROPOSAL-4 W6 / IP4). The contract gates
// (oasdiff / go-apidiff / vacuum) and every other dev/CI tool are provisioned via
// the flake dev shell, NEVER as a go.mod require — Go has no dev/prod dep split,
// so the only mechanical enforcement is this audit over the direct require set.
//
// Keep this in lockstep with go.mod's direct requires; adding one here is a
// deliberate decision that this is part of the contract's runtime/library
// surface, not a tool that leaked in.
var allowedDirectRequires = map[string]bool{
	"github.com/dayvidpham/bestiary":           true,
	"github.com/santhosh-tekuri/jsonschema/v5": true,
	"github.com/swaggest/jsonschema-go":        true,
	"github.com/swaggest/openapi-go":           true,
	"golang.org/x/crypto":                      true,
	"gopkg.in/yaml.v3":                         true,
}

// parseDirectRequires extracts the DIRECT (non-`// indirect`) require paths from
// go.mod bytes. It is a pure function (testable against fixtures) and parses
// go.mod by hand ON PURPOSE: importing golang.org/x/mod/modfile would itself add
// a go.mod require and violate the very invariant this test guards.
//
// It handles both forms:
//
//	require golang.org/x/crypto v0.48.0
//	require ( ... )   // block, one "path version" per line
//
// Lines carrying a trailing `// indirect` comment are transitive deps (allowed)
// and are skipped.
func parseDirectRequires(data []byte) []string {
	var direct []string
	inBlock := false
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// Indirect deps are transitive — not part of the audited direct set.
		isIndirect := strings.Contains(line, "// indirect")
		// Strip any trailing comment for path extraction.
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "require ("):
			inBlock = true
			continue
		case inBlock && line == ")":
			inBlock = false
			continue
		case strings.HasPrefix(line, "require "):
			// Single-line require: "require path version".
			fields := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(fields) >= 1 && !isIndirect {
				direct = append(direct, fields[0])
			}
		case inBlock:
			// Block line: "path version".
			fields := strings.Fields(line)
			if len(fields) >= 1 && !isIndirect {
				direct = append(direct, fields[0])
			}
		}
	}
	sort.Strings(direct)
	return direct
}

// auditDirectRequires returns the require paths that are NOT in the allowed set.
func auditDirectRequires(direct []string, allowed map[string]bool) []string {
	var violations []string
	for _, path := range direct {
		if !allowed[path] {
			violations = append(violations, path)
		}
	}
	return violations
}

// TestLeafAudit_GoModRequiresAreAllowed asserts the real go.mod's direct require
// set is a subset of allowedDirectRequires — the leaf-audit gate (W6/IP4). It
// fails the moment a dev/CI tool (oasdiff, go-apidiff, vacuum, …) leaks into
// go.mod instead of the flake dev shell.
func TestLeafAudit_GoModRequiresAreAllowed(t *testing.T) {
	root := moduleRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	direct := parseDirectRequires(data)
	if len(direct) == 0 {
		t.Fatal("parsed zero direct requires from go.mod — the parser is broken (go.mod has at least the bestiary + swaggest deps)")
	}

	violations := auditDirectRequires(direct, allowedDirectRequires)
	if len(violations) > 0 {
		t.Errorf(
			"go.mod has direct require(s) outside the allowed leaf set: %v\n"+
				"  what: this contract-only leaf may import directly only %v.\n"+
				"  why:  dev/CI tools (oasdiff, go-apidiff, vacuum, …) belong in the flake dev shell, NEVER go.mod — Go has no dev/prod dep split.\n"+
				"  fix:  remove the require (run `go mod tidy` after deleting any tools.go / `go get -tool` use), or, if this is genuinely part of the contract library surface, add it to allowedDirectRequires with justification.",
			violations, sortedKeys(allowedDirectRequires))
	}
}

// TestLeafAudit_NegativeFixtureFails proves the audit FIRES: a polluted go.mod
// whose direct requires include a contract-gate tool (oasdiff) must be rejected.
func TestLeafAudit_NegativeFixtureFails(t *testing.T) {
	polluted := []byte(`module github.com/peasant-labs/schema

go 1.25.5

require (
	github.com/dayvidpham/bestiary v0.1.1
	github.com/oasdiff/oasdiff v1.19.1
	github.com/swaggest/openapi-go v0.2.60
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
)
`)

	direct := parseDirectRequires(polluted)
	violations := auditDirectRequires(direct, allowedDirectRequires)
	if len(violations) == 0 {
		t.Fatal("leaf audit did NOT fire on a go.mod that directly requires github.com/oasdiff/oasdiff — the gate is broken")
	}
	if !containsString(violations, "github.com/oasdiff/oasdiff") {
		t.Fatalf("expected the oasdiff tool to be flagged, got violations: %v", violations)
	}
	// The indirect dep must NOT be flagged (transitive deps are allowed).
	if containsString(violations, "github.com/google/uuid") {
		t.Fatalf("indirect dep github.com/google/uuid was wrongly flagged as a direct violation: %v", violations)
	}
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
