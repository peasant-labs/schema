package intent_test

// PR classifier suite: drives intent.ClassifyPR over the classifier corpus.

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/classifier_corpus.yaml
var classifierCorpusYAML []byte

type classifierInput struct {
	Author           string `yaml:"author"`
	HeadBranch       string `yaml:"head_branch"`
	AutomationAuthor string `yaml:"automation_author"`
	ReleaseBranch    string `yaml:"release_branch"`
	VendorBranch     string `yaml:"vendor_branch"`
	CutBranch        string `yaml:"cut_branch"`
}

type classifierExpected struct {
	Kind string `yaml:"kind"`
}

func TestClassifyPR(t *testing.T) {
	corpus := loadNamedCorpus[classifierInput, classifierExpected](t, classifierCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			cfg := intent.Config{
				AutomationAuthor: c.Input.AutomationAuthor,
				BotBranches: map[string]string{
					"release":     c.Input.ReleaseBranch,
					"vendor-hash": c.Input.VendorBranch,
					"cut-release": c.Input.CutBranch,
				},
			}
			got := intent.ClassifyPR(cfg, c.Input.Author, c.Input.HeadBranch)
			if string(got) != c.Expected.Kind {
				t.Errorf("ClassifyPR(%q, %q) = %q, want %q",
					c.Input.Author, c.Input.HeadBranch, got, c.Expected.Kind)
			}
		})
	}
}
