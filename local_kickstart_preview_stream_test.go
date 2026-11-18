package schema_test

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/kickstart_preview_stream.yaml
var previewStreamFixture []byte

//go:embed testdata/local-api/kickstart_preview_stream_manifest.yaml
var previewStreamManifest []byte

//go:embed testdata/local-api/kickstart_preview_requests_manifest.yaml
var previewRequestsManifest []byte

//go:embed testdata/local-api/kickstart_preview_http_manifest.yaml
var previewHTTPManifest []byte

type previewFixture struct {
	Frames map[string]map[string]any `yaml:"frames"`
	Cases  []struct {
		Name           string   `yaml:"name"`
		Frames         []string `yaml:"frames"`
		Valid          bool     `yaml:"valid"`
		ExpectSegments *int     `yaml:"expectSegments"`
		ExpectUnknown  *int     `yaml:"expectUnknown"`
		ExpectMetadata *int     `yaml:"expectMetadata"`
		ExpectHasNewer *bool    `yaml:"expectHasNewer"`
		TSValid        *bool    `yaml:"tsValid"`
		ProducerOnly   string   `yaml:"producerOnly"`
		Mutate         *struct {
			Frame int    `yaml:"frame"`
			Path  string `yaml:"path"`
			Value any    `yaml:"value"`
		} `yaml:"mutate"`
	} `yaml:"cases"`
	RequestCases []struct {
		Name              string         `yaml:"name"`
		Request           map[string]any `yaml:"request"`
		RepeatedNodeBytes int            `yaml:"repeatedNodeBytes"`
		MaxEnvelope       bool           `yaml:"maxEnvelope"`
		Valid             bool           `yaml:"valid"`
	} `yaml:"requestCases"`
	PreviewCases []struct {
		Name    string         `yaml:"name"`
		Preview map[string]any `yaml:"preview"`
		Valid   bool           `yaml:"valid"`
	} `yaml:"previewCases"`
}

func TestLocalKickstartPreviewHTTP(t *testing.T) {
	var fixture previewFixture
	if err := yaml.Unmarshal(previewStreamFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(fixture.PreviewCases))
	for i, c := range fixture.PreviewCases {
		names[i] = c.Name
	}
	caseassert.RequireManifestNames(t, "HTTP preview cases", previewHTTPManifest, names)
	for _, c := range fixture.PreviewCases {
		t.Run(c.Name, func(t *testing.T) {
			raw, err := json.Marshal(c.Preview)
			if err != nil {
				t.Fatal(err)
			}
			var value schema.LocalKickstartPreview
			err = json.Unmarshal(raw, &value)
			if err == nil {
				err = value.Validate()
			}
			if (err == nil) != c.Valid {
				t.Fatalf("valid = %v, expected %v: %v", err == nil, c.Valid, err)
			}
		})
	}
}

func TestLocalKickstartPreviewRequests(t *testing.T) {
	var fixture previewFixture
	if err := yaml.Unmarshal(previewStreamFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(fixture.RequestCases))
	for i, c := range fixture.RequestCases {
		names[i] = c.Name
	}
	caseassert.RequireManifestNames(t, "preview request cases", previewRequestsManifest, names)
	for _, c := range fixture.RequestCases {
		t.Run(c.Name, func(t *testing.T) {
			if c.RepeatedNodeBytes > 0 {
				c.Request["nodeId"] = strings.Repeat("\x01", c.RepeatedNodeBytes)
			}
			if c.MaxEnvelope {
				c.Request["requestId"] = strings.Repeat("r", 32)
				c.Request["sessionId"] = strings.Repeat("s", 128)
			}
			raw, err := json.Marshal(c.Request)
			if err != nil {
				t.Fatal(err)
			}
			var request schema.LocalKickstartPreviewRequest
			err = json.Unmarshal(raw, &request)
			if err == nil {
				err = request.Validate()
			}
			if (err == nil) != c.Valid {
				t.Fatalf("valid = %v, expected %v: %v", err == nil, c.Valid, err)
			}
			if c.Valid {
				wire, err := json.Marshal(schema.ClientMessage{Type: schema.MsgKickstartPreview, Preview: &request})
				if err != nil || len(wire) > 4096 {
					t.Fatalf("valid request exceeds read limit: %d, %v", len(wire), err)
				}
			}
		})
	}
}

func TestLocalKickstartPreviewStream(t *testing.T) {
	var fixture previewFixture
	if err := yaml.Unmarshal(previewStreamFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(fixture.Cases))
	for i, c := range fixture.Cases {
		names[i] = c.Name
	}
	caseassert.RequireManifestNames(t, "preview stream cases", previewStreamManifest, names)
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var state schema.LocalKickstartPreviewState
			var result error
			for i, name := range c.Frames {
				raw, err := json.Marshal(fixture.Frames[name])
				if err != nil {
					t.Fatal(err)
				}
				var object map[string]any
				if err := json.Unmarshal(raw, &object); err != nil {
					t.Fatal(err)
				}
				if c.Mutate != nil && c.Mutate.Frame == i {
					var target any = object
					parts := strings.Split(c.Mutate.Path, ".")
					for _, part := range parts[:len(parts)-1] {
						switch v := target.(type) {
						case map[string]any:
							target = v[part]
						case []any:
							n, err := strconv.Atoi(part)
							if err != nil {
								t.Fatal(err)
							}
							target = v[n]
						default:
							t.Fatal("fixture path does not address an object or array")
						}
					}
					last := parts[len(parts)-1]
					if targetMap, ok := target.(map[string]any); ok {
						targetMap[last] = c.Mutate.Value
					} else {
						t.Fatal("fixture mutation target is not an object")
					}
				}
				raw, err = json.Marshal(object)
				if err != nil {
					t.Fatal(err)
				}
				var frame schema.LocalKickstartPreviewEvent
				if err := json.Unmarshal(raw, &frame); err != nil {
					result = err
					break
				}
				oldSegments, err := json.Marshal(state.Segments)
				if err != nil {
					t.Fatal(err)
				}
				oldCount := len(state.Segments)
				state, result = schema.MergeLocalKickstartPreview(state, frame)
				if result != nil {
					t.Logf("frame %d (%s): %v", i, name, result)
					break
				}
				if oldCount > 0 {
					newSegments, err := json.Marshal(state.Segments[:oldCount])
					if err != nil || string(newSegments) != string(oldSegments) {
						t.Fatal("previously sent segments changed during merge")
					}
				}
			}
			if (result == nil) != c.Valid {
				t.Fatalf("valid = %v, expected %v: %v", result == nil, c.Valid, result)
			}
			unknown, metadata := 0, 0
			for _, segment := range state.Segments {
				unknown += len(segment.RetainedUnknown)
				metadata += len(segment.NativeMetadata)
			}
			if c.ExpectSegments != nil && len(state.Segments) != *c.ExpectSegments {
				t.Fatal("merged segment count differs from fixture")
			}
			if c.ExpectUnknown != nil && unknown != *c.ExpectUnknown {
				t.Fatal("retained unknown evidence was not preserved")
			}
			if c.ExpectMetadata != nil && metadata != *c.ExpectMetadata {
				t.Fatal("native metadata evidence was not preserved")
			}
			if c.ExpectHasNewer != nil && (len(state.Segments) == 0 || state.Segments[len(state.Segments)-1].HasNewer != *c.ExpectHasNewer) {
				t.Fatal("segment snapshot changed")
			}
			if result != nil && (strings.Contains(result.Error(), "ses_test") || strings.Contains(result.Error(), "/private/")) {
				t.Fatal("error disclosed source data")
			}
		})
	}
}
