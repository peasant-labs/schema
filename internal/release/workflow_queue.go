package release

import "fmt"

// QueueRule names one merge-queue workflow rule the shared checker applies to
// every repo. The set is closed: adding a rule is a contract change, so callers
// compare against these constants rather than bare strings and diagnostics
// carry the member that fired.
type QueueRule string

const (
	// QueueRuleRequiredContexts pins every required check the merge queue waits
	// on: each entry names a workflow file in the workflows directory and the
	// job inside it that reports the context. The checker returns the list for
	// the operator to compare against the queue ruleset.
	QueueRuleRequiredContexts QueueRule = "required-contexts"
	// QueueRuleMergeGroup requires every workflow that owns a required context
	// to trigger on merge_group as well as on pull_request.
	QueueRuleMergeGroup QueueRule = "merge-group-on-required-workflows"
	// QueueRuleConcurrency requires every workflow that owns a required context
	// to key its concurrency group on the ref or the run id, and to cancel
	// in-progress runs only for pull_request. A group keyed only on the PR
	// number is empty on a merge group, so every group build would share one
	// group and cancel all but one pending run.
	QueueRuleConcurrency QueueRule = "required-workflows-concurrency-per-ref"
	// QueueRuleReleaseActions gates every step of the release job that can
	// write on the action the tool printed in the same run, in the canonical
	// verdict form. A tool output is not a gate by itself: a needs edge is
	// satisfied by any job that succeeds and an exit 0 in one step does not
	// stop the next step.
	QueueRuleReleaseActions QueueRule = "release-actions-gated"
	// QueueRuleTitleGuard keeps the title guard honest: the guard job carries
	// no job-level if and no continue-on-error, and the guard workflow never
	// reads the pull-request action or the changed-file list. A run that ends
	// green without checking would replace a red run, and a body edit does not
	// change the head commit.
	QueueRuleTitleGuard QueueRule = "title-guard-no-early-success"
	// QueueRuleRangeFloor requires every workflow step that reads a commit
	// range (plan, release-pr, freeze-plan, check-contract --until) to pass
	// --floor from the RELEASE_BATCH_FLOOR repository variable, so a subject
	// the tool cannot parse leaves parsing everywhere at once.
	QueueRuleRangeFloor QueueRule = "range-readers-take-the-floor"
	// QueueRuleBotWriter requires each bot branch to have exactly one writer
	// job: the release branch belongs to the release job and every other bot
	// branch to its own bot job, so two jobs can never push the same branch.
	QueueRuleBotWriter QueueRule = "bot-branch-one-writer"
	// QueueRuleTokenMint requires every job that mints the releaser App token
	// to check that github.ref is the default branch before it mints. A
	// dispatch on any branch would otherwise run that branch's workflow file
	// with the App key.
	QueueRuleTokenMint QueueRule = "token-mint-ref-check"
)

// AllQueueRules is the canonical ordered closed set, used by generators and
// boundary validators.
var AllQueueRules = []QueueRule{
	QueueRuleRequiredContexts,
	QueueRuleMergeGroup,
	QueueRuleConcurrency,
	QueueRuleReleaseActions,
	QueueRuleTitleGuard,
	QueueRuleRangeFloor,
	QueueRuleBotWriter,
	QueueRuleTokenMint,
}

// String renders the rule for diagnostic messages and CLI output.
func (r QueueRule) String() string { return string(r) }

// IsValid reports whether the rule is one of the known members.
func (r QueueRule) IsValid() bool {
	for _, known := range AllQueueRules {
		if r == known {
			return true
		}
	}
	return false
}

// Diagnostic is one merge-queue policy violation. Rule names the closed rule
// that fired; Workflow names the file carrying the fault (empty for a fault in
// the policy itself); Job names the offending job when the fault is scoped to
// one; Message states what is wrong, where, why it matters, and how to fix it.
type Diagnostic struct {
	Rule     QueueRule
	Workflow string
	Job      string
	Message  string
}

// RequiredContext is one required check the merge queue waits on: the context
// name the ruleset requires, the workflow file that reports it, and the job
// inside that file. Aggregator marks the job that always reports so a
// docs-only change is never stuck in the queue; Guard marks the title-guard
// job that must run the full check on every event with no early exit.
type RequiredContext struct {
	Context    string `yaml:"context"`
	Workflow   string `yaml:"workflow"`
	Job        string `yaml:"job"`
	Aggregator bool   `yaml:"aggregator"`
	Guard      bool   `yaml:"guard"`
}

// ReleaseJobRule locates the release job: the workflow file and job that carry
// out the one action the planner prints, the head branch of the standing
// release PR (named here and nowhere else), and the ordered pre-tag step ids
// the canonical verdict reads.
type ReleaseJobRule struct {
	Workflow    string   `yaml:"workflow"`
	Job         string   `yaml:"job"`
	Branch      string   `yaml:"branch"`
	PreTagSteps []string `yaml:"pre_tag_steps"`
}

