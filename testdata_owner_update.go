package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

// OwnerUpdateFixtures is the segmented owner-update corpus: one named arm per
// behavior under test. The arms are heterogeneous on purpose. Request validation
// works on raw client bytes, encoding works on a Go-side intent, and the spec
// probes work on the generated document, so a single input type would have to be
// a union that hides which behavior a row actually exercises.
type OwnerUpdateFixtures struct {
	// RequiredBehaviours is the closed set, declared in the corpus itself so both
	// language bindings read ONE source instead of each keeping a copy. A
	// hand-mirrored list in the other language protects only the members someone
	// remembered to copy, while claiming both sides are guarded.
	RequiredBehaviours []OwnerUpdateRequestBehaviour `yaml:"required_behaviours"`
	// RequiredDescriptionAnchors is the count the anchors row must carry. The
	// anchors are prose, so a closed set does not fit them; a count in the corpus
	// at least makes a silent deletion from the row fail.
	RequiredDescriptionAnchors int                                                               `yaml:"required_description_anchors"`
	RequestValidations         testcase.Corpus[string, OwnerUpdateValidationExpectation]         `yaml:"request_validations"`
	Encodings                  testcase.Corpus[OwnerUpdateEncodingInput, string]                 `yaml:"encodings"`
	SpecExpectations           testcase.Corpus[OwnerUpdateSpecProbe, OwnerUpdateSpecExpectation] `yaml:"spec_expectations"`
}

// OwnerUpdateRequestBehaviour is the closed set of contract behaviours the
// request_validations arm must cover.
//
// It exists because a row-count floor cannot protect a corpus. A minimum with
// slack in it lets rows carrying a whole behaviour be deleted without any
// assertion noticing, and a fixture row is not code, so no mutation of the
// production source can reach it. Naming the behaviours and asserting coverage
// per member makes the corpus itself guarded input: losing every null-refusal
// row now fails, where losing six of twenty rows previously did not.
type OwnerUpdateRequestBehaviour string

const (
	// OwnerUpdateBehaviourVisibilityAccepted covers an accepted visibility value.
	OwnerUpdateBehaviourVisibilityAccepted OwnerUpdateRequestBehaviour = "visibility_accepted"
	// OwnerUpdateBehaviourVisibilityRefused covers a visibility outside the menu.
	OwnerUpdateBehaviourVisibilityRefused OwnerUpdateRequestBehaviour = "visibility_refused"
	// OwnerUpdateBehaviourLicenseAccepted covers a canonical menu license.
	OwnerUpdateBehaviourLicenseAccepted OwnerUpdateRequestBehaviour = "license_accepted"
	// OwnerUpdateBehaviourLicenseClearAccepted covers the clear sentinel, the
	// value that makes un-licensing requestable and therefore refusable.
	OwnerUpdateBehaviourLicenseClearAccepted OwnerUpdateRequestBehaviour = "license_clear_accepted"
	// OwnerUpdateBehaviourLicenseRefused covers a license off the menu.
	OwnerUpdateBehaviourLicenseRefused OwnerUpdateRequestBehaviour = "license_refused"
	// OwnerUpdateBehaviourNullRefused covers an explicit JSON null, which the
	// server would read as preserve rather than the clear a caller intends.
	OwnerUpdateBehaviourNullRefused OwnerUpdateRequestBehaviour = "null_refused"
	// OwnerUpdateBehaviourUnknownFieldRefused covers a field the server accepts
	// and silently discards.
	OwnerUpdateBehaviourUnknownFieldRefused OwnerUpdateRequestBehaviour = "unknown_field_refused"
	// OwnerUpdateBehaviourOmissionAccepted covers omission meaning unchanged.
	OwnerUpdateBehaviourOmissionAccepted OwnerUpdateRequestBehaviour = "omission_accepted"
	// OwnerUpdateBehaviourEmptyStringAccepted covers the empty string as a real
	// clearing value rather than an omission.
	OwnerUpdateBehaviourEmptyStringAccepted OwnerUpdateRequestBehaviour = "empty_string_accepted"
	// OwnerUpdateBehaviourTitleBoundEnforced covers the declared title bound at
	// its two decisive points. The bound was previously asserted only as text in
	// the published document, so the Go validator accepted a title the generated
	// JavaScript validator refused.
	OwnerUpdateBehaviourTitleBoundEnforced OwnerUpdateRequestBehaviour = "title_bound_enforced"
)

