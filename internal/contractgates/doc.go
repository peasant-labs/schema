// Package contractgates holds the SYNTHETIC-BREAK tests for the schema repo's
// breaking-change contract gates: oasdiff for the OpenAPI specs and go-apidiff
// for the exported Go API.
//
// Each breaking gate (oasdiff for the OpenAPI specs, go-apidiff for the exported
// Go API) ships a test here that applies a KNOWN breaking mutation and asserts
// the gate binary exits non-zero / reports the break. This proves the gate
// actually fires — a gate that never fails is no gate at all.
//
// The tests invoke the gate binaries directly (they are provisioned by the flake
// dev shell, not go.mod). When a binary is not on PATH — e.g. a bare `go test`
// run outside `nix develop` — the test t.Skip()s with an ACTIONABLE message
// telling the developer how to get the tool, rather than failing spuriously.
// Inside `nix develop -c make check` the binaries are present and the tests run
// for real.
package contractgates
