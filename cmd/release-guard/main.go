// Command release-guard validates release tags for the release workflows.
// parse-tag writes version and kind to stdout and, when set, $GITHUB_OUTPUT.
package main

import (
	"fmt"
	"os"

	"github.com/peasant-labs/schema/internal/release"
)

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: release-guard parse-tag \"<git tag>\"")
	}
	if os.Args[1] != "parse-tag" {
		fatalf("unknown subcommand %q: expected parse-tag", os.Args[1])
	}
	if err := runParseTag(os.Args[2:]); err != nil {
		fatalf("%v", err)
	}
}

func runParseTag(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: release-guard parse-tag \"<git tag>\"")
	}
	v, kind, err := release.ParseTag(args[0])
	if err != nil {
		return err
	}
	return emit(v, kind)
}

// --- output helpers ---

// formatOutput renders the version/kind lines exactly as written to
// $GITHUB_OUTPUT and stdout. It is the byte-parity surface: the format is
// "version=<v>\nkind=<k>\n", unchanged across the go-github swap.
func formatOutput(v release.Version, kind release.ReleaseKind) string {
	return fmt.Sprintf("version=%s\nkind=%s\n", v, kind)
}

// emit writes version/kind to $GITHUB_OUTPUT (when set) and stdout.
func emit(v release.Version, kind release.ReleaseKind) error {
	line := formatOutput(v, kind)
	if path := os.Getenv("GITHUB_OUTPUT"); path != "" {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("cannot open $GITHUB_OUTPUT (%s): %v", path, err)
		}
		defer f.Close()
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("cannot write to $GITHUB_OUTPUT (%s): %v", path, err)
		}
	}
	fmt.Print(line)
	return nil
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "release-guard: "+format+"\n", a...)
	os.Exit(1)
}
