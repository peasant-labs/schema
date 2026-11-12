package release

import (
	"regexp"
	"strconv"
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
// --until) to pass --floor in its quoted form, and every plan step to pass
// --mode quoted too: --floor "${{ vars.RELEASE_BATCH_FLOOR }}", --mode
// "${{ vars.RELEASE_BATCH_MODE }}", or the same values through env
// (--floor "$FLOOR" with FLOOR mapped from the variable). An unquoted
// expansion vanishes when the variable is unset -- the normal state, set only
// for recovery -- turning the flag into a missing value or swallowing the
// next flag, so the job fails closed red on every run.
func (c *queueChecker) checkRangeFloor() {
	c.eachDoc(QueueRuleRangeFloor, func(name string, doc *yaml.Node) {
		for _, jobName := range jobNames(doc) {
			job := jobNode(doc, jobName)
			for _, step := range stepsOf(job) {
				run := stepRunText(step)
				readsRange := rangeReader.MatchString(run) ||
					(strings.Contains(run, "check-contract") && strings.Contains(run, "--until"))
				if !readsRange {
					continue
				}
				c.checkQuotedFlag(name, job, jobName, step, run, "floor", "RELEASE_BATCH_FLOOR")
				if planRunner.MatchString(run) {
					c.checkQuotedFlag(name, job, jobName, step, run, "mode", "RELEASE_BATCH_MODE")
				}
			}
		}
	})
}

// checkQuotedFlag requires one range-reader argument in its quoted form: the
// flag followed by "${{ vars.VAR }}" directly, or by "$ENV" with ENV mapped
// from the variable in the step or job env.
func (c *queueChecker) checkQuotedFlag(name string, job *yaml.Node, jobName string, step *yaml.Node, run, flag, variable string) {
	value, ok := flagValue(run, flag)
	if !ok || !c.quotedVariable(step, job, value, variable) {
		c.emit(QueueRuleRangeFloor, name, stepID(step),
			"a step reads a commit range without a quoted --"+flag+" during merge-queue policy validation. Pass --"+flag+" \"${{ vars."+variable+" }}\" (or --"+flag+" \"$VAR\" with VAR mapped from the variable in env): an unquoted expansion vanishes when the variable is unset and the flag eats the next argument")
	}
}

// quotedVariable reports whether value carries the variable: the double-quoted
// ${{ vars.VAR }} form directly, or a double-quoted $ENV mapped from the
// variable in the step or job env. Single quotes and bare words never expand.
func (c *queueChecker) quotedVariable(step, job *yaml.Node, value, variable string) bool {
	if len(value) >= 2 && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		inner := value[1 : len(value)-1]
		if strings.Contains(inner, "vars."+variable) {
			return true
		}
		if strings.HasPrefix(inner, "$") {
			if v := envValue(step, inner[1:]); strings.Contains(v, "vars."+variable) {
				return true
			}
			if v := envValue(job, inner[1:]); strings.Contains(v, "vars."+variable) {
				return true
			}
		}
	}
	return false
}

// envValue returns the value of the named env entry on a step or job, or "".
func envValue(node *yaml.Node, name string) string {
	if env := mappingValue(node, "env"); env != nil {
		if v := mappingValue(env, name); v != nil {
			return v.Value
		}
	}
	return ""
}

// flagValue returns the shell word following --flag in the run line, keeping
// quoted spans whole, or false when the flag is absent or valueless.
func flagValue(run, flag string) (string, bool) {
	words := shellWords(run)
	for i, w := range words {
		if w == "--"+flag && i+1 < len(words) {
			return words[i+1], true
		}
		if rest, ok := strings.CutPrefix(w, "--"+flag+"="); ok {
			return rest, true
		}
	}
	return "", false
}

// shellWords is the checker's one shell-word splitter. It splits a run line
// on unquoted whitespace and keeps three spans whole: a single-quoted span, a
// double-quoted span, and an unquoted GitHub expression (${{ ... }}), which
// may hold spaces. Quotes stay in the word so a caller can tell quoted from
// bare; bareWord strips them where the bare value is needed.
func shellWords(s string) []string {
	var out []string
	var cur strings.Builder
	var quote byte
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case quote != 0:
			cur.WriteByte(ch)
			if ch == quote {
				quote = 0
			}
		case ch == '"' || ch == '\'':
			quote = ch
			cur.WriteByte(ch)
		case strings.HasPrefix(s[i:], "${{"):
			end := strings.Index(s[i:], "}}")
			if end < 0 {
				cur.WriteString(s[i:])
				i = len(s)
				break
			}
			cur.WriteString(s[i : i+end+2])
			i += end + 1
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			flush()
		default:
			cur.WriteByte(ch)
		}
	}
	flush()
	return out
}

