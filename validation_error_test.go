package schema_test

import (
	"testing"

	"github.com/peasant-labs/schema/internal/testutil"
)

func requireValidationErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	testutil.RequireValidationErrorContains(t, err, want)
}

func requireActionableValidationError(t *testing.T, err error, wantContains ...string) {
	t.Helper()
	testutil.RequireActionableValidationError(t, err, wantContains...)
}
