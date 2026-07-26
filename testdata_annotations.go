package schema

import (
	"bytes"
	"fmt"
	"io"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

// AnnotationFixtures is the top-level YAML structure for annotation test data.
type AnnotationFixtures struct {
	Summaries         []AnnotationFixtureSummary                                                   `yaml:"annotation_summaries"`
	Validations       []SubscriptionValidationCase                                                 `yaml:"subscription_validations"`
	TargetValidations testcase.Corpus[AnnotationSummary, bool]                                     `yaml:"annotation_target_validations"`
	TargetRepairs     testcase.Corpus[annotationTargetRepairInput, annotationTargetRepairExpected] `yaml:"annotation_target_repairs"`
}

// AnnotationFixtureSummary is a named annotation template in YAML.
type AnnotationFixtureSummary struct {
	Name                string `yaml:"name"`
	ID                  string `yaml:"id"`
	TargetKind          string `yaml:"targetKind"`
	TargetSessionID     string `yaml:"targetSessionId,omitempty"`
	TargetEntryIndex    *int   `yaml:"targetEntryIndex,omitempty"`
	TargetEntryEndIndex *int   `yaml:"targetEntryEndIndex,omitempty"`
	TargetAnnotationID  string `yaml:"targetAnnotationId,omitempty"`
	TargetProjectHash   string `yaml:"targetProjectHash,omitempty"`
	TargetAssociationID string `yaml:"targetAssociationId,omitempty"`
	TargetFilePath      string `yaml:"targetFilePath,omitempty"`
	TargetContentHash   string `yaml:"targetContentHash,omitempty"`
	IsPrimary           bool   `yaml:"isPrimary"`
	AnnotatorKind       string `yaml:"annotatorKind"`
	AnnotatorName       string `yaml:"annotatorName"`
	TypeID              string `yaml:"typeId"`
	TypeName            string `yaml:"typeName"`
	Value               string `yaml:"value"`
	CreatedAt           int64  `yaml:"createdAt"`
}

// ToAnnotationSummary converts the fixture to a real AnnotationSummary.
func (f *AnnotationFixtureSummary) ToAnnotationSummary() AnnotationSummary {
	s := AnnotationSummary{
		ID:            f.ID,
		TargetKind:    TargetKind(f.TargetKind),
		IsPrimary:     f.IsPrimary,
		AnnotatorKind: AnnotatorKind(f.AnnotatorKind),
		AnnotatorName: f.AnnotatorName,
		TypeID:        f.TypeID,
		TypeName:      f.TypeName,
		Value:         f.Value,
		CreatedAt:     f.CreatedAt,
	}
	if f.TargetSessionID != "" {
		sid := f.TargetSessionID
		s.TargetSessionID = &sid
	}
	if f.TargetEntryIndex != nil {
		s.TargetEntryIndex = f.TargetEntryIndex
	}
	if f.TargetEntryEndIndex != nil {
		s.TargetEntryEndIndex = f.TargetEntryEndIndex
	}
	if f.TargetAnnotationID != "" {
		value := f.TargetAnnotationID
		s.TargetAnnotID = &value
	}
	if f.TargetProjectHash != "" {
		value := ProjectHash(f.TargetProjectHash)
		s.TargetProjectHash = &value
	}
	if f.TargetAssociationID != "" {
		value := AssociationID(f.TargetAssociationID)
		s.TargetAssociationID = &value
	}
	if f.TargetFilePath != "" {
		value := f.TargetFilePath
		s.TargetFilePath = &value
	}
	if f.TargetContentHash != "" {
		value := f.TargetContentHash
		s.TargetContentHash = &value
	}
	return s
}

// annotationTargetRepairKind names a fixture-owned in-place repair for an
// invalid annotation target fixture.
type annotationTargetRepairKind string

const (
	annotationTargetRepairClearSessionID annotationTargetRepairKind = "clear_target_session_id"
)

// annotationTargetRepairInput selects a rejected target-validation fixture and
// declares the one repair applied to it.
type annotationTargetRepairInput struct {
	SourceCase string                     `yaml:"sourceCase"`
	Kind       annotationTargetRepairKind `yaml:"kind"`
}

// annotationTargetRepairExpected records the production validator result
// before and after a target repair.
type annotationTargetRepairExpected struct {
	OriginalErrorContains string `yaml:"originalErrorContains"`
	PostMutationValid     bool   `yaml:"postMutationValid"`
}

// SubscriptionValidationCase is a test case for ValidateSubscription.
type SubscriptionValidationCase struct {
	Name  string `yaml:"name"`
	Topic string `yaml:"topic"`
	Axis  string `yaml:"axis,omitempty"`
	ID    string `yaml:"id,omitempty"`
	Valid bool   `yaml:"valid"`
}

// LoadAnnotationFixtures parses AnnotationsYAML into structured fixtures.
func LoadAnnotationFixtures() (*AnnotationFixtures, error) {
	var f AnnotationFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(AnnotationsYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&f); err != nil {
		return nil, fmt.Errorf("decode annotation fixtures: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("decode trailing annotation fixture document: %w", err)
		}
		return nil, fmt.Errorf("decode annotation fixtures: multiple YAML documents are not allowed")
	}
	if err := f.TargetValidations.Validate(); err != nil {
		return nil, fmt.Errorf("validate annotation target fixtures: %w", err)
	}
	if err := f.TargetRepairs.Validate(); err != nil {
		return nil, fmt.Errorf("validate annotation target repair fixtures: %w", err)
	}
	return &f, nil
}

// FindSummary returns the fixture annotation with the given name, or nil.
func (f *AnnotationFixtures) FindSummary(name string) *AnnotationFixtureSummary {
	for i := range f.Summaries {
		if f.Summaries[i].Name == name {
			return &f.Summaries[i]
		}
	}
	return nil
}
