package main

import (
	"os"
	"testing"

	"github.com/peasant-labs/schema/internal/testutil"
)

func TestMain(m *testing.M) { os.Exit(testutil.RunWithIsolatedGitConfig(m)) }
