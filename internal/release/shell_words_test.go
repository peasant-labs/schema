package release

import (
	_ "embed"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/shell_words.yaml
var shellWordsFixtureYAML []byte

// shellWordsCase is one splitter case: the input line, its words with quotes
// kept, the bare words, and the refspecs a git push with these arguments
// resolves to.
type shellWordsCase struct {
	Name         string   `yaml:"name"`
	Input        string   `yaml:"input"`
	Words        []string `yaml:"words"`
	Bare         []string `yaml:"bare"`
	PushRefspecs []string `yaml:"push_refspecs"`
}

type shellWordsFixtures struct {
	RequiredNames []string         `yaml:"required_names"`
	Cases         []shellWordsCase `yaml:"cases"`
}

// loadShellWordsFixtures loads the splitter corpus and holds it to its
// required-name manifest: every case listed, every listed case present.
func loadShellWordsFixtures(t *testing.T) shellWordsFixtures {
	t.Helper()
	var fx shellWordsFixtures
	if err := yaml.Unmarshal(shellWordsFixtureYAML, &fx); err != nil {
		t.Fatalf("parse testdata/shell_words.yaml: %v", err)
	}
	seen := map[string]bool{}
	for _, c := range fx.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("shell-words case name %q is empty or repeated", c.Name)
		}
		seen[c.Name] = true
		if !slices.Contains(fx.RequiredNames, c.Name) {
			t.Fatalf("shell-words case %q is not in required_names", c.Name)
		}
	}
	for _, name := range fx.RequiredNames {
		if !seen[name] {
			t.Fatalf("required shell-words case %q is missing from the corpus", name)
		}
	}
	return fx
}

// TestShellWords drives the one splitter and the two readers built on it over
// the corpus: the split keeps quoted spans and GitHub expressions whole, the
// bare form drops the delimiting quotes, and a git push line resolves to the
// same operands.
func TestShellWords(t *testing.T) {
	t.Parallel()

	for _, tc := range loadShellWordsFixtures(t).Cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			words := shellWords(tc.Input)
			if !slices.Equal(words, tc.Words) {
				t.Fatalf("shellWords(%q) = %q, want %q", tc.Input, words, tc.Words)
			}
			bare := make([]string, 0, len(words))
			for _, w := range words {
				bare = append(bare, bareWord(w))
			}
			if !slices.Equal(bare, tc.Bare) {
				t.Fatalf("bare words of %q = %q, want %q", tc.Input, bare, tc.Bare)
			}
			if got := splitPushArgs(tc.Input); !slices.Equal(got, tc.PushRefspecs) {
				t.Fatalf("splitPushArgs(%q) = %q, want %q", tc.Input, got, tc.PushRefspecs)
			}
		})
	}
}
