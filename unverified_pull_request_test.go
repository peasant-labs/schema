package schema_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	tcassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/unverified_pull_requests.yaml
var unverifiedWireYAML []byte

type wireProbe struct {
	Kind string `yaml:"kind"`
	JSON string `yaml:"json"`
}
type wireExpected struct {
	Valid bool `yaml:"valid"`
}

func TestUnverifiedPullRequestWire(t *testing.T) {
	var f struct {
		RequiredNames []string                                 `yaml:"required_names"`
		Cases         []testcase.Case[wireProbe, wireExpected] `yaml:"cases"`
	}
	if err := yaml.Unmarshal(unverifiedWireYAML, &f); err != nil {
		t.Fatal(err)
	}
	corpus := testcase.Corpus[wireProbe, wireExpected]{Cases: f.Cases}
	tcassert.RequireValid(t, corpus)
	names := map[string]bool{}
	for _, c := range f.Cases {
		names[c.Name] = true
	}
	for _, n := range f.RequiredNames {
		if !names[n] {
			t.Errorf("required case %q is missing", n)
		}
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var v interface{ Validate() error }
			switch c.Input.Kind {
			case "local":
				v = &schema.SessionPullRequestsResponse{}
			case "village":
				v = &schema.VillageTranscriptPullRequestsResponse{}
			default:
				t.Fatalf("unknown kind %s", c.Input.Kind)
			}
			err := json.Unmarshal([]byte(c.Input.JSON), v)
			if err == nil {
				err = v.Validate()
			}
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%t error=%v", c.Expected.Valid, err)
			}
			if err == nil {
				encoded, e := json.Marshal(v)
				if e != nil {
					t.Fatal(e)
				}
				var before, after any
				if e = json.Unmarshal([]byte(c.Input.JSON), &before); e != nil {
					t.Fatal(e)
				}
				if e = json.Unmarshal(encoded, &after); e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(before, after) {
					t.Fatalf("wire changed: %s", encoded)
				}
			}
		})
	}
}

//go:embed testdata/unverified_pull_request_rows.yaml
var unverifiedRowsYAML []byte

type rowProbe struct {
	JSON string `yaml:"json"`
}

func TestUnverifiedPullRequestRows(t *testing.T) {
	var f struct {
		RequiredNames []string                                `yaml:"required_names"`
		Cases         []testcase.Case[rowProbe, wireExpected] `yaml:"cases"`
	}
	if err := yaml.Unmarshal(unverifiedRowsYAML, &f); err != nil {
		t.Fatal(err)
	}
	corpus := testcase.Corpus[rowProbe, wireExpected]{Cases: f.Cases}
	tcassert.RequireValid(t, corpus)
	names := map[string]bool{}
	for _, c := range f.Cases {
		names[c.Name] = true
	}
	for _, n := range f.RequiredNames {
		if !names[n] {
			t.Errorf("missing required row case %q", n)
		}
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var row schema.UnverifiedPullRequest
			err := json.Unmarshal([]byte(c.Input.JSON), &row)
			if err == nil {
				err = row.Validate()
			}
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%t, error=%v", c.Expected.Valid, err)
			}
		})
	}
}
