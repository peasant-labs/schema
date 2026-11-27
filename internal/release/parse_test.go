package release_test

// The grammar table CASES live in testdata/grammar/versions.yaml (loaded via
// loadGrammarFixtures); these tests iterate the fixture rows. The assertions are
// identical to the pre-extraction inline tables. A required-name manifest
// guards against accidental truncation.

import (
	_ "embed"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

//go:embed testdata/grammar/versions_manifest.yaml
var versionsManifestYAML []byte

func TestNewVersion(t *testing.T) {
	cases := loadGrammarFixtures(t).NewVersion
	names := make([]string, 0, len(cases))
	for _, c := range cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "version cases", versionsManifestYAML, names)
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
