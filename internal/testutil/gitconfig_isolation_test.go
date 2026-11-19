package testutil

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) { os.Exit(RunWithIsolatedGitConfig(m)) }
