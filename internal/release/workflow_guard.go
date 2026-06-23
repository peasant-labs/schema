package release

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

// --- Policy-driven (generalized) release-workflow guard ---------------------
//
// The block above (CheckReleaseWorkflowFile / CheckReleaseWorkflow) hardcodes
// the schema repo's own job graph. The descriptor + functions below are the
// repo-agnostic successor: the same YAML-parse/assert logic, parameterized by a
// per-repo WorkflowPolicy loaded from .github/release-guard.policy.yml. They are
// ADDITIVE for SLICE-1 — the hardcoded entry point stays live (cmd/release-guard
// and `make check` still call it) until SLICE-4 cuts main.go over to the policy
// API and removes the old path.

// WorkflowPolicy is the per-repo, declarative projection of the release-workflow
// assertions that were previously hardcoded in each repo's workflow_guard.go.
// It is data, not a DSL: each JobRule states a job that must exist, the
// needs-edges it must declare, and (optionally) the reusable-workflow shape it
// must take. schema and peasant supply different policy files; the shared tool
// holds no repo-specific job-shape knowledge.
type WorkflowPolicy struct {
	Jobs []JobRule `yaml:"jobs"`
}

// JobRule constrains a single workflow job. Name is required; Needs lists the
// required needs-edges; Reusable (when non-nil) requires the job to be a
// reusable-workflow gate matching the ReusableRule.
type JobRule struct {
	Name     string        `yaml:"name"`
	Needs    []string      `yaml:"needs"`
	Reusable *ReusableRule `yaml:"reusable"`
}

// ReusableRule constrains a reusable-workflow gate job (peasant's e2e /
// release-e2e shape): it must `uses:` the named reusable workflow, pass
// `secrets: inherit` (when SecretsInherit), and carry no `if:` condition (when
// ForbidIf) so the gate runs on every release tag rather than selected paths.
type ReusableRule struct {
	Uses           string `yaml:"uses"`
	SecretsInherit bool   `yaml:"secretsInherit"`
	ForbidIf       bool   `yaml:"forbidIf"`
}

// LoadWorkflowPolicy reads and parses a repo's release-guard policy file
// (.github/release-guard.policy.yml) into a WorkflowPolicy. Unknown fields are
// rejected so a typo'd policy key fails loudly rather than silently disabling a
// gate.
func LoadWorkflowPolicy(path string) (WorkflowPolicy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return WorkflowPolicy{}, fmt.Errorf("load workflow policy: cannot read %s during release workflow validation: %w. Create the release-guard policy file (.github/release-guard.policy.yml) or fix the --policy path", path, err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var policy WorkflowPolicy
	if err := dec.Decode(&policy); err != nil && !errors.Is(err, io.EOF) {
		return WorkflowPolicy{}, fmt.Errorf("load workflow policy: cannot parse %s as a release-guard policy: %w. Expected a top-level 'jobs:' sequence of {name, needs, reusable} entries", path, err)
	}
	if len(policy.Jobs) == 0 {
		return WorkflowPolicy{}, fmt.Errorf("load workflow policy: %s declares no jobs during release workflow validation. A policy must list at least the publish job and its required gates under 'jobs:'", path)
	}
	for i, j := range policy.Jobs {
		if j.Name == "" {
			return WorkflowPolicy{}, fmt.Errorf("load workflow policy: %s jobs[%d] has an empty name during release workflow validation. Every policy job entry must set 'name:' to the workflow job it constrains", path, i)
		}
	}
	return policy, nil
}

// CheckReleaseWorkflowFileWithPolicy is the generalized, policy-driven successor
// to CheckReleaseWorkflowFile: it validates the workflow file at workflowPath
// against the supplied per-repo WorkflowPolicy. SLICE-4 renames this to
// CheckReleaseWorkflowFile (and drops the hardcoded pair) once main.go is cut
// over; it carries the transitional name here only to avoid colliding with the
// still-live hardcoded entry point.
func CheckReleaseWorkflowFileWithPolicy(workflowPath string, policy WorkflowPolicy) error {
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		return fmt.Errorf("check workflow: cannot read %s during release workflow validation: %w. Fix the path or run from the repository root", workflowPath, err)
	}
	return checkReleaseWorkflowWithPolicy(workflowPath, data, policy)
}

