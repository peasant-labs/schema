package contractgates

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// moduleRoot walks up from the test working directory to the schema module root
// (the dir holding go.mod).
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

// skipIfMissing t.Skip()s with an actionable message when a gate binary is not on
// PATH (e.g. a bare `go test` outside `nix develop`).
func skipIfMissing(t *testing.T, bin string) {
	t.Helper()
	if _, err := exec.LookPath(bin); err != nil {
		t.Skipf(
			"%s not on PATH: the %s contract gate is provisioned by the flake dev shell, not go.mod.\n"+
				"  run this test inside the dev shell — `nix develop -c make check` (or `direnv allow` then `go test ./...`).",
			bin, bin)
	}
}

// TestOasdiffSyntheticBreak proves the oasdiff gate fires: it removes an endpoint
// from a committed golden OpenAPI spec and asserts `oasdiff breaking --fail-on
// ERR` exits non-zero and reports the breaking change.
func TestOasdiffSyntheticBreak(t *testing.T) {
	skipIfMissing(t, "oasdiff")

	root := moduleRoot(t)
	goldenPath := filepath.Join(root, "generated", "village-api-0.2.0.json")
	goldenBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden spec %s: %v", goldenPath, err)
	}

	// Parse with stdlib encoding/json (no go.mod pollution) and remove ONE path
	// (endpoint) — removing an endpoint that exists in base but not revision is an
	// ERR-level breaking change (api-path-removed).
	var spec map[string]any
	if err := json.Unmarshal(goldenBytes, &spec); err != nil {
		t.Fatalf("unmarshal golden spec: %v", err)
	}
	paths, ok := spec["paths"].(map[string]any)
	if !ok || len(paths) == 0 {
		t.Fatalf("golden spec %s has no paths object to mutate (got %T)", goldenPath, spec["paths"])
	}
	// Delete a deterministic path (the lexicographically first) so the mutation is
	// stable across runs.
	var removed string
	for p := range paths {
		if removed == "" || p < removed {
			removed = p
		}
	}
	delete(paths, removed)
	mutated, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		t.Fatalf("marshal mutated spec: %v", err)
	}

	dir := t.TempDir()
	basePath := filepath.Join(dir, "base.json")
	revPath := filepath.Join(dir, "revision.json")
	if err := os.WriteFile(basePath, goldenBytes, 0o644); err != nil {
		t.Fatalf("write base spec: %v", err)
	}
	if err := os.WriteFile(revPath, mutated, 0o644); err != nil {
		t.Fatalf("write revision spec: %v", err)
	}

	cmd := exec.Command("oasdiff", "breaking", basePath, revPath, "--fail-on", "ERR")
	out, err := cmd.CombinedOutput()
	t.Logf("oasdiff output (removed endpoint %q):\n%s", removed, out)

	if err == nil {
		t.Fatalf("oasdiff exited 0 after removing endpoint %q — the breaking gate did NOT fire", removed)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("oasdiff failed to run (not an exit error): %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Fatalf("expected oasdiff to exit 1 on an ERR-level breaking change, got exit %d", exitErr.ExitCode())
	}
	// oasdiff reports the break as an `error`-severity change (e.g.
	// "1 error" / "[api-path-removed-without-deprecation]"). Assert the
	// breaking change was actually reported, not merely a non-zero exit.
	low := strings.ToLower(string(out))
	if !strings.Contains(low, "error") || !strings.Contains(low, "removed") {
		t.Fatalf("oasdiff exited non-zero but its output did not report the removed-endpoint breaking change:\n%s", out)
	}
}

