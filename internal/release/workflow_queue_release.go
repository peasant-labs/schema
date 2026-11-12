package release

import (
	"regexp"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
	"gopkg.in/yaml.v3"
)

// This file is the release-actions-gated rule: every step of the release job
// that can write runs only under an `if:` on the one action the planner
// printed in the same run, the tag goes out once on the verdict's tag action,
// the verdict carries every pre-tag step in the canonical form, each pre-tag
// step is a single gated command, later steps carry the status function so a
// failed pre-tag step rebuilds instead of skipping, writes test the push the
// write step printed, nothing that writes uses continue-on-error, and the arm
// and hold steps test the gate. A tool output is not a gate by itself: a
// needs edge is satisfied by any job that succeeds, and an exit 0 in one step
// does not stop the next step.

var (
	// actionGate matches an if that compares one plan step's action output
	// with == against a literal, e.g. steps.plan.outputs.action == 'open-pr'.
	actionGate = regexp.MustCompile(`steps\.([A-Za-z0-9_-]+)\.outputs\.action\s*==\s*'([^']+)'`)
	// gateTest matches an if that compares the queue-gate output, e.g.
	// steps.gate.outputs.gate == 'arm'. The alternation is built from the
	// planner's closed gate set so a new gate reaches this check with the set.
	gateTest = regexp.MustCompile(`steps\.([A-Za-z0-9_-]+)\.outputs\.gate\s*==\s*'` + gateAlternatives() + `'`)
	// pushTest matches an if that tests the push the write step printed, e.g.
	// steps.write.outputs.push == 'true'.
	pushTest = regexp.MustCompile(`steps\.([A-Za-z0-9_-]+)\.outputs\.push\b`)
	// verdictArg extracts the --verdict argument: the canonical ${{ }} form
	// or the bare taggable literal for a repo with no pre-tag steps.
	verdictArg = regexp.MustCompile(`--verdict\s+(\$\{\{.*?\}\}|\S+)`)
	// planRunner matches a step that runs the planner, e.g.
	// release-intent plan --branch develop (also behind a dev-shell wrapper).
	planRunner = regexp.MustCompile(`release-intent\s+plan\b`)
	// rangeReader matches a step that reads a commit range: plan, release-pr,
	// or freeze-plan.
	rangeReader = regexp.MustCompile(`release-intent\s+(plan|release-pr|freeze-plan)\b`)
	// prWrite matches a run line that opens, edits, closes, merges, or
	// releases a PR, or pushes a ref.
	prWrite = regexp.MustCompile(`git\s+push\b|\bgh\s+pr\s+(create|edit|close|merge)\b|\bgh\s+release\s+create\b`)
	// tagPush matches a run line that pushes a tag: a refs/tags refspec, an
	// explicit tag, or a release creation.
	tagPush = regexp.MustCompile(`refs/tags/|\bgit\s+tag\b|\bgh\s+release\s+create\b`)
)

// gateAlternatives renders the planner's closed gate set as a regex
// alternation, so the arm/hold check knows exactly the gates the tool prints.
func gateAlternatives() string {
	alts := make([]string, 0, len(intent.AllGates))
	for _, g := range intent.AllGates {
		alts = append(alts, regexp.QuoteMeta(g.String()))
	}
	return "(" + strings.Join(alts, "|") + ")"
}

// closedActions is the planner's closed action set: the only values a write
// step's action gate may compare against. It is built from the intent package
// so the two sets are one object, not two copies.
var closedActions = func() map[string]bool {
	set := make(map[string]bool, len(intent.AllActions))
	for _, a := range intent.AllActions {
		set[a.String()] = true
	}
	return set
}()

// gateValues renders the planner's closed gate set for messages.
func gateValues() []string {
	vals := make([]string, 0, len(intent.AllGates))
	for _, g := range intent.AllGates {
		vals = append(vals, g.String())
	}
	return vals
}

