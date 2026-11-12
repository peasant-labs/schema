package main

import (
	"fmt"
	"regexp"
	"strconv"
)

// groupRefPattern matches merge-group refs:
// refs/heads/gh-readonly-queue/<branch>/pr-<N>-<sha>
// where branch may contain slashes.
var groupRefPattern = regexp.MustCompile(`^refs/heads/gh-readonly-queue/(.+)/pr-(\d+)-[0-9a-fA-F]+$`)

// parseGroupRef extracts the PR number from a merge-group ref. Any other ref
// yields "" with no error: the PR-time check stays authoritative and a
// ref-format change must not block the queue.
func parseGroupRef(ref string) string {
	m := groupRefPattern.FindStringSubmatch(ref)
	if m == nil {
		return ""
	}
	n, err := strconv.Atoi(m[2])
	if err != nil || n <= 0 {
		return ""
	}
	return strconv.Itoa(n)
}

// runGroupPR implements: release-intent group-pr --ref REF -> pr=<number>.
func runGroupPR(args []string) error {
	var ref string
	var have bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--ref":
			val, next, err := requireFlagValue(args, i, "--ref")
			if err != nil {
				return err
			}
			ref = val
			have = true
			i = next
		default:
			return usageErr("unknown flag %q for group-pr: expected --ref REF (usage: release-intent group-pr --ref REF)", args[i])
		}
	}
	if !have {
		return usageErr("usage: release-intent group-pr --ref REF (missing required --ref): pass the merge-group ref, e.g. refs/heads/gh-readonly-queue/main/pr-12-abc123")
	}
	return emitLines([]string{"pr=" + parseGroupRef(ref)})
}

// groupRefBranch extracts the branch portion of a merge-group ref for tests.
func groupRefBranch(ref string) string {
	m := groupRefPattern.FindStringSubmatch(ref)
	if m == nil {
		return ""
	}
	return m[1]
}

var _ = fmt.Sprint
