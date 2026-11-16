package schema

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestPositionedSegmentWriterMemberOrder(t *testing.T) {
	typ := reflect.TypeFor[PositionedSegment]()
	names := make([]string, typ.NumField())
	for i := range names {
		names[i] = strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
	}
	if !slices.Equal(names, positionedSegmentMembers) {
		t.Fatalf("segment fields = %v, writer fields = %v", names, positionedSegmentMembers)
	}
}