// checkReleaseWorkflowWithPolicy validates parsed workflow YAML against policy.
// It reuses the shared mappingValue/needsList/requireNeeds helpers so the
// assertions are byte-equivalent to the hardcoded checks they generalize.
func checkReleaseWorkflowWithPolicy(path string, data []byte, policy WorkflowPolicy) error {
	if len(policy.Jobs) == 0 {
		return fmt.Errorf("check workflow: the policy for %s declares no jobs during release workflow validation. A release-guard policy must list at least the publish job and its required gates", path)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("check workflow: cannot parse %s as YAML during release workflow validation: %w. Fix the workflow syntax before publishing can be guarded", path, err)
	}
	if len(root.Content) == 0 {
		return fmt.Errorf("check workflow: %s is empty during release workflow validation. Restore the release workflow before publishing", path)
	}

	jobs := mappingValue(root.Content[0], "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return fmt.Errorf("check workflow: %s has no jobs mapping during release workflow validation. Add %s before publishing", path, policyJobNames(policy))
	}

	for _, rule := range policy.Jobs {
		job := mappingValue(jobs, rule.Name)
		if job == nil || job.Kind != yaml.MappingNode {
			return fmt.Errorf("check workflow: %s is missing jobs.%s during release workflow validation. The policy (.github/release-guard.policy.yml) requires this job; add it so the release pipeline matches the declared gates", path, rule.Name)
		}
		if rule.Reusable != nil {
			if err := checkReusableJobAgainstRule(path, rule.Name, job, rule.Reusable); err != nil {
				return err
			}
		}
		if len(rule.Needs) > 0 {
			if err := requireNeeds(path, "jobs."+rule.Name, job, rule.Needs); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkReusableJobAgainstRule asserts a reusable-workflow gate job matches rule:
// no `if:` (when ForbidIf), the exact `uses:` target, and `secrets: inherit`
// (when SecretsInherit). Mirrors peasant's checkReusableWorkflowJob byte-for-
// byte, parameterized by the rule.
func checkReusableJobAgainstRule(path, jobName string, job *yaml.Node, rule *ReusableRule) error {
	if rule.ForbidIf {
		if ifNode := mappingValue(job, "if"); ifNode != nil {
			return fmt.Errorf("check workflow: %s jobs.%s has an if condition %q during release workflow validation. Remove it so the gate runs on every release tag, not only on rc or selected paths", path, jobName, ifNode.Value)
		}
	}
	if rule.Uses != "" {
		if uses := mappingValue(job, "uses"); uses == nil || uses.Value != rule.Uses {
			got := "<missing>"
			if uses != nil {
				got = uses.Value
			}
			return fmt.Errorf("check workflow: %s jobs.%s uses %s during release workflow validation. Point it at %s so release tags reuse the required workflow", path, jobName, got, rule.Uses)
		}
	}
	if rule.SecretsInherit {
		if secrets := mappingValue(job, "secrets"); secrets == nil || secrets.Kind != yaml.ScalarNode || secrets.Value != "inherit" {
			got := "<missing>"
			if secrets != nil {
				got = secrets.Value
			}
			return fmt.Errorf("check workflow: %s jobs.%s secrets is %s during release workflow validation. Use secrets: inherit so the reusable %s workflow receives the credentials it needs", path, jobName, got, jobName)
		}
	}
	return nil
}

// policyJobNames renders the policy's job names as "jobs.a, jobs.b" for the
// no-jobs-mapping diagnostic.
func policyJobNames(policy WorkflowPolicy) string {
	parts := make([]string, 0, len(policy.Jobs))
	for _, j := range policy.Jobs {
		parts = append(parts, "jobs."+j.Name)
	}
	return strings.Join(parts, ", ")
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
