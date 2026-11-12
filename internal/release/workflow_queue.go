package release

import (
	"fmt"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// QueueRule names one merge-queue workflow rule the shared checker applies to
// every repo. The set is closed: adding a rule is a contract change, so callers
// compare against these constants rather than bare strings and diagnostics
// carry the member that fired.
//
// Each member names the workflow shape it pins: which workflows trigger on
// the merge queue and report unfiltered, how the release job gates every
// write on the planner's action, how the title guard stays honest, how range
// readers take the floor, how each bot branch keeps exactly one writer, and
// how the tool pin, the cut button, and the title checks stay exact.
// Required-contexts entries must resolve, and the operator compares the
// printed list with the ruleset. The token rule and the classify rule keep
// one classifier, no branch literals, and the ref checked before the App key
// exists.
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
	// QueueRuleNoPush requires every branch push to be a listed job pushing
	// its own branch: the release job pushes the release branch and each bot
	// job pushes its own kind's branch, resolved from the release config.
	// Anything else, including a second job on one bot kind, is reported.
	QueueRuleNoPush QueueRule = "no-default-branch-push-from-workflows"
	// QueueRuleTokenMint requires every job that mints the releaser App token
	// to check that github.ref is the default branch, in a job-level or
	// mint-step if, and to mint only when tag mode is on. A dispatch on any
	// branch would otherwise run that branch's workflow file with the App key.
	QueueRuleTokenMint QueueRule = "token-mint-ref-check"
	// QueueRuleUnfiltered requires every workflow that owns a required context
	// to run on every pull_request and merge_group event: no paths,
	// paths-ignore, or branches-ignore filter that could skip the check the
	// queue waits on.
	QueueRuleUnfiltered QueueRule = "required-workflows-unfiltered"
	// QueueRuleMergeGate requires the aggregator job that always reports so a
	// change touching no heavy path is never stuck in the queue: it carries
	// if: always() and a needs list.
	QueueRuleMergeGate QueueRule = "merge-gate-always-reports"
	// QueueRuleTitles requires that no run or env block interpolates a PR
	// title or a commit message expression: titles are untrusted input.
	QueueRuleTitles QueueRule = "titles-never-interpolated"
	// QueueRuleCutRelease requires the cut-release job to trigger only on
	// workflow_dispatch, to take its level from a choice of exactly patch and
	// minor, and to write only when tag mode is on.
	QueueRuleCutRelease QueueRule = "cut-release-is-dispatch-only"
	// QueueRuleToolPin requires every checkout of the tool to name a 40-hex
	// commit, and all of them in one repo to name the same commit.
	QueueRuleToolPin QueueRule = "tool-pin-is-a-sha"
	// QueueRuleClassify requires workflows to call the release-intent
	// classify command instead of comparing a bot branch literal: no if or
	// run compares == or != against a configured bot branch, and the guard
	// workflow invokes classify to tell the release PR apart.
	QueueRuleClassify QueueRule = "bot-branch-via-classify"
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
	QueueRuleNoPush,
	QueueRuleTokenMint,
	QueueRuleUnfiltered,
	QueueRuleMergeGate,
	QueueRuleTitles,
	QueueRuleCutRelease,
	QueueRuleToolPin,
	QueueRuleClassify,
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
// out the one action the planner prints, and the ordered pre-tag step ids the
// canonical verdict reads. The release PR's head branch is not named here: it
// comes from the release config's bot_branches release entry, the one source,
// resolved by the checker.
type ReleaseJobRule struct {
	Workflow    string   `yaml:"workflow"`
	Job         string   `yaml:"job"`
	PreTagSteps []string `yaml:"pre_tag_steps"`
}

// BotJobRule locates one bot job that may push a branch: the workflow file,
// the job, and the bot kind whose branch it pushes. The branch itself is not
// named here: it comes from the release config's bot_branches entry for that
// kind, the one source, resolved by the checker.
type BotJobRule struct {
	Workflow string        `yaml:"workflow"`
	Job      string        `yaml:"job"`
	Kind     intent.PRKind `yaml:"kind"`
}

// botKinds are the bot PR kinds a policy bot job may claim. The release kind
// belongs to the release job, never to a bot entry; ordinary is not a bot PR.
func botKindValid(kind intent.PRKind) bool {
	return kind == intent.PRKindCutRelease || kind == intent.PRKindVendorHash
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
			return fmt.Errorf("%s has an empty job during release workflow validation. Set 'job:' to the bot job that pushes its kind's branch", where)
		}
		if !botKindValid(bot.Kind) {
			return fmt.Errorf("%s names kind %q during release workflow validation. A bot job claims one of the bot PR kinds, cut-release or vendor-hash, so the checker can resolve its branch from the release config; the release kind belongs to the release job", where, bot.Kind.String())
		}
	}
	return nil
}
