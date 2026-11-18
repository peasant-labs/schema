package schema

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/freeze_annotation_validator.yaml
var freezeAnnotationValidatorCases []byte

//go:embed testdata/freeze_annotation_validator_manifest.yaml
var freezeAnnotationValidatorManifest []byte

type freezeAnnotationInput struct {
	Body    string `yaml:"body"`
	Require string `yaml:"require"`
}

type freezeAnnotationExpected struct {
	FrozenAccepts  bool `yaml:"frozen_accepts"`
	ChangedAccepts bool `yaml:"changed_accepts"`
}

func TestFrozenAndWorkingAnnotationValidators(t *testing.T) {
	corpus, err := testcase.LoadCorpus[freezeAnnotationInput, freezeAnnotationExpected](freezeAnnotationValidatorCases)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	var names []string
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "annotation freeze validator", freezeAnnotationValidatorManifest, names)
	working, err := workingSpec("annotation-push-request-next.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	frozen := AnnotationPushRequestSchemaJSON()
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var document map[string]any
			if err := json.Unmarshal(working, &document); err != nil {
				t.Fatal(err)
			}
			required, ok := document["required"].([]any) // unstring-allow: go-json-key-assert — fixture mutates arbitrary documents
			if !ok {
				t.Fatal("working request schema has no required list")
			}
			document["required"] = append(required, c.Input.Require)
			changed, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			internal, err := newAnnotationPushRequestSchema(changed)
			if err != nil {
				t.Fatal(err)
			}
			var body any
			if err := json.Unmarshal([]byte(c.Input.Body), &body); err != nil {
				t.Fatal(err)
			}
			if got := internal.Validate(body) == nil; got != c.Expected.ChangedAccepts {
				t.Fatalf("working validator accepts = %v, want %v", got, c.Expected.ChangedAccepts)
			}
			if got := ValidateAnnotationPushRequest([]byte(c.Input.Body)) == nil; got != c.Expected.FrozenAccepts {
				t.Fatalf("exported validator accepts = %v, want %v", got, c.Expected.FrozenAccepts)
			}
			if bytes.Equal(working, frozen) {
				unmodified, err := newAnnotationPushRequestSchema(working)
				if err != nil {
					t.Fatal(err)
				}
				if got := unmodified.Validate(body) == nil; got != c.Expected.FrozenAccepts {
					t.Fatalf("clean working validator accepts = %v, want %v", got, c.Expected.FrozenAccepts)
				}
			}
		})
	}
}
