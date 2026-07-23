package release

import "errors"

// GitRef is a git reference resolved to the commit SHA it points at. It is the
// own-type projection of GitHub's Git Data "get a reference" response (mirroring
// Review / WorkflowRun / CollaboratorPermission) so the squash-merge bubble
// orchestrator never touches a *github.Reference and the go-github pointer-field
// nil-guards stay at the cmd/release-guard wrapper boundary.
type GitRef struct {
	// SHA is the object SHA the reference points at — the tip the bubbler reads
	// before it fast-forwards the ref.
	SHA string
}

// GitCommit is the subset of a git commit object the bubble needs: its own SHA,
// the tree it points at, and its parents (first-parent first). It is the decoded
// shape of both "get a commit" (GET git/commits/{sha}) and the "create a commit"
// (POST git/commits) response, so the same own-type serves the not-a-squash /
// already-bubbled decisions and the created merge commit.
type GitCommit struct {
	// SHA is the commit's own object SHA (server-assigned on create).
	SHA string
	// TreeSHA is the SHA of the tree this commit points at. A bubble merge commit
	// M over squash S carries S's tree, so M.TreeSHA == S.TreeSHA.
	TreeSHA string
	// ParentSHAs are the parent commit SHAs in order. For a bubble merge commit M
	// over squash S onto tip T, ParentSHAs is [T, S] (first-parent T).
	ParentSHAs []string
}

// NewCommit is the input to GitHubClient.CreateCommit: the fields the caller
// supplies to build a git commit via the Git Data API. Kept distinct from
// GitCommit (the response own-type) so the created SHA is not something a caller
// can pretend to know before the server assigns it.
type NewCommit struct {
	// Message is the commit message.
	Message string
	// TreeSHA is the SHA of the tree the new commit points at.
	TreeSHA string
	// ParentSHAs are the parent commit SHAs in order (first-parent first). For a
	// bubble merge commit M over squash S onto tip T, this is [T, S].
	ParentSHAs []string
}

// Pull is the subset of a GitHub pull request the bubbler needs to build per-PR
// provenance in the bubble merge message: the number and title (body carried for
// context). Own-type projection of the "get a pull request" response so the
// policy layer stays free of any go-github import.
type Pull struct {
	// Number is the pull request number.
	Number int
	// Title is the pull request title.
	Title string
	// Body is the pull request body (markdown), available for provenance context.
	Body string
}

// ErrNotFastForward is returned (wrapped) by GitHubClient.UpdateRefFastForward
// when GitHub rejects a ref update because it would not be a fast-forward (HTTP
// 422 whose message says "...is not a fast forward"). With the update forced to
// a non-clobbering fast-forward compare-and-swap, this is the server-side signal
// that the ref advanced since its tip was read: the bubble caller re-reads the
// ref tip and retries. A 422 for any OTHER reason (bad SHA, malformed ref) is
// surfaced as its own actionable error, never misreported as non-fast-forward.
var ErrNotFastForward = errors.New("release: ref update rejected as non-fast-forward — the ref advanced since its tip was read; re-read the ref tip and retry the bubble")
