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

// runEvaluateAPIDiff pipes go-apidiff output through the SHIPPED incompatible-decision
// seam (scripts/contract-gates.sh evaluate-apidiff) and returns (exit, stdout). Under
// the pre-1.0 advisory policy the exit is two-state — 0 = satisfied (clean OR warned),
// 1 = fail-closed — and the WARN vs CLEAN distinction rides the STDOUT payload
// (non-empty non-exempt incompatible bullets = warn; empty = clean). STDOUT is captured
// SEPARATELY from stderr so the returned string is the pure payload channel (the human
// verdict / ::warning:: text lands on stderr). It needs only bash + the committed gate
// script — NOT the go-apidiff binary — so canned-string cases can exercise the filter
// even where go-apidiff is absent.
func runEvaluateAPIDiff(t *testing.T, root, apidiff string) (int, string) {
	t.Helper()
	gateScript := filepath.Join(root, "scripts", "contract-gates.sh")
	cmd := exec.Command("bash", gateScript, "evaluate-apidiff")
	cmd.Dir = root
	cmd.Stdin = strings.NewReader(apidiff)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	t.Logf("evaluate-apidiff stdout(payload):\n%s\nstderr(verdict):\n%s", stdout.String(), stderr.String())
	if err == nil {
		return 0, stdout.String()
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), stdout.String()
	}
	t.Fatalf("evaluate-apidiff failed to run: %v\nstderr:\n%s", err, stderr.String())
	return -1, ""
}

// runExtractCompatible pipes go-apidiff output through the SHIPPED compatible-extraction
// seam (scripts/contract-gates.sh extract-compatible) and returns its STDOUT — the
// compatible (non-breaking/additive) bullet lines (empty if none). This seam is purely
// informational: it must ALWAYS exit 0 and never affects the gate decision, so a
// non-zero exit is a hard failure here. Like runEvaluateAPIDiff it needs only bash + the
// committed script (no go-apidiff binary), driving the shipped seam rather than a
// reimplementation.
func runExtractCompatible(t *testing.T, root, apidiff string) string {
	t.Helper()
	gateScript := filepath.Join(root, "scripts", "contract-gates.sh")
	cmd := exec.Command("bash", gateScript, "extract-compatible")
	cmd.Dir = root
	cmd.Stdin = strings.NewReader(apidiff)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("extract-compatible must ALWAYS exit 0 (informational seam), got: %v\nstderr:\n%s", err, stderr.String())
	}
	t.Logf("extract-compatible stdout:\n%s", stdout.String())
	return stdout.String()
}

