package testutil

import (
	"fmt"
	"os"
)

// RunWithIsolatedGitConfig makes every git process the calling test binary
// starts read no configuration outside the repository it runs in, then runs
// the package's tests and returns the exit code for os.Exit. Call it from
// TestMain in every package whose tests run git directly, through a shell
// script, or through production code:
//
//	func TestMain(m *testing.M) { os.Exit(testutil.RunWithIsolatedGitConfig(m)) }
//
// A package whose TestMain does more passes a value whose Run does the rest.
//
// Without it, a test inherits the developer's configuration. A global
// commit.gpgsign=true starts a signer that can block on a pinentry prompt, and
// a url.<base>.insteadOf rewrite changes which remote git contacts, so a test
// that passes in CI fails or hangs on a developer machine.
//
// The isolation is process-wide and is inherited by every child built from
// os.Environ(), including shell scripts and nested go commands:
//
//   - GIT_CONFIG_GLOBAL=/dev/null replaces ~/.gitconfig and
//     $XDG_CONFIG_HOME/git/config;
//   - GIT_CONFIG_NOSYSTEM=1 skips the system file;
//   - GIT_CONFIG_COUNT and GIT_CONFIG_PARAMETERS are removed, because
//     configuration injected through the environment bypasses both files.
//
// A test that needs a specific ambient configuration (for example, to prove
// that a helper strips hostile configuration) still sets it with t.Setenv,
// which overrides this baseline for that test only. Production code is not
// changed: only the test binary's environment is.
func RunWithIsolatedGitConfig(m interface{ Run() int }) int {
	if err := isolateGitConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	return m.Run()
}

func isolateGitConfig() error {
	for name, value := range map[string]string{
		"GIT_CONFIG_GLOBAL":   os.DevNull,
		"GIT_CONFIG_NOSYSTEM": "1",
	} {
		if err := os.Setenv(name, value); err != nil {
			return fmt.Errorf("isolate test git config: set %s=%s in the test process environment: %w; "+
				"git children would read the developer's configuration, so the package's tests cannot run hermetically", name, value, err)
		}
	}
	for _, name := range []string{"GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS"} {
		if err := os.Unsetenv(name); err != nil {
			return fmt.Errorf("isolate test git config: unset %s in the test process environment: %w; "+
				"git children would read configuration injected through the environment, so the package's tests cannot run hermetically", name, err)
		}
	}
	return nil
}
