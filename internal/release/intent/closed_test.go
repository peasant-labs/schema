package intent_test

// Closed-set membership suite: every closed set in the package is asserted by
// exact membership, never by a count. A dropped member, a stray addition, and
// a reorder all redden the suite.

import (
	"reflect"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
	"gopkg.in/yaml.v3"
)

func TestClosedSets(t *testing.T) {
	if got, want := intent.AllTypes, []intent.Type{
		intent.TypeFeat, intent.TypeFix, intent.TypeRelease,
		intent.TypeRefactor, intent.TypePerf, intent.TypeBuild, intent.TypeRevert,
		intent.TypeDocs, intent.TypeTest, intent.TypeCI, intent.TypeChore, intent.TypeStyle,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllTypes = %v, want %v", got, want)
	}
	if got, want := intent.AllClasses, []intent.Class{
		intent.ClassMustRelease, intent.ClassMayRelease, intent.ClassNeverRelease,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllClasses = %v, want %v", got, want)
	}
	if got, want := intent.AllLevels, []intent.Level{
		intent.LevelNone, intent.LevelPatch, intent.LevelMinor, intent.LevelMajor,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllLevels = %v, want %v", got, want)
	}
	if got, want := intent.AllModes, []intent.Mode{
		intent.ModePlanOnly, intent.ModeTag,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllModes = %v, want %v", got, want)
	}
	if got, want := intent.AllActions, []intent.Action{
		intent.ActionNone, intent.ActionVerify, intent.ActionTag, intent.ActionOpenPR, intent.ActionClosePR,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllActions = %v, want %v", got, want)
	}
	if got, want := intent.AllVerdicts, []intent.Verdict{
		intent.VerdictNone, intent.VerdictTaggable, intent.VerdictRejected,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllVerdicts = %v, want %v", got, want)
	}
	if got, want := intent.AllGates, []intent.Gate{
		intent.GateArm, intent.GateHold,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllGates = %v, want %v", got, want)
	}
	if got, want := intent.AllFreshness, []intent.Freshness{
		intent.FreshnessUnknown, intent.FreshnessFresh, intent.FreshnessStale,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllFreshness = %v, want %v", got, want)
	}
	if got, want := intent.AllPRKinds, []intent.PRKind{
		intent.PRKindOrdinary, intent.PRKindRelease, intent.PRKindVendorHash, intent.PRKindCutRelease,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllPRKinds = %v, want %v", got, want)
	}
	if got, want := intent.AllErrorCodes, []intent.ErrorCode{
		intent.ErrMalformed,
		intent.ErrUnknownType,
		intent.ErrMissingBump,
		intent.ErrBumpOnNonReleasing,
		intent.ErrBangBumpMismatch,
		intent.ErrOldReleaseTitle,
		intent.ErrReleasesDisabled,
		intent.ErrProtectedPath,
		intent.ErrNoFinalTag,
		intent.ErrEpochMissing,
		intent.ErrFloorNotOnChain,
		intent.ErrReleaseContentStale,
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("AllErrorCodes = %v, want %v", got, want)
	}
}

// TestErrorCodeCoverage proves every code the intent package can reject has a
// reject fixture with six non-empty parts. Each corpus suite already asserts
// the six parts at its reject sites through requireSixParts; this test
// re-reads every corpus and proves each code appears in at least one
// must-fail row, independent of test execution order. ReleaseContentStale is
// judged at the command level, where the release PR's content is compared;
// the CLI fixtures own its reject case.
func TestErrorCodeCoverage(t *testing.T) {
	seen := map[string]bool{}
	for _, data := range [][]byte{
		titleCorpusYAML, arithmeticCorpusYAML, semverCorpusYAML,
		classifierCorpusYAML, configCorpusYAML, protectedCorpusYAML,
		planCorpusYAML, generateCorpusYAML,
		queueDecodeCorpusYAML, queueGateCorpusYAML,
		highestTagCorpusYAML, versionFieldCorpusYAML,
	} {
		var doc struct {
			Cases []struct {
				Classification string         `yaml:"classification"`
				Expected       map[string]any `yaml:"expected"`
			} `yaml:"cases"`
		}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			t.Fatalf("scan corpus for codes: %v", err)
		}
		for _, c := range doc.Cases {
			if c.Classification == "must-fail" {
				if code, ok := c.Expected["code"].(string); ok && code != "" {
					seen[code] = true
				}
			}
		}
	}
	commandLevel := map[string]bool{
		string(intent.ErrReleaseContentStale): true,
	}
	// pendingCodes lists codes whose reject fixtures arrive with a later
	// layer: the command level owns the content comparison. Each layer
	// deletes its codes here as it lands them.
	pendingCodes := map[string]bool{}
	for _, code := range intent.AllErrorCodes {
		if commandLevel[string(code)] || pendingCodes[string(code)] {
			continue
		}
		if !seen[string(code)] {
			t.Errorf("error code %q has no reject fixture in the intent corpora", code)
		}
	}
}
