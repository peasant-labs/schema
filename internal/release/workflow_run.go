package release

// WorkflowRunStatus is the `.status` field of a GitHub Actions workflow run.
// Typed per the repo's no-stringly-typed rule so the release-run predicate
// compares against named constants, not bare strings. The wrapper at the
// cmd/release-guard go-github seam maps the upstream *github.WorkflowRun.Status
// onto these values; internal/release stays free of any go-github import.
type WorkflowRunStatus string

const (
	// WorkflowRunCompleted is the only status a green run can have — the run has
	// finished and its Conclusion is meaningful.
	WorkflowRunCompleted WorkflowRunStatus = "completed"
	// WorkflowRunInProgress is a run that is still executing (no conclusion yet).
	WorkflowRunInProgress WorkflowRunStatus = "in_progress"
	// WorkflowRunQueued is a run accepted but not yet started.
	WorkflowRunQueued WorkflowRunStatus = "queued"
	// WorkflowRunRequested is a run requested but not yet queued.
	WorkflowRunRequested WorkflowRunStatus = "requested"
	// WorkflowRunWaiting is a run paused waiting on a deployment gate/approval.
	WorkflowRunWaiting WorkflowRunStatus = "waiting"
	// WorkflowRunPending is a run pending (concurrency or other hold).
	WorkflowRunPending WorkflowRunStatus = "pending"
)

// String renders the workflow-run status for CLI output and error messages.
func (s WorkflowRunStatus) String() string { return string(s) }

// WorkflowRunConclusion is the `.conclusion` field of a GitHub Actions workflow
// run — meaningful only once Status is WorkflowRunCompleted. Typed for the same
// reason as WorkflowRunStatus.
type WorkflowRunConclusion string

const (
	// WorkflowRunSuccess is the only conclusion a green run can have.
	WorkflowRunSuccess WorkflowRunConclusion = "success"
	// WorkflowRunFailure is a completed run that failed.
	WorkflowRunFailure WorkflowRunConclusion = "failure"
	// WorkflowRunNeutral is a completed run that neither passed nor failed.
	WorkflowRunNeutral WorkflowRunConclusion = "neutral"
	// WorkflowRunCancelled is a completed run that was cancelled.
	WorkflowRunCancelled WorkflowRunConclusion = "cancelled"
	// WorkflowRunSkipped is a completed run that was skipped.
	WorkflowRunSkipped WorkflowRunConclusion = "skipped"
	// WorkflowRunTimedOut is a completed run that exceeded its time limit.
	WorkflowRunTimedOut WorkflowRunConclusion = "timed_out"
	// WorkflowRunActionRequired is a completed run that needs manual action.
	WorkflowRunActionRequired WorkflowRunConclusion = "action_required"
	// WorkflowRunStale is a completed run whose result is considered stale.
	WorkflowRunStale WorkflowRunConclusion = "stale"
	// WorkflowRunNoConclusion is the empty conclusion of a not-yet-completed run.
	WorkflowRunNoConclusion WorkflowRunConclusion = ""
)

// String renders the workflow-run conclusion for CLI output and error messages.
func (c WorkflowRunConclusion) String() string { return string(c) }

// WorkflowRun is the minimal projection of a GitHub Actions workflow run the
// release-final gate consumes: the commit it ran against and its terminal
// state. It is an own-type (mirroring release.Review and
// release.CollaboratorPermission) so the policy predicate never touches a
// *github.WorkflowRun and the go-github pointer-field nil-guards stay at the
// wrapper boundary.
type WorkflowRun struct {
	HeadSHA    string
	Status     WorkflowRunStatus
	Conclusion WorkflowRunConclusion
}

// RunGreenForCommit reports whether runs contains a completed, successful run of
// the workflow for commitSHA. It is the pure, table-testable predicate lifted
// out of the old runGreen's inline scan: a tag run's HeadSHA is the tagged
// commit, so a later success on the same commit is also accepted. Matching on
// the commit SHA (rather than the latest run) keeps the rule independent of run
// ordering or the now-dropped 100-run client-side cap.
func RunGreenForCommit(runs []WorkflowRun, commitSHA string) bool {
	for _, r := range runs {
		if r.HeadSHA == commitSHA && r.Status == WorkflowRunCompleted && r.Conclusion == WorkflowRunSuccess {
			return true
		}
	}
	return false
}
