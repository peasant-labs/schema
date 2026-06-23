package main

import (
	"context"
	"fmt"
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
// (the same auth the previous `gh` shell-outs relied on via GITHUB_TOKEN) and,
// in go-github v88, rejects an empty token at construction — so this is also a
// fail-fast on a missing GITHUB_TOKEN before any API call is attempted.
//
// NOTE: go-github v88's NewClient returns (*Client, error) (the constructor was
// reshaped to options funcs), so this returns an error too. main (SLICE-4), the
// sole composition root, surfaces it as the actionable fail-fast.
func newGitHubClient(token string) (GitHubClient, error) {
	gh, err := github.NewClient(github.WithAuthToken(token))
	if err != nil {
		return nil, fmt.Errorf("github client: cannot construct the GitHub API client for release-guard: %w. Set GITHUB_TOKEN to a non-empty token with repo read access before invoking a check-* subcommand", err)
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
		return "", fmt.Errorf("github client: cannot read the collaborator permission of %q on %s/%s during release-guard: %w. Confirm the user exists and GITHUB_TOKEN can read repo collaborators", user, owner, name, err)
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
			return nil, fmt.Errorf("github client: cannot list runs of %s for commit %s on %s/%s during check-final: %w. Confirm the workflow file name and that GITHUB_TOKEN can read Actions", workflowFile, commitSHA, owner, name, err)
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
			return nil, fmt.Errorf("github client: cannot list reviews for PR #%d on %s/%s during check-approval: %w. Confirm the PR number and that GITHUB_TOKEN can read pull requests", prNumber, owner, name, err)
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
