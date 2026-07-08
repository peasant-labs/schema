package release_test

// The grammar table CASES live in testdata/grammar/versions.yaml (loaded via
// loadGrammarFixtures); these tests iterate the fixture rows. The assertions are
// identical to the pre-extraction inline tables — only the data moved out. Each
// test asserts the exact row count so an accidental fixture truncation fails
// loudly rather than silently dropping coverage.

import (
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

func TestNewVersion(t *testing.T) {
	cases := loadGrammarFixtures(t).NewVersion
	if len(cases) != 15 {
		t.Fatalf("grammar fixture new_version has %d rows, want 15 (fixture truncated?)", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := release.NewVersion(tc.Raw)
			if tc.WantErr {
				if err == nil {
					t.Fatalf("NewVersion(%q): expected error, got %q", tc.Raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewVersion(%q): unexpected error: %v", tc.Raw, err)
			}
			if got != tc.Want {
				t.Errorf("NewVersion(%q) = %q, want %q", tc.Raw, got, tc.Want)
			}
		})
	}
}

func TestVersionKindBaseIsRC(t *testing.T) {
	cases := loadGrammarFixtures(t).VersionKind
	if len(cases) != 3 {
		t.Fatalf("grammar fixture version_kind has %d rows, want 3 (fixture truncated?)", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Raw, func(t *testing.T) {
			v, err := release.NewVersion(tc.Raw)
			if err != nil {
				t.Fatalf("NewVersion(%q): %v", tc.Raw, err)
			}
			if got := v.Kind(); got != tc.WantKind {
				t.Errorf("Kind() = %q, want %q", got, tc.WantKind)
			}
			if got := v.Base(); got != tc.WantBase {
				t.Errorf("Base() = %q, want %q", got, tc.WantBase)
			}
			if got := v.IsRC(); got != tc.WantRC {
				t.Errorf("IsRC() = %v, want %v", got, tc.WantRC)
			}
		})
	}
}

func TestParseReleaseTitle(t *testing.T) {
	cases := loadGrammarFixtures(t).ParseTitle
	if len(cases) != 13 {
		t.Fatalf("grammar fixture parse_title has %d rows, want 13 (fixture truncated?)", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			v, k, err := release.ParseReleaseTitle(tc.Title)
			if tc.WantErr {
				if err == nil {
					t.Fatalf("ParseReleaseTitle(%q): expected error, got (%q, %q)", tc.Title, v, k)
				}
				if k != release.KindInvalid {
					t.Errorf("ParseReleaseTitle(%q): want KindInvalid on error, got %q", tc.Title, k)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseReleaseTitle(%q): unexpected error: %v", tc.Title, err)
			}
			if v != tc.WantVer || k != tc.WantKind {
				t.Errorf("ParseReleaseTitle(%q) = (%q, %q), want (%q, %q)", tc.Title, v, k, tc.WantVer, tc.WantKind)
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	cases := loadGrammarFixtures(t).ParseTag
	if len(cases) != 7 {
		t.Fatalf("grammar fixture parse_tag has %d rows, want 7 (fixture truncated?)", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			v, k, err := release.ParseTag(tc.Tag)
			if tc.WantErr {
				if err == nil {
					t.Fatalf("ParseTag(%q): expected error, got (%q, %q)", tc.Tag, v, k)
				}
				if k != release.KindInvalid {
					t.Errorf("ParseTag(%q): want KindInvalid on error, got %q", tc.Tag, k)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTag(%q): unexpected error: %v", tc.Tag, err)
			}
			if v != tc.WantVer || k != tc.WantKind {
				t.Errorf("ParseTag(%q) = (%q, %q), want (%q, %q)", tc.Tag, v, k, tc.WantVer, tc.WantKind)
			}
		})
	}
}