// bareWord returns a shell word with its delimiting quotes removed, the value
// the shell would pass to the command (expansions left as written).
func bareWord(word string) string {
	var out strings.Builder
	var quote byte
	for i := 0; i < len(word); i++ {
		ch := word[i]
		switch {
		case quote != 0 && ch == quote:
			quote = 0
		case quote == 0 && (ch == '"' || ch == '\''):
			quote = ch
		default:
			out.WriteByte(ch)
		}
	}
	return out.String()
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
				"the policy names bot kind "+strconv.Quote(bot.Kind.String())+" for job "+strconv.Quote(bot.Job)+" during merge-queue policy validation, but the release config names no branch for that kind. Add the bot_branches entry so the checker can tell the bot branch apart")
			continue
		}
		if owner, dup := seen[bot.Kind]; dup {
			c.emit(rule, "", bot.Job,
				"bot kind "+strconv.Quote(bot.Kind.String())+" is written by "+owner+" and by job "+strconv.Quote(bot.Job)+" during merge-queue policy validation. Give each bot kind exactly one writer job so two jobs can never push the same branch")
			continue
		}
		seen[bot.Kind] = "job " + strconv.Quote(bot.Job)
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

// splitPushArgs splits a git push command line into its bare refspec
// operands, skipping flags and the remote name. Quoted spans and GitHub
// expression spans stay one operand each.
func splitPushArgs(rest string) []string {
	var out []string
	skipValue := map[string]bool{"--repo": true, "--remote": true}
	remoteSkipped := false
	words := shellWords(rest)
	for i := 0; i < len(words); i++ {
		w := bareWord(words[i])
		if skipValue[w] {
			i++
			continue
		}
		if strings.HasPrefix(w, "-") {
			continue
		}
		if !remoteSkipped {
			remoteSkipped = true
			continue
		}
		out = append(out, w)
	}
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

var (
	// shellSegments splits a run line into shell segments on ||, &&, ;, and
	// pipes, so one push per segment is found.
	shellSegments = regexp.MustCompile(`\|\||&&|[|;&]`)
	// prCreateHead extracts the --head branch of a gh pr create command.
	prCreateHead = regexp.MustCompile(`\bgh\s+pr\s+create\b(.*?)(?:\|\||&&|;|$)`)
)

// gitPushSites returns every git push in the run text: branch targets, tag
// pushes, and pushes with no parseable refspec. One push per shell segment is
// assumed; a segment is split on ||, &&, ;, and pipes.
func gitPushSites(run string) []pushSite {
	var out []pushSite
	for _, segment := range shellSegments.Split(run, -1) {
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
	segment := prCreateHead.FindStringSubmatch(run)
	if segment == nil {
		return "", false
	}
	head, ok := flagValue(segment[1], "head")
	if !ok {
		return "", false
	}
	return bareWord(head), true
}

// prTargets returns the (verb, target) pairs for gh pr merge, close, and
// edit invocations in the run text: the first positional argument after the
// verb, or "" when the command names none (a number, a URL, or a flag-only
// call). Quoted spans stay one word, and flag values are skipped.
func prTargets(run string) [][2]string {
	var out [][2]string
	for _, m := range prMergeClose.FindAllStringSubmatch(run, -1) {
		verb, rest := m[1], m[2]
		target := ""
		skipNext := false
		for _, word := range shellWords(rest) {
			tok := bareWord(word)
			if skipNext {
				skipNext = false
				continue
			}
			if strings.HasPrefix(tok, "-") {
				skipNext = prValueFlags[tok]
				continue
			}
			target = tok
			break
		}
		out = append(out, [2]string{verb, target})
	}
	return out
}

// checkNoPush (no-default-branch-push-from-workflows) allows exactly the
// listed pushes: the release job pushes the release branch, each bot job
// pushes its kind's branch, and tag publishes live only in the release job's
// tag step. Every other push is reported: a branch push to any other branch
// (including the default branch), an unresolvable push, a tag push outside
// the tag step, a gh release creation outside the tag step (except the
// tag-triggered publish, see tagPublishFault), a gh pr merge
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
	c.eachDoc(QueueRuleNoPush, func(name string, doc *yaml.Node) {
		for _, jobName := range jobNames(doc) {
			job := jobNode(doc, jobName)
			branch, listed := byFile[name][jobName]
			for _, step := range stepsOf(job) {
				c.checkPushStep(doc, name, jobName, step, branch, listed, name == releaseFile && jobName == releaseJob)
			}
		}
	})
}

// checkPushStep applies the push allowlist to one step. listed tells whether
// the policy lists this (file, job); branch is its branch when listed;
// isReleaseJob tells whether this is the release job of the release workflow.
func (c *queueChecker) checkPushStep(doc *yaml.Node, workflow, job string, step *yaml.Node, branch string, listed, isReleaseJob bool) {
	run := stepRunText(step)
	id := stepID(step)
	for _, site := range gitPushSites(run) {
		switch {
		case site.isTag:
			if !(workflow == releaseWorkflow(c) && isTagStep(step)) {
				c.emit(QueueRuleNoPush, workflow, id,
					"a step pushes a tag outside the release job's tag step during merge-queue policy validation ("+strconv.Quote(site.raw)+"). Tag publishes live only in the release tag step; move this push there or drop it")
			}
		case site.unresolved:
			c.emit(QueueRuleNoPush, workflow, id,
				"a step pushes without a parseable branch refspec during merge-queue policy validation ("+strconv.Quote(site.raw)+"). Name the branch explicitly so the checker can tell it is the job's own branch")
		case !listed || site.target != branch:
			c.emit(QueueRuleNoPush, workflow, id,
				"a step pushes branch "+strconv.Quote(site.target)+" during merge-queue policy validation, but the policy lists no such push. Only the release job pushes the release branch and each bot job pushes its kind's branch; point the push at the job's own branch")
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
				"a step opens a PR on branch "+strconv.Quote(head)+" during merge-queue policy validation, but the policy lists no such push. Only the release job opens on the release branch and each bot job on its kind's branch")
		}
	}
	for _, pair := range prTargets(run) {
		verb, target := pair[0], pair[1]
		if verb == "merge" {
			if isArmHoldStep(step) {
				continue
			}
			if listed && target != "" && target == branch {
				continue
			}
			c.emit(QueueRuleNoPush, workflow, id,
				"a step merges a PR outside an arm/hold step during merge-queue policy validation. Only the gate-tested arm and hold steps may merge, and a listed bot job may merge only its own branch ("+strconv.Quote(branch)+"); gate this step on the queue-gate output or drop it")
			continue
		}
		if verb == "close" {
			if isReleaseJob {
				continue
			}
			if listed && target != "" && target == branch {
				continue
			}
			c.emit(QueueRuleNoPush, workflow, id,
				"a step closes a PR outside the release job during merge-queue policy validation. Only the release job closes the release PR, and a listed bot job closes only its own branch ("+strconv.Quote(branch)+"); a bot job otherwise updates its PR by pushing its branch")
		}
	}
	if prEdit.FindStringSubmatch(run) != nil && !isReleaseJob {
		c.emit(QueueRuleNoPush, workflow, id,
			"a step edits a PR outside the release job during merge-queue policy validation. Only the release job edits the release PR; a bot job updates its own PR by pushing its branch")
	}
	if releaseCreate.MatchString(run) && !(isReleaseJob && isTagStep(step)) {
		if fault := c.tagPublishFault(doc, jobNode(doc, job), step, run); fault != "" {
			c.emit(QueueRuleNoPush, workflow, id,
				"a step creates a GitHub Release outside the release job's tag step during merge-queue policy validation: "+fault+". A Release is created only by the release tag step, or by a workflow that runs only on a pushed release tag, writes no ref and no PR, and passes --verify-tag so it can never create a tag")
		}
	}
}

