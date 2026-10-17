package schema_test

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/home_queries.yaml
var homeQueriesYAML []byte

type homeQueryInput struct {
	Kind string `yaml:"kind"`
	JSON string `yaml:"json"`
}

type homeQueryExpected struct {
	Valid         bool   `yaml:"valid"`
	ErrorContains string `yaml:"error_contains"`
}

func TestLocalHomeQueries(t *testing.T) {
	var fx struct {
		RequiredNames []string                                           `yaml:"required_names"`
		Cases         []testcase.Case[homeQueryInput, homeQueryExpected] `yaml:"cases"`
	}
	if err := yaml.Unmarshal(homeQueriesYAML, &fx); err != nil {
		t.Fatal(err)
	}
	corpus := testcase.Corpus[homeQueryInput, homeQueryExpected]{Cases: fx.Cases}
	assertcase.RequireMin(t, corpus, 1)
	assertcase.RequireValid(t, corpus)
	names := map[string]bool{}
	for _, c := range fx.Cases {
		names[c.Name] = true
	}
	for _, name := range fx.RequiredNames {
		if !names[name] {
			t.Fatalf("missing required Home query case %s", name)
		}
		delete(names, name)
	}
	if len(names) != 0 || len(fx.RequiredNames) == 0 {
		t.Fatal("Home query required-name manifest differs from cases")
	}
	for _, c := range fx.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var err error
			switch c.Input.Kind {
			case "sessions":
				var q schema.LocalHomeSessionsQuery
				err = json.Unmarshal([]byte(c.Input.JSON), &q)
				if err == nil {
					err = q.Validate()
				}
			case "stats":
				var q schema.LocalHomeStatsQuery
				err = json.Unmarshal([]byte(c.Input.JSON), &q)
				if err == nil {
					err = q.Validate()
				}
			default:
				t.Fatalf("unknown query kind %q", c.Input.Kind)
			}
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v: %v", c.Expected.Valid, err)
			}
			if c.Expected.ErrorContains != "" && (err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains)) {
				t.Fatalf("want error containing %q, got %v", c.Expected.ErrorContains, err)
			}
		})
	}
}
