package release_test

// The version_kind and parse_tag grammar corpora, migrated onto the canonical
// github.com/peasant-labs/schema/testcase corpus standard: the rows load through
// testcase.LoadCorpus, carry classification + provenance + mutation metadata
// (Corpus.Validate enforces non-vacuity), and each suite guards coverage two
// ways. An EXACT case-count control catches a silent drop or a stray add (a
// min-floor RequireMin would pass a drop that stays above the floor). Because an
// exact count cannot see a NET-SAME swap (drop a real case, add a filler, count
// unchanged), a lean present-by-value coverage check asserts the load-bearing
// cases are still there. All prior assertions from the pre-migration table tests
// are preserved.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
	"github.com/peasant-labs/schema/testcase"
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

func TestVersionKindBaseIsRC(t *testing.T) {
	corpus, err := testcase.LoadCorpus[string, versionKindExpected](versionKindCorpusYAML)
	if err != nil {
		t.Fatalf("load version_kind corpus: %v", err)
	}
	// Exact case-count control: a silent drop or a stray add reddens (a min floor
	// would not). It cannot see a net-same swap, so the value-coverage check below
	// guards the load-bearing cases.
	if got := len(corpus.Cases); got != 3 {
		t.Fatalf("version_kind corpus has %d cases, want exactly 3", got)
	}
	if err := corpus.Validate(); err != nil {
		t.Fatalf("version_kind corpus is under-populated: %v", err)
	}
	// Value coverage: all three load-bearing cases must remain present BY INPUT
	// VALUE, so a net-same swap that drops any of them reddens. For this 3-case
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
	if got := len(corpus.Cases); got != 7 {
		t.Fatalf("parse_tag corpus has %d cases, want exactly 7", got)
	}
	if err := corpus.Validate(); err != nil {
		t.Fatalf("parse_tag corpus is under-populated: %v", err)
	}
	// Value coverage: the load-bearing must-fail inputs must remain present by
	// value, so a net-same swap that drops a real negative (for example the
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
