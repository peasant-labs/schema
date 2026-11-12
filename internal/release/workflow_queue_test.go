package release

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

// The policy loader corpus lives in testdata/workflow_policy_load.yaml: one
// inline policy document per row with its expected failure fragment. The
// runner writes each row to a temp file and runs the production loader.

func writeQueuePolicyFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "release-guard.policy.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write policy fixture: %v", err)
	}
	return path
}

const queuePolicyYAML = `jobs:
  - name: guard
queue:
  required:
    - { context: merge-gate, workflow: ci.yml, job: merge-gate, aggregator: true }
    - { context: title-guard, workflow: title-guard.yml, job: title-guard, guard: true }
  release_job:
    workflow: release-batch.yml
    job: release
    pre_tag_steps: [verify-contract, verify-freeze]
  bot_jobs:
    - { workflow: cut-release.yml, job: cut, kind: cut-release }
`

func TestLoadWorkflowPolicy_QueueSection(t *testing.T) {
	t.Parallel()

	policy, err := LoadWorkflowPolicy(writeQueuePolicyFile(t, queuePolicyYAML))
	if err != nil {
		t.Fatalf("LoadWorkflowPolicy(queue policy) failed: %v", err)
	}
	if policy.Queue == nil {
		t.Fatalf("loaded policy has a nil queue section, want the parsed required/release/bot entries")
	}
	if len(policy.Queue.Required) != 2 {
		t.Fatalf("loaded queue has %d required entries, want 2", len(policy.Queue.Required))
	}
	gate := policy.Queue.Required[0]
	if gate.Context != "merge-gate" || gate.Workflow != "ci.yml" || gate.Job != "merge-gate" || !gate.Aggregator || gate.Guard {
		t.Fatalf("loaded required[0] = %+v, want the merge-gate aggregator entry", gate)
	}
	guard := policy.Queue.Required[1]
	if guard.Context != "title-guard" || guard.Workflow != "title-guard.yml" || guard.Job != "title-guard" || guard.Aggregator || !guard.Guard {
		t.Fatalf("loaded required[1] = %+v, want the title-guard guard entry", guard)
	}
	if policy.Queue.ReleaseJob == nil {
		t.Fatalf("loaded queue has a nil release job, want release-batch.yml/release")
	}
	rel := policy.Queue.ReleaseJob
	if rel.Workflow != "release-batch.yml" || rel.Job != "release" {
		t.Fatalf("loaded release job = %+v, want the release-batch entry", rel)
	}
	if len(rel.PreTagSteps) != 2 || rel.PreTagSteps[0] != "verify-contract" || rel.PreTagSteps[1] != "verify-freeze" {
		t.Fatalf("loaded pre_tag_steps = %q, want [verify-contract verify-freeze]", rel.PreTagSteps)
	}
	if len(policy.Queue.BotJobs) != 1 {
		t.Fatalf("loaded queue has %d bot jobs, want 1", len(policy.Queue.BotJobs))
	}
	bot := policy.Queue.BotJobs[0]
	if bot.Workflow != "cut-release.yml" || bot.Job != "cut" || bot.Kind != intent.PRKindCutRelease {
		t.Fatalf("loaded bot job = %+v, want the cut-release kind entry", bot)
	}
}

func TestLoadWorkflowPolicy_LegacyJobsOnlyHasNoQueue(t *testing.T) {
	t.Parallel()

	// Today's jobs-only shape must load with a nil queue section, so the
	// queue half constrains nothing until a repo adopts it.
	policy, err := LoadWorkflowPolicy(writeQueuePolicyFile(t, "jobs:\n  - name: guard\n"))
	if err != nil {
		t.Fatalf("LoadWorkflowPolicy(jobs-only) failed: %v", err)
	}
	if policy.Queue != nil {
		t.Fatalf("jobs-only policy loaded with queue %+v, want nil", policy.Queue)
	}
}

