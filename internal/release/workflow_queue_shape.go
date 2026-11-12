package release

import (
	"regexp"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
	"gopkg.in/yaml.v3"
)

// This file holds the workflow-shape rules every repo shares: required
// workflows run unfiltered, the aggregator always reports, titles are never
// interpolated, the cut-release job is dispatch-only with its tag-mode writes,
// the tool checkout is a pinned SHA, and workflows call classify instead of
// comparing a bot branch literal.

// checkUnfiltered (required-workflows-unfiltered) requires every workflow that
// owns a required context to run on every pull_request and merge_group event:
// no paths, paths-ignore, or branches-ignore filter that could skip the check
// the queue waits on and stick the queue.
func (c *queueChecker) checkUnfiltered() {
	for _, req := range c.policy.Queue.Required {
		doc := c.needFile(QueueRuleUnfiltered, req.Workflow)
		if doc == nil {
			continue
		}
		on := mappingValue(doc, "on")
		if on == nil || on.Kind != yaml.MappingNode {
			continue
		}
		for _, event := range []string{"pull_request", "merge_group"} {
			trigger := mappingValue(on, event)
			if trigger == nil || trigger.Kind != yaml.MappingNode {
				continue
			}
			for _, key := range []string{"paths", "paths-ignore", "branches-ignore"} {
				if mappingValue(trigger, key) != nil {
					c.emit(QueueRuleUnfiltered, req.Workflow, req.Job,
						"the "+event+" trigger filters with "+key+" during merge-queue policy validation, but the workflow owns required context "+req.Context+". Run on every event with no path filter, or a change that skips the check leaves the queue waiting forever")
				}
			}
		}
	}
}

// checkMergeGate (merge-gate-always-reports) requires the aggregator job that
// always reports so a change touching no heavy path is never stuck in the
// queue: the required entry marked aggregator has if: always() and a needs
// list.
func (c *queueChecker) checkMergeGate() {
	for _, req := range c.policy.Queue.Required {
		if !req.Aggregator {
			continue
		}
		doc := c.needFile(QueueRuleMergeGate, req.Workflow)
		if doc == nil {
			continue
		}
		job := jobNode(doc, req.Job)
		if job == nil {
			continue
		}
		if cond := mappingValue(job, "if"); cond == nil || !isAlways(cond.Value) {
			c.emit(QueueRuleMergeGate, req.Workflow, req.Job,
				"the aggregator job has no if: always() during merge-queue policy validation. Report always so a change that skips every needed job still satisfies the queue")
		}
		needs, err := needsList(job)
		if err != nil || len(needs) == 0 {
			c.emit(QueueRuleMergeGate, req.Workflow, req.Job,
				"the aggregator job names no needs during merge-queue policy validation. List the jobs it gates so a failure upstream blocks the queue instead of passing silently")
		}
	}
}

// isAlways reports whether an if is exactly the always() status function.
func isAlways(cond string) bool {
	cond = strings.TrimSpace(cond)
	return cond == "always()" || cond == "${{ always() }}"
}

var titleInterpolation = []string{
	"pull_request.title",
	"head_commit",
	"event.commits",
}

// checkTitles (titles-never-interpolated) requires that no run or env block
// interpolates a PR title or a commit message expression. Titles are
// untrusted input: interpolating one into a shell command runs attacker
// text, and interpolating one into env smuggles it into every later run.
func (c *queueChecker) checkTitles() {
	for _, name := range c.sortedNames() {
		f := c.files[name]
		if f.missing || f.fault != "" {
			if !f.reported && f.fault != "" {
				f.reported = true
				c.emit(QueueRuleTitles, name, "", f.fault)
			}
			continue
		}
		for _, jobName := range jobNames(f.doc) {
			for _, step := range stepsOf(jobNode(f.doc, jobName)) {
				for _, field := range []struct {
					key   string
					value string
				}{
					{"run", stepRunText(step)},
					{"env", envText(step)},
				} {
					for _, pattern := range titleInterpolation {
						if strings.Contains(field.value, pattern) {
							c.emit(QueueRuleTitles, name, stepID(step),
								"a step "+field.key+" interpolates "+quote(pattern)+" during merge-queue policy validation. Titles are untrusted input: pass the title through the tool's arguments from git instead of interpolating the event payload")
							break
						}
					}
				}
			}
		}
	}
}

