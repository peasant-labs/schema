package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_download_text.yaml
var downloadTextYAML []byte

//go:embed testdata/transcript_download_text_manifest.yaml
var downloadTextManifestYAML []byte

type downloadTextInput struct {
	Subject        string   `yaml:"subject"`
	Replay         []string `yaml:"replay"`
	Raw            string   `yaml:"raw"`
	Depth          int      `yaml:"depth"`
	KeyBytes       int      `yaml:"keyBytes"`
	Bytes          int      `yaml:"bytes"`
	SyntheticBytes bool     `yaml:"syntheticBytes"`
}

type downloadTextExpected struct {
	Outcome  string `yaml:"outcome"`
	Failure  string `yaml:"failure"`
	Contains string `yaml:"contains"`
}

func downloadTextFixture(t *testing.T) testcase.Corpus[downloadTextInput, downloadTextExpected] {
	t.Helper()
	corpus := decodeStrictYAML[testcase.Corpus[downloadTextInput, downloadTextExpected]](t, downloadTextYAML, "transcript download fixture")
	caseassert.RequireValid(t, corpus)
	var names []string
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
		if len(c.Input.Replay) == 0 || !slices.Contains(c.Input.Replay, "go") {
			t.Fatalf("%s has no Go replay", c.Name)
		}
	}
	caseassert.RequireManifestNames(t, "transcript download", downloadTextManifestYAML, names)
	return corpus
}

func expandDownloadInput(t *testing.T, input downloadTextInput) []byte {
	t.Helper()
	if input.SyntheticBytes {
		const prefix = `{"padding":"`
		const suffix = `","sessionDetail":{"id":"legacy","turns":[]},"kind":"session_detail","contractVersion":"0.1.1"}`
		if input.Bytes < len(prefix)+len(suffix) {
			t.Fatal("synthetic body size is too small")
		}
		return []byte(prefix + strings.Repeat("x", input.Bytes-len(prefix)-len(suffix)) + suffix)
	}
	raw := strings.ReplaceAll(input.Raw, "{{key}}", strings.Repeat("k", input.KeyBytes))
	if input.Depth > 0 {
		raw = strings.Repeat("[", input.Depth) + raw + strings.Repeat("]", input.Depth)
	}
	return []byte(raw)
}

func TestTranscriptDownloadText(t *testing.T) {
	for _, c := range downloadTextFixture(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Input.SyntheticBytes && testing.Short() {
				t.Skip("large boundary case")
			}
			raw := expandDownloadInput(t, c.Input)
			if c.Input.SyntheticBytes && len(raw) != c.Input.Bytes {
				t.Fatalf("body size = %d, want %d", len(raw), c.Input.Bytes)
			}
			var result schema.TranscriptDownloadRead
			var err error
			switch c.Input.Subject {
			case "download-read":
				result, err = schema.DecodeTranscriptDownloadRaw(raw)
			case "source-entry":
				_, err = schema.DecodeTranscriptContentRaw(raw)
			case "source-detail-entry":
				_, err = schema.DecodeSessionDetailPayloadRaw(raw)
			default:
				t.Fatalf("unknown subject %q", c.Input.Subject)
			}
			if c.Expected.Outcome == "refused" {
				if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(c.Expected.Contains)) {
					t.Fatalf("refusal = %v, want %q (%s)", err, c.Expected.Contains, c.Expected.Failure)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if string(result.Kind) != c.Expected.Outcome {
				t.Fatalf("outcome = %q, want %q", result.Kind, c.Expected.Outcome)
			}
			if result.Kind == schema.TranscriptDownloadRecords {
				var expected []json.RawMessage
				for _, line := range bytes.Split(raw, []byte("\n")) {
					if len(bytes.TrimSpace(line)) > 0 {
						expected = append(expected, append(json.RawMessage(nil), line...))
					}
				}
				if !slices.EqualFunc(result.Records, expected, func(a, b json.RawMessage) bool { return bytes.Equal(a, b) }) || result.Detail != nil {
					t.Fatalf("record output does not preserve the served JSONL")
				}
				return
			}
			expected := json.RawMessage(raw)
			var object map[string]json.RawMessage
			if json.Unmarshal(raw, &object) == nil && object["sessionDetail"] != nil {
				expected = object["sessionDetail"]
			}
			if !bytes.Equal(result.Detail, expected) || result.Records != nil {
				t.Fatal("download did not preserve the served detail bytes")
			}
		})
	}
}

func TestTranscriptDownloadStrictness(t *testing.T) {
	for _, c := range downloadTextFixture(t).Cases {
		if c.Input.Subject != "download-read" || c.Input.SyntheticBytes ||
			(c.Expected.Outcome != "strict" && c.Expected.Outcome != "legacy") {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			raw := expandDownloadInput(t, c.Input)
			var object map[string]json.RawMessage
			if err := json.Unmarshal(raw, &object); err != nil {
				t.Fatal(err)
			}
			if object["sessionDetail"] != nil {
				raw = object["sessionDetail"]
			}
			got := schema.TranscriptContentStrict(raw)
			if want := c.Expected.Outcome == "strict"; got != want {
				t.Fatalf("strictness = %v, want %v", got, want)
			}
		})
	}
}
