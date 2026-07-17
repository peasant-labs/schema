package release

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// The workflow-shape YAML fixtures, the peasant policy file, and the
// check-workflow case table now live under testdata/workflow (loaded via the
// embedded FS in workflow_fixtures_test.go). The policy Go VALUES stay here:
// they are the canonical expected shapes (also asserted by the LoadWorkflowPolicy
// round-trip) and are referenced by name from the case fixture.

// --- Policy fixtures (Go values) --------------------------------------------

// schemaShapePolicy is the schema repo's WorkflowPolicy: guard; nix-vendor-hash
// needs[guard]; contract-gates needs[guard, nix-vendor-hash]; npm-publish
// needs[guard, nix-vendor-hash, contract-gates], requires its own
// permissions.id-token: write (OIDC trusted publishing), and must run under the
// "npm-publish" GitHub Actions environment; release needs[guard,
// nix-vendor-hash, contract-gates]. npm-publish and release are independent
// siblings behind the same three gates.
func schemaShapePolicy() WorkflowPolicy {
	return WorkflowPolicy{Jobs: []JobRule{
		{Name: "guard"},
		{Name: "nix-vendor-hash", Needs: []string{"guard"}},
		{Name: "contract-gates", Needs: []string{"guard", "nix-vendor-hash"}},
		{Name: "npm-publish", Needs: []string{"guard", "nix-vendor-hash", "contract-gates"}, Permissions: &PermissionsRule{IDToken: true}, Environment: "npm-publish"},
		{Name: "release", Needs: []string{"guard", "nix-vendor-hash", "contract-gates"}},
	}}
}

// peasantShapePolicy is the peasant repo's WorkflowPolicy: the reusable e2e and
// release-e2e gates (uses + secrets:inherit + forbid-if + needs[guard,
// nix-vendor-hash]); release needs[guard, nix-vendor-hash, e2e, release-e2e].
func peasantShapePolicy() WorkflowPolicy {
	return WorkflowPolicy{Jobs: []JobRule{
		{Name: "guard"},
		{Name: "nix-vendor-hash", Needs: []string{"guard"}},
		{Name: "e2e", Needs: []string{"guard", "nix-vendor-hash"},
			Reusable: &ReusableRule{Uses: "./.github/workflows/e2e.yml", SecretsInherit: true, ForbidIf: true}},
		{Name: "release-e2e", Needs: []string{"guard", "nix-vendor-hash"},
			Reusable: &ReusableRule{Uses: "./.github/workflows/release-e2e.yml", SecretsInherit: true, ForbidIf: true}},
		{Name: "release", Needs: []string{"guard", "nix-vendor-hash", "e2e", "release-e2e"}},
	}}
}

// --- Fixture-driven policy/workflow validation cases ------------------------
//
// One table over testdata/workflow/cases.yaml: the accept pairs must validate,
// and every reject pair must fail with a message containing all its substrings.
// This subsumes the former OwnPairsPass / CrossFedFailsActionably / Rejects /
// RejectsWrongUsesTarget tables — same cases, same assertions, data moved out.

func TestCheckReleaseWorkflow_Cases(t *testing.T) {
	t.Parallel()

	cases := loadCheckWorkflowCases(t)
	if len(cases.Accept) != 2 || len(cases.Reject) != 18 {
		t.Fatalf("check-workflow fixture has accept=%d reject=%d, want accept=2 reject=18 (fixture truncated?)",
			len(cases.Accept), len(cases.Reject))
	}

	t.Run("accept", func(t *testing.T) {
		t.Parallel()
		for _, tc := range cases.Accept {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				t.Parallel()
				data := readWorkflowFixture(t, tc.Workflow)
				if err := checkReleaseWorkflow(tc.Workflow, data, policyByName(t, tc.Policy)); err != nil {
					t.Fatalf("accept case %q rejected: %v", tc.Name, err)
				}
			})
		}
	})

	t.Run("reject", func(t *testing.T) {
		t.Parallel()
		for _, tc := range cases.Reject {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				t.Parallel()
				if len(tc.WantContains) == 0 {
					t.Fatalf("reject case %q declares no wantContains substrings (fixture error)", tc.Name)
				}
				data := readWorkflowFixture(t, tc.Workflow)
				err := checkReleaseWorkflow(tc.Workflow, data, policyByName(t, tc.Policy))
				if err == nil {
					t.Fatalf("expected rejection for %q, got nil", tc.Name)
				}
				for _, want := range tc.WantContains {
					if !strings.Contains(err.Error(), want) {
						t.Fatalf("error %q does not contain expected substring %q", err.Error(), want)
					}
				}
			})
		}
	})
}