// envText renders a step's env mapping as text for interpolation scans.
func envText(step *yaml.Node) string {
	env := mappingValue(step, "env")
	if env == nil {
		return ""
	}
	return strings.Join(nodeScalars(env), "\n")
}

// cutLevels are the only levels the cut-release button may offer, from the
// intent package: a level, never a number.
func cutLevels() []string {
	return []string{intent.LevelPatch.String(), intent.LevelMinor.String()}
}

// checkCutRelease (cut-release-is-dispatch-only) requires the cut-release job
// to trigger only on workflow_dispatch, to take its level from a choice of
// exactly patch and minor, and to write only when tag mode is on. The button
// acts as the bot on any branch otherwise, and an unconstrained level or an
// untagged write cuts a release the plan never approved.
func (c *queueChecker) checkCutRelease() {
	kind := string(intent.PRKindCutRelease)
	for _, bot := range c.policy.Queue.BotJobs {
		if string(bot.Kind) != kind {
			continue
		}
		doc := c.needFile(QueueRuleCutRelease, bot.Workflow)
		if doc == nil {
			continue
		}
		on := mappingValue(doc, "on")
		if on == nil {
			c.emit(QueueRuleCutRelease, bot.Workflow, bot.Job,
				"the cut-release workflow has no trigger during merge-queue policy validation. Trigger only on workflow_dispatch so a push can never cut a release")
			continue
		}
		if !isDispatchOnly(on) {
			c.emit(QueueRuleCutRelease, bot.Workflow, bot.Job,
				"the cut-release workflow triggers on more than workflow_dispatch during merge-queue policy validation. Trigger only on workflow_dispatch so a push can never cut a release")
		}
		c.checkCutLevel(bot.Workflow, bot.Job, on)
		job := jobNode(doc, bot.Job)
		if job == nil {
			continue
		}
		for _, step := range stepsOf(job) {
			if !isWriteStep(step) || isMintStep(step) {
				continue
			}
			// The write step itself tests tag mode: a job-level gate can be
			// edited away from the payload, so the check reads the step if.
			if hasTagModeGate(stepIfText(step)) {
				continue
			}
			c.emit(QueueRuleCutRelease, bot.Workflow, bot.Job,
				"a cut-release write step has no tag-mode gate during merge-queue policy validation. Test "+quoteTagMode()+" in the step if so the button writes nothing while tag mode is off")
		}
	}
}

// isDispatchOnly reports whether the on: trigger is exactly workflow_dispatch.
func isDispatchOnly(on *yaml.Node) bool {
	switch on.Kind {
	case yaml.ScalarNode:
		return on.Value == "workflow_dispatch"
	case yaml.MappingNode:
		if len(on.Content) != 2 {
			return false
		}
		return on.Content[0].Value == "workflow_dispatch"
	default:
		return false
	}
}

