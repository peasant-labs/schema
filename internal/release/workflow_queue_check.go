package release

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// This file is the merge-queue half of the workflow policy checker. It reads
// every workflow file in a directory and applies the eight closed queue rules
// from the policy's queue section: required contexts resolve, required
// workflows trigger on merge_group with per-ref concurrency, the release job's
// write steps are gated on the planner's action in the canonical verdict form,
// the title guard cannot pass early, range readers take the floor, each bot
// branch has one writer, and token-minting jobs check the ref first.
//
// A policy without a queue section constrains nothing here. A repo with no
// release job skips the release-job clauses. Every diagnostic is actionable:
// it names the file, the job, and the rule, says why the shape matters, and
// how to fix it.

// CheckWorkflows validates every workflow file in dir against the policy's
// queue section and returns the violations plus the required contexts in
// policy order (for the operator to compare against the queue ruleset). A
// policy without a queue section constrains nothing here and yields no
// diagnostics, so today's jobs-only policies keep passing unchanged.
//
// A non-empty diagnostic list means the directory is rejected; an empty one
// means it satisfies every rule. A file that cannot be read or parsed is
// reported under the first rule that needs it, so one run names every broken
// file instead of stopping at the first.
func CheckWorkflows(policy WorkflowPolicy, dir string) ([]Diagnostic, []RequiredContext) {
	if policy.Queue == nil {
		return nil, nil
	}
	c := &queueChecker{policy: policy, dir: dir}
	if err := c.load(); err != nil {
		return []Diagnostic{{
			Rule:     QueueRuleRequiredContexts,
			Workflow: dir,
			Message:  "check workflows: " + dir + ": rule required-contexts: cannot list the workflows directory during merge-queue policy validation: " + err.Error() + ". Fix the --dir path or run from the repository root",
		}}, append([]RequiredContext(nil), policy.Queue.Required...)
	}
	c.checkRequiredContexts()
	c.checkMergeGroup()
	c.checkConcurrency()
	c.checkReleaseActions()
	c.checkTitleGuard()
	c.checkRangeFloor()
	c.checkBotWriter()
	c.checkTokenMint()
	return c.diagnostics, append([]RequiredContext(nil), policy.Queue.Required...)
}

// queueChecker holds one CheckWorkflows run: the parsed workflow files keyed by
// base name, the faults already reported, and the diagnostics so far. Files
// are parsed once up front; a read or parse fault is attributed to the first
// rule that needs the file and skipped by later rules.
type queueChecker struct {
	policy      WorkflowPolicy
	dir         string
	files       map[string]*queueFile
	diagnostics []Diagnostic
}

// queueFile is one parsed workflow file, or the fault that kept it from
// parsing with whether a rule has already reported that fault.
type queueFile struct {
	name     string
	raw      []byte
	doc      *yaml.Node
	fault    string
	reported bool
	missing  bool
}

// load reads every *.yml and *.yaml file in the directory in sorted order. A
// directory that cannot be listed is a hard failure for the caller; a file
// that cannot be read or parsed becomes a fault reported by the first rule
// that needs it.
func (c *queueChecker) load() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}
	c.files = map[string]*queueFile{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		path := filepath.Join(c.dir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			c.files[name] = &queueFile{name: name, fault: "cannot read " + path + " during merge-queue policy validation: " + err.Error() + ". Restore the file or fix its permissions"}
			continue
		}
		var doc yaml.Node
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			c.files[name] = &queueFile{name: name, raw: raw, fault: "cannot parse " + path + " as YAML during merge-queue policy validation: " + err.Error() + ". Fix the workflow syntax before it can be guarded"}
			continue
		}
		if len(doc.Content) == 0 {
			c.files[name] = &queueFile{name: name, raw: raw, fault: "empty workflow " + path + " during merge-queue policy validation. Restore the workflow before it can be guarded"}
			continue
		}
		c.files[name] = &queueFile{name: name, raw: raw, doc: doc.Content[0]}
	}
	return nil
}

// needFile returns the parsed workflow file a rule needs. A missing file and
// a faulted file are reported once, under the first rule that needs them;
// later rules skip the file silently so one broken file yields one diagnostic.
func (c *queueChecker) needFile(rule QueueRule, name string) *yaml.Node {
	f, ok := c.files[name]
	if !ok {
		c.emit(rule, name, "", "the required workflow "+name+" is missing from "+c.dir+" during merge-queue policy validation. Add the file so the queue can run the required checks")
		c.files[name] = &queueFile{name: name, missing: true, reported: true}
		return nil
	}
	if f.missing {
		return nil
	}
	if f.fault != "" {
		if !f.reported {
			f.reported = true
			c.emit(rule, name, "", f.fault)
		}
		return nil
	}
	return f.doc
}

