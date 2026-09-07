package schema_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/metadata/adapter_versions.yaml
var adapterVersionCases []byte

//go:embed testdata/metadata/adapter_versions_manifest.yaml
var adapterVersionManifest []byte

type adapterVersionExpected struct {
	Accepted       bool   `yaml:"accepted"`
	SchemaVersion  int    `yaml:"schemaVersion"`
	AdapterVersion *int   `yaml:"adapterVersion"`
	LegacyHash     string `yaml:"legacyHash"`
}

func loadAdapterVersionFixtures(t *testing.T) (testcase.Corpus[string, adapterVersionExpected], string) {
	t.Helper()
	corpus, err := testcase.LoadCorpus[string, adapterVersionExpected](adapterVersionCases)
	if err != nil {
		t.Fatalf("load metadata adapter version fixtures: %v", err)
	}
	var manifest struct {
		Names        []string `yaml:"requiredCaseNames"`
		BaseMetadata string   `yaml:"baseMetadata"`
	}
	if err := yaml.Unmarshal(adapterVersionManifest, &manifest); err != nil {
		t.Fatalf("load adapter version required-name manifest: %v", err)
	}
	names := make(map[string]bool)
	for _, c := range corpus.Cases {
		names[c.Name] = true
	}
	for _, name := range manifest.Names {
		if !names[name] {
			t.Fatalf("adapter version fixtures missing required case %q", name)
		}
		delete(names, name)
	}
	if len(names) != 0 || len(manifest.Names) == 0 {
		t.Fatalf("adapter version fixture manifest does not match case names: %v", names)
	}
	return corpus, manifest.BaseMetadata
}

func TestUnifiedMetadataAdapterVersionJSON(t *testing.T) {
	corpus, base := loadAdapterVersionFixtures(t)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var meta schema.UnifiedMetadata
			if err := json.Unmarshal([]byte(base), &meta); err != nil {
				t.Fatalf("decode baseline metadata: %v", err)
			}
			err := json.Unmarshal([]byte(c.Input), &meta)
			if (err == nil) != c.Expected.Accepted {
				t.Fatalf("metadata decode acceptance = %v, want %v: %v", err == nil, c.Expected.Accepted, err)
			}
			if !c.Expected.Accepted {
				if !strings.Contains(err.Error(), "adapterVersion") {
					t.Fatalf("metadata error must name adapterVersion: %v", err)
				}
				return
			}
			if meta.SchemaVersion != c.Expected.SchemaVersion || !reflect.DeepEqual(meta.AdapterVersion, c.Expected.AdapterVersion) {
				t.Fatalf("decoded metadata version/provenance = %d/%v, want %d/%v", meta.SchemaVersion, meta.AdapterVersion, c.Expected.SchemaVersion, c.Expected.AdapterVersion)
			}
			encoded, err := json.Marshal(meta)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), `"adapterVersion"`) != (c.Expected.AdapterVersion != nil) {
				t.Fatalf("producer omission changed on encode: %s", encoded)
			}
			var roundtrip schema.UnifiedMetadata
			if err := json.Unmarshal(encoded, &roundtrip); err != nil || !reflect.DeepEqual(roundtrip.AdapterVersion, meta.AdapterVersion) {
				t.Fatalf("producer revision changed on round-trip: %v", err)
			}
		})
	}
}

func TestUnifiedMetadataAdapterVersionHash(t *testing.T) {
	base := schema.NewUnifiedMetadata()
	base.SchemaVersion = 9
	base.SessionID = "99d59925-36bc-424c-a789-8be54d9702ba"
	corpus, _ := loadAdapterVersionFixtures(t)
	for _, c := range corpus.Cases {
		if !c.Expected.Accepted {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			meta := base
			meta.AdapterVersion = c.Expected.AdapterVersion
			changed := schema.ComputeMetadataHash(&meta) != schema.ComputeMetadataHash(&base)
			if changed != (meta.AdapterVersion != nil) {
				t.Fatalf("adapter producer hash change = %v, want %v", changed, meta.AdapterVersion != nil)
			}
			if c.Expected.LegacyHash != "" && schema.ComputeMetadataHash(&meta) != c.Expected.LegacyHash {
				t.Fatalf("omitted producer changed historic metadata hash: got %s, want %s", schema.ComputeMetadataHash(&meta), c.Expected.LegacyHash)
			}
		})
	}
}

func TestUnifiedMetadataAdapterVersionInvalidWrites(t *testing.T) {
	corpus, _ := loadAdapterVersionFixtures(t)
	for _, c := range corpus.Cases {
		var input struct {
			AdapterVersion *int `json:"adapterVersion"`
		}
		if json.Unmarshal([]byte(c.Input), &input) != nil || input.AdapterVersion == nil || *input.AdapterVersion > 0 {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			meta := schema.NewUnifiedMetadata()
			meta.AdapterVersion = input.AdapterVersion
			if _, err := json.Marshal(meta); err == nil || !strings.Contains(err.Error(), "adapterVersion") {
				t.Fatalf("invalid producer must fail serialization with a field diagnostic: %v", err)
			}
			if hash := schema.ComputeMetadataHash(&meta); hash != "" {
				t.Fatalf("invalid metadata must not produce a success hash: %s", hash)
			}
		})
	}
}
