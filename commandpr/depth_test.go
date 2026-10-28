package commandpr

import (
	"encoding/json"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// The named shell examples prove the recursion budget, not merely a parse error.
func TestWrapperDepthBound(t *testing.T) {
	data, err := os.ReadFile("testdata/detection.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name  string `yaml:"name"`
			Input struct {
				Calls []struct {
					Arguments string `yaml:"arguments"`
				} `yaml:"calls"`
			} `yaml:"input"`
		} `yaml:"cases"`
	}
	if err = yaml.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	lookup := func(name string) string {
		t.Helper()
		for _, c := range fixture.Cases {
			if c.Name == name {
				if len(c.Input.Calls) != 1 {
					t.Fatalf("%s needs one call", name)
				}
				var args map[string]string
				if err := json.Unmarshal([]byte(c.Input.Calls[0].Arguments), &args); err != nil {
					t.Fatal(err)
				}
				return args["command"]
			}
		}
		t.Fatalf("missing %s", name)
		return ""
	}
	four, five := lookup("bash -c nested four levels"), lookup("bash -c nested five levels")
	if !commandInvokesGHPRCreate(four, maxWrapperDepth) || commandInvokesGHPRCreate(five, maxWrapperDepth) || !commandInvokesGHPRCreate(five, maxWrapperDepth+1) {
		t.Fatal("shell wrapper recursion budget does not match the named examples")
	}
}
