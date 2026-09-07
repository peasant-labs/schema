package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/contract/raw_json_boundaries.yaml
var rawJSONBoundaryYAML []byte

type rawJSONBoundaryCase struct {
	Name                   string                     `yaml:"name"`
	Operation              string                     `yaml:"operation"`
	ErrorContains          string                     `yaml:"errorContains"`
	ErrorCategory          string                     `yaml:"errorCategory"`
	Raw                    string                     `yaml:"raw"`
	GoOwnerBytesHex        string                     `yaml:"goOwnerBytesHex"`
	GoNamespaceBytesHex    string                     `yaml:"goNamespaceBytesHex"`
	Targets                string                     `yaml:"targets"`
	Accepted               bool                       `yaml:"accepted"`
	Pointers               []string                   `yaml:"pointers"`
	MaxDepth               int                        `yaml:"maxDepth"`
	Generate               *rawJSONGeneration         `yaml:"generate"`
	RequiredCapabilities   []schema.ContentCapability `yaml:"requiredCapabilities"`
	ExpectedDataBytes      int                        `yaml:"expectedDataBytes"`
	ExpectedAggregateBytes int                        `yaml:"expectedAggregateBytes"`
}

// Generation expands only fixture input, never implements validation. Raw
// templates retain whitespace and escapes until the production scanner runs.
type rawJSONGeneration struct {
	Unit       string `yaml:"unit"`
	Repeat     int    `yaml:"repeat"`
	Count      int    `yaml:"count"`
	Depth      int    `yaml:"depth"`
	Whitespace int    `yaml:"whitespace"`
	Copies     int    `yaml:"copies"`
	Padding    int    `yaml:"padding"`
	LastRepeat *int   `yaml:"lastRepeat"`
}

type rawJSONBoundaryFixtures struct {
	RequiredCaseNames []string              `yaml:"requiredCaseNames"`
	Cases             []rawJSONBoundaryCase `yaml:"cases"`
}

func loadRawJSONBoundaryFixtures(t *testing.T) rawJSONBoundaryFixtures {
	t.Helper()
	var document yaml.Node
	if err := yaml.Unmarshal(rawJSONBoundaryYAML, &document); err != nil {
		t.Fatal(err)
	}
	rejectDuplicateYAMLKeys(t, &document)
	var fixtures rawJSONBoundaryFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(rawJSONBoundaryYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatal(err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		t.Fatalf("fixture must contain one YAML document: %v", err)
	}
	seen := map[string]bool{}
	for _, c := range fixtures.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("empty or duplicate case name %q", c.Name)
		}
		seen[c.Name] = true
		switch c.Operation {
		case "scanner", "metadata", "detail", "turn", "metadata-records", "metadata-value", "owner-unicode", "namespace-unicode":
		default:
			t.Fatalf("case %q has unknown operation %q", c.Name, c.Operation)
		}
		if !c.Accepted && c.ErrorCategory != "lexical" && c.ErrorCategory != "validation" {
			t.Fatalf("case %q needs a closed error category", c.Name)
		}
		if c.Generate != nil && (c.Generate.Repeat < 0 || c.Generate.Count < 0 || c.Generate.Depth < 0 || c.Generate.Whitespace < 0 || c.Generate.Copies < 0 || c.Generate.Padding < 0) {
			t.Fatalf("case %q has invalid input expansion", c.Name)
		}
		if g := c.Generate; g != nil && g.LastRepeat != nil && (*g.LastRepeat < 0 || g.Copies == 0 || g.Count != 0 || g.Depth != 0 || g.Padding != 0) {
			t.Fatalf("case %q needs nonnegative lastRepeat with scalar copies", c.Name)
		}
	}
	required := map[string]bool{}
	for _, name := range fixtures.RequiredCaseNames {
		if name == "" || required[name] || !seen[name] {
			t.Fatalf("required case %q empty, repeated, or missing", name)
		}
		required[name] = true
	}
	if !reflect.DeepEqual(seen, required) {
		t.Fatal("required manifest must name every case exactly once")
	}
	return fixtures
}

func rejectDuplicateYAMLKeys(t *testing.T, node *yaml.Node) {
	t.Helper()
	if node.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			if seen[key] {
				t.Fatalf("duplicate YAML key %q", key)
			}
			seen[key] = true
		}
	}
	for _, child := range node.Content {
		rejectDuplicateYAMLKeys(t, child)
	}
}

