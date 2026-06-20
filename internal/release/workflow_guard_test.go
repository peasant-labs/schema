package release

import (
	"strings"
	"testing"
)

// validReleaseWorkflow is the minimal release.yml shape the schema repo's guard
// accepts: a guard job, a nix-vendor-hash job, a contract-gates job that needs
// the guard, and a publish job behind all three.
const validReleaseWorkflow = `
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

func TestCheckReleaseWorkflow_Valid(t *testing.T) {
	if err := CheckReleaseWorkflow("release.yml", []byte(validReleaseWorkflow)); err != nil {
		t.Fatalf("valid release workflow rejected: %v", err)
	}
}

func TestCheckReleaseWorkflow_Rejects(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantSub string // substring the actionable error must contain
	}{
		{
			name:    "no jobs mapping",
			yaml:    "name: Release\non:\n  push:\n    tags: [\"v*\"]\n",
			wantSub: "no jobs mapping",
		},
		{
			name: "missing contract-gates job",
			yaml: `
jobs:
  guard:
    runs-on: ubuntu-latest
  nix-vendor-hash:
    needs: guard
    runs-on: ubuntu-latest
  release:
    needs: [guard, nix-vendor-hash]
    runs-on: ubuntu-latest
`,
			wantSub: "missing jobs.contract-gates",
		},
		{
			name: "publish not behind contract-gates",
			yaml: `
jobs:
  guard:
    runs-on: ubuntu-latest
  nix-vendor-hash:
    needs: guard
    runs-on: ubuntu-latest
  contract-gates:
    needs: guard
    runs-on: ubuntu-latest
  release:
    needs: [guard, nix-vendor-hash]
    runs-on: ubuntu-latest
`,
			wantSub: "jobs.release.needs is missing contract-gates",
		},
		{
			name: "contract-gates not behind guard",
			yaml: `
jobs:
  guard:
    runs-on: ubuntu-latest
  nix-vendor-hash:
    needs: guard
    runs-on: ubuntu-latest
  contract-gates:
    runs-on: ubuntu-latest
  release:
    needs: [guard, nix-vendor-hash, contract-gates]
    runs-on: ubuntu-latest
`,
			wantSub: "jobs.contract-gates.needs is missing guard",
		},
		{
			name: "missing nix-vendor-hash job",
			yaml: `
jobs:
  guard:
    runs-on: ubuntu-latest
  contract-gates:
    needs: guard
    runs-on: ubuntu-latest
  release:
    needs: [guard, contract-gates]
    runs-on: ubuntu-latest
`,
			wantSub: "missing jobs.nix-vendor-hash",
		},
		{
			name:    "malformed yaml",
			yaml:    "jobs: [this is: not valid: mapping",
			wantSub: "cannot parse",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckReleaseWorkflow("release.yml", []byte(tc.yaml))
			if err == nil {
				t.Fatalf("expected rejection for %q, got nil", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain expected substring %q", err.Error(), tc.wantSub)
			}
		})
	}
}
