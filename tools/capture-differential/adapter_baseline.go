//go:build capture_baseline

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	schema "github.com/peasant-labs/schema"
)

// adapterName names the build. Only the baseline build may record fixture
// expectations.
const adapterName = "baseline"

const captureWhere = "transcript turn capture failed at schema.NewTranscriptCookTurns: "

var turnClause = regexp.MustCompile(`^turn (\d+) (has invalid or repeated members: |is not one JSON object|cannot be decoded as a turn)`)

// classify maps the baseline's refusal text to a refusal class through a
// closed table of message clauses. A message outside the table fails closed.
func classify(err error) (string, int, error) {
	text := err.Error()
	rest, ok := strings.CutPrefix(text, captureWhere)
	if !ok {
		return "", 0, fmt.Errorf("capture-differential baseline classifier: refusal %q lacks the capture prefix; extend the clause table only after reading the baseline source", text)
	}
	switch {
	case strings.HasPrefix(rest, "session detail is not one valid UTF-8 JSON object"):
		return "document-invalid", -1, nil
	case strings.HasPrefix(rest, "session detail is not one valid JSON object: a JSON object is required"):
		return "document-not-object", -1, nil
	case strings.HasPrefix(rest, "session detail is not one valid JSON object: duplicate object key "):
		return "document-duplicate-member", -1, nil
	case strings.HasPrefix(rest, "an envelope was supplied"):
		return "document-envelope", -1, nil
	case strings.HasPrefix(rest, "turns is not an array"):
		return "turns-not-array", -1, nil
	}
	if m := turnClause.FindStringSubmatch(rest); m != nil {
		turn, _ := strconv.Atoi(m[1])
		cause := rest[len(m[0]):]
		switch m[2] {
		case "has invalid or repeated members: ":
			const scan = "raw JSON validation failed at schema.ScanRawJSONDocument: "
			switch {
			case strings.HasPrefix(cause, scan+"unpaired Unicode surrogate escape"),
				strings.HasPrefix(cause, scan+"invalid Unicode surrogate pair"):
				return "turn-unicode-escape", turn, nil
			case strings.HasPrefix(cause, "raw JSON validation failed at schema.ScanRawJSONDocument during syntax scanning: "+scan+"duplicate object key "):
				return "turn-duplicate-member", turn, nil
			}
		case "is not one JSON object":
			return "turn-not-object", turn, nil
		case "cannot be decoded as a turn":
			return "turn-type-mismatch", turn, nil
		}
	}
	return "", 0, fmt.Errorf("capture-differential baseline classifier: refusal %q matches no clause of the class table; an unreachable clause or a new refusal must be investigated before recording", text)
}

func recordedIndex(t schema.TranscriptCookTurn) (schema.TranscriptRecordedIndexState, *int) {
	return schema.DecodeTranscriptRecordedIndex(t.Raw())
}

// turnFacts derives a partition index's per-turn values from Detail(), as
// Village's partition build does on the released API.
func turnFacts(t schema.TranscriptCookTurn) facts {
	d := t.Detail()
	usage := d.Usage != nil
	for _, tool := range d.ToolCalls {
		usage = usage || tool.Usage != nil
	}
	return facts{Index: d.Index, Role: string(d.Role), EntryType: string(d.EntryType), Depth: d.Depth, Usage: usage,
		Conversation: schema.IsConversationTurn(d.Role, d.EntryType, d.Depth), ByteLength: schema.TranscriptTurnByteLength(d)}
}

// cutKept cuts the kept turns' details.
func cutKept(turns []schema.TranscriptCookTurn, verdicts []schema.CookedTurn, budget int) ([]schema.TranscriptSegmentRef, error) {
	details := make([]schema.TurnDetail, 0, len(turns))
	for i, v := range verdicts {
		if v.Keep {
			details = append(details, turns[i].Detail())
		}
	}
	return schema.CutTranscriptSegments(details, budget), nil
}

// partitionPass is Village's partition build on the released API.
func partitionPass(doc []byte) (int, error) {
	turns, err := schema.NewTranscriptCookTurns(doc)
	if err != nil {
		return 0, err
	}
	verdicts, err := schema.CookTranscriptRows(turns, schema.HarnessClaudeCode)
	if err != nil {
		return 0, err
	}
	details := make([]schema.TurnDetail, 0, len(turns))
	sink := 0
	for i, turn := range turns {
		if !verdicts[i].Keep {
			continue
		}
		detail := turn.Detail()
		details = append(details, detail)
		if detail.Usage != nil {
			sink++
		}
		for _, tool := range detail.ToolCalls {
			if tool.Usage != nil {
				sink++
			}
		}
		if _, index := schema.DecodeTranscriptRecordedIndex(turn.Raw()); index != nil {
			sink += *index
		}
		if schema.IsConversationTurn(detail.Role, detail.EntryType, detail.Depth) {
			sink++
		}
		sink += schema.TranscriptTurnByteLength(detail)
	}
	return sink + len(schema.CutTranscriptSegments(details, schema.TranscriptSegmentBudgetBytes)), nil
}
