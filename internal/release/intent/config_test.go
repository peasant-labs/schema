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

type configExpected struct {
	intent.Config `yaml:",inline"`
	Code          string `yaml:"code"`
}

func TestLoadConfig(t *testing.T) {
	corpus := loadNamedCorpus[configInput, configExpected](t, configCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "release-intent.yaml")
			if err := os.WriteFile(path, []byte(c.Input.Document), 0o600); err != nil {
				t.Fatalf("write temp config: %v", err)
			}
			got, err := intent.LoadConfig(path)
			// Bytes-vs-path parity: the same document through ParseConfig must
			// reach the same verdict, since LoadConfig reads the file and
			// calls ParseConfig.
			parsed, parseErr := intent.ParseConfig([]byte(c.Input.Document))
			if (err == nil) != (parseErr == nil) {
				t.Fatalf("LoadConfig error = %v, ParseConfig error = %v: entry points disagree", err, parseErr)
			}
			if c.Classification == "must-fail" {
				if err == nil {
					t.Fatalf("LoadConfig: expected an error, got %+v", got)
				}
				requireSixParts(t, err, c.Expected.Code)
				if parseErr == nil {
					t.Fatalf("ParseConfig: expected an error, got %+v", parsed)
				}
				requireSixParts(t, parseErr, c.Expected.Code)
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig: unexpected error: %v", err)
			}
			if parseErr != nil {
				t.Fatalf("ParseConfig: unexpected error: %v", parseErr)
			}
			if !reflect.DeepEqual(parsed, c.Expected.Config) {
				t.Errorf("ParseConfig = %+v, want %+v", parsed, c.Expected.Config)
			}
			if !reflect.DeepEqual(got, c.Expected.Config) {
				t.Errorf("LoadConfig = %+v, want %+v", got, c.Expected.Config)
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