// TestGoAPIDiffStampExemption is the REAL-go-apidiff FORMAT ANCHOR for the stamp
// exemption: it proves the two cases through actual go-apidiff v0.8.3 output (from a
// throwaway repo, where go-apidiff — unlike a linked worktree — analyses cleanly),
// confirming the canned strings in TestGoAPIDiffStampExemptionFilter mirror the tool's
// current format:
//
//	(a) a stamp-ONLY value change PASSES the gate; and
//	(b) a stamp change ACCOMPANIED by a real incompatible change (a removed exported
//	    symbol) still FAILS.
//
// The EXHAUSTIVE tightness/fail-closed cases (non-stamp value change, stamp removal,
// header-without-bullets) live in TestGoAPIDiffStampExemptionFilter as canned-string
// cases through the SAME seam — kept separate so they run without the go-apidiff
// binary. Both drive the SHIPPED gate decision (contract-gates.sh evaluate-apidiff),
// not a reimplementation of it.
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

		// --print-compatible mirrors the SHIPPED gate invocation (gate_go_apidiff), so
		// this real-format anchor exercises the exact output the runner parses. The two
		// cases below have no additive change, so the compatible section stays empty and
		// the incompatible verdict is unaffected.
		cmd := exec.Command("go-apidiff", base, head, "--print-compatible", "--repo-path", dir)
		// go-apidiff resolves the module from the working directory, so run it inside
		// the throwaway repo (see TestGoAPIDiffSyntheticBreak).
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
		out, runErr := cmd.CombinedOutput()
		t.Logf("go-apidiff output (exit %v):\n%s", runErr, out)
		return string(out)
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
	// (a) exempt stamp-only bump -> exit 0 AND an EMPTY incompatible payload (the stamp
	// bullet is filtered out, so nothing is reported to warn on).
	if exit, payload := runEvaluateAPIDiff(t, root, aOut); exit != 0 || strings.TrimSpace(payload) != "" {
		t.Fatalf("(a) stamp-only bump: gate exit = %d, payload = %q; want exit 0 + EMPTY payload — the stamp exemption did not apply", exit, payload)
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
	// (b) stamp bump + a REAL break -> WARN: exit 0, and the payload names the break
	// (Removed) but NOT the exempt stamp value change ("value changed"). This is the
	// no-leak guarantee at the real-format anchor — the exemption filters the stamp
	// bullet while the real break still surfaces (it does not mask a real break).
	if exit, payload := runEvaluateAPIDiff(t, root, bOut); exit != 0 ||
		!strings.Contains(payload, "Removed") || strings.Contains(payload, "value changed") {
		t.Fatalf("(b) stamp bump + removed exported func: gate exit = %d, payload = %q; want exit 0 + payload containing \"Removed\" and NOT \"value changed\"", exit, payload)
	}
}

