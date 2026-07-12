// Package assert is the testing seam for the test-case corpus: helpers that take
// *testing.T and fail a test when a corpus violates its own invariants. It lives
// in its own subpackage so the pure-data package
// github.com/peasant-labs/schema/testcase never imports testing.
package assert

import (
	"testing"

	"github.com/peasant-labs/schema/testcase"
)

// RequireMin fails the test (t.Fatalf) unless the corpus holds at least n cases.
// It is the loud, testing-side wrapper around Corpus.CheckMin: a corpus that was
// silently gutted, or never loaded, trips it here instead of passing vacuously.
// The size logic itself lives in the pure CheckMin so it can be negative-tested
// without a *testing.T.
func RequireMin[I any, E any](t *testing.T, corpus testcase.Corpus[I, E], n int) {
	t.Helper()
	if err := corpus.CheckMin(n); err != nil {
		t.Fatalf("RequireMin: %v", err)
	}
}

// RequireValid fails the test (t.Fatalf) unless every case in the corpus is
// non-vacuous: an in-set classification and provenance source, a non-empty
// provenance ref, and a non-empty mutation description. It is the loud,
// testing-side wrapper around the pure Corpus.Validate, symmetric to RequireMin
// around CheckMin: the validity logic lives in Validate so it can be
// negative-tested without a *testing.T. A guard call names the arm it guards by
// its call site, so a segmented fixture pairs it with RequireMin per arm.
func RequireValid[I any, E any](t *testing.T, corpus testcase.Corpus[I, E]) {
	t.Helper()
	if err := corpus.Validate(); err != nil {
		t.Fatalf("RequireValid: %v", err)
	}
}
