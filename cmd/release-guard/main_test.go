package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

func TestRunGreenFiltersRunsByRCCommit(t *testing.T) {
	const (
		workflow = "release.yml"
		rcTag    = "v1.2.3-rc1"
		rcCommit = "0123456789abcdef0123456789abcdef01234567"
	)

	var calls [][]string
	restoreCommandOutput(t, func(name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		switch {
		case name == "git" && slices.Equal(args, []string{"rev-list", "-n", "1", rcTag}):
			return []byte(rcCommit + "\n"), nil
		case name == "gh":
			return []byte(fmt.Sprintf(`[{"headSha":%q,"status":"completed","conclusion":"success"}]`, rcCommit)), nil
		default:
			return nil, fmt.Errorf("unexpected command: %s %v", name, args)
		}
	})

	green, err := runGreen(workflow, rcTag)
	if err != nil {
		t.Fatalf("runGreen(%q, %q): %v", workflow, rcTag, err)
	}
	if !green {
		t.Fatalf("runGreen(%q, %q) = false, want true", workflow, rcTag)
	}

	wantGH := []string{
		"gh", "run", "list",
		"--workflow", workflow,
		"--commit", rcCommit,
		"--limit", "100",
		"--json", "headSha,status,conclusion",
	}
	if len(calls) != 2 {
		t.Fatalf("command calls = %v, want 2 calls", calls)
	}
	if !slices.Equal(calls[1], wantGH) {
		t.Fatalf("gh run list args = %v, want %v", calls[1], wantGH)
	}
}

func TestRunGreenRequiresCompletedSuccess(t *testing.T) {
	const (
		workflow = "release.yml"
		rcTag    = "v1.2.3-rc1"
		rcCommit = "0123456789abcdef0123456789abcdef01234567"
	)

	restoreCommandOutput(t, func(name string, args ...string) ([]byte, error) {
		switch name {
		case "git":
			return []byte(rcCommit + "\n"), nil
		case "gh":
			return []byte(fmt.Sprintf(`[
				{"headSha":%q,"status":"in_progress","conclusion":"success"},
				{"headSha":%q,"status":"completed","conclusion":"failure"},
				{"headSha":"fedcba9876543210fedcba9876543210fedcba98","status":"completed","conclusion":"success"}
			]`, rcCommit, rcCommit)), nil
		default:
			return nil, fmt.Errorf("unexpected command: %s %v", name, args)
		}
	})

	green, err := runGreen(workflow, rcTag)
	if err != nil {
		t.Fatalf("runGreen(%q, %q): %v", workflow, rcTag, err)
	}
	if green {
		t.Fatalf("runGreen(%q, %q) = true, want false", workflow, rcTag)
	}
}

func TestParseReviewsFlattensSlurpedPages(t *testing.T) {
	data := []byte(`[
		[
			{"user":{"login":"alice"},"state":"APPROVED"}
		],
		[
			{"user":{"login":"alice"},"state":"CHANGES_REQUESTED"},
			{"user":{"login":"bob"},"state":"APPROVED"}
		]
	]`)

	reviews, err := parseReviews(data)
	if err != nil {
		t.Fatalf("parseReviews(): %v", err)
	}
	if len(reviews) != 3 {
		t.Fatalf("parseReviews() returned %d reviews, want 3", len(reviews))
	}
	if got, want := reviews[0].User.Login, "alice"; got != want {
		t.Fatalf("first review login = %q, want %q", got, want)
	}
	if got, want := reviews[1].State, release.ReviewStateChangesRequested; got != want {
		t.Fatalf("second review state = %q, want %q", got, want)
	}

	approvers := release.LatestApprovers(reviews)
	if !slices.Equal(approvers, []string{"bob"}) {
		t.Fatalf("LatestApprovers(parseReviews(slurped pages)) = %v, want [bob]", approvers)
	}
}

func TestParseReviewsAcceptsSingleArray(t *testing.T) {
	data := []byte(`[
		{"user":{"login":"alice"},"state":"APPROVED"},
		{"user":{"login":"alice"},"state":"COMMENTED"}
	]`)

	reviews, err := parseReviews(data)
	if err != nil {
		t.Fatalf("parseReviews(): %v", err)
	}
	if len(reviews) != 2 {
		t.Fatalf("parseReviews() returned %d reviews, want 2", len(reviews))
	}
	if got, want := release.LatestApprovers(reviews), []string{"alice"}; !slices.Equal(got, want) {
		t.Fatalf("LatestApprovers(parseReviews(single array)) = %v, want %v", got, want)
	}
}