// TestGoAPIDiffStampExemptionFilter exhaustively pins the tightness of the stamp
// exemption through the SAME shipped seam (contract-gates.sh evaluate-apidiff) using
// CANNED go-apidiff v0.8.3-format strings — so these cases run without the go-apidiff
// binary (only bash is required) and are the durable regression guard behind the
// script's "cannot mask a real break" claim. TestGoAPIDiffStampExemption validates
// that these strings mirror go-apidiff's REAL current format; if a future go-apidiff
// bump changes it, that test surfaces the drift.
func TestGoAPIDiffStampExemptionFilter(t *testing.T) {
	skipIfMissing(t, "bash")
	root := moduleRoot(t)

	// The COMMON additive case: no "Incompatible changes" header at all (go-apidiff run
	// WITH --print-compatible emits only a Compatible section for a purely-additive PR).
	// A valid positive control at the incompatible seam — exit 0, EMPTY payload — proving
	// the seam is not stuck-on-warn.
	cleanNoSection := `example.com/cap
  Compatible changes:
  - Added: added
`

	// Only a stamp const's VALUE CHANGE is exempt — this is the CLEAN control at the
	// seam (mirrors TestGoAPIDiffStampExemption case (a) without needing a repo): the
	// stamp bullet is filtered out, so the payload is empty (nothing to warn on).
	stampValueChange := `example.com/cap
  Incompatible changes:
  - VillageAPIVersion: value changed from "0.3.0" to "0.4.0"
`

	// A NON-stamp const value change: identical quoted value-change FORM, only the
	// name differs. This is the case (b) does NOT cover — (b)'s break is a func
	// removal (a different bullet form), so a regex over-broadened to
	// `.*: value changed from` would still exempt (a) and (this row); ONLY this row
	// catches that over-broadening (it must WARN, i.e. land on the payload). (A string
	// const is used deliberately: an int const renders unquoted — `value changed from 3
	// to 5` — and would not exercise the quoted-form tightness.)
	nonStampValueChange := `example.com/cap
  Incompatible changes:
  - OtherVersion: value changed from "0.3.0" to "0.4.0"
`

	// A stamp REMOVED (not a value change) is a real break — proves the exemption is
	// tight to the value-change FORM, not just the name (it must WARN).
	stampRemoved := `example.com/cap
  Incompatible changes:
  - VillageAPIVersion: removed
`

	// A stamp value bump ACCOMPANYING a real break: the exempt stamp bullet is dropped
	// but the real break still surfaces on the payload — the payload has the break
	// (Removed) and NOT the exempt stamp bullet ("value changed").
	mixedStampAndBreak := `example.com/cap
  Incompatible changes:
  - VillageAPIVersion: value changed from "0.3.0" to "0.4.0"
  - Removed: removed
`

	// A real break ACCOMPANIED by a compatible (additive) change: the incompatible
	// collector closes its block on the "Compatible changes:" heading, so the compatible
	// bullet must NOT leak into the incompatible payload (payload has Removed, NOT Added).
	mixedBreakAndCompatible := `example.com/cap
  Incompatible changes:
  - Removed: removed
  Compatible changes:
  - Added: added
`

	// An "Incompatible changes" header with no parseable bullets. go-apidiff v0.8.3
	// never emits this (it always renders ≥1 bullet); it is a SYNTHETIC stand-in for a
	// future output-format drift, which is exactly what the fail-closed branch guards
	// against — silently passing a real break it could no longer parse.
	headerNoBullets := `example.com/cap
  Incompatible changes:
`

	for _, tc := range []struct {
		name     string
		apidiff  string
		wantExit int // 0 = satisfied (warn/clean), 1 = fail-closed
		// wantContains: non-empty => the STDOUT incompatible payload must contain it
		// (WARN); empty => the payload must be blank (clean/exempt/fail-closed).
		wantContains string
		// wantNotContains: non-empty => the payload must NOT contain it (no-leak).
		wantNotContains string
	}{
		{"additive-only, no incompatible section (PASS)", cleanNoSection, 0, "", ""},
		{"stamp value-change is exempt (PASS)", stampValueChange, 0, "", ""},
		{"non-stamp value-change is NOT exempt (WARN)", nonStampValueChange, 0, "OtherVersion", ""},
		{"stamp removal is NOT exempt (WARN)", stampRemoved, 0, "VillageAPIVersion", ""},
		{"mixed stamp bump + real break (WARN)", mixedStampAndBreak, 0, "Removed", "value changed"},
		{"mixed break + compatible: compatible does NOT leak (WARN)", mixedBreakAndCompatible, 0, "Removed", "Added"},
		{"fail-closed on header without bullets (FAIL)", headerNoBullets, 1, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exit, payload := runEvaluateAPIDiff(t, root, tc.apidiff)
			if exit != tc.wantExit {
				t.Fatalf("%s: evaluate-apidiff exit = %d, want %d\npayload:\n%s", tc.name, exit, tc.wantExit, payload)
			}
			if tc.wantContains == "" {
				if strings.TrimSpace(payload) != "" {
					t.Fatalf("%s: expected an EMPTY incompatible payload on stdout, got:\n%s", tc.name, payload)
				}
			} else if !strings.Contains(payload, tc.wantContains) {
				t.Fatalf("%s: incompatible payload missing %q:\n%s", tc.name, tc.wantContains, payload)
			}
			if tc.wantNotContains != "" && strings.Contains(payload, tc.wantNotContains) {
				t.Fatalf("%s: incompatible payload unexpectedly contains %q (leak):\n%s", tc.name, tc.wantNotContains, payload)
			}
		})
	}
}

