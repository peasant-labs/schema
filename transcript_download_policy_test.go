package schema

import (
	"reflect"
	"testing"
)

func TestTranscriptSourcePoliciesMatchExistingBoundaries(t *testing.T) {
	if !reflect.DeepEqual(TranscriptSourceReadPolicy, transcriptRawPolicy()) {
		t.Fatal("envelope source policy differs from the existing boundary")
	}
	if !reflect.DeepEqual(TranscriptSourceDetailReadPolicy, sessionDetailRawPolicy()) {
		t.Fatal("bare-detail source policy differs from the existing boundary")
	}
}
