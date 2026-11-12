package release

import (
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

// The merge-queue policy corpus drives the production CheckWorkflows over one
// materialized tree per case: base plus the case overlay. The policy and the
// release config load through the production loaders, and the directory is a
// real temp dir the production reader lists. An accept case must yield no
// diagnostics; a reject case must trip only its expected closed rule, with
// every expected fragment in the diagnostics.

func TestCheckWorkflows_Cases(t *testing.T) {
	t.Parallel()

	fx := loadQueueWorkflowFixtures(t)
	corpus := fx.corpus()
	assert.RequireValid(t, corpus)

	for _, row := range fx.Rows {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			t.Parallel()
			dir := materializeQueueFixture(t, row.Input.Overlay)
			policy, err := LoadWorkflowPolicy(dir + "/policy.yml")
			if err != nil {
				t.Fatalf("case %q policy does not load: %v", row.Name, err)
			}
			cfg, err := intent.LoadConfig(dir + "/release-intent.yaml")
			if err != nil {
				t.Fatalf("case %q release config does not load: %v", row.Name, err)
			}
			diags, required := CheckWorkflows(policy, cfg, dir+"/workflows")
			assertQueueRule(t, row, policy, diags, required)
		})
	}
}

// assertQueueRule checks one corpus case against its classification: an accept
// case passes only with zero diagnostics and the policy's required contexts
// back in order; a reject case passes only when every diagnostic carries the
// expected rule and every expected fragment appears in the diagnostics.
func assertQueueRule(t *testing.T, row queueCaseRow, policy WorkflowPolicy, diags []Diagnostic, required []RequiredContext) {
	t.Helper()
	if row.Classification == testcase.MustPass {
		if len(diags) != 0 {
			msgs := make([]string, 0, len(diags))
			for _, d := range diags {
				msgs = append(msgs, d.Message)
			}
			t.Fatalf("accept case %q yielded %d diagnostics:\n%s", row.Name, len(diags), strings.Join(msgs, "\n"))
		}
		if len(required) != len(policy.Queue.Required) {
			t.Fatalf("accept case %q returned %d required contexts, want the %d policy entries in order", row.Name, len(required), len(policy.Queue.Required))
		}
		for i := range required {
			if required[i] != policy.Queue.Required[i] {
				t.Fatalf("accept case %q required[%d] = %+v, want %+v", row.Name, i, required[i], policy.Queue.Required[i])
			}
		}
		return
	}
	rule := QueueRule(row.Expected.Rule)
	if len(row.Expected.WantContains) == 0 {
		t.Fatalf("reject case %q declares no want_contains fragments (fixture error)", row.Name)
	}
	var msgs []string
	for _, d := range diags {
		msgs = append(msgs, d.Message)
		if d.Rule != rule {
			t.Fatalf("reject case %q tripped rule %s, want only %s:\n%s", row.Name, d.Rule, rule, d.Message)
		}
	}
	if len(diags) == 0 {
		t.Fatalf("reject case %q tripped no %s diagnostic", row.Name, rule)
	}
	joined := strings.Join(msgs, "\n")
	for _, want := range row.Expected.WantContains {
		if !strings.Contains(joined, want) {
			t.Fatalf("reject case %q diagnostics do not contain %q:\n%s", row.Name, want, joined)
		}
	}
}

// TestCheckWorkflows_Coverage pins the load-bearing behavior per rule: every
// closed rule needs at least one accept and one reject sample, so a
// count-preserving swap that drops a rule's real sample reddens.
func TestCheckWorkflows_Coverage(t *testing.T) {
	t.Parallel()

	fx := loadQueueWorkflowFixtures(t)
	passRules := map[QueueRule]bool{}
	failRules := map[QueueRule]bool{}
	for _, row := range fx.Rows {
		if row.Classification == testcase.MustPass {
			passRules[QueueRule(row.Expected.Rule)] = true
			continue
		}
		failRules[QueueRule(row.Expected.Rule)] = true
	}
	for _, r := range AllQueueRules {
		if !passRules[r] {
			t.Errorf("no accept case samples rule %s", r)
		}
		if !failRules[r] {
			t.Errorf("no reject case samples rule %s", r)
		}
	}
}

// TestCheckWorkflows_NoOrphanDirs fails when an overlay directory has no
// corpus case or a case names a missing overlay, so samples cannot hide
// outside the manifest.
func TestCheckWorkflows_NoOrphanDirs(t *testing.T) {
	t.Parallel()

	fx := loadQueueWorkflowFixtures(t)
	byOverlay := map[string]bool{}
	for _, row := range fx.Rows {
		if row.Input.Overlay != "" {
			byOverlay[row.Input.Overlay] = true
		}
	}
	entries, err := workflowQueueFixtureFS.ReadDir("testdata/workflow_queue/overlays")
	if err != nil {
		t.Fatalf("list queue overlays: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			t.Errorf("overlay entry %q is not a directory", e.Name())
			continue
		}
		if !byOverlay[e.Name()] {
			t.Errorf("overlay directory %q has no corpus case (add it to cases.yaml or remove it)", e.Name())
		}
	}
	for _, row := range fx.Rows {
		if row.Input.Overlay == "" {
			continue
		}
		if _, err := workflowQueueFixtureFS.ReadDir("testdata/workflow_queue/overlays/" + row.Input.Overlay); err != nil {
			t.Errorf("case %q names missing overlay %q", row.Name, row.Input.Overlay)
		}
	}
}
