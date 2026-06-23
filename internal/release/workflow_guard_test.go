package release

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// --- Workflow fixtures ------------------------------------------------------

// schemaShapeReleaseWorkflow is the schema repo's release.yml shape: a guard
// job, a nix-vendor-hash job, a contract-gates job behind the guard, and a
// publish job behind all three. No reusable gates, no goreleaser.
const schemaShapeReleaseWorkflow = `
name: Release
on:
  push:
    tags: ["v*"]
jobs:
  guard:
    runs-on: ubuntu-latest
    steps: [{ run: "go run ./cmd/release-guard parse-tag $TAG" }]
  nix-vendor-hash:
    needs: guard
    runs-on: ubuntu-latest
    steps: [{ run: "make nix-vendor-hash" }]
  contract-gates:
    needs: [guard, nix-vendor-hash]
    runs-on: ubuntu-latest
    steps: [{ run: "make gates" }]
  release:
    needs: [guard, nix-vendor-hash, contract-gates]
    runs-on: ubuntu-latest
    steps: [{ run: "gh release create $TAG generated/*" }]
`

// peasantShapeReleaseWorkflow is the peasant repo's release.yml shape: the
// goreleaser publish job sits behind two reusable-workflow gates (e2e and
// release-e2e), each using its reusable workflow, passing secrets: inherit, and
// carrying no `if:`.
const peasantShapeReleaseWorkflow = `
name: Release
on:
  push:
    tags: ["v*"]
jobs:
  guard:
    runs-on: ubuntu-latest
    steps: [{ run: "go run github.com/peasant-labs/schema/cmd/release-guard parse-tag $TAG" }]
  nix-vendor-hash:
    needs: guard
    runs-on: ubuntu-latest
    steps: [{ run: "make nix-vendor-hash" }]
  e2e:
    needs: [guard, nix-vendor-hash]
    uses: ./.github/workflows/e2e.yml
    secrets: inherit
  release-e2e:
    needs: [guard, nix-vendor-hash]
    uses: ./.github/workflows/release-e2e.yml
    secrets: inherit
  release:
    needs: [guard, nix-vendor-hash, e2e, release-e2e]
    runs-on: ubuntu-latest
    steps: [{ run: "goreleaser release" }]
`

// --- Policy fixtures (Go values) --------------------------------------------

// schemaShapePolicy is the schema repo's WorkflowPolicy: guard; nix-vendor-hash
// needs[guard]; contract-gates needs[guard, nix-vendor-hash]; release
// needs[guard, nix-vendor-hash, contract-gates].
func schemaShapePolicy() WorkflowPolicy {
	return WorkflowPolicy{Jobs: []JobRule{
		{Name: "guard"},
		{Name: "nix-vendor-hash", Needs: []string{"guard"}},
		{Name: "contract-gates", Needs: []string{"guard", "nix-vendor-hash"}},
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

// --- Each repo's own policy validates its own release.yml -------------------

func TestCheckReleaseWorkflowWithPolicy_OwnPairsPass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		policy   WorkflowPolicy
		workflow string
	}{
		{"schema policy vs schema release.yml", schemaShapePolicy(), schemaShapeReleaseWorkflow},
		{"peasant policy vs peasant release.yml", peasantShapePolicy(), peasantShapeReleaseWorkflow},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := checkReleaseWorkflowWithPolicy("release.yml", []byte(tc.workflow), tc.policy); err != nil {
				t.Fatalf("own policy/workflow pair rejected: %v", err)
			}
		})
	}
}

// --- Cross-fed wrong policy fails actionably (BDD #5) -----------------------

func TestCheckReleaseWorkflowWithPolicy_CrossFedFailsActionably(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		policy   WorkflowPolicy
		workflow string
		wantSub  string // observable, names the offending job
	}{
		{
			name:     "schema policy rejects peasant release.yml (no contract-gates job)",
			policy:   schemaShapePolicy(),
			workflow: peasantShapeReleaseWorkflow,
			wantSub:  "missing jobs.contract-gates",
		},
		{
			name:     "peasant policy rejects schema release.yml (no e2e reusable gate)",
			policy:   peasantShapePolicy(),
			workflow: schemaShapeReleaseWorkflow,
			wantSub:  "missing jobs.e2e",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := checkReleaseWorkflowWithPolicy("release.yml", []byte(tc.workflow), tc.policy)
			if err == nil {
				t.Fatalf("expected cross-fed pair to be rejected, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not name the offending job (want substring %q)", err.Error(), tc.wantSub)
			}
		})
	}
}

// --- Per-assertion rejections ------------------------------------------------

