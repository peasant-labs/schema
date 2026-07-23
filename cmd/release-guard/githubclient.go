package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v88/github"
	"github.com/peasant-labs/schema/internal/release"
)

// GitHubClient is the release-guard's narrow view of the GitHub REST API: the
// three calls the release gates make. It returns internal/release own-types
// (never go-github types) so the policy layer stays free of any go-github
// import — the production wrapper below is the single go-github seam.
type GitHubClient interface {
	// CollaboratorPermission returns user's permission level on repo
	// ("owner/repo"). Backs check-maintainer / check-approval.
	CollaboratorPermission(ctx context.Context, repo, user string) (release.CollaboratorPermission, error)
	// WorkflowRunsForCommit returns every run of workflowFile whose head commit
	// is commitSHA, filtered server-side. Backs check-final (RunGreenForCommit).
	WorkflowRunsForCommit(ctx context.Context, repo, workflowFile, commitSHA string) ([]release.WorkflowRun, error)
	// PullReviews returns every review on PR prNumber, across all pages, in the
	// API's ascending order. Backs check-approval (LatestApprovers).
	PullReviews(ctx context.Context, repo string, prNumber int) ([]release.Review, error)

	// Ref resolves a git reference (e.g. "heads/develop") to the commit SHA it
	// points at — the tip the squash-merge bubble reads before fast-forwarding.
	Ref(ctx context.Context, repo, ref string) (release.GitRef, error)
	// Commit reads a git commit object, exposing its tree SHA and parent SHAs
	// (used to test the not-a-squash and already-bubbled bubble decisions).
	Commit(ctx context.Context, repo, sha string) (release.GitCommit, error)
	// Pull returns the number/title/body of a single pull request, for building
	// per-PR provenance in the bubble merge message.
	Pull(ctx context.Context, repo string, number int) (release.Pull, error)
	// CreateCommit creates a git commit via the Git Data API and returns it
	// (including its server-assigned SHA). Commits created this way are signed by
	// GitHub and show as "Verified".
	CreateCommit(ctx context.Context, repo string, in release.NewCommit) (release.GitCommit, error)
	// UpdateRefFastForward advances a git reference to newSHA as a server-side
	// fast-forward compare-and-swap (force is always false, so mainline history is
	// never clobbered). It returns release.ErrNotFastForward when GitHub rejects
	// the update as non-fast-forward (the caller re-reads the tip and retries).
	UpdateRefFastForward(ctx context.Context, repo, ref, newSHA string) error
}

// githubClient is the production GitHubClient: a thin wrapper over the go-github
// *github.Client. It is the ONLY go-github seam in release-guard — it maps
// go-github response types onto internal/release own-types and nil-guards the
// pointer fields here so the policy layer never sees a nil deref.
type githubClient struct {
	gh *github.Client
}

// newGitHubClient builds the production GitHubClient from a GitHub token.
// github.WithAuthToken sends the token as a bearer credential on every request
// (the same auth the previous `gh` shell-outs relied on via GH_TOKEN) and,
// in go-github v88, rejects an empty token at construction — so this is also a
// fail-fast on a missing GH_TOKEN before any API call is attempted.
//
// NOTE: go-github v88's NewClient returns (*Client, error) (the constructor was
// reshaped to options funcs), so this returns an error too. main, the sole
// composition root, surfaces it as the actionable fail-fast.
func newGitHubClient(token string) (GitHubClient, error) {
	gh, err := github.NewClient(github.WithAuthToken(token))
	if err != nil {
		return nil, fmt.Errorf("github client: cannot construct the GitHub API client for release-guard: %w. Set GH_TOKEN to a non-empty token with repo read access before invoking a check-* subcommand", err)
	}
	return &githubClient{gh: gh}, nil
}

