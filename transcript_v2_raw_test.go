package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"testing"

	schema "github.com/peasant-labs/schema"
)

//go:embed testdata/transcript_segment_v2_raw_golden.json
var transcriptSegmentV2RawGolden []byte

func TestTranscriptSegmentV2RawGolden(t *testing.T) {
	// The fixture's final newline is a file terminator, not a response byte.
	want := bytes.TrimSuffix(transcriptSegmentV2RawGolden, []byte("\n"))
	var segment schema.TranscriptSegmentV2
	if err := json.Unmarshal(want, &segment); err != nil {
		t.Fatal(err)
	}
	got, err := schema.MarshalTranscriptSegmentV2(segment)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("segment bytes differ from the stored-turn golden:\ngot  %s\nwant %s", got, want)
	}
	var decoded schema.TranscriptSegmentV2
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatal(err)
	}
	for i := range segment.Rows {
		if !bytes.Equal(segment.Rows[i].Turn.Raw(), decoded.Rows[i].Turn.Raw()) {
			t.Fatalf("turn %d changed during segment round trip", i)
		}
	}
	if _, err := json.Marshal(segment); !errors.Is(err, schema.ErrTranscriptSegmentV2NotEncodable) {
		t.Fatalf("plain segment marshal error = %v", err)
	}
	if err := json.NewEncoder(new(bytes.Buffer)).Encode(segment); !errors.Is(err, schema.ErrTranscriptSegmentV2NotEncodable) {
		t.Fatalf("plain segment encoder error = %v", err)
	}
	if _, err := json.Marshal(segment.Rows[0]); !errors.Is(err, schema.ErrTranscriptSegmentRowNotEncodable) || errors.Is(err, schema.ErrTranscriptSegmentV2NotEncodable) {
		t.Fatalf("plain row marshal error = %v", err)
	}
	if err := json.NewEncoder(new(bytes.Buffer)).Encode(segment.Rows[0]); !errors.Is(err, schema.ErrTranscriptSegmentRowNotEncodable) || errors.Is(err, schema.ErrTranscriptSegmentV2NotEncodable) {
		t.Fatalf("plain row encoder error = %v", err)
	}
}
