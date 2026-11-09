package schema_test

import (
	_ "embed"
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
	Page          int    `yaml:"page"`
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
			if err == nil && q.Home() {
				if int(q.EffectivePageSize()) != c.Expected.MaxItems {
					t.Fatalf("perPage=%d want %d", q.EffectivePageSize(), c.Expected.MaxItems)
				}
				if q.EffectivePage() != c.Expected.Page {
					t.Fatalf("page=%d want %d", q.EffectivePage(), c.Expected.Page)
				}
			}
		})
	}
}

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
		for _, name := range []string{"items", "nextItems", "page", "perPage", "total", "revision", "totals"} {
			if properties[name] == nil {
				t.Fatalf("%s page lacks %q: %v", document, name, properties)
			}
		}
		if len(properties) != 7 {
			t.Fatalf("%s page is not exactly the six offset fields plus totals: %v", document, properties)
		}
		items := properties["items"].(map[string]any)
		if items["type"] != "array" {
			t.Fatalf("%s items is not a non-null array", document)
		}
		ref := items["items"].(map[string]any)["$ref"].(string)
		if !strings.HasSuffix(ref, "/LocalSyncSessionRow") && !strings.HasSuffix(ref, "/SchemaLocalSyncSessionRow") && !strings.HasSuffix(ref, "/SchemaSyncSessionRow") {
			t.Fatalf("%s item ref %q", document, ref)
		}
		perPage := properties["perPage"].(map[string]any)
		if ref, ok := perPage["$ref"].(string); ok {
			perPage = components[strings.TrimPrefix(ref, "#/components/schemas/")].(map[string]any)
		}
		if perPage["type"] != "integer" || len(perPage["enum"].([]any)) != 3 {
			t.Fatalf("%s perPage is not the numeric 25/50/100 enum: %v", document, perPage)
		}
	}
	paths := h.documents["local"]["paths"].(map[string]any)
	if paths["/api/v1/home/sessions"] != nil {
		t.Fatal("unshipped offset Home route remains")
	}
}