// BotJobRule locates one bot job that may push a branch: the workflow file and
// job, and the branch that is its own (no other job pushes it).
type BotJobRule struct {
	Workflow string `yaml:"workflow"`
	Job      string `yaml:"job"`
	Branch   string `yaml:"branch"`
}

// QueueSection is the merge-queue half of a repo's release-guard policy. It is
// optional: a policy without it constrains only the tag-triggered release
// workflow's job graph, exactly as before. When present, Required must be
// non-empty; ReleaseJob is absent only in a repo with no release job, and
// BotJobs lists every other job that may push a branch.
type QueueSection struct {
	Required   []RequiredContext `yaml:"required"`
	ReleaseJob *ReleaseJobRule   `yaml:"release_job"`
	BotJobs    []BotJobRule      `yaml:"bot_jobs"`
}

// CheckWorkflows validates every workflow file in dir against the policy's
// queue section and returns the violations plus the required contexts in
// policy order (for the operator to compare against the queue ruleset). A
// policy without a queue section constrains nothing here and yields no
// diagnostics, so today's jobs-only policies keep passing unchanged.
//
// A non-empty diagnostic list means the directory is rejected; an empty one
// means it satisfies every rule. Directory-level and file-level faults that
// keep the checker from reading what a rule needs are reported as diagnostics
// under the first rule that needs the file, so one run names every broken
// file instead of stopping at the first.
func CheckWorkflows(policy WorkflowPolicy, dir string) ([]Diagnostic, []RequiredContext) {
	if policy.Queue == nil {
		return nil, nil
	}
	return []Diagnostic{{
		Rule:     QueueRuleRequiredContexts,
		Workflow: dir,
		Message:  "check workflows: " + dir + ": rule required-contexts: the merge-queue rule checks are not wired yet, so no workflow directory can be accepted during merge-queue policy validation. Wait for the rule implementation before enforcing a queue section",
	}}, append([]RequiredContext(nil), policy.Queue.Required...)
}

// validateQueueSection rejects a malformed queue section with an actionable
// error naming the entry and the fix. Semantic checks (a guard entry exists,
// one writer per bot branch, the workflow shapes) belong to the rule checker,
// which reports them as diagnostics; this only guards the structure a typo'd
// policy would otherwise silently disable.
func validateQueueSection(path string, queue *QueueSection) error {
	prefix := "load workflow policy: " + path + " queue"
	if len(queue.Required) == 0 {
		return fmt.Errorf("%s declares no required contexts during release workflow validation. A queue section must list at least the title-guard context under 'required:'; remove the empty 'queue:' key to keep a jobs-only policy", prefix)
	}
	seen := make(map[string]bool, len(queue.Required))
	for i, req := range queue.Required {
		where := fmt.Sprintf("%s required[%d]", prefix, i)
		if req.Context == "" {
			return fmt.Errorf("%s has an empty context during release workflow validation. Every required entry must set 'context:' to the check name the merge queue waits on", where)
		}
		if req.Workflow == "" {
			return fmt.Errorf("%s (context %q) has an empty workflow during release workflow validation. Set 'workflow:' to the file in .github/workflows that reports it", where, req.Context)
		}
		if req.Job == "" {
			return fmt.Errorf("%s (context %q) has an empty job during release workflow validation. Set 'job:' to the job in that file that reports the context", where, req.Context)
		}
		if seen[req.Context] {
			return fmt.Errorf("%s declares context %q twice during release workflow validation. Each required context must appear once so the printed list compares exactly with the queue ruleset", where, req.Context)
		}
		seen[req.Context] = true
	}
	if queue.ReleaseJob != nil {
		where := prefix + " release_job"
		if queue.ReleaseJob.Workflow == "" {
			return fmt.Errorf("%s has an empty workflow during release workflow validation. Set 'workflow:' to the release workflow file", where)
		}
		if queue.ReleaseJob.Job == "" {
			return fmt.Errorf("%s has an empty job during release workflow validation. Set 'job:' to the release job that carries out the planner's action", where)
		}
		if queue.ReleaseJob.Branch == "" {
			return fmt.Errorf("%s has an empty branch during release workflow validation. Set 'branch:' to the release PR's head branch, named here and nowhere else", where)
		}
		for i, step := range queue.ReleaseJob.PreTagSteps {
			if step == "" {
				return fmt.Errorf("%s pre_tag_steps[%d] is empty during release workflow validation. Name the pre-tag step ids in verdict order, or drop the key when the repo has no pre-tag steps", where, i)
			}
		}
	}
	for i, bot := range queue.BotJobs {
		where := fmt.Sprintf("%s bot_jobs[%d]", prefix, i)
		if bot.Workflow == "" {
			return fmt.Errorf("%s has an empty workflow during release workflow validation. Set 'workflow:' to the file holding the bot job", where)
		}
		if bot.Job == "" {
			return fmt.Errorf("%s has an empty job during release workflow validation. Set 'job:' to the bot job that pushes the branch", where)
		}
		if bot.Branch == "" {
			return fmt.Errorf("%s has an empty branch during release workflow validation. Set 'branch:' to the branch only this job pushes", where)
		}
	}
	return nil
}