// TestGoAPIDiffCompatibleExtraction pins the compatible-extraction seam
// (contract-gates.sh extract-compatible) and cross-checks it against the incompatible
// seam on the SAME input, proving the two channels are DISJOINT: a compatible bullet
// lands only on the compatible channel and never in the incompatible payload, and vice
// versa. cleanEmpty is the positive control that the compatible seam is not stuck-on-
// emit. Drives the SHIPPED seams (bash only, no go-apidiff binary).
func TestGoAPIDiffCompatibleExtraction(t *testing.T) {
	skipIfMissing(t, "bash")
	root := moduleRoot(t)

	additiveOnly := "example.com/cap\n  Compatible changes:\n  - Added: added\n"
	mixed := "example.com/cap\n  Incompatible changes:\n  - Removed: removed\n  Compatible changes:\n  - Added: added\n"
	incompatOnly := "example.com/cap\n  Incompatible changes:\n  - Removed: removed\n"
	cleanEmpty := ""

	for _, tc := range []struct {
		name    string
		apidiff string
		// compatContains: non-empty => the extract-compatible stdout must contain it;
		// empty => that stdout must be blank.
		compatContains    string
		compatNotContains string
		// payloadContains: cross-check on evaluate-apidiff — non-empty => the
		// incompatible payload must contain it; empty => it must be blank. Exit is
		// always 0 here (none of these inputs is fail-closed).
		payloadContains    string
		payloadNotContains string
	}{
		{"additive-only -> compatible has Added; incompatible empty", additiveOnly, "Added", "", "", ""},
		{"mixed -> compatible has Added not Removed; incompatible has Removed not Added", mixed, "Added", "Removed", "Removed", "Added"},
		{"incompat-only -> compatible empty; incompatible has Removed", incompatOnly, "", "", "Removed", ""},
		{"clean -> both channels empty", cleanEmpty, "", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compat := runExtractCompatible(t, root, tc.apidiff)
			if tc.compatContains == "" {
				if strings.TrimSpace(compat) != "" {
					t.Fatalf("%s: expected an EMPTY compatible payload, got:\n%s", tc.name, compat)
				}
			} else if !strings.Contains(compat, tc.compatContains) {
				t.Fatalf("%s: compatible payload missing %q:\n%s", tc.name, tc.compatContains, compat)
			}
			if tc.compatNotContains != "" && strings.Contains(compat, tc.compatNotContains) {
				t.Fatalf("%s: compatible payload unexpectedly contains %q (leak):\n%s", tc.name, tc.compatNotContains, compat)
			}

			// Cross-check: the incompatible seam on the SAME input never sees the
			// compatible bullet, and always exits 0 (no fail-closed input here).
			exit, payload := runEvaluateAPIDiff(t, root, tc.apidiff)
			if exit != 0 {
				t.Fatalf("%s: evaluate-apidiff exit = %d, want 0\npayload:\n%s", tc.name, exit, payload)
			}
			if tc.payloadContains == "" {
				if strings.TrimSpace(payload) != "" {
					t.Fatalf("%s: expected an EMPTY incompatible payload, got:\n%s", tc.name, payload)
				}
			} else if !strings.Contains(payload, tc.payloadContains) {
				t.Fatalf("%s: incompatible payload missing %q:\n%s", tc.name, tc.payloadContains, payload)
			}
			if tc.payloadNotContains != "" && strings.Contains(payload, tc.payloadNotContains) {
				t.Fatalf("%s: incompatible payload unexpectedly contains %q (leak):\n%s", tc.name, tc.payloadNotContains, payload)
			}
		})
	}
}

// readIfExists reports whether path exists and, if so, its content. An empty path (the
// "env intentionally unset" case) reports absent. A read error other than not-exist is
// fatal — it would mask a real gate defect.
func readIfExists(t *testing.T, path string) (bool, string) {
	t.Helper()
	if path == "" {
		return false, ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, ""
		}
		t.Fatalf("read %s: %v", path, err)
	}
	return true, string(b)
}

