package schema

import (
	"fmt"
	"strings"
)

// --- Session commits ---

// SessionCommitsResponse is the response of GET /api/v1/sessions/{id}/commits.
//
// The rows are the commits a stored session recorded, which is the same
// session-to-commit binding the review surfaces read rather than a git log
// replay. The envelope is a read projection: it carries no git context and no
// rewrite or ghost mapping, because a caller that needs those reads the session
// detail instead.
//
// A stored session with no recorded commits answers an empty array, never null.
// An identifier that names no stored session is a 404 rather than an empty
// list, so "this session recorded nothing" and "this session is not here" stay
// distinguishable to the caller.
type SessionCommitsResponse struct {
	Commits []CommitInfo `json:"commits" nullable:"false"`
}

// Validate checks that the envelope carries an array. A null array would let a
// caller read "no commits recorded" and "the read never produced a list" as the
// same answer.
func (r SessionCommitsResponse) Validate() error {
	if r.Commits == nil {
		return fmt.Errorf("session commits validation failed at schema.SessionCommitsResponse.Validate during local read construction: commits is null; a stored session with no recorded commits must emit an empty array; initialize the slice so an empty answer cannot be misread as a failed read")
	}
	return nil
}

// --- Session pull requests ---

// SessionPullRequest is one pull request detected in a stored session.
//
// A row is detected from the session's own recorded tool calls, never from the
// repository's current state: it exists only when one recorded call both
// invokes `gh pr create` and returns a GitHub pull request URL. The answer is
// therefore display-only evidence of what the session did. It is not an access
// boundary, it never widens who may read the transcript, and nothing is posted
// to GitHub on the caller's behalf.
type SessionPullRequest struct {
	Owner  string `json:"owner" minLength:"1"` // GitHub repository owner (user or organization) the pull request was opened against
	Repo   string `json:"repo" minLength:"1"`  // GitHub repository name the pull request was opened against
	Number int    `json:"number" minimum:"1"`  // pull request number within that repository
	// URL is the canonical pull request address recorded in the tool call
	// result. It may carry no sub-page, so it is the identity of the pull
	// request rather than a link to one of its tabs.
	URL string `json:"url" minLength:"1"`
	// Source is an open string naming how the row was detected, so a later
	// detector can be added without rewriting the rows already stored in a
	// consumer's cache. The only value the contract currently defines is
	// gh_pr_create: a recorded `gh pr create` invocation whose result carried a
	// GitHub pull request URL. It is deliberately not a closed set: a consumer
	// reads the rows it knows and does not refuse the response over a value it
	// does not know, and an empty string is not a detector at all.
	Source string `json:"source" minLength:"1"`
}

// Validate checks that the row names its repository, a positive pull request
// number, the canonical address, and the detector that found it. A row missing
// any of them cannot be shown as evidence of what the session did, and the
// fields cannot be recovered from one another.
func (p SessionPullRequest) Validate() error {
	if strings.TrimSpace(p.Owner) == "" {
		return fmt.Errorf("session pull request validation failed at schema.SessionPullRequest.Validate during local read construction: owner is empty; a caller cannot attribute the row to a GitHub repository; emit the repository owner recorded in the detected pull request URL")
	}
	if strings.TrimSpace(p.Repo) == "" {
		return fmt.Errorf("session pull request validation failed at schema.SessionPullRequest.Validate during local read construction: repo is empty for %q; a caller cannot attribute the row to a repository; emit the repository name recorded in the detected pull request URL", p.Owner)
	}
	if p.Number < 1 {
		return fmt.Errorf("session pull request validation failed at schema.SessionPullRequest.Validate during local read construction: number is %d for %q/%q; a pull request number is one or greater; emit the number recorded in the detected pull request URL", p.Number, p.Owner, p.Repo)
	}
	if strings.TrimSpace(p.URL) == "" {
		return fmt.Errorf("session pull request validation failed at schema.SessionPullRequest.Validate during local read construction: url is empty for %q/%q#%d; the row is not evidence without its canonical address; emit the pull request URL the recorded tool call returned", p.Owner, p.Repo, p.Number)
	}
	if strings.TrimSpace(p.Source) == "" {
		return fmt.Errorf("session pull request validation failed at schema.SessionPullRequest.Validate during local read construction: source is empty for %q/%q#%d; an unknown detector cannot be told apart from a missing one; emit the detector label, currently gh_pr_create", p.Owner, p.Repo, p.Number)
	}
	return nil
}

// SessionPullRequestsResponse is the response of
// GET /api/v1/sessions/{id}/pull-requests.
//
// A stored session with no detected pull request answers an empty array, never
// null. An identifier that names no stored session is a 404 rather than an
// empty list, for the same reason the commits read distinguishes them.
type SessionPullRequestsResponse struct {
	PullRequests []SessionPullRequest `json:"pullRequests" nullable:"false"`
}

// Validate checks that the envelope carries an array and that every row names a
// pull request and the detector that found it, so an empty detection and a
// detection that never ran stay distinguishable.
func (r SessionPullRequestsResponse) Validate() error {
	if r.PullRequests == nil {
		return fmt.Errorf("session pull requests validation failed at schema.SessionPullRequestsResponse.Validate during local read construction: pullRequests is null; a stored session with no detected pull request must emit an empty array; initialize the slice so an empty answer cannot be misread as a failed detection")
	}
	for _, pull := range r.PullRequests {
		if err := pull.Validate(); err != nil {
			return err
		}
	}
	return nil
}