// checkCutLevel requires the cut-release level input to be a choice of exactly
// patch and minor.
func (c *queueChecker) checkCutLevel(workflow, job string, on *yaml.Node) {
	inputs := func() *yaml.Node {
		if on.Kind != yaml.MappingNode {
			return nil
		}
		dispatch := mappingValue(on, "workflow_dispatch")
		if dispatch == nil || dispatch.Kind != yaml.MappingNode {
			return nil
		}
		return mappingValue(dispatch, "inputs")
	}()
	level := func() *yaml.Node {
		if inputs == nil || inputs.Kind != yaml.MappingNode {
			return nil
		}
		return mappingValue(inputs, "level")
	}()
	var typ string
	var options []string
	if level != nil && level.Kind == yaml.MappingNode {
		if t := mappingValue(level, "type"); t != nil {
			typ = t.Value
		}
		if o := mappingValue(level, "options"); o != nil && o.Kind == yaml.SequenceNode {
			for _, item := range o.Content {
				options = append(options, item.Value)
			}
		}
	}
	if typ != "choice" || !sameStrings(options, cutLevels()) {
		c.emit(QueueRuleCutRelease, workflow, job,
			"the cut-release level input is not a choice of exactly patch and minor during merge-queue policy validation. Offer a closed level choice so the button can never cut a numbered release")
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, s := range a {
		seen[s]++
	}
	for _, s := range b {
		seen[s]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}

// toolCheckoutSHA matches a 40-hex commit.
var toolCheckoutSHA = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

// checkToolPin (tool-pin-is-a-sha) requires every checkout of the tool to name
// a 40-hex commit, and all of them in one repo to name the same commit. A tag
// or branch pin moves under the workflows and runs code no review approved.
func (c *queueChecker) checkToolPin() {
	pins := map[string][]string{}
	for _, name := range c.sortedNames() {
		f := c.files[name]
		if f.missing || f.fault != "" {
			if !f.reported && f.fault != "" {
				f.reported = true
				c.emit(QueueRuleToolPin, name, "", f.fault)
			}
			continue
		}
		for _, jobName := range jobNames(f.doc) {
			for _, step := range stepsOf(jobNode(f.doc, jobName)) {
				uses := stepUsesText(step)
				if !strings.Contains(uses, "checkout") {
					continue
				}
				with := mappingValue(step, "with")
				if with == nil {
					continue
				}
				repo := mappingValue(with, "repository")
				if repo == nil || repo.Value == "" {
					continue
				}
				ref := ""
				if r := mappingValue(with, "ref"); r != nil {
					ref = r.Value
				}
				if !toolCheckoutSHA.MatchString(ref) {
					c.emit(QueueRuleToolPin, name, stepID(step),
						"a tool checkout pins "+quote(ref)+" during merge-queue policy validation, which is not a 40-hex commit. Pin the tool checkout to the reviewed commit SHA so the workflows run exactly the code the review approved")
					continue
				}
				key := strings.ToLower(ref)
				pins[key] = append(pins[key], name+" jobs."+jobName)
			}
		}
	}
	if len(pins) > 1 {
		var where []string
		for _, locs := range pins {
			where = append(where, locs...)
		}
		c.emit(QueueRuleToolPin, "", "",
			"tool checkouts name different commits during merge-queue policy validation ("+strings.Join(where, ", ")+"). Pin every tool checkout to the same commit SHA so every workflow runs the same tool")
	}
}

// checkClassify (bot-branch-via-classify) requires workflows to call the
// release-intent classify command instead of comparing a bot branch literal:
// no if or run compares == or != against a configured bot branch, and the
// guard workflow invokes classify to tell the release PR apart. A literal in
// one workflow drifts from the config the classifier reads.
func (c *queueChecker) checkClassify() {
	var branches []string
	for _, branch := range c.cfg.BotBranches {
		if branch != "" {
			branches = append(branches, branch)
		}
	}
	literalComparison := func() *regexp.Regexp {
		if len(branches) == 0 {
			return regexp.MustCompile(`$.^`)
		}
		alts := make([]string, 0, len(branches))
		for _, b := range branches {
			alts = append(alts, regexp.QuoteMeta(b))
		}
		any := "(?:" + strings.Join(alts, "|") + ")"
		return regexp.MustCompile(`(==|!=)\s*['"]?` + any + `|['"]?` + any + `['"\s]\s*(==|!=)`)
	}()
	for _, name := range c.sortedNames() {
		f := c.files[name]
		if f.missing || f.fault != "" {
			if !f.reported && f.fault != "" {
				f.reported = true
				c.emit(QueueRuleClassify, name, "", f.fault)
			}
			continue
		}
		for _, jobName := range jobNames(f.doc) {
			for _, step := range stepsOf(jobNode(f.doc, jobName)) {
				for _, field := range []struct {
					key   string
					value string
				}{
					{"if", stepIfText(step)},
					{"run", stepRunText(step)},
				} {
					if literalComparison.MatchString(field.value) {
						c.emit(QueueRuleClassify, name, stepID(step),
							"a step "+field.key+" compares a bot branch literal during merge-queue policy validation. Call release-intent classify instead of comparing the branch: a literal in one workflow drifts from the release config the classifier reads")
					}
				}
			}
		}
	}
	for _, req := range c.policy.Queue.Required {
		if !req.Guard {
			continue
		}
		doc := c.needFile(QueueRuleClassify, req.Workflow)
		if doc == nil {
			continue
		}
		if !invokesClassify(doc) {
			c.emit(QueueRuleClassify, req.Workflow, req.Job,
				"the guard workflow never calls release-intent classify during merge-queue policy validation. Call classify to tell the release PR apart instead of comparing its head branch, so the protected-path exception follows the one classifier")
		}
	}
}

// invokesClassify reports whether any step runs the classify command.
func invokesClassify(doc *yaml.Node) bool {
	classify := regexp.MustCompile(`release-intent\s+classify\b`)
	for _, jobName := range jobNames(doc) {
		for _, step := range stepsOf(jobNode(doc, jobName)) {
			if classify.MatchString(stepRunText(step)) {
				return true
			}
		}
	}
	return false
}