// TestGoAPIDiffGateRunner drives the SHIPPED runner (gate_go_apidiff) end-to-end across
// the tri-state incompatible signal (APIDIFF_CHANGES_FILE) and the compatible channel
// (APIDIFF_COMPATIBLE_FILE). It copies the COMMITTED contract-gates.sh (real shipped
// bytes) into a throwaway git repo so the runner's repo_root (the script's own ../)
// resolves to that repo, then runs the real go-apidiff 0.8.3 binary across two commits —
// so case (d) doubles as the real-format ANCHOR for --print-compatible + the compatible
// seam. The signal files are placed under a not-yet-existing subdirectory so the runner's
// `mkdir -p "$(dirname …)"` guard is exercised on every write.
func TestGoAPIDiffGateRunner(t *testing.T) {
	skipIfMissing(t, "go-apidiff")
	skipIfMissing(t, "git")
	skipIfMissing(t, "go")
	skipIfMissing(t, "bash")

	const stableOnly = `package gaterunner

func Stable() string { return "stable" }
`
	const stablePlusRemoved = `package gaterunner

func Stable() string { return "stable" }

func Removed() string { return "removed" }
`
	const stablePlusAdded = `package gaterunner

func Stable() string { return "stable" }

func Added() string { return "added" }
`

	root := moduleRoot(t)
	committed, err := os.ReadFile(filepath.Join(root, "scripts", "contract-gates.sh"))
	if err != nil {
		t.Fatalf("read committed gate script: %v", err)
	}

	// buildRepoWithScript builds a throwaway git repo whose base commit holds baseAPI and
	// head commit holds headAPI, with the COMMITTED gate script copied to
	// scripts/contract-gates.sh so the runner's repo_root resolves to this repo. Returns
	// (dir, baseSHA); go-apidiff single-arg mode diffs baseSHA against HEAD.
	buildRepoWithScript := func(t *testing.T, baseAPI, headAPI string) (string, string) {
		t.Helper()
		dir := t.TempDir()
		writeFile := func(name, content string) {
			p := filepath.Join(dir, name)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatalf("mkdir for %s: %v", name, err)
			}
			if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
		git := func(args ...string) string {
			cmd := exec.Command("git", args...)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(),
				"GIT_AUTHOR_NAME=gaterunner", "GIT_AUTHOR_EMAIL=gaterunner@invalid",
				"GIT_COMMITTER_NAME=gaterunner", "GIT_COMMITTER_EMAIL=gaterunner@invalid",
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
			}
			return strings.TrimSpace(string(out))
		}
		writeFile(filepath.Join("scripts", "contract-gates.sh"), string(committed))
		writeFile("go.mod", "module example.com/gaterunner\n\ngo 1.25\n")
		writeFile("api.go", baseAPI)
		git("init", "-q")
		git("add", ".")
		git("commit", "-q", "-m", "base")
		base := git("rev-parse", "HEAD")
		writeFile("api.go", headAPI)
		git("add", ".")
		// --allow-empty so the CLEAN case (headAPI == baseAPI, identical tree) still
		// produces a head commit; go-apidiff then diffs an identical API -> no changes.
		git("commit", "-q", "--allow-empty", "-m", "head")
		return dir, base
	}

	// runGate runs `bash scripts/contract-gates.sh go-apidiff <base>` in dir with the
	// chosen env files (an empty path leaves that env var UNSET), returning the exit code
	// plus each file's existence + content.
	runGate := func(t *testing.T, dir, base, changesFile, compatFile string) (exit int, changesExists bool, changesContent string, compatExists bool, compatContent string) {
		t.Helper()
		script := filepath.Join(dir, "scripts", "contract-gates.sh")
		cmd := exec.Command("bash", script, "go-apidiff", base)
		cmd.Dir = dir
		env := append(os.Environ(), "GOFLAGS=-mod=mod")
		if changesFile != "" {
			env = append(env, "APIDIFF_CHANGES_FILE="+changesFile)
		}
		if compatFile != "" {
			env = append(env, "APIDIFF_COMPATIBLE_FILE="+compatFile)
		}
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		t.Logf("gate runner output (exit %v):\n%s", err, out)
		if err != nil {
			ee, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("gate runner failed to run: %v\n%s", err, out)
			}
			exit = ee.ExitCode()
		}
		changesExists, changesContent = readIfExists(t, changesFile)
		compatExists, compatContent = readIfExists(t, compatFile)
		return
	}

	// (a) real break + BOTH envs set -> exit 0; changes file EXISTS + non-empty (names
	//     "Removed") = the warn signal; compatible file EXISTS + empty (no additive).
	t.Run("real break, both envs -> warn payload + empty compatible", func(t *testing.T) {
		dir, base := buildRepoWithScript(t, stablePlusRemoved, stableOnly)
		outDir := t.TempDir()
		changes := filepath.Join(outDir, "nested", "changes.txt")
		compat := filepath.Join(outDir, "nested", "compat.txt")
		exit, chExists, chBody, coExists, coBody := runGate(t, dir, base, changes, compat)
		if exit != 0 {
			t.Fatalf("real break: exit = %d, want 0 (advisory warn, non-blocking)", exit)
		}
		if !chExists || strings.TrimSpace(chBody) == "" {
			t.Fatalf("real break: changes file exists=%v content=%q; want EXISTS + non-empty (the warn signal)", chExists, chBody)
		}
		if !strings.Contains(chBody, "Removed") {
			t.Fatalf("real break: changes file does not name the removed symbol:\n%s", chBody)
		}
		if !coExists || strings.TrimSpace(coBody) != "" {
			t.Fatalf("real break: compatible file exists=%v content=%q; want EXISTS + empty (no additive change)", coExists, coBody)
		}
	})

	// (b) real break + envs UNSET -> exit 0 on the local no-op path. "No files written"
	// is a STRUCTURAL guarantee, not something this case can observe: every write in
	// gate_go_apidiff is guarded by `[[ -n "${APIDIFF_*_FILE:-}" ]]`, so with the env
	// unset there is no path to write to at all. The WRITE paths are meaningfully
	// asserted with the env SET by (a)/(c)/(d), and no-write-on-fail-closed (the
	// load-bearing empty-vs-absent distinction) by TestGoAPIDiffGateRunnerFailClosed.
	// So the only meaningful assertion here is that an advisory break still exits 0.
	t.Run("real break, envs unset -> exit 0 (local no-op path)", func(t *testing.T) {
		dir, base := buildRepoWithScript(t, stablePlusRemoved, stableOnly)
		exit, _, _, _, _ := runGate(t, dir, base, "", "")
		if exit != 0 {
			t.Fatalf("real break (envs unset): exit = %d, want 0 (advisory, non-blocking)", exit)
		}
	})

	// (c) CLEAN repo + envs set -> exit 0; changes file EXISTS + EMPTY (the "cleanliness
	//     established" signal, distinct from ABSENT = fail-closed); compatible file
	//     EXISTS + empty.
	t.Run("clean repo, envs set -> empty changes file (cleanliness established)", func(t *testing.T) {
		dir, base := buildRepoWithScript(t, stableOnly, stableOnly)
		outDir := t.TempDir()
		changes := filepath.Join(outDir, "nested", "changes.txt")
		compat := filepath.Join(outDir, "nested", "compat.txt")
		exit, chExists, chBody, coExists, coBody := runGate(t, dir, base, changes, compat)
		if exit != 0 {
			t.Fatalf("clean repo: exit = %d, want 0", exit)
		}
		if !chExists {
			t.Fatalf("clean repo: changes file was NOT created; want EXISTS + EMPTY (cleanliness established, not absent=fail-closed)")
		}
		if strings.TrimSpace(chBody) != "" {
			t.Fatalf("clean repo: changes file is not empty:\n%s", chBody)
		}
		if !coExists || strings.TrimSpace(coBody) != "" {
			t.Fatalf("clean repo: compatible file exists=%v content=%q; want EXISTS + empty", coExists, coBody)
		}
	})

	// (d) ADDITIVE-only repo + envs set -> exit 0; changes file EXISTS + EMPTY; compatible
	//     file EXISTS + non-empty containing "Added". Runs the REAL go-apidiff
	//     --print-compatible, so this is the format ANCHOR for the compatible channel:
	//     additive lands EMPTY on the incompatible channel and non-empty on the compatible
	//     one (the no-leak guarantee at the runner). The raw "- Added" bullet is asserted
	//     so a future format drift that stops emitting it fails here.
	t.Run("additive-only repo, envs set -> empty changes + compatible names Added", func(t *testing.T) {
		dir, base := buildRepoWithScript(t, stableOnly, stablePlusAdded)
		outDir := t.TempDir()
		changes := filepath.Join(outDir, "nested", "changes.txt")
		compat := filepath.Join(outDir, "nested", "compat.txt")
		exit, chExists, chBody, coExists, coBody := runGate(t, dir, base, changes, compat)
		if exit != 0 {
			t.Fatalf("additive-only: exit = %d, want 0", exit)
		}
		if !chExists || strings.TrimSpace(chBody) != "" {
			t.Fatalf("additive-only: changes file exists=%v content=%q; want EXISTS + EMPTY (no incompatible change)", chExists, chBody)
		}
		if !coExists || !strings.Contains(coBody, "Added") {
			t.Fatalf("additive-only: compatible file exists=%v content=%q; want EXISTS + containing \"Added\"", coExists, coBody)
		}
		if !strings.Contains(coBody, "- Added") {
			t.Fatalf("additive-only: compatible file lacks the raw \"- Added\" bullet (format drift?):\n%s", coBody)
		}
	})
}

