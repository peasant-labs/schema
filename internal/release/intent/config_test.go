package intent_test

// Release-config loader suite: writes each document to a temp file and
// drives intent.LoadConfig over the config corpus. A missing file is covered
// directly: it must fail with all six parts.

import (
	_ "embed"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/config_corpus.yaml
var configCorpusYAML []byte

type configInput struct {
	Document string `yaml:"document"`
}

func TestLoadConfig(t *testing.T) {
	corpus := loadNamedCorpus[configInput, intent.Config](t, configCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "release-intent.yaml")
			if err := os.WriteFile(path, []byte(c.Input.Document), 0o600); err != nil {
				t.Fatalf("write temp config: %v", err)
			}
			got, err := intent.LoadConfig(path)
			if c.Classification == "must-fail" {
				if err == nil {
					t.Fatalf("LoadConfig: expected an error, got %+v", got)
				}
				requireSixLabels(t, err)
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, c.Expected) {
				t.Errorf("LoadConfig = %+v, want %+v", got, c.Expected)
			}
		})
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := intent.LoadConfig(filepath.Join(t.TempDir(), "no-such-file.yaml"))
	if err == nil {
		t.Fatalf("LoadConfig: expected an error for a missing file")
	}
	requireSixLabels(t, err)
}
