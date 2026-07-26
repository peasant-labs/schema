package schema

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

type timelineFixtureOracleEntry struct {
	Family                  string                           `yaml:"family"`
	Name                    string                           `yaml:"name"`
	Classification          testcase.Classification          `yaml:"classification"`
	RequiredRewrittenCommit *timelineRequiredRewrittenCommit `yaml:"required_rewritten_commit"`
	RequiredRepairKind      timelineFixtureRepairKind        `yaml:"required_repair_kind"`
	LedgerTargetSessionID   SessionID                        `yaml:"ledger_target_session_id"`
}

type timelineRequiredRewrittenCommit struct {
	GhostHash string `yaml:"ghost_hash"`
}

type timelineOracleMutationInput struct {
	Kind                      string                  `yaml:"kind"`
	Target                    string                  `yaml:"target"`
	ReplacementName           string                  `yaml:"replacement_name"`
	ReplacementClassification testcase.Classification `yaml:"replacement_classification"`
}

type timelineFixtureOracle struct {
	Cases     []timelineFixtureOracleEntry                       `yaml:"cases"`
	Mutations testcase.Corpus[timelineOracleMutationInput, bool] `yaml:"mutations"`
}

func TestTimelineFixtureValidationRejectsVacuousMutation(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	fixtures.Cases[0].Mutation.Description = " "
	if err := validateTimelineFixtures(fixtures); err == nil || !strings.Contains(err.Error(), "mutation description is empty") {
		t.Fatalf("vacuous mutation error = %v, want mutation-description rejection", err)
	}
}

func TestTimelineFixtureValidationRejectsContradictoryOutcome(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	fixtures.Cases[0].Classification = testcase.MustFail
	if err := validateTimelineFixtures(fixtures); err == nil || !strings.Contains(err.Error(), "no error_contains") {
		t.Fatalf("contradictory outcome error = %v, want missing error expectation rejection", err)
	}
}

func TestTimelineFixtureValidationRejectsMissingCanonicalCase(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatalf("LoadTimelineFixtures: %v", err)
	}
	fixtures.Cases = fixtures.Cases[:len(fixtures.Cases)-1]
	if err := validateTimelineFixtures(fixtures); err == nil || !strings.Contains(err.Error(), "want at least 23") {
		t.Fatalf("missing-case error = %v, want exact-count rejection", err)
	}
}

func TestTimelineFixtureOracleRejectsCountPreservingMutations(t *testing.T) {
	fixtures, err := LoadTimelineFixtures()
	if err != nil {
		t.Fatal(err)
	}
	oracle := loadTimelineFixtureOracle(t)
	if err := validateTimelineFixtureOracle(fixtures, oracle); err != nil {
		t.Fatalf("canonical timeline fixture differs from its test-only oracle: %v", err)
	}
	for _, mutation := range oracle.Mutations.Cases {
		t.Run(mutation.Name, func(t *testing.T) {
			mutated := fixtures
			mutated.Cases = append([]TimelineFixtureCase(nil), fixtures.Cases...)
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
				t.Fatalf("unknown oracle mutation kind %q", mutation.Input.Kind)
			}
			accepted := validateTimelineFixtureOracle(mutated, oracle) == nil
			if accepted != mutation.Expected {
				t.Fatalf("accepted=%v, want %v", accepted, mutation.Expected)
			}
		})
	}
}

func loadTimelineFixtureOracle(t *testing.T) timelineFixtureOracle {
	t.Helper()
	data, err := os.ReadFile("testdata/local-api/timeline_manifest.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var oracle timelineFixtureOracle
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&oracle); err != nil {
		t.Fatal(err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			t.Fatalf("decode trailing timeline oracle document: %v", err)
		}
		t.Fatal("timeline oracle contains multiple YAML documents")
	}
	if len(oracle.Cases) != timelineFixtureCaseCount {
		t.Fatalf("timeline oracle has %d families, want exactly %d", len(oracle.Cases), timelineFixtureCaseCount)
	}
	if err := oracle.Mutations.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(oracle.Mutations.Cases) != 2 {
		t.Fatalf("timeline oracle has %d mutations, want exactly 2", len(oracle.Mutations.Cases))
	}
	return oracle
}

func validateTimelineFixtureOracle(fixtures TimelineFixtureCorpus, oracle timelineFixtureOracle) error {
	if len(fixtures.Cases) != len(oracle.Cases) {
		return fmt.Errorf("timeline fixture has %d rows, test-only oracle has %d", len(fixtures.Cases), len(oracle.Cases))
	}
	for index, fixture := range fixtures.Cases {
		expected := oracle.Cases[index]
		if fixture.Family != expected.Family || fixture.Name != expected.Name || fixture.Classification != expected.Classification {
			return fmt.Errorf("timeline row %d identity=(%q,%q,%q), want test-only oracle identity=(%q,%q,%q)", index, fixture.Family, fixture.Name, fixture.Classification, expected.Family, expected.Name, expected.Classification)
		}
		if err := validateTimelineFixtureOracleSemantics(fixture, expected); err != nil {
			return fmt.Errorf("timeline row %d %q semantic oracle: %w", index, fixture.Name, err)
		}
	}
	return nil
}

func validateTimelineFixtureOracleSemantics(fixture TimelineFixtureCase, expected timelineFixtureOracleEntry) error {
	switch expected.Name {
	case "rewrite_ledger_references_bound_session":
		if expected.RequiredRewrittenCommit == nil {
			return fmt.Errorf("required_rewritten_commit is absent; declare the rewrite-ledger identity this row protects")
		}
	case "rewrite_ledger_reference_requires_binding_truth":
		if expected.RequiredRepairKind == "" {
			return fmt.Errorf("required_repair_kind is absent; declare the repair this row protects")
		}
		if expected.LedgerTargetSessionID == "" {
			return fmt.Errorf("ledger_target_session_id is absent; declare the ledger-referenced repair target")
		}
	}
	if expected.RequiredRewrittenCommit != nil {
		if strings.TrimSpace(expected.RequiredRewrittenCommit.GhostHash) == "" {
			return fmt.Errorf("required_rewritten_commit.ghost_hash is empty")
		}
		if !fixture.Input.hasRewrittenCommit(expected.RequiredRewrittenCommit.GhostHash) {
			return fmt.Errorf("requires rewritten commit with ghostHash %q, but the fixture input does not contain it", expected.RequiredRewrittenCommit.GhostHash)
		}
	}
	if expected.RequiredRepairKind != "" {
		if fixture.Expected.Repair == nil {
			return fmt.Errorf("requires repair kind %q, but the fixture has no expected repair", expected.RequiredRepairKind)
		}
		if fixture.Expected.Repair.Kind != expected.RequiredRepairKind {
			return fmt.Errorf("repair kind=%q, want %q", fixture.Expected.Repair.Kind, expected.RequiredRepairKind)
		}
	}
	if expected.LedgerTargetSessionID != "" {
		if !fixture.Input.rewriteLedgerReferencesSession(expected.LedgerTargetSessionID) {
			return fmt.Errorf("ledger target sessionId %q is not referenced by any rewritten commit", expected.LedgerTargetSessionID)
		}
		if fixture.Expected.Repair == nil {
			return fmt.Errorf("ledger target sessionId %q requires an expected repair", expected.LedgerTargetSessionID)
		}
		if fixture.Expected.Repair.SessionID != expected.LedgerTargetSessionID {
			return fmt.Errorf("repair sessionId=%q, want ledger target sessionId %q", fixture.Expected.Repair.SessionID, expected.LedgerTargetSessionID)
		}
	}
	return nil
}
