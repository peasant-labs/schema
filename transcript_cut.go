package schema

import (
	"encoding/json"
	"errors"
	"fmt"
)

// This file holds the producer rules of the segmented transcript reads: how
// large a turn is, how a session's turns are cut into segments, and how an
// anchor resolves to a window. They are pure functions over wire types. They
// live in the contract module because every server that serves the reads must
// cut the same turns into the same segments; see the recorded exception in
// AGENTS.md.

// TranscriptSegmentBudgetBytes is the target byte size of one transcript
// segment. A segment fills with whole turns until the next turn would push it
// past this size. A single turn larger than the budget is alone in its segment.
//
// It is a contract value, not a tuning knob: every server cuts with it, so
// changing it moves the segment boundaries of every transcript on every server,
// and a directory cut under the old value no longer matches a segment cut under
// the new one. Change it only as a deliberate contract change.
const TranscriptSegmentBudgetBytes = 1 << 20

// The window span is the number of assistant turns collected on each side of
// the anchor. A caller may narrow it; it cannot widen the design value.
const (
	// TranscriptWindowSpanDefault is the span served when the caller names none.
	TranscriptWindowSpanDefault = 10
	// TranscriptWindowSpanMinimum is the narrowest span a caller may request.
	TranscriptWindowSpanMinimum = 1
	// TranscriptWindowSpanMaximum is the widest span a caller may request.
	TranscriptWindowSpanMaximum = 10
)

// TranscriptJSONStringByteLength is the number of bytes the JSON encoder writes
// for s, quotes included. It is measured with the encoder itself, so escaping,
// HTML-escaping and invalid-UTF-8 replacement are counted exactly as the wire
// carries them. It is the string term of TranscriptTurnByteLength; a server
// that bounds served text against the same measure calls it instead of keeping
// a copy.
func TranscriptJSONStringByteLength(s string) int {
	encoded, err := json.Marshal(s)
	if err != nil {
		// json.Marshal of a string does not fail; a future change that made it
		// fail must not silently under-count the document.
		return len(s) + 2
	}
	return len(encoded)
}

// TranscriptTurnByteLength returns the pre-bound JSON byte length of a turn:
// its content, its text parts, its image metadata, and every tool call's
// arguments and result, measured with the JSON encoder so escaping is counted
// as the wire carries it. It is the one measure of the cut and of a segment's
// byteLength. It is the size before any served-text bound, so an oversized turn
// stays detectable in the cut after its text has been shortened.
func TranscriptTurnByteLength(turn TurnDetail) int {
	total := TranscriptJSONStringByteLength(turn.Content)
	for _, part := range turn.ContentParts {
		if part.Type == TurnPartTypeText {
			total += TranscriptJSONStringByteLength(part.Text)
		} else if part.Type == TurnPartTypeImage && part.Image != nil {
			total += transcriptImagePartByteLength(part.Image)
		}
	}
	for i := range turn.ToolCalls {
		total += TranscriptJSONStringByteLength(turn.ToolCalls[i].Arguments)
		total += TranscriptJSONStringByteLength(turn.ToolCalls[i].Result)
	}
	return total
}

// transcriptImagePartByteLength is the number of bytes the JSON encoder writes
// for an image part's metadata object.
func transcriptImagePartByteLength(image *ImagePart) int {
	encoded, err := json.Marshal(image)
	if err != nil {
		// Cannot happen: ImagePart holds only strings and integers and has no custom marshaler.
		panic(fmt.Sprintf("measure image part: %v", err))
	}
	return len(encoded)
}

// CutTranscriptSegments cuts a session's turns, in order, into segment refs.
// A turn is appended whole or deferred; a turn larger than budget is alone in
// its segment. The function never splits a turn, so it never splits an entry or
// a user turn. Each ref's ByteLength is the sum of its turns'
// TranscriptTurnByteLength, and each ref carries ConversationTurns
// (IsConversationTurn): Before is the session-wide count of conversation turns
// in the earlier segments, so no directory page restarts it. Two servers that
// cut the same turns with the same budget produce the same refs. No turns give
// an empty, non-nil slice.
func CutTranscriptSegments(turns []TurnDetail, budget int) []TranscriptSegmentRef {
	if len(turns) == 0 {
		return []TranscriptSegmentRef{}
	}
	refs := make([]TranscriptSegmentRef, 0, len(turns))
	var current *TranscriptSegmentRef
	conversationTurns := 0
	for i := range turns {
		size := TranscriptTurnByteLength(turns[i])
		if current != nil && current.ByteLength+size > budget {
			refs = append(refs, *current)
			current = nil
		}
		if current == nil {
			current = &TranscriptSegmentRef{
				Index:             len(refs),
				FirstTurn:         turns[i].Index,
				LastTurn:          turns[i].Index,
				ConversationTurns: &ConversationTurnSpan{Before: conversationTurns},
			}
		}
		current.LastTurn = turns[i].Index
		current.ByteLength += size
		if IsConversationTurn(turns[i].Role, turns[i].EntryType, turns[i].Depth) {
			current.ConversationTurns.Count++
			conversationTurns++
		}
	}
	if current != nil {
		refs = append(refs, *current)
	}
	return refs
}

