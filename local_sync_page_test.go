package schema_test

import (
	_ "embed"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/sync_page_queries.yaml
var syncPageQueriesYAML []byte

type syncPageQueryInput struct {
	Query      string `yaml:"query"`
	TokenBytes int    `yaml:"token_bytes"`
}

type syncPageQueryExpected struct {
	Valid         bool   `yaml:"valid"`
	ErrorContains string `yaml:"error_contains"`
	MaxItems      int    `yaml:"max_items"`
	Token         string `yaml:"token"`
}

func TestLocalSyncPageQueries(t *testing.T) {
	var fx struct {
		RequiredNames []string                                                   `yaml:"required_names"`
		Cases         []testcase.Case[syncPageQueryInput, syncPageQueryExpected] `yaml:"cases"`
	}
	if err := yaml.Unmarshal(syncPageQueriesYAML, &fx); err != nil {
		t.Fatal(err)
	}
	corpus := testcase.Corpus[syncPageQueryInput, syncPageQueryExpected]{Cases: fx.Cases}
	assertcase.RequireMin(t, corpus, 1)
	assertcase.RequireValid(t, corpus)
	names := map[string]bool{}
	for _, c := range fx.Cases {
		names[c.Name] = true
	}
	for _, name := range fx.RequiredNames {
		if !names[name] {
			t.Fatalf("missing required sync query case %s", name)
		}
		delete(names, name)
	}
	if len(names) != 0 || len(fx.RequiredNames) == 0 {
		t.Fatal("sync query required-name manifest differs from cases")
	}
	for _, c := range fx.Cases {
		t.Run(c.Name, func(t *testing.T) {
			values, err := url.ParseQuery(c.Input.Query)
			if err != nil {
				t.Fatal(err)
			}
			if c.Input.TokenBytes > 0 {
				values.Set("pageToken", strings.Repeat("x", c.Input.TokenBytes))
			}
			q, err := schema.ParseLocalSyncSessionsQuery(values)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v: %v", c.Expected.Valid, err)
			}
			if c.Expected.ErrorContains != "" && (err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains)) {
				t.Fatalf("want %q, got %v", c.Expected.ErrorContains, err)
			}
			if err == nil && q.View == "home" {
				request := q.PageRequest()
				if request.MaxItems != c.Expected.MaxItems {
					t.Fatalf("MaxItems=%d want %d", request.MaxItems, c.Expected.MaxItems)
				}
				if c.Input.TokenBytes == 0 && string(request.Token) != c.Expected.Token {
					t.Fatalf("token=%q want %q", request.Token, c.Expected.Token)
				}
			}
		})
	}
}

// Assignment in both directions proves the public alias is the real generic
// instantiation, not an independently declared lookalike response struct.
var _ schema.LocalSyncSessionPage = schema.Page[schema.LocalSyncSessionRow]{}
var _ schema.Page[schema.LocalSyncSessionRow] = schema.LocalSyncSessionPage{}

func TestSyncPageConcreteCatalogIdentity(t *testing.T) {
	h := newBoundaryHarness(t)
	for _, document := range []string{"types", "local"} {
		components := h.documents[document]["components"].(map[string]any)["schemas"].(map[string]any)
		for name := range components {
			if strings.ContainsAny(name, "[]") || strings.Contains(name, "PageSchema") || strings.Contains(name, "LocalHome") {
				t.Fatalf("%s leaks unstable or superseded name %q", document, name)
			}
		}
		page, ok := components["LocalSyncSessionPage"].(map[string]any)
		if !ok {
			t.Fatalf("%s lacks concrete page", document)
		}
		properties := page["properties"].(map[string]any)
		if len(properties) != 2 || properties["items"] == nil || properties["next"] == nil {
			t.Fatalf("%s page is not exactly items/next: %v", document, properties)
		}
		items := properties["items"].(map[string]any)
		if items["type"] != "array" {
			t.Fatalf("%s items is not a non-null array", document)
		}
		ref := items["items"].(map[string]any)["$ref"].(string)
		if !strings.HasSuffix(ref, "/LocalSyncSessionRow") && !strings.HasSuffix(ref, "/SchemaLocalSyncSessionRow") {
			t.Fatalf("%s item ref %q", document, ref)
		}
		token := h.documents["types"]["components"].(map[string]any)["schemas"].(map[string]any)["PageToken"]
		want, _ := json.Marshal(token)
		got, _ := json.Marshal(properties["next"])
		if string(got) != string(want) {
			t.Fatalf("%s next scalar differs from PageToken: %s != %s", document, got, want)
		}
	}
	paths := h.documents["local"]["paths"].(map[string]any)
	if paths["/api/v1/home/sessions"] != nil {
		t.Fatal("unshipped offset Home route remains")
	}
}
