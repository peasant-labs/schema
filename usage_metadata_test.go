package schema_test

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	schema "github.com/peasant-labs/schema"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/contract/usage_metadata_raw.yaml
var usageMetadataFS embed.FS

type usageMetadataFixture struct {
	RequiredCaseNames []string `yaml:"requiredCaseNames"`
	Cases             []struct {
		Name           string `yaml:"name"`
		Operation      string `yaml:"operation"`
		Completeness   string `yaml:"completeness"`
		CostTotal      string `yaml:"costTotal"`
		ErrorContains  string `yaml:"errorContains"`
		Raw            string `yaml:"raw"`
		Accepted       bool   `yaml:"accepted"`
		DuplicateOwner bool   `yaml:"duplicateOwner"`
		WrongToolScope bool   `yaml:"wrongToolScope"`
		Tokens         struct {
			Input       *int64 `yaml:"input"`
			Output      *int64 `yaml:"output"`
			CacheRead   *int64 `yaml:"cacheRead"`
			CacheWrite  *int64 `yaml:"cacheWrite"`
			Reasoning   *int64 `yaml:"reasoning"`
			TotalTokens *int64 `yaml:"totalTokens"`
		} `yaml:"tokens"`
		Capabilities []schema.ContentCapability `yaml:"capabilities"`
	} `yaml:"cases"`
}

func TestUsageMetadataRawContract(t *testing.T) {
	raw, err := usageMetadataFS.ReadFile("testdata/contract/usage_metadata_raw.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var f usageMetadataFixture
	var document yaml.Node
	if err := yaml.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	rejectDuplicateYAMLKeys(t, &document)
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err = decoder.Decode(&f); err != nil {
		t.Fatal(err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		t.Fatalf("fixture must contain one YAML document: %v", err)
	}
	seen := map[string]bool{}
	for _, c := range f.Cases {
		c := c
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("duplicate fixture case name %q", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			var got error
			switch c.Operation {
			case "usage":
				u := schema.UsageDetail{OwnerID: "owner", SourceEntryRef: "entry", Scope: schema.UsageScopeAssistant, Completeness: schema.UsageCompleteness(c.Completeness)}
				if c.Tokens.Input != nil || c.Tokens.Output != nil || c.Tokens.CacheRead != nil || c.Tokens.CacheWrite != nil || c.Tokens.Reasoning != nil || c.Tokens.TotalTokens != nil {
					u.Tokens = &schema.TokenUsageDetail{Input: c.Tokens.Input, Output: c.Tokens.Output, CacheRead: c.Tokens.CacheRead, CacheWrite: c.Tokens.CacheWrite, Reasoning: c.Tokens.Reasoning, TotalTokens: c.Tokens.TotalTokens}
				}
				if c.CostTotal != "" {
					v := schema.RecordedCostAmount(c.CostTotal)
					u.Cost = &schema.RecordedCostDetail{Total: &v, Source: schema.RecordedCostSourceHarnessEstimate}
				}
				got = schema.ValidateUsageDetail(u)
			case "raw":
				got = schema.ScanRawJSONDocument([]byte(c.Raw), schema.RawJSONPathPolicy{MaxDocumentBytes: 1024, MaxDocumentDepth: 8})
			case "metadataRaw":
				_, got = schema.DecodeNativeMetadataDataRaw([]byte(c.Raw))
			case "payload":
				u := schema.UsageDetail{OwnerID: "owner", SourceEntryRef: "entry", Scope: schema.UsageScopeAssistant, Completeness: schema.UsageUnknown}
				p := schema.SessionDetailPayload{Harness: schema.HarnessClaudeCode, Turns: []schema.TurnDetail{{Index: 1, Role: schema.RoleAssistant, SourceEntryRef: "entry", Usage: &u}}}
				if c.DuplicateOwner {
					p.Turns = append(p.Turns, schema.TurnDetail{Index: 2, Role: schema.RoleAssistant, SourceEntryRef: "entry", Usage: &u})
				}
				if c.WrongToolScope {
					p.Turns[0].Usage = nil
					p.Turns[0].ToolCalls = []schema.ToolCallDetail{{ID: "tool", Usage: &u}}
				}
				got = schema.ValidateSessionDetailPayload(p)
			case "metadataTarget":
				idx := 1
				records := []schema.NativeMetadataRecord{{ID: "metadata", Kind: schema.NativeMetadataPiToolResultDetails, Source: schema.NativeSourceRef{EntryRef: "result", SourceType: schema.NativeSourcePiMessage, MessageRole: schema.NativePiMessageRoleToolResult}, Attachment: &schema.NativeAttachmentRef{TurnIndex: &idx, ToolCallID: "missing"}, Data: json.RawMessage(`{}`)}}
				got = schema.ValidateNativeMetadataRecords(records, []schema.TurnDetail{{Index: 1, Role: schema.RoleAssistant, Timestamp: time.Unix(0, 0)}})
			case "capabilities":
				u := schema.UsageDetail{OwnerID: "owner", SourceEntryRef: "entry", Scope: schema.UsageScopeAssistant, Completeness: schema.UsageUnknown}
				p := schema.SessionDetailPayload{Turns: []schema.TurnDetail{{Index: 1, Role: schema.RoleAssistant, ObservedModel: "model", Usage: &u}}, NativeMetadata: []schema.NativeMetadataRecord{{ID: "metadata", Kind: schema.NativeMetadataPiCustomData, Source: schema.NativeSourceRef{EntryRef: "custom", SourceType: schema.NativeSourcePiCustom}, CustomType: "extension", Data: json.RawMessage(`{}`)}}}
				actual := schema.RequiredContentCapabilities(p)
				if strings.Join(stringSlice(actual), ",") != strings.Join(stringSlice(c.Capabilities), ",") {
					t.Fatalf("capabilities %v", actual)
				}
			default:
				t.Fatalf("unknown fixture operation %q", c.Operation)
			}
			if c.Accepted && (got != nil) {
				t.Fatalf("unexpected error: %v", got)
			}
			if !c.Accepted && (got == nil || !strings.Contains(got.Error(), c.ErrorContains)) {
				t.Fatalf("error %v, want %q", got, c.ErrorContains)
			}
		})
	}
	required := map[string]bool{}
	for _, name := range f.RequiredCaseNames {
		if name == "" || required[name] || !seen[name] {
			t.Fatalf("required case %q missing", name)
		}
		required[name] = true
	}
	if !reflect.DeepEqual(seen, required) {
		t.Fatal("required manifest must name every case exactly once")
	}
}
func stringSlice(v []schema.ContentCapability) []string {
	r := make([]string, len(v))
	for i := range v {
		r[i] = string(v[i])
	}
	return r
}
