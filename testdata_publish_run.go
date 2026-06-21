package schema

import (
	"strings"
	"testing"
)

// RunPublishVerdicts drives a publish-body validator over the ENTIRE shared
// publish-verdict corpus (testdata/publish/verdicts.yaml), in two passes — first
// the Acceptances() (schema_accepts:true), then the Rejections()
// (schema_accepts:false) — asserting each row's expectation with
// ASSERT-WHERE-PRESENT semantics (FINDING-2a):
//
//   - an accepted row MUST validate with no error;
//   - a rejected row MUST error;
//   - WHERE a rejected row pins error_contains, the validator message MUST contain
//     that exact (minimal/stable) substring — but rows WITHOUT error_contains are
//     NOT required to carry one (no false-fail on the ~half the corpus that pins
//     only error_category); and
//   - EVERY rejected row MUST carry a non-empty error_category (no vacuous reject).
//
// It is the one-call, reusable form of the accept/reject loop every corpus
// consumer would otherwise open-code, so the corpus stays the SINGLE source of
// publish-validation cases. Pass the validator under test — e.g.
// schema.RunPublishVerdicts(t, schema.ValidatePublishRequest) for the root client
// validator, or a compiled-schema closure for the extracted-schema side.
//
// ERROR-STRING BREADCRUMB (FINDING-2b): the pinned error_contains substrings are
// coupled to the message text of github.com/santhosh-tekuri/jsonschema/v5, pinned
// at v5.3.1 (see go.mod). They are kept MINIMAL/stable on purpose — e.g.
// "value must be one of", "missing properties: 'model'"/'harness'", "does not
// match pattern". If a future jsonschema/v5 bump changes the wording, the
// error_contains assertions here (and TestPublishVerdicts_ErrorContainsPinnedStrings)
// fail loudly so the drift is diagnosable at the dep, not silently tolerated.
//
// NOTE on placement: this helper takes *testing.T, so it necessarily imports
// "testing". It is isolated in this dedicated file (not the production-logic
// files) and is consumed only by _test packages (publish_validate_test, openapi);
// production code never calls it.
func RunPublishVerdicts(t *testing.T, validate func([]byte) error) {
	t.Helper()
	fixtures, err := LoadPublishVerdictFixtures()
	if err != nil {
		t.Fatalf("RunPublishVerdicts: load publish verdict fixtures: %v", err)
	}

	acc := fixtures.Acceptances()
	rej := fixtures.Rejections()

	// GUARD 1 — partition completeness. The two-pass driver only exercises
	// acc ∪ rej, so every corpus row MUST land in exactly one of the two filters.
	// If a future change to Acceptances()/Rejections() ever dropped a row from BOTH
	// passes (e.g. a third verdict state), that row would silently stop being
	// validated and ONLY this equality catches it. It is the completeness invariant
	// the driver's correctness depends on — not a tautology.
	if len(acc)+len(rej) != len(fixtures.Cases) {
		t.Fatalf("RunPublishVerdicts: corpus partition incomplete — Acceptances()=%d + Rejections()=%d != Cases=%d; "+
			"every row must fall in exactly one filter, else it is silently unvalidated",
			len(acc), len(rej), len(fixtures.Cases))
	}

	// GUARD 2 — both passes have teeth. An empty corpus, or one that is all-accept
	// or all-reject, would let a no-op validator pass vacuously; require BOTH sides
	// to be exercised.
	if len(acc) == 0 || len(rej) == 0 {
		t.Fatalf("RunPublishVerdicts: corpus must exercise BOTH accept and reject — acc=%d rej=%d", len(acc), len(rej))
	}

	// Pass 1: every accepted row must validate cleanly. Flat subtests (no
	// "accepted"/"rejected" grouping) keep -run paths stable per row name.
	for _, tc := range acc {
		t.Run(tc.Name, func(t *testing.T) {
			assertAccepted(t, tc, validate)
		})
	}
	// Pass 2: every rejected row must error (and match its pinned substring).
	for _, tc := range rej {
		t.Run(tc.Name, func(t *testing.T) {
			assertRejected(t, tc, validate)
		})
	}
}

// assertAccepted requires the validator to ACCEPT a corpus row marked
// schema_accepts:true (no error).
func assertAccepted(t *testing.T, tc PublishVerdictCase, validate func([]byte) error) {
	t.Helper()
	if err := validate([]byte(tc.Body)); err != nil {
		t.Errorf("row %q: validator REJECTED a body the corpus marks accepted: %v", tc.Name, err)
	}
}

// assertRejected requires the validator to REJECT a corpus row marked
// schema_accepts:false, that the row carries a non-empty error_category (no
// vacuous reject), and — WHERE the row pins error_contains — that the validator
// message carries that exact substring (ASSERT-WHERE-PRESENT, FINDING-2a).
func assertRejected(t *testing.T, tc PublishVerdictCase, validate func([]byte) error) {
	t.Helper()
	// Every reject row must classify its failure — guards against a row that
	// rejects "for some reason" with no documented category (vacuous reject).
	if tc.Expect.ErrorCategory == "" {
		t.Errorf("row %q: reject row has empty error_category — every rejection must classify its failure", tc.Name)
	}
	err := validate([]byte(tc.Body))
	if err == nil {
		t.Errorf("row %q: validator ACCEPTED a body the corpus marks rejected", tc.Name)
		return // MUST return before touching err — no nil-deref below.
	}
	if tc.Expect.ErrorContains != "" && !strings.Contains(err.Error(), tc.Expect.ErrorContains) {
		t.Errorf("row %q: validator message missing pinned error_contains %q\n got: %s",
			tc.Name, tc.Expect.ErrorContains, err.Error())
	}
}