// TestGoAPIDiffGateRunnerFailClosed shims go-apidiff on PATH to emit an "Incompatible
// changes" header with NO bullets (a synthetic format drift), and asserts the SHIPPED
// runner exits 1 (fail-closed) AND writes NEITHER signal file. An ABSENT changes file is
// precisely the "go-apidiff established nothing / gate blind" signal the comment step
// keys on (distinct from an EMPTY = clean file), because the runner returns BEFORE any
// write. No real go-apidiff binary is needed.
func TestGoAPIDiffGateRunnerFailClosed(t *testing.T) {
	skipIfMissing(t, "bash")
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Fatalf(
			"resolve bash for TestGoAPIDiffGateRunnerFailClosed after its interpreter precondition: %v; "+
				"the hermetic fake go-apidiff cannot be constructed or executed, so the caller cannot verify that the contract gate fails closed; "+
				"make bash available on PATH and rerun the test",
			err,
		)
	}

	dir := t.TempDir()
	root := moduleRoot(t)
	committed, err := os.ReadFile(filepath.Join(root, "scripts", "contract-gates.sh"))
	if err != nil {
		t.Fatalf("read committed gate script: %v", err)
	}
	scriptsDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("mkdir scripts: %v", err)
	}
	script := filepath.Join(scriptsDir, "contract-gates.sh")
	if err := os.WriteFile(script, committed, 0o644); err != nil {
		t.Fatalf("write gate script: %v", err)
	}

	// PATH shim: a fake go-apidiff that emits a header with no parseable bullets. This is
	// the UNPARSEABLE incompatible section that must trip the fail-closed guard.
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	shim := "#!" + bashPath + "\n" +
		"printf '%s\\n' 'example.com/x'\n" +
		"printf '%s\\n' '  Incompatible changes:'\n" +
		"exit 0\n"
	if err := os.WriteFile(filepath.Join(binDir, "go-apidiff"), []byte(shim), 0o755); err != nil {
		t.Fatalf("write go-apidiff shim: %v", err)
	}

	changesFile := filepath.Join(dir, "changes.txt")
	compatFile := filepath.Join(dir, "compat.txt")

	cmd := exec.Command("bash", script, "go-apidiff", "dummybase")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"APIDIFF_CHANGES_FILE="+changesFile,
		"APIDIFF_COMPATIBLE_FILE="+compatFile,
	)
	out, err := cmd.CombinedOutput()
	t.Logf("fail-closed runner output (exit %v):\n%s", err, out)

	exit := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("fail-closed runner failed to run: %v\n%s", err, out)
		}
		exit = ee.ExitCode()
	}
	if exit != 1 {
		t.Fatalf("fail-closed: exit = %d, want 1 (the gate must stop the line when it can no longer parse the incompatible section)", exit)
	}
	if _, statErr := os.Stat(changesFile); !os.IsNotExist(statErr) {
		t.Fatalf("fail-closed: APIDIFF_CHANGES_FILE was written (stat err = %v); want ABSENT (established nothing)", statErr)
	}
	// Belt: the compatible file must also be absent — the runner returns before BOTH
	// writes, so a fail-closed run touches neither channel.
	if _, statErr := os.Stat(compatFile); !os.IsNotExist(statErr) {
		t.Fatalf("fail-closed: APIDIFF_COMPATIBLE_FILE was written (stat err = %v); want ABSENT", statErr)
	}
}

