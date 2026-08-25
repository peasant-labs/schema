package schema

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/session_origin.yaml
var sessionOriginYAML []byte

// SessionOriginBehaviour is the closed set of contract behaviours the declared
// session-origin corpus must cover.
//
// It exists because a row-count floor cannot protect a corpus: rows carrying a
// whole behaviour can be deleted with no assertion noticing, and a fixture row
// is not code, so no mutation of the production source can reach it. Naming the
// behaviours and asserting coverage per member makes the corpus itself guarded
// input.
type SessionOriginBehaviour string

const (
	// SessionOriginBehaviourMenuMemberAccepted covers a declared menu member.
	SessionOriginBehaviourMenuMemberAccepted SessionOriginBehaviour = "menu_member_accepted"
	// SessionOriginBehaviourAbsentDeclarationAccepted covers the empty string:
	// an absent declaration is valid and means the producer expressed no opinion.
	SessionOriginBehaviourAbsentDeclarationAccepted SessionOriginBehaviour = "absent_declaration_accepted"
	// SessionOriginBehaviourOutOfMenuRejected covers a token outside the menu,
	// which neither the Go predicate nor the published schema may accept.
	SessionOriginBehaviourOutOfMenuRejected SessionOriginBehaviour = "out_of_menu_rejected"
	// SessionOriginBehaviourDetailRoundTrip covers the field surviving a JSON
	// round trip on the session detail payload.
	SessionOriginBehaviourDetailRoundTrip SessionOriginBehaviour = "detail_round_trip"
	// SessionOriginBehaviourSummaryRoundTrip covers the same on the session
	// summary, which carries the field independently of the detail payload.
	SessionOriginBehaviourSummaryRoundTrip SessionOriginBehaviour = "summary_round_trip"
	// SessionOriginBehaviourAbsentFieldOmitted covers omitempty: an undeclared
	// origin must not appear on the wire as an empty string, because a consumer
	// would then have to treat "" as a fourth token.
	SessionOriginBehaviourAbsentFieldOmitted SessionOriginBehaviour = "absent_field_omitted"
	// SessionOriginBehaviourPublishedSchemaRefusal covers the published Types
	// component refusing an out-of-menu token, so a non-Go consumer inherits the
	// same closed set.
	SessionOriginBehaviourPublishedSchemaRefusal SessionOriginBehaviour = "published_schema_refusal"
)

// AllSessionOriginBehaviours is the canonical ordered closed set. The corpus
// declares the same set, and the loader compares the two by MEMBERSHIP, so
// adding a behaviour without a row fails rather than silently reducing coverage.
var AllSessionOriginBehaviours = []SessionOriginBehaviour{
	SessionOriginBehaviourMenuMemberAccepted,
	SessionOriginBehaviourAbsentDeclarationAccepted,
	SessionOriginBehaviourOutOfMenuRejected,
	SessionOriginBehaviourDetailRoundTrip,
	SessionOriginBehaviourSummaryRoundTrip,
	SessionOriginBehaviourAbsentFieldOmitted,
	SessionOriginBehaviourPublishedSchemaRefusal,
}

// IsValid reports whether b is a member of the closed behaviour set.
func (b SessionOriginBehaviour) IsValid() bool {
	for _, known := range AllSessionOriginBehaviours {
		if b == known {
			return true
		}
	}
	return false
}

func (b SessionOriginBehaviour) String() string { return string(b) }

// SessionOriginPayloadKind names which payload a round-trip row exercises. It is
// a closed set because the two payloads carry the field independently: a row
// that could name an unimplemented payload would assert nothing.
type SessionOriginPayloadKind string

const (
	// SessionOriginPayloadDetail is SessionDetailPayload.
	SessionOriginPayloadDetail SessionOriginPayloadKind = "session_detail"
	// SessionOriginPayloadSummary is SessionSummary.
	SessionOriginPayloadSummary SessionOriginPayloadKind = "session_summary"
)

// AllSessionOriginPayloadKinds is the canonical ordered closed set.
var AllSessionOriginPayloadKinds = []SessionOriginPayloadKind{
	SessionOriginPayloadDetail,
	SessionOriginPayloadSummary,
}

