package main

import (
	"embed"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/group-pr.yaml testdata/cli-exit.yaml testdata/check-title.yaml testdata/freshness.yaml
var cliFixtureYAML embed.FS

//go:embed testdata/queue/*.json
var queueJSONFS embed.FS

type GroupPRInput struct {
	Ref string `yaml:"ref"`
}

type GroupPRExpected struct {
	PR string `yaml:"pr"`
}

type GroupPRFixtures struct {
	RequiredNames []string                                       `yaml:"required_names"`
	Cases         testcase.Corpus[GroupPRInput, GroupPRExpected] `yaml:"cases"`
}

type CLIInput struct {
	Argv         []string `yaml:"argv"`
	Stdin        string   `yaml:"stdin"`
	Setup        string   `yaml:"setup"`
	QueueFile    string   `yaml:"queueFile"`
	GithubOutput bool     `yaml:"githubOutput"`
}

type CLIExpected struct {
	Exit   int      `yaml:"exit"`
	Stdout []string `yaml:"stdout"`
	Absent []string `yaml:"absent"`
	Stderr []string `yaml:"stderr"`
}

type CLIFixtures struct {
	RequiredNames []string                               `yaml:"required_names"`
	Cases         testcase.Corpus[CLIInput, CLIExpected] `yaml:"cases"`
}

func loadFixtures[T any](t *testing.T, name string) T {
	t.Helper()
	data, err := cliFixtureYAML.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	var f T
	if err := yaml.Unmarshal(data, &f); err != nil {
		t.Fatalf("parse fixture %q: %v", name, err)
	}
	return f
}

func checkRequiredNames(t *testing.T, required []string, seen map[string]struct{}, what string) {
	t.Helper()
	if len(required) == 0 {
		t.Fatalf("%s: required_names is empty; a corpus with no named case can lose coverage silently", what)
	}
	for _, name := range required {
		if _, ok := seen[name]; !ok {
			t.Fatalf("%s: required case %q is missing", what, name)
		}
	}
	if len(seen) != len(required) {
		t.Fatalf("%s: required_names names %d cases but the corpus carries %d", what, len(required), len(seen))
	}
}

func loadGroupPRFixtures(t *testing.T) GroupPRFixtures {
	t.Helper()
	f := loadFixtures[GroupPRFixtures](t, "group-pr.yaml")
	if err := f.Cases.Validate(); err != nil {
		t.Fatalf("group-pr fixtures: %v", err)
	}
	seen := map[string]struct{}{}
	for _, c := range f.Cases.Cases {
		seen[c.Name] = struct{}{}
	}
	checkRequiredNames(t, f.RequiredNames, seen, "group-pr fixtures")
	return f
}

func loadCLIFixtures(t *testing.T) CLIFixtures {
	t.Helper()
	f := loadFixtures[CLIFixtures](t, "cli-exit.yaml")
	if err := f.Cases.Validate(); err != nil {
		t.Fatalf("cli-exit fixtures: %v", err)
	}
	seen := map[string]struct{}{}
	for _, c := range f.Cases.Cases {
		seen[c.Name] = struct{}{}
	}
	checkRequiredNames(t, f.RequiredNames, seen, "cli-exit fixtures")
	data, err := cliFixtureYAML.ReadFile("testdata/cli-exit.yaml")
	if err != nil {
		t.Fatalf("read cli-exit fixture for siblings: %v", err)
	}
	checkCorpusSiblings(t, data, "cli-exit fixtures")
	return f
}

// checkCorpusSiblings enforces the accept-sibling rule: every must-fail row
// names an accept sibling in the same corpus, so a reject proves it fails for
// the named fault and not for an unrelated input error.
func checkCorpusSiblings(t *testing.T, data []byte, what string) {
	t.Helper()
	var doc struct {
		Cases struct {
			Cases []struct {
				Name           string `yaml:"name"`
				Classification string `yaml:"classification"`
				Sibling        string `yaml:"sibling"`
			} `yaml:"cases"`
		} `yaml:"cases"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("%s: cannot parse sibling references: %v", what, err)
	}
	byName := map[string]string{}
	for _, c := range doc.Cases.Cases {
		byName[c.Name] = c.Classification
	}
	for _, c := range doc.Cases.Cases {
		if c.Classification != "must-fail" {
			continue
		}
		if c.Sibling == "" {
			t.Fatalf("%s: must-fail case %q names no accept sibling", what, c.Name)
		}
		class, ok := byName[c.Sibling]
		if !ok {
			t.Fatalf("%s: must-fail case %q names missing sibling %q", what, c.Name, c.Sibling)
		}
		if class != "must-pass" {
			t.Fatalf("%s: must-fail case %q names sibling %q which is %q, not must-pass", what, c.Name, c.Sibling, class)
		}
	}
}
