package schema

import (
	"bytes"
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
	RequestValidations testcase.Corpus[string, OwnerUpdateValidationExpectation]         `yaml:"request_validations"`
	Encodings          testcase.Corpus[OwnerUpdateEncodingInput, string]                 `yaml:"encodings"`
	SpecExpectations   testcase.Corpus[OwnerUpdateSpecProbe, OwnerUpdateSpecExpectation] `yaml:"spec_expectations"`
}

// OwnerUpdateValidationExpectation is the verdict for one raw request body:
// whether the contract accepts it and, for a refusal, a distinctive fragment of
// the actionable message. ErrorContains is empty exactly when Accepted is true.
type OwnerUpdateValidationExpectation struct {
	Accepted      bool   `yaml:"accepted"`
	ErrorContains string `yaml:"error_contains"`
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
