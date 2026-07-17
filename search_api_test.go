package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewSearchPayload_EmptyResultsMarshalsAsArray(t *testing.T) {
	b, err := json.Marshal(NewSearchPayload("photosynthesis"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, `"results":[]`) {
		t.Errorf("empty Results should marshal as [], got %s", got)
	}
	if strings.Contains(got, "null") {
		t.Errorf("payload contains null (nil slice?): %s", got)
	}
	if !strings.Contains(got, `"query":"photosynthesis"`) {
		t.Errorf("query not echoed: %s", got)
	}
}

func TestSearchResult_CamelCaseJSON(t *testing.T) {
	projectHash := ProjectHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	p := NewSearchPayload("q")
	p.Results = append(p.Results, SearchResult{
		SessionID:   "s1",
		Project:     "/repo",
		ProjectHash: projectHash,
		EntryIndex:  3,
		Role:        "user",
		Snippet:     "the [match] here",
		Score:       1.5,
	})
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	for _, field := range []string{
		`"sessionId":"s1"`, `"project":"/repo"`, `"projectHash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
		`"entryIndex":3`, `"role":"user"`, `"snippet":"the [match] here"`, `"score":1.5`,
	} {
		if !strings.Contains(got, field) {
			t.Errorf("missing %s in %s", field, got)
		}
	}
}

func TestSearchResult_ProjectHashOmittedWhenEmpty(t *testing.T) {
	b, _ := json.Marshal(SearchResult{SessionID: "s1", EntryIndex: 0, Role: "user"})
	if strings.Contains(string(b), "projectHash") {
		t.Errorf("empty projectHash should be omitted: %s", b)
	}
}