// checkReleaseActions applies release-actions-gated to the release job. A
// policy with no release job (a repo with no release PR) skips the rule.
func (c *queueChecker) checkReleaseActions() {
	rel := c.policy.Queue.ReleaseJob
	if rel == nil {
		return
	}
	doc := c.needFile(QueueRuleReleaseActions, rel.Workflow)
	if doc == nil {
		return
	}
	job := jobNode(doc, rel.Job)
	if job == nil {
		c.emit(QueueRuleReleaseActions, rel.Workflow, rel.Job,
			"jobs."+rel.Job+" is missing during merge-queue policy validation, but the policy names it as the release job. Add the job that carries out the planner's action")
		return
	}
	r := &releaseChecker{c: c, workflow: rel.Workflow, doc: doc, job: job, preTag: rel.PreTagSteps}
	r.collectSteps()
	r.checkVerdictForm()
	r.checkWriteStepsGated()
	r.checkMintGated()
	r.checkTagPush()
	r.checkPreTagSteps()
	r.checkStatusFunctions()
	r.checkPushOutputs()
	r.checkContinueOnError()
	r.checkArmHold()
}

// releaseChecker holds one release-actions-gated evaluation: the release job's
// steps in order with their ids, and the step indexes each clause flags.
type releaseChecker struct {
	c        *queueChecker
	workflow string
	doc      *yaml.Node
	job      *yaml.Node
	preTag   []string
	steps    []*yaml.Node
	ids      map[string]int
	flagged  map[int]bool
}

// collectSteps indexes the release job's steps in order. A step without an id
// still participates (by position); clauses that need a named step report the
// missing id.
func (r *releaseChecker) collectSteps() {
	r.steps = stepsOf(r.job)
	r.ids = map[string]int{}
	r.flagged = map[int]bool{}
	for i, step := range r.steps {
		if id := stepID(step); id != "" {
			r.ids[id] = i
		}
	}
}

// stepIf returns the step's if text, or "" when it has none.
func stepIfText(step *yaml.Node) string {
	if cond := mappingValue(step, "if"); cond != nil {
		return cond.Value
	}
	return ""
}

// stepRunText returns the step's run text, or "" for a uses-only step.
func stepRunText(step *yaml.Node) string {
	if run := mappingValue(step, "run"); run != nil {
		return run.Value
	}
	return ""
}

// stepUsesText returns the step's uses target, or "" for a run step.
func stepUsesText(step *yaml.Node) string {
	if uses := mappingValue(step, "uses"); uses != nil {
		return uses.Value
	}
	return ""
}

// stepScalars returns every scalar value under the step, for content scans
// (token-output references, dequeue calls, event reads).
func stepScalars(step *yaml.Node) []string {
	var out []string
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		if n.Kind == yaml.ScalarNode {
			out = append(out, n.Value)
			return
		}
		for _, child := range n.Content {
			walk(child)
		}
	}
	walk(step)
	return out
}

// isMintStep reports whether the step mints a GitHub App token.
func isMintStep(step *yaml.Node) bool {
	return strings.Contains(stepUsesText(step), "create-github-app-token")
}

// isWriteStep reports whether the step can write: it mints the App token, it
// consumes a minted token output, or its run pushes a ref or writes a PR.
func isWriteStep(step *yaml.Node) bool {
	if isMintStep(step) {
		return true
	}
	for _, s := range stepScalars(step) {
		if strings.Contains(s, "outputs.token") {
			return true
		}
	}
	return prWrite.MatchString(stepRunText(step))
}

// isArmHoldStep reports whether the step is gated on the queue-gate output
// rather than on the planner's action; those steps belong to clause (8) and
// are exempt from the action-gate clauses.
func isArmHoldStep(step *yaml.Node) bool {
	return gateTest.MatchString(stepIfText(step))
}

// isTagStep reports whether the step pushes a tag.
func isTagStep(step *yaml.Node) bool {
	return tagPush.MatchString(stepRunText(step))
}

// isBranchPushStep reports whether the step pushes a branch (a push that is
// not a tag push).
func isBranchPushStep(step *yaml.Node) bool {
	run := stepRunText(step)
	return strings.Contains(run, "git push") && !isTagStep(step)
}

// isPROpenEditStep reports whether the step opens or edits the release PR.
func isPROpenEditStep(step *yaml.Node) bool {
	return regexp.MustCompile(`\bgh\s+pr\s+(create|edit)\b`).MatchString(stepRunText(step))
}

// isQueueReadStep reports whether the step reads the merge queue or runs the
// queue gate.
func isQueueReadStep(step *yaml.Node) bool {
	run := stepRunText(step)
	return strings.Contains(run, "mergeQueue") || strings.Contains(run, "queue-gate")
}

