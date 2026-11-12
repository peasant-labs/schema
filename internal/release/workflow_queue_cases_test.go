package release

import (
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

// The merge-queue policy corpus drives the production CheckWorkflows over one
// materialized directory per case: the policy loads through the production
// LoadWorkflowPolicy, and the directory is a real temp dir the production
// reader lists. Accept cases must yield no diagnostics; reject cases must trip
// the expected closed rule with every expected fragment in the diagnostics.

func TestCheckWorkflows_Cases(t *testing.T) {
	t.Parallel()

	fx := loadQueueWorkflowFixtures(t)
	corpus := fx.corpus()
	assert.RequireMin(t, corpus, 28)
	assert.RequireValid(t, corpus)

	for _, c := range fx.corpus().Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			dir := copyQueueFixture(t, c.Input.Dir)
			policy, err := LoadWorkflowPolicy(dir + "/policy.yml")
			if err != nil {
				t.Fatalf("case %q policy does not load: %v", c.Name, err)
			}
			diags, required := CheckWorkflows(policy, dir+"/workflows")
			assertQueueRule(t, c, policy, diags, required)
		})
	}
}

// assertQueueRule checks one corpus case against its classification: an
// accept case passes only with zero diagnostics and the policy's required
// contexts back in order; a reject case passes only when the expected closed
// rule fires and every expected fragment appears in the diagnostics.
func assertQueueRule(t *testing.T, c testcase.Case[queueCaseInput, queueCaseExpected], policy WorkflowPolicy, diags []Diagnostic, required []RequiredContext) {
	t.Helper()
	if c.Classification == testcase.MustPass {
		if len(diags) != 0 {
			msgs := make([]string, 0, len(diags))
			for _, d := range diags {
				msgs = append(msgs, d.Message)
			}
			t.Fatalf("accept case %q yielded %d diagnostics:\n%s", c.Name, len(diags), strings.Join(msgs, "\n"))
		}
		if len(required) != len(policy.Queue.Required) {
			t.Fatalf("accept case %q returned %d required contexts, want the %d policy entries in order", c.Name, len(required), len(policy.Queue.Required))
		}
		for i := range required {
			if required[i] != policy.Queue.Required[i] {
				t.Fatalf("accept case %q required[%d] = %+v, want %+v", c.Name, i, required[i], policy.Queue.Required[i])
			}
		}
		return
	}
	rule := QueueRule(c.Expected.Rule)
	if !rule.IsValid() {
		t.Fatalf("reject case %q names unknown rule %q (fixture error: not a closed member)", c.Name, c.Expected.Rule)
	}
	if len(c.Expected.WantContains) == 0 {
		t.Fatalf("reject case %q declares no want_contains fragments (fixture error)", c.Name)
	}
	fired := false
	var msgs []string
	for _, d := range diags {
		msgs = append(msgs, d.Message)
		if d.Rule == rule {
			fired = true
		}
	}
	if !fired {
		t.Fatalf("reject case %q tripped no %s diagnostic (got %d diagnostics):\n%s", c.Name, rule, len(diags), strings.Join(msgs, "\n"))
	}
	joined := strings.Join(msgs, "\n")
	for _, want := range c.Expected.WantContains {
		if !strings.Contains(joined, want) {
			t.Fatalf("reject case %q diagnostics do not contain %q:\n%s", c.Name, want, joined)
		}
	}
}

// TestCheckWorkflows_Coverage pins the load-bearing behavior per rule: every
// closed rule needs at least one accept and one reject sample, so a
// count-preserving swap that drops a rule's real sample reddens even at a
// fixed count.
func TestCheckWorkflows_Coverage(t *testing.T) {
	t.Parallel()

	fx := loadQueueWorkflowFixtures(t)
	passRules := map[QueueRule]bool{}
	failRules := map[QueueRule]bool{}
	for _, c := range fx.corpus().Cases {
		if c.Classification == testcase.MustPass {
			// An accept case exercises every rule at once (the run checks all
			// eight), so it counts as the accept sample for each member.
			for _, r := range AllQueueRules {
				passRules[r] = true
			}
			continue
		}
		failRules[QueueRule(c.Expected.Rule)] = true
	}
	for _, r := range AllQueueRules {
		if !passRules[r] {
			t.Errorf("no accept sample covers rule %s", r)
		}
		if !failRules[r] {
			t.Errorf("no reject sample covers rule %s", r)
		}
	}
}

// TestCheckWorkflows_NoOrphanDirs fails when a case directory has no corpus
// case or a case names a missing directory, so samples cannot hide outside
// the manifest.
func TestCheckWorkflows_NoOrphanDirs(t *testing.T) {
	t.Parallel()

	fx := loadQueueWorkflowFixtures(t)
	byDir := map[string]bool{}
	for _, c := range fx.corpus().Cases {
		byDir[c.Input.Dir] = true
	}
	entries, err := workflowQueueFixtureFS.ReadDir("testdata/workflow_queue")
	if err != nil {
		t.Fatalf("list queue fixtures: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if !byDir[e.Name()] {
			t.Errorf("fixture directory %q has no corpus case (add it to cases.yaml or remove it)", e.Name())
		}
		if _, err := workflowQueueFixtureFS.ReadFile("testdata/workflow_queue/" + e.Name() + "/policy.yml"); err != nil {
			t.Errorf("fixture directory %q has no policy.yml: %v", e.Name(), err)
		}
	}
	for _, c := range fx.corpus().Cases {
		if _, err := workflowQueueFixtureFS.ReadFile("testdata/workflow_queue/" + c.Input.Dir + "/policy.yml"); err != nil {
			t.Errorf("case %q names missing directory %q", c.Name, c.Input.Dir)
		}
	}
}
