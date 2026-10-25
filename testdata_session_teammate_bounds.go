package schema

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/typescript/session_teammate_bounds.yaml
var sessionTeammateBoundsYAML []byte

// SessionTeammateBoundsInput is one identity offered to every layer that bounds
// it: the team, the name, and both length measurements so the metric itself is
// asserted rather than assumed.
type SessionTeammateBoundsInput struct {
	Team       string `json:"team" yaml:"team"`
	Name       string `json:"name" yaml:"name"`
	Evidence   string `json:"evidence" yaml:"evidence"`
	CodePoints int    `json:"codePoints" yaml:"code_points"`
	UTF16Units int    `json:"utf16Units" yaml:"utf16_units"`
}

// SessionTeammateBoundsExpectation is the verdict every layer must give.
type SessionTeammateBoundsExpectation struct {
	Valid bool `json:"valid" yaml:"valid"`
}

// SessionTeammateBoundsFixtures is the declared identity bound and the corpus
// shared by the Go, JSON Schema, and generated Zod checks.
type SessionTeammateBoundsFixtures struct {
	// Bound is the declared code-point bound. The loader requires it to equal
	// SessionTeammateMaxIdentityRunes so the corpus cannot drift from the rule.
	Bound int `yaml:"bound"`
	// RequiredNames lists the cases that must exist.
	RequiredNames []string                                                                      `yaml:"required_names"`
	Cases         testcase.Corpus[SessionTeammateBoundsInput, SessionTeammateBoundsExpectation] `yaml:"cases"`
}

// LoadSessionTeammateBoundsFixtures parses and boundary-validates the corpus.
func LoadSessionTeammateBoundsFixtures() (*SessionTeammateBoundsFixtures, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(sessionTeammateBoundsYAML))
	decoder.KnownFields(true)

	var f SessionTeammateBoundsFixtures
	if err := decoder.Decode(&f); err != nil {
		return nil, fmt.Errorf("load session teammate bounds: decode corpus document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("load session teammate bounds: decode trailing document: %w", err)
		}
		return nil, fmt.Errorf("load session teammate bounds: multiple YAML documents are not allowed")
	}
	if f.Bound != SessionTeammateMaxIdentityRunes {
		return nil, fmt.Errorf("load session teammate bounds: bound is %d, but schema.SessionTeammateMaxIdentityRunes is %d; the corpus pins every layer to one rule, so both must move together", f.Bound, SessionTeammateMaxIdentityRunes)
	}
	if len(f.RequiredNames) == 0 {
		return nil, fmt.Errorf("load session teammate bounds: required_names is empty; a corpus with no named case can lose coverage silently")
	}
	if err := f.Cases.Validate(); err != nil {
		return nil, fmt.Errorf("load session teammate bounds: cases: %w", err)
	}
	seen := make(map[string]struct{}, len(f.Cases.Cases))
	for _, c := range f.Cases.Cases {
		seen[c.Name] = struct{}{}
	}
	for _, name := range f.RequiredNames {
		if _, ok := seen[name]; !ok {
			return nil, fmt.Errorf("load session teammate bounds: required case %q is missing", name)
		}
	}
	if len(seen) != len(f.RequiredNames) {
		return nil, fmt.Errorf("load session teammate bounds: required_names names %d cases but the corpus carries %d", len(f.RequiredNames), len(seen))
	}
	return &f, nil
}
