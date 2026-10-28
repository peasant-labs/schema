package commandpr_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/commandpr"
	"github.com/peasant-labs/schema/testcase"
	tcassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/detection.yaml
var detectionYAML []byte

//go:embed testdata/required-names.yaml
var requiredYAML []byte

//go:embed testdata/session-cases.yaml
var sessionYAML []byte

type call struct {
	Arguments string `yaml:"arguments"`
	Result    string `yaml:"result"`
	IsError   bool   `yaml:"isError"`
}
type callInput struct {
	Calls []call `yaml:"calls"`
}
type row struct {
	Owner  string `yaml:"owner"`
	Repo   string `yaml:"repo"`
	Number int    `yaml:"number"`
	URL    string `yaml:"url"`
}

func TestRecordedCommands(t *testing.T) {
	corpus, err := testcase.LoadCorpus[callInput, []row](detectionYAML)
	if err != nil {
		t.Fatal(err)
	}
	tcassert.RequireValid(t, corpus)
	var manifest struct {
		Names []string `yaml:"names"`
	}
	if err := yaml.Unmarshal(requiredYAML, &manifest); err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, c := range corpus.Cases {
		names[c.Name] = true
	}
	for _, n := range manifest.Names {
		if !names[n] {
			t.Errorf("missing required case %q", n)
		}
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			calls := make([]schema.ToolCallDetail, len(c.Input.Calls))
			for i, v := range c.Input.Calls {
				calls[i] = schema.ToolCallDetail{Arguments: v.Arguments, Result: v.Result, IsError: v.IsError}
			}
			got := commandpr.DetectSession(schema.SessionDetailPayload{Turns: []schema.TurnDetail{{ToolCalls: calls}}})
			rows := make([]row, 0, len(got))
			for _, p := range got {
				rows = append(rows, row{p.Owner, p.Repo, p.Number, p.URL})
			}
			if len(rows) == 0 && len(c.Expected) == 0 {
				return
			}
			if !reflect.DeepEqual(rows, c.Expected) {
				t.Fatalf("got %#v want %#v", rows, c.Expected)
			}
		})
	}
}

type sessionInput struct {
	Detail string `yaml:"detail"`
}
type sessionExpected struct {
	Rows string `yaml:"rows"`
}

func TestSessionProvenance(t *testing.T) {
	var required struct {
		RequiredNames []string `yaml:"required_names"`
	}
	if err := yaml.Unmarshal(sessionYAML, &required); err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(sessionYAML, &document); err != nil {
		t.Fatal(err)
	}
	delete(document, "required_names")
	data, err := yaml.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := testcase.LoadCorpus[sessionInput, sessionExpected](data)
	if err != nil {
		t.Fatal(err)
	}
	tcassert.RequireValid(t, corpus)
	names := map[string]bool{}
	for _, c := range corpus.Cases {
		names[c.Name] = true
	}
	for _, n := range required.RequiredNames {
		if !names[n] {
			t.Errorf("missing required case %q", n)
		}
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var detail schema.SessionDetailPayload
			var expected []schema.UnverifiedPullRequest
			if err := json.Unmarshal([]byte(c.Input.Detail), &detail); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(c.Expected.Rows), &expected); err != nil {
				t.Fatal(err)
			}
			got := commandpr.DetectSession(detail)
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("got %#v want %#v", got, expected)
			}
		})
	}
}
