package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/authoritative_publish_metadata_limits.yaml
var authoritativePublishMetadataLimitsYAML []byte

type metadataLimitInput struct {
	RequestCase   string `yaml:"request_case"`
	DocumentBytes int    `yaml:"document_bytes"`
	EntryRecipe   string `yaml:"entry_recipe"`
	EntryCount    int    `yaml:"entry_count"`
}

type metadataLimitExpected struct {
	Accepted             bool   `yaml:"accepted"`
	ErrorContains        string `yaml:"error_contains"`
	MinimumDocumentBytes int    `yaml:"minimum_document_bytes"`
	MaximumDocumentBytes int    `yaml:"maximum_document_bytes"`
}

func loadAuthoritativePublishMetadataLimits(data []byte) (testcase.Corpus[metadataLimitInput, metadataLimitExpected], error) {
	return testcase.LoadCorpus[metadataLimitInput, metadataLimitExpected](data)
}

func TestAuthoritativePublishMetadataLimits(t *testing.T) {
	corpus, err := loadAuthoritativePublishMetadataLimits(authoritativePublishMetadataLimitsYAML)
	if err != nil {
		t.Fatal(err)
	}
	assert.RequireMin(t, corpus, 1)
	assert.RequireValid(t, corpus)
	requiredNames := []string{
		"one byte below metadata cap",
		"exact metadata cap",
		"one byte above metadata cap",
		"past former metadata cap",
		"many entry identities survive decoding",
	}
	names := make(map[string]bool)
	for _, row := range corpus.Cases {
		names[row.Name] = true
	}
	for _, name := range requiredNames {
		if !names[name] {
			t.Fatalf("authoritative metadata corpus missing required case %q", name)
		}
	}
	publication := loadPublicationCorpus(t)
	for _, row := range corpus.Cases {
		t.Run(row.Name, func(t *testing.T) {
			var base string
			for _, recipe := range publication.ParentIdentity.Cases {
				if recipe.Name == row.Input.RequestCase {
					base = recipe.Input
				}
			}
			if base == "" {
				t.Fatalf("missing publication request recipe %q", row.Input.RequestCase)
			}
			request, err := schema.DecodeAuthoritativePublishRequest([]byte(base))
			if err != nil {
				t.Fatalf("decode publication request recipe: %v", err)
			}
			if row.Input.EntryCount > 0 {
				var entries []schema.AuthoritativeSessionEntry
				for _, recipe := range publication.FingerprintMutations.Cases {
					if recipe.Name == row.Input.EntryRecipe && recipe.Path == "replacement.entries" {
						if err := json.Unmarshal([]byte(recipe.Value), &entries); err != nil {
							t.Fatalf("decode publication entry recipe: %v", err)
						}
					}
				}
				if len(entries) != 1 {
					t.Fatalf("publication entry recipe %q must supply one template", row.Input.EntryRecipe)
				}
				request.Entries = make([]schema.AuthoritativeSessionEntry, row.Input.EntryCount)
				for i := range request.Entries {
					request.Entries[i] = entries[0]
					request.Entries[i].EntryIndex = i
				}
			}
			raw, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			// The publication marshaler materializes canonical defaults. Compare
			// against that encoded request, independently of the raw size wrapper.
			want, err := schema.DecodeAuthoritativePublishRequest(raw)
			if err != nil {
				t.Fatalf("decode compact publication recipe: %v", err)
			}
			if row.Input.DocumentBytes > 0 {
				if len(raw) > row.Input.DocumentBytes {
					t.Fatal("request recipe exceeds whitespace padding target")
				}
				raw = append(raw, bytes.Repeat([]byte(" "), row.Input.DocumentBytes-len(raw))...)
				if len(raw) != row.Input.DocumentBytes {
					t.Fatalf("encoded length=%d want %d", len(raw), row.Input.DocumentBytes)
				}
			} else if len(raw) < row.Expected.MinimumDocumentBytes || len(raw) > row.Expected.MaximumDocumentBytes {
				t.Fatalf("many-entry encoded length=%d outside fixture bounds [%d, %d]", len(raw), row.Expected.MinimumDocumentBytes, row.Expected.MaximumDocumentBytes)
			}
			t.Logf("encoded document bytes=%d entries=%d", len(raw), len(request.Entries))
			decoded, err := schema.DecodeAuthoritativePublishMetadataRaw(raw)
			if (err == nil) != row.Expected.Accepted {
				t.Fatalf("accepted=%v want %v: %v", err == nil, row.Expected.Accepted, err)
			}
			if err != nil {
				if row.Expected.ErrorContains == "" || !strings.Contains(err.Error(), row.Expected.ErrorContains) {
					t.Fatalf("refusal=%v want %q", err, row.Expected.ErrorContains)
				}
				return
			}
			if !reflect.DeepEqual(decoded, want) {
				t.Fatal("raw decoder changed the complete authoritative request or entry identities")
			}
			if len(decoded.Entries) != len(request.Entries) {
				t.Fatalf("decoded entry count=%d want %d", len(decoded.Entries), len(request.Entries))
			}
			for i, entry := range decoded.Entries {
				if entry.EntryIndex != i || entry.SessionID != request.Entries[i].SessionID {
					t.Fatalf("decoded entry %d lost its identity", i)
				}
			}
		})
	}
}