// emit appends one diagnostic. The message carries the rule so CLI output and
// fixture assertions can name what fired without parsing the text.
func (c *queueChecker) emit(rule QueueRule, workflow, job, what string) {
	where := workflow
	if where == "" {
		where = "policy queue"
	} else {
		where = c.dir + "/" + workflow
	}
	c.diagnostics = append(c.diagnostics, Diagnostic{
		Rule:     rule,
		Workflow: workflow,
		Job:      job,
		Message:  "check workflows: " + where + ": rule " + rule.String() + ": " + what,
	})
}

// sortedNames returns the loaded file names in sorted order for deterministic
// whole-directory scans.
func (c *queueChecker) sortedNames() []string {
	names := make([]string, 0, len(c.files))
	for name := range c.files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// checkRequiredContexts pins every required check the queue waits on: the
// workflow file must exist in the directory and hold the named job. The
// equality with the queue ruleset itself is an operator step over the returned
// required contexts, not something this checker can observe.
func (c *queueChecker) checkRequiredContexts() {
	for _, req := range c.policy.Queue.Required {
		doc := c.needFile(QueueRuleRequiredContexts, req.Workflow)
		if doc == nil {
			continue
		}
		if jobNode(doc, req.Job) == nil {
			c.emit(QueueRuleRequiredContexts, req.Workflow, req.Job,
				"jobs."+req.Job+" is missing during merge-queue policy validation, but the policy requires context "+req.Context+" from it. Add the job so the context the queue waits on is actually reported")
		}
	}
}

// checkMergeGroup requires every workflow that owns a required context to
// trigger on merge_group as well as on pull_request, so every queued entry is
// checked again after enqueue.
func (c *queueChecker) checkMergeGroup() {
	for _, req := range c.policy.Queue.Required {
		doc := c.needFile(QueueRuleMergeGroup, req.Workflow)
		if doc == nil {
			continue
		}
		if !hasTrigger(doc, "merge_group") {
			c.emit(QueueRuleMergeGroup, req.Workflow, req.Job,
				"no merge_group trigger during merge-queue policy validation, but the workflow owns required context "+req.Context+". Trigger on merge_group so the queue build checks the entry that is actually landing")
		}
	}
}

// checkConcurrency requires every workflow that owns a required context to key
// its concurrency group on the ref or the run id, and to cancel in-progress
// runs only for pull_request. A group keyed only on the PR number is empty on
// a merge group, so every group build would share one group and cancel all
// but one pending run.
func (c *queueChecker) checkConcurrency() {
	for _, req := range c.policy.Queue.Required {
		doc := c.needFile(QueueRuleConcurrency, req.Workflow)
		if doc == nil {
			continue
		}
		concurrency := mappingValue(doc, "concurrency")
		if concurrency == nil {
			continue
		}
		var group, cancel string
		switch concurrency.Kind {
		case yaml.ScalarNode:
			group = concurrency.Value
		case yaml.MappingNode:
			if g := mappingValue(concurrency, "group"); g != nil {
				group = g.Value
			}
			if k := mappingValue(concurrency, "cancel-in-progress"); k != nil {
				cancel = k.Value
			}
		default:
			c.emit(QueueRuleConcurrency, req.Workflow, "",
				"concurrency is neither a group name nor a mapping during merge-queue policy validation. Key the group on the ref or the run id so merge-group builds never share a group")
			continue
		}
		if !strings.Contains(group, "github.ref") && !strings.Contains(group, "github.run_id") {
			c.emit(QueueRuleConcurrency, req.Workflow, "",
				"concurrency group "+quote(group)+" is not keyed on the ref or the run id during merge-queue policy validation. A group keyed only on the PR number is empty on a merge group, so every group build would share one group and cancel all but one pending run; key it on github.ref or github.run_id")
		}
		switch {
		case cancel == "", cancel == "false":
		case cancel == "true":
			c.emit(QueueRuleConcurrency, req.Workflow, "",
				"cancel-in-progress is true during merge-queue policy validation. Cancelling applies to the merge-group builds too; scope it to pull_request with ${{ github.event_name == 'pull_request' }} or drop it")
		default:
			if !strings.Contains(cancel, "pull_request") {
				c.emit(QueueRuleConcurrency, req.Workflow, "",
					"cancel-in-progress expression "+quote(cancel)+" is not scoped to pull_request during merge-queue policy validation. Scope the cancellation to ${{ github.event_name == 'pull_request' }} so merge-group builds are never cancelled")
			}
		}
	}
}

// hasTrigger reports whether the workflow's `on:` includes the named event,
// in scalar, sequence, or mapping form.
func hasTrigger(doc *yaml.Node, event string) bool {
	on := mappingValue(doc, "on")
	if on == nil {
		return false
	}
	switch on.Kind {
	case yaml.ScalarNode:
		return on.Value == event
	case yaml.SequenceNode:
		for _, item := range on.Content {
			if item.Value == event {
				return true
			}
		}
	case yaml.MappingNode:
		return mappingValue(on, event) != nil
	}
	return false
}

func quote(s string) string {
	if s == "" {
		return "<empty>"
	}
	return `"` + s + `"`
}
