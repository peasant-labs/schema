// Package intent holds the pure decision logic behind the release-intent
// tool: the PR-title grammar, the PR classifier, the version arithmetic, the
// release-config loader, the release planner, the release-content generator,
// and the merge-queue gate plus its queue-response decoder.
//
// Everything here is side-effect free. The one impure seam, reading commits,
// tags and files from git, sits behind the GitRunner interface, so every
// decision is a pure function of text that fixtures can pin. The thin CLI in
// cmd/release-intent wires these functions to real repositories.
package intent
