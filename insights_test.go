package schema_test

import (
	_ "embed"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/local-api/insights.yaml
var insightsCasesYAML []byte

// insightFixtures is the segmented multi-axis fixture for SessionInsight
// (the insight envelope contract): a typed struct of named testcase.Corpus arms sharing
// one (schema.SessionInsight, bool) shape but a distinct behavioral role
// each, per the module's segmented-fixture convention (TESTING.md).
type insightFixtures struct {
	// Mechanical covers every InsightKind under the current mechanical
	// producers (unusualSignals, frictionClusters,
	// retry-loop detection).
	Mechanical testcase.Corpus[schema.SessionInsight, bool] `yaml:"mechanical"`
	// Mined previews the reserved mined shape with no evidence-count requirement
	// (that requirement is scoped to provenance=mechanical only).
	Mined testcase.Corpus[schema.SessionInsight, bool] `yaml:"mined"`
	// ClassificationMustBeNil enforces the current contract invariant:
	// every case here is must-fail, and each mutation is "populate
	// Classification", proving the rule is not vacuous.
	ClassificationMustBeNil testcase.Corpus[schema.SessionInsight, bool] `yaml:"classification_must_be_nil"`
	// Rejections covers every other closed-set/nil-array failure mode.
	Rejections testcase.Corpus[schema.SessionInsight, bool] `yaml:"rejections"`
}

func loadInsightFixtures(t *testing.T) insightFixtures {
	t.Helper()
	var fx insightFixtures
	if err := yaml.Unmarshal(insightsCasesYAML, &fx); err != nil {
		t.Fatalf("load insights fixtures (testdata/local-api/insights.yaml): %v", err)
	}
	assert.RequireMin(t, fx.Mechanical, len(schema.AllInsightKinds))
	assert.RequireValid(t, fx.Mechanical)
	assert.RequireMin(t, fx.Mined, 2)
	assert.RequireValid(t, fx.Mined)
	assert.RequireMin(t, fx.ClassificationMustBeNil, 2)
	assert.RequireValid(t, fx.ClassificationMustBeNil)
	assert.RequireMin(t, fx.Rejections, 5)
	assert.RequireValid(t, fx.Rejections)
	return fx
}

func runInsightArm(t *testing.T, arm string, corpus testcase.Corpus[schema.SessionInsight, bool]) {
	t.Helper()
	t.Run(arm, func(t *testing.T) {
		for _, c := range corpus.Cases {
			t.Run(c.Name, func(t *testing.T) {
				err := c.Input.Validate()
				valid := err == nil
				if valid != c.Expected {
					t.Fatalf("Validate() error=%v (valid=%v), want valid=%v", err, valid, c.Expected)
				}
				if c.Classification == testcase.MustFail {
					requireActionableValidationError(t, err)
				}
			})
		}
	})
}

// TestSessionInsight_FixtureContract drives every arm of the segmented
// insight corpus against the real SessionInsight.Validate.
func TestSessionInsight_FixtureContract(t *testing.T) {
	fx := loadInsightFixtures(t)
	runInsightArm(t, "mechanical", fx.Mechanical)
	runInsightArm(t, "mined", fx.Mined)
	runInsightArm(t, "classification_must_be_nil", fx.ClassificationMustBeNil)
	runInsightArm(t, "rejections", fx.Rejections)
}

// TestSessionInsight_MechanicalArmCoversEveryInsightKind guards the
// mechanical arm's exhaustion claim: every InsightKind member has at least
// one must-pass mechanical case, so a fifth kind added without a fixture
// update reddens here rather than silently under-covering the wire.
func TestSessionInsight_MechanicalArmCoversEveryInsightKind(t *testing.T) {
	fx := loadInsightFixtures(t)
	covered := map[schema.InsightKind]bool{}
	for _, c := range fx.Mechanical.Cases {
		if c.Classification == testcase.MustPass {
			covered[c.Input.Kind] = true
		}
	}
	for _, kind := range schema.AllInsightKinds {
		if !covered[kind] {
			t.Errorf("InsightKind member %q has no must-pass mechanical fixture case", kind)
		}
	}
}

// TestSessionInsight_ClassificationMustBeNilArmIsMutationProvable proves the
// classification_must_be_nil arm is not vacuous: every case there must-fail
// specifically because Classification is non-nil (not for some unrelated
// reason), and clearing Classification on each case's input must make it
// pass. This is the fixture-level mutation proof the current contract
// invariant needs.
func TestSessionInsight_ClassificationMustBeNilArmIsMutationProvable(t *testing.T) {
	fx := loadInsightFixtures(t)
	for _, c := range fx.ClassificationMustBeNil.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Input.Classification == nil {
				t.Fatalf("case %q has a nil Classification; this arm exists to prove the must-be-nil rule and every case must set one", c.Name)
			}
			if err := c.Input.Validate(); err == nil || !strings.Contains(err.Error(), "classification is non-nil") {
				t.Fatalf("Validate() error=%v, want a classification-is-non-nil rejection", err)
			}
			cleared := c.Input
			cleared.Classification = nil
			if err := cleared.Validate(); err != nil {
				t.Fatalf("clearing Classification still failed validation: %v; the case must fail for exactly the must-be-nil reason", err)
			}
		})
	}
}
