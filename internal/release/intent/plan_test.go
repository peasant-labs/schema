package intent_test

// Planner suite: drives intent.Plan over the plan corpus.

import (
	_ "embed"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/plan_corpus.yaml
var planCorpusYAML []byte

type planInput struct {
	Config intent.Config `yaml:"config"`
	Facts  intent.Facts  `yaml:"facts"`
}

type planExpected struct {
	Action         string                 `yaml:"action"`
	RangeBase      string                 `yaml:"range_base"`
	VersionBase    string                 `yaml:"version_base"`
	Candidate      string                 `yaml:"candidate"`
	CandidateFresh string                 `yaml:"candidate_fresh"`
	Pending        []intent.Title         `yaml:"pending"`
	Level          intent.Level           `yaml:"level"`
	Breaking       bool                   `yaml:"breaking"`
	Version        string                 `yaml:"version"`
	Tag            string                 `yaml:"tag"`
	Before         intent.BeforeCandidate `yaml:"before"`
	Full           intent.BeforeCandidate `yaml:"full"`
	Date           string                 `yaml:"date"`
	Code           string                 `yaml:"code"`
	FixContains    string                 `yaml:"fix_contains"`
}

func TestPlan(t *testing.T) {
	corpus := loadNamedCorpus[planInput, planExpected](t, planCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := intent.Plan(c.Input.Config, c.Input.Facts)
			if c.Classification == "must-fail" {
				if err == nil {
					t.Fatalf("Plan: expected an error, got %+v", got)
				}
				requireSixParts(t, err, c.Expected.Code)
				if c.Expected.FixContains != "" && !strings.Contains(err.Error(), c.Expected.FixContains) {
					t.Errorf("error message is missing %q: %v", c.Expected.FixContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Plan: unexpected error: %v", err)
			}
			if string(got.Action) != c.Expected.Action {
				t.Errorf("Action = %q, want %q", got.Action, c.Expected.Action)
			}
			if got.RangeBase != c.Expected.RangeBase {
				t.Errorf("RangeBase = %q, want %q", got.RangeBase, c.Expected.RangeBase)
			}
			if got.VersionBase != c.Expected.VersionBase {
				t.Errorf("VersionBase = %q, want %q", got.VersionBase, c.Expected.VersionBase)
			}
			if got.Candidate != c.Expected.Candidate {
				t.Errorf("Candidate = %q, want %q", got.Candidate, c.Expected.Candidate)
			}
			if string(got.CandidateFresh) != c.Expected.CandidateFresh {
				t.Errorf("CandidateFresh = %q, want %q", got.CandidateFresh, c.Expected.CandidateFresh)
			}
			if len(got.Pending) != len(c.Expected.Pending) {
				t.Fatalf("Pending has %d titles, want %d", len(got.Pending), len(c.Expected.Pending))
			}
			for i := range got.Pending {
				want := c.Expected.Pending[i]
				if !reflect.DeepEqual(got.Pending[i], want) {
					t.Errorf("Pending[%d] = %+v, want %+v", i, got.Pending[i], want)
				}
			}
			if got.Level != c.Expected.Level {
				t.Errorf("Level = %q, want %q", got.Level, c.Expected.Level)
			}
			if got.Breaking != c.Expected.Breaking {
				t.Errorf("Breaking = %v, want %v", got.Breaking, c.Expected.Breaking)
			}
			if string(got.Version) != c.Expected.Version {
				t.Errorf("Version = %q, want %q", got.Version, c.Expected.Version)
			}
			if got.Tag != c.Expected.Tag {
				t.Errorf("Tag = %q, want %q", got.Tag, c.Expected.Tag)
			}
			if string(got.Date) != c.Expected.Date {
				t.Errorf("Date = %q, want %q", got.Date, c.Expected.Date)
			}
			compareRangeSet(t, "Before", got.Before, c.Expected.Before)
			compareRangeSet(t, "Full", got.Full, c.Expected.Full)
		})
	}
}

func compareRangeSet(t *testing.T, what string, got, want intent.BeforeCandidate) {
	t.Helper()
	if len(got.Pending) != len(want.Pending) {
		t.Fatalf("%s.Pending has %d titles, want %d", what, len(got.Pending), len(want.Pending))
	}
	for i := range got.Pending {
		if !reflect.DeepEqual(got.Pending[i], want.Pending[i]) {
			t.Errorf("%s.Pending[%d] = %+v, want %+v", what, i, got.Pending[i], want.Pending[i])
		}
	}
	if string(got.Version) != string(want.Version) {
		t.Errorf("%s.Version = %q, want %q", what, got.Version, want.Version)
	}
	if got.Tag != want.Tag {
		t.Errorf("%s.Tag = %q, want %q", what, got.Tag, want.Tag)
	}
	if string(got.Date) != string(want.Date) {
		t.Errorf("%s.Date = %q, want %q", what, got.Date, want.Date)
	}
}