// TranscriptWindowTurn is the identity and role of one turn: all a window
// needs, so a server need not keep turn content to resolve one.
type TranscriptWindowTurn struct {
	Index int
	Role  Role
}

// ErrTranscriptWindowNoUserTurn reports that the turns hold no user turn to
// centre a window on. ResolveTranscriptWindow wraps it with the session and the
// requested anchor; match it with errors.Is.
var ErrTranscriptWindowNoUserTurn = errors.New("the transcript holds no user turn to centre a window on")

// ResolveTranscriptWindow resolves an anchor to a user turn (the anchor itself,
// else the nearest lower user turn, else the nearest higher), collects up to
// span assistant turns on each side, and names the contiguous segment run that
// covers the range. turns are in transcript order and refs are the directory
// cut from the same turns. It never reads turn content. The result passes
// TranscriptWindow.Validate and ValidateTranscriptWindowCoverage; a span outside
// TranscriptWindowSpanMinimum..TranscriptWindowSpanMaximum is refused by the
// former.
func ResolveTranscriptWindow(sessionID SessionID, revision string, requestedAnchor, span int, turns []TranscriptWindowTurn, refs []TranscriptSegmentRef) (TranscriptWindow, error) {
	const where = "transcript window resolution failed at schema.ResolveTranscriptWindow: "
	anchorPos, ok := resolveTranscriptWindowAnchor(turns, requestedAnchor)
	if !ok {
		return TranscriptWindow{}, fmt.Errorf("%ssession %s has no user turn among its %d turns, so a window cannot be centred on turn %d; no window was resolved; a window requires a user turn: %w", where, sessionID, len(turns), requestedAnchor, ErrTranscriptWindowNoUserTurn)
	}

	firstPos := anchorPos
	seen := 0
	for pos := anchorPos - 1; pos >= 0; pos-- {
		firstPos = pos
		if turns[pos].Role == RoleAssistant {
			seen++
			if seen == span {
				break
			}
		}
	}
	lastPos := anchorPos
	seen = 0
	for pos := anchorPos + 1; pos < len(turns); pos++ {
		lastPos = pos
		if turns[pos].Role == RoleAssistant {
			seen++
			if seen == span {
				break
			}
		}
	}

	firstTurn := turns[firstPos].Index
	lastTurn := turns[lastPos].Index
	segments, covered := coveringSegmentRun(refs, firstTurn, lastTurn)
	if !covered {
		return TranscriptWindow{}, fmt.Errorf("%sthe window [%d, %d] of session %s is not covered by the %d segment refs supplied; no window was resolved; pass the directory cut from the same turns, and rebuild it after the session is re-indexed", where, firstTurn, lastTurn, sessionID, len(refs))
	}

	window := TranscriptWindow{
		SessionID:       sessionID,
		Revision:        revision,
		RequestedAnchor: requestedAnchor,
		AnchorTurn:      turns[anchorPos].Index,
		Span:            span,
		FirstTurn:       firstTurn,
		LastTurn:        lastTurn,
		Segments:        segments,
		HasOlder:        firstPos > 0,
		HasNewer:        lastPos < len(turns)-1,
	}
	if err := window.Validate(); err != nil {
		return TranscriptWindow{}, err
	}
	if err := ValidateTranscriptWindowCoverage(window, refs); err != nil {
		return TranscriptWindow{}, err
	}
	return window, nil
}

// resolveTranscriptWindowAnchor returns the position of the user turn a window
// is centred on: the anchor itself when it is a user turn, otherwise the
// nearest lower user turn, otherwise the nearest higher one.
func resolveTranscriptWindowAnchor(turns []TranscriptWindowTurn, anchor int) (int, bool) {
	lower := -1
	higher := -1
	for pos := range turns {
		if turns[pos].Role != RoleUser {
			continue
		}
		if turns[pos].Index == anchor {
			return pos, true
		}
		if turns[pos].Index < anchor {
			lower = pos
			continue
		}
		if higher < 0 {
			higher = pos
		}
	}
	if lower >= 0 {
		return lower, true
	}
	if higher >= 0 {
		return higher, true
	}
	return 0, false
}

// coveringSegmentRun returns the ordered contiguous run of segment ordinals
// from the ref containing firstTurn to the ref containing lastTurn. It reports
// false when either edge lies in no ref. The window resolver and the coverage
// validator share it, so a resolved window always names the run the validator
// expects.
func coveringSegmentRun(refs []TranscriptSegmentRef, firstTurn, lastTurn int) ([]int, bool) {
	first, last := -1, -1
	for i := range refs {
		if refs[i].FirstTurn <= firstTurn && firstTurn <= refs[i].LastTurn {
			first = refs[i].Index
		}
		if refs[i].FirstTurn <= lastTurn && lastTurn <= refs[i].LastTurn {
			last = refs[i].Index
		}
	}
	if first < 0 || last < 0 {
		return nil, false
	}
	run := make([]int, 0, last-first+1)
	for segment := first; segment <= last; segment++ {
		run = append(run, segment)
	}
	return run, true
}
