package main

import (
	"os"
	"testing"

	"github.com/peasant-labs/schema/internal/testutil"
)

// TestMain gives every git process these tests start, directly or through a
// script or production code, no configuration outside its repository, so a
// developer's signing key, URL rewrites, or injected GIT_CONFIG_* values cannot
// change a result.
func TestMain(m *testing.M) { os.Exit(testutil.RunWithIsolatedGitConfig(m)) }
