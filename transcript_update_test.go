package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase/assert"
)

func loadOwnerUpdateFixtures(t *testing.T) *schema.OwnerUpdateFixtures {
	t.Helper()
	fixtures, err := schema.LoadOwnerUpdateFixtures()
	if err != nil {
		t.Fatalf("load owner update fixtures: %v", err)
	}
	return fixtures
}

// TestOwnerUpdateRequestValidation drives raw client bytes through the real
// decode-then-validate path a caller uses, so the closed enum sets are proven on
// the shipped code rather than on a test-local reimplementation.
func TestOwnerUpdateRequestValidation(t *testing.T) {
	fixtures := loadOwnerUpdateFixtures(t)
	assert.RequireValid(t, fixtures.RequestValidations)
	assert.RequireMin(t, fixtures.RequestValidations, 14)

	for _, c := range fixtures.RequestValidations.Cases {
		t.Run(c.Name, func(t *testing.T) {
			// A body can be refused at either boundary: the strict decoder
			// rejects unknown fields and explicit nulls, and Validate rejects
			// out-of-set enum values. A caller experiences both as "the
			// contract refused this", so the row asserts against whichever
			// fires rather than caring which one did.
			var request schema.TranscriptUpdateRequest
			err := json.Unmarshal([]byte(c.Input), &request)
			if err == nil {
				err = request.Validate()
			}
			if c.Expected.Accepted {
				if err != nil {
					t.Fatalf("body %s must be accepted, got refusal: %v", c.Input, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("body %s must be refused, but validation accepted it", c.Input)
			}
			if !strings.Contains(err.Error(), c.Expected.ErrorContains) {
				t.Fatalf("refusal for %s must explain the failure with %q, got: %v", c.Input, c.Expected.ErrorContains, err)
			}
		})
	}
}

// TestOwnerUpdateEncoding pins the exact bytes each update intent produces. It is
// the evidence that preserve, clear, and replace stay three distinguishable
// states on the wire: if any two collapsed, the server could not implement the
// rule that a granted license is irrevocable.
func TestOwnerUpdateEncoding(t *testing.T) {
	fixtures := loadOwnerUpdateFixtures(t)
	assert.RequireValid(t, fixtures.Encodings)
	assert.RequireMin(t, fixtures.Encodings, 6)

	for _, c := range fixtures.Encodings.Cases {
		t.Run(c.Name, func(t *testing.T) {
			encoded, err := json.Marshal(c.Input.ToRequest())
			if err != nil {
				t.Fatalf("encode update intent: %v", err)
			}
			if string(encoded) != c.Expected {
				t.Fatalf("encoded bytes must be exactly %s, got %s", c.Expected, encoded)
			}
		})
	}
}

// TestOwnerUpdateLicenseStatesAreDistinguishable states the tri-state property
// directly rather than leaving it implied by the byte fixtures, because it is the
// single property the irrevocability gate depends on. A decoded body must report
// preserve, clear, and replace as three different observations.
func TestOwnerUpdateLicenseStatesAreDistinguishable(t *testing.T) {
	decode := func(body string) schema.TranscriptUpdateRequest {
		t.Helper()
		var r schema.TranscriptUpdateRequest
		if err := json.Unmarshal([]byte(body), &r); err != nil {
			t.Fatalf("decode %s: %v", body, err)
		}
		return r
	}

	preserve := decode(`{"title":"unrelated"}`)
	if preserve.License != nil {
		t.Fatalf("an omitted license must decode to no license intent, got %q", preserve.License.String())
	}

	clear := decode(`{"license":""}`)
	if clear.License == nil {
		t.Fatal("an explicit empty license must decode to a present intent, not an omission; collapsing the two would make clearing unrequestable")
	}
	if *clear.License != schema.TranscriptUpdateLicenseClear {
		t.Fatalf("an explicit empty license must decode to the clear sentinel, got %q", clear.License.String())
	}

	replace := decode(`{"license":"CC0-1.0"}`)
	if replace.License == nil {
		t.Fatal("a menu license must decode to a present intent")
	}
	if *replace.License == schema.TranscriptUpdateLicenseClear {
		t.Fatal("a menu license must not decode to the clear sentinel; replacement and clearing must stay distinguishable")
	}
}

// TestOwnerUpdateVisibilityExcludesDeferredOrganizationScope pins the narrowing
// itself. The general Visibility enum keeps its third member; this operation must
// not accept it, and the general enum must not be altered to achieve that.
func TestOwnerUpdateVisibilityExcludesDeferredOrganizationScope(t *testing.T) {
	if len(schema.AllTranscriptUpdateVisibilities) != 2 {
		t.Fatalf("the owner update visibility menu must hold exactly the two accepted members, got %v", schema.AllTranscriptUpdateVisibilities)
	}
	if schema.TranscriptUpdateVisibility(schema.VisibilityGroup).IsValid() {
		t.Fatal("the owner update operation must not accept organization-scoped visibility; it is a deferred capability and the village refuses it")
	}
	if !schema.VisibilityGroup.IsValid() {
		t.Fatal("narrowing this operation must not narrow the general Visibility enum")
	}
}

// TestOwnerUpdateLicenseMenuDerivesFromCanonical proves the operation's license
// set is derived rather than duplicated, so widening the canonical menu widens
// this operation without a second edit that could be forgotten.
func TestOwnerUpdateLicenseMenuDerivesFromCanonical(t *testing.T) {
	// The headline claim lives inside the loop below, so an empty canonical menu
	// would satisfy this test without ever evaluating the derivation. Guard the
	// iteration source before trusting what iterating it proves.
	if len(schema.AllLicenses) == 0 {
		t.Fatal("the canonical license menu is empty, so the derivation loop below would assert nothing while reporting success")
	}
	if len(schema.AllTranscriptUpdateLicenses) != len(schema.AllLicenses)+1 {
		t.Fatalf("the operation menu holds %d values but the canonical menu holds %d; the operation menu must be the canonical menu plus the clear sentinel, or it is not derived at all", len(schema.AllTranscriptUpdateLicenses), len(schema.AllLicenses))
	}
	for _, license := range schema.AllLicenses {
		if !schema.TranscriptUpdateLicense(license).IsValid() {
			t.Fatalf("canonical menu license %q must be accepted by the owner update operation", license)
		}
	}
	if !schema.TranscriptUpdateLicenseClear.IsValid() {
		t.Fatal("the clear sentinel must be an accepted value; it is what makes un-licensing requestable and therefore refusable")
	}
	if schema.TranscriptUpdateLicense("MIT").IsValid() {
		t.Fatal("a license outside the canonical menu must be refused")
	}
}
