package release

import (
	"strings"
	"testing"
)

// L3 pins the legacy contract: today's jobs-only policies keep working
// unchanged. The repo's own release-guard policy and release workflow load
// and validate through the original single-file path, and the new
// directory checker treats a queue-less policy as constraining nothing.

func TestLegacyOwnPolicyAndWorkflowKeepWorking(t *testing.T) {
	t.Parallel()

	// The repo's own files, read from the working tree: the policy loads,
	// the release workflow satisfies it, and the directory checker yields
	// no diagnostics and no required contexts for a policy without a queue
	// section.
	policy, err := LoadWorkflowPolicy("../../.github/release-guard.policy.yml")
	if err != nil {
		t.Fatalf("LoadWorkflowPolicy(repo policy) failed: %v", err)
	}
	if err := CheckReleaseWorkflowFile("../../.github/workflows/release.yml", policy); err != nil {
		t.Fatalf("repo release.yml stopped satisfying the repo policy: %v", err)
	}
	diags, required := CheckWorkflows(policy, loadBaseConfig(t), "../../.github/workflows")
	if len(diags) != 0 {
		msgs := make([]string, 0, len(diags))
		for _, d := range diags {
			msgs = append(msgs, d.Message)
		}
		t.Fatalf("queue-less repo policy yielded %d queue diagnostics:\n%s", len(diags), strings.Join(msgs, "\n"))
	}
	if len(required) != 0 {
		t.Fatalf("queue-less repo policy yielded %d required contexts, want none", len(required))
	}
}

func TestLegacyQueuePolicyIgnoresQueueInSingleFileCheck(t *testing.T) {
	t.Parallel()

	// The original single-file checker reads only the jobs section: adding
	// a queue section must not change what it accepts or rejects.
	policy, err := LoadWorkflowPolicy(writeQueuePolicyFile(t, queuePolicyYAML))
	if err != nil {
		t.Fatalf("LoadWorkflowPolicy(queue policy) failed: %v", err)
	}
	if policy.Queue == nil {
		t.Fatalf("queue policy loaded with a nil queue section")
	}
	if err := checkReleaseWorkflow("release.yml", readWorkflowFixture(t, "schema-release.yml"), policy); err != nil {
		t.Fatalf("single-file check of a queue policy rejected the schema shape: %v", err)
	}
}
