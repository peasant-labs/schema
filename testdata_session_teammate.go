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

//go:embed testdata/local-api/session_teammate.yaml
var sessionTeammateYAML []byte

// SessionTeammateSchemaProbe is one evidence token offered to the published
// SessionTeammateEvidence component from the generated Types document.
type SessionTeammateSchemaProbe struct {
	Evidence string `json:"evidence" yaml:"evidence"`
}

// SessionTeammateSchemaExpectation is the published schema's verdict on a
// probe. ErrorContains is authored in the corpus as an independent literal,
// never taken from a production message constant.
type SessionTeammateSchemaExpectation struct {
	Accepted      bool   `json:"accepted" yaml:"accepted"`
	ErrorContains string `json:"errorContains" yaml:"error_contains"`
}

// SessionTeammateIdentityProbe is one complete identity object offered to the
// published SessionTeammate component, so the team, name, and evidence rules are
// proven together rather than one field at a time.
type SessionTeammateIdentityProbe struct {
	Team     string `json:"team" yaml:"team"`
	Name     string `json:"name" yaml:"name"`
	Evidence string `json:"evidence" yaml:"evidence"`
}

// SessionTeammateFixtures is the declared published shape of the local
// teammate identity: which component carries it, which closed set its evidence
// field references, the identity bound and required members, and the tokens and
// objects offered to the published component.
type SessionTeammateFixtures struct {
	// Component is the published component carrying the identity object.
	Component string `yaml:"component"`
	// EvidenceField is the property whose value must reference the closed set.
	EvidenceField string `yaml:"evidence_field"`
	// EvidenceComponent is the closed-set component the field must reference.
	EvidenceComponent string `yaml:"evidence_component"`
	// EvidenceRef is the exact $ref the field must publish.
	EvidenceRef string `yaml:"evidence_ref"`
	// IdentityMaxLength is the declared character bound of team and name. The
	// loader requires it to equal SessionTeammateMaxIdentityRunes, so the
	// contract and the Go rule cannot drift apart silently.
	IdentityMaxLength int `yaml:"identity_max_length"`
	// IdentityFields are the properties the published bound applies to.
	IdentityFields []string `yaml:"identity_fields"`
	// IdentityRequired are the members the published component must require.
	IdentityRequired []string `yaml:"identity_required"`
	// RequiredMenu is the declared evidence menu, compared to
	// AllSessionTeammateEvidences by membership.
	RequiredMenu []SessionTeammateEvidence `yaml:"required_menu"`
	// SchemaProbes drives the published evidence component.
	SchemaProbes testcase.Corpus[SessionTeammateSchemaProbe, SessionTeammateSchemaExpectation] `yaml:"schema_probes"`
	// IdentityProbes drives the published identity object.
	IdentityProbes testcase.Corpus[SessionTeammateIdentityProbe, SessionTeammateSchemaExpectation] `yaml:"identity_probes"`
}

