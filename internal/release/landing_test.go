package release_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

// GOLDEN SOURCE REF (G-DRY breadcrumb ↔ ~/dotfiles/scripts/git-prmerge):
//
//	dotfiles repo ref : b7167775a9eaefc764d37d39b0f97618a815ddc7
//	git-prmerge blob  : bf10fdc84b5ffd290bdce66bd235b6831f693efe
//	git-prmerge commit: 104bef2 (2026-06-07)
//
// testdata/squash_message.golden pins OUR SquashMessage format as a ONE-TIME
// authoring-parity snapshot of git-prmerge's step-2 message construction.
// git-prmerge is an external dotfiles tool NEVER run by schema CI, so this is a
// snapshot — NOT a continuous-parity check. If you intentionally change the
// format in release.SquashMessage, update this golden AND git-prmerge step 2 in
// the same change (see the breadcrumb on SquashMessage).
func TestSquashMessageGolden(t *testing.T) {
	const (
		prTitle = "feat: add auto-land triangle"
		prBody  = "Implements the CI auto-land squash+merge-commit triangle for schema PR landings.\n\nCloses #114"
	)
	headlines := []string{
		"feat(release): SquashMessage golden-mirror of git-prmerge",
		"feat(release-guard): land subcommand (reshape -> poll-clean -> recheck -> merge)",
		"ci: thin auto-land.yml workflow",
	}

	got := release.SquashMessage(prTitle, prBody, headlines)

	goldenPath := filepath.Join("testdata", "squash_message.golden")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden %s: %v", goldenPath, err)
	}
	if got != string(want) {
		t.Fatalf("SquashMessage output drifted from the git-prmerge snapshot golden.\n--- got ---\n%q\n--- want (%s) ---\n%q\nIf this change is intentional, update the golden AND git-prmerge step 2 together.",
			got, goldenPath, string(want))
	}
}

// TestSquashMessageStructure asserts the structural invariants of the format
// independently of the golden bytes: title, blank line, body, the "---" rule,
// the "Squashed commits:" header, and oldest-first "- headline" lines.
func TestSquashMessageStructure(t *testing.T) {
	got := release.SquashMessage("My title", "Body line.", []string{"first", "second"})
	want := "My title\n\nBody line.\n\n---\nSquashed commits:\n- first\n- second\n"
	if got != want {
		t.Fatalf("SquashMessage structure = %q, want %q", got, want)
	}
}

func TestSquashMessageEmptyBody(t *testing.T) {
	// git-prmerge feeds `.body // ""` so an empty body is a real input.
	got := release.SquashMessage("Title only", "", []string{"only commit"})
	want := "Title only\n\n\n\n---\nSquashed commits:\n- only commit\n"
	if got != want {
		t.Fatalf("SquashMessage(empty body) = %q, want %q", got, want)
	}
}

func TestMergeCommitTitle(t *testing.T) {
	got := release.MergeCommitTitle("feat--auto-land-ci", "develop")
	want := "Merge branch 'feat--auto-land-ci' into 'develop'"
	if got != want {
		t.Fatalf("MergeCommitTitle = %q, want %q", got, want)
	}
}

func TestMergeCommitBody(t *testing.T) {
	tests := []struct {
		name        string
		closesIssue string
		prNumber    int
		want        string
	}{
		{name: "with closes", closesIssue: "114", prNumber: 207, want: "Closes #114\n\nSee PR #207"},
		{name: "no closes", closesIssue: "", prNumber: 207, want: "See PR #207"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := release.MergeCommitBody(tt.closesIssue, tt.prNumber); got != tt.want {
				t.Fatalf("MergeCommitBody(%q, %d) = %q, want %q", tt.closesIssue, tt.prNumber, got, tt.want)
			}
		})
	}
}

// TestClassifyMergeGate exercises the load-bearing F1 decision directly. The
// two discriminating safety rows are: a "blocked" PR (whose `.mergeable` boolean
// would be TRUE) is REFUSED, and a "clean" PR whose checks have NOT settled (the
// stale pre-push clean) is told to WAIT, not proceed.
func TestClassifyMergeGate(t *testing.T) {
	tests := []struct {
		name          string
		state         release.MergeableState
		headMatches   bool
		checksSettled bool
		want          release.MergeGateDecision
	}{
		{name: "clean + settled -> proceed", state: release.MergeClean, headMatches: true, checksSettled: true, want: release.MergeGateProceed},
		{name: "clean but checks NOT settled (stale pre-push clean) -> wait", state: release.MergeClean, headMatches: true, checksSettled: false, want: release.MergeGateWait},
		{name: "clean but push not reflected yet -> wait", state: release.MergeClean, headMatches: false, checksSettled: true, want: release.MergeGateWait},
		{name: "blocked (mergeable boolean would be true) -> refuse", state: release.MergeBlocked, headMatches: true, checksSettled: true, want: release.MergeGateRefuse},
		{name: "unknown (recomputing) -> wait", state: release.MergeUnknown, headMatches: true, checksSettled: true, want: release.MergeGateWait},
		{name: "empty state -> wait", state: release.MergeableState(""), headMatches: true, checksSettled: true, want: release.MergeGateWait},
		{name: "behind (base advanced) -> reshape", state: release.MergeBehind, headMatches: true, checksSettled: true, want: release.MergeGateReshape},
		{name: "dirty -> refuse", state: release.MergeDirty, headMatches: true, checksSettled: true, want: release.MergeGateRefuse},
		{name: "unstable -> refuse", state: release.MergeUnstable, headMatches: true, checksSettled: true, want: release.MergeGateRefuse},
		{name: "draft -> refuse", state: release.MergeDraft, headMatches: true, checksSettled: true, want: release.MergeGateRefuse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := release.ClassifyMergeGate(tt.state, tt.headMatches, tt.checksSettled)
			if got != tt.want {
				t.Fatalf("ClassifyMergeGate(%q, head=%v, settled=%v) = %v, want %v", tt.state, tt.headMatches, tt.checksSettled, got, tt.want)
			}
		})
	}
}

func TestExtractClosesIssue(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "closes", body: "blah\n\nCloses #114", want: "114"},
		{name: "fixes lowercase", body: "fixes #7 in the parser", want: "7"},
		{name: "resolved", body: "Resolved #42.", want: "42"},
		{name: "first wins", body: "Closes #1\nalso fixes #2", want: "1"},
		{name: "none", body: "no references here #notanissue", want: ""},
		{name: "empty", body: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := release.ExtractClosesIssue(tt.body); got != tt.want {
				t.Fatalf("ExtractClosesIssue(%q) = %q, want %q", tt.body, got, tt.want)
			}
		})
	}
}
