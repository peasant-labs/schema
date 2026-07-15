package schema

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
)

//go:embed testdata/quality/loader_cases.yaml
var qualityLoaderCasesYAML []byte

type qualityLoaderMutation struct {
	Old    string `yaml:"old"`
	New    string `yaml:"new"`
	Append string `yaml:"append"`
}

func TestQualityLoaderRejectsFixtureDrivenMutations(t *testing.T) {
	matrix, err := testcase.LoadCorpus[qualityLoaderMutation, bool](qualityLoaderCasesYAML)
	if err != nil {
		t.Fatalf("load quality mutation corpus: %v", err)
	}
	if err := matrix.Validate(); err != nil {
		t.Fatalf("validate quality mutation corpus: %v", err)
	}
	if len(matrix.Cases) != 7 {
		t.Fatalf("quality mutation corpus has %d rows, want 7", len(matrix.Cases))
	}
	for _, tc := range matrix.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			source := string(QualitySessionsYAML)
			if tc.Input.Old != "" {
				if !strings.Contains(source, tc.Input.Old) {
					t.Fatalf("mutation source does not contain %q", tc.Input.Old)
				}
				source = strings.Replace(source, tc.Input.Old, tc.Input.New, 1)
			}
			source += tc.Input.Append
			_, err := loadQualityFixtures([]byte(source))
			if got := err == nil; got != tc.Expected {
				t.Fatalf("accepted=%v, want %v (error: %v)", got, tc.Expected, err)
			}
		})
	}
}
