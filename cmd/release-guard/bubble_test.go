package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

// fakeGraph is an in-memory GitHub commit graph + develop ref backing the bubble
// orchestration tests. It models exactly the reads/writes the bubbler makes
// (Ref/Commit/Pull/PullReviews/CreateCommit/UpdateRefFastForward) so runBubble is
// exercised end-to-end against a scriptable dependency — the SUT (bubbler) is
// never mocked, only its GitHubClient seam is.
type fakeGraph struct {
	commits map[string]release.GitCommit
	pulls   map[int]release.Pull
	reviews map[int][]release.Review

	ref string // current develop tip sha

	created  []release.NewCommit // every CreateCommit call, in order
	mCount   int                 // assigns deterministic M shas
	upserts  int                 // UpdateRefFastForward call count
	ffReject int                 // reject the first N UpdateRefFastForward calls as non-FF
	hardErr  error               // if set, UpdateRefFastForward returns this (non-FF) error
	onReject func(g *fakeGraph)  // invoked on each non-FF rejection (tip churn)
}

func newFakeGraph() *fakeGraph {
	return &fakeGraph{
		commits: map[string]release.GitCommit{},
		pulls:   map[int]release.Pull{},
		reviews: map[int][]release.Review{},
	}
}

func (g *fakeGraph) addCommit(c release.GitCommit) { g.commits[c.SHA] = c }

// newBubbleMock wires a *mockGitHubClient (the seam named in main_test.go) to the
// graph, so the bubbler talks to the graph through the real GitHubClient methods.
func newBubbleMock(g *fakeGraph) *mockGitHubClient {
	return &mockGitHubClient{
		refFn: func(_ context.Context, _, _ string) (release.GitRef, error) {
			return release.GitRef{SHA: g.ref}, nil
		},
		commitFn: func(_ context.Context, _, sha string) (release.GitCommit, error) {
			c, ok := g.commits[sha]
			if !ok {
				return release.GitCommit{}, fmt.Errorf("fake: no commit %s", sha)
			}
			return c, nil
		},
		pullFn: func(_ context.Context, _ string, n int) (release.Pull, error) {
			p, ok := g.pulls[n]
			if !ok {
				return release.Pull{}, fmt.Errorf("fake: no PR #%d", n)
			}
			return p, nil
		},
		pullReviewsFn: func(_ context.Context, _ string, n int) ([]release.Review, error) {
			return g.reviews[n], nil
		},
		createCommitFn: func(_ context.Context, _ string, in release.NewCommit) (release.GitCommit, error) {
			g.created = append(g.created, in)
			g.mCount++
			sha := fmt.Sprintf("M%d", g.mCount)
			c := release.GitCommit{SHA: sha, TreeSHA: in.TreeSHA, ParentSHAs: in.ParentSHAs, Message: in.Message}
			g.addCommit(c)
			return c, nil
		},
		updateRefFastForwardFn: func(_ context.Context, _, _, newSHA string) error {
			g.upserts++
			if g.upserts <= g.ffReject {
				if g.onReject != nil {
					g.onReject(g)
				}
				return fmt.Errorf("fake: rejected: %w", release.ErrNotFastForward)
			}
			if g.hardErr != nil {
				return g.hardErr
			}
			g.ref = newSHA
			return nil
		},
	}
}

func review(login, state string) release.Review {
	return release.Review{User: &release.ReviewUser{Login: login}, State: release.ReviewState(state)}
}

// mergeBoundary returns a two-parent merge commit T (the drain-all boundary).
func mergeBoundary(sha string) release.GitCommit {
	return release.GitCommit{SHA: sha, TreeSHA: "treeT", ParentSHAs: []string{"p0", "p1"}, Message: "Merge boundary"}
}

// --- happy path: single squash bubbles into one signed merge commit ------------

