package release_test

import (
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

// TestDecideBubbleTopologyFacts (B1) checks the happy-path mapping of topology
// facts to decisions: a single-parent squash sitting at develop's tip proceeds,
// while each non-proceed shape is reached from a realistic fact set.
func TestDecideBubbleTopologyFacts(t *testing.T) {
	cases := []struct {
		name  string
		facts release.BubbleFacts
		want  release.BubbleDecision
	}{
		{
			name: "tip is the squash, ready to bubble",
			// develop just fast-forwarded T -> S; tip == S (single-parent, S.tree).
			facts: release.BubbleFacts{
				SParentCount:   1,
				TipParentCount: 1,
				TipTreeEqualsS: true,
				TipEqualsS:     true,
			},
			want: release.BubbleProceed,
		},
		{
			name: "tip is the already-built bubble merge-commit",
			// tip == M: two parents [T, S], M.tree == S.tree, distinct from S.
			facts: release.BubbleFacts{
				SParentCount:   1,
				TipParentCount: 2,
				TipTreeEqualsS: true,
				TipEqualsS:     false,
			},
			want: release.BubbleSkipAlreadyBubbled,
		},
		{
			name: "tip advanced to an unrelated commit",
			// someone pushed more work; tip's tree differs and it is not S.
			facts: release.BubbleFacts{
				SParentCount:   1,
				TipParentCount: 1,
				TipTreeEqualsS: false,
				TipEqualsS:     false,
			},
			want: release.BubbleRetryTipAdvanced,
		},
		{
			name: "target is not a single-parent squash",
			facts: release.BubbleFacts{
				SParentCount:   2,
				TipParentCount: 1,
				TipTreeEqualsS: true,
				TipEqualsS:     true,
			},
			want: release.BubbleSkipNotSquash,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := release.DecideBubble(tc.facts); got != tc.want {
				t.Fatalf("DecideBubble(%+v) = %v, want %v", tc.facts, got, tc.want)
			}
		})
	}
}

// TestDecideBubbleNotSquash (B4) checks the R-B rule: ANY parent count other
// than 1 means the target is not a native squash-merge, regardless of the tip
// topology — and NotSquash is the highest-precedence decision.
func TestDecideBubbleNotSquash(t *testing.T) {
	for _, parents := range []int{0, 2, 3} {
		facts := release.BubbleFacts{
			SParentCount: parents,
			// Tip facts that would otherwise say Proceed — NotSquash must win.
			TipParentCount: 1,
			TipTreeEqualsS: true,
			TipEqualsS:     true,
		}
		if got := release.DecideBubble(facts); got != release.BubbleSkipNotSquash {
			t.Fatalf("DecideBubble(SParentCount=%d) = %v, want BubbleSkipNotSquash", parents, got)
		}
	}
}

// TestDecideBubblePrecedence (B12) checks the pinned precedence on OVERLAPPING
// facts: a tip that is not S yet is a genuine bubble (two parents, tree==S.tree)
// must resolve to AlreadyBubbled, not TipAdvanced; and NotSquash dominates even
// that already-bubbled shape.
func TestDecideBubblePrecedence(t *testing.T) {
	t.Run("already-bubbled wins over tip-advanced", func(t *testing.T) {
		// Overlap: tip != S (could read as "advanced") AND tip.tree == S.tree
		// with two parents (the real bubble M). AlreadyBubbled must win.
		facts := release.BubbleFacts{
			SParentCount:   1,
			TipParentCount: 2,
			TipTreeEqualsS: true,
			TipEqualsS:     false,
		}
		if got := release.DecideBubble(facts); got != release.BubbleSkipAlreadyBubbled {
			t.Fatalf("DecideBubble(%+v) = %v, want BubbleSkipAlreadyBubbled", facts, got)
		}
	})

	t.Run("not-squash dominates already-bubbled", func(t *testing.T) {
		// Same already-bubbled tip shape, but S is not a squash. NotSquash wins.
		facts := release.BubbleFacts{
			SParentCount:   2,
			TipParentCount: 2,
			TipTreeEqualsS: true,
			TipEqualsS:     false,
		}
		if got := release.DecideBubble(facts); got != release.BubbleSkipNotSquash {
			t.Fatalf("DecideBubble(%+v) = %v, want BubbleSkipNotSquash", facts, got)
		}
	})

	t.Run("same-tree single-parent non-S tip is not already-bubbled", func(t *testing.T) {
		// A non-merge commit that happens to share S's tree is NOT a bubble; the
		// safe call is to re-read (TipAdvanced), never a silent AlreadyBubbled no-op.
		facts := release.BubbleFacts{
			SParentCount:   1,
			TipParentCount: 1,
			TipTreeEqualsS: true,
			TipEqualsS:     false,
		}
		if got := release.DecideBubble(facts); got != release.BubbleRetryTipAdvanced {
			t.Fatalf("DecideBubble(%+v) = %v, want BubbleRetryTipAdvanced", facts, got)
		}
	})
}