// IsValid reports whether k is a member of the closed payload set.
func (k SessionOriginPayloadKind) IsValid() bool {
	for _, known := range AllSessionOriginPayloadKinds {
		if k == known {
			return true
		}
	}
	return false
}

func (k SessionOriginPayloadKind) String() string { return string(k) }

// SessionOriginValidityExpectation is the verdict SessionOrigin.IsValid must
// return for a raw token, plus the behaviour the row contributes.
type SessionOriginValidityExpectation struct {
	Valid     bool                   `json:"valid" yaml:"valid"`
	Behaviour SessionOriginBehaviour `json:"behaviour" yaml:"behaviour"`
}

// SessionOriginRoundTripInput names a payload and the origin token to encode
// into it.
type SessionOriginRoundTripInput struct {
	Payload SessionOriginPayloadKind `json:"payload" yaml:"payload"`
	Origin  string                   `json:"origin" yaml:"origin"`
}

// SessionOriginRoundTripExpectation says what the encoded payload must look
// like and what must survive decoding it again. FieldPresent false asserts the
// omitempty behaviour: the key is absent, not present and empty.
type SessionOriginRoundTripExpectation struct {
	FieldPresent bool                   `json:"fieldPresent" yaml:"field_present"`
	EncodedJSON  string                 `json:"encodedJson" yaml:"encoded_json"`
	Decoded      string                 `json:"decoded" yaml:"decoded"`
	Behaviour    SessionOriginBehaviour `json:"behaviour" yaml:"behaviour"`
}

// SessionOriginSchemaProbe is one token offered to the published SessionOrigin
// component from the generated Types document.
type SessionOriginSchemaProbe struct {
	Origin string `json:"origin" yaml:"origin"`
}

// SessionOriginSchemaExpectation is the published schema's verdict on a probe.
//
// ErrorContains is authored in the corpus as an INDEPENDENT literal and is
// never taken from a production message constant: a needle derived from the code
// under test cannot detect that code changing.
type SessionOriginSchemaExpectation struct {
	Accepted      bool                   `json:"accepted" yaml:"accepted"`
	ErrorContains string                 `json:"errorContains" yaml:"error_contains"`
	Behaviour     SessionOriginBehaviour `json:"behaviour" yaml:"behaviour"`
}

// SessionSummariesByIDAnchor names one clause the by-id session-summaries
// operation description must state. The description is prose, so a closed set
// cannot be derived from it; naming each required clause makes deleting one fail
// here instead of quietly weakening a stated contract.
type SessionSummariesByIDAnchor string

const (
	// SessionSummariesByIDAnchorResolvesLinks: the operation exists to resolve
	// links, which is why the two scopes below do not apply to it.
	SessionSummariesByIDAnchorResolvesLinks SessionSummariesByIDAnchor = "resolves_links"
	// SessionSummariesByIDAnchorNoOriginScope: origin scope is a discovery
	// boundary and must not reach a by-id resolution.
	SessionSummariesByIDAnchorNoOriginScope SessionSummariesByIDAnchor = "no_origin_scope"
	// SessionSummariesByIDAnchorNoSelectionScope: selection scope is the same
	// kind of boundary and is excluded for the same reason.
	SessionSummariesByIDAnchorNoSelectionScope SessionSummariesByIDAnchor = "no_selection_scope"
	// SessionSummariesByIDAnchorUnknownIDOmitted: an identifier that names no
	// local session is omitted rather than failing the whole batch.
	SessionSummariesByIDAnchorUnknownIDOmitted SessionSummariesByIDAnchor = "unknown_identifier_omitted"
)

// AllSessionSummariesByIDAnchors is the canonical ordered closed set. The corpus
// declares one row per member and the loader compares the two by membership.
var AllSessionSummariesByIDAnchors = []SessionSummariesByIDAnchor{
	SessionSummariesByIDAnchorResolvesLinks,
	SessionSummariesByIDAnchorNoOriginScope,
	SessionSummariesByIDAnchorNoSelectionScope,
	SessionSummariesByIDAnchorUnknownIDOmitted,
}

// IsValid reports whether a is a member of the closed anchor set.
func (a SessionSummariesByIDAnchor) IsValid() bool {
	for _, known := range AllSessionSummariesByIDAnchors {
		if a == known {
			return true
		}
	}
	return false
}