// TestGoAPIDiffGateRunnerInvocationFailure proves that the SHIPPED runner treats a
// non-zero go-apidiff exit without an incompatible-report header as a tooling failure
// rather than parsing its potentially partial output. The fake emits an ordinary-looking
// compatible report before exiting non-zero, so restoring output-suppressing error
// handling would incorrectly make this test green and write the clean/compatible signals.
func TestGoAPIDiffGateRunnerInvocationFailure(t *testing.T) {
	skipIfMissing(t, "bash")
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Fatalf(
			"resolve bash for TestGoAPIDiffGateRunnerInvocationFailure after its interpreter precondition: %v; "+
				"the hermetic fake go-apidiff cannot be constructed or executed, so the caller cannot verify that an invocation failure fails closed; "+
				"make bash available on PATH and rerun the test",
			err,
		)
	}

	dir := t.TempDir()
	root := moduleRoot(t)
	committed, err := os.ReadFile(filepath.Join(root, "scripts", "contract-gates.sh"))
	if err != nil {
		t.Fatalf("read committed gate script: %v", err)
	}
	scriptsDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("mkdir scripts: %v", err)
	}
	script := filepath.Join(scriptsDir, "contract-gates.sh")
	if err := os.WriteFile(script, committed, 0o644); err != nil {
		t.Fatalf("write gate script: %v", err)
	}

	// PATH shim: ordinary-looking compatible output is not a valid report when the
	// tool itself exits non-zero. The runner must return before parsing or signaling.
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	shim := "#!" + bashPath + "\n" +
		"printf '%s\\n' 'example.com/x'\n" +
		"printf '%s\\n' '  Compatible changes:'\n" +
		"printf '%s\\n' '  - Added: example.com/x.New'\n" +
		"exit 1\n"
	if err := os.WriteFile(filepath.Join(binDir, "go-apidiff"), []byte(shim), 0o755); err != nil {
		t.Fatalf("write go-apidiff shim: %v", err)
	}

	changesFile := filepath.Join(dir, "changes.txt")
	compatFile := filepath.Join(dir, "compat.txt")
	cmd := exec.Command("bash", script, "go-apidiff", "dummybase")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"APIDIFF_CHANGES_FILE="+changesFile,
		"APIDIFF_COMPATIBLE_FILE="+compatFile,
	)
	out, err := cmd.CombinedOutput()
	t.Logf("invocation-failure runner output (exit %v):\n%s", err, out)

	exit := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("invocation-failure runner failed to run: %v\n%s", err, out)
		}
		exit = ee.ExitCode()
	}
	if exit != 1 {
		t.Fatalf("invocation failure: exit = %d, want 1 (a non-zero go-apidiff exit must fail closed)", exit)
	}
	if !strings.Contains(string(out), "Compatible changes") || !strings.Contains(string(out), "exit 1") {
		t.Fatalf("invocation failure: expected ordinary-looking tool output and the captured exit status, got:\n%s", out)
	}
	if _, statErr := os.Stat(changesFile); !os.IsNotExist(statErr) {
		t.Fatalf("invocation failure: APIDIFF_CHANGES_FILE was written (stat err = %v); want ABSENT (tool did not establish a report)", statErr)
	}
	if _, statErr := os.Stat(compatFile); !os.IsNotExist(statErr) {
		t.Fatalf("invocation failure: APIDIFF_COMPATIBLE_FILE was written (stat err = %v); want ABSENT (tool did not establish a report)", statErr)
	}
}
