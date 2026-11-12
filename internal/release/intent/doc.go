// Package intent holds the pure decision logic behind the release-intent
// tool: the PR-title grammar, the PR classifier, the version arithmetic, the
// release-config loader, the release planner, the release-content generator,
// the merge-queue gate plus its queue-response decoder, and the byte-exact
// version-field reader and writer.
//
// Everything here is side-effect free: every decision is a pure function of
// text. Reading commits, tags and files from git lives in the caller, which
// fills Facts and calls back in.
package intent
