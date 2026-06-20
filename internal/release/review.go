package release

// ReviewState is the `.state` field returned by the GitHub pull-request reviews
// API for a submitted review.
type ReviewState string

const (
	ReviewStateApproved         ReviewState = "APPROVED"
	ReviewStateCommented        ReviewState = "COMMENTED"
	ReviewStateChangesRequested ReviewState = "CHANGES_REQUESTED"
	ReviewStateDismissed        ReviewState = "DISMISSED"
)

// String renders the review state for CLI output and error messages.
func (s ReviewState) String() string { return string(s) }

// ReviewUser is the GitHub user shape embedded in a pull-request review.
type ReviewUser struct {
	Login string `json:"login"`
}

// Review is the subset of a GitHub pull-request review needed by the release
// approval gate.
type Review struct {
	User  *ReviewUser `json:"user"`
	State ReviewState `json:"state"`
}

// LatestApprovers returns logins whose latest non-COMMENTED review is APPROVED.
// Plain comments do not withdraw or shadow a standing approval.
//
// The "latest" review is determined by input slice order: the loop below is
// intentionally last-write-wins per reviewer. The release-guard caller feeds
// this from GitHub's pull-request reviews API, which returns reviews in
// ascending chronological/id order, and that order is preserved through
// `gh api --paginate --slurp` and parseReviews flattening. This id/input order
// is preferred over sorting by submitted_at because GitHub IDs are stable and
// order-preserving even when multiple reviews have same-second submitted_at
// timestamps.
func LatestApprovers(reviews []Review) []string {
	latest := make(map[string]ReviewState)
	order := make([]string, 0, len(reviews))

	for _, review := range reviews {
		if review.State == ReviewStateCommented || review.User == nil || review.User.Login == "" {
			continue
		}
		login := review.User.Login
		if _, seen := latest[login]; !seen {
			order = append(order, login)
		}
		latest[login] = review.State
	}

	approvers := make([]string, 0, len(latest))
	for _, login := range order {
		if latest[login] == ReviewStateApproved {
			approvers = append(approvers, login)
		}
	}
	return approvers
}