// canonicalVerdict builds the exact verdict expression the policy's pre-tag
// steps require: one conjunct per step in order, each reading that step's
// outcome, and nothing else. With no pre-tag steps the verdict is the
// taggable literal. The literals come from the intent package, the same
// objects the planner prints.
func canonicalVerdict(preTag []string) string {
	taggable := string(intent.VerdictTaggable)
	if len(preTag) == 0 {
		return taggable
	}
	parts := make([]string, 0, len(preTag))
	for _, name := range preTag {
		parts = append(parts, "steps."+name+".outcome == 'success'")
	}
	return "${{ (" + strings.Join(parts, " && ") + ") && '" + taggable + "' || '" + string(intent.VerdictRejected) + "' }}"
}

// closedActionList renders the closed action set for messages.
func closedActionList() string {
	vals := make([]string, 0, len(intent.AllActions))
	for _, a := range intent.AllActions {
		vals = append(vals, a.String())
	}
	return strings.Join(vals, ", ")
}

// hasDisjunction reports whether an if spreads its gate across a disjunction:
// ||, always(), or success() alongside the comparison. A gate inside a
// disjunction opens on runs the comparison never approved.
func hasDisjunction(cond string) bool {
	return strings.Contains(cond, "||") || strings.Contains(cond, "always()") || strings.Contains(cond, "success()")
}

// checkVerdictForm (clause 3) requires every --verdict argument to be the
// canonical form: the taggable literal with no pre-tag steps, otherwise the
// exact expression built from the policy's pre-tag steps in order.
func (r *releaseChecker) checkVerdictForm() {
	want := canonicalVerdict(r.preTag)
	for i, step := range r.steps {
		if !strings.Contains(stepRunText(step), "--verdict") {
			continue
		}
		r.flagged[i] = true
		got := ""
		if m := verdictArg.FindStringSubmatch(stepRunText(step)); m != nil {
			got = strings.TrimSpace(m[1])
		}
		if got != want {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"the --verdict argument is "+quote(got)+" during merge-queue policy validation, want the canonical "+quote(want)+". A verdict that leaves out a pre-tag step or reads conclusion instead of outcome can tag a commit its checks rejected; pass the exact expression the policy's pre_tag_steps build")
		}
	}
}

// checkWriteStepsGated (clause 1) requires every step that can write to run
// only under an if that compares one step's action output with == against one
// value of the closed action set, as a single conjunction. The arm and hold
// steps test the gate instead and belong to clause (8); the token-minting
// step is gated on tag mode instead (checkMintGated), because no single action
// covers every run that needs a token.
func (r *releaseChecker) checkWriteStepsGated() {
	for i, step := range r.steps {
		if !isWriteStep(step) || isArmHoldStep(step) || isMintStep(step) {
			continue
		}
		r.flagged[i] = true
		cond := stepIfText(step)
		m := actionGate.FindStringSubmatch(cond)
		if m == nil {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"a step that can write has no action-gated if during merge-queue policy validation. A needs edge is satisfied by any job that succeeds and an exit 0 in an earlier step does not stop this one; gate the step on steps.<plan>.outputs.action == '<one of "+closedActionList()+">'")
			continue
		}
		if hasDisjunction(cond) {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"an action gate sits inside a disjunction during merge-queue policy validation. Parse the if as a conjunction: a gate with ||, always() or success() opens on runs the comparison never approved; gate on the one action alone")
			continue
		}
		if !closedActions[m[2]] {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"the action gate compares against "+quote(m[2])+" during merge-queue policy validation, which is not one of "+closedActionList()+". Compare against the closed action the planner prints, or the gate never matches a real run")
			continue
		}
		if _, ok := r.ids[m[1]]; !ok {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"the action gate reads steps."+m[1]+" during merge-queue policy validation, but no step has that id. Point the gate at the plan step that prints the action in this run")
		}
	}
}