// AllOwnerUpdateRequestBehaviours is the canonical ordered closed set. The suite
// derives its minimum from this list AND asserts every member is exercised, so
// adding a behaviour without a row fails rather than silently reducing coverage.
var AllOwnerUpdateRequestBehaviours = []OwnerUpdateRequestBehaviour{
	OwnerUpdateBehaviourVisibilityAccepted,
	OwnerUpdateBehaviourVisibilityRefused,
	OwnerUpdateBehaviourLicenseAccepted,
	OwnerUpdateBehaviourLicenseClearAccepted,
	OwnerUpdateBehaviourLicenseRefused,
	OwnerUpdateBehaviourNullRefused,
	OwnerUpdateBehaviourUnknownFieldRefused,
	OwnerUpdateBehaviourOmissionAccepted,
	OwnerUpdateBehaviourEmptyStringAccepted,
	OwnerUpdateBehaviourTitleBoundEnforced,
}

// IsValid reports whether the behaviour is a known member.
func (b OwnerUpdateRequestBehaviour) IsValid() bool {
	for _, known := range AllOwnerUpdateRequestBehaviours {
		if b == known {
			return true
		}
	}
	return false
}

// OwnerUpdateValidationExpectation is the verdict for one raw request body:
// whether the contract accepts it and, for a refusal, a distinctive fragment of
// the actionable message. ErrorContains is empty exactly when Accepted is true.
type OwnerUpdateValidationExpectation struct {
	Accepted      bool                        `yaml:"accepted"`
	ErrorContains string                      `yaml:"error_contains"`
	Behaviour     OwnerUpdateRequestBehaviour `yaml:"behaviour"`
}

// OwnerUpdateEncodingInput is a Go-side update intent expressed so the fixture
// can distinguish "field not set" from "field set to an empty value". Each
// set_* flag makes the corresponding value participate; without it the field
// stays nil and must not reach the wire. That separation is the whole point of
// the arm: it is what proves preserve, clear, and replace stay distinguishable.
type OwnerUpdateEncodingInput struct {
	SetTitle       bool   `yaml:"set_title"`
	Title          string `yaml:"title"`
	SetDescription bool   `yaml:"set_description"`
	Description    string `yaml:"description"`
	SetVisibility  string `yaml:"set_visibility"`
	SetLicense     bool   `yaml:"set_license"`
	License        string `yaml:"license"`
}

// ToRequest builds the real TranscriptUpdateRequest a caller would construct.
// It returns the production type rather than a test-local mirror so the encoding
// arm exercises the shipped marshaling behavior.
func (in OwnerUpdateEncodingInput) ToRequest() TranscriptUpdateRequest {
	var r TranscriptUpdateRequest
	if in.SetTitle {
		title := in.Title
		r.Title = &title
	}
	if in.SetDescription {
		description := in.Description
		r.Description = &description
	}
	if in.SetVisibility != "" {
		visibility := TranscriptUpdateVisibility(in.SetVisibility)
		r.Visibility = &visibility
	}
	if in.SetLicense {
		license := TranscriptUpdateLicense(in.License)
		r.License = &license
	}
	return r
}

// OwnerUpdateSpecProbeKind is the closed set of generated-document properties the
// spec arm can assert. It is closed so a fixture cannot silently name a probe no
// test implements, which would let a row pass by doing nothing.
type OwnerUpdateSpecProbeKind string

