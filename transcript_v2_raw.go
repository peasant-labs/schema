package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// ErrTranscriptSegmentV2NotEncodable identifies an attempt to encode a segment
// through encoding/json, which would rewrite its turns' preserved bytes.
var ErrTranscriptSegmentV2NotEncodable = errors.New("a v2 segment must be written by schema.MarshalTranscriptSegmentV2")

// ErrTranscriptSegmentRowNotEncodable identifies an attempt to encode a row
// outside its segment, which would rewrite the row's preserved turn bytes.
var ErrTranscriptSegmentRowNotEncodable = errors.New("a v2 segment row must be written by schema.writeTranscriptSegmentRow")

func (TranscriptSegmentV2) MarshalJSON() ([]byte, error) {
	return nil, ErrTranscriptSegmentV2NotEncodable
}

func (TranscriptSegmentRow) MarshalJSON() ([]byte, error) {
	return nil, ErrTranscriptSegmentRowNotEncodable
}

var transcriptSegmentV2Members = []string{
	"context", "partition", "index", "rows", "byteLength", "hasOlder", "hasNewer", "nativeMetadata",
}

// ValidateTranscriptSegmentV2Boundaries compares the segment's served rows
// with the directory ref it was cut from. Validate checks the segment alone;
// this function checks the independent directory coordinates.
func ValidateTranscriptSegmentV2Boundaries(seg TranscriptSegmentV2, ref TranscriptSegmentRefV2) error {
	const where = "v2 segment boundary validation failed at schema.ValidateTranscriptSegmentV2Boundaries: "
	if ref.Index != seg.Index {
		return fmt.Errorf("%ssegment ordinal %d differs from ref ordinal %d; the rows cannot be placed under this ref; use the ref the segment was cut from", where, seg.Index, ref.Index)
	}
	if ref.SourceFirstPosition < 0 || ref.SourceLastPosition < ref.SourceFirstPosition ||
		ref.SourceLastPosition-ref.SourceFirstPosition != len(seg.Rows)-1 {
		return fmt.Errorf("%sthe %d rows do not match sourceFirstPosition %d and sourceLastPosition %d; the ref describes a different extent; rebuild the segment from that ref", where, len(seg.Rows), ref.SourceFirstPosition, ref.SourceLastPosition)
	}
	for i, row := range seg.Rows {
		if row.Position-ref.SourceFirstPosition != i {
			return fmt.Errorf("%srow %d carries position %d for sourceFirstPosition %d and sourceLastPosition %d; the row is outside its ref's served order; rebuild the segment from that ref", where, i, row.Position, ref.SourceFirstPosition, ref.SourceLastPosition)
		}
	}
	return nil
}

// MarshalTranscriptSegmentV2 validates and writes one complete response. The
// turn objects are written directly, retaining their original served bytes.
// The caller writes the returned bytes as the entire JSON response body.
func MarshalTranscriptSegmentV2(seg TranscriptSegmentV2) ([]byte, error) {
	const where = "v2 segment encoding failed at schema.MarshalTranscriptSegmentV2: "
	if err := seg.Validate(); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString(`{"context":`)
	if err := writeTranscriptJSONValue(&buf, seg.Context); err != nil {
		return nil, fmt.Errorf("%scontext: %w", where, err)
	}
	buf.WriteString(`,"partition":`)
	if err := writeTranscriptJSONValue(&buf, seg.Partition); err != nil {
		return nil, fmt.Errorf("%spartition: %w", where, err)
	}
	buf.WriteString(`,"index":`)
	buf.WriteString(strconv.Itoa(seg.Index))
	buf.WriteString(`,"rows":[`)
	for i := range seg.Rows {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := writeTranscriptSegmentRow(&buf, seg.Rows[i]); err != nil {
			return nil, fmt.Errorf("%srows[%d]: %w", where, i, err)
		}
	}
	buf.WriteString(`],"byteLength":`)
	buf.WriteString(strconv.Itoa(seg.ByteLength))
	buf.WriteString(`,"hasOlder":`)
	buf.WriteString(strconv.FormatBool(seg.HasOlder))
	buf.WriteString(`,"hasNewer":`)
	buf.WriteString(strconv.FormatBool(seg.HasNewer))
	if len(seg.NativeMetadata) > 0 {
		buf.WriteString(`,"nativeMetadata":`)
		if err := writeTranscriptJSONValue(&buf, seg.NativeMetadata); err != nil {
			return nil, fmt.Errorf("%snativeMetadata: %w", where, err)
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func writeTranscriptSegmentRow(buf *bytes.Buffer, row TranscriptSegmentRow) error {
	const where = "v2 segment row encoding failed at schema.writeTranscriptSegmentRow: "
	buf.WriteString(`{"position":`)
	buf.WriteString(strconv.Itoa(row.Position))
	buf.WriteString(`,"recordedIndexState":`)
	if err := writeTranscriptJSONValue(buf, row.RecordedIndexState); err != nil {
		return fmt.Errorf("%srecordedIndexState: %w", where, err)
	}
	if row.RecordedIndex != nil {
		buf.WriteString(`,"recordedIndex":`)
		buf.WriteString(strconv.Itoa(*row.RecordedIndex))
	}
	if row.DisplayNumber != "" {
		buf.WriteString(`,"displayNumber":`)
		if err := writeTranscriptJSONValue(buf, row.DisplayNumber); err != nil {
			return fmt.Errorf("%sdisplayNumber: %w", where, err)
		}
	}
	buf.WriteString(`,"turn":`)
	buf.Write(row.Turn.Raw())
	buf.WriteByte('}')
	return nil
}

func writeTranscriptJSONValue(buf *bytes.Buffer, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = buf.Write(encoded)
	return err
}