// checkMintGated requires the token-minting step to mint only when tag mode
// is on: its own if, or the job-level if, compares RELEASE_BATCH_MODE with ==
// against 'tag' as a single comparison. With tag mode off the action is always
// none, so the job mints no token and nothing downstream can write.
func (r *releaseChecker) checkMintGated() {
	jobCond := ""
	if cond := mappingValue(r.job, "if"); cond != nil {
		jobCond = cond.Value
	}
	for i, step := range r.steps {
		if !isMintStep(step) {
			continue
		}
		r.flagged[i] = true
		if hasTagModeGate(stepIfText(step)) || hasTagModeGate(jobCond) {
			continue
		}
		r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
			"the token-minting step has no tag-mode gate during merge-queue policy validation. Mint only when tag mode is on, with "+quoteTagMode()+" in the step if or the job if, so a tag-mode-off run mints no token")
	}
}

// tagModeGate is the canonical tag-mode comparison a mint gate must carry.
func quoteTagMode() string {
	return "vars.RELEASE_BATCH_MODE == 'tag'"
}

// hasTagModeGate reports whether an if carries the tag-mode comparison as a
// single comparison: the variable, ==, and 'tag', with no disjunction.
func hasTagModeGate(cond string) bool {
	return strings.Contains(cond, "RELEASE_BATCH_MODE") &&
		strings.Contains(cond, "==") &&
		strings.Contains(cond, "'tag'") &&
		!strings.Contains(cond, "||")
}

// checkTagPush (clause 2) requires exactly one step to push a tag, gated on
// the tag action of a plan step that passes --verdict, with no other step or
// job pushing a tag.
func (r *releaseChecker) checkTagPush() {
	for _, jobName := range jobNames(r.doc) {
		job := jobNode(r.doc, jobName)
		for _, step := range stepsOf(job) {
			if !isTagStep(step) {
				continue
			}
			r.flagged[r.stepIndex(step)] = true
			cond := stepIfText(step)
			m := actionGate.FindStringSubmatch(cond)
			if m == nil || m[2] != string(intent.ActionTag) || hasDisjunction(cond) {
				r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
					"a step that pushes a tag is not gated on the tag action during merge-queue policy validation. Gate it on steps.<plan>.outputs.action == 'tag' as a single conjunction, with no ||, always() or success(), so only the verdict's tag action pushes the tag")
				continue
			}
			idx, ok := r.ids[m[1]]
			var planRun string
			if ok {
				planRun = stepRunText(r.steps[idx])
			}
			if !ok || !strings.Contains(planRun, "--verdict") {
				r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
					"the tag step is gated on steps."+m[1]+" during merge-queue policy validation, which is not a plan step that passes --verdict. Gate the tag on the verdict's tag action so a stale candidate is never tagged")
			}
		}
	}
	tagCount := 0
	for _, jobName := range jobNames(r.doc) {
		for _, step := range stepsOf(jobNode(r.doc, jobName)) {
			if isTagStep(step) {
				tagCount++
			}
		}
	}
	if tagCount == 0 {
		r.c.emit(QueueRuleReleaseActions, r.workflow, "",
			"no step pushes a tag during merge-queue policy validation. One step must push the tag on the verdict's tag action, or the release job can never publish")
	} else if tagCount > 1 {
		r.c.emit(QueueRuleReleaseActions, r.workflow, "",
			"more than one step pushes a tag during merge-queue policy validation. Exactly one step pushes the tag so a failed run cannot double-publish; move every other tag push out of this workflow")
	}
}

// stepIndex returns the release-job step index of step, or -1 for a step in
// another job (only used for flagging release-job steps).
func (r *releaseChecker) stepIndex(step *yaml.Node) int {
	for i, s := range r.steps {
		if s == step {
			return i
		}
	}
	return -1
}