// splitRepo splits a "$GITHUB_REPOSITORY" value ("owner/repo") into the separate
// owner and repo arguments the go-github methods take. The interface carries a
// single repo string (matching the old gh shell-outs); the split happens here,
// at the wrapper boundary.
func splitRepo(repo string) (owner, name string, err error) {
	owner, name, found := strings.Cut(repo, "/")
	if !found || owner == "" || name == "" || strings.Contains(name, "/") {
		return "", "", fmt.Errorf("github client: repository %q is not in owner/repo form during a release-guard API call; set $GITHUB_REPOSITORY to e.g. peasant-labs/schema so the call can address the repository", repo)
	}
	return owner, name, nil
}

func (c *githubClient) CollaboratorPermission(ctx context.Context, repo, user string) (release.CollaboratorPermission, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return "", err
	}
	level, _, err := c.gh.Repositories.GetPermissionLevel(ctx, owner, name, user)
	if err != nil {
		return "", fmt.Errorf("github client: cannot read the collaborator permission of %q on %s/%s during release-guard: %w. Confirm the user exists and GH_TOKEN can read repo collaborators", user, owner, name, err)
	}
	return release.CollaboratorPermission(level.GetPermission()), nil
}

func (c *githubClient) WorkflowRunsForCommit(ctx context.Context, repo, workflowFile, commitSHA string) ([]release.WorkflowRun, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	// HeadSHA filters server-side, so the old client-side 100-run scan (and its
	// silent cap) is gone. Paginate so a commit with many reruns is fully
	// covered.
	opts := &github.ListWorkflowRunsOptions{
		HeadSHA:     commitSHA,
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var runs []release.WorkflowRun
	for {
		page, resp, err := c.gh.Actions.ListWorkflowRunsByFileName(ctx, owner, name, workflowFile, opts)
		if err != nil {
			return nil, fmt.Errorf("github client: cannot list runs of %s for commit %s on %s/%s during check-final: %w. Confirm the workflow file name and that GH_TOKEN can read Actions", workflowFile, commitSHA, owner, name, err)
		}
		for _, r := range page.WorkflowRuns {
			if r == nil {
				continue
			}
			runs = append(runs, release.WorkflowRun{
				HeadSHA:    r.GetHeadSHA(),
				Status:     release.WorkflowRunStatus(r.GetStatus()),
				Conclusion: release.WorkflowRunConclusion(r.GetConclusion()),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return runs, nil
}

func (c *githubClient) PullReviews(ctx context.Context, repo string, prNumber int) ([]release.Review, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	// ListReviews returns one page at a time; the NextPage loop collects ALL
	// pages in ascending order so LatestApprovers sees the same review sequence
	// the previous `gh api --paginate --slurp` produced.
	opts := &github.ListOptions{PerPage: 100}
	var reviews []release.Review
	for {
		page, resp, err := c.gh.PullRequests.ListReviews(ctx, owner, name, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("github client: cannot list reviews for PR #%d on %s/%s during check-approval: %w. Confirm the PR number and that GH_TOKEN can read pull requests", prNumber, owner, name, err)
		}
		for _, r := range page {
			if r == nil {
				continue
			}
			var user *release.ReviewUser
			if r.User != nil {
				user = &release.ReviewUser{Login: r.User.GetLogin()}
			}
			reviews = append(reviews, release.Review{
				User:  user,
				State: release.ReviewState(r.GetState()),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return reviews, nil
}

func (c *githubClient) Ref(ctx context.Context, repo, ref string) (release.GitRef, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return release.GitRef{}, err
	}
	// go-github's GetRef trims a leading "refs/" and escapes each ref segment, so
	// "heads/develop" (or "refs/heads/develop") both address the same ref.
	r, _, err := c.gh.Git.GetRef(ctx, owner, name, ref)
	if err != nil {
		return release.GitRef{}, fmt.Errorf("github client: cannot resolve git ref %q on %s/%s during release bubble: %w. Confirm the ref exists and GH_TOKEN can read the repository's git data", ref, owner, name, err)
	}
	return release.GitRef{SHA: r.GetObject().GetSHA()}, nil
}

func (c *githubClient) Commit(ctx context.Context, repo, sha string) (release.GitCommit, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return release.GitCommit{}, err
	}
	commit, _, err := c.gh.Git.GetCommit(ctx, owner, name, sha)
	if err != nil {
		return release.GitCommit{}, fmt.Errorf("github client: cannot read git commit %s on %s/%s during release bubble: %w. Confirm the commit SHA and that GH_TOKEN can read the repository's git data", sha, owner, name, err)
	}
	return mapGitCommit(commit), nil
}

func (c *githubClient) Pull(ctx context.Context, repo string, number int) (release.Pull, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return release.Pull{}, err
	}
	pr, _, err := c.gh.PullRequests.Get(ctx, owner, name, number)
	if err != nil {
		return release.Pull{}, fmt.Errorf("github client: cannot read pull request #%d on %s/%s during release bubble: %w. Confirm the PR number and that GH_TOKEN can read pull requests", number, owner, name, err)
	}
	return release.Pull{
		Number: pr.GetNumber(),
		Title:  pr.GetTitle(),
		Body:   pr.GetBody(),
	}, nil
}

func (c *githubClient) CreateCommit(ctx context.Context, repo string, in release.NewCommit) (release.GitCommit, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return release.GitCommit{}, err
	}
	parents := make([]*github.Commit, 0, len(in.ParentSHAs))
	for _, p := range in.ParentSHAs {
		parents = append(parents, &github.Commit{SHA: github.Ptr(p)})
	}
	commit := github.Commit{
		Message: github.Ptr(in.Message),
		Tree:    &github.Tree{SHA: github.Ptr(in.TreeSHA)},
		Parents: parents,
	}
	created, _, err := c.gh.Git.CreateCommit(ctx, owner, name, commit, nil)
	if err != nil {
		return release.GitCommit{}, fmt.Errorf("github client: cannot create git commit (tree %s over %d parent(s)) on %s/%s during release bubble: %w. Confirm the tree and parent SHAs exist and that GH_TOKEN can write the repository's git data", in.TreeSHA, len(in.ParentSHAs), owner, name, err)
	}
	return mapGitCommit(created), nil
}

func (c *githubClient) UpdateRefFastForward(ctx context.Context, repo, ref, newSHA string) error {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return err
	}
	// Force is pinned false so this is a server-side fast-forward compare-and-swap:
	// a non-fast-forward update (which would rewrite mainline history) is rejected
	// by GitHub, never forced through.
	update := github.UpdateRef{SHA: newSHA, Force: github.Ptr(false)}
	if _, _, err := c.gh.Git.UpdateRef(ctx, owner, name, ref, update); err != nil {
		// Precisely distinguish "the ref advanced under us" (retryable) from any
		// other 422 (bad SHA, malformed ref) — keyed on the 422 status AND the
		// documented "is not a fast forward" message, exactly as the retired
		// stdlib client keyed on the response body.
		var errResp *github.ErrorResponse
		if errors.As(err, &errResp) &&
			errResp.Response != nil &&
			errResp.Response.StatusCode == http.StatusUnprocessableEntity &&
			strings.Contains(strings.ToLower(errResp.Message), "fast forward") {
			return fmt.Errorf("github client: cannot fast-forward ref %q to %s on %s/%s during release bubble: %w", ref, newSHA, owner, name, release.ErrNotFastForward)
		}
		return fmt.Errorf("github client: cannot update git ref %q to %s on %s/%s during release bubble: %w. Confirm the ref and SHA exist and that GH_TOKEN can write the repository's git data", ref, newSHA, owner, name, err)
	}
	return nil
}

// mapGitCommit projects a go-github *Commit onto the release.GitCommit own-type,
// nil-guarding the pointer fields (tree, parents, and each parent's SHA) here at
// the wrapper boundary so the policy layer never sees a nil deref.
func mapGitCommit(commit *github.Commit) release.GitCommit {
	if commit == nil {
		return release.GitCommit{}
	}
	parents := make([]string, 0, len(commit.Parents))
	for _, p := range commit.Parents {
		if p == nil {
			continue
		}
		parents = append(parents, p.GetSHA())
	}
	return release.GitCommit{
		SHA:        commit.GetSHA(),
		TreeSHA:    commit.GetTree().GetSHA(),
		ParentSHAs: parents,
		Message:    commit.GetMessage(),
	}
}
