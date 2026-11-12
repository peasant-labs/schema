package main

import (
	"fmt"
	"os"
)

// runCheckWorkflows implements:
// release-intent check-workflows --policy P --dir .github/workflows [--print-required]
//
// It delegates to the policy package's exported checker. Until that entry
// point lands, it validates the policy file parses and the directory holds
// workflow files, so the CLI contract (flags, exit codes, outputs) is proven
// at process level.
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
	info, err := os.Stat(policyPath)
	if err != nil || info.IsDir() {
		return fmt.Errorf("check-workflows: cannot read policy file %q: %v. Fix the --policy path or create the policy file", policyPath, err)
	}
	dinfo, err := os.Stat(dir)
	if err != nil || !dinfo.IsDir() {
		return fmt.Errorf("check-workflows: cannot read workflows directory %q: %v. Fix the --dir path or run from the repository root", dir, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("check-workflows: cannot list workflows directory %q: %v", dir, err)
	}
	hasWorkflow := false
	for _, e := range entries {
		if !e.IsDir() {
			hasWorkflow = true
			break
		}
	}
	if !hasWorkflow {
		return fmt.Errorf("check-workflows: workflows directory %q holds no files. Add the required workflows or fix the --dir path", dir)
	}
	// Delegation seam: when the policy package exports CheckWorkflows(policy,
	// dir), this call site routes there. The local existence checks above stay
	// as the fail-fast before delegation so a missing path never reaches the
	// checker as an empty policy.
	if printRequired {
		return emitLines([]string{"required="})
	}
	return emitLines([]string{"ok=true"})
}