func (a SessionSummariesByIDAnchor) String() string { return string(a) }

// SessionSummariesByIDDescriptionAnchor is one required clause: its closed-set
// name and the literal text the published description must contain.
type SessionSummariesByIDDescriptionAnchor struct {
	Name SessionSummariesByIDAnchor `json:"name" yaml:"name"`
	Text string                     `json:"text" yaml:"text"`
}

// SessionSummariesByIDContract is the declared shape of the by-id
// session-summaries operation in the Local API document.
type SessionSummariesByIDContract struct {
	Path               string                                  `json:"path" yaml:"path"`
	Method             string                                  `json:"method" yaml:"method"`
	OperationID        string                                  `json:"operationId" yaml:"operation_id"`
	QueryParameter     string                                  `json:"queryParameter" yaml:"query_parameter"`
	ResponseStatus     string                                  `json:"responseStatus" yaml:"response_status"`
	MediaType          string                                  `json:"mediaType" yaml:"media_type"`
	ResponseRef        string                                  `json:"responseRef" yaml:"response_ref"`
	DescriptionAnchors []SessionSummariesByIDDescriptionAnchor `json:"descriptionAnchors" yaml:"description_anchors"`
}

// sameAnchorSet compares the corpus-declared anchors to the Go closed set by
// membership in both directions, and refuses an anchor whose text is empty: an
// empty needle is satisfied by any description at all.
func sameAnchorSet(declared []SessionSummariesByIDDescriptionAnchor) error {
	want := make(map[SessionSummariesByIDAnchor]struct{}, len(AllSessionSummariesByIDAnchors))
	for _, anchor := range AllSessionSummariesByIDAnchors {
		want[anchor] = struct{}{}
	}
	seen := make(map[SessionSummariesByIDAnchor]struct{}, len(declared))
	for _, anchor := range declared {
		if _, ok := want[anchor.Name]; !ok {
			return fmt.Errorf("the corpus declares description anchor %q, which is not a member of AllSessionSummariesByIDAnchors", anchor.Name)
		}
		if _, dup := seen[anchor.Name]; dup {
			return fmt.Errorf("the corpus declares description anchor %q twice", anchor.Name)
		}
		if anchor.Text == "" {
			return fmt.Errorf("description anchor %q names no text; an empty needle is satisfied by any description, so the clause would be unguarded", anchor.Name)
		}
		seen[anchor.Name] = struct{}{}
	}
	for _, anchor := range AllSessionSummariesByIDAnchors {
		if _, ok := seen[anchor]; !ok {
			return fmt.Errorf("the Go closed set carries description anchor %q, which the corpus does not declare; deleting a stated clause must fail here", anchor)
		}
	}
	return nil
}

// SessionOriginFixtures is the segmented declared-session-origin corpus: one
// named arm per behaviour family. The arms are heterogeneous on purpose. The
// predicate arm works on raw tokens, the round-trip arm on encoded payloads, and
// the schema arm on the generated document, so a single input type would have to
// be a union that hides which behaviour a row actually exercises.
type SessionOriginFixtures struct {
	// RequiredBehaviours is the closed set, declared in the corpus itself so a
	// reader of the fixture sees the coverage contract without reading Go. The
	// loader compares it to AllSessionOriginBehaviours by membership.
	RequiredBehaviours []SessionOriginBehaviour `yaml:"required_behaviours"`
	// RequiredMenu is the declared menu the validity arm must exercise in full.
	// The loader compares it to AllSessionOrigins by membership, so widening the
	// enum without adding a row fails here.
	// ByIDOperation is the declared shape of the by-id session-summaries
	// operation, the link-resolution surface a hidden session still reaches.
	ByIDOperation SessionSummariesByIDContract                                                    `yaml:"by_id_operation"`
	RequiredMenu  []SessionOrigin                                                                 `yaml:"required_menu"`
	Validities    testcase.Corpus[string, SessionOriginValidityExpectation]                       `yaml:"validities"`
	RoundTrips    testcase.Corpus[SessionOriginRoundTripInput, SessionOriginRoundTripExpectation] `yaml:"round_trips"`
	SchemaProbes  testcase.Corpus[SessionOriginSchemaProbe, SessionOriginSchemaExpectation]       `yaml:"schema_probes"`
}