// TestOasdiffNoBreakOnIdenticalSpec is the POSITIVE control for the oasdiff gate:
// diffing a committed golden spec against ITSELF must exit 0 and report no
// breaking change. Without this, a gate that errored unconditionally (stuck-on-
// fail) would still pass TestOasdiffSyntheticBreak for the wrong reason; this
// proves the non-zero exit there is attributable to the synthetic break, not a
// permanently-red gate.
func TestOasdiffNoBreakOnIdenticalSpec(t *testing.T) {
	skipIfMissing(t, "oasdiff")

	root := moduleRoot(t)
	goldenPath := filepath.Join(root, "generated", "village-api-0.2.0.json")
	if _, err := os.Stat(goldenPath); err != nil {
		t.Fatalf("golden spec %s missing: %v", goldenPath, err)
	}

	cmd := exec.Command("oasdiff", "breaking", goldenPath, goldenPath, "--fail-on", "ERR")
	out, err := cmd.CombinedOutput()
	t.Logf("oasdiff output (identical spec):\n%s", out)
	if err != nil {
		t.Fatalf("oasdiff exited non-zero diffing an IDENTICAL spec against itself — the gate is stuck-on-fail (a permanently-red gate would falsely satisfy the synthetic-break test): %v\n%s", err, out)
	}
}