// LoadSessionTeammateFixtures parses the corpus and boundary-validates it. It
// rejects unknown fields, trailing documents, an under-populated component
// block, an evidence ref that does not name the declared component, a menu that
// disagrees with the Go closed set, a verdict inconsistent with its
// classification, and a refusal naming no expected fragment or a fragment that
// omits a menu member.
func LoadSessionTeammateFixtures() (*SessionTeammateFixtures, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(sessionTeammateYAML))
	decoder.KnownFields(true)

	var f SessionTeammateFixtures
	if err := decoder.Decode(&f); err != nil {
		return nil, fmt.Errorf("load session teammate fixtures: decode corpus document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("load session teammate fixtures: decode trailing document: %w", err)
		}
		return nil, fmt.Errorf("load session teammate fixtures: multiple YAML documents are not allowed")
	}

	if f.Component == "" || f.EvidenceField == "" || f.EvidenceComponent == "" || f.EvidenceRef == "" {
		return nil, fmt.Errorf("load session teammate fixtures: the published component block is under-populated (component=%q evidenceField=%q evidenceComponent=%q evidenceRef=%q); an empty field would make its probe assert nothing", f.Component, f.EvidenceField, f.EvidenceComponent, f.EvidenceRef)
	}
	wantRef := "#/components/schemas/" + f.EvidenceComponent
	if f.EvidenceRef != wantRef {
		return nil, fmt.Errorf("load session teammate fixtures: evidence ref %q does not name the declared component; the probe compares the published $ref against this literal, so it must be %q", f.EvidenceRef, wantRef)
	}
	if err := sameTeammateEvidenceMenu(f.RequiredMenu); err != nil {
		return nil, fmt.Errorf("load session teammate fixtures: %w", err)
	}
	if f.IdentityMaxLength != SessionTeammateMaxIdentityRunes {
		return nil, fmt.Errorf("load session teammate fixtures: identity_max_length is %d, but schema.SessionTeammateMaxIdentityRunes is %d; the corpus pins the published bound to the Go rule, so both must move together", f.IdentityMaxLength, SessionTeammateMaxIdentityRunes)
	}
	if err := requireUniqueNonEmpty("identity_fields", f.IdentityFields); err != nil {
		return nil, fmt.Errorf("load session teammate fixtures: %w", err)
	}
	if err := requireUniqueNonEmpty("identity_required", f.IdentityRequired); err != nil {
		return nil, fmt.Errorf("load session teammate fixtures: %w", err)
	}
	if err := f.SchemaProbes.Validate(); err != nil {
		return nil, fmt.Errorf("load session teammate fixtures: schema_probes: %w", err)
	}
	if err := f.IdentityProbes.Validate(); err != nil {
		return nil, fmt.Errorf("load session teammate fixtures: identity_probes: %w", err)
	}

	accepted := make(map[string]struct{}, len(f.RequiredMenu))
	for _, c := range f.SchemaProbes.Cases {
		if c.Expected.Accepted != (c.Classification == testcase.MustPass) {
			return nil, fmt.Errorf("load session teammate fixtures: schema_probes case %q is classified %q but its expectation says accepted=%t", c.Name, c.Classification, c.Expected.Accepted)
		}
		if c.Expected.Accepted {
			if c.Expected.ErrorContains != "" {
				return nil, fmt.Errorf("load session teammate fixtures: schema_probes case %q is accepted but names expected error fragment %q; an accepted value produces no error", c.Name, c.Expected.ErrorContains)
			}
			accepted[c.Input.Evidence] = struct{}{}
			continue
		}
		if c.Expected.ErrorContains == "" {
			return nil, fmt.Errorf("load session teammate fixtures: schema_probes case %q is a refusal but names no expected error fragment; without one the row would pass on any message and prove nothing about why the value was refused", c.Name)
		}
		// A refusal needle must name the whole menu: substring matching would let
		// a truncated needle keep passing while a member was dropped.
		for _, evidence := range AllSessionTeammateEvidences {
			if !strings.Contains(c.Expected.ErrorContains, evidence.String()) {
				return nil, fmt.Errorf("load session teammate fixtures: schema_probes case %q names refusal fragment %q, which does not mention menu member %q", c.Name, c.Expected.ErrorContains, evidence)
			}
		}
	}
	for _, evidence := range f.RequiredMenu {
		if _, ok := accepted[evidence.String()]; !ok {
			return nil, fmt.Errorf("load session teammate fixtures: menu member %q is accepted by no probe; a closed set with no accepted row can be emptied without the corpus noticing", evidence)
		}
	}

	acceptedIdentities, refusedIdentities := 0, 0
	menuMembers := make(map[string]struct{}, len(AllSessionTeammateEvidences))
	for _, evidence := range AllSessionTeammateEvidences {
		menuMembers[evidence.String()] = struct{}{}
	}
	for _, c := range f.IdentityProbes.Cases {
		if c.Expected.Accepted != (c.Classification == testcase.MustPass) {
			return nil, fmt.Errorf("load session teammate fixtures: identity_probes case %q is classified %q but its expectation says accepted=%t", c.Name, c.Classification, c.Expected.Accepted)
		}
		if c.Expected.Accepted {
			if c.Expected.ErrorContains != "" {
				return nil, fmt.Errorf("load session teammate fixtures: identity_probes case %q is accepted but names expected error fragment %q; an accepted value produces no error", c.Name, c.Expected.ErrorContains)
			}
			if _, ok := menuMembers[c.Input.Evidence]; !ok {
				return nil, fmt.Errorf("load session teammate fixtures: identity_probes case %q is accepted with evidence %q, which is outside AllSessionTeammateEvidences; an accepted object must be one the wire can carry", c.Name, c.Input.Evidence)
			}
			acceptedIdentities++
			continue
		}
		if c.Expected.ErrorContains == "" {
			return nil, fmt.Errorf("load session teammate fixtures: identity_probes case %q is a refusal but names no expected error fragment; without one the row would pass on any message and prove nothing about why the object was refused", c.Name)
		}
		refusedIdentities++
	}
	if acceptedIdentities == 0 || refusedIdentities == 0 {
		return nil, fmt.Errorf("load session teammate fixtures: identity_probes has %d accepted and %d refused cases; the arm must prove both that a recorded identity is accepted and that a broken one is refused", acceptedIdentities, refusedIdentities)
	}
	return &f, nil
}

// requireUniqueNonEmpty rejects a declared name list that is empty, holds a
// blank entry, or repeats one: each entry is a property the published probe
// looks up, so a blank or duplicated name would assert nothing.
func requireUniqueNonEmpty(axis string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s is empty; the published probe would assert nothing", axis)
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s holds a blank entry", axis)
		}
		if _, dup := seen[value]; dup {
			return fmt.Errorf("%s names %q twice", axis, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

// sameTeammateEvidenceMenu compares the corpus-declared menu to the Go closed
// set by membership in both directions, so widening the set without a row fails
// here rather than silently reducing coverage.
func sameTeammateEvidenceMenu(declared []SessionTeammateEvidence) error {
	want := make(map[SessionTeammateEvidence]struct{}, len(AllSessionTeammateEvidences))
	for _, evidence := range AllSessionTeammateEvidences {
		want[evidence] = struct{}{}
	}
	seen := make(map[SessionTeammateEvidence]struct{}, len(declared))
	for _, evidence := range declared {
		if _, ok := want[evidence]; !ok {
			return fmt.Errorf("the corpus declares teammate evidence %q, which is not a member of AllSessionTeammateEvidences", evidence)
		}
		if _, dup := seen[evidence]; dup {
			return fmt.Errorf("the corpus declares teammate evidence %q twice", evidence)
		}
		seen[evidence] = struct{}{}
	}
	for _, evidence := range AllSessionTeammateEvidences {
		if _, ok := seen[evidence]; !ok {
			return fmt.Errorf("AllSessionTeammateEvidences carries %q, which the corpus does not declare; widening the set requires a row for the new member", evidence)
		}
	}
	return nil
}