func TestFetchReviewsUsesPaginateSlurp(t *testing.T) {
	const (
		repo = "peasant-labs/schema"
		pr   = "94"
	)

	var calls [][]string
	restoreCommandOutput(t, func(name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		if name != "gh" {
			return nil, fmt.Errorf("unexpected command: %s %v", name, args)
		}
		return []byte(`[[{"user":{"login":"alice"},"state":"APPROVED"}]]`), nil
	})

	reviews, err := fetchReviews(repo, pr)
	if err != nil {
		t.Fatalf("fetchReviews(%q, %q): %v", repo, pr, err)
	}
	if len(reviews) != 1 {
		t.Fatalf("fetchReviews(%q, %q) returned %d reviews, want 1", repo, pr, len(reviews))
	}

	want := []string{
		"gh", "api",
		"repos/peasant-labs/schema/pulls/94/reviews",
		"--paginate",
		"--slurp",
	}
	if len(calls) != 1 {
		t.Fatalf("command calls = %v, want 1 call", calls)
	}
	if !slices.Equal(calls[0], want) {
		t.Fatalf("fetchReviews gh args = %v, want %v", calls[0], want)
	}
}

func TestFindMaintainerApproverChecksApproversInOrder(t *testing.T) {
	const repo = "peasant-labs/schema"

	var calls [][]string
	restoreCommandOutput(t, func(name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		if name != "gh" {
			return nil, fmt.Errorf("unexpected command: %s %v", name, args)
		}
		switch args[1] {
		case "repos/peasant-labs/schema/collaborators/alice/permission":
			return []byte("write\n"), nil
		case "repos/peasant-labs/schema/collaborators/bob/permission":
			return []byte("maintain\n"), nil
		default:
			return nil, fmt.Errorf("unexpected gh args: %v", args)
		}
	})

	approver, perm, err := findMaintainerApprover(repo, []string{"alice", "bob", "carol"})
	if err != nil {
		t.Fatalf("findMaintainerApprover(): %v", err)
	}
	if approver != "bob" || perm != release.PermMaintain {
		t.Fatalf("findMaintainerApprover() = (%q, %q), want (bob, %q)", approver, perm, release.PermMaintain)
	}
	if len(calls) != 2 {
		t.Fatalf("command calls = %v, want 2 calls", calls)
	}
}

func TestRunCheckApprovalExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		wantOK   bool
	}{
		{
			name:     "no standing approver",
			scenario: "no-approver",
			wantOK:   false,
		},
		{
			name:     "approvals but no maintainer approver",
			scenario: "no-maintainer",
			wantOK:   false,
		},
		{
			name:     "fetch error",
			scenario: "fetch-error",
			wantOK:   false,
		},
		{
			name:     "standing maintainer approval",
			scenario: "maintainer",
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=TestRunCheckApprovalExitCodesHelper")
			cmd.Env = append(os.Environ(),
				"RELEASE_GUARD_CHECK_APPROVAL_HELPER="+tt.scenario,
				"GITHUB_REPOSITORY=peasant-labs/schema",
			)
			out, err := cmd.CombinedOutput()
			gotOK := err == nil
			if gotOK != tt.wantOK {
				t.Fatalf("check-approval exit success = %v, want %v; err=%v output:\n%s", gotOK, tt.wantOK, err, out)
			}
			if !tt.wantOK {
				if _, ok := err.(*exec.ExitError); !ok {
					t.Fatalf("check-approval error = %T %[1]v, want exec.ExitError; output:\n%s", err, out)
				}
			}
		})
	}
}

func TestRunCheckApprovalExitCodesHelper(t *testing.T) {
	scenario := os.Getenv("RELEASE_GUARD_CHECK_APPROVAL_HELPER")
	if scenario == "" {
		return
	}

	commandOutput = func(name string, args ...string) ([]byte, error) {
		if name != "gh" {
			return nil, fmt.Errorf("unexpected command: %s %v", name, args)
		}
		if slices.Equal(args, []string{"api", "repos/peasant-labs/schema/pulls/94/reviews", "--paginate", "--slurp"}) {
			switch scenario {
			case "no-approver":
				return []byte(`[[{"user":{"login":"alice"},"state":"CHANGES_REQUESTED"}]]`), nil
			case "no-maintainer", "maintainer":
				return []byte(`[[{"user":{"login":"alice"},"state":"APPROVED"}]]`), nil
			case "fetch-error":
				return nil, fmt.Errorf("review fetch failed")
			default:
				return nil, fmt.Errorf("unknown helper scenario %q", scenario)
			}
		}
		if slices.Equal(args, []string{"api", "repos/peasant-labs/schema/collaborators/alice/permission", "--jq", ".permission"}) {
			switch scenario {
			case "no-maintainer":
				return []byte("write\n"), nil
			case "maintainer":
				return []byte("maintain\n"), nil
			default:
				return nil, fmt.Errorf("unexpected collaborator lookup for scenario %q", scenario)
			}
		}
		return nil, fmt.Errorf("unexpected gh args: %v", args)
	}

	os.Args = []string{"release-guard", "check-approval", "--pr", "94"}
	main()
	os.Exit(0)
}

func restoreCommandOutput(t *testing.T, fn func(string, ...string) ([]byte, error)) {
	t.Helper()
	orig := commandOutput
	commandOutput = fn
	t.Cleanup(func() {
		commandOutput = orig
	})
}
