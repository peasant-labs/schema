package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/peasant-labs/schema/internal/release/intent"
)

// RejectError is a rejected input: exit 2 with a machine line and a six-part
// human message. The code comes from the contract package so the two never
// drift.
type RejectError struct {
	Code  intent.ErrorCode
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

// intentErrorToReject maps a contract error to the command-layer rejection so
// the machine channel stays error=<code> with exit 2.
func intentErrorToReject(err error) error {
	var ie *intent.Error
	if errors.As(err, &ie) {
		return &RejectError{
			Code: ie.Code, What: ie.What, Why: ie.Why,
			Where: ie.Where, When: ie.When, Means: ie.Means, Fix: ie.Fix,
		}
	}
	return err
}
