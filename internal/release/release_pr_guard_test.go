package release

// Fixture-driven guard for the release-PR concurrency contract: a body edit
// must queue behind the running push run instead of cancelling it. The
// fixtures live under testdata/release-pr (embedded only into the test
// binary, never into the production release-guard tool); cases.yaml names the
// accept/reject pairs. TestCheckReleasePRWorkflow_Cases proves each rejection
// fires, and TestCheckReleasePRWorkflow_CurrentFile pins the real
// .github/workflows/release-pr.yml to the same contract (the peasant-style
// read-the-real-workflow pin).

import (
	"embed"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/release-pr
var releasePRFixtureFS embed.FS

// releasePRCase is one trigger-plus-concurrency validation case. Workflow
// names an embedded testdata/release-pr/*.yml file. An empty WantContains
// means the pair must be ACCEPTED (no error); a non-empty WantContains means
// the resulting error MUST contain every listed substring.
type releasePRCase struct {
	Name         string   `yaml:"name"`
	Workflow     string   `yaml:"workflow"`
	WantContains []string `yaml:"wantContains"`
}

// releasePRCases is the parsed testdata/release-pr/cases.yaml corpus.
type releasePRCases struct {
	Accept []releasePRCase `yaml:"accept"`
	Reject []releasePRCase `yaml:"reject"`
}

// loadReleasePRCases parses the embedded release-PR case table.
func loadReleasePRCases(t *testing.T) releasePRCases {
	t.Helper()
	data, err := releasePRFixtureFS.ReadFile("testdata/release-pr/cases.yaml")
	if err != nil {
		t.Fatalf("read release-PR cases: %v", err)
	}
	var c releasePRCases
	if err := yaml.Unmarshal(data, &c); err != nil {
		t.Fatalf("load release-PR cases (testdata/release-pr/cases.yaml): %v", err)
	}
	return c
}

// readReleasePRFixture returns the bytes of testdata/release-pr/<name>.
func readReleasePRFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := releasePRFixtureFS.ReadFile("testdata/release-pr/" + name)
	if err != nil {
		t.Fatalf("read release-PR fixture %q: %v", name, err)
	}
	return data
}

func TestCheckReleasePRWorkflow_Cases(t *testing.T) {
	t.Parallel()

	cases := loadReleasePRCases(t)
	if len(cases.Accept) != 1 || len(cases.Reject) != 4 {
		t.Fatalf("release-PR fixture has accept=%d reject=%d, want accept=1 reject=4 (fixture truncated?)",
			len(cases.Accept), len(cases.Reject))
	}

	t.Run("accept", func(t *testing.T) {
		t.Parallel()
		for _, tc := range cases.Accept {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				t.Parallel()
				data := readReleasePRFixture(t, tc.Workflow)
				if err := checkReleasePRWorkflow(tc.Workflow, data); err != nil {
					t.Fatalf("accept case %q rejected: %v", tc.Name, err)
				}
			})
		}
	})

	t.Run("reject", func(t *testing.T) {
		t.Parallel()
		for _, tc := range cases.Reject {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				t.Parallel()
				if len(tc.WantContains) == 0 {
					t.Fatalf("reject case %q declares no wantContains substrings (fixture error)", tc.Name)
				}
				data := readReleasePRFixture(t, tc.Workflow)
				err := checkReleasePRWorkflow(tc.Workflow, data)
				if err == nil {
					t.Fatalf("expected rejection for %q, got nil", tc.Name)
				}
				for _, want := range tc.WantContains {
					if !strings.Contains(err.Error(), want) {
						t.Fatalf("error %q does not contain expected substring %q", err.Error(), want)
					}
				}
			})
		}
	})
}

// TestCheckReleasePRWorkflow_CurrentFile pins the real workflow to the same
// contract: the committed .github/workflows/release-pr.yml must keep the
// edited trigger and the edited-exempting cancel-in-progress expression.
func TestCheckReleasePRWorkflow_CurrentFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", ".github", "workflows", "release-pr.yml")
	if err := CheckReleasePRWorkflowFile(path); err != nil {
		t.Fatalf("real release-PR workflow rejected: %v", err)
	}
}
