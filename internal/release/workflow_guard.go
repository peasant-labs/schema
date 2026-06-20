package release

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// This is the schema repo's OWN release-workflow guard. It deliberately replaces
// peasant's internal/release/workflow_guard.go (which encoded peasant's
// goreleaser + full-stack-e2e + installed-package-e2e release.yml shape and
// imported peasant's internal/defaults). PROPOSAL-4 W4 SUBTRACTS all of those
// CLI-binary stages: the schema module ships NO binary, so it has NO goreleaser,
// no e2e, no release-e2e. Instead its release.yml publishes a GitHub Release with
// the OpenAPI specs as assets, behind the CONTRACT gates.
//
// The job names below are the schema release pipeline's:
const (
	// releaseWorkflowGuardJob parses/classifies the tag and applies the
	// final-requires-green-ancestor-rc rule.
	releaseWorkflowGuardJob = "guard"
	// releaseWorkflowNixHashJob asserts the flake vendorHash is current at the
	// tagged commit.
	releaseWorkflowNixHashJob = "nix-vendor-hash"
	// releaseWorkflowGatesJob runs the contract gates (oasdiff / go-apidiff /
	// vacuum) — the schema repo's CI value-add that replaces peasant's e2e gates.
	releaseWorkflowGatesJob = "contract-gates"
	// releaseWorkflowPublishJob creates the GitHub Release and uploads the
	// generated/ OpenAPI specs as assets.
	releaseWorkflowPublishJob = "release"
)

var (
	// The contract-gates job must run after the guard (so a malformed/non-schema
	// tag never reaches the gates).
	requiredGatesNeeds = []string{releaseWorkflowGuardJob}
	// The publish job must sit behind the guard, the vendorHash freshness gate,
	// AND the contract gates — nothing publishes a release that hasn't proven the
	// contract is non-breaking and the source tree is hash-current.
	requiredPublishNeeds = []string{releaseWorkflowGuardJob, releaseWorkflowNixHashJob, releaseWorkflowGatesJob}
)

// CheckReleaseWorkflowFile validates the release workflow's publication gates.
// It is intentionally narrow: the workflow may evolve, but the GitHub-Release
// publish job must stay behind the guard, the Nix vendorHash gate, and the
// contract gates.
func CheckReleaseWorkflowFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("check workflow: cannot read %s during release workflow validation: %w. Fix the path or run from the repository root", path, err)
	}
	return CheckReleaseWorkflow(path, data)
}

// CheckReleaseWorkflow validates release.yml from parsed YAML bytes.
func CheckReleaseWorkflow(path string, data []byte) error {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("check workflow: cannot parse %s as YAML during release workflow validation: %w. Fix the workflow syntax before publishing can be guarded", path, err)
	}
	if len(root.Content) == 0 {
		return fmt.Errorf("check workflow: %s is empty during release workflow validation. Restore the release workflow before publishing", path)
	}

	jobs := mappingValue(root.Content[0], "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return fmt.Errorf("check workflow: %s has no jobs mapping during release workflow validation. Add jobs.%s, jobs.%s, jobs.%s, and jobs.%s before publishing",
			path, releaseWorkflowGuardJob, releaseWorkflowNixHashJob, releaseWorkflowGatesJob, releaseWorkflowPublishJob)
	}

	// The guard job must exist (the entry gate).
	if guardJob := mappingValue(jobs, releaseWorkflowGuardJob); guardJob == nil || guardJob.Kind != yaml.MappingNode {
		return fmt.Errorf("check workflow: %s is missing jobs.%s. Add the tag guard (parse-tag + final-requires-ancestor-rc) so a malformed or non-schema tag never publishes a release",
			path, releaseWorkflowGuardJob)
	}

	// The vendorHash freshness job must exist.
	if hashJob := mappingValue(jobs, releaseWorkflowNixHashJob); hashJob == nil || hashJob.Kind != yaml.MappingNode {
		return fmt.Errorf("check workflow: %s is missing jobs.%s. Add the flake-vendorHash freshness gate so a tag is never published from a hash-stale source tree",
			path, releaseWorkflowNixHashJob)
	}

	// The contract-gates job must exist and run after the guard.
	gatesJob := mappingValue(jobs, releaseWorkflowGatesJob)
	if gatesJob == nil || gatesJob.Kind != yaml.MappingNode {
		return fmt.Errorf("check workflow: %s is missing jobs.%s. Add the contract gates (oasdiff / go-apidiff / vacuum) so a breaking contract change cannot publish a release",
			path, releaseWorkflowGatesJob)
	}
	if err := requireNeeds(path, "jobs."+releaseWorkflowGatesJob, gatesJob, requiredGatesNeeds); err != nil {
		return err
	}

	// The publish job must exist and sit behind all required gates.
	releaseJob := mappingValue(jobs, releaseWorkflowPublishJob)
	if releaseJob == nil || releaseJob.Kind != yaml.MappingNode {
		return fmt.Errorf("check workflow: %s is missing jobs.%s. Restore the GitHub-Release publish job (uploads the generated/ OpenAPI specs as assets) behind the required gates",
			path, releaseWorkflowPublishJob)
	}
	if err := requireNeeds(path, "jobs."+releaseWorkflowPublishJob, releaseJob, requiredPublishNeeds); err != nil {
		return err
	}

	return nil
}

func requireNeeds(path, jobName string, job *yaml.Node, required []string) error {
	needs, err := needsList(job)
	if err != nil {
		return fmt.Errorf("check workflow: %s %s has invalid needs during release workflow validation: %w. Use a scalar or YAML sequence of job names", path, jobName, err)
	}
	have := make(map[string]bool, len(needs))
	for _, need := range needs {
		have[need] = true
	}
	var missing []string
	for _, need := range required {
		if !have[need] {
			missing = append(missing, need)
		}
	}
	if len(missing) > 0 {
		sort.Strings(needs)
		return fmt.Errorf("check workflow: %s %s.needs is missing %s during release workflow validation; current needs are [%s]. Add the missing gate(s) so the release publish cannot run before all required contract checks pass",
			path, jobName, strings.Join(missing, ", "), strings.Join(needs, ", "))
	}
	return nil
}

func needsList(job *yaml.Node) ([]string, error) {
	needs := mappingValue(job, "needs")
	if needs == nil {
		return nil, nil
	}
	switch needs.Kind {
	case yaml.ScalarNode:
		if needs.Value == "" {
			return nil, fmt.Errorf("empty scalar")
		}
		return []string{needs.Value}, nil
	case yaml.SequenceNode:
		values := make([]string, 0, len(needs.Content))
		for _, node := range needs.Content {
			if node.Kind != yaml.ScalarNode || node.Value == "" {
				return nil, fmt.Errorf("sequence contains a non-scalar or empty item")
			}
			values = append(values, node.Value)
		}
		return values, nil
	default:
		return nil, fmt.Errorf("expected scalar or sequence, got YAML kind %d", needs.Kind)
	}
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
