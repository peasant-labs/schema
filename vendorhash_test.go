package schema_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// localReplacePattern matches a go.mod `replace` directive whose target is a
// LOCAL filesystem path (starts with ./ or ../). These are the only replaces
// that fold first-party source into the module graph.
var localReplacePattern = regexp.MustCompile(`=>\s*\.\.?/`)

// TestVendorHashStableOnFirstPartyEdit is the #119 vendorHash-stability proof,
// owned by THIS repo (PROPOSAL-4 review refinement #2: the proof lives in the
// schema repo; peasant's #119 fix is the replace-removal in SLICE-C).
//
// The Nix `buildGoModule` vendorHash is a fixed-output hash over the THIRD-PARTY
// module graph only (go.mod + go.sum). A first-party edit — e.g. changing a
// schema testdata YAML or a .go source file — changes neither, so it can NEVER
// drift the vendorHash. The ONLY way a first-party edit leaks into the vendor
// computation is a local-path `replace` directive (`=> ./...` / `=> ../...`):
// that was exactly the #119 pathology in peasant (its root go.mod had
// `replace …/pkg/schema => ./pkg/schema`, so every schema edit re-hashed the
// vendor). This contract-only leaf has no such replace by construction.
//
// This test pins that invariant structurally and hermetically (no Nix needed):
// it fails if go.mod ever grows a local-path replace, which is the single thing
// that would make a first-party edit drift the vendorHash. It is the durable
// guard behind the flake comment on `vendorHash`.
func TestVendorHashStableOnFirstPartyEdit(t *testing.T) {
	root := moduleRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	for i, line := range strings.Split(string(data), "\n") {
		// Ignore comments.
		code := line
		if idx := strings.Index(code, "//"); idx >= 0 {
			code = code[:idx]
		}
		if !strings.Contains(code, "replace") && !strings.Contains(code, "=>") {
			continue
		}
		if localReplacePattern.MatchString(code) {
			t.Errorf(
				"go.mod:%d has a LOCAL-path replace directive: %q\n"+
					"  what: a `replace => ./...`/`../...` folds first-party source into the Nix vendor graph.\n"+
					"  why:  that is the #119 pathology — every first-party edit would then drift the flake vendorHash.\n"+
					"  fix:  remove the local replace; this leaf module must depend on published versions only.",
				i+1, strings.TrimSpace(line))
		}
	}
}

// moduleRoot walks up from the test's working directory to the directory holding
// go.mod (the schema module root).
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate go.mod walking up from the test working directory")
		}
		dir = parent
	}
}
