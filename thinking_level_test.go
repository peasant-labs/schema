package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/metadata/thinking_level.yaml
var thinkingLevelMetadataYAML []byte

//go:embed testdata/metadata/thinking_level_manifest.yaml
var thinkingLevelMetadataManifestYAML []byte

//go:embed testdata/metadata/thinking_level_raw.yaml
var thinkingLevelRawYAML []byte

//go:embed testdata/metadata/thinking_level_raw_manifest.yaml
var thinkingLevelRawManifestYAML []byte

type thinkingLevelMetadataInput struct {
	Operation string `yaml:"operation"`
	JSON      string `yaml:"json,omitempty"`
	Level     string `yaml:"level,omitempty"`
	Raw       string `yaml:"raw,omitempty"`
	RawBase64 string `yaml:"rawBase64,omitempty"`
}

type thinkingLevelMetadataExpected struct {
	Accepted         bool   `yaml:"accepted"`
	SchemaVersion    int    `yaml:"schemaVersion,omitempty"`
	ThinkingLevel    string `yaml:"thinkingLevel,omitempty"`
	ThinkingLevelRaw string `yaml:"thinkingLevelRaw,omitempty"`
	ErrorContains    string `yaml:"errorContains,omitempty"`
}

type thinkingLevelMetadataManifest struct {
	BaseMetadata      string   `yaml:"baseMetadata"`
	ExpectedCaseCount int      `yaml:"expectedCaseCount"`
	RequiredCaseNames []string `yaml:"requiredCaseNames"`
}

type thinkingLevelMetadataCorpus struct {
	BaseMetadata string                                                                       `yaml:"baseMetadata"`
	Cases        []testcase.Case[thinkingLevelMetadataInput, thinkingLevelMetadataExpected] `yaml:"cases"`
}

func loadThinkingLevelMetadataFixtures(t *testing.T) (testcase.Corpus[thinkingLevelMetadataInput, thinkingLevelMetadataExpected], string) {
	t.Helper()
	var doc thinkingLevelMetadataCorpus
	decoder := yaml.NewDecoder(bytes.NewReader(thinkingLevelMetadataYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&doc); err != nil {
		t.Fatalf("decode thinking level metadata corpus: %v", err)
	}
	corpus := testcase.Corpus[thinkingLevelMetadataInput, thinkingLevelMetadataExpected]{Cases: doc.Cases}
	assert.RequireValid(t, corpus)
	manifest, err := decodeTurnModelFixtureManifest(thinkingLevelMetadataManifestYAML)
	if err != nil {
		t.Fatalf("load thinking level metadata manifest: %v", err)
	}
	if err := validateCorpusInventory("thinking level metadata", corpus, manifest); err != nil {
		t.Fatalf("validate thinking level metadata inventory: %v", err)
	}
	if strings.TrimSpace(doc.BaseMetadata) == "" {
		t.Fatal("thinking level metadata corpus must carry baseMetadata (a v11 sidecar)")
	}
	return corpus, doc.BaseMetadata
}

func TestUnifiedMetadataThinkingLevelFixture(t *testing.T) {
	corpus, base := loadThinkingLevelMetadataFixtures(t)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var meta schema.UnifiedMetadata
			var err error
			switch c.Input.Operation {
			case "decode":
				if err := json.Unmarshal([]byte(base), &meta); err != nil {
					t.Fatalf("decode v11 baseline metadata: %v", err)
				}
				err = json.Unmarshal([]byte(c.Input.JSON), &meta)
			case "marshal":
				raw := c.Input.Raw
				if c.Input.RawBase64 != "" {
					b, decodeErr := base64.StdEncoding.DecodeString(c.Input.RawBase64)
					if decodeErr != nil {
						t.Fatalf("fixture rawBase64: %v", decodeErr)
					}
					raw = string(b)
				}
				meta = schema.NewUnifiedMetadata()
				meta.ThinkingLevel = schema.ThinkingLevel(c.Input.Level)
				meta.ThinkingLevelRaw = schema.ThinkingLevelRaw(raw)
				var encoded []byte
				encoded, err = json.Marshal(meta)
				if err == nil {
					meta = schema.UnifiedMetadata{}
					if err := json.Unmarshal(encoded, &meta); err != nil {
						t.Fatalf("re-decode marshaled metadata: %v", err)
					}
				}
			default:
				t.Fatalf("unknown fixture operation %q", c.Input.Operation)
			}
			if (err == nil) != c.Expected.Accepted {
				t.Fatalf("accepted=%v, want %v: %v", err == nil, c.Expected.Accepted, err)
			}
			if err != nil {
				if !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("error %q does not name %q", err, c.Expected.ErrorContains)
				}
				return
			}
			if meta.SchemaVersion != c.Expected.SchemaVersion {
				t.Fatalf("schemaVersion=%d, want %d (reading must not rewrite the declared version)", meta.SchemaVersion, c.Expected.SchemaVersion)
			}
			if string(meta.ThinkingLevel) != c.Expected.ThinkingLevel || string(meta.ThinkingLevelRaw) != c.Expected.ThinkingLevelRaw {
				t.Fatalf("level/raw=%q/%q, want %q/%q", meta.ThinkingLevel, meta.ThinkingLevelRaw, c.Expected.ThinkingLevel, c.Expected.ThinkingLevelRaw)
			}
			encoded, err := json.Marshal(meta)
			if err != nil {
				t.Fatalf("re-encode accepted metadata: %v", err)
			}
			var object map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &object); err != nil {
				t.Fatal(err)
			}
			for key, want := range map[string]string{"thinkingLevel": c.Expected.ThinkingLevel, "thinkingLevelRaw": c.Expected.ThinkingLevelRaw} {
				value, present := object[key]
				if present != (want != "") {
					t.Fatalf("%s present=%v on encode, want %v (absence must stay absent): %s", key, present, want != "", encoded)
				}
				if present {
					var got string
					if err := json.Unmarshal(value, &got); err != nil || got != want {
						t.Fatalf("%s bytes=%q, want %q", key, got, want)
					}
				}
			}
		})
	}
}