// checkPreTagSteps (clause 4) requires each policy pre-tag step to exist by
// id, to run only on the verify action, to carry the status function on every
// step after the first, and to run exactly one planner command with no
// chaining, no shell override, and no multi-line run.
func (r *releaseChecker) checkPreTagSteps() {
	for i, name := range r.preTag {
		idx, ok := r.ids[name]
		if !ok {
			r.c.emit(QueueRuleReleaseActions, r.workflow, name,
				"the policy names pre_tag_steps entry "+quote(name)+" during merge-queue policy validation, but no release-job step has that id. Add the step so the verdict reads a real outcome")
			continue
		}
		r.flagged[idx] = true
		step := r.steps[idx]
		cond := stepIfText(step)
		m := actionGate.FindStringSubmatch(cond)
		if m == nil || m[2] != string(intent.ActionVerify) {
			r.c.emit(QueueRuleReleaseActions, r.workflow, name,
				"pre-tag step "+quote(name)+" does not run only on the verify action during merge-queue policy validation. Gate it on steps.<plan>.outputs.action == 'verify' so a stale candidate never reaches the checks")
		} else if i > 0 && !strings.Contains(cond, "!cancelled()") {
			r.c.emit(QueueRuleReleaseActions, r.workflow, name,
				"pre-tag step "+quote(name)+" runs after a pre-tag step without the status function during merge-queue policy validation. Carry !cancelled() so a failed earlier check still rebuilds the release PR instead of skipping every later step")
		}
		run := stepRunText(step)
		switch {
		case strings.Contains(run, "\n"):
			r.c.emit(QueueRuleReleaseActions, r.workflow, name,
				"pre-tag step "+quote(name)+" has a multi-line run during merge-queue policy validation. Run exactly one command so the step's exit status is the check's verdict and nothing else can mask it")
		case mappingValue(step, "shell") != nil:
			r.c.emit(QueueRuleReleaseActions, r.workflow, name,
				"pre-tag step "+quote(name)+" overrides the shell during merge-queue policy validation. Drop the shell override so the step runs the one planner command with the default failure semantics")
		case !strings.Contains(run, "release-intent") && !strings.Contains(run, "schema-gen"):
			r.c.emit(QueueRuleReleaseActions, r.workflow, name,
				"pre-tag step "+quote(name)+" runs "+quote(firstLine(run))+" during merge-queue policy validation instead of a release-intent or schema-gen subcommand. Run the check itself here; a wrapper script that exits zero on a break would tag a rejected commit")
		case strings.Contains(run, "||") || strings.Contains(run, "&&") || strings.Contains(run, ";") || strings.Contains(run, "|") || strings.Contains(run, "set +e"):
			r.c.emit(QueueRuleReleaseActions, r.workflow, name,
				"pre-tag step "+quote(name)+" chains commands during merge-queue policy validation. Run exactly one command with no ||, &&, ;, pipe, or set +e so the step's exit status is the check's verdict")
		}
	}
}

// checkStatusFunctions (clause 5) requires every step after the first pre-tag
// step that runs the planner or reads an action to carry !cancelled(). Without
// the status function GitHub adds success() to every later if, so a failed
// pre-tag step would skip the rebuild instead of running it.
func (r *releaseChecker) checkStatusFunctions() {
	first := -1
	for i, step := range r.steps {
		if _, ok := r.ids[stepID(step)]; ok && isPreTag(r.preTag, stepID(step)) && (first == -1 || i < first) {
			first = i
		}
	}
	if first == -1 {
		return
	}
	for i, step := range r.steps {
		if i <= first {
			continue
		}
		if !planRunner.MatchString(stepRunText(step)) && !strings.Contains(stepIfText(step), "outputs.action") {
			continue
		}
		r.flagged[i] = true
		if !strings.Contains(stepIfText(step), "!cancelled()") {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"a step after the first pre-tag step has no status function during merge-queue policy validation. Carry !cancelled() so a failed check rebuilds the release PR instead of skipping every step after it")
		}
	}
}

// checkPushOutputs (clause 6) requires each step that pushes the release
// branch, opens or edits the release PR, or reads the merge queue to also test
// the push output the write step printed in the same run. A step with
// !cancelled() runs after any failure, so without the push test it would push
// or touch the PR when the write step never ran.
func (r *releaseChecker) checkPushOutputs() {
	write := -1
	for i, step := range r.steps {
		if strings.Contains(stepRunText(step), "release-pr") && strings.Contains(stepRunText(step), "--write") {
			write = i
			break
		}
	}
	for i, step := range r.steps {
		if !isBranchPushStep(step) && !isPROpenEditStep(step) && !isQueueReadStep(step) {
			continue
		}
		r.flagged[i] = true
		if write == -1 {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"a push, PR, or queue-read step has no release-pr --write step in this run during merge-queue policy validation. Write the release content first so these steps can test the push it printed")
			continue
		}
		m := pushTest.FindStringSubmatch(stepIfText(step))
		writeID := stepID(r.steps[write])
		if m == nil || (writeID != "" && m[1] != writeID) {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"a push, PR, or queue-read step does not test the write step's push output during merge-queue policy validation. Test steps."+displayID(writeID)+"outputs.push in the if so the step runs only when this run wrote the release content")
		}
	}
}

