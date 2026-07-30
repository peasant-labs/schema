package schema

import (
	"fmt"
	"slices"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// --- Owner transcript update ---
//
// PATCH /api/v1/transcripts/{id} lets a transcript's OWNER revise the metadata
// and governance axes of an already-published transcript. The village has served
// this operation since the governance work landed; this contract declares it so
// a CLI client can call it from the published spec instead of from knowledge of
// the handler. Ownership is enforced server-side: a non-owner is refused with
// 403 and no state or audit row moves.
//
// Every field is OPTIONAL and absence means "leave unchanged". That is what
// makes this a partial update rather than a replacement: a client that wants to
// change only visibility sends only visibility, and the server resolves every
// omitted field against the locked pre-image inside its transaction, so a
// concurrent edit is never silently reverted.

// TranscriptUpdateVisibility is the closed set of visibility values the owner
// update operation accepts.
//
// It is deliberately NARROWER than the general Visibility enum. Visibility
// carries a third member for organization-scoped access, which this operation
// does not declare: organization visibility is a deferred capability, not merely
// an unimplemented one, so declaring it here would publish deferred scope into
// the contract and advertise a value a client cannot successfully send. The
// general Visibility enum is unchanged; only what THIS operation declares is
// constrained.
type TranscriptUpdateVisibility string

const (
	// TranscriptUpdateVisibilityPrivate restricts the transcript to its owner.
	TranscriptUpdateVisibilityPrivate TranscriptUpdateVisibility = "private"
	// TranscriptUpdateVisibilityPublic makes the transcript openly readable.
	TranscriptUpdateVisibilityPublic TranscriptUpdateVisibility = "public"
)

// AllTranscriptUpdateVisibilities is the canonical menu of visibility values
// this operation accepts, in declaration order. The generated schema enum is
// derived from it so a widening is a one-line edit that cannot drift from the
// documented contract.
var AllTranscriptUpdateVisibilities = []TranscriptUpdateVisibility{
	TranscriptUpdateVisibilityPrivate, TranscriptUpdateVisibilityPublic,
}

// IsValid reports whether the value is one of the accepted variants.
func (v TranscriptUpdateVisibility) IsValid() bool {
	return slices.Contains(AllTranscriptUpdateVisibilities, v)
}

func (v TranscriptUpdateVisibility) String() string { return string(v) }

// TranscriptUpdateVisibilityMenu renders the accepted values for help text and
// error messages, derived from AllTranscriptUpdateVisibilities so a message can
// never drift from the enum.
func TranscriptUpdateVisibilityMenu() string {
	values := make([]string, len(AllTranscriptUpdateVisibilities))
	for i, v := range AllTranscriptUpdateVisibilities {
		values[i] = string(v)
	}
	return strings.Join(values, ", ")
}

// JSONSchema implements jsonschema.Exposer.
func (TranscriptUpdateVisibility) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("TranscriptUpdateVisibility")
	s.WithDescription("Visibility values accepted by the owner transcript update operation. " +
		"Organization-scoped visibility is deferred and is deliberately not offered here.")
	enum := make([]any, len(AllTranscriptUpdateVisibilities))
	for i, v := range AllTranscriptUpdateVisibilities {
		enum[i] = string(v)
	}
	s.WithEnum(enum...)
	s.WithExamples("public", "private")
	return s, nil
}

// TranscriptUpdateLicense is the license value the owner update operation
// accepts: any license on the canonical License menu, PLUS the empty string,
// which CLEARS the license.
//
// The empty string is a real member of this operation's value set rather than a
// missing value, and that is the whole point of the type: it is what lets a
// caller distinguish "do not touch the license" (omit the field entirely) from
// "remove the license" (send ""). Collapsing the two would make the
// irrevocability rule below unexpressible.
//
// Clearing is NOT always permitted. A license that was actually granted can
// never be cleared, because a Creative Commons grant is irrevocable for anyone
// who already received the work, so an un-license would misrepresent the legal
// state. The server refuses that with 400. Clearing a transcript that was never
// licensed is an accepted no-op, and replacing one menu license with a different
// one stays allowed.
type TranscriptUpdateLicense string

// TranscriptUpdateLicenseClear is the value that clears the license. It is
// spelled as a named constant so callers express the intent rather than an
// unexplained empty string.
const TranscriptUpdateLicenseClear TranscriptUpdateLicense = ""