const (
	// OwnerUpdateProbeOperationDeclared asserts the operation exists and carries
	// its declared operation id.
	OwnerUpdateProbeOperationDeclared OwnerUpdateSpecProbeKind = "operation_declared"
	// OwnerUpdateProbePathParameter asserts the transcript id is a required path
	// parameter.
	OwnerUpdateProbePathParameter OwnerUpdateSpecProbeKind = "path_parameter"
	// OwnerUpdateProbeBodyIsReference asserts the request body is a component
	// reference rather than an inline object.
	OwnerUpdateProbeBodyIsReference OwnerUpdateSpecProbeKind = "body_is_reference"
	// OwnerUpdateProbeResponseStatuses asserts the declared refusal status set.
	OwnerUpdateProbeResponseStatuses OwnerUpdateSpecProbeKind = "response_statuses"
	// OwnerUpdateProbeVisibilityEnum asserts the declared visibility members.
	OwnerUpdateProbeVisibilityEnum OwnerUpdateSpecProbeKind = "visibility_enum"
	// OwnerUpdateProbeLicenseEnum asserts the declared license members.
	OwnerUpdateProbeLicenseEnum OwnerUpdateSpecProbeKind = "license_enum"
	// OwnerUpdateProbeBodyProperties asserts the exact property set the body
	// declares. Without it the payload, which is the operation's whole subject,
	// has nothing pinning its field names.
	OwnerUpdateProbeBodyProperties OwnerUpdateSpecProbeKind = "body_properties"
	// OwnerUpdateProbePathParameterIsCanonical asserts the path parameter is the
	// canonical validated identifier rather than a bare string, so the contract
	// stops describing ids the server refuses outright.
	OwnerUpdateProbePathParameterIsCanonical OwnerUpdateSpecProbeKind = "path_parameter_is_canonical"
	// OwnerUpdateProbeTitleMaxLength asserts the declared title bound matches the
	// storage column, so an over-long title is refused before it becomes an
	// opaque server error.
	OwnerUpdateProbeTitleMaxLength OwnerUpdateSpecProbeKind = "title_max_length"
	// OwnerUpdateProbeBodyIsRequired asserts the request body is declared
	// required. Reflection defaults it to optional, which would describe a
	// request the server always refuses.
	OwnerUpdateProbeBodyIsRequired OwnerUpdateSpecProbeKind = "body_is_required"
	// OwnerUpdateProbeDescriptionAnchors asserts the operation description still
	// carries the claims a consumer relies on. The description is the only place
	// several contract decisions are explained, and prose is otherwise unguarded:
	// deleting it changes no assertion and breaks no build.
	OwnerUpdateProbeDescriptionAnchors OwnerUpdateSpecProbeKind = "description_anchors"
	// OwnerUpdateProbeRefusalEnvelopeProperties asserts the EXACT property set of
	// the refusal envelope. Asserting that a property exists cannot detect a
	// property that appears: an invented field on the envelope shipped with the
	// whole suite green, because nothing said which properties the envelope has.
	// This is the response-side twin of the request body's property probe.
	OwnerUpdateProbeRefusalEnvelopeProperties OwnerUpdateSpecProbeKind = "refusal_envelope_properties"
	// OwnerUpdateProbeBodyIsClosed asserts the body rejects unknown properties
	// rather than accepting and discarding them.
	OwnerUpdateProbeBodyIsClosed OwnerUpdateSpecProbeKind = "body_is_closed"
	// OwnerUpdateProbeSuccessHasNoBody asserts the success status is declared but
	// carries no content schema. It guards the deliberate omission in BOTH
	// directions: dropping the status, or inventing a success body the village
	// does not serve, each fail.
	OwnerUpdateProbeSuccessHasNoBody OwnerUpdateSpecProbeKind = "success_has_no_body"
)

// AllOwnerUpdateSpecProbeKinds is the canonical ordered closed set. The spec test
// asserts the corpus covers every member, so adding a probe kind without a
// fixture row fails rather than silently reducing coverage.
var AllOwnerUpdateSpecProbeKinds = []OwnerUpdateSpecProbeKind{
	OwnerUpdateProbeOperationDeclared,
	OwnerUpdateProbePathParameter,
	OwnerUpdateProbeBodyIsReference,
	OwnerUpdateProbeResponseStatuses,
	OwnerUpdateProbeVisibilityEnum,
	OwnerUpdateProbeLicenseEnum,
	OwnerUpdateProbeSuccessHasNoBody,
	OwnerUpdateProbeBodyProperties,
	OwnerUpdateProbeBodyIsClosed,
	OwnerUpdateProbeDescriptionAnchors,
	OwnerUpdateProbeBodyIsRequired,
	OwnerUpdateProbePathParameterIsCanonical,
	OwnerUpdateProbeTitleMaxLength,
	OwnerUpdateProbeRefusalEnvelopeProperties,
}

