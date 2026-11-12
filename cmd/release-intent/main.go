// Command release-intent is the merge-queue release tool every repo runs from a
// pinned schema checkout. It holds every release decision: title grammar,
// version arithmetic, release config loading, Plan, Generate, QueueGate and
// the git fact collection behind them.
//
// The tool has no network access and never imports the GitHub client. The one
// impure seam is the git runner, which reads commits, tags and files from git
// through a hardened structured argv.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/peasant-labs/schema/internal/release/intent"
)

func main() {
	if len(os.Args) < 2 {
		usageErrorf("usage: release-intent <check-config|check-workflows|check-title|classify|group-pr|plan|release-pr|queue-gate> ...")
	}
	sub := os.Args[1]
	args := os.Args[2:]

	var err error
	switch sub {
	case "check-config":
		err = runCheckConfig(args)
	case "check-workflows":
		err = runCheckWorkflows(args)
	case "check-title":
		err = runCheckTitle(args)
	case "classify":
		err = runClassify(args)
	case "group-pr":
		err = runGroupPR(args)
	case "plan":
		err = runPlan(args)
	case "release-pr":
		err = runReleasePR(args)
	case "queue-gate":
		err = runQueueGate(args)
	default:
		usageErrorf("unknown subcommand %q: expected check-config, check-workflows, check-title, classify, group-pr, plan, release-pr or queue-gate", sub)
	}
	if err != nil {
		exitWith(err)
	}
}

// exitWith maps a handler error to the process exit status and streams.
//
// Exit contract: 0 is accepted, 2 is rejected input, 1 is usage or internal
// error. A rejection prints its machine channel (error=<code> and friends) on
// stdout and a six-part human message on stderr. A usage or internal failure
// prints an actionable message on stderr and no machine line unless the
// handler already emitted one.
func exitWith(err error) {
	if r, ok := asReject(err); ok {
		for _, line := range r.machineLines() {
			fmt.Println(line)
		}
		if path := os.Getenv("GITHUB_OUTPUT"); path != "" {
			_ = appendOutputFile(path, r.machineLines())
		}
		fmt.Fprintf(os.Stderr, "%s\n", r.humanMessage())
		os.Exit(2)
	}
	var ie *intent.Error
	if errors.As(err, &ie) {
		if r, ok := asReject(rejectAs("release-intent", err)); ok {
			for _, line := range r.machineLines() {
				fmt.Println(line)
			}
			if path := os.Getenv("GITHUB_OUTPUT"); path != "" {
				_ = appendOutputFile(path, r.machineLines())
			}
			fmt.Fprintf(os.Stderr, "%s\n", r.humanMessage())
			os.Exit(2)
		}
	}
	if u, ok := asUsage(err); ok {
		fmt.Fprintf(os.Stderr, "release-intent: %s\n", u.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "release-intent: %v\n", err)
	os.Exit(1)
}

func usageErrorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "release-intent: "+format+"\n", args...)
	os.Exit(1)
}
