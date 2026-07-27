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
	PublishRequestValid    bool `yaml:"publishRequestValid"`
	AnnotationRequestValid bool `yaml:"annotationRequestValid"`
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
	TargetKind          string  `yaml:"targetKind"`
	TargetAssociationID *string `yaml:"targetAssociationId,omitempty"`
	SessionID           *string `yaml:"sessionId,omitempty"`
	AnnotationID        *string `yaml:"annotationId,omitempty"`
	ProjectHash         *string `yaml:"projectHash,omitempty"`
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
