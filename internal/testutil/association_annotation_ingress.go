package testutil

import (
	"bytes"
	"fmt"
	"io"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

// AssociationAnnotationIngressFixtures is the typed form of the shared YAML
// corpus for association publication and annotation ingress validation.
type AssociationAnnotationIngressFixtures struct {
	Cases                   []testcase.Case[AssociationAnnotationIngressInput, AssociationAnnotationIngressExpected] `yaml:"cases"`
	AnnotationRequestShapes testcase.Corpus[AnnotationRequestShape, bool]                                            `yaml:"annotation_request_shapes"`
	StrictDecoding          testcase.Corpus[string, bool]                                                            `yaml:"strict_decoding"`
}

// CaseCorpus returns the typed main validation arm in the same canonical form
// as the nested strict-decoding arm. The YAML keeps its main cases at the
// top-level so TypeScript can consume the one cross-language corpus directly.
func (f *AssociationAnnotationIngressFixtures) CaseCorpus() testcase.Corpus[AssociationAnnotationIngressInput, AssociationAnnotationIngressExpected] {
	return testcase.Corpus[AssociationAnnotationIngressInput, AssociationAnnotationIngressExpected]{Cases: f.Cases}
}

// AssociationAnnotationIngressInput is one producer request fragment and one
// annotation item validated against it. Its field strings deliberately model
// raw YAML; callers convert them to production wire types before validation.
type AssociationAnnotationIngressInput struct {
	Associations   []PublishedAssociationFixture               `yaml:"associations"`
	Annotation     AssociationAnnotationIngressAnnotation      `yaml:"annotation"`
	HashComparison *AssociationAnnotationIngressHashComparison `yaml:"hashComparison,omitempty"`
}

// AssociationAnnotationIngressExpected records the independent request
// boundary verdicts that one shared ingress row must exercise.
type AssociationAnnotationIngressExpected struct {
	PublishRequestValid            bool  `yaml:"publishRequestValid"`
	AnnotationRequestValid         bool  `yaml:"annotationRequestValid"`
	AnnotationOperationSchemaValid *bool `yaml:"annotationOperationSchemaValid,omitempty"`
	AnnotationEntryTargetValid     *bool `yaml:"annotationEntryTargetValid,omitempty"`
}

// OperationSchemaValid returns the expected result from the generated Village
// operation schema. Most rows match the typed request verdict; rows that prove
// a relational rule JSON Schema cannot express declare an explicit override.
func (e AssociationAnnotationIngressExpected) OperationSchemaValid() bool {
	if e.AnnotationOperationSchemaValid == nil {
		return e.AnnotationRequestValid
	}
	return *e.AnnotationOperationSchemaValid
}

// AnnotationRequestShape is the nullability arm of the shared ingress corpus.
// A nil slice represents YAML null; an allocated empty slice represents [].
type AnnotationRequestShape struct {
	Annotations []any `yaml:"annotations"`
}

// PublishedAssociationFixture preserves the fixture's exact wire names before
// the root contract test converts it to schema.PublishedAssociation.
type PublishedAssociationFixture struct {
	ID                 string `yaml:"id"`
	ObservedCommitHash string `yaml:"observedCommitHash"`
}

// AssociationAnnotationIngressAnnotation is the fixture-shaped form of an
// AnnotationPushItem. It keeps YAML field names explicit without importing the
// root contract package into this shared test helper.
type AssociationAnnotationIngressAnnotation struct {
	TargetKind          string                        `yaml:"targetKind"`
	TargetAssociationID *string                       `yaml:"targetAssociationId,omitempty"`
	SessionID           *string                       `yaml:"sessionId,omitempty"`
	EntryTarget         *AnnotationEntryTargetFixture `yaml:"entryTarget,omitempty"`
	AnnotationID        *string                       `yaml:"annotationId,omitempty"`
	ProjectHash         *string                       `yaml:"projectHash,omitempty"`

	explicitNullTargetFields map[string]struct{}
}

// AnnotationEntryTargetFixture is the fixture-shaped entry target before tests
// convert it to the public AnnotationEntryTarget wire type.
type AnnotationEntryTargetFixture struct {
	SessionID  string `yaml:"sessionId"`
	EntryIndex int    `yaml:"entryIndex"`
	EndIndex   int    `yaml:"endIndex"`
}

type associationAnnotationIngressAnnotationWire struct {
	TargetKind          string                        `yaml:"targetKind"`
	TargetAssociationID *string                       `yaml:"targetAssociationId,omitempty"`
	SessionID           *string                       `yaml:"sessionId,omitempty"`
	EntryTarget         *AnnotationEntryTargetFixture `yaml:"entryTarget,omitempty"`
	AnnotationID        *string                       `yaml:"annotationId,omitempty"`
	ProjectHash         *string                       `yaml:"projectHash,omitempty"`
}

// UnmarshalYAML preserves whether a nullable target was explicitly supplied as
// null. The typed Go validator receives the same nil pointer either way, while
// the JSON-backed OpenAPI test must retain explicit null to verify parity with
// the served schema.
func (a *AssociationAnnotationIngressAnnotation) UnmarshalYAML(node *yaml.Node) error {
	if err := requireExactFixtureMapping(node, "annotation", []string{"targetKind", "targetAssociationId", "sessionId", "entryTarget", "annotationId", "projectHash"}); err != nil {
		return err
	}
	for index := 0; index < len(node.Content); index += 2 {
		if node.Content[index].Value != "entryTarget" || node.Content[index+1].Tag == "!!null" {
			continue
		}
		if err := requireExactFixtureMapping(node.Content[index+1], "annotation.entryTarget", []string{"sessionId", "entryIndex", "endIndex"}); err != nil {
			return err
		}
	}

	var wire associationAnnotationIngressAnnotationWire
	if err := node.Decode(&wire); err != nil {
		return fmt.Errorf("decode annotation: %w", err)
	}
	*a = AssociationAnnotationIngressAnnotation{
		TargetKind:          wire.TargetKind,
		TargetAssociationID: wire.TargetAssociationID,
		SessionID:           wire.SessionID,
		EntryTarget:         wire.EntryTarget,
		AnnotationID:        wire.AnnotationID,
		ProjectHash:         wire.ProjectHash,
	}
	for index := 0; index < len(node.Content); index += 2 {
		name, value := node.Content[index].Value, node.Content[index+1]
		if value.Tag != "!!null" || !isAnnotationTargetField(name) {
			continue
		}
		if a.explicitNullTargetFields == nil {
			a.explicitNullTargetFields = make(map[string]struct{})
		}
		a.explicitNullTargetFields[name] = struct{}{}
	}
	return nil
}

// HasExplicitNullTargetField reports whether the fixture explicitly contained
// a JSON-null value for the named target property.
func (a AssociationAnnotationIngressAnnotation) HasExplicitNullTargetField(name string) bool {
	_, exists := a.explicitNullTargetFields[name]
	return exists
}

// AnnotationPushItemTargetJSON returns the fixture's exact target fields for
// a real AnnotationPushItem JSON body, retaining explicit null target values.
func AnnotationPushItemTargetJSON(annotation AssociationAnnotationIngressAnnotation) map[string]any {
	item := map[string]any{"targetKind": annotation.TargetKind}
	addNullableStringTarget(item, "targetAssociationId", annotation.TargetAssociationID, annotation.HasExplicitNullTargetField("targetAssociationId"))
	addNullableStringTarget(item, "sessionId", annotation.SessionID, annotation.HasExplicitNullTargetField("sessionId"))
	if annotation.EntryTarget != nil {
		item["entryTarget"] = map[string]any{
			"sessionId":  annotation.EntryTarget.SessionID,
			"entryIndex": annotation.EntryTarget.EntryIndex,
			"endIndex":   annotation.EntryTarget.EndIndex,
		}
	} else if annotation.HasExplicitNullTargetField("entryTarget") {
		item["entryTarget"] = nil
	}
	addNullableStringTarget(item, "annotationId", annotation.AnnotationID, annotation.HasExplicitNullTargetField("annotationId"))
	addNullableStringTarget(item, "projectHash", annotation.ProjectHash, annotation.HasExplicitNullTargetField("projectHash"))
	return item
}

func addNullableStringTarget(item map[string]any, name string, value *string, explicitNull bool) {
	if value != nil {
		item[name] = *value
	} else if explicitNull {
		item[name] = nil
	}
}

func isAnnotationTargetField(name string) bool {
	switch name {
	case "targetAssociationId", "sessionId", "entryTarget", "annotationId", "projectHash":
		return true
	}
	return false
}

func requireExactFixtureMapping(node *yaml.Node, location string, allowed []string) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("%s must be a mapping", location)
	}
	allowedFields := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedFields[name] = struct{}{}
	}
	seen := make(map[string]struct{}, len(node.Content)/2)
	for index := 0; index < len(node.Content); index += 2 {
		name := node.Content[index].Value
		if _, exists := allowedFields[name]; !exists {
			return fmt.Errorf("%s has unknown field %q", location, name)
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("%s repeats field %q", location, name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

// AssociationAnnotationIngressHashComparison declares a target-ID mutation
// whose production content hash must have the expected equality relationship.
type AssociationAnnotationIngressHashComparison struct {
	AlternateTargetAssociationID string `yaml:"alternateTargetAssociationId"`
	Distinct                     bool   `yaml:"distinct"`
}

// DecodeAssociationAnnotationIngressFixtures validates a raw corpus source.
// It lets strict-decoding cases exercise the same loader that reads the
// committed corpus.
func DecodeAssociationAnnotationIngressFixtures(data []byte) (*AssociationAnnotationIngressFixtures, error) {
	var fixtures AssociationAnnotationIngressFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		return nil, fmt.Errorf("decode association annotation ingress fixtures: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("decode trailing association annotation ingress fixture document: %w", err)
		}
		return nil, fmt.Errorf("decode association annotation ingress fixtures: multiple YAML documents are not allowed")
	}
	if err := fixtures.CaseCorpus().Validate(); err != nil {
		return nil, fmt.Errorf("validate association annotation ingress cases: %w", err)
	}
	if err := fixtures.AnnotationRequestShapes.Validate(); err != nil {
		return nil, fmt.Errorf("validate annotation request shape cases: %w", err)
	}
	if err := fixtures.StrictDecoding.Validate(); err != nil {
		return nil, fmt.Errorf("validate association annotation ingress strict decoding cases: %w", err)
	}
	return &fixtures, nil
}
