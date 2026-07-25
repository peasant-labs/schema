package schema_test

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/internal/testutil"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/actionable_validation_error_policy.yaml
var actionableValidationErrorPolicyYAML []byte

type actionableValidationErrorPolicyFixture struct {
	Baseline  actionableValidationErrorPolicyBaseline   `yaml:"baseline"`
	Mutations []actionableValidationErrorPolicyMutation `yaml:"mutations"`
}

type actionableValidationErrorPolicyBaseline struct {
	Validator string `yaml:"validator"`
	Input     string `yaml:"input"`
}

type actionableValidationErrorPolicyMutation struct {
	Name      string                                      `yaml:"name"`
	Dimension testutil.ActionableValidationErrorDimension `yaml:"dimension"`
}

func loadActionableValidationErrorPolicy(t *testing.T) actionableValidationErrorPolicyFixture {
	t.Helper()
	fx, err := decodeActionableValidationErrorPolicy(actionableValidationErrorPolicyYAML)
	if err != nil {
		t.Fatalf("load actionable validation error policy fixture: %v", err)
	}
	if len(fx.Mutations) != 6 {
		t.Fatalf("actionable validation error policy has %d mutations, want exactly 6", len(fx.Mutations))
	}
	seen := map[testutil.ActionableValidationErrorDimension]struct{}{}
	for _, mutation := range fx.Mutations {
		if strings.TrimSpace(mutation.Name) == "" {
			t.Fatal("actionable validation error policy contains an empty mutation name")
		}
		if _, exists := seen[mutation.Dimension]; exists {
			t.Fatalf("actionable validation error policy repeats dimension %q", mutation.Dimension)
		}
		seen[mutation.Dimension] = struct{}{}
	}
	return fx
}

func decodeActionableValidationErrorPolicy(data []byte) (actionableValidationErrorPolicyFixture, error) {
	var fx actionableValidationErrorPolicyFixture
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fx); err != nil {
		return actionableValidationErrorPolicyFixture{}, fmt.Errorf("decode actionable validation error policy fixture: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return actionableValidationErrorPolicyFixture{}, fmt.Errorf("decode trailing actionable validation error policy fixture document: %w", err)
		}
		return actionableValidationErrorPolicyFixture{}, fmt.Errorf("decode actionable validation error policy fixture: multiple YAML documents are not allowed")
	}
	return fx, nil
}

func baselineActionableValidationError(t *testing.T, fx actionableValidationErrorPolicyBaseline) error {
	t.Helper()
	switch fx.Validator {
	case "read_state_grade":
		return schema.ReadStateGrade(fx.Input).Validate()
	default:
		t.Fatalf("unknown actionable validation error baseline validator %q", fx.Validator)
		return nil
	}
}

func requireActionableValidationErrorFails(t *testing.T, err error, wantContains ...string) bool {
	t.Helper()
	return testing.RunTests(func(_, _ string) (bool, error) { return true, nil }, []testing.InternalTest{{
		Name: "RequireActionableValidationError",
		F: func(t *testing.T) {
			testutil.RequireActionableValidationError(t, err, wantContains...)
		},
	}})
}

func TestActionableValidationErrorPolicyMutationProof(t *testing.T) {
	fx := loadActionableValidationErrorPolicy(t)
	baselineErr := baselineActionableValidationError(t, fx.Baseline)
	parts, err := testutil.ParseActionableValidationError(baselineErr.Error())
	if err != nil {
		t.Fatalf("parse representative actionable validation error: %v", err)
	}
	testutil.RequireActionableValidationError(t, baselineErr, parts.Fragments()...)

	for _, mutation := range fx.Mutations {
		t.Run(mutation.Name, func(t *testing.T) {
			mutatedMsg, err := testutil.StripActionableValidationDimension(baselineErr.Error(), mutation.Dimension)
			if err != nil {
				t.Fatalf("strip %q dimension: %v", mutation.Dimension, err)
			}
			mutatedErr := errors.New(mutatedMsg)
			if mutation.Dimension == testutil.ActionableValidationErrorRemediation {
				if ok := requireActionableValidationErrorFails(t, mutatedErr, parts.Fragments()...); ok {
					t.Fatal("remediation-stripped actionable validation error unexpectedly satisfied RequireActionableValidationError")
				}
				return
			}
			missing := testutil.ActionableValidationErrorViolations(mutatedErr, parts.Fragments()...)
			if len(missing) == 0 {
				t.Fatalf("mutated actionable validation error still satisfied the guard after removing %q", mutation.Dimension)
			}
		})
	}
}

func TestActionableValidationErrorPolicyLoaderRejectsUnknownField(t *testing.T) {
	_, err := decodeActionableValidationErrorPolicy(append(append([]byte{}, actionableValidationErrorPolicyYAML...), []byte("\nunexpected: true\n")...))
	if err == nil {
		t.Fatal("decodeActionableValidationErrorPolicy accepted an unknown top-level field")
	}
}

func TestActionableValidationErrorPolicyLoaderRejectsTrailingDocument(t *testing.T) {
	_, err := decodeActionableValidationErrorPolicy(append(append([]byte{}, actionableValidationErrorPolicyYAML...), []byte("\n---\nextra: true\n")...))
	if err == nil {
		t.Fatal("decodeActionableValidationErrorPolicy accepted a trailing YAML document")
	}
}