func TestBubbleRun_SingleSquash(t *testing.T) {
	g := newFakeGraph()
	g.addCommit(mergeBoundary("T"))
	g.addCommit(release.GitCommit{SHA: "S", TreeSHA: "treeS", ParentSHAs: []string{"T"}, Message: "feat: add thing (#42)"})
	g.ref = "S"
	g.pulls[42] = release.Pull{Number: 42, Title: "Add thing", Body: "Closes #7"}
	g.reviews[42] = []release.Review{review("alice", "APPROVED"), review("bob", "COMMENTED")}

	var stdout, stderr bytes.Buffer
	err := bubbleRun(context.Background(), newBubbleMock(g), "peasant-labs/schema", &stdout, &stderr, []string{"--pr", "42"})
	if err != nil {
		t.Fatalf("bubbleRun: %v\nstderr:\n%s", err, stderr.String())
	}

	if len(g.created) != 1 {
		t.Fatalf("CreateCommit called %d times, want 1: %+v", len(g.created), g.created)
	}
	m := g.created[0]
	if m.TreeSHA != "treeS" {
		t.Fatalf("M.tree = %q, want treeS (M.tree == S.tree)", m.TreeSHA)
	}
	if len(m.ParentSHAs) != 2 || m.ParentSHAs[0] != "T" || m.ParentSHAs[1] != "S" {
		t.Fatalf("M.parents = %v, want [T S] (first-parent boundary T)", m.ParentSHAs)
	}
	if !strings.HasPrefix(m.Message, "Merge PR #42: Add thing") {
		t.Fatalf("M.message = %q, want it to start with the PR subject", m.Message)
	}
	if !strings.Contains(m.Message, "Closes #7") || !strings.Contains(m.Message, "Approved-by: alice") || !strings.Contains(m.Message, "Reviewed-by: bob") {
		t.Fatalf("M.message missing expected trailers:\n%s", m.Message)
	}
	if g.ref != "M1" {
		t.Fatalf("develop advanced to %q, want M1", g.ref)
	}
	if !strings.Contains(stdout.String(), "bubble complete") {
		t.Fatalf("stdout missing success line:\n%s", stdout.String())
	}
}

// --- drain-all: two stacked squashes chain oldest-first ------------------------

func TestBubbleRun_StackedSquashesChainOldestFirst(t *testing.T) {
	g := newFakeGraph()
	g.addCommit(mergeBoundary("T"))
	g.addCommit(release.GitCommit{SHA: "S1", TreeSHA: "tree1", ParentSHAs: []string{"T"}, Message: "first (#1)"})
	g.addCommit(release.GitCommit{SHA: "S2", TreeSHA: "tree2", ParentSHAs: []string{"S1"}, Message: "second (#2)"})
	g.ref = "S2"
	g.pulls[1] = release.Pull{Number: 1, Title: "First"}
	g.pulls[2] = release.Pull{Number: 2, Title: "Second"}

	var stdout, stderr bytes.Buffer
	if err := bubbleRun(context.Background(), newBubbleMock(g), "peasant-labs/schema", &stdout, &stderr, nil); err != nil {
		t.Fatalf("bubbleRun: %v\nstderr:\n%s", err, stderr.String())
	}

	if len(g.created) != 2 {
		t.Fatalf("CreateCommit called %d times, want 2 (drain-all): %+v", len(g.created), g.created)
	}
	// M1 = merge[T, S1] (oldest first); M2 = merge[M1, S2].
	if got := g.created[0].ParentSHAs; len(got) != 2 || got[0] != "T" || got[1] != "S1" {
		t.Fatalf("M1.parents = %v, want [T S1]", got)
	}
	if got := g.created[1].ParentSHAs; len(got) != 2 || got[0] != "M1" || got[1] != "S2" {
		t.Fatalf("M2.parents = %v, want [M1 S2] (first-parent chain)", got)
	}
	if g.created[0].TreeSHA != "tree1" || g.created[1].TreeSHA != "tree2" {
		t.Fatalf("M trees = [%q %q], want [tree1 tree2]", g.created[0].TreeSHA, g.created[1].TreeSHA)
	}
	if g.ref != "M2" {
		t.Fatalf("develop advanced to %q, want M2 (final M)", g.ref)
	}
}

// --- retry: a non-FF rejection re-reads the tip and rebuilds, then succeeds -----

func TestBubbleRun_RetriesOnNonFastForward(t *testing.T) {
	g := newFakeGraph()
	g.addCommit(mergeBoundary("T"))
	g.addCommit(release.GitCommit{SHA: "S", TreeSHA: "treeS", ParentSHAs: []string{"T"}, Message: "feat (#9)"})
	g.ref = "S"
	g.pulls[9] = release.Pull{Number: 9, Title: "Nine"}
	g.ffReject = 1 // first FF-CAS rejected; second succeeds

	var stdout, stderr bytes.Buffer
	if err := bubbleRun(context.Background(), newBubbleMock(g), "peasant-labs/schema", &stdout, &stderr, nil); err != nil {
		t.Fatalf("bubbleRun: %v\nstderr:\n%s", err, stderr.String())
	}

	if g.upserts != 2 {
		t.Fatalf("UpdateRefFastForward called %d times, want 2 (one reject + one success)", g.upserts)
	}
	// The plan was rebuilt on the retry: CreateCommit ran once per attempt.
	if len(g.created) != 2 {
		t.Fatalf("CreateCommit called %d times, want 2 (rebuilt M-chain across 2 attempts)", len(g.created))
	}
	if g.ref != "M2" {
		t.Fatalf("develop advanced to %q, want M2 (second attempt's M)", g.ref)
	}
	if !strings.Contains(stderr.String(), "retrying") {
		t.Fatalf("stderr missing retry notice:\n%s", stderr.String())
	}
}