type thinkingLevelRawInput struct {
	Value  *string `yaml:"value,omitempty"`
	Base64 string  `yaml:"base64,omitempty"`
}

type thinkingLevelRawExpected struct {
	Accepted bool `yaml:"accepted"`
}

func loadThinkingLevelRawFixtures(t *testing.T) testcase.Corpus[thinkingLevelRawInput, thinkingLevelRawExpected] {
	t.Helper()
	corpus, err := testcase.LoadCorpus[thinkingLevelRawInput, thinkingLevelRawExpected](thinkingLevelRawYAML)
	if err != nil {
		t.Fatalf("load thinking level raw corpus: %v", err)
	}
	assert.RequireValid(t, corpus)
	manifest, err := decodeTurnModelFixtureManifest(thinkingLevelRawManifestYAML)
	if err != nil {
		t.Fatalf("load thinking level raw manifest: %v", err)
	}
	if err := validateCorpusInventory("thinking level raw", corpus, manifest); err != nil {
		t.Fatalf("validate thinking level raw inventory: %v", err)
	}
	return corpus
}

func thinkingLevelRawValue(t *testing.T, in thinkingLevelRawInput) string {
	t.Helper()
	if in.Base64 != "" {
		b, err := base64.StdEncoding.DecodeString(in.Base64)
		if err != nil {
			t.Fatalf("fixture base64: %v", err)
		}
		return string(b)
	}
	if in.Value == nil {
		t.Fatal("raw fixture must carry value or base64")
	}
	return *in.Value
}

func compileThinkingLevelRawSchema(t *testing.T, compiler *jsonschema.Compiler) *jsonschema.Schema {
	t.Helper()
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := compiler.AddResource("types.json", bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile("types.json#/components/schemas/ThinkingLevelRaw")
	if err != nil {
		t.Fatalf("compile ThinkingLevelRaw component: %v", err)
	}
	return compiled
}

// TestThinkingLevelRawJSONSchemaFormatEnforced compiles the published
// ThinkingLevelRaw component through the production compiler and proves that
// it agrees with NewThinkingLevelRaw/IsValid on every corpus row.
func TestThinkingLevelRawJSONSchemaFormatEnforced(t *testing.T) {
	compiled := compileThinkingLevelRawSchema(t, schema.NewJSONSchemaCompiler())
	for _, c := range loadThinkingLevelRawFixtures(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			value := thinkingLevelRawValue(t, c.Input)
			schemaErr := compiled.Validate(value)
			if (schemaErr == nil) != c.Expected.Accepted {
				t.Fatalf("compiled schema accepted=%v, want %v: %v", schemaErr == nil, c.Expected.Accepted, schemaErr)
			}
			constructed, err := schema.NewThinkingLevelRaw(value)
			if (err == nil) != c.Expected.Accepted || schema.ThinkingLevelRaw(value).IsValid() != c.Expected.Accepted {
				t.Fatalf("constructor/IsValid disagree with the compiled schema: err=%v", err)
			}
			if err == nil && constructed.String() != value {
				t.Fatalf("constructor changed accepted bytes: %q -> %q", value, constructed)
			}
			if err != nil && !strings.Contains(err.Error(), "schema.NewThinkingLevelRaw") {
				t.Fatalf("constructor error does not name its location: %v", err)
			}
		})
	}
}

// TestThinkingLevelRawByteFormatIsLoadBearing proves the byte rows fail only
// because the registered format asserts bytes: a compiler without the
// registration accepts them, since maxLength counts code points.
func TestThinkingLevelRawByteFormatIsLoadBearing(t *testing.T) {
	bare := jsonschema.NewCompiler()
	bare.AssertFormat = true
	compiled := compileThinkingLevelRawSchema(t, bare)
	corpus := loadThinkingLevelRawFixtures(t)
	checked := 0
	for _, name := range []string{"multibyte-129-bytes-rejected", "four-byte-64-code-points-rejected"} {
		for _, c := range corpus.Cases {
			if c.Name != name {
				continue
			}
			checked++
			if err := compiled.Validate(thinkingLevelRawValue(t, c.Input)); err != nil {
				t.Fatalf("%s: rejected without the byte format, so the corpus row does not exercise it: %v", name, err)
			}
		}
	}
	if checked != 2 {
		t.Fatalf("byte-format rows missing from the corpus: checked %d of 2", checked)
	}
}