// --- Public file wrapper: read-error path -----------------------------------

func TestCheckReleaseWorkflowFile_ReadError(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "does-not-exist.yml")
	err := CheckReleaseWorkflowFile(missing, schemaShapePolicy())
	if err == nil {
		t.Fatalf("expected a read error for a nonexistent workflow path, got nil")
	}
	// Actionable: names the file, says it could not be read, and how to fix it.
	for _, want := range []string{"cannot read", missing, "Fix the path or run from the repository root"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("read error %q does not contain expected substring %q", err.Error(), want)
		}
	}
}

func TestCheckReleaseWorkflowFile_ReadsAndValidates(t *testing.T) {
	t.Parallel()

	// The wrapper reads a real file and validates it end-to-end (happy path), so
	// the os.ReadFile success branch is exercised alongside the error branch. The
	// bytes come from the embedded schema-shape workflow fixture.
	path := filepath.Join(t.TempDir(), "release.yml")
	if err := os.WriteFile(path, readWorkflowFixture(t, "schema-release.yml"), 0o644); err != nil {
		t.Fatalf("write workflow fixture: %v", err)
	}
	if err := CheckReleaseWorkflowFile(path, schemaShapePolicy()); err != nil {
		t.Fatalf("wrapper rejected a valid schema-shape workflow file: %v", err)
	}
}

// --- LoadWorkflowPolicy round-trips a policy file ----------------------------

func writePolicyFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "release-guard.policy.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write policy fixture: %v", err)
	}
	return path
}

func TestLoadWorkflowPolicy_RoundTripsAndValidates(t *testing.T) {
	t.Parallel()

	// The peasant policy YAML now lives in testdata/workflow/peasant.policy.yml;
	// LoadWorkflowPolicy must reconstruct the peasantShapePolicy() Go value from
	// it, and that loaded policy must validate the peasant-shape workflow.
	path := writePolicyFile(t, string(readWorkflowFixture(t, "peasant.policy.yml")))
	policy, err := LoadWorkflowPolicy(path)
	if err != nil {
		t.Fatalf("LoadWorkflowPolicy(%s) failed: %v", path, err)
	}
	if !reflect.DeepEqual(policy, peasantShapePolicy()) {
		t.Fatalf("loaded policy does not match expected:\n got: %+v\nwant: %+v", policy, peasantShapePolicy())
	}
	// End-to-end: the loaded policy validates the peasant-shape workflow.
	if err := checkReleaseWorkflow("release.yml", readWorkflowFixture(t, "peasant-release.yml"), policy); err != nil {
		t.Fatalf("loaded policy rejected its own workflow shape: %v", err)
	}
}

func TestLoadWorkflowPolicy_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string // "" sentinel means: do not write a file (missing path)
		write   bool
		wantSub string
	}{
		{
			name:    "missing file",
			write:   false,
			wantSub: "cannot read",
		},
		{
			name:    "empty policy declares no jobs",
			content: "",
			write:   true,
			wantSub: "declares no jobs",
		},
		{
			name:    "unknown field rejected",
			content: "jobs:\n  - name: guard\n    unexpected: true\n",
			write:   true,
			wantSub: "cannot parse",
		},
		{
			name:    "empty job name",
			content: "jobs:\n  - needs: [guard]\n",
			write:   true,
			wantSub: "has an empty name",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "release-guard.policy.yml")
			if tc.write {
				if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
					t.Fatalf("write policy fixture: %v", err)
				}
			}
			_, err := LoadWorkflowPolicy(path)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain expected substring %q", err.Error(), tc.wantSub)
			}
		})
	}
}