// IsValid reports whether the probe kind is a known member.
func (k OwnerUpdateSpecProbeKind) IsValid() bool {
	for _, known := range AllOwnerUpdateSpecProbeKinds {
		if k == known {
			return true
		}
	}
	return false
}

func (k OwnerUpdateSpecProbeKind) String() string { return string(k) }

// OwnerUpdateSpecProbe selects which generated-document property a row asserts.
type OwnerUpdateSpecProbe struct {
	Probe OwnerUpdateSpecProbeKind `yaml:"probe"`
}

// OwnerUpdateSpecExpectation is the exact set of strings a probe must observe.
// It is an exact set rather than a subset so a widening (an extra enum member, an
// extra declared status) fails the row and has to be stated deliberately.
type OwnerUpdateSpecExpectation struct {
	Strings []string `yaml:"strings"`
}

// verifyBehaviourMatchesRow checks that a row actually EXHIBITS the behaviour it
// claims. Without this the coverage guard's only input is a hand-written label
// nothing cross-checks, so deleting every row of a behaviour and moving its label
// onto an unrelated surviving row leaves the guard reporting the behaviour
// covered while no row exercises it. The label is data like any other and needs
// its own validation.
func verifyBehaviourMatchesRow(name, input string, expect OwnerUpdateValidationExpectation) error {
	known := map[string]struct{}{"title": {}, "description": {}, "visibility": {}, "license": {}}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &decoded); err != nil {
		return fmt.Errorf("request_validations case %q has an input that is not a JSON object, so no behaviour claim about it can be checked: %w", name, err)
	}

	hasNull := false
	hasUnknown := false
	for field, raw := range decoded {
		if string(raw) == "null" {
			hasNull = true
		}
		if _, ok := known[field]; !ok {
			hasUnknown = true
		}
	}

	switch expect.Behaviour {
	case OwnerUpdateBehaviourNullRefused:
		if !hasNull {
			return fmt.Errorf("request_validations case %q claims behaviour %q but its input %s carries no explicit JSON null; the coverage guard would report the behaviour exercised while nothing exercises it", name, expect.Behaviour, input)
		}
	case OwnerUpdateBehaviourUnknownFieldRefused:
		if !hasUnknown {
			return fmt.Errorf("request_validations case %q claims behaviour %q but every field in its input %s is a declared field; the coverage guard would report the behaviour exercised while nothing exercises it", name, expect.Behaviour, input)
		}
	case OwnerUpdateBehaviourLicenseClearAccepted:
		if string(decoded["license"]) != `""` {
			return fmt.Errorf("request_validations case %q claims behaviour %q but its input %s does not send the clear sentinel", name, expect.Behaviour, input)
		}
	case OwnerUpdateBehaviourVisibilityAccepted, OwnerUpdateBehaviourVisibilityRefused:
		if _, ok := decoded["visibility"]; !ok {
			return fmt.Errorf("request_validations case %q claims behaviour %q but its input %s carries no visibility field", name, expect.Behaviour, input)
		}
	case OwnerUpdateBehaviourLicenseAccepted, OwnerUpdateBehaviourLicenseRefused:
		if _, ok := decoded["license"]; !ok {
			return fmt.Errorf("request_validations case %q claims behaviour %q but its input %s carries no license field", name, expect.Behaviour, input)
		}
	case OwnerUpdateBehaviourTitleBoundEnforced:
		var title string
		if err := json.Unmarshal(decoded["title"], &title); err != nil {
			return fmt.Errorf("request_validations case %q claims behaviour %q but carries no string title", name, expect.Behaviour)
		}
		if n := len([]rune(title)); n != TranscriptUpdateTitleMaxLength && n != TranscriptUpdateTitleMaxLength+1 {
			return fmt.Errorf("request_validations case %q claims behaviour %q but its title is %d characters; a bound is proven at the bound and one past it, not at an arbitrary length", name, expect.Behaviour, n)
		}
	case OwnerUpdateBehaviourEmptyStringAccepted:
		found := false
		for _, raw := range decoded {
			if string(raw) == `""` {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("request_validations case %q claims behaviour %q but its input %s carries no empty-string value", name, expect.Behaviour, input)
		}
	}
	return nil
}

