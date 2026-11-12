package release

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// This file holds the last four queue rules: the title guard cannot pass
// early, range readers take the floor, each bot branch has exactly one writer,
// and token-minting jobs check the ref first.

// checkTitleGuard (title-guard-no-early-success) fails the policy when no
// required entry is marked as the guard, and otherwise requires the guard
// workflow's guard job to carry no job-level if and no continue-on-error. The
// guard workflow must never read the pull-request action or the changed-file
// list: a body edit does not change the head commit, so a run that ends green
// without checking would replace a red run. The event name may be tested in a
// step-level if and in the concurrency expression (to scope cancellation to
// pull_request); anywhere else, and any read of the event file, is rejected.
func (c *queueChecker) checkTitleGuard() {
	var guards []RequiredContext
	for _, req := range c.policy.Queue.Required {
		if req.Guard {
			guards = append(guards, req)
		}
	}
	if len(guards) == 0 {
		c.emit(QueueRuleTitleGuard, "", "",
			"the policy names no guard entry during merge-queue policy validation. Mark the title-guard required context with guard: true so one workflow runs the full check on every event with no early exit")
		return
	}
	for _, req := range guards {
		doc := c.needFile(QueueRuleTitleGuard, req.Workflow)
		if doc == nil {
			continue
		}
		job := jobNode(doc, req.Job)
		if job == nil {
			continue
		}
		if cond := mappingValue(job, "if"); cond != nil {
			c.emit(QueueRuleTitleGuard, req.Workflow, req.Job,
				"the guard job has a job-level if during merge-queue policy validation. Drop it so the full check runs on every event, including an edited PR body that changes no commit")
		}
		if cont := mappingValue(job, "continue-on-error"); cont != nil && cont.Value == "true" {
			c.emit(QueueRuleTitleGuard, req.Workflow, req.Job,
				"the guard job uses continue-on-error during merge-queue policy validation. Drop it so a failed check stays red instead of passing the queue")
		}
		for _, scalar := range fileScalars(doc) {
			if strings.Contains(scalar, "github.event.action") || strings.Contains(scalar, "github.event.changes") {
				c.emit(QueueRuleTitleGuard, req.Workflow, req.Job,
					"the guard workflow reads "+eventRead(scalar)+" during merge-queue policy validation. Never read the action or the changed-file list here: a body edit changes no commit, so a skipped run would replace a red check with green; check the title from git on every event")
				break
			}
			if strings.Contains(scalar, "GITHUB_EVENT_PATH") {
				c.emit(QueueRuleTitleGuard, req.Workflow, req.Job,
					"the guard workflow reads the event file during merge-queue policy validation. Read titles and SHAs from git through the tool's arguments instead, so a re-run of an old event cannot replace a newer result")
				break
			}
		}
		for _, name := range eventNameMisuse(doc) {
			c.emit(QueueRuleTitleGuard, req.Workflow, req.Job,
				"the guard workflow tests github.event_name in "+name+" during merge-queue policy validation. Test the event name only in a step-level if or in the concurrency expression")
		}
	}
}

// eventNameMisuse returns the locations outside a step-level if and the
// concurrency expression where the workflow tests the event name: job-level
// ifs, run blocks, and the workflow header. Step-level ifs and the
// concurrency subtree are the two allowed scopes.
func eventNameMisuse(doc *yaml.Node) []string {
	var out []string
	for _, scalar := range docScalarsExcept(doc, "jobs", "concurrency", "on") {
		if strings.Contains(scalar, "github.event_name") {
			out = append(out, "the workflow header")
			break
		}
	}
	for _, jobName := range jobNames(doc) {
		job := jobNode(doc, jobName)
		for i := 0; i+1 < len(job.Content); i += 2 {
			key, val := job.Content[i].Value, job.Content[i+1]
			switch key {
			case "concurrency", "steps":
				continue
			}
			for _, scalar := range nodeScalars(val) {
				if strings.Contains(scalar, "github.event_name") {
					out = append(out, "jobs."+jobName+"."+key)
					break
				}
			}
		}
		for _, step := range stepsOf(job) {
			for i := 0; i+1 < len(step.Content); i += 2 {
				key, val := step.Content[i].Value, step.Content[i+1]
				if key == "if" {
					continue
				}
				for _, scalar := range nodeScalars(val) {
					if strings.Contains(scalar, "github.event_name") {
						out = append(out, "a step "+key+" in jobs."+jobName)
						break
					}
				}
			}
		}
	}
	return out
}

// docScalarsExcept returns every scalar in the document outside the named
// top-level keys.
func docScalarsExcept(doc *yaml.Node, except ...string) []string {
	skip := map[string]bool{}
	for _, k := range except {
		skip[k] = true
	}
	var out []string
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if skip[doc.Content[i].Value] {
			continue
		}
		out = append(out, nodeScalars(doc.Content[i+1])...)
	}
	return out
}

