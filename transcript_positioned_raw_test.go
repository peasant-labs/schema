package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	schema "github.com/peasant-labs/schema"
)

//go:embed testdata/transcript_positioned_segment_raw_golden.json
var transcriptPositionedRawGolden []byte

func TestPositionedSegmentRawGolden(t *testing.T) {
	// The fixture's final newline is a file terminator, not a response byte.
	want := bytes.TrimSuffix(transcriptPositionedRawGolden, []byte("\n"))
	var segment schema.PositionedSegment
	if err := json.Unmarshal(want, &segment); err != nil {
		t.Fatal(err)
	}
	got, err := schema.MarshalPositionedSegment(segment)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("segment bytes differ from the stored-turn golden:\ngot  %s\nwant %s", got, want)
	}
	var decoded schema.PositionedSegment
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatal(err)
	}
	for i := range segment.Rows {
		if !bytes.Equal(segment.Rows[i].Turn.Raw(), decoded.Rows[i].Turn.Raw()) {
			t.Fatalf("turn %d changed during segment round trip", i)
		}
	}
	if _, err := json.Marshal(segment); !errors.Is(err, schema.ErrPositionedSegmentNotEncodable) {
		t.Fatalf("plain segment marshal error = %v", err)
	}
	if err := json.NewEncoder(new(bytes.Buffer)).Encode(segment); !errors.Is(err, schema.ErrPositionedSegmentNotEncodable) {
		t.Fatalf("plain segment encoder error = %v", err)
	}
	if _, err := json.Marshal(segment.Rows[0]); !errors.Is(err, schema.ErrTranscriptSegmentRowNotEncodable) || errors.Is(err, schema.ErrPositionedSegmentNotEncodable) {
		t.Fatalf("plain row marshal error = %v", err)
	}
	if err := json.NewEncoder(new(bytes.Buffer)).Encode(segment.Rows[0]); !errors.Is(err, schema.ErrTranscriptSegmentRowNotEncodable) || errors.Is(err, schema.ErrPositionedSegmentNotEncodable) {
		t.Fatalf("plain row encoder error = %v", err)
	}
}

func TestPositionedSegmentRawFixture(t *testing.T) {
	for _, c := range loadPositionedSegmentCases(t).Raw.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var segment schema.PositionedSegment
			if err := json.Unmarshal(bytes.TrimSuffix(transcriptPositionedRawGolden, []byte("\n")), &segment); err != nil {
				t.Fatal(err)
			}
			encoded, err := schema.MarshalPositionedSegment(segment)
			if err != nil {
				t.Fatal(err)
			}
			switch c.Input.Mode {
			case "members":
				decoder := json.NewDecoder(bytes.NewReader(encoded))
				if _, err := decoder.Token(); err != nil {
					t.Fatal(err)
				}
				var names []string
				for decoder.More() {
					name, err := decoder.Token()
					if err != nil {
						t.Fatal(err)
					}
					names = append(names, name.(string))
					var value json.RawMessage
					if err := decoder.Decode(&value); err != nil {
						t.Fatal(err)
					}
				}
				if !reflect.DeepEqual(names, c.Expected.Members) {
					t.Fatalf("writer members = %v, want %v", names, c.Expected.Members)
				}
			case "segment-refusal":
				_, err := json.Marshal(segment)
				if !errors.Is(err, schema.ErrPositionedSegmentNotEncodable) {
					t.Fatalf("segment marshal error = %v", err)
				}
				err = json.NewEncoder(new(bytes.Buffer)).Encode(segment)
				if !errors.Is(err, schema.ErrPositionedSegmentNotEncodable) {
					t.Fatalf("segment encoder error = %v", err)
				}
			case "row-refusal":
				row := segment.Rows[0]
				_, err := json.Marshal(row)
				if !errors.Is(err, schema.ErrTranscriptSegmentRowNotEncodable) || errors.Is(err, schema.ErrPositionedSegmentNotEncodable) {
					t.Fatalf("row marshal error = %v", err)
				}
				err = json.NewEncoder(new(bytes.Buffer)).Encode(row)
				if !errors.Is(err, schema.ErrTranscriptSegmentRowNotEncodable) || errors.Is(err, schema.ErrPositionedSegmentNotEncodable) {
					t.Fatalf("row encoder error = %v", err)
				}
			case "round-trip":
				var decoded schema.PositionedSegment
				if err := json.Unmarshal(encoded, &decoded); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(decoded, segment) {
					t.Fatal("the assembled segment did not round-trip every member and raw turn")
				}
			case "turn":
				var decoded schema.PositionedSegment
				if err := json.Unmarshal(encoded, &decoded); err != nil {
					t.Fatal(err)
				}
				if c.Input.Row < 0 || c.Input.Row >= len(decoded.Rows) {
					t.Fatal("fixture row is outside the segment")
				}
				if got := string(decoded.Rows[c.Input.Row].Turn.Raw()); got != c.Expected.Raw {
					t.Fatalf("turn bytes = %q, want %q", got, c.Expected.Raw)
				}
			case "omissions":
				var decoded struct {
					Rows   []map[string]json.RawMessage `json:"rows"`
					Native json.RawMessage              `json:"nativeMetadata"`
				}
				if err := json.Unmarshal(encoded, &decoded); err != nil {
					t.Fatal(err)
				}
				if c.Input.Row < 0 || c.Input.Row >= len(decoded.Rows) {
					t.Fatal("fixture row is outside the segment")
				}
				row := decoded.Rows[c.Input.Row]
				if _, ok := row["recordedIndex"]; ok {
					t.Fatal("absent recordedIndex was emitted")
				}
				if _, ok := row["displayNumber"]; ok {
					t.Fatal("absent displayNumber was emitted")
				}
				if decoded.Native != nil {
					t.Fatal("absent nativeMetadata was emitted")
				}
			default:
				t.Fatalf("unknown raw fixture mode %q", c.Input.Mode)
			}
		})
	}
}
