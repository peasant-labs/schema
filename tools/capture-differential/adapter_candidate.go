//go:build !capture_baseline

package main

import (
	"errors"
	"fmt"

	schema "github.com/peasant-labs/schema"
)

const adapterName = "candidate"

// classify reads the candidate's refusal class through errors.As, never text.
func classify(err error) (string, int, error) {
	var refusal *schema.TranscriptCaptureError
	if !errors.As(err, &refusal) || !refusal.Refusal.IsValid() {
		return "", 0, fmt.Errorf("capture-differential candidate classifier: refusal %q is not a *schema.TranscriptCaptureError with a valid class", err)
	}
	return string(refusal.Refusal), refusal.Turn, nil
}

func recordedIndex(t schema.TranscriptCookTurn) (schema.TranscriptRecordedIndexState, *int) {
	return t.RecordedIndex()
}