func TestCheckReleaseWorkflowWithPolicy_Rejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		policy   WorkflowPolicy
		workflow string
		wantSub  string
	}{
		{
			name:     "empty policy",
			policy:   WorkflowPolicy{},
			workflow: schemaShapeReleaseWorkflow,
			wantSub:  "declares no jobs",
		},
		{
			name:     "no jobs mapping",
			policy:   schemaShapePolicy(),
			workflow: "name: Release\non:\n  push:\n    tags: [\"v*\"]\n",
			wantSub:  "has no jobs mapping",
		},
		{
			name:   "missing required needs edge",
			policy: schemaShapePolicy(),
			workflow: `
jobs:
  guard:
    runs-on: ubuntu-latest
  nix-vendor-hash:
    needs: guard
    runs-on: ubuntu-latest
  contract-gates:
    needs: [guard, nix-vendor-hash]
    runs-on: ubuntu-latest
  release:
    needs: [guard, nix-vendor-hash]
    runs-on: ubuntu-latest
`,
			wantSub: "jobs.release.needs is missing contract-gates",
		},
		{
			name:   "reusable gate missing uses",
			policy: peasantShapePolicy(),
			workflow: `
jobs:
  guard:
    runs-on: ubuntu-latest
  nix-vendor-hash:
    needs: guard
    runs-on: ubuntu-latest
  e2e:
    needs: [guard, nix-vendor-hash]
    secrets: inherit
  release-e2e:
    needs: [guard, nix-vendor-hash]
    uses: ./.github/workflows/release-e2e.yml
    secrets: inherit
  release:
    needs: [guard, nix-vendor-hash, e2e, release-e2e]
    runs-on: ubuntu-latest
`,
			wantSub: "jobs.e2e uses <missing>",
		},
		{
			name:   "reusable gate missing secrets inherit",
			policy: peasantShapePolicy(),
			workflow: `
jobs:
  guard:
    runs-on: ubuntu-latest
  nix-vendor-hash:
    needs: guard
    runs-on: ubuntu-latest
  e2e:
    needs: [guard, nix-vendor-hash]
    uses: ./.github/workflows/e2e.yml
  release-e2e:
    needs: [guard, nix-vendor-hash]
    uses: ./.github/workflows/release-e2e.yml
    secrets: inherit
  release:
    needs: [guard, nix-vendor-hash, e2e, release-e2e]
    runs-on: ubuntu-latest
`,
			wantSub: "jobs.e2e secrets is <missing>",
		},
		{
			name:   "reusable gate has forbidden if condition",
			policy: peasantShapePolicy(),
			workflow: `
jobs:
  guard:
    runs-on: ubuntu-latest
  nix-vendor-hash:
    needs: guard
    runs-on: ubuntu-latest
  e2e:
    if: startsWith(github.ref, 'refs/tags/v')
    needs: [guard, nix-vendor-hash]
    uses: ./.github/workflows/e2e.yml
    secrets: inherit
  release-e2e:
    needs: [guard, nix-vendor-hash]
    uses: ./.github/workflows/release-e2e.yml
    secrets: inherit
  release:
    needs: [guard, nix-vendor-hash, e2e, release-e2e]
    runs-on: ubuntu-latest
`,
			wantSub: "jobs.e2e has an if condition",
		},
		{
			name:     "malformed yaml",
			policy:   schemaShapePolicy(),
			workflow: "jobs: [this is: not valid: mapping",
			wantSub:  "cannot parse",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := checkReleaseWorkflowWithPolicy("release.yml", []byte(tc.workflow), tc.policy)
			if err == nil {
				t.Fatalf("expected rejection for %q, got nil", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain expected substring %q", err.Error(), tc.wantSub)
			}
		})
	}
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
	// the os.ReadFile success branch is exercised alongside the error branch.
	path := filepath.Join(t.TempDir(), "release.yml")
	if err := os.WriteFile(path, []byte(schemaShapeReleaseWorkflow), 0o644); err != nil {
		t.Fatalf("write workflow fixture: %v", err)
	}
	if err := CheckReleaseWorkflowFile(path, schemaShapePolicy()); err != nil {
		t.Fatalf("wrapper rejected a valid schema-shape workflow file: %v", err)
	}
}

// --- LoadWorkflowPolicy round-trips a policy file ----------------------------

const peasantShapePolicyYAML = `
jobs:
  - name: guard
  - name: nix-vendor-hash
    needs: [guard]
  - name: e2e
    needs: [guard, nix-vendor-hash]
    reusable:
      uses: ./.github/workflows/e2e.yml
      secretsInherit: true
      forbidIf: true
  - name: release-e2e
    needs: [guard, nix-vendor-hash]
    reusable:
      uses: ./.github/workflows/release-e2e.yml
      secretsInherit: true
      forbidIf: true
  - name: release
    needs: [guard, nix-vendor-hash, e2e, release-e2e]
`

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

	path := writePolicyFile(t, peasantShapePolicyYAML)
	policy, err := LoadWorkflowPolicy(path)
	if err != nil {
		t.Fatalf("LoadWorkflowPolicy(%s) failed: %v", path, err)
	}
	if !reflect.DeepEqual(policy, peasantShapePolicy()) {
		t.Fatalf("loaded policy does not match expected:\n got: %+v\nwant: %+v", policy, peasantShapePolicy())
	}
	// End-to-end: the loaded policy validates the peasant-shape workflow.
	if err := checkReleaseWorkflowWithPolicy("release.yml", []byte(peasantShapeReleaseWorkflow), policy); err != nil {
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
