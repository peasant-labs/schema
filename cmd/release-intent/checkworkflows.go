package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/peasant-labs/schema/internal/release"
	"github.com/peasant-labs/schema/internal/release/intent"
)

// runCheckWorkflows implements:
// release-intent check-workflows --policy P --dir .github/workflows [--print-required]
//
// It delegates to the policy package's exported checker. A policy violation
// exits 2 with a typed line; every diagnostic is printed on stderr.
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
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fmt.Errorf("check-workflows: cannot read workflows directory %q: %v. Fix the --dir path or run from the repository root", dir, err)
	}
	policy, err := release.LoadWorkflowPolicy(policyPath)
	if err != nil {
		return err
	}
	diagnostics, required := release.CheckWorkflows(policy, dir)
	if printRequired {
		names := make([]string, 0, len(required))
		for _, r := range required {
			names = append(names, r.Context)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Println(name)
		}
		return nil
	}
	if len(diagnostics) > 0 {
		for _, d := range diagnostics {
			where := d.Workflow
			if d.Job != "" {
				where += " job " + d.Job
			}
			fmt.Fprintf(os.Stderr, "release-intent check-workflows: rule %s in %s: %s\n", d.Rule, where, d.Message)
		}
		first := diagnostics[0]
		where := first.Workflow
		if first.Job != "" {
			where += " job " + first.Job
		}
		return rejectAs("check-workflows", &intent.Error{
			Code:  intent.ErrorCode("policy-violation"),
			What:  first.Message,
			Why:   "every required workflow must satisfy the merge-queue policy",
			Where: "workflows in " + dir + ", checked by release-intent check-workflows against " + policyPath,
			When:  "checking workflow policy",
			Means: "the check-workflows gate is red",
			Fix:   "align the workflows with the policy file",
		})
	}
	return emitLines([]string{"ok=true"})
}
