package intent_test

// Shared corpus harness for the intent package: every fixture file carries a
// required_names manifest (the deletion guard) plus a cases list in the
// testcase schema. The loader checks the manifest by exact set equality, so a
// dropped case and a filler case both redden the suite, then validates every
// case for non-vacuity.

import (
	"bytes"
	"errors"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
	"gopkg.in/yaml.v3"
)

// namedCorpus is one fixture file: its manifest plus its typed cases.
type namedCorpus[I any, E any] struct {
	RequiredNames []string           `yaml:"required_names"`
	Cases         []corpusCase[I, E] `yaml:"cases"`
}

// corpusCase mirrors testcase.Case without importing testing into the shape.
type corpusCase[I any, E any] struct {
	Name           string `yaml:"name"`
	Input          I      `yaml:"input"`
	Expected       E      `yaml:"expected"`
	Classification string `yaml:"classification"`
	// Sibling names the accept case a must-fail row is one fault away
	// from: the suite proves the sibling exists, passes, and stays passing,
	// so the reject fails for the named fault and not an unrelated error.
	Sibling    string `yaml:"sibling"`
	Provenance struct {
		Source string `yaml:"source"`
		Ref    string `yaml:"ref"`
	} `yaml:"provenance"`
	Mutation struct {
		Description string `yaml:"description"`
	} `yaml:"mutation"`
}

func loadNamedCorpus[I any, E any](t *testing.T, data []byte) namedCorpus[I, E] {
	t.Helper()
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var c namedCorpus[I, E]
	if err := decoder.Decode(&c); err != nil {
		t.Fatalf("decode corpus document: %v", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			t.Fatalf("decode trailing corpus document: %v", err)
		}
		t.Fatalf("multiple YAML documents are not allowed")
	}
	if len(c.RequiredNames) == 0 {
		t.Fatalf("required_names is empty; a corpus with no named case can lose coverage silently")
	}
	seen := make(map[string]bool, len(c.Cases))
	for _, cse := range c.Cases {
		if strings.TrimSpace(cse.Name) == "" {
			t.Fatalf("corpus case has an empty name")
		}
		if seen[cse.Name] {
			t.Fatalf("duplicate case name %q", cse.Name)
		}
		seen[cse.Name] = true
		if cse.Classification != "must-pass" && cse.Classification != "must-fail" {
			t.Fatalf("case %q: classification %q is not must-pass/must-fail", cse.Name, cse.Classification)
		}
		if strings.TrimSpace(cse.Provenance.Source) == "" || strings.TrimSpace(cse.Provenance.Ref) == "" {
			t.Fatalf("case %q: provenance must name a source and a ref", cse.Name)
		}
		if strings.TrimSpace(cse.Mutation.Description) == "" {
			t.Fatalf("case %q: mutation description is empty", cse.Name)
		}
	}
	classOf := make(map[string]string, len(c.Cases))
	for _, cse := range c.Cases {
		classOf[cse.Name] = cse.Classification
	}
	for _, cse := range c.Cases {
		if cse.Classification == "must-fail" {
			if strings.TrimSpace(cse.Sibling) == "" {
				t.Fatalf("case %q: a reject case must name its accept sibling", cse.Name)
			}
			siblingClass, ok := classOf[cse.Sibling]
			if !ok {
				t.Fatalf("case %q: sibling %q is not in the corpus", cse.Name, cse.Sibling)
			}
			if siblingClass != "must-pass" {
				t.Fatalf("case %q: sibling %q is %s, want must-pass", cse.Name, cse.Sibling, siblingClass)
			}
		} else if strings.TrimSpace(cse.Sibling) != "" {
			t.Fatalf("case %q: an accept case names no sibling", cse.Name)
		}
	}
	want := append([]string(nil), c.RequiredNames...)
	sort.Strings(want)
	got := make([]string, 0, len(seen))
	for name := range seen {
		got = append(got, name)
	}
	sort.Strings(got)
	if len(want) != len(got) {
		t.Fatalf("manifest names %d cases but the corpus carries %d", len(want), len(got))
	}
	for i := range want {
		if want[i] != got[i] {
			t.Fatalf("manifest mismatch: want %q, corpus has %q", want[i], got[i])
		}
	}
	return c
}

// requireSixParts fails unless an intent rejection carries every part.
func requireSixParts(t *testing.T, err error, wantCode string) {
	t.Helper()
	var rej *intent.Error
	if !errors.As(err, &rej) {
		t.Fatalf("expected a typed intent.Error, got %T: %v", err, err)
	}
	if string(rej.Code) != wantCode {
		t.Fatalf("code = %q, want %q", rej.Code, wantCode)
	}
	parts := map[string]string{
		"what": rej.What, "why": rej.Why, "where": rej.Where,
		"when": rej.When, "means": rej.Means, "fix": rej.Fix,
	}
	for name, value := range parts {
		if strings.TrimSpace(value) == "" {
			t.Errorf("rejection %q has an empty %s part", wantCode, name)
		}
	}
}

// requireSixLabels fails unless a plain error message carries all six labels.
// Plain errors are the exit-1 faults (unreadable files, malformed values),
// not typed rejections.
func requireSixLabels(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	var rej *intent.Error
	if errors.As(err, &rej) {
		t.Fatalf("expected a plain error, got a typed rejection with code %q", rej.Code)
	}
	for _, label := range []string{"why:", "where:", "when:", "means:", "fix:"} {
		if !strings.Contains(err.Error(), label) {
			t.Errorf("error message is missing the %q part: %v", label, err)
		}
	}
}
