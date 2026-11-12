package main

import (
	"embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/group-pr.yaml testdata/cli-exit.yaml
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
	Argv  []string `yaml:"argv"`
	Stdin string   `yaml:"stdin"`
	Setup string   `yaml:"setup"`
}

type CLIExpected struct {
	Exit   int      `yaml:"exit"`
	Stdout []string `yaml:"stdout"`
	NoGate bool     `yaml:"noGate"`
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
	return f
}

func TestClosedSetsAreExact(t *testing.T) {
	t.Parallel()
	wantActions := map[intent.Action]bool{"none": true, "verify": true, "tag": true, "open-pr": true, "close-pr": true}
	if len(intent.AllActions) != len(wantActions) {
		t.Fatalf("intent.AllActions has %d members, want %d", len(intent.AllActions), len(wantActions))
	}
	for _, a := range intent.AllActions {
		if !wantActions[a] {
			t.Fatalf("intent.AllActions carries unexpected member %q", a)
		}
	}
	wantKinds := map[intent.PRKind]bool{"ordinary": true, "release": true, "vendor-hash": true, "cut-release": true}
	if len(intent.AllPRKinds) != len(wantKinds) {
		t.Fatalf("intent.AllPRKinds has %d members, want %d", len(intent.AllPRKinds), len(wantKinds))
	}
	for _, k := range intent.AllPRKinds {
		if !wantKinds[k] {
			t.Fatalf("intent.AllPRKinds carries unexpected member %q", k)
		}
	}
	wantCodes := map[intent.ErrorCode]bool{
		"malformed": true, "unknown-type": true, "missing-bump": true, "bump-on-non-releasing": true,
		"bang-bump-mismatch": true, "old-release-title": true, "releases-disabled": true, "protected-path": true,
		"no-final-tag": true, "epoch-missing": true, "floor-not-on-chain": true, "release-content-stale": true,
	}
	if len(intent.AllErrorCodes) != len(wantCodes) {
		t.Fatalf("intent.AllErrorCodes has %d members, want %d", len(intent.AllErrorCodes), len(wantCodes))
	}
	for _, c := range intent.AllErrorCodes {
		if !wantCodes[c] {
			t.Fatalf("intent.AllErrorCodes carries unexpected member %q", c)
		}
	}
	wantGates := map[intent.Gate]bool{"arm": true, "hold": true}
	if len(intent.AllGates) != len(wantGates) {
		t.Fatalf("intent.AllGates has %d members, want %d", len(intent.AllGates), len(wantGates))
	}
	for _, g := range intent.AllGates {
		if !wantGates[g] {
			t.Fatalf("intent.AllGates carries unexpected member %q", g)
		}
	}
	wantFresh := map[intent.Freshness]bool{"unknown": true, "fresh": true, "stale": true}
	if len(intent.AllFreshness) != len(wantFresh) {
		t.Fatalf("intent.AllFreshness has %d members, want %d", len(intent.AllFreshness), len(wantFresh))
	}
	for _, f := range intent.AllFreshness {
		if !wantFresh[f] {
			t.Fatalf("intent.AllFreshness carries unexpected member %q", f)
		}
	}
}
