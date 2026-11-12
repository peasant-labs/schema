package intent

import (
	"strconv"
)

// This file holds the one workflow-policy rejection shape: the typed code
// and six parts every policy-breach report carries. The checker lives
// wherever the workflows are judged; every report site calls this
// constructor so the code is never re-spelled as a bare string.

// PolicyViolation builds the typed rejection for a workflow-policy rule
// breach. The rule, the place, and the checker's message ride in the report;
// the code and the recovery stay identical at every site.
func PolicyViolation(rule, where, message string) *Error {
	return &Error{
		Code:  ErrPolicyViolation,
		What:  "workflow policy rule " + strconv.Quote(rule) + " is violated: " + message,
		Why:   "every required workflow must satisfy the merge-queue policy",
		Where: where,
		When:  "checking workflow policy",
		Means: "the check-workflows gate is red",
		Fix:   "align the workflows with the policy file",
	}
}