func expandRawCase(c rawJSONBoundaryCase) string {
	raw := c.Raw
	if g := c.Generate; g != nil {
		value := strings.Repeat(g.Unit, g.Repeat)
		if g.Count > 0 {
			items := make([]string, g.Count)
			for i := range items {
				items[i] = `"` + value + `"`
			}
			value = "[" + strings.Join(items, ","+strings.Repeat(" ", g.Whitespace)) + "]"
		}
		for i := 0; i < g.Depth; i++ {
			value = `{"x":` + value + `}`
		}
		raw = strings.ReplaceAll(raw, "{{value}}", value)
		raw = strings.ReplaceAll(raw, "{{padding}}", strings.Repeat("p", g.Padding))
		if g.Copies > 0 {
			items := make([]string, g.Copies)
			for i := range items {
				items[i] = strings.ReplaceAll(raw, "{{index}}", fmt.Sprint(i))
				if i == len(items)-1 && g.LastRepeat != nil {
					items[i] = strings.ReplaceAll(strings.ReplaceAll(c.Raw, "{{value}}", strings.Repeat(g.Unit, *g.LastRepeat)), "{{index}}", fmt.Sprint(i))
				}
			}
			raw = "[" + strings.Join(items, ",") + "]"
		}
	}
	return raw
}

func rawDetail(turns, metadata, harness string) string {
	extra := ""
	if metadata != "" {
		extra = `,"nativeMetadata":` + metadata
	}
	return `{"id":"s","harness":"` + harness + `","startTime":"2020-01-01T00:00:00Z","endTime":"2020-01-01T00:00:00Z","durationMins":0,"totalTokens":0,"tokensIn":0,"tokensOut":0,"turnCount":0,"toolCallCount":0,"turns":` + turns + extra + `}`
}

func assertRawOutcome(t *testing.T, c rawJSONBoundaryCase, got any, err error, expected string) {
	t.Helper()
	if c.Accepted {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			encoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var actualValue, expectedValue any
			if err := json.Unmarshal(encoded, &actualValue); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(expected), &expectedValue); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actualValue, expectedValue) {
				t.Fatalf("accepted output changed:\ngot %s\nwant %s", encoded, expected)
			}
		}
		return
	}
	if err == nil {
		t.Fatal("invalid fixture was accepted")
	}
	category := "validation"
	if strings.Contains(strings.ToLower(err.Error()), "raw json validation failed") {
		category = "lexical"
	}
	if category != c.ErrorCategory {
		t.Fatalf("error category %s, want %s: %v", category, c.ErrorCategory, err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(c.ErrorContains)) {
		t.Fatalf("error %v, want %q", err, c.ErrorContains)
	}
}

