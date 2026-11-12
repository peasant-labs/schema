package main

import (
	"fmt"
	"os"
)

// emitLines writes key=value lines to stdout and appends them to the file
// named by $GITHUB_OUTPUT when set, so workflow steps can consume them as
// step outputs. The file is opened append-only and never truncated: it
// accumulates outputs across a job's steps.
func emitLines(lines []string) error {
	for _, line := range lines {
		fmt.Println(line)
	}
	if path := os.Getenv("GITHUB_OUTPUT"); path != "" {
		if err := appendOutputFile(path, lines); err != nil {
			return fmt.Errorf("cannot write to $GITHUB_OUTPUT (%s): %w. Unset $GITHUB_OUTPUT to print to stdout only, or fix the path permissions", path, err)
		}
	}
	return nil
}

func appendOutputFile(path string, lines []string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range lines {
		if _, err := f.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return nil
}