func TestLoadWorkflowPolicy_QueueOnly(t *testing.T) {
	t.Parallel()

	// A queue section with no jobs loads, so a repo with no tag-triggered
	// release workflow can adopt the queue half on its own.
	policy, err := LoadWorkflowPolicy(writeQueuePolicyFile(t, `queue:
  required:
    - { context: title-guard, workflow: title-guard.yml, job: title-guard, guard: true }
`))
	if err != nil {
		t.Fatalf("LoadWorkflowPolicy(queue-only) failed: %v", err)
	}
	if policy.Queue == nil {
		t.Fatalf("queue-only policy loaded with a nil queue section")
	}
	if len(policy.Jobs) != 0 {
		t.Fatalf("queue-only policy loaded with %d jobs, want none", len(policy.Jobs))
	}
}

//go:embed testdata/workflow_policy_load.yaml
var workflowPolicyLoadYAML []byte

// loadCaseInput is one inline policy document from the loader corpus.
type loadCaseInput struct {
	Policy string `yaml:"policy"`
}

// loadCaseExpected is the failure fragment a must-fail loader row must carry.
type loadCaseExpected struct {
	Rule         string   `yaml:"rule"`
	WantContains []string `yaml:"want_contains"`
}

// loadCaseRow is one loader corpus row.
type loadCaseRow struct {
	Name           string                  `yaml:"name"`
	Input          loadCaseInput           `yaml:"input"`
	Expected       loadCaseExpected        `yaml:"expected"`
	Classification testcase.Classification `yaml:"classification"`
	Provenance     testcase.Provenance     `yaml:"provenance"`
	Mutation       testcase.Mutation       `yaml:"mutation"`
	Sibling        string                  `yaml:"sibling"`
}

func (r loadCaseRow) siblingName() string                   { return r.Name }
func (r loadCaseRow) siblingClass() testcase.Classification { return r.Classification }
func (r loadCaseRow) siblingOf() string                     { return r.Sibling }

// loadPolicyFixtures is the parsed loader corpus with its manifest.
type loadPolicyFixtures struct {
	RequiredNames []string      `yaml:"required_names"`
	Rows          []loadCaseRow `yaml:"cases"`
}

func loadLoadPolicyFixtures(t *testing.T) loadPolicyFixtures {
	t.Helper()
	dec := yaml.NewDecoder(bytes.NewReader(workflowPolicyLoadYAML))
	dec.KnownFields(true)
	var f loadPolicyFixtures
	if err := dec.Decode(&f); err != nil {
		t.Fatalf("load policy loader cases (testdata/workflow_policy_load.yaml): %v", err)
	}
	cases := make([]testcase.Case[loadCaseInput, loadCaseExpected], 0, len(f.Rows))
	rows := make([]siblingCased, 0, len(f.Rows))
	for _, row := range f.Rows {
		row := row
		cases = append(cases, testcase.Case[loadCaseInput, loadCaseExpected]{
			Name: row.Name, Input: row.Input, Expected: row.Expected,
			Classification: row.Classification, Provenance: row.Provenance, Mutation: row.Mutation,
		})
		rows = append(rows, row)
	}
	corpus := testcase.Corpus[loadCaseInput, loadCaseExpected]{Cases: cases}
	assert.RequireValid(t, corpus)
	checkSiblings(t, rows)
	seen := map[string]bool{}
	for _, row := range f.Rows {
		seen[row.Name] = true
	}
	for _, name := range f.RequiredNames {
		if !seen[name] {
			t.Fatalf("required loader case %q is missing", name)
		}
	}
	if len(seen) != len(f.RequiredNames) {
		t.Fatalf("required_names names %d cases but the corpus carries %d", len(f.RequiredNames), len(seen))
	}
	return f
}

func TestLoadWorkflowPolicy_LoadCases(t *testing.T) {
	t.Parallel()

	for _, row := range loadLoadPolicyFixtures(t).Rows {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadWorkflowPolicy(writeQueuePolicyFile(t, row.Input.Policy))
			if row.Classification == testcase.MustPass {
				if err != nil {
					t.Fatalf("accept case %q failed to load: %v", row.Name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("reject case %q loaded without error", row.Name)
			}
			for _, want := range row.Expected.WantContains {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error %q does not contain expected fragment %q", err.Error(), want)
				}
			}
		})
	}
}

