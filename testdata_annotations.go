package schema

import "gopkg.in/yaml.v3"

// AnnotationFixtures is the top-level YAML structure for annotation test data.
type AnnotationFixtures struct {
	Summaries   []AnnotationFixtureSummary   `yaml:"annotation_summaries"`
	Validations []SubscriptionValidationCase `yaml:"subscription_validations"`
}

// AnnotationFixtureSummary is a named annotation template in YAML.
type AnnotationFixtureSummary struct {
	Name            string `yaml:"name"`
	ID              string `yaml:"id"`
	TargetKind      string `yaml:"targetKind"`
	TargetSessionID string `yaml:"targetSessionId,omitempty"`
	IsPrimary       bool   `yaml:"isPrimary"`
	AnnotatorKind   string `yaml:"annotatorKind"`
	AnnotatorName   string `yaml:"annotatorName"`
	TypeID          string `yaml:"typeId"`
	TypeName        string `yaml:"typeName"`
	Value           string `yaml:"value"`
	CreatedAt       int64  `yaml:"createdAt"`
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
	return s
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
	if err := yaml.Unmarshal(AnnotationsYAML, &f); err != nil {
		return nil, err
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
