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

// This file is the repo-agnostic, policy-driven release-workflow guard. Each
// consumer repo (schema, peasant, …) ships a .github/release-guard.policy.yml
// describing its own job graph; the shared tool holds NO repo-specific job-shape
// knowledge. The WorkflowPolicy descriptor is a precise data projection of the
// assertions that USED to be hardcoded per repo (schema: guard / nix-vendor-hash
// / contract-gates / GitHub-Release publish; peasant: goreleaser behind reusable
// e2e + release-e2e gates) — not a speculative DSL.

// WorkflowPolicy is the per-repo, declarative projection of the release-workflow
// assertions that were previously hardcoded in each repo's workflow_guard.go.
// It is data, not a DSL: each JobRule states a job that must exist, the
// needs-edges it must declare, and (optionally) the reusable-workflow shape or
// the job permissions/environment binding it must carry. schema and peasant
// supply different policy files; the shared tool holds no repo-specific
// job-shape knowledge.
type WorkflowPolicy struct {
	Jobs []JobRule `yaml:"jobs"`
}

// JobRule constrains a single workflow job. Name is required; Needs lists the
// required needs-edges; Reusable (when non-nil) requires the job to be a
// reusable-workflow gate matching the ReusableRule; Permissions (when non-nil)
// requires the job's own `permissions:` block to grant the scopes in
// PermissionsRule; Environment (when non-empty) requires the job's `environment:`
// to equal it - scoping which GitHub Actions environment (and therefore, for an
// OIDC-trusted-publishing job, which environment's protection rules) the job runs
// under.
//
// Permissions/Environment recognize only the exact forms this repo's workflows
// use - the explicit `permissions:` map (not the `read-all`/`write-all` bare
// scalar shorthand) and a bare scalar `environment:` (not the `{name, url}`
// mapping form). GitHub accepts all of those forms; this checker intentionally
// does not parse the unrecognized ones (rather than silently mis-accept or
// mis-reject them) and instead fails closed with an error naming the
// unsupported form, per checkJobPermissionsAgainstRule /
// checkJobEnvironmentAgainstRule below.
type JobRule struct {
	Name        string           `yaml:"name"`
	Needs       []string         `yaml:"needs"`
	Reusable    *ReusableRule    `yaml:"reusable"`
	Permissions *PermissionsRule `yaml:"permissions"`
	Environment string           `yaml:"environment"`
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

// PermissionsRule constrains a job's own `permissions:` block. IDToken (when
// true) requires `permissions.id-token: write` on the job - the scope an OIDC
// trusted-publishing step (e.g. `pnpm publish` to npm) needs to mint its token.
// Job-level `permissions:` REPLACES the workflow-level default entirely (GitHub
// Actions does not merge the two), so this checks the job's own block, not the
// workflow top level.
type PermissionsRule struct {
	IDToken bool `yaml:"idToken"`
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
		return WorkflowPolicy{}, fmt.Errorf("load workflow policy: cannot parse %s as a release-guard policy: %w. Expected a top-level 'jobs:' sequence of {name, needs, reusable, permissions, environment} entries", path, err)
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

// CheckReleaseWorkflowFile validates the workflow file at workflowPath against
// the supplied per-repo WorkflowPolicy: every JobRule's job must exist, declare
// the required needs-edges, and (for reusable gates or a required permission
// scope) match the ReusableRule / PermissionsRule.
func CheckReleaseWorkflowFile(workflowPath string, policy WorkflowPolicy) error {
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		return fmt.Errorf("check workflow: cannot read %s during release workflow validation: %w. Fix the path or run from the repository root", workflowPath, err)
	}
	return checkReleaseWorkflow(workflowPath, data, policy)
}

// checkReleaseWorkflow validates parsed workflow YAML against policy.
// It reuses the shared mappingValue/needsList/requireNeeds helpers so the
// assertions are byte-equivalent to the hardcoded checks they generalize.
func checkReleaseWorkflow(path string, data []byte, policy WorkflowPolicy) error {
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
		if rule.Permissions != nil {
			if err := checkJobPermissionsAgainstRule(path, rule.Name, job, rule.Permissions); err != nil {
				return err
			}
		}
		if rule.Environment != "" {
			if err := checkJobEnvironmentAgainstRule(path, rule.Name, job, rule.Environment); err != nil {
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

// checkJobPermissionsAgainstRule asserts a job's OWN `permissions:` block
// grants the scopes rule requires. Job-level `permissions:` replaces the
// workflow-level default rather than merging with it, so this reads only
// jobs.<name>.permissions, never the workflow top level.
//
// Only the explicit map form (`permissions: { id-token: write, ... }`) is
// parsed. GitHub also accepts a bare `permissions: read-all` / `write-all`
// shorthand that implicitly grants every scope including id-token: write; that
// form is deliberately NOT recognized (rather than silently mis-detected as
// either granting or lacking the scope) - a job using it is rejected with a
// message naming the shorthand and the fix, distinct from the "missing
// entirely" message.
func checkJobPermissionsAgainstRule(path, jobName string, job *yaml.Node, rule *PermissionsRule) error {
	if !rule.IDToken {
		return nil
	}
	permsNode := mappingValue(job, "permissions")
	if permsNode != nil && permsNode.Kind != yaml.MappingNode {
		return fmt.Errorf("check workflow: %s jobs.%s permissions is the shorthand scalar %q during release workflow validation. This checker only parses the explicit map form; rewrite jobs.%s.permissions to { id-token: write, ... } (even though GitHub itself also accepts the %q shorthand here) so the required id-token scope is checkable", path, jobName, permsNode.Value, jobName, permsNode.Value)
	}
	idToken := mappingValue(permsNode, "id-token")
	if idToken == nil || idToken.Kind != yaml.ScalarNode || idToken.Value != "write" {
		got := "<missing>"
		if idToken != nil {
			got = idToken.Value
		}
		return fmt.Errorf("check workflow: %s jobs.%s permissions.id-token is %s during release workflow validation. OIDC trusted publishing needs an id-token: write permission scoped to this job to mint its short-lived token; add permissions: { id-token: write } (alongside any other scopes the job needs) to jobs.%s", path, jobName, got, jobName)
	}
	return nil
}

// checkJobEnvironmentAgainstRule asserts a job's `environment:` scalar equals
// want. Used to keep an OIDC-trusted-publishing job bound to the GitHub
// Actions environment its npm Trusted Publisher registration is scoped to.
//
// Only the bare scalar form (`environment: npm-publish`) is parsed. GitHub
// also accepts a mapping form (`environment: { name: npm-publish, url: ... }`)
// that surfaces a "View deployment" link in the Actions UI; that form is
// deliberately NOT recognized (rather than silently mis-detected) - a job
// using it is rejected with a message naming the mapping form and the fix,
// distinct from the "missing entirely" message.
func checkJobEnvironmentAgainstRule(path, jobName string, job *yaml.Node, want string) error {
	env := mappingValue(job, "environment")
	if env != nil && env.Kind == yaml.MappingNode {
		return fmt.Errorf("check workflow: %s jobs.%s environment is the {name, url} mapping form during release workflow validation. This checker only parses the bare scalar form; rewrite jobs.%s.environment to the scalar %s (even though GitHub itself also accepts an environment: { name: ..., url: ... } mapping here, e.g. to surface a deployment link) so the required environment binding is checkable", path, jobName, jobName, want)
	}
	if env == nil || env.Kind != yaml.ScalarNode || env.Value != want {
		got := "<missing>"
		if env != nil {
			got = env.Value
		}
		return fmt.Errorf("check workflow: %s jobs.%s environment is %s during release workflow validation. Set environment: %s so this job stays bound to the GitHub Actions environment its trusted-publisher registration expects", path, jobName, got, want)
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