func TestQueueRuleClosedSet(t *testing.T) {
	t.Parallel()

	// Closed sets assert exact membership: the declared list must equal the
	// plan's rule set, not merely count it.
	want := []QueueRule{
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
	if !reflect.DeepEqual(AllQueueRules, want) {
		t.Fatalf("AllQueueRules = %v, want exactly %v", AllQueueRules, want)
	}
	if (QueueRule("no-such-rule")).IsValid() {
		t.Fatalf("unknown rule passes IsValid")
	}
}

func TestClosedSetsMatchIntent(t *testing.T) {
	t.Parallel()

	// The checker's action and gate sets are the same objects as the intent
	// package's, not a second copy: sorted values must match exactly.
	actions := make([]string, 0, len(intent.AllActions))
	for _, a := range intent.AllActions {
		actions = append(actions, a.String())
	}
	gotActions := make([]string, 0, len(closedActions))
	for a := range closedActions {
		gotActions = append(gotActions, a)
	}
	if !reflect.DeepEqual(sortedStrings(gotActions), sortedStrings(actions)) {
		t.Fatalf("checker actions %v do not equal intent.AllActions %v", gotActions, actions)
	}
	var gateVals []string
	for _, g := range intent.AllGates {
		gateVals = append(gateVals, g.String())
	}
	if !reflect.DeepEqual(gateValues(), gateVals) {
		t.Fatalf("checker gates %v do not equal intent.AllGates %v", gateValues(), gateVals)
	}
}

func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

func TestCheckWorkflows_LegacyPassesEmpty(t *testing.T) {
	t.Parallel()

	// The entry-point contract for jobs-only policies: no queue section means
	// no diagnostics and no required contexts.
	diags, required := CheckWorkflows(WorkflowPolicy{Jobs: []JobRule{{Name: "guard"}}}, intent.Config{}, t.TempDir())
	if len(diags) != 0 {
		t.Fatalf("jobs-only policy yielded %d diagnostics, want none: %+v", len(diags), diags)
	}
	if len(required) != 0 {
		t.Fatalf("jobs-only policy yielded %d required contexts, want none", len(required))
	}
}

func TestCheckWorkflows_QueuePresentPassesWhenValid(t *testing.T) {
	t.Parallel()

	// The entry-point contract for a queue policy: a directory whose
	// workflows satisfy every rule yields no diagnostics, and the required
	// contexts come back in policy order for --print-required. (The corpus
	// accept cases prove each rule's accept path through fixtures; this pins
	// the contract directly.)
	policy, err := LoadWorkflowPolicy(writeQueuePolicyFile(t, `jobs:
  - name: guard
queue:
  required:
    - { context: merge-gate, workflow: ci.yml, job: merge-gate, aggregator: true }
    - { context: title-guard, workflow: title-guard.yml, job: title-guard, guard: true }
`))
	if err != nil {
		t.Fatalf("LoadWorkflowPolicy(queue policy) failed: %v", err)
	}
	dir := t.TempDir()
	workflows := map[string]string{
		"ci.yml":          "name: CI\non:\n  pull_request:\n  merge_group:\nconcurrency:\n  group: ci-${{ github.ref }}\njobs:\n  lib:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make lib\n  merge-gate:\n    if: always()\n    needs: [lib]\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo gate\n",
		"title-guard.yml": "name: Title guard\non:\n  pull_request:\n  merge_group:\njobs:\n  title-guard:\n    runs-on: ubuntu-latest\n    steps:\n      - run: release-intent classify --author x --head-branch y\n      - run: release-intent check-title --base x --head y --author z\n",
	}
	for name, content := range workflows {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write workflow fixture: %v", err)
		}
	}
	diags, required := CheckWorkflows(policy, loadBaseConfig(t), dir)
	if len(diags) != 0 {
		msgs := make([]string, 0, len(diags))
		for _, d := range diags {
			msgs = append(msgs, d.Message)
		}
		t.Fatalf("valid directory yielded %d diagnostics:\n%s", len(diags), strings.Join(msgs, "\n"))
	}
	if len(required) != 2 || required[0].Context != "merge-gate" || required[1].Context != "title-guard" {
		t.Fatalf("required contexts = %+v, want the two policy entries in order", required)
	}
}