// TestGoAPIDiffSyntheticBreak proves the go-apidiff gate fires: it builds a
// throwaway git repo whose head commit REMOVES an exported function, then asserts
// go-apidiff reports an incompatible change.
func TestGoAPIDiffSyntheticBreak(t *testing.T) {
	skipIfMissing(t, "go-apidiff")
	skipIfMissing(t, "git")
	skipIfMissing(t, "go")

	dir := t.TempDir()

	writeFile := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	gitCmd := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		// Deterministic, config-independent identity so the test never depends on
		// the developer's global git config.
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=synthbreak", "GIT_AUTHOR_EMAIL=synthbreak@invalid",
			"GIT_COMMITTER_NAME=synthbreak", "GIT_COMMITTER_EMAIL=synthbreak@invalid",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	revParse := func(ref string) string {
		t.Helper()
		cmd := exec.Command("git", "rev-parse", ref)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git rev-parse %s: %v\n%s", ref, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	writeFile("go.mod", "module example.com/synthbreak\n\ngo 1.25\n")
	// Base API: two exported functions.
	writeFile("api.go", `package synthbreak

// Stable stays put across the break.
func Stable() string { return "stable" }

// Removed is deleted in the head commit — an incompatible exported-API change.
func Removed() string { return "removed" }
`)
	gitCmd("init", "-q")
	gitCmd("add", ".")
	gitCmd("commit", "-q", "-m", "base API")
	baseSHA := revParse("HEAD")

	// Head API: remove the exported function (breaking change).
	writeFile("api.go", `package synthbreak

// Stable stays put across the break.
func Stable() string { return "stable" }
`)
	gitCmd("add", ".")
	gitCmd("commit", "-q", "-m", "remove exported Removed()")
	headSHA := revParse("HEAD")

	cmd := exec.Command("go-apidiff", baseSHA, headSHA, "--repo-path", dir)
	// Run WITH cwd inside the synthetic repo: go-apidiff resolves the module from
	// the working directory, so running it from the schema module root (a
	// different module) yields no analysis.
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, runErr := cmd.CombinedOutput()
	t.Logf("go-apidiff output (exit: %v):\n%s", runErr, out)

	if !strings.Contains(string(out), "Incompatible changes") {
		t.Fatalf("go-apidiff did NOT report an incompatible change after removing an exported function — the breaking gate did not fire:\n%s", out)
	}
	if !strings.Contains(string(out), "Removed") {
		t.Fatalf("go-apidiff reported incompatible changes but did not mention the removed symbol:\n%s", out)
	}
}

// TestGoAPIDiffStampExemption is the load-bearing proof that the go-apidiff gate's
// spec-version-stamp exemption does NOT punch a hole. go-apidiff sees a bump of a
// version-marker const (VillageAPIVersion / PeasantLocalAPIVersion / TypesVersion) as
// an "incompatible" value change; the gate exempts exactly that so an intentional
// version bump does not red the contract gate. This test proves:
//
//	(a) a stamp-ONLY value change PASSES the gate; but
//	(b) a stamp change ACCOMPANIED by a real incompatible change (a removed exported
//	    symbol) still FAILS.
//
// It drives the SHIPPED gate decision — scripts/contract-gates.sh evaluate-apidiff —
// with REAL go-apidiff output produced from a throwaway repo (where go-apidiff, unlike
// a linked worktree, analyses cleanly), so it exercises the real filter, not a
// reimplementation of it.
func TestGoAPIDiffStampExemption(t *testing.T) {
	skipIfMissing(t, "go-apidiff")
	skipIfMissing(t, "git")
	skipIfMissing(t, "go")
	skipIfMissing(t, "bash")

	root := moduleRoot(t)
	gateScript := filepath.Join(root, "scripts", "contract-gates.sh")
	if _, err := os.Stat(gateScript); err != nil {
		t.Fatalf("gate script %s missing: %v", gateScript, err)
	}

	// apidiffOutput builds a throwaway git repo (base API, then head API), runs
	// go-apidiff across the two commits, and returns its raw output.
	apidiffOutput := func(t *testing.T, baseAPI, headAPI string) string {
		t.Helper()
		dir := t.TempDir()
		writeFile := func(name, content string) {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
		git := func(args ...string) string {
			cmd := exec.Command("git", args...)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(),
				"GIT_AUTHOR_NAME=stampexempt", "GIT_AUTHOR_EMAIL=stampexempt@invalid",
				"GIT_COMMITTER_NAME=stampexempt", "GIT_COMMITTER_EMAIL=stampexempt@invalid",
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
			}
			return strings.TrimSpace(string(out))
		}
		writeFile("go.mod", "module example.com/stampexempt\n\ngo 1.25\n")
		writeFile("api.go", baseAPI)
		git("init", "-q")
		git("add", ".")
		git("commit", "-q", "-m", "base")
		base := git("rev-parse", "HEAD")
		writeFile("api.go", headAPI)
		git("add", ".")
		git("commit", "-q", "-m", "head")
		head := git("rev-parse", "HEAD")

		cmd := exec.Command("go-apidiff", base, head, "--repo-path", dir)
		// go-apidiff resolves the module from the working directory, so run it inside
		// the throwaway repo (see TestGoAPIDiffSyntheticBreak).
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
		out, runErr := cmd.CombinedOutput()
		t.Logf("go-apidiff output (exit %v):\n%s", runErr, out)
		return string(out)
	}

	// evaluate pipes go-apidiff output through the REAL gate decision and returns its
	// exit code (0 = gate PASS, 1 = gate FAIL).
	evaluate := func(t *testing.T, apidiff string) int {
		t.Helper()
		cmd := exec.Command("bash", gateScript, "evaluate-apidiff")
		cmd.Dir = root
		cmd.Stdin = strings.NewReader(apidiff)
		out, err := cmd.CombinedOutput()
		t.Logf("evaluate-apidiff verdict:\n%s", out)
		if err == nil {
			return 0
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		t.Fatalf("evaluate-apidiff failed to run: %v\n%s", err, out)
		return -1
	}

	// (a) stamp-ONLY value change -> gate PASSES.
	aOut := apidiffOutput(t,
		`package stampexempt

const VillageAPIVersion = "0.3.0"

func Stable() string { return "s" }
`,
		`package stampexempt

const VillageAPIVersion = "0.4.0"

func Stable() string { return "s" }
`,
	)
	if !strings.Contains(aOut, "Incompatible changes") || !strings.Contains(aOut, "VillageAPIVersion") {
		t.Fatalf("precondition (a): go-apidiff did not flag the VillageAPIVersion bump as incompatible:\n%s", aOut)
	}
	if got := evaluate(t, aOut); got != 0 {
		t.Fatalf("(a) stamp-only bump: gate exit = %d, want 0 (PASS) — the stamp exemption did not apply", got)
	}

	// (b) stamp value change + a REAL break (removed exported func) -> gate FAILS.
	bOut := apidiffOutput(t,
		`package stampexempt

const VillageAPIVersion = "0.3.0"

func Stable() string { return "s" }
func Removed() string { return "r" }
`,
		`package stampexempt

const VillageAPIVersion = "0.4.0"

func Stable() string { return "s" }
`,
	)
	if !strings.Contains(bOut, "Removed") {
		t.Fatalf("precondition (b): go-apidiff did not flag the removed exported func:\n%s", bOut)
	}
	if got := evaluate(t, bOut); got != 1 {
		t.Fatalf("(b) stamp bump + removed exported func: gate exit = %d, want 1 (FAIL) — the exemption masked a REAL break", got)
	}
}