// --- terminal: persistent non-FF exhausts the bounded retry, fails loud ---------

func TestBubbleRun_TerminalAfterPersistentNonFastForward(t *testing.T) {
	g := newFakeGraph()
	g.addCommit(mergeBoundary("T"))
	g.addCommit(release.GitCommit{SHA: "S", TreeSHA: "treeS", ParentSHAs: []string{"T"}, Message: "feat (#3)"})
	g.ref = "S"
	g.pulls[3] = release.Pull{Number: 3, Title: "Three"}
	g.ffReject = 99 // never succeeds

	var stdout, stderr bytes.Buffer
	err := bubbleRun(context.Background(), newBubbleMock(g), "peasant-labs/schema", &stdout, &stderr, []string{"--max-attempts", "3"})
	if err == nil {
		t.Fatalf("bubbleRun succeeded, want a TerminalError")
	}
	var terminal *TerminalError
	if !errors.As(err, &terminal) {
		t.Fatalf("error is %T (%v), want *TerminalError", err, err)
	}
	if terminal.Attempts != 3 {
		t.Fatalf("TerminalError.Attempts = %d, want 3", terminal.Attempts)
	}
	if g.upserts != 3 {
		t.Fatalf("UpdateRefFastForward attempted %d times, want 3 (bounded)", g.upserts)
	}
	if g.ref != "S" {
		t.Fatalf("develop tip = %q, want S unchanged (no partial write on give-up)", g.ref)
	}
	if !strings.Contains(stdout.String(), "::error::") {
		t.Fatalf("stdout missing GitHub Actions ::error:: annotation:\n%s", stdout.String())
	}
}

// --- a genuine non-FF-agnostic UpdateRef error is NOT retried -------------------

func TestBubbleRun_NonRetryableUpdateErrorFailsImmediately(t *testing.T) {
	g := newFakeGraph()
	g.addCommit(mergeBoundary("T"))
	g.addCommit(release.GitCommit{SHA: "S", TreeSHA: "treeS", ParentSHAs: []string{"T"}, Message: "feat (#5)"})
	g.ref = "S"
	g.pulls[5] = release.Pull{Number: 5, Title: "Five"}
	g.hardErr = errors.New("github client: cannot update git ref: 403 forbidden")

	var stdout, stderr bytes.Buffer
	err := bubbleRun(context.Background(), newBubbleMock(g), "peasant-labs/schema", &stdout, &stderr, nil)
	if err == nil {
		t.Fatalf("bubbleRun succeeded, want the hard UpdateRef error surfaced")
	}
	if errors.As(err, new(*TerminalError)) {
		t.Fatalf("a non-FF hard error was misreported as a TerminalError: %v", err)
	}
	if g.upserts != 1 {
		t.Fatalf("UpdateRefFastForward attempted %d times, want 1 (no retry on a non-FF error)", g.upserts)
	}
	if !strings.Contains(err.Error(), "403 forbidden") {
		t.Fatalf("error %q does not carry the underlying cause", err.Error())
	}
}

// --- skip: already-bubbled tip is a no-op --------------------------------------

func TestBubbleRun_SkipAlreadyBubbled(t *testing.T) {
	g := newFakeGraph()
	g.addCommit(release.GitCommit{SHA: "S", TreeSHA: "treeS", ParentSHAs: []string{"T"}, Message: "feat (#1)"})
	// tip M: two parents [T, S], M.tree == S.tree, S single-parent -> AlreadyBubbled.
	g.addCommit(release.GitCommit{SHA: "M", TreeSHA: "treeS", ParentSHAs: []string{"T", "S"}, Message: "Merge PR #1"})
	g.ref = "M"

	var stdout, stderr bytes.Buffer
	if err := bubbleRun(context.Background(), newBubbleMock(g), "peasant-labs/schema", &stdout, &stderr, nil); err != nil {
		t.Fatalf("bubbleRun: %v", err)
	}
	if len(g.created) != 0 || g.upserts != 0 {
		t.Fatalf("already-bubbled tip must be a no-op; created=%d upserts=%d", len(g.created), g.upserts)
	}
	if g.ref != "M" {
		t.Fatalf("develop tip = %q, want M unchanged", g.ref)
	}
}

// --- skip: a stable non-bubble two-parent merge tip is NotSquash, not a retry ---