// releaseCreate matches a gh release create invocation.
var releaseCreate = regexp.MustCompile(`\bgh\s+release\s+create\b`)

// tagPublishFault says why a gh release create in this workflow is not the
// allowed tag-triggered publish, or "" when it is. The publish is allowed
// only when all three hold: the workflow's one trigger is push: tags: with
// every pattern starting with the release config's tag prefix, so it runs
// only for a release tag that already exists; no step pushes a ref, writes
// refs/tags/, or writes a PR; and every create in the step passes
// --verify-tag, so gh refuses to create a missing tag, and names the
// triggering tag as its tag argument (see triggeringTag). Such a workflow can
// publish a Release for the pushed tag and nothing else; gh release edit
// needs no further condition there.
func (c *queueChecker) tagPublishFault(doc, job, step *yaml.Node, run string) string {
	if fault := tagOnlyTriggerFault(doc, c.cfg.TagPrefix); fault != "" {
		return fault
	}
	for _, jobName := range jobNames(doc) {
		for _, step := range stepsOf(jobNode(doc, jobName)) {
			text := stepRunText(step)
			if refOrPRWrite.MatchString(text) {
				return "the tag-triggered workflow also pushes a ref or writes a PR (job " + strconv.Quote(jobName) + ")"
			}
		}
	}
	for _, words := range releaseCreateCommands(run) {
		verified := false
		for _, w := range words {
			if w == "--verify-tag" {
				verified = true
			}
		}
		if !verified {
			return "the create does not pass --verify-tag, so it would create the tag when it is missing"
		}
		tag := releaseCreateTag(words)
		if !triggeringTag(tag, step, job, doc) {
			return "the create targets " + strconv.Quote(tag) + ", which is not the triggering tag. Pass \"$GITHUB_REF_NAME\", ${{ github.ref_name }}, or an env variable mapped from ${{ github.ref_name }} as the tag argument"
		}
	}
	return ""
}

