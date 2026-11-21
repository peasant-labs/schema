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

func turnFacts(t schema.TranscriptCookTurn) facts {
	f := t.Facts()
	return facts{Index: f.Index, Role: string(f.Role), EntryType: string(f.EntryType), Depth: f.Depth, Usage: f.UsagePresent,
		Conversation: f.ConversationTurn, ByteLength: f.ByteLength}
}

func cutKept(turns []schema.TranscriptCookTurn, verdicts []schema.CookedTurn, budget int) ([]schema.TranscriptSegmentRef, error) {
	return schema.CutTranscriptCookTurns(turns, verdicts, budget)
}

// partitionPass is the same partition build on the captured-turn facts.
func partitionPass(doc []byte) (int, error) {
	turns, err := schema.NewTranscriptCookTurns(doc)
	if err != nil {
		return 0, err
	}
	verdicts, err := schema.CookTranscriptRows(turns, schema.HarnessClaudeCode)
	if err != nil {
		return 0, err
	}
	sink := 0
	for i, turn := range turns {
		if !verdicts[i].Keep {
			continue
		}
		f := turn.Facts()
		if f.UsagePresent {
			sink++
		}
		if _, index := turn.RecordedIndex(); index != nil {
			sink += *index
		}
		if f.ConversationTurn {
			sink++
		}
		sink += f.ByteLength
	}
	refs, err := schema.CutTranscriptCookTurns(turns, verdicts, schema.TranscriptSegmentBudgetBytes)
	if err != nil {
		return 0, err
	}
	return sink + len(refs), nil
}
