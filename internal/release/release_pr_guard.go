package release

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// This guard pins the release-PR concurrency contract: a body edit must queue
// behind the running push run instead of cancelling it. The workflow triggers
// on `edited` (to revalidate the title) with the same concurrency group as the
// push run, so `cancel-in-progress: true` lets an edit on an open release PR
// cancel the in-progress tests while the edited run skips them (tests only run
// on opened/reopened/synchronize), and lets an edit after merge cancel the tag
// job. The expression below exempts `edited` runs from cancelling, matching
// peasant's release-pr.yml, while pushes still supersede stale runs.

// CheckReleasePRWorkflowFile validates the release-PR workflow at
// workflowPath: the pull_request trigger keeps `edited`, and
// concurrency.cancel-in-progress exempts `edited` instead of cancelling
// in-progress runs.
func CheckReleasePRWorkflowFile(workflowPath string) error {
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		return fmt.Errorf("check release-pr workflow: cannot read %s during release-PR validation: %w. Fix the path or run from the repository root", workflowPath, err)
	}
	return checkReleasePRWorkflow(workflowPath, data)
}

// checkReleasePRWorkflow validates parsed release-PR workflow YAML: the
// pull_request trigger keeps `edited` (title revalidation), and
// concurrency.cancel-in-progress is an expression that exempts `edited` so a
// body edit queues behind the running push run instead of cancelling it.
func checkReleasePRWorkflow(path string, data []byte) error {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("check release-pr workflow: cannot parse %s as YAML during release-PR validation: %w. Fix the workflow syntax before the release-PR gate can be trusted", path, err)
	}
	if len(root.Content) == 0 {
		return fmt.Errorf("check release-pr workflow: %s is empty during release-PR validation. Restore the release-PR workflow before the gate can be trusted", path)
	}

	on := mappingValue(root.Content[0], "on")
	if on == nil || on.Kind != yaml.MappingNode {
		return fmt.Errorf("check release-pr workflow: %s has no on.pull_request trigger during release-PR validation. Restore the pull_request trigger (with edited) so the title is still revalidated on edit", path)
	}
	pr := mappingValue(on, "pull_request")
	if pr == nil || pr.Kind != yaml.MappingNode {
		return fmt.Errorf("check release-pr workflow: %s has no on.pull_request trigger during release-PR validation. Restore the pull_request trigger (with edited) so the title is still revalidated on edit", path)
	}
	types := mappingValue(pr, "types")
	if types == nil || types.Kind != yaml.SequenceNode {
		return fmt.Errorf("check release-pr workflow: %s on.pull_request has no types sequence during release-PR validation. Restore the types list (with edited) so the title is still revalidated on edit", path)
	}
	var haveEdited bool
	for _, item := range types.Content {
		if item.Kind == yaml.ScalarNode && item.Value == "edited" {
			haveEdited = true
			break
		}
	}
	if !haveEdited {
		return fmt.Errorf("check release-pr workflow: %s on.pull_request.types is missing edited during release-PR validation. Keep edited in the trigger types so the title is still revalidated on edit", path)
	}

	concurrency := mappingValue(root.Content[0], "concurrency")
	if concurrency == nil || concurrency.Kind != yaml.MappingNode {
		return fmt.Errorf("check release-pr workflow: %s has no concurrency mapping during release-PR validation. Restore the concurrency group with cancel-in-progress exempting edited so a body edit queues behind the running push run", path)
	}
	cancel := mappingValue(concurrency, "cancel-in-progress")
	if cancel == nil || cancel.Kind != yaml.ScalarNode {
		return fmt.Errorf("check release-pr workflow: %s is missing concurrency.cancel-in-progress during release-PR validation. Set cancel-in-progress: ${{ github.event.action != 'edited' }} so a body edit queues behind the running push run instead of cancelling it", path)
	}
	// A literal boolean reintroduces the regression: an edit on an open
	// release PR cancels the in-progress tests (the edited run skips them),
	// and an edit after merge can cancel the tag job.
	if cancel.Tag == "!!bool" || cancel.Value == "true" || cancel.Value == "false" {
		return fmt.Errorf("check release-pr workflow: %s concurrency.cancel-in-progress is %s during release-PR validation. A literal boolean lets an edit cancel in-progress tests (the edited run skips them) and lets an edit after merge cancel tag; use ${{ github.event.action != 'edited' }} so a body edit queues behind the running push run", path, cancel.Value)
	}
	expr := cancel.Value
	if !(strings.Contains(expr, "github.event.action") && strings.Contains(expr, "edited") && strings.Contains(expr, "!=")) {
		return fmt.Errorf("check release-pr workflow: %s concurrency.cancel-in-progress is %q during release-PR validation. It must exempt edited (${{ github.event.action != 'edited' }}) so a body edit queues behind the running push run instead of cancelling it", path, expr)
	}
	return nil
}
