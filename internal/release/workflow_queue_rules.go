package release

import (
	"regexp"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
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

// allowedPush maps one policy job to the branch it may push: the release job
// pushes the release branch and each bot job pushes its kind's branch, all
// resolved from the release config's bot_branches, the one source.
type allowedPush struct {
	workflow, job, branch string
}

// resolvePushes maps every policy job to its branch. A policy kind the config
// does not name is a policy-level diagnostic under the caller's rule: nothing
// may push for a branch that does not exist.
func (c *queueChecker) resolvePushes(rule QueueRule) []allowedPush {
	var allowed []allowedPush
	seen := map[intent.PRKind]string{}
	if rel := c.policy.Queue.ReleaseJob; rel != nil {
		branch, ok := c.cfg.BotBranches[intent.PRKindRelease]
		if !ok {
			c.emit(rule, "", rel.Job,
				"the policy names a release job during merge-queue policy validation, but the release config names no branch for the release kind. Set bot_branches release so the checker can tell the release branch apart")
		} else {
			allowed = append(allowed, allowedPush{rel.Workflow, rel.Job, branch})
		}
	}
	for _, bot := range c.policy.Queue.BotJobs {
		branch, ok := c.cfg.BotBranches[bot.Kind]
		if !ok {
			c.emit(rule, "", bot.Job,
				"the policy names bot kind "+quote(bot.Kind.String())+" for job "+quote(bot.Job)+" during merge-queue policy validation, but the release config names no branch for that kind. Add the bot_branches entry so the checker can tell the bot branch apart")
			continue
		}
		if owner, dup := seen[bot.Kind]; dup {
			c.emit(rule, "", bot.Job,
				"bot kind "+quote(bot.Kind.String())+" is written by "+owner+" and by job "+quote(bot.Job)+" during merge-queue policy validation. Give each bot kind exactly one writer job so two jobs can never push the same branch")
			continue
		}
		seen[bot.Kind] = "job " + quote(bot.Job)
		allowed = append(allowed, allowedPush{bot.Workflow, bot.Job, branch})
	}
	return allowed
}

// pushTarget returns the branch a git push refspec targets, or "" when the
// refspec names no branch (a tag, or an unresolvable expression).
func pushTarget(refspec string) (string, bool) {
	dst := refspec
	if i := strings.LastIndex(dst, ":"); i >= 0 {
		dst = dst[i+1:]
	}
	dst = strings.TrimSpace(dst)
	if dst == "" || strings.HasPrefix(dst, "refs/tags/") {
		return "", false
	}
	dst = strings.TrimPrefix(dst, "refs/heads/")
	if dst == "" || strings.ContainsAny(dst, "${} \t") {
		return "", false
	}
	return dst, true
}

// splitPushArgs splits a git push command line into its refspec operands,
// skipping flags and the remote name. GitHub expression spans (${{ }}) may
// contain spaces; they are glued back together before splitting so a dynamic
// refspec stays one operand.
func splitPushArgs(rest string) []string {
	fields := glueExpressions(strings.Fields(rest))
	var out []string
	skipValue := map[string]bool{"--repo": true, "--remote": true}
	remoteSkipped := false
	for i := 0; i < len(fields); i++ {
		f := fields[i]
		if skipValue[f] {
			i++
			continue
		}
		if strings.HasPrefix(f, "-") {
			continue
		}
		if !remoteSkipped {
			remoteSkipped = true
			continue
		}
		out = append(out, f)
	}
	return out
}

// glueExpressions joins field-split tokens back into whole ${{ }} spans.
func glueExpressions(fields []string) []string {
	var out []string
	var cur strings.Builder
	depth := 0
	flush := func() {
		if s := strings.TrimSpace(cur.String()); s != "" {
			out = append(out, s)
		}
		cur.Reset()
	}
	for _, f := range fields {
		opens := strings.Count(f, "${{")
		closes := strings.Count(f, "}}")
		if depth == 0 && opens == 0 && closes == 0 && cur.Len() == 0 {
			out = append(out, f)
			continue
		}
		if cur.Len() > 0 {
			cur.WriteByte(' ')
		}
		cur.WriteString(f)
		depth += opens - closes
		if depth <= 0 {
			depth = 0
			flush()
		}
	}
	flush()
	return out
}

// pushSite is one push found in a run block: a branch push with its target,
// a tag push, or a push the checker cannot resolve to a branch.
type pushSite struct {
	target     string
	isTag      bool
	unresolved bool
	raw        string
}

// gitPushSites returns every git push in the run text: branch targets, tag
// pushes, and pushes with no parseable refspec. One push per shell segment is
// assumed; a segment is split on ||, &&, ;, and pipes.
func gitPushSites(run string) []pushSite {
	var out []pushSite
	segments := regexp.MustCompile(`\|\||&&|[|;&]`).Split(run, -1)
	for _, segment := range segments {
		idx := strings.Index(segment, "git push")
		if idx < 0 {
			continue
		}
		refspecs := splitPushArgs(segment[idx+len("git push"):])
		if len(refspecs) == 0 {
			out = append(out, pushSite{raw: strings.TrimSpace(segment), unresolved: true})
		}
		for _, refspec := range refspecs {
			if strings.Contains(refspec, "refs/tags/") {
				out = append(out, pushSite{isTag: true, raw: refspec})
				continue
			}
			if target, ok := pushTarget(refspec); ok {
				out = append(out, pushSite{target: target, raw: refspec})
			} else {
				out = append(out, pushSite{raw: refspec, unresolved: true})
			}
		}
	}
	return out
}

// prHeadTarget extracts the --head branch of a gh pr create command, or "".
func prHeadTarget(run string) (string, bool) {
	segment := regexp.MustCompile(`\bgh\s+pr\s+create\b(.*?)(?:\|\||&&|;|$)`).FindStringSubmatch(run)
	if segment == nil {
		return "", false
	}
	h := regexp.MustCompile(`--head\s+(\S+)`).FindStringSubmatch(segment[1])
	if h == nil {
		return "", false
	}
	return h[1], true
}

// checkNoPush (no-default-branch-push-from-workflows) allows exactly the
// listed pushes: the release job pushes the release branch, each bot job
// pushes its kind's branch, and tag publishes live only in the release job's
// tag step. Every other push is reported: a branch push to any other branch
// (including the default branch), an unresolvable push, a tag push outside
// the tag step, a gh release creation outside the tag step, a gh pr merge
// outside an arm/hold step, a gh pr close or edit outside the release job, and
// a gh pr create without --head or with a head that is not the job's branch.
func (c *queueChecker) checkNoPush() {
	allowed := c.resolvePushes(QueueRuleNoPush)
	byFile := map[string]map[string]string{}
	for _, a := range allowed {
		if byFile[a.workflow] == nil {
			byFile[a.workflow] = map[string]string{}
		}
		byFile[a.workflow][a.job] = a.branch
	}
	releaseFile, releaseJob := "", ""
	if rel := c.policy.Queue.ReleaseJob; rel != nil {
		releaseFile, releaseJob = rel.Workflow, rel.Job
	}
	for _, name := range c.sortedNames() {
		f := c.files[name]
		if f.missing || f.fault != "" {
			if !f.reported && f.fault != "" {
				f.reported = true
				c.emit(QueueRuleNoPush, name, "", f.fault)
			}
			continue
		}
		for _, jobName := range jobNames(f.doc) {
			job := jobNode(f.doc, jobName)
			branch, listed := byFile[name][jobName]
			for _, step := range stepsOf(job) {
				c.checkPushStep(name, jobName, step, branch, listed, name == releaseFile && jobName == releaseJob)
			}
		}
	}
}

// checkPushStep applies the push allowlist to one step. listed tells whether
// the policy lists this (file, job); branch is its branch when listed;
// isReleaseJob tells whether this is the release job of the release workflow.
func (c *queueChecker) checkPushStep(workflow, job string, step *yaml.Node, branch string, listed, isReleaseJob bool) {
	run := stepRunText(step)
	id := stepID(step)
	for _, site := range gitPushSites(run) {
		switch {
		case site.isTag:
			if !(workflow == releaseWorkflow(c) && isTagStep(step)) {
				c.emit(QueueRuleNoPush, workflow, id,
					"a step pushes a tag outside the release job's tag step during merge-queue policy validation ("+quote(site.raw)+"). Tag publishes live only in the release tag step; move this push there or drop it")
			}
		case site.unresolved:
			c.emit(QueueRuleNoPush, workflow, id,
				"a step pushes without a parseable branch refspec during merge-queue policy validation ("+quote(site.raw)+"). Name the branch explicitly so the checker can tell it is the job's own branch")
		case !listed || site.target != branch:
			c.emit(QueueRuleNoPush, workflow, id,
				"a step pushes branch "+quote(site.target)+" during merge-queue policy validation, but the policy lists no such push. Only the release job pushes the release branch and each bot job pushes its kind's branch; point the push at the job's own branch")
		}
	}
	if prCreate.FindStringSubmatch(run) != nil {
		head, ok := prHeadTarget(run)
		switch {
		case !ok:
			c.emit(QueueRuleNoPush, workflow, id,
				"a step opens a PR without --head during merge-queue policy validation. Name the head branch explicitly so the checker can tell it is the job's own branch")
		case !listed || head != branch:
			c.emit(QueueRuleNoPush, workflow, id,
				"a step opens a PR on branch "+quote(head)+" during merge-queue policy validation, but the policy lists no such push. Only the release job opens on the release branch and each bot job on its kind's branch")
		}
	}
	if prMerge.FindStringSubmatch(run) != nil && !isArmHoldStep(step) {
		c.emit(QueueRuleNoPush, workflow, id,
			"a step merges a PR outside an arm/hold step during merge-queue policy validation. Only the gate-tested arm and hold steps may merge; gate this step on the queue-gate output or drop it")
	}
	if (prCloseOrEdit.FindStringSubmatch(run) != nil) && !isReleaseJob {
		c.emit(QueueRuleNoPush, workflow, id,
			"a step closes or edits a PR outside the release job during merge-queue policy validation. Only the release job touches the release PR; a bot job updates its own PR by pushing its branch")
	}
	if strings.Contains(run, "gh release create") && !(isReleaseJob && isTagStep(step)) {
		c.emit(QueueRuleNoPush, workflow, id,
			"a step creates a GitHub Release outside the release job's tag step during merge-queue policy validation. Releases are published only by the release tag step")
	}
}

// releaseWorkflow returns the release workflow file, or "" when the policy
// names no release job.
func releaseWorkflow(c *queueChecker) string {
	if rel := c.policy.Queue.ReleaseJob; rel != nil {
		return rel.Workflow
	}
	return ""
}

var (
	// prCreate matches a gh pr create invocation.
	prCreate = regexp.MustCompile(`\bgh\s+pr\s+create\b`)
	// prMerge matches a gh pr merge invocation.
	prMerge = regexp.MustCompile(`\bgh\s+pr\s+merge\b`)
	// prCloseOrEdit matches a gh pr close or edit invocation.
	prCloseOrEdit = regexp.MustCompile(`\bgh\s+pr\s+(close|edit)\b`)
)

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
		var mint *yaml.Node
		for _, step := range stepsOf(job) {
			if isMintStep(step) {
				mint = step
				break
			}
		}
		if mint == nil {
			continue
		}
		jobCond := ""
		if cond := mappingValue(job, "if"); cond != nil {
			jobCond = cond.Value
		}
		if !hasRefCheck(stepIfText(mint)) && !hasRefCheck(jobCond) {
			c.emit(QueueRuleTokenMint, target.workflow, target.job,
				"job "+quote(target.job)+" mints the App token without checking github.ref during merge-queue policy validation. Compare github.ref == 'refs/heads/<default>' in the mint step's own if or the job if, or a dispatch on any branch runs that branch's file with the App key")
		}
		if !hasTagModeGate(stepIfText(mint)) && !hasTagModeGate(jobCond) {
			c.emit(QueueRuleTokenMint, target.workflow, target.job,
				"job "+quote(target.job)+" mints the App token without a tag-mode gate during merge-queue policy validation. Mint only when tag mode is on, with "+quoteTagMode()+" in the mint step's if or the job if, so a tag-mode-off run mints no token")
		}
	}
}

// refComparison matches the canonical default-branch comparison: github.ref
// == 'refs/heads/<default>', or the format form against the repository's
// default branch.
var refComparison = regexp.MustCompile(`github\.ref\s*==\s*'refs/heads/[^']+'|github\.ref\s*==\s*format\('refs/heads/\{0\}'[^)]*default_branch`)

// hasRefCheck reports whether an if carries the canonical default-branch
// comparison as a single comparison: the ref comparison, with no disjunction.
// A run-block test does not stop the job, so only an if counts.
func hasRefCheck(cond string) bool {
	return refComparison.MatchString(cond) && !strings.Contains(cond, "||")
}
