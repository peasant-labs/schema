package schema

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

const timelineFixtureCaseCount = 16

// TimelineFixtureInput is one normalized session and commit relationship.
type TimelineFixtureInput struct {
	Sessions []TimelineSessionRef `json:"sessions" yaml:"sessions"`
	Commits  []CommitRef          `json:"commits" yaml:"commits"`
}

// TimelineFixtureExpected records the validation error for a rejected input.
// A must-pass case leaves ErrorContains empty.
type TimelineFixtureExpected struct {
	ErrorContains string `json:"errorContains,omitempty" yaml:"error_contains"`
}

// TimelineFixtureCorpus is the project Git timeline validation corpus.
type TimelineFixtureCorpus = testcase.Corpus[TimelineFixtureInput, TimelineFixtureExpected]

// TimelineFixtureManifestEntry pins one canonical behavioral family to its
// exact case name and expected classification.
type TimelineFixtureManifestEntry struct {
	Family         string                  `json:"family" yaml:"family"`
	Name           string                  `json:"name" yaml:"name"`
	Classification testcase.Classification `json:"classification" yaml:"classification"`
}

// TimelineManifestMutationInput describes a count-preserving manifest mutation.
type TimelineManifestMutationInput struct {
	Kind                      string                  `json:"kind" yaml:"kind"`
	Target                    string                  `json:"target" yaml:"target"`
	ReplacementName           string                  `json:"replacementName" yaml:"replacement_name"`
	ReplacementClassification testcase.Classification `json:"replacementClassification" yaml:"replacement_classification"`
}

// TimelineManifestMutationCorpus proves that exact family identity survives
// count-preserving renames and replacements.
type TimelineManifestMutationCorpus = testcase.Corpus[TimelineManifestMutationInput, bool]

// TimelineFixtureManifest is the independent identity oracle for the corpus.
type TimelineFixtureManifest struct {
	Cases     []TimelineFixtureManifestEntry `json:"cases" yaml:"cases"`
	Mutations TimelineManifestMutationCorpus `json:"mutations" yaml:"mutations"`
}

// LoadTimelineFixtures parses and validates the shared timeline corpus.
func LoadTimelineFixtures() (TimelineFixtureCorpus, error) {
	fixtures, err := testcase.LoadCorpus[TimelineFixtureInput, TimelineFixtureExpected](TimelineYAML)
	if err != nil {
		return TimelineFixtureCorpus{}, fmt.Errorf("load timeline fixtures: %w", err)
	}
	manifest, err := LoadTimelineFixtureManifest()
	if err != nil {
		return TimelineFixtureCorpus{}, err
	}
	if err := validateTimelineFixtures(fixtures, manifest); err != nil {
		return TimelineFixtureCorpus{}, err
	}
	return fixtures, nil
}

// LoadTimelineFixtureManifest parses and validates the exact timeline family manifest.
func LoadTimelineFixtureManifest() (TimelineFixtureManifest, error) {
	var manifest TimelineFixtureManifest
	decoder := yaml.NewDecoder(bytes.NewReader(TimelineManifestYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return TimelineFixtureManifest{}, fmt.Errorf("load timeline fixture manifest: decode document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return TimelineFixtureManifest{}, fmt.Errorf("load timeline fixture manifest: decode trailing document: %w", err)
		}
		return TimelineFixtureManifest{}, fmt.Errorf("load timeline fixture manifest: multiple YAML documents are not allowed")
	}
	if err := validateTimelineFixtureManifest(manifest); err != nil {
		return TimelineFixtureManifest{}, err
	}
	return manifest, nil
}

func validateTimelineFixtureManifest(manifest TimelineFixtureManifest) error {
	if len(manifest.Cases) != timelineFixtureCaseCount {
		return fmt.Errorf("load timeline fixture manifest: has %d families, want exactly %d", len(manifest.Cases), timelineFixtureCaseCount)
	}
	families := make(map[string]struct{}, len(manifest.Cases))
	names := make(map[string]struct{}, len(manifest.Cases))
	for index, entry := range manifest.Cases {
		if strings.TrimSpace(entry.Family) == "" || strings.TrimSpace(entry.Name) == "" {
			return fmt.Errorf("load timeline fixture manifest: entry %d has an empty family or name; identify every canonical behavior explicitly", index)
		}
		if !entry.Classification.IsValid() {
			return fmt.Errorf("load timeline fixture manifest: family %q has invalid classification %q; use a value from testcase.AllClassifications", entry.Family, entry.Classification)
		}
		if _, duplicate := families[entry.Family]; duplicate {
			return fmt.Errorf("load timeline fixture manifest: duplicate family %q; each behavior must have one identity", entry.Family)
		}
		if _, duplicate := names[entry.Name]; duplicate {
			return fmt.Errorf("load timeline fixture manifest: duplicate case name %q; each behavior must have one case", entry.Name)
		}
		families[entry.Family] = struct{}{}
		names[entry.Name] = struct{}{}
	}
	if err := manifest.Mutations.Validate(); err != nil {
		return fmt.Errorf("load timeline fixture manifest: validate mutation corpus: %w", err)
	}
	if len(manifest.Mutations.Cases) != 2 {
		return fmt.Errorf("load timeline fixture manifest: mutation corpus has %d cases, want exactly 2 count-preserving proofs", len(manifest.Mutations.Cases))
	}
	return nil
}

func validateTimelineFixtures(fixtures TimelineFixtureCorpus, manifest TimelineFixtureManifest) error {
	if err := fixtures.CheckMin(timelineFixtureCaseCount); err != nil {
		return fmt.Errorf("load timeline fixtures: %w", err)
	}
	if len(fixtures.Cases) != timelineFixtureCaseCount {
		return fmt.Errorf("load timeline fixtures: corpus has %d cases, want exactly %d canonical relationship cases", len(fixtures.Cases), timelineFixtureCaseCount)
	}
	passCount, failCount := 0, 0
	for index, fixture := range fixtures.Cases {
		identity := manifest.Cases[index]
		if fixture.Name != identity.Name || fixture.Classification != identity.Classification {
			return fmt.Errorf("load timeline fixtures: case %d identity=(%q,%q), want manifest identity=(%q,%q) for family %q; restore the canonical case rather than substituting a filler", index, fixture.Name, fixture.Classification, identity.Name, identity.Classification, identity.Family)
		}
		if len(fixture.Input.Sessions) == 0 && len(fixture.Input.Commits) == 0 {
			return fmt.Errorf("load timeline fixtures: case %q has no sessions or commits; add relationship data so the case cannot pass vacuously", fixture.Name)
		}
		switch fixture.Classification {
		case testcase.MustPass:
			passCount++
			if strings.TrimSpace(fixture.Expected.ErrorContains) != "" {
				return fmt.Errorf("load timeline fixtures: must-pass case %q unexpectedly declares error_contains; remove the contradictory error expectation", fixture.Name)
			}
		case testcase.MustFail:
			failCount++
			if strings.TrimSpace(fixture.Expected.ErrorContains) == "" {
				return fmt.Errorf("load timeline fixtures: must-fail case %q has no error_contains; name the validation failure the mutation must trigger", fixture.Name)
			}
		}
	}
	if passCount != 5 || failCount != 11 {
		return fmt.Errorf("load timeline fixtures: canonical outcome coverage changed; got %d must-pass and %d must-fail cases, want 5 and 11", passCount, failCount)
	}
	return nil
}