// TestBubbleFactsFromCommits checks the own-types realignment: the mechanical
// projection of (squash S, tip) release.GitCommit own-types onto the topology
// facts DecideBubble consumes, and that the projected facts drive the expected
// decision end-to-end.
func TestBubbleFactsFromCommits(t *testing.T) {
	const treeS = "5dc4ea0f1122"

	cases := []struct {
		name         string
		squash, tip  release.GitCommit
		wantFacts    release.BubbleFacts
		wantDecision release.BubbleDecision
	}{
		{
			name:         "tip is the single-parent squash (proceed)",
			squash:       release.GitCommit{SHA: "S", TreeSHA: treeS, ParentSHAs: []string{"T"}},
			tip:          release.GitCommit{SHA: "S", TreeSHA: treeS, ParentSHAs: []string{"T"}},
			wantFacts:    release.BubbleFacts{SParentCount: 1, TipParentCount: 1, TipTreeEqualsS: true, TipEqualsS: true},
			wantDecision: release.BubbleProceed,
		},
		{
			name:         "tip is the two-parent bubble M over S (already-bubbled)",
			squash:       release.GitCommit{SHA: "S", TreeSHA: treeS, ParentSHAs: []string{"T"}},
			tip:          release.GitCommit{SHA: "M", TreeSHA: treeS, ParentSHAs: []string{"T", "S"}},
			wantFacts:    release.BubbleFacts{SParentCount: 1, TipParentCount: 2, TipTreeEqualsS: true, TipEqualsS: false},
			wantDecision: release.BubbleSkipAlreadyBubbled,
		},
		{
			name:         "S is not a single-parent squash (not-squash)",
			squash:       release.GitCommit{SHA: "S", TreeSHA: treeS, ParentSHAs: []string{"P1", "P2"}},
			tip:          release.GitCommit{SHA: "S", TreeSHA: treeS, ParentSHAs: []string{"P1", "P2"}},
			wantFacts:    release.BubbleFacts{SParentCount: 2, TipParentCount: 2, TipTreeEqualsS: true, TipEqualsS: true},
			wantDecision: release.BubbleSkipNotSquash,
		},
		{
			name:         "tip advanced past S to a different single-parent commit",
			squash:       release.GitCommit{SHA: "S", TreeSHA: treeS, ParentSHAs: []string{"T"}},
			tip:          release.GitCommit{SHA: "X", TreeSHA: "otherTree", ParentSHAs: []string{"S"}},
			wantFacts:    release.BubbleFacts{SParentCount: 1, TipParentCount: 1, TipTreeEqualsS: false, TipEqualsS: false},
			wantDecision: release.BubbleRetryTipAdvanced,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := release.BubbleFactsFromCommits(tc.squash, tc.tip)
			if got != tc.wantFacts {
				t.Fatalf("BubbleFactsFromCommits = %+v, want %+v", got, tc.wantFacts)
			}
			if d := release.DecideBubble(got); d != tc.wantDecision {
				t.Fatalf("DecideBubble(projected facts) = %v, want %v", d, tc.wantDecision)
			}
		})
	}
}

// TestBubbleMergeMessage (B6) checks the subject + trailer rendering: fixed
// trailer order, the empty-trailer (subject-only) case, the unresolved-title
// fallback, and single-line subject hardening against special characters.
func TestBubbleMergeMessage(t *testing.T) {
	cases := []struct {
		name string
		prov release.PRProvenance
		want string
	}{
		{
			name: "full provenance renders subject then ordered trailers",
			prov: release.PRProvenance{
				Number:       42,
				Title:        "add bubble logic",
				ClosesIssues: []int{7, 9},
				ApprovedBy:   []string{"alice"},
				ReviewedBy:   []string{"bob"},
				CoAuthoredBy: []string{"Carol <carol@example.com>"},
			},
			want: "Merge PR #42: add bubble logic\n\n" +
				"Closes #7\n" +
				"Closes #9\n" +
				"Approved-by: alice\n" +
				"Reviewed-by: bob\n" +
				"Co-authored-by: Carol <carol@example.com>",
		},
		{
			name: "empty trailers yield a subject-only message",
			prov: release.PRProvenance{
				Number: 5,
				Title:  "tidy docs",
			},
			want: "Merge PR #5: tidy docs",
		},
		{
			name: "unresolved title falls back to numbered subject",
			prov: release.PRProvenance{
				Number:     11,
				Title:      "",
				ApprovedBy: []string{"alice"},
			},
			want: "Merge PR #11\n\nApproved-by: alice",
		},
		{
			name: "whitespace-only title is treated as unresolved",
			prov: release.PRProvenance{
				Number: 12,
				Title:  "   ",
			},
			want: "Merge PR #12",
		},
		{
			name: "special characters in title pass through but newlines are flattened",
			prov: release.PRProvenance{
				Number:       3,
				Title:        "fix: handle #hash & \"quotes\"\nDROP TABLE",
				ClosesIssues: []int{1},
			},
			want: "Merge PR #3: fix: handle #hash & \"quotes\" DROP TABLE\n\nCloses #1",
		},
		{
			name: "CRLF and tab in title are flattened to single spaces",
			prov: release.PRProvenance{
				Number: 4,
				Title:  "line one\r\nline two\ttabbed",
			},
			want: "Merge PR #4: line one line two tabbed",
		},
		{
			name: "newline in a trailer value cannot forge an extra trailer",
			prov: release.PRProvenance{
				Number:       6,
				Title:        "guard trailers",
				ApprovedBy:   []string{"alice\nReviewed-by: mallory"},
				CoAuthoredBy: []string{"Carol <carol@example.com>\r\nDROP"},
			},
			want: "Merge PR #6: guard trailers\n\n" +
				"Approved-by: alice Reviewed-by: mallory\n" +
				"Co-authored-by: Carol <carol@example.com> DROP",
		},
		{
			name: "is not the PR body — only provenance trailers",
			prov: release.PRProvenance{
				Number:       8,
				Title:        "feature",
				CoAuthoredBy: []string{"Dan <dan@example.com>", "Eve <eve@example.com>"},
			},
			want: "Merge PR #8: feature\n\n" +
				"Co-authored-by: Dan <dan@example.com>\n" +
				"Co-authored-by: Eve <eve@example.com>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := release.BubbleMergeMessage(tc.prov); got != tc.want {
				t.Fatalf("BubbleMergeMessage() =\n%q\nwant\n%q", got, tc.want)
			}
		})
	}
}