// releaseCreateValueFlags are the gh release create flags that take a value,
// so the word after one is not the tag argument.
var releaseCreateValueFlags = map[string]bool{
	"--title": true, "-t": true, "--notes": true, "-n": true, "--notes-file": true, "-F": true,
	"--target": true, "--repo": true, "-R": true, "--discussion-category": true,
	"--notes-start-tag": true,
}

// releaseCreateTag returns the tag argument of one gh release create command,
// its first positional word with quotes kept, or "" when it names none.
func releaseCreateTag(words []string) string {
	for i := 0; i < len(words); i++ {
		w := words[i]
		if strings.HasPrefix(w, "-") {
			if releaseCreateValueFlags[w] {
				i++
			}
			continue
		}
		return w
	}
	return ""
}

// refNameExpression matches the GitHub expression for the triggering ref's
// short name, alone.
var refNameExpression = regexp.MustCompile(`^\$\{\{\s*github\.ref_name\s*\}\}$`)

// triggeringTag reports whether a create's tag argument is the tag that
// triggered the workflow: ${{ github.ref_name }}, the runner's
// $GITHUB_REF_NAME, or an env variable the step, the job, or the workflow
// maps from ${{ github.ref_name }}, quoted or bare. A single-quoted shell
// variable never expands, so it does not count; any literal or other
// variable is rejected.
func triggeringTag(word string, step, job, doc *yaml.Node) bool {
	bare := bareWord(word)
	if refNameExpression.MatchString(bare) {
		return true
	}
	if strings.Contains(word, "'") {
		return false
	}
	name, ok := strings.CutPrefix(bare, "$")
	if !ok {
		return false
	}
	if inner, braced := strings.CutPrefix(name, "{"); braced {
		name, ok = strings.CutSuffix(inner, "}")
		if !ok {
			return false
		}
	}
	if !shellName.MatchString(name) {
		return false
	}
	for _, scope := range []*yaml.Node{step, job, doc} {
		if scope == nil {
			continue
		}
		if v := envValue(scope, name); v != "" {
			return refNameExpression.MatchString(strings.TrimSpace(v))
		}
	}
	return name == "GITHUB_REF_NAME"
}

// shellName matches a plain shell variable name.
var shellName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// refOrPRWrite matches a run line that pushes a ref, writes a tag ref, or
// writes a PR.
var refOrPRWrite = regexp.MustCompile(`\bgit\s+push\b|refs/tags/|\bgit\s+tag\b|\bgh\s+pr\s+(create|edit|close|merge)\b`)

