package schema_test

import (
	_ "embed"
	"errors"
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
	var fx actionableValidationErrorPolicyFixture
	decoder := yaml.NewDecoder(strings.NewReader(string(actionableValidationErrorPolicyYAML)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fx); err != nil {
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
			missing := testutil.ActionableValidationErrorViolations(mutatedErr, parts.Fragments()...)
			if len(missing) == 0 {
				t.Fatalf("mutated actionable validation error still satisfied the guard after removing %q", mutation.Dimension)
			}
		})
	}
}