// checkContinueOnError (clause 7) forbids continue-on-error on every step the
// earlier clauses flag: a masked failure would make outcome lie to the
// verdict.
func (r *releaseChecker) checkContinueOnError() {
	for i, step := range r.steps {
		if !r.flagged[i] {
			continue
		}
		if cont := mappingValue(step, "continue-on-error"); cont != nil && cont.Value == "true" {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"a gated step uses continue-on-error during merge-queue policy validation. Drop it so a failed write, check, or tag fails the run and its outcome stays honest for the verdict")
		}
	}
}

// checkArmHold (clause 8) requires the auto-merge arm step to test gate ==
// 'arm' and the hold step to test gate == 'hold', both on the queue-gate
// output of the same run. On hold the job only turns auto-merge off; no step
// may call a dequeue.
func (r *releaseChecker) checkArmHold() {
	seen := map[string]bool{}
	for i, step := range r.steps {
		m := gateTest.FindStringSubmatch(stepIfText(step))
		if m == nil {
			continue
		}
		r.flagged[i] = true
		seen[m[2]] = true
		idx, ok := r.ids[m[1]]
		var gateRun string
		if ok {
			gateRun = stepRunText(r.steps[idx])
		}
		if !ok || !strings.Contains(gateRun, "queue-gate") {
			r.c.emit(QueueRuleReleaseActions, r.workflow, stepID(step),
				"the arm/hold step tests steps."+m[1]+" during merge-queue policy validation, which is not the queue-gate output of this run. Test the gate the queue-gate command printed so arming follows the hold rule")
		}
	}
	for _, want := range gateValues() {
		if !seen[want] {
			r.c.emit(QueueRuleReleaseActions, r.workflow, "",
				"no step tests gate == '"+want+"' during merge-queue policy validation. Arm auto-merge only when the gate says "+string(intent.GateArm)+", and take the arming back when it says "+string(intent.GateHold))
		}
	}
	for _, scalar := range fileScalars(r.doc) {
		if strings.Contains(scalar, "dequeuePullRequest") {
			r.c.emit(QueueRuleReleaseActions, r.workflow, "",
				"the workflow calls a dequeue during merge-queue policy validation. On hold the job only turns auto-merge off; a queued release PR is never taken out of the queue")
			break
		}
	}
}

func isPreTag(preTag []string, id string) bool {
	for _, name := range preTag {
		if name == id {
			return true
		}
	}
	return false
}

func displayID(id string) string {
	if id == "" {
		return "<the write step>."
	}
	return id + "."
}

func firstLine(s string) string {
	if i := strings.Index(s, "\n"); i >= 0 {
		return s[:i]
	}
	return s
}

// jobNames returns every job name in the workflow in mapping order.
func jobNames(doc *yaml.Node) []string {
	jobs := mappingValue(doc, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return nil
	}
	var names []string
	for i := 0; i+1 < len(jobs.Content); i += 2 {
		names = append(names, jobs.Content[i].Value)
	}
	return names
}

// jobNode returns the named job's mapping node, or nil.
func jobNode(doc *yaml.Node, name string) *yaml.Node {
	jobs := mappingValue(doc, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return nil
	}
	job := mappingValue(jobs, name)
	if job == nil || job.Kind != yaml.MappingNode {
		return nil
	}
	return job
}

// stepsOf returns the job's steps in order, or nil when it has none.
func stepsOf(job *yaml.Node) []*yaml.Node {
	if job == nil {
		return nil
	}
	steps := mappingValue(job, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return nil
	}
	var out []*yaml.Node
	for _, step := range steps.Content {
		if step.Kind == yaml.MappingNode {
			out = append(out, step)
		}
	}
	return out
}

// stepID returns the step's id, or "" for an unnamed step.
func stepID(step *yaml.Node) string {
	if id := mappingValue(step, "id"); id != nil {
		return id.Value
	}
	return ""
}

// fileScalars returns every scalar value in the document, for content scans.
func fileScalars(doc *yaml.Node) []string {
	var out []string
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		if n.Kind == yaml.ScalarNode {
			out = append(out, n.Value)
			return
		}
		for _, child := range n.Content {
			walk(child)
		}
	}
	walk(doc)
	return out
}
