package schema_test

import (
	"testing"

	"github.com/peasant-labs/schema"
)

func TestLoadQualityFixtures(t *testing.T) {
	fixtures, err := schema.LoadQualityFixtures()
	if err != nil {
		t.Fatalf("LoadQualityFixtures: %v", err)
	}

	row, ok := fixtures.SessionByName(schema.QualityFixtureResolvedTypical)
	if !ok {
		t.Fatalf("SessionByName(%q) not found", schema.QualityFixtureResolvedTypical)
	}

	got := row.ToQualitySession()
	if got.ID != "sess-000" {
		t.Fatalf("ID = %q, want sess-000", got.ID)
	}
	if got.Project != "fortuna" {
		t.Fatalf("Project = %q, want fortuna", got.Project)
	}
	if got.RetryTokensWasted != 0 {
		t.Fatalf("RetryTokensWasted = %d, want 0", got.RetryTokensWasted)
	}
	if got.WithinSessionReverts != 1 {
		t.Fatalf("WithinSessionReverts = %d, want 1", got.WithinSessionReverts)
	}
}

func TestQualityFixtures_SetByName(t *testing.T) {
	fixtures, err := schema.LoadQualityFixtures()
	if err != nil {
		t.Fatalf("LoadQualityFixtures: %v", err)
	}

	sessions, err := fixtures.QualitySessionsForSet(schema.QualityFixtureSetProjectMix)
	if err != nil {
		t.Fatalf("QualitySessionsForSet(%q): %v", schema.QualityFixtureSetProjectMix, err)
	}
	if len(sessions) != 2 {
		t.Fatalf("len(sessions) = %d, want 2", len(sessions))
	}
	if sessions[0].ID != "sess-000" {
		t.Fatalf("sessions[0].ID = %q, want sess-000", sessions[0].ID)
	}
	if sessions[1].ID != "sess-002" {
		t.Fatalf("sessions[1].ID = %q, want sess-002", sessions[1].ID)
	}
}

// TestLoadQualityFixturesMedians proves the Go aggregate reads the same
// personal-medians corpus the browser baseline exports: the exclusion cases and
// the empty sample are carried verbatim, not recomputed here.
func TestLoadQualityFixturesMedians(t *testing.T) {
	fixtures, err := schema.LoadQualityFixtures()
	if err != nil {
		t.Fatalf("LoadQualityFixtures: %v", err)
	}
	cases := map[string]schema.QualityMedianCase{}
	for _, medianCase := range fixtures.Medians {
		cases[medianCase.Name] = medianCase
	}
	empty, ok := cases["empty_set_returns_no_medians"]
	if !ok {
		t.Fatal("medians corpus is missing the empty sample")
	}
	if len(empty.Sessions) != 0 || empty.Expected != (schema.PersonalMedians{}) {
		t.Fatalf("empty medians sample = %+v, want [] and no expected values", empty)
	}
	negatives, ok := cases["negative_reverts_excluded"]
	if !ok {
		t.Fatal("medians corpus is missing the negative revert sample")
	}
	if negatives.Expected.WithinSessionReverts == nil || *negatives.Expected.WithinSessionReverts != 2 {
		t.Fatalf("negative revert median = %v, want 2 after the negative count is excluded", negatives.Expected.WithinSessionReverts)
	}
	zeroTokens, ok := cases["zero_token_excluded_from_retry_share"]
	if !ok {
		t.Fatal("medians corpus is missing the zero-token sample")
	}
	if zeroTokens.Expected.RetryShare == nil || *zeroTokens.Expected.RetryShare != 0.1 {
		t.Fatalf("zero-token retry share median = %v, want 0.1 from the one nonzero-token session", zeroTokens.Expected.RetryShare)
	}
	nonPositive, ok := cases["non_positive_spec_score_excluded"]
	if !ok {
		t.Fatal("medians corpus is missing the non-positive score sample")
	}
	if nonPositive.Expected.SpecQualityScore == nil || *nonPositive.Expected.SpecQualityScore != 60 {
		t.Fatalf("non-positive score median = %v, want 60 from the single eligible score", nonPositive.Expected.SpecQualityScore)
	}
	if nonPositive.Expected.RetryShare == nil || *nonPositive.Expected.RetryShare != 0 {
		t.Fatalf("non-positive score retry share = %v, want 0 from three sampled sessions with no wasted retries", nonPositive.Expected.RetryShare)
	}
}
