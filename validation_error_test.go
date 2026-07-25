package schema_test

import (
	"strings"
	"testing"
)

func requireValidationErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("validation succeeded, want an error containing %q", want)
	}
	if strings.TrimSpace(want) == "" {
		t.Fatal("validation error expectation must not be empty")
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("validation error %q is missing %q", err, want)
	}
}

func requireActionableValidationError(t *testing.T, err error) {
	t.Helper()
	requireValidationErrorContains(t, err, " at schema.")
	requireValidationErrorContains(t, err, "during wire-boundary validation")
}
