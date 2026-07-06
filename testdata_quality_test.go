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