// LoadSessionOriginFixtures parses the corpus and boundary-validates it. It
// rejects unknown fields, trailing documents, a case missing its classification,
// provenance or mutation metadata, a verdict inconsistent with its
// classification, a refusal naming no expected error fragment, and any declared
// menu or behaviour the Go closed sets do not carry.
func LoadSessionOriginFixtures() (*SessionOriginFixtures, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(sessionOriginYAML))
	decoder.KnownFields(true)

	var f SessionOriginFixtures
	if err := decoder.Decode(&f); err != nil {
		return nil, fmt.Errorf("load session origin fixtures: decode corpus document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("load session origin fixtures: decode trailing document: %w", err)
		}
		return nil, fmt.Errorf("load session origin fixtures: multiple YAML documents are not allowed")
	}

	if err := sameBehaviourSet(f.RequiredBehaviours); err != nil {
		return nil, fmt.Errorf("load session origin fixtures: %w", err)
	}
	if err := sameOriginMenu(f.RequiredMenu); err != nil {
		return nil, fmt.Errorf("load session origin fixtures: %w", err)
	}
	if err := sameAnchorSet(f.ByIDOperation.DescriptionAnchors); err != nil {
		return nil, fmt.Errorf("load session origin fixtures: %w", err)
	}
	if f.ByIDOperation.Path == "" || f.ByIDOperation.Method == "" || f.ByIDOperation.OperationID == "" || f.ByIDOperation.QueryParameter == "" || f.ByIDOperation.ResponseRef == "" {
		return nil, fmt.Errorf("load session origin fixtures: the by-id operation block is under-populated (path=%q method=%q operationId=%q queryParameter=%q responseRef=%q); an empty field would make its probe assert nothing", f.ByIDOperation.Path, f.ByIDOperation.Method, f.ByIDOperation.OperationID, f.ByIDOperation.QueryParameter, f.ByIDOperation.ResponseRef)
	}

	if err := f.Validities.Validate(); err != nil {
		return nil, fmt.Errorf("load session origin fixtures: validities: %w", err)
	}
	if err := f.RoundTrips.Validate(); err != nil {
		return nil, fmt.Errorf("load session origin fixtures: round_trips: %w", err)
	}
	if err := f.SchemaProbes.Validate(); err != nil {
		return nil, fmt.Errorf("load session origin fixtures: schema_probes: %w", err)
	}

	for _, c := range f.Validities.Cases {
		if !c.Expected.Behaviour.IsValid() {
			return nil, fmt.Errorf("load session origin fixtures: validities case %q names behaviour %q, which is not a member of the closed set; a row naming no known behaviour cannot contribute to coverage, so a whole behaviour could be deleted unnoticed", c.Name, c.Expected.Behaviour)
		}
		if err := agreesWithClassification(c.Name, "validities", c.Classification, c.Expected.Valid); err != nil {
			return nil, fmt.Errorf("load session origin fixtures: %w", err)
		}
	}

	menuTokens := make(map[string]struct{}, len(AllSessionOrigins))
	for _, origin := range AllSessionOrigins {
		menuTokens[origin.String()] = struct{}{}
	}
	for _, c := range f.RoundTrips.Cases {
		if !c.Expected.Behaviour.IsValid() {
			return nil, fmt.Errorf("load session origin fixtures: round_trips case %q names behaviour %q, which is not a member of the closed set", c.Name, c.Expected.Behaviour)
		}
		if !c.Input.Payload.IsValid() {
			return nil, fmt.Errorf("load session origin fixtures: round_trips case %q names payload %q, which is not a member of the closed set; the row would assert a shape no test can build", c.Name, c.Input.Payload)
		}
		if c.Classification != testcase.MustPass {
			return nil, fmt.Errorf("load session origin fixtures: round_trips case %q is classified %q; every round trip is an encoding a producer must be able to make, so the arm carries must-pass rows only", c.Name, c.Classification)
		}
		if c.Expected.Decoded != c.Input.Origin {
			return nil, fmt.Errorf("load session origin fixtures: round_trips case %q encodes origin %q but expects to decode %q; a round trip that does not return its input proves nothing", c.Name, c.Input.Origin, c.Expected.Decoded)
		}
		if c.Expected.FieldPresent {
			if _, ok := menuTokens[c.Input.Origin]; !ok {
				return nil, fmt.Errorf("load session origin fixtures: round_trips case %q expects the field on the wire for origin %q, which is not a declared menu member", c.Name, c.Input.Origin)
			}
			if c.Expected.EncodedJSON == "" {
				return nil, fmt.Errorf("load session origin fixtures: round_trips case %q expects the field on the wire but names no encoded fragment; without one the row would pass on any encoding", c.Name)
			}
		} else {
			if c.Input.Origin != "" {
				return nil, fmt.Errorf("load session origin fixtures: round_trips case %q expects no field on the wire but declares origin %q; only an absent declaration is omitted", c.Name, c.Input.Origin)
			}
			if c.Expected.EncodedJSON != "" {
				return nil, fmt.Errorf("load session origin fixtures: round_trips case %q expects no field on the wire but names encoded fragment %q", c.Name, c.Expected.EncodedJSON)
			}
		}
	}

	for _, c := range f.SchemaProbes.Cases {
		if !c.Expected.Behaviour.IsValid() {
			return nil, fmt.Errorf("load session origin fixtures: schema_probes case %q names behaviour %q, which is not a member of the closed set", c.Name, c.Expected.Behaviour)
		}
		if err := agreesWithClassification(c.Name, "schema_probes", c.Classification, c.Expected.Accepted); err != nil {
			return nil, fmt.Errorf("load session origin fixtures: %w", err)
		}
		if c.Expected.Accepted && c.Expected.ErrorContains != "" {
			return nil, fmt.Errorf("load session origin fixtures: schema_probes case %q is accepted but names expected error fragment %q; an accepted value produces no error", c.Name, c.Expected.ErrorContains)
		}
		if !c.Expected.Accepted && c.Expected.ErrorContains == "" {
			return nil, fmt.Errorf("load session origin fixtures: schema_probes case %q is a refusal but names no expected error fragment; without one the row would pass on any message and prove nothing about why the value was refused", c.Name)
		}
		// A refusal needle must name the WHOLE menu. Substring matching means a
		// truncated needle still passes against the real message, so a row could
		// keep proving a refusal while quietly dropping a member from the menu it
		// claims the component publishes.
		for _, origin := range AllSessionOrigins {
			if c.Expected.Accepted {
				continue
			}
			if !strings.Contains(c.Expected.ErrorContains, origin.String()) {
				return nil, fmt.Errorf("load session origin fixtures: schema_probes case %q names refusal fragment %q, which does not mention menu member %q; a truncated needle still matches the real message by substring, so it would stop guarding the dropped member", c.Name, c.Expected.ErrorContains, origin)
			}
		}
	}

	if err := coversEveryBehaviour(&f); err != nil {
		return nil, fmt.Errorf("load session origin fixtures: %w", err)
	}
	if err := coversEveryMenuMember(&f); err != nil {
		return nil, fmt.Errorf("load session origin fixtures: %w", err)
	}
	return &f, nil
}

