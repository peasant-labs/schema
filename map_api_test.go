package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewReviewListPayload_TimelineCollectionsMarshalAsArrays(t *testing.T) {
	b, err := json.Marshal(NewReviewListPayload("project"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, `"recentCommits":[]`) || !strings.Contains(got, `"sessions":[]`) {
		t.Errorf("timeline collections should marshal as arrays, got %s", got)
	}
}

func TestNewCommitRef_SessionIDsMarshalAsArray(t *testing.T) {
	b, err := json.Marshal(NewCommitRef("hash", "subject"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(b); !strings.Contains(got, `"sessionIds":[]`) {
		t.Errorf("sessionIds should marshal as an array, got %s", got)
	}
}

func TestNewChangeDetailPayload_FrictionsMarshalsAsArray(t *testing.T) {
	b, err := json.Marshal(NewChangeDetailPayload("feat/x"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, `"frictions":[]`) {
		t.Errorf("empty Frictions should marshal as [], got %s", got)
	}
	if strings.Contains(got, "null") {
		t.Errorf("payload contains null (nil slice?): %s", got)
	}
}

func TestNewMapNodeDetailPayload_ConnectionsMarshalAsArrays(t *testing.T) {
	b, err := json.Marshal(NewMapNodeDetailPayload("internal/api"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, `"dependsOn":[]`) || !strings.Contains(got, `"usedBy":[]`) {
		t.Errorf("dependsOn/usedBy should marshal as [], got %s", got)
	}
	if strings.Contains(got, "null") {
		t.Errorf("payload contains null (nil slice?): %s", got)
	}
}

func TestFrictionCluster_CamelCaseJSON(t *testing.T) {
	b, err := json.Marshal(FrictionCluster{
		Kind:     "retryLoop",
		Label:    "retry loops",
		File:     "internal/api/server.go",
		Count:    3,
		Sessions: 2,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	for _, field := range []string{
		`"kind":"retryLoop"`, `"label":"retry loops"`,
		`"file":"internal/api/server.go"`, `"count":3`, `"sessions":2`,
	} {
		if !strings.Contains(got, field) {
			t.Errorf("missing %s in %s", field, got)
		}
	}
}
