package main

import (
	"embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/group-pr.yaml testdata/classify.yaml testdata/queue-gate.yaml testdata/title.yaml testdata/cli-exit.yaml
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

type ClassifyInput struct {
	Author      string            `yaml:"author"`
	HeadBranch  string            `yaml:"headBranch"`
	BotBranches map[string]string `yaml:"botBranches"`
}

type ClassifyExpected struct {
	Kind string `yaml:"kind"`
}

type ClassifyFixtures struct {
	RequiredNames []string                                         `yaml:"required_names"`
	Cases         testcase.Corpus[ClassifyInput, ClassifyExpected] `yaml:"cases"`
}

type QueueGateInput struct {
	File      string `yaml:"file"`
	ReleasePR int    `yaml:"releasePR"`
}

type QueueGateExpected struct {
	Gate string `yaml:"gate"`
	Exit int    `yaml:"exit"`
}

type QueueGateFixtures struct {
	RequiredNames []string                                           `yaml:"required_names"`
	Cases         testcase.Corpus[QueueGateInput, QueueGateExpected] `yaml:"cases"`
}

type TitleInput struct {
	Title    string `yaml:"title"`
	Releases bool   `yaml:"releases"`
	BelowOne bool   `yaml:"belowOne"`
}

type TitleExpected struct {
	Code string `yaml:"code"`
}

type TitleFixtures struct {
	RequiredNames []string                                   `yaml:"required_names"`
	Cases         testcase.Corpus[TitleInput, TitleExpected] `yaml:"cases"`
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

func loadClassifyFixtures(t *testing.T) ClassifyFixtures {
	t.Helper()
	f := loadFixtures[ClassifyFixtures](t, "classify.yaml")
	if err := f.Cases.Validate(); err != nil {
		t.Fatalf("classify fixtures: %v", err)
	}
	seen := map[string]struct{}{}
	for _, c := range f.Cases.Cases {
		seen[c.Name] = struct{}{}
	}
	checkRequiredNames(t, f.RequiredNames, seen, "classify fixtures")
	return f
}

func loadQueueGateFixtures(t *testing.T) QueueGateFixtures {
	t.Helper()
	f := loadFixtures[QueueGateFixtures](t, "queue-gate.yaml")
	if err := f.Cases.Validate(); err != nil {
		t.Fatalf("queue-gate fixtures: %v", err)
	}
	seen := map[string]struct{}{}
	for _, c := range f.Cases.Cases {
		seen[c.Name] = struct{}{}
	}
	checkRequiredNames(t, f.RequiredNames, seen, "queue-gate fixtures")
	return f
}

func loadTitleFixtures(t *testing.T) TitleFixtures {
	t.Helper()
	f := loadFixtures[TitleFixtures](t, "title.yaml")
	if err := f.Cases.Validate(); err != nil {
		t.Fatalf("title fixtures: %v", err)
	}
	seen := map[string]struct{}{}
	for _, c := range f.Cases.Cases {
		seen[c.Name] = struct{}{}
	}
	checkRequiredNames(t, f.RequiredNames, seen, "title fixtures")
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
	wantActions := map[Action]bool{"none": true, "verify": true, "tag": true, "open-pr": true, "close-pr": true}
	if len(AllActions) != len(wantActions) {
		t.Fatalf("AllActions has %d members, want %d", len(AllActions), len(wantActions))
	}
	for _, a := range AllActions {
		if !wantActions[a] {
			t.Fatalf("AllActions carries unexpected member %q", a)
		}
	}
	// The command-layer action set must match the contract package so the two
	// never drift: same members, same spelling.
	if len(intent.AllActions) != len(AllActions) {
		t.Fatalf("intent.AllActions has %d members, command AllActions has %d: the two action sets drifted", len(intent.AllActions), len(AllActions))
	}
	for _, a := range intent.AllActions {
		if !wantActions[Action(a)] {
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
	wantCodes := map[ErrorCode]bool{
		"malformed": true, "unknown-type": true, "missing-bump": true, "bump-on-non-releasing": true,
		"bang-bump-mismatch": true, "old-release-title": true, "releases-disabled": true, "protected-path": true,
		"no-final-tag": true, "epoch-missing": true, "floor-not-on-chain": true, "release-content-stale": true,
	}
	if len(AllErrorCodes) != len(wantCodes) {
		t.Fatalf("AllErrorCodes has %d members, want %d", len(AllErrorCodes), len(wantCodes))
	}
	for _, c := range AllErrorCodes {
		if !wantCodes[c] {
			t.Fatalf("AllErrorCodes carries unexpected member %q", c)
		}
	}
	if len(intent.AllErrorCodes) != len(AllErrorCodes) {
		t.Fatalf("intent.AllErrorCodes has %d members, command AllErrorCodes has %d: the two code sets drifted", len(intent.AllErrorCodes), len(AllErrorCodes))
	}
	for _, c := range intent.AllErrorCodes {
		if !wantCodes[ErrorCode(c)] {
			t.Fatalf("intent.AllErrorCodes carries unexpected member %q", c)
		}
	}
}