// AllTranscriptUpdateLicenses is the canonical menu of values this operation
// accepts: the clear sentinel first, then every canonical menu license in order.
// It is DERIVED from AllLicenses rather than restated, so widening the license
// menu widens this operation with no second edit to forget.
var AllTranscriptUpdateLicenses = func() []TranscriptUpdateLicense {
	values := make([]TranscriptUpdateLicense, 0, len(AllLicenses)+1)
	values = append(values, TranscriptUpdateLicenseClear)
	for _, license := range AllLicenses {
		values = append(values, TranscriptUpdateLicense(license))
	}
	return values
}()

// IsValid reports whether the value is either the clear sentinel or a license on
// the canonical menu. It derives the menu from AllLicenses rather than repeating
// it, so widening the license menu widens this operation automatically.
func (l TranscriptUpdateLicense) IsValid() bool {
	if l == TranscriptUpdateLicenseClear {
		return true
	}
	return License(l).IsValid()
}

func (l TranscriptUpdateLicense) String() string { return string(l) }

// JSONSchema implements jsonschema.Exposer.
func (TranscriptUpdateLicense) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("TranscriptUpdateLicense")
	s.WithDescription("License value for the owner transcript update operation: a canonical menu " +
		"license, or the empty string to clear. Clearing a license that was actually granted is " +
		"refused, because a granted Creative Commons license is irrevocable.")
	enum := make([]any, 0, len(AllTranscriptUpdateLicenses))
	for _, license := range AllTranscriptUpdateLicenses {
		enum = append(enum, string(license))
	}
	s.WithEnum(enum...)
	s.WithExamples("CC-BY-4.0", "")
	return s, nil
}

// TranscriptUpdateRequest is the body of PATCH /api/v1/transcripts/{id}.
//
// Every field is a pointer and every field is optional, because omission is
// meaningful: a nil field means "leave this unchanged", NOT "set this to the
// zero value". Title and Description are therefore settable to the empty string
// (a caller pointing at "" clears the text) without that being confusable with
// omitting them.
//
// License carries the third state described on TranscriptUpdateLicense: nil
// preserves, a pointer to the clear sentinel removes, and a pointer to a menu
// value replaces.
type TranscriptUpdateRequest struct {
	// Title replaces the transcript title when non-nil.
	Title *string `json:"title,omitempty"`
	// Description replaces the transcript description when non-nil.
	Description *string `json:"description,omitempty"`
	// Visibility replaces the access level when non-nil.
	Visibility *TranscriptUpdateVisibility `json:"visibility,omitempty"`
	// License preserves when nil, clears when it points at the clear sentinel,
	// and replaces when it points at a canonical menu license.
	License *TranscriptUpdateLicense `json:"license,omitempty"`
}

// Validate checks the closed-set fields a client can get wrong locally, so a
// caller learns about a bad enum value before spending a network round trip on a
// guaranteed 400. It deliberately does NOT attempt to predict the server's
// ownership check or the irrevocability refusal: both depend on stored state
// this request cannot see, and pretending otherwise would produce a local
// success that the server then contradicts.
func (r TranscriptUpdateRequest) Validate() error {
	if r.Visibility != nil && !r.Visibility.IsValid() {
		return fmt.Errorf("transcript update validation failed at schema.TranscriptUpdateRequest.Validate during owner-update request validation: visibility %q is not an accepted value; the owner update operation accepts only %s, because organization-scoped visibility is deferred; send one of the accepted values or omit visibility to leave it unchanged", r.Visibility.String(), TranscriptUpdateVisibilityMenu())
	}
	if r.License != nil && !r.License.IsValid() {
		return fmt.Errorf("transcript update validation failed at schema.TranscriptUpdateRequest.Validate during owner-update request validation: license %q is not on the canonical menu; the owner update operation accepts %s, or the empty string to clear a license that was never granted; send a menu license or omit license to leave it unchanged", r.License.String(), LicenseMenu())
	}
	return nil
}

// TranscriptUpdateErrorResponse is the body the owner update operation returns
// on every refusal. The village serves one uniform error envelope, so each
// declared non-success status carries this same shape and a client can read the
// reason from one field regardless of which refusal it hit.
//
// It is scoped to this operation rather than declared village-wide because this
// is currently the only operation whose refusals are declared; promoting it to a
// shared envelope is a deliberate decision for whichever change declares the
// next one, not a side effect of this one.
type TranscriptUpdateErrorResponse struct {
	// Error is the human-readable, actionable refusal reason.
	Error string `json:"error"`
}

// IsEmpty reports whether the request would change nothing. A body with every
// field omitted is accepted by the server and is a no-op, so a caller that built
// a request from an unchanged form can detect that and skip the call rather than
// spend a round trip and an audit evaluation on it.
func (r TranscriptUpdateRequest) IsEmpty() bool {
	return r.Title == nil && r.Description == nil && r.Visibility == nil && r.License == nil
}
