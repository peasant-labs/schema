package main

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCode is the closed wave-1 set of machine-readable rejection codes.
// Contract codes for later milestones enter this set in their milestone with
// their reject case.
type ErrorCode string

const (
	CodeMalformed           ErrorCode = "malformed"
	CodeUnknownType         ErrorCode = "unknown-type"
	CodeMissingBump         ErrorCode = "missing-bump"
	CodeBumpOnNonReleasing  ErrorCode = "bump-on-non-releasing"
	CodeBangBumpMismatch    ErrorCode = "bang-bump-mismatch"
	CodeOldReleaseTitle     ErrorCode = "old-release-title"
	CodeReleasesDisabled    ErrorCode = "releases-disabled"
	CodeProtectedPath       ErrorCode = "protected-path"
	CodeNoFinalTag          ErrorCode = "no-final-tag"
	CodeEpochMissing        ErrorCode = "epoch-missing"
	CodeFloorNotOnChain     ErrorCode = "floor-not-on-chain"
	CodeReleaseContentStale ErrorCode = "release-content-stale"
)

// AllErrorCodes is the canonical closed set. Tests assert exact membership.
var AllErrorCodes = []ErrorCode{
	CodeMalformed,
	CodeUnknownType,
	CodeMissingBump,
	CodeBumpOnNonReleasing,
	CodeBangBumpMismatch,
	CodeOldReleaseTitle,
	CodeReleasesDisabled,
	CodeProtectedPath,
	CodeNoFinalTag,
	CodeEpochMissing,
	CodeFloorNotOnChain,
	CodeReleaseContentStale,
}

// RejectError is a rejected input: exit 2 with a machine line and a six-part
// human message.
type RejectError struct {
	Code  ErrorCode
	What  string
	Why   string
	Where string
	When  string
	Means string
	Fix   string
}

func (e *RejectError) Error() string { return string(e.Code) + ": " + e.What }

func (e *RejectError) machineLines() []string { return []string{"error=" + string(e.Code)} }

func (e *RejectError) humanMessage() string {
	return fmt.Sprintf("title-guard: rejected input %q\n  what:  %s (%s)\n  why:   %s\n  where: %s\n  when:  %s\n  means: %s\n  fix:   %s",
		e.Code, e.What, e.Code, e.Why, e.Where, e.When, e.Means, e.Fix)
}

// UsageError is a CLI usage failure: exit 1 with an actionable message.
type UsageError struct{ msg string }

func (e *UsageError) Error() string { return e.msg }

func usageErr(format string, args ...any) error {
	return &UsageError{msg: fmt.Sprintf(format, args...)}
}

func asReject(err error) (*RejectError, bool) {
	var r *RejectError
	if errors.As(err, &r) {
		return r, true
	}
	return nil, false
}

func asUsage(err error) (*UsageError, bool) {
	var u *UsageError
	if errors.As(err, &u) {
		return u, true
	}
	return nil, false
}

// requireFlagValue pops --flag value from args, failing closed on a missing
// value or a value that looks like another flag.
func requireFlagValue(args []string, i int, flag string) (string, int, error) {
	if i+1 >= len(args) {
		return "", i, usageErr("usage: %s requires a value: pass %s <value> (got end of arguments after %s)", flag, flag, flag)
	}
	val := args[i+1]
	if strings.HasPrefix(val, "-") && val != "-" {
		return "", i, usageErr("usage: %s requires a value: got flag-like %q (pass a value that does not start with '-')", flag, val)
	}
	return val, i + 1, nil
}