func testDetailRawExits(t *testing.T, c rawJSONBoundaryCase, raw string) {
	t.Helper()
	t.Run("detail", func(t *testing.T) {
		value, err := schema.DecodeSessionDetailPayloadRaw([]byte(raw))
		assertRawOutcome(t, c, value, err, raw)
		if c.Accepted && c.RequiredCapabilities != nil && !slices.Equal(schema.RequiredContentCapabilities(value), c.RequiredCapabilities) {
			t.Fatalf("decoded detail requirements %v, want %v", schema.RequiredContentCapabilities(value), c.RequiredCapabilities)
		}
	})
	t.Run("transcript", func(t *testing.T) {
		envelope := `{"kind":"session_detail","contractVersion":"1.0.0","sessionDetail":` + raw + `}`
		value, err := schema.DecodeTranscriptContentRaw([]byte(envelope))
		assertRawOutcome(t, c, value, err, envelope)
		if c.Accepted && c.RequiredCapabilities != nil && !slices.Equal(schema.RequiredContentCapabilities(*value.SessionDetail), c.RequiredCapabilities) {
			t.Fatalf("decoded transcript requirements %v, want %v", schema.RequiredContentCapabilities(*value.SessionDetail), c.RequiredCapabilities)
		}
	})
	if c.Accepted {
		t.Run("typed-roundtrip", func(t *testing.T) {
			var value schema.SessionDetailPayload
			if err := json.Unmarshal([]byte(raw), &value); err != nil {
				t.Fatal(err)
			}
			assertRawOutcome(t, c, value, schema.ValidateSessionDetailPayload(value), raw)
			envelope := schema.TranscriptContent{Kind: schema.ContentKindSessionDetail, SessionDetail: &value}
			if err := schema.ValidateTranscriptContent(envelope); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRawJSONBoundariesSharedCorpus(t *testing.T) {
	for _, c := range loadRawJSONBoundaryFixtures(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			raw := expandRawCase(c)
			if c.ExpectedDataBytes > 0 || c.ExpectedAggregateBytes > 0 {
				var records []struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal([]byte(raw), &records); err != nil {
					t.Fatal(err)
				}
				total := 0
				for _, record := range records {
					if c.ExpectedDataBytes > 0 && len(record.Data) != c.ExpectedDataBytes {
						t.Fatalf("fixture data bytes %d, want %d", len(record.Data), c.ExpectedDataBytes)
					}
					total += len(record.Data)
				}
				if c.ExpectedAggregateBytes > 0 && total != c.ExpectedAggregateBytes {
					t.Fatalf("fixture aggregate data bytes %d, want %d", total, c.ExpectedAggregateBytes)
				}
			}
			switch c.Operation {
			case "namespace-unicode":
				var value schema.SessionDetailPayload
				if err := json.Unmarshal([]byte(rawDetail("["+raw+"]", "", "claude-code")), &value); err != nil {
					t.Fatal(err)
				}
				invalid, err := hex.DecodeString(c.GoNamespaceBytesHex)
				if err != nil {
					t.Fatal(err)
				}
				namespace := string(invalid)
				value.Turns[0].ToolCalls[0].Namespace = &namespace
				assertRawOutcome(t, c, nil, schema.ValidateSessionDetailPayload(value), raw)
				assertRawOutcome(t, c, nil, schema.ValidateTranscriptContent(schema.TranscriptContent{Kind: schema.ContentKindSessionDetail, SessionDetail: &value}), raw)
			case "scanner":
				depth := c.MaxDepth
				if depth == 0 {
					depth = 64
				}
				err := schema.ScanRawJSONDocument([]byte(raw), schema.RawJSONPathPolicy{MaxDocumentBytes: 8 << 20, MaxDocumentDepth: depth, OpaqueMetadataPointers: c.Pointers})
				assertRawOutcome(t, c, nil, err, raw)
			case "metadata":
				value, err := schema.DecodeNativeMetadataDataRaw([]byte(raw))
				assertRawOutcome(t, c, value, err, raw)
			case "turn":
				testDetailRawExits(t, c, rawDetail("["+raw+"]", "", "claude-code"))
			case "detail":
				testDetailRawExits(t, c, raw)
			case "metadata-records":
				var targets []schema.TurnDetail
				if err := json.Unmarshal([]byte(c.Targets), &targets); err != nil {
					t.Fatal(err)
				}
				value, err := schema.DecodeNativeMetadataRecordsRaw([]byte(raw), targets)
				assertRawOutcome(t, c, value, err, raw)
				if c.ErrorCategory == "lexical" {
					t.Run("raw-public-bounds", func(t *testing.T) { testDetailRawExits(t, c, rawDetail(c.Targets, raw, "claude-code")) })
				}
				t.Run("pi-public-roots", func(t *testing.T) {
					testDetailRawExits(t, c, rawDetail(c.Targets, raw, "pi"))
				})
			case "metadata-value":
				var records []schema.NativeMetadataRecord
				var targets []schema.TurnDetail
				if err := json.Unmarshal([]byte(raw), &records); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte(c.Targets), &targets); err != nil {
					t.Fatal(err)
				}
				assertRawOutcome(t, c, records, schema.ValidateNativeMetadataRecords(records, targets), raw)
				payload := schema.SessionDetailPayload{Harness: schema.HarnessPi, Turns: targets, NativeMetadata: records}
				assertRawOutcome(t, c, nil, schema.ValidateSessionDetailPayload(payload), "")
				assertRawOutcome(t, c, nil, schema.ValidateTranscriptContent(schema.TranscriptContent{Kind: schema.ContentKindSessionDetail, SessionDetail: &payload}), "")
			case "owner-unicode":
				// Go strings can contain invalid UTF-8; JSON cannot represent those
				// bytes. Exercise the public value boundary before serialization.
				var value schema.SessionDetailPayload
				if err := json.Unmarshal([]byte(rawDetail("["+raw+"]", "", "claude-code")), &value); err != nil {
					t.Fatal(err)
				}
				ownerBytes, err := hex.DecodeString(c.GoOwnerBytesHex)
				if err != nil {
					t.Fatal(err)
				}
				value.Turns[0].Usage.OwnerID = schema.UsageOwnerID(string(ownerBytes))
				assertRawOutcome(t, c, nil, schema.ValidateSessionDetailPayload(value), "")
			}
		})
	}
}
