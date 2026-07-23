package main

// Test-only fixture loader for the go-github seam corpus. The //go:embed lives in
// a _test.go file so the testdata/github tree (canned response bodies + the case
// table) compiles ONLY into the test binary, never into the production tool.

import (
	"embed"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/github
var githubFixtureFS embed.FS

// readGithubFixture returns the bytes of testdata/github/<name>.
func readGithubFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := githubFixtureFS.ReadFile("testdata/github/" + name)
	if err != nil {
		t.Fatalf("read github fixture %q: %v", name, err)
	}
	return data
}

// githubSeamCase is one single-response seam case: serve `Response` with `Status`
// and drive `Method` against the real wrapper. Input fields (User/Commit/PR) feed
// the selected method; a non-empty WantErrContains marks an error case, otherwise
// the method-specific Want* result fields are asserted.
type githubSeamCase struct {
	Name     string `yaml:"name"`
	Method   string `yaml:"method"`
	Status   int    `yaml:"status"`
	Response string `yaml:"response"`

	// inputs
	User   string `yaml:"user"`
	Commit string `yaml:"commit"`
	PR     int    `yaml:"pr"`

	// git-data inputs (Ref/Commit/Pull/CreateCommit/UpdateRefFastForward)
	Ref     string   `yaml:"ref"`
	SHA     string   `yaml:"sha"`
	Tree    string   `yaml:"tree"`
	Parents []string `yaml:"parents"`
	Message string   `yaml:"message"`
	NewSHA  string   `yaml:"newSHA"`

	// expected outcomes
	WantErrContains []string `yaml:"wantErrContains"`
	WantPath        string   `yaml:"wantPath"`
	WantPermission  string   `yaml:"wantPermission"`
	WantRunCount    int      `yaml:"wantRunCount"`
	WantGreen       bool     `yaml:"wantGreen"`
	WantReviewCount int      `yaml:"wantReviewCount"`
	WantApprovers   []string `yaml:"wantApprovers"`

	// git-data expected outcomes
	WantSHA            string   `yaml:"wantSHA"`
	WantTreeSHA        string   `yaml:"wantTreeSHA"`
	WantParents        []string `yaml:"wantParents"`
	WantTitle          string   `yaml:"wantTitle"`
	WantNumber         int      `yaml:"wantNumber"`
	WantNotFastForward bool     `yaml:"wantNotFastForward"`
}

// githubSeamCases is the parsed testdata/github/cases.yaml corpus.
type githubSeamCases struct {
	Cases []githubSeamCase `yaml:"cases"`
}

// loadGithubSeamCases parses the embedded seam case table.
func loadGithubSeamCases(t *testing.T) []githubSeamCase {
	t.Helper()
	data, err := githubFixtureFS.ReadFile("testdata/github/cases.yaml")
	if err != nil {
		t.Fatalf("read github seam cases: %v", err)
	}
	var c githubSeamCases
	if err := yaml.Unmarshal(data, &c); err != nil {
		t.Fatalf("load github seam cases (testdata/github/cases.yaml): %v", err)
	}
	return c.Cases
}