// tagOnlyTriggerFault says why the workflow's on: is not exactly push: tags:
// with every pattern under the tag prefix, or "" when it is.
func tagOnlyTriggerFault(doc *yaml.Node, prefix string) string {
	on := mappingValue(doc, "on")
	if on == nil || on.Kind != yaml.MappingNode || len(on.Content) != 2 || on.Content[0].Value != "push" {
		return "the workflow's only trigger is not push: tags:"
	}
	push := on.Content[1]
	if push.Kind != yaml.MappingNode || len(push.Content) != 2 || push.Content[0].Value != "tags" {
		return "the workflow's push trigger is not limited to tags: alone"
	}
	var patterns []string
	switch tags := push.Content[1]; tags.Kind {
	case yaml.ScalarNode:
		patterns = []string{tags.Value}
	case yaml.SequenceNode:
		for _, item := range tags.Content {
			patterns = append(patterns, item.Value)
		}
	}
	if len(patterns) == 0 {
		return "the workflow's push: tags: names no pattern"
	}
	if prefix == "" {
		return "the release config sets no tag_prefix to hold the tag patterns to"
	}
	for _, pattern := range patterns {
		if !strings.HasPrefix(pattern, prefix) {
			return "the tag pattern " + strconv.Quote(pattern) + " does not start with the release config's tag_prefix " + strconv.Quote(prefix)
		}
	}
	return ""
}

// releaseCreateCommands returns the shell words of every gh release create
// command in the run text. A command runs to the end of its line, with
// backslash continuations joined, or to the next shell separator.
func releaseCreateCommands(run string) [][]string {
	joined := strings.ReplaceAll(strings.ReplaceAll(run, "\\\r\n", " "), "\\\n", " ")
	var out [][]string
	for _, line := range strings.Split(joined, "\n") {
		for _, loc := range releaseCreate.FindAllStringIndex(line, -1) {
			rest := line[loc[1]:]
			if sep := shellSegments.FindStringIndex(rest); sep != nil {
				rest = rest[:sep[0]]
			}
			out = append(out, shellWords(rest))
		}
	}
	return out
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
	// prEdit matches a gh pr edit invocation.
	prEdit = regexp.MustCompile(`\bgh\s+pr\s+edit\b`)
	// prMergeClose matches a gh pr merge or close invocation and captures the
	// verb and the rest of the shell segment.
	prMergeClose = regexp.MustCompile(`\bgh\s+pr\s+(merge|close)\b([\s\S]*?)(?:\|\||&&|;|$)`)
	// prValueFlags take a value, so the token after one is not the command's
	// positional target.
	prValueFlags = map[string]bool{
		"--title": true, "-t": true, "--body": true, "-b": true, "--body-file": true, "-F": true,
		"--comment": true, "-c": true, "--label": true, "-l": true, "--milestone": true, "-m": true,
		"--project": true, "-p": true, "--reviewer": true, "-r": true, "--assignee": true, "-a": true,
		"--repo": true, "-R": true, "--head": true, "--base": true, "-H": true, "-B": true,
		"--field": true, "--jq": true,
	}
)

func (c *queueChecker) checkTokenMint() {
	type target struct {
		workflow, job string
		kind          intent.PRKind
		isRelease     bool
	}
	var jobs []target
	if rel := c.policy.Queue.ReleaseJob; rel != nil {
		jobs = append(jobs, target{rel.Workflow, rel.Job, intent.PRKindRelease, true})
	}
	for _, bot := range c.policy.Queue.BotJobs {
		jobs = append(jobs, target{bot.Workflow, bot.Job, bot.Kind, false})
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
				"job "+strconv.Quote(target.job)+" mints the App token without checking github.ref during merge-queue policy validation. Compare github.ref == 'refs/heads/<default>' in the mint step's own if or the job if, or a dispatch on any branch runs that branch's file with the App key")
		}
		// Only the release job and the cut-release job mint behind tag mode:
		// the vendor-hash fix lands whenever its tree is stale, tag mode or
		// not.
		if (target.isRelease || target.kind == intent.PRKindCutRelease) &&
			!hasTagModeGate(stepIfText(mint)) && !hasTagModeGate(jobCond) {
			c.emit(QueueRuleTokenMint, target.workflow, target.job,
				"job "+strconv.Quote(target.job)+" mints the App token without a tag-mode gate during merge-queue policy validation. Mint only when tag mode is on, with "+quoteTagMode()+" in the mint step's if or the job if, so a tag-mode-off run mints no token")
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
