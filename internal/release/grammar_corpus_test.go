package release_test

// The version-kind and parse-tag corpora cover the frozen tag grammar.
// The tag corpus uses a required-name manifest and preserves value coverage.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

// versionKindExpected is the derived output of one version_kind case.
type versionKindExpected struct {
	Kind release.ReleaseKind `yaml:"kind"`
	Base release.Version     `yaml:"base"`
	RC   bool                `yaml:"rc"`
}

// parseResult is the parsed output of one parse_tag must-pass case.
type parseResult struct {
	Ver  release.Version     `yaml:"ver"`
	Kind release.ReleaseKind `yaml:"kind"`
}

//go:embed testdata/grammar/version_kind_corpus.yaml
var versionKindCorpusYAML []byte

//go:embed testdata/grammar/parse_tag_corpus.yaml
var parseTagCorpusYAML []byte

//go:embed testdata/grammar/version_kind_manifest.yaml
var versionKindManifestYAML []byte

//go:embed testdata/grammar/parse_tag_manifest.yaml
var parseTagManifestYAML []byte

func TestVersionKindBaseIsRC(t *testing.T) {
	corpus, err := testcase.LoadCorpus[string, versionKindExpected](versionKindCorpusYAML)
	if err != nil {
		t.Fatalf("load version_kind corpus: %v", err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "version kind cases", versionKindManifestYAML, names)
	if err := corpus.Validate(); err != nil {
		t.Fatalf("version_kind corpus is under-populated: %v", err)
	}
	// Value coverage: all three load-bearing cases must remain present BY INPUT
	// VALUE, so a count-preserving swap that drops any of them reddens. For this 3-case
	// corpus the set is complete-but-minimal, each input exercising a distinct
	// behavior: the final case (v0.1.0, the only IsRC=false path), the rc case
	// (v0.1.0-rc1), and the multi-digit-rc boundary case (v3.2.1-rc9).
	present := map[string]bool{}
	for _, c := range corpus.Cases {
		present[c.Input] = true
	}
	for _, want := range []string{"v0.1.0", "v0.1.0-rc1", "v3.2.1-rc9"} {
		if !present[want] {
			t.Errorf("value coverage lost: version_kind case for input %q is missing", want)
		}
	}

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			v, err := release.NewVersion(c.Input)
			if err != nil {
				t.Fatalf("NewVersion(%q): %v", c.Input, err)
			}
			if got := v.Kind(); got != c.Expected.Kind {
				t.Errorf("Kind() = %q, want %q", got, c.Expected.Kind)
			}
			if got := v.Base(); got != c.Expected.Base {
				t.Errorf("Base() = %q, want %q", got, c.Expected.Base)
			}
			if got := v.IsRC(); got != c.Expected.RC {
				t.Errorf("IsRC() = %v, want %v", got, c.Expected.RC)
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	corpus, err := testcase.LoadCorpus[string, parseResult](parseTagCorpusYAML)
	if err != nil {
		t.Fatalf("load parse_tag corpus: %v", err)
	}
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "parse tag cases", parseTagManifestYAML, names)
	if err := corpus.Validate(); err != nil {
		t.Fatalf("parse_tag corpus is under-populated: %v", err)
	}
	// Value coverage: the load-bearing must-fail inputs must remain present by
	// value, so a count-preserving swap that drops a real negative (for example the
	// namespaced tag that must never parse as a release) reddens.
	present := map[string]bool{}
	for _, c := range corpus.Cases {
		present[c.Input] = true
	}
	for _, want := range []string{"pkg/schema/v1.2.3", "v0.1.0-rc", ""} {
		if !present[want] {
			t.Errorf("value coverage lost: must-fail case for input %q is missing", want)
		}
	}

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			v, k, err := release.ParseTag(c.Input)
			if c.Classification == testcase.MustFail {
				if err == nil {
					t.Fatalf("ParseTag(%q): expected error, got (%q, %q)", c.Input, v, k)
				}
				if k != release.KindInvalid {
					t.Errorf("ParseTag(%q): want KindInvalid on error, got %q", c.Input, k)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTag(%q): unexpected error: %v", c.Input, err)
			}
			if v != c.Expected.Ver || k != c.Expected.Kind {
				t.Errorf("ParseTag(%q) = (%q, %q), want (%q, %q)", c.Input, v, k, c.Expected.Ver, c.Expected.Kind)
			}
		})
	}
}
