package main

import (
	"fmt"

	"github.com/peasant-labs/schema/internal/release"
)

// runCheckWorkflows implements:
// release-intent check-workflows --policy P --dir .github/workflows [--print-required]
//
// It delegates to the policy package's exported checker.
func runCheckWorkflows(args []string) error {
	policyPath := ""
	dir := ""
	printRequired := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--policy":
			val, next, err := requireFlagValue(args, i, "--policy")
			if err != nil {
				return err
			}
			policyPath = val
			i = next
		case "--dir":
			val, next, err := requireFlagValue(args, i, "--dir")
			if err != nil {
				return err
			}
			dir = val
			i = next
		case "--print-required":
			printRequired = true
		default:
			return usageErr("unknown flag %q for check-workflows: expected --policy P --dir DIR [--print-required]", args[i])
		}
	}
	if policyPath == "" {
		return usageErr("usage: release-intent check-workflows --policy P --dir DIR [--print-required] (missing required --policy): pass the policy file path, e.g. .github/release-guard.policy.yml")
	}
	if dir == "" {
		return usageErr("usage: release-intent check-workflows --policy P --dir DIR [--print-required] (missing required --dir): pass the workflows directory, e.g. .github/workflows")
	}
	policy, err := release.LoadWorkflowPolicy(policyPath)
	if err != nil {
		return err
	}
	diagnostics, required := release.CheckWorkflows(policy, dir)
	if printRequired {
		if len(required) == 0 {
			return emitLines([]string{"required="})
		}
		lines := make([]string, 0, len(required))
		for _, r := range required {
			lines = append(lines, "required="+r.Context)
		}
		return emitLines(lines)
	}
	if len(diagnostics) > 0 {
		msg := diagnostics[0].Message
		if msg == "" {
			msg = fmt.Sprintf("workflow policy violation in %s", diagnostics[0].Workflow)
		}
		return fmt.Errorf("check-workflows: policy violations in %q: %s. Fix: align the workflows with the policy file", policyPath, msg)
	}
	return emitLines([]string{"ok=true"})
}