// LoadOwnerUpdateFixtures parses OwnerUpdateYAML into the segmented corpus. It
// rejects unknown fields, trailing documents, and any case missing its
// classification, provenance, or mutation metadata, so a row cannot enter the
// corpus without recording why it exists and what it changes.
func LoadOwnerUpdateFixtures() (*OwnerUpdateFixtures, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(OwnerUpdateYAML))
	decoder.KnownFields(true)

	var f OwnerUpdateFixtures
	if err := decoder.Decode(&f); err != nil {
		return nil, fmt.Errorf("load owner update fixtures: decode corpus document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("load owner update fixtures: decode trailing document: %w", err)
		}
		return nil, fmt.Errorf("load owner update fixtures: multiple YAML documents are not allowed")
	}

	if len(f.RequiredBehaviours) != len(AllOwnerUpdateRequestBehaviours) {
		return nil, fmt.Errorf("load owner update fixtures: the corpus declares %d required behaviours but the Go closed set has %d; the two must agree because both language bindings derive coverage from the corpus", len(f.RequiredBehaviours), len(AllOwnerUpdateRequestBehaviours))
	}
	for _, declared := range f.RequiredBehaviours {
		if !declared.IsValid() {
			return nil, fmt.Errorf("load owner update fixtures: the corpus declares required behaviour %q, which is not a member of the Go closed set", declared)
		}
	}
	if f.RequiredDescriptionAnchors <= 0 {
		return nil, fmt.Errorf("load owner update fixtures: the corpus declares no required description-anchor count, so anchors could be deleted from the row one at a time with nothing noticing")
	}
	if err := f.RequestValidations.Validate(); err != nil {
		return nil, fmt.Errorf("load owner update fixtures: request_validations: %w", err)
	}
	if err := f.Encodings.Validate(); err != nil {
		return nil, fmt.Errorf("load owner update fixtures: encodings: %w", err)
	}
	if err := f.SpecExpectations.Validate(); err != nil {
		return nil, fmt.Errorf("load owner update fixtures: spec_expectations: %w", err)
	}

	// A verdict must be internally consistent: an accepted body cannot also name
	// an expected error fragment, and a refusal that names none would pass on any
	// message at all, which is the classic vacuous negative.
	for _, c := range f.RequestValidations.Cases {
		if c.Expected.Accepted && c.Expected.ErrorContains != "" {
			return nil, fmt.Errorf("load owner update fixtures: request_validations case %q is accepted but names an expected error fragment %q; an accepted body produces no error", c.Name, c.Expected.ErrorContains)
		}
		if err := verifyBehaviourMatchesRow(c.Name, c.Input, c.Expected); err != nil {
			return nil, fmt.Errorf("load owner update fixtures: %w", err)
		}
		if !c.Expected.Behaviour.IsValid() {
			return nil, fmt.Errorf("load owner update fixtures: request_validations case %q names behaviour %q, which is not a member of the closed set; a row that names no known behaviour cannot contribute to coverage and would let a whole behaviour be deleted unnoticed", c.Name, c.Expected.Behaviour)
		}
		wantClassification := testcase.MustFail
		if c.Expected.Accepted {
			wantClassification = testcase.MustPass
		}
		if c.Classification != wantClassification {
			return nil, fmt.Errorf("load owner update fixtures: request_validations case %q is classified %q but its expectation says accepted=%t; the classification must agree with the verdict or it is decoration a reader would trust and a mutation could flip unnoticed; use %q", c.Name, c.Classification, c.Expected.Accepted, wantClassification)
		}
		if !c.Expected.Accepted && c.Expected.ErrorContains == "" {
			return nil, fmt.Errorf("load owner update fixtures: request_validations case %q is a refusal but names no expected error fragment; without one the row would pass on any message and prove nothing about why the body was refused", c.Name)
		}
	}
	for _, c := range f.SpecExpectations.Cases {
		if !c.Input.Probe.IsValid() {
			return nil, fmt.Errorf("load owner update fixtures: spec_expectations case %q names unknown probe %q; the probe set is closed so a row cannot assert a property no test implements; use one of the declared probe kinds", c.Name, c.Input.Probe)
		}
		if len(c.Expected.Strings) == 0 {
			return nil, fmt.Errorf("load owner update fixtures: spec_expectations case %q expects no strings; an empty expectation is satisfied by anything and would prove nothing", c.Name)
		}
	}
	return &f, nil
}
