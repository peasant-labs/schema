package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

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
//
// JSON null is REFUSED on every field, and the refusal is the point. A caller
// reaching for null almost always means "clear this", but the village decodes a
// null into the same nil pointer an omitted field produces, so null would
// silently mean PRESERVE - the opposite of the obvious reading, on exactly the
// fields where clearing is what a caller wants. Rather than declare that trap,
// this contract rejects null and gives each intent its own unambiguous spelling:
// omit the field to leave it unchanged, and send the empty string to clear a
// title, a description, or a license. This is a deliberate narrowing of what the
// village tolerates, in the same spirit as the visibility menu: the contract
// never promises behavior the server lacks, it only declines to bless a spelling
// whose meaning would surprise the caller.
//
// Unknown fields are refused for the same reason. The village accepts and
// silently discards a "tags" field, so a client that sent tags would receive a
// success and believe they applied; refusing the field turns that silent no-op
// into an actionable error.
type TranscriptUpdateRequest struct {
	// Title replaces the transcript title when non-nil. The bound mirrors the
	// storage column, which is VARCHAR(500) and is enforced by nothing else on
	// the way in: an over-long title reaches the database and surfaces as an
	// opaque server error rather than a refusal naming the limit. Declaring it
	// lets a client catch the problem before it sends the request.
	//
	// The two limits count differently, and the difference is recorded here
	// rather than discovered later. The column counts CODE POINTS; a JavaScript
	// validator generated from this bound counts UTF-16 CODE UNITS, so an astral
	// character costs two there and one in the column. The direction is the safe
	// one: a JavaScript count is never smaller than the code-point count, so
	// nothing this bound accepts can be rejected by the column. The cost is at
	// the other end, where a title of more than 250 astral characters is refused
	// client-side though the column would take it. That band was judged worth
	// the far commoner case of an ordinary over-long title turning into a caught
	// validation error instead of an opaque server failure.
	Title *string `json:"title,omitempty" maxLength:"500"`
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
// TranscriptUpdateTitleMaxLength is the declared bound on a title, mirroring the
// storage column. It is a named constant because three places must agree: this
// validator, the reflector tag that emits maxLength into both documents, and the
// corpus rows that exercise the boundary.
//
// It counts RUNES, matching the column's code-point counting. A generated
// JavaScript validator counts UTF-16 code units instead and is therefore
// stricter on astral input; that divergence is deliberate and documented on
// Title above.
const TranscriptUpdateTitleMaxLength = 500

func (r TranscriptUpdateRequest) Validate() error {
	if r.Title != nil && utf8.RuneCountInString(*r.Title) > TranscriptUpdateTitleMaxLength {
		return fmt.Errorf("transcript update validation failed at schema.TranscriptUpdateRequest.Validate during owner-update request validation: title is %d characters, exceeding the %d the storage column accepts; without this check the request reaches the database and returns an opaque server error instead of naming the limit; shorten the title to %d characters or fewer", utf8.RuneCountInString(*r.Title), TranscriptUpdateTitleMaxLength, TranscriptUpdateTitleMaxLength)
	}
	if r.Visibility != nil && !r.Visibility.IsValid() {
		return fmt.Errorf("transcript update validation failed at schema.TranscriptUpdateRequest.Validate during owner-update request validation: visibility %q is not an accepted value; the owner update operation accepts only %s, because organization-scoped visibility is deferred; send one of the accepted values or omit visibility to leave it unchanged", r.Visibility.String(), TranscriptUpdateVisibilityMenu())
	}
	if r.License != nil && !r.License.IsValid() {
		return fmt.Errorf("transcript update validation failed at schema.TranscriptUpdateRequest.Validate during owner-update request validation: license %q is not on the canonical menu; the owner update operation accepts %s, or the empty string to clear a license that was never granted; send a menu license or omit license to leave it unchanged", r.License.String(), LicenseMenu())
	}
	return nil
}

// transcriptUpdateFields is the closed set of property names this operation
// accepts, paired with the JSON tags on TranscriptUpdateRequest. It is the one
// place the wire vocabulary is written down, so the strict decoder and the
// generated schema cannot disagree about which names exist.
var transcriptUpdateFields = []string{"title", "description", "visibility", "license"}

// UnmarshalJSON decodes an owner-update body at the trust boundary, refusing
// two things the permissive default would wave through: an unknown field, and
// an explicit null.
//
// Both refusals exist because the alternative is a SILENT wrong answer rather
// than a loud one. An unknown field (notably "tags", which the village decodes
// and discards) would otherwise be dropped and reported as success. An explicit
// null would otherwise decode to the same nil pointer an omitted field
// produces, so a caller who sent null meaning "clear this" would get "leave it
// unchanged" and no indication anything was ignored.
func (r *TranscriptUpdateRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&raw); err != nil {
		return fmt.Errorf("transcript update decoding failed at schema.TranscriptUpdateRequest.UnmarshalJSON during owner-update request decoding: the body is not a JSON object: %w; send an object carrying any of %s, or an empty object to change nothing", err, strings.Join(transcriptUpdateFields, ", "))
	}

	known := make(map[string]struct{}, len(transcriptUpdateFields))
	for _, field := range transcriptUpdateFields {
		known[field] = struct{}{}
	}
	for name, value := range raw {
		if _, ok := known[name]; !ok {
			return fmt.Errorf("transcript update decoding failed at schema.TranscriptUpdateRequest.UnmarshalJSON during owner-update request decoding: %q is not a field of this operation; the server would accept and silently discard it, so a caller would believe it applied; the accepted fields are %s", name, strings.Join(transcriptUpdateFields, ", "))
		}
		if string(bytes.TrimSpace(value)) == "null" {
			return fmt.Errorf("transcript update decoding failed at schema.TranscriptUpdateRequest.UnmarshalJSON during owner-update request decoding: %q is explicitly null, which this operation refuses because null would mean preserve rather than the clear a caller usually intends; omit %q to leave it unchanged, or send an empty string to clear it", name, name)
		}
	}

	type alias TranscriptUpdateRequest
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("transcript update decoding failed at schema.TranscriptUpdateRequest.UnmarshalJSON during owner-update request decoding: %w; check that each field carries the type the contract declares", err)
	}
	*r = TranscriptUpdateRequest(decoded)
	return nil
}
