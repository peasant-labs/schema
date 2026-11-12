package intent

import (
	"fmt"
	"strconv"
)

// inputError formats a six-part usage error for failures that are not typed
// rejections: unreadable files, malformed values, and other exit-1 faults.
// Typed rejections use *Error with a member of AllErrorCodes instead.
func inputError(what, why, where, when, means, fix string) error {
	return fmt.Errorf("release-intent: %s [why: %s; where: %s; when: %s; means: %s; fix: %s]",
		what, why, where, when, means, fix)
}

// rejection builds a typed six-part rejection for one intent title.
func rejection(code ErrorCode, raw, what, why, means, fix string) *Error {
	return &Error{
		Code:  code,
		What:  what,
		Why:   why,
		Where: "intent title " + strconv.Quote(raw) + ", parsed by intent.ParseTitle",
		When:  "checking a PR title before it can be enqueued",
		Means: means,
		Fix:   fix,
	}
}
