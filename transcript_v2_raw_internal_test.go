package schema

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestTranscriptSegmentV2WriterMemberOrder(t *testing.T) {
	typ := reflect.TypeFor[TranscriptSegmentV2]()
	names := make([]string, typ.NumField())
	for i := range names {
		names[i] = strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
	}
	if !slices.Equal(names, transcriptSegmentV2Members) {
		t.Fatalf("segment fields = %v, writer fields = %v", names, transcriptSegmentV2Members)
	}
}
