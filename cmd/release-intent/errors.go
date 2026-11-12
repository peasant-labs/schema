package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// RejectError attaches the invoking subcommand to a contract rejection so the
// human message names the actual command. It wraps *intent.Error instead of
// copying its fields: the contract package owns the six parts.
type RejectError struct {
	Subcommand string
	Err        *intent.Error
}

func (e *RejectError) Error() string { return e.Err.Error() }

func (e *RejectError) Unwrap() error { return e.Err }

func (e *RejectError) machineLines() []string {
	return []string{"error=" + e.Err.Code.String()}
}

func (e *RejectError) humanMessage() string {
	return fmt.Sprintf("release-intent %s: rejected %q\n  what:  %s (%s)\n  why:   %s\n  where: %s\n  when:  %s\n  means: %s\n  fix:   %s",
		e.Subcommand, e.Err.Code, e.Err.What, e.Err.Code, e.Err.Why, e.Err.Where, e.Err.When, e.Err.Means, e.Err.Fix)
}

// rejectAs wraps a contract rejection with the invoking subcommand for the
// exit-2 path. Every other error passes through unchanged for exit 1.
func rejectAs(subcommand string, err error) error {
	var ie *intent.Error
	if errors.As(err, &ie) {
		return &RejectError{Subcommand: subcommand, Err: ie}
	}
	return err
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