// agreesWithClassification keeps a row's classification from becoming decoration
// a reader trusts and a mutation could flip unnoticed.
func agreesWithClassification(name, arm string, classification testcase.Classification, accepted bool) error {
	want := testcase.MustFail
	if accepted {
		want = testcase.MustPass
	}
	if classification != want {
		return fmt.Errorf("%s case %q is classified %q but its expectation says accepted=%t; use %q", arm, name, classification, accepted, want)
	}
	return nil
}

// sameBehaviourSet compares the corpus-declared behaviours to the Go closed set
// by membership in both directions. Membership, not a count: a count churns on
// every addition and cannot tell a swap from a match.
func sameBehaviourSet(declared []SessionOriginBehaviour) error {
	want := make(map[SessionOriginBehaviour]struct{}, len(AllSessionOriginBehaviours))
	for _, behaviour := range AllSessionOriginBehaviours {
		want[behaviour] = struct{}{}
	}
	seen := make(map[SessionOriginBehaviour]struct{}, len(declared))
	for _, behaviour := range declared {
		if _, ok := want[behaviour]; !ok {
			return fmt.Errorf("the corpus declares required behaviour %q, which is not a member of the Go closed set AllSessionOriginBehaviours", behaviour)
		}
		if _, dup := seen[behaviour]; dup {
			return fmt.Errorf("the corpus declares required behaviour %q twice", behaviour)
		}
		seen[behaviour] = struct{}{}
	}
	for _, behaviour := range AllSessionOriginBehaviours {
		if _, ok := seen[behaviour]; !ok {
			return fmt.Errorf("the Go closed set carries behaviour %q, which the corpus does not declare; both language bindings derive coverage from the corpus, so the two must agree", behaviour)
		}
	}
	return nil
}