// nodeScalars returns every scalar value under n.
func nodeScalars(n *yaml.Node) []string {
	var out []string
	var walk func(node *yaml.Node)
	walk = func(node *yaml.Node) {
		if node.Kind == yaml.ScalarNode {
			out = append(out, node.Value)
			return
		}
		for _, child := range node.Content {
			walk(child)
		}
	}
	walk(n)
	return out
}

func eventRead(scalar string) string {
	if strings.Contains(scalar, "github.event.action") {
		return "github.event.action"
	}
	return "github.event.changes"
}

// checkRangeFloor (range-readers-take-the-floor) requires every workflow step
// that reads a commit range (plan, release-pr, freeze-plan, check-contract
// --until) to pass --floor from the RELEASE_BATCH_FLOOR repository variable. A
// subject the tool cannot parse stops every range reader at the same commit;
// a reader without the floor would keep parsing it and keep the release
// stopped.
func (c *queueChecker) checkRangeFloor() {
	for _, name := range c.sortedNames() {
		f := c.files[name]
		if f.missing || f.fault != "" {
			if !f.reported && f.fault != "" {
				f.reported = true
				c.emit(QueueRuleRangeFloor, name, "", f.fault)
			}
			continue
		}
		for _, jobName := range jobNames(f.doc) {
			for _, step := range stepsOf(jobNode(f.doc, jobName)) {
				run := stepRunText(step)
				readsRange := rangeReader.MatchString(run) ||
					(strings.Contains(run, "check-contract") && strings.Contains(run, "--until"))
				if !readsRange {
					continue
				}
				if !strings.Contains(run, "--floor") || !strings.Contains(run, "RELEASE_BATCH_FLOOR") {
					c.emit(QueueRuleRangeFloor, name, jobName,
						"a step reads a commit range without the floor during merge-queue policy validation. Pass --floor ${{ vars.RELEASE_BATCH_FLOOR }} so an unparsable subject leaves parsing everywhere at once")
				}
			}
		}
	}
}

// checkBotWriter (bot-branch-one-writer) requires each bot branch to have
// exactly one writer job: the release branch belongs to the release job and
// every other bot branch to its own bot job.
func (c *queueChecker) checkBotWriter() {
	owners := map[string]string{}
	if rel := c.policy.Queue.ReleaseJob; rel != nil {
		owners[rel.Branch] = "the release job"
	}
	for _, bot := range c.policy.Queue.BotJobs {
		if owner, dup := owners[bot.Branch]; dup {
			c.emit(QueueRuleBotWriter, "", bot.Job,
				"bot branch "+quote(bot.Branch)+" is written by "+owner+" and by job "+quote(bot.Job)+" during merge-queue policy validation. Give each bot branch exactly one writer job so two jobs can never push the same branch")
			continue
		}
		owners[bot.Branch] = "job " + quote(bot.Job)
	}
}

// checkTokenMint (token-mint-ref-check) requires every job that mints the
// releaser App token, in the release workflow and in every bot workflow, to
// check that github.ref is the default branch before it mints. A dispatch on
// any branch would otherwise run that branch's workflow file with the App key.
func (c *queueChecker) checkTokenMint() {
	type named struct {
		workflow, job string
	}
	var jobs []named
	if rel := c.policy.Queue.ReleaseJob; rel != nil {
		jobs = append(jobs, named{rel.Workflow, rel.Job})
	}
	for _, bot := range c.policy.Queue.BotJobs {
		jobs = append(jobs, named{bot.Workflow, bot.Job})
	}
	for _, target := range jobs {
		doc := c.needFile(QueueRuleTokenMint, target.workflow)
		if doc == nil {
			continue
		}
		job := jobNode(doc, target.job)
		if job == nil {
			continue
		}
		steps := stepsOf(job)
		mint := -1
		for i, step := range steps {
			if isMintStep(step) {
				mint = i
				break
			}
		}
		if mint == -1 {
			continue
		}
		if refChecked(job, steps, mint) {
			continue
		}
		c.emit(QueueRuleTokenMint, target.workflow, target.job,
			"job "+quote(target.job)+" mints the App token without checking github.ref during merge-queue policy validation. Check that github.ref is the default branch in the job if or in a step before the mint, or a dispatch on any branch runs that branch's file with the App key")
	}
}

// refChecked reports whether the job gates the mint on the default branch: a
// job-level if naming the ref, or a step at or before the mint whose if or
// run compares the ref against a branch or the default-branch value.
func refChecked(job *yaml.Node, steps []*yaml.Node, mint int) bool {
	if cond := mappingValue(job, "if"); cond != nil && mentionsRefCheck(cond.Value) {
		return true
	}
	for i := 0; i <= mint && i < len(steps); i++ {
		if mentionsRefCheck(stepIfText(steps[i])) || mentionsRefCheck(stepRunText(steps[i])) {
			return true
		}
	}
	return false
}

func mentionsRefCheck(s string) bool {
	return strings.Contains(s, "github.ref") &&
		(strings.Contains(s, "refs/heads/") || strings.Contains(s, "default_branch"))
}
