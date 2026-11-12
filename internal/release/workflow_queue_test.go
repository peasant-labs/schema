package release

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// L1 covers the policy model's queue section and the exported checker entry
// point: the loader accepts and validates the optional queue key with strict
// fields, jobs-only policies keep loading with a nil queue, and the checker
// entry point honors its legacy contract (no queue, no diagnostics) while the
// rule checks are still unwired.

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
    branch: automation/release
    pre_tag_steps: [verify-contract, verify-freeze]
  bot_jobs:
    - { workflow: cut-release.yml, job: cut, branch: automation/cut-release }
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
	if rel.Workflow != "release-batch.yml" || rel.Job != "release" || rel.Branch != "automation/release" {
		t.Fatalf("loaded release job = %+v, want the release-batch entry", rel)
	}
	if len(rel.PreTagSteps) != 2 || rel.PreTagSteps[0] != "verify-contract" || rel.PreTagSteps[1] != "verify-freeze" {
		t.Fatalf("loaded pre_tag_steps = %q, want [verify-contract verify-freeze]", rel.PreTagSteps)
	}
	if len(policy.Queue.BotJobs) != 1 {
		t.Fatalf("loaded queue has %d bot jobs, want 1", len(policy.Queue.BotJobs))
	}
	bot := policy.Queue.BotJobs[0]
	if bot.Workflow != "cut-release.yml" || bot.Job != "cut" || bot.Branch != "automation/cut-release" {
		t.Fatalf("loaded bot job = %+v, want the cut-release entry", bot)
	}
}

func TestLoadWorkflowPolicy_QueueStrictFields(t *testing.T) {
	t.Parallel()

	// A typo'd queue key must fail loudly, exactly like a typo'd jobs key.
	for _, tc := range []struct {
		name    string
		content string
	}{
		{
			name:    "unknown queue section key",
			content: "jobs:\n  - name: guard\nqueue:\n  require: []\n",
		},
		{
			name:    "unknown required entry key",
			content: "jobs:\n  - name: guard\nqueue:\n  required:\n    - { context: title-guard, workflow: title-guard.yml, job: title-guard, gare: true }\n",
		},
		{
			name:    "unknown release job key",
			content: "jobs:\n  - name: guard\nqueue:\n  required:\n    - { context: title-guard, workflow: title-guard.yml, job: title-guard }\n  release_job:\n    workflow: release-batch.yml\n    job: release\n    brunch: automation/release\n",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := LoadWorkflowPolicy(writeQueuePolicyFile(t, tc.content)); err == nil {
				t.Fatalf("expected a strict-fields error for %q, got nil", tc.name)
			} else if !strings.Contains(err.Error(), "cannot parse") {
				t.Fatalf("error %q does not name the parse failure", err.Error())
			}
		})
	}
}

func TestLoadWorkflowPolicy_QueueValidation(t *testing.T) {
	t.Parallel()

	base := "jobs:\n  - name: guard\nqueue:\n"
	for _, tc := range []struct {
		name    string
		queue   string
		wantSub string
	}{
		{
			name:    "empty required",
			queue:   "  required: []\n",
			wantSub: "declares no required contexts",
		},
		{
			name:    "required entry without context",
			queue:   "  required:\n    - { workflow: title-guard.yml, job: title-guard }\n",
			wantSub: "has an empty context",
		},
		{
			name:    "required entry without workflow",
			queue:   "  required:\n    - { context: title-guard, job: title-guard }\n",
			wantSub: "has an empty workflow",
		},
		{
			name:    "required entry without job",
			queue:   "  required:\n    - { context: title-guard, workflow: title-guard.yml }\n",
			wantSub: "has an empty job",
		},
		{
			name: "duplicate context",
			queue: "  required:\n" +
				"    - { context: title-guard, workflow: a.yml, job: title-guard }\n" +
				"    - { context: title-guard, workflow: b.yml, job: title-guard }\n",
			wantSub: "declares context \"title-guard\" twice",
		},
		{
			name: "release job without branch",
			queue: "  required:\n    - { context: title-guard, workflow: title-guard.yml, job: title-guard }\n" +
				"  release_job:\n    workflow: release-batch.yml\n    job: release\n",
			wantSub: "has an empty branch",
		},
		{
			name: "bot job without job",
			queue: "  required:\n    - { context: title-guard, workflow: title-guard.yml, job: title-guard }\n" +
				"  bot_jobs:\n    - { workflow: cut-release.yml, branch: automation/cut-release }\n",
			wantSub: "has an empty job",
		},
		{
			name: "empty pre-tag step",
			queue: "  required:\n    - { context: title-guard, workflow: title-guard.yml, job: title-guard }\n" +
				"  release_job:\n    workflow: release-batch.yml\n    job: release\n    branch: automation/release\n    pre_tag_steps: [\"\"]\n",
			wantSub: "pre_tag_steps[0] is empty",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadWorkflowPolicy(writeQueuePolicyFile(t, base+tc.queue))
			if err == nil {
				t.Fatalf("expected a validation error for %q, got nil", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain expected substring %q", err.Error(), tc.wantSub)
			}
		})
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

func TestQueueRuleClosedSet(t *testing.T) {
	t.Parallel()

	if len(AllQueueRules) != 8 {
		t.Fatalf("AllQueueRules has %d members, want 8", len(AllQueueRules))
	}
	seen := map[QueueRule]bool{}
	for _, r := range AllQueueRules {
		if seen[r] {
			t.Fatalf("duplicate rule %q in AllQueueRules", r)
		}
		seen[r] = true
		if !r.IsValid() {
			t.Fatalf("rule %q fails its own IsValid", r)
		}
		if r.String() == "" {
			t.Fatalf("rule renders empty")
		}
	}
	if (QueueRule("no-such-rule")).IsValid() {
		t.Fatalf("unknown rule passes IsValid")
	}
}

func TestCheckWorkflows_LegacyPassesEmpty(t *testing.T) {
	t.Parallel()

	// The entry-point contract for jobs-only policies: no queue section means
	// no diagnostics and no required contexts.
	diags, required := CheckWorkflows(WorkflowPolicy{Jobs: []JobRule{{Name: "guard"}}}, t.TempDir())
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
		"ci.yml":          "name: CI\non:\n  pull_request:\n  merge_group:\nconcurrency:\n  group: ci-${{ github.ref }}\njobs:\n  merge-gate:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo gate\n",
		"title-guard.yml": "name: Title guard\non:\n  pull_request:\n  merge_group:\njobs:\n  title-guard:\n    runs-on: ubuntu-latest\n    steps:\n      - run: release-intent check-title --base x --head y --author z\n",
	}
	for name, content := range workflows {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write workflow fixture: %v", err)
		}
	}
	diags, required := CheckWorkflows(policy, dir)
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