// sameOriginMenu compares the corpus-declared menu to AllSessionOrigins by
// membership, so the corpus cannot drift from the wire enum it guards.
func sameOriginMenu(declared []SessionOrigin) error {
	want := make(map[SessionOrigin]struct{}, len(AllSessionOrigins))
	for _, origin := range AllSessionOrigins {
		want[origin] = struct{}{}
	}
	seen := make(map[SessionOrigin]struct{}, len(declared))
	for _, origin := range declared {
		if _, ok := want[origin]; !ok {
			return fmt.Errorf("the corpus declares menu member %q, which is not a member of AllSessionOrigins", origin)
		}
		if _, dup := seen[origin]; dup {
			return fmt.Errorf("the corpus declares menu member %q twice", origin)
		}
		seen[origin] = struct{}{}
	}
	for _, origin := range AllSessionOrigins {
		if _, ok := seen[origin]; !ok {
			return fmt.Errorf("AllSessionOrigins carries %q, which the corpus does not declare; widening the enum requires a row for the new member", origin)
		}
	}
	return nil
}

// coversEveryBehaviour asserts each declared behaviour is exercised by at least
// one row somewhere in the corpus.
func coversEveryBehaviour(f *SessionOriginFixtures) error {
	covered := make(map[SessionOriginBehaviour]struct{}, len(AllSessionOriginBehaviours))
	for _, c := range f.Validities.Cases {
		covered[c.Expected.Behaviour] = struct{}{}
	}
	for _, c := range f.RoundTrips.Cases {
		covered[c.Expected.Behaviour] = struct{}{}
	}
	for _, c := range f.SchemaProbes.Cases {
		covered[c.Expected.Behaviour] = struct{}{}
	}
	for _, behaviour := range f.RequiredBehaviours {
		if _, ok := covered[behaviour]; !ok {
			return fmt.Errorf("required behaviour %q is exercised by no case; deleting the rows that carried it must fail here rather than silently reducing coverage", behaviour)
		}
	}
	return nil
}

// coversEveryMenuMember asserts every declared origin is accepted by a
// must-pass validity row AND survives a round trip on BOTH payloads, so the two
// placements are proven independently.
func coversEveryMenuMember(f *SessionOriginFixtures) error {
	accepted := make(map[string]struct{})
	for _, c := range f.Validities.Cases {
		if c.Expected.Valid {
			accepted[c.Input] = struct{}{}
		}
	}
	roundTripped := make(map[SessionOriginPayloadKind]map[string]struct{}, len(AllSessionOriginPayloadKinds))
	for _, kind := range AllSessionOriginPayloadKinds {
		roundTripped[kind] = make(map[string]struct{})
	}
	for _, c := range f.RoundTrips.Cases {
		roundTripped[c.Input.Payload][c.Input.Origin] = struct{}{}
	}
	for _, origin := range f.RequiredMenu {
		token := origin.String()
		if _, ok := accepted[token]; !ok {
			return fmt.Errorf("menu member %q is accepted by no validity case; membership is derived from AllSessionOrigins, so every member needs a row", token)
		}
		for _, kind := range AllSessionOriginPayloadKinds {
			if _, ok := roundTripped[kind][token]; !ok {
				return fmt.Errorf("menu member %q never round-trips on payload %q; the two placements carry the field independently and are proven independently", token, kind)
			}
		}
	}
	for _, kind := range AllSessionOriginPayloadKinds {
		if _, ok := roundTripped[kind][""]; !ok {
			return fmt.Errorf("payload %q has no absent-declaration round trip; omitempty is what keeps the empty string off the wire as a fourth token", kind)
		}
	}
	return nil
}
