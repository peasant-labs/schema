package schema

import (
	"strconv"
	"strings"
)

// CookedTurn describes one stored row's membership and its identity in the
// served partition. A removed row has ServedPosition -1 and no DisplayNumber.
type CookedTurn struct {
	Position       int
	Keep           bool
	ServedPosition int
	DisplayNumber  string
}

// CookTranscriptRows filters, deduplicates and numbers one partition using the
// same whole-document rule as the reader. Inputs must have passed the evidence
// validation applicable to the served document. Pi rows bypass the legacy
// filter. The input slice and its turns are not modified.
func CookTranscriptRows(turns []TurnDetail, harness Harness) []CookedTurn {
	result := make([]CookedTurn, len(turns))
	kept := make([]int, 0, len(turns))
	for i := range turns {
		result[i] = CookedTurn{Position: i, ServedPosition: -1}
		if harness != HarnessPi {
			if !keepTranscriptTurn(turns[i]) {
				continue
			}
			if len(kept) != 0 {
				prev := kept[len(kept)-1]
				if duplicateTranscriptTurns(turns[prev], turns[i]) {
					if len(turns[i].ToolCalls) > 0 && len(turns[prev].ToolCalls) == 0 {
						kept[len(kept)-1] = i
					}
					continue
				}
			}
		}
		kept = append(kept, i)
	}

	task, sub := 0, 0
	for served, stored := range kept {
		turn := turns[stored]
		startsTask := turn.Role == RoleUser && turn.Depth == 0 && trimTranscriptText(turn.Content) != ""
		if served == 0 || startsTask {
			task++
			sub = 0
		} else {
			sub++
		}
		label := strconv.Itoa(task)
		if sub > 0 {
			label += transcriptNumberSuffix(sub)
		}
		result[stored].Keep = true
		result[stored].ServedPosition = served
		result[stored].DisplayNumber = label
	}
	return result
}

func keepTranscriptTurn(turn TurnDetail) bool {
	if protectedTranscriptTurn(turn) || (turn.ControlKind != nil && *turn.ControlKind != "") || transcriptHasImage(turn) {
		return true
	}
	content := trimTranscriptText(turn.Content)
	if content == "" && len(turn.ToolCalls) == 0 {
		return false
	}
	if turn.Role == RoleSystem && len(turn.ToolCalls) == 0 && transcriptUTF16LengthBelowEight(content) {
		return false
	}
	return true
}

func protectedTranscriptTurn(turn TurnDetail) bool {
	if turn.SourceEntryRef != "" || turn.Provenance != nil || turn.Usage != nil {
		return true
	}
	for _, call := range turn.ToolCalls {
		if call.CallEntryRef != "" || call.ResultEntryRef != "" || call.Usage != nil ||
			call.CallProvenance != nil || call.ResultProvenance != nil {
			return true
		}
	}
	return false
}

func transcriptHasImage(turn TurnDetail) bool {
	for _, part := range turn.ContentParts {
		if part.Type == TurnPartTypeImage && part.Image != nil {
			return true
		}
	}
	return false
}

func duplicateTranscriptTurns(prev, curr TurnDetail) bool {
	return !protectedTranscriptTurn(prev) && !protectedTranscriptTurn(curr) &&
		!transcriptHasImage(prev) && !transcriptHasImage(curr) &&
		prev.Role == curr.Role && prev.Content == curr.Content &&
		trimTranscriptText(prev.Content) != ""
}

// trimTranscriptText uses ECMAScript WhiteSpace and LineTerminator, the set
// used by the reader's String.trim, rather than Go's broader IsSpace set.
func trimTranscriptText(text string) string {
	return strings.TrimFunc(text, func(r rune) bool {
		switch r {
		case '\u0009', '\u000a', '\u000b', '\u000c', '\u000d', '\u0020',
			'\u00a0', '\u1680', '\u2028', '\u2029', '\u202f', '\u205f',
			'\u3000', '\ufeff':
			return true
		default:
			return r >= '\u2000' && r <= '\u200a'
		}
	})
}

// The short-banner rule uses JavaScript string length, measured in UTF-16 code
// units, so four supplementary characters are not mistaken for four units.
func transcriptUTF16LengthBelowEight(text string) bool {
	units := 0
	for _, r := range text {
		units++
		if r > 0xffff {
			units++
		}
		if units >= 8 {
			return false
		}
	}
	return true
}

func transcriptNumberSuffix(n int) string {
	var buf [32]byte
	i := len(buf)
	for n > 0 {
		n--
		i--
		buf[i] = byte('a' + n%26)
		n /= 26
	}
	return string(buf[i:])
}
