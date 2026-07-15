package schema

import (
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
)

func TestTimelineFixtureValidationRejectsVacuousMutation(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	fixtures.Cases[0].Mutation.Description = " "
	if err := fixtures.Validate(); err == nil || !strings.Contains(err.Error(), "mutation description is empty") {
		t.Fatalf("vacuous mutation error = %v, want mutation-description rejection", err)
	}
}

func TestTimelineFixtureValidationRejectsContradictoryOutcome(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	fixtures.Cases[0].Classification = testcase.MustFail
	manifest, err := LoadTimelineFixtureManifest()
	if err != nil {
		t.Fatal(err)
	}
	if err := validateTimelineFixtures(fixtures, manifest); err == nil || !strings.Contains(err.Error(), "want manifest identity") {
		t.Fatalf("contradictory outcome error = %v, want missing error expectation rejection", err)
	}
}

func TestTimelineFixtureValidationRejectsMissingCanonicalCase(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	fixtures.Cases = fixtures.Cases[:len(fixtures.Cases)-1]
	manifest, err := LoadTimelineFixtureManifest()
	if err != nil {
		t.Fatal(err)
	}
	if err := validateTimelineFixtures(fixtures, manifest); err == nil || !strings.Contains(err.Error(), "want at least 16") {
		t.Fatalf("missing-case error = %v, want exact-count rejection", err)
	}
}

func TestTimelineFixtureManifestRejectsCountPreservingMutations(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := LoadTimelineFixtureManifest()
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range manifest.Mutations.Cases {
		t.Run(mutation.Name, func(t *testing.T) {
			mutated := fixtures
			mutated.Cases = append([]testcase.Case[TimelineFixtureInput, TimelineFixtureExpected](nil), fixtures.Cases...)
			target := -1
			for index := range mutated.Cases {
				if mutated.Cases[index].Name == mutation.Input.Target {
					target = index
					break
				}
			}
			if target < 0 {
				t.Fatalf("mutation target %q is absent from canonical corpus", mutation.Input.Target)
			}
			switch mutation.Input.Kind {
			case "rename":
				mutated.Cases[target].Name = mutation.Input.ReplacementName
			case "replace":
				mutated.Cases[target].Name = mutation.Input.ReplacementName
				mutated.Cases[target].Classification = mutation.Input.ReplacementClassification
			default:
				t.Fatalf("unknown manifest mutation kind %q", mutation.Input.Kind)
			}
			accepted := validateTimelineFixtures(mutated, manifest) == nil
			if accepted != mutation.Expected {
				t.Fatalf("accepted=%v, want %v", accepted, mutation.Expected)
			}
		})
	}
}