func TestBubbleRun_SkipStableMergeTip(t *testing.T) {
	g := newFakeGraph()
	// Second parent S' is itself a two-parent merge (not a single-parent squash),
	// so the tip is a regular merge, not our bubble -> NotSquash skip.
	g.addCommit(release.GitCommit{SHA: "S2p", TreeSHA: "treeX", ParentSHAs: []string{"a", "b"}, Message: "sub-merge"})
	g.addCommit(release.GitCommit{SHA: "MG", TreeSHA: "treeMG", ParentSHAs: []string{"T", "S2p"}, Message: "regular merge"})
	g.ref = "MG"

	var stdout, stderr bytes.Buffer
	if err := bubbleRun(context.Background(), newBubbleMock(g), "peasant-labs/schema", &stdout, &stderr, []string{"--max-attempts", "2"}); err != nil {
		t.Fatalf("bubbleRun: %v", err)
	}
	if len(g.created) != 0 || g.upserts != 0 {
		t.Fatalf("stable merge tip must be a no-op skip; created=%d upserts=%d", len(g.created), g.upserts)
	}
}

// --- no-PR fallback: a squash with no (#n) suffix renders "Merge commit <sha>" --

func TestBubbleRun_NoPRSuffixFallback(t *testing.T) {
	g := newFakeGraph()
	g.addCommit(mergeBoundary("T"))
	g.addCommit(release.GitCommit{
		SHA: "abcdef0123456789", TreeSHA: "treeS", ParentSHAs: []string{"T"},
		Message: "hotfix with no pr suffix\n\nCo-authored-by: Carol <carol@example.com>\nCloses #12",
	})
	g.ref = "abcdef0123456789"

	var stdout, stderr bytes.Buffer
	if err := bubbleRun(context.Background(), newBubbleMock(g), "peasant-labs/schema", &stdout, &stderr, nil); err != nil {
		t.Fatalf("bubbleRun: %v\nstderr:\n%s", err, stderr.String())
	}
	if len(g.created) != 1 {
		t.Fatalf("CreateCommit called %d times, want 1", len(g.created))
	}
	msg := g.created[0].Message
	if !strings.HasPrefix(msg, "Merge commit abcdef012345:") {
		t.Fatalf("no-PR message subject = %q, want a \"Merge commit <short-sha>:\" prefix", msg)
	}
	if !strings.Contains(msg, "Closes #12") || !strings.Contains(msg, "Co-authored-by: Carol <carol@example.com>") {
		t.Fatalf("no-PR message missing trailers parsed from the commit body:\n%s", msg)
	}
}

// --- flag validation -----------------------------------------------------------

func TestBubbleRun_FlagValidation(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"unknown flag", []string{"--nope"}, "unknown flag"},
		{"missing --pr value", []string{"--pr"}, "--pr requires a value"},
		{"non-numeric --pr", []string{"--pr", "abc"}, "--pr must be a positive integer"},
		{"bad --max-attempts", []string{"--max-attempts", "0"}, "--max-attempts must be a positive integer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newFakeGraph()
			var stdout, stderr bytes.Buffer
			err := bubbleRun(context.Background(), newBubbleMock(g), "peasant-labs/schema", &stdout, &stderr, tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("bubbleRun(%v) error = %v, want it to contain %q", tc.args, err, tc.want)
			}
			if len(g.created) != 0 || g.upserts != 0 {
				t.Fatalf("a flag error must touch no git state")
			}
		})
	}
}

// --- pure provenance parsers ---------------------------------------------------

func TestProvenanceParsers(t *testing.T) {
	if n, ok := parseTrailingPRNumber("feat: thing (#42)"); !ok || n != 42 {
		t.Fatalf("parseTrailingPRNumber = (%d,%v), want (42,true)", n, ok)
	}
	if _, ok := parseTrailingPRNumber("no suffix here"); ok {
		t.Fatalf("parseTrailingPRNumber matched a subject with no suffix")
	}
	if got := stripTrailingPRNumber("feat: thing (#42)"); got != "feat: thing" {
		t.Fatalf("stripTrailingPRNumber = %q, want %q", got, "feat: thing")
	}
	if got := parseCoAuthors("body\nCo-authored-by: A <a@x>\nCo-authored-by: B <b@x>"); len(got) != 2 || got[0] != "A <a@x>" || got[1] != "B <b@x>" {
		t.Fatalf("parseCoAuthors = %v, want [A <a@x> B <b@x>]", got)
	}
	if got := parseClosesIssues("Fixes #3 and closes #3 and resolves #5"); len(got) != 2 || got[0] != 3 || got[1] != 5 {
		t.Fatalf("parseClosesIssues = %v, want [3 5] (deduped, ordered)", got)
	}
	if got := distinctReviewers([]release.Review{review("a", "APPROVED"), review("a", "COMMENTED"), review("b", "APPROVED")}); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("distinctReviewers = %v, want [a b]", got)
	}
}
