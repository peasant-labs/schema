package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"
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

// CutVersion names the producer rules for the turn byte measure, segment cut,
// ResolveTranscriptWindow, ResolvePositionedWindow, cooking, numbering, and
// TranscriptSegmentBudgetBytes.
// Changing served rows, segment boundaries, or display numbers changes this
// identity and adds a new version's cut and numbering fixture expectations.
const CutVersion = "c1"

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
// for s, quotes included. It counts the encoder's escaping, HTML-escaping and
// invalid-UTF-8 replacement exactly as the wire carries them, without
// encoding; FuzzTranscriptJSONStringByteLength holds it equal to json.Marshal. It is the string term of TranscriptTurnByteLength; a server
// that bounds served text against the same measure calls it instead of keeping
// a copy.
func TranscriptJSONStringByteLength(s string) int {
	return transcriptEncodedStringLength(s)
}

// transcriptEncodedStringLength counts, without encoding, the bytes
// encoding/json writes for s with HTML escaping: quotes, two-byte escapes for
// quote, backslash, \b, \f, \n, \r and \t, a six-byte \u escape for other
// control bytes, for <, > and &, for U+2028 and U+2029, and for each invalid
// UTF-8 byte (written as \ufffd). Every other byte is written as is.
func transcriptEncodedStringLength(s string) int {
	n := 2
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			switch {
			case c == '"' || c == '\\' || c == '\b' || c == '\f' || c == '\n' || c == '\r' || c == '\t':
				n += 2
			case c < 0x20 || c == '<' || c == '>' || c == '&':
				n += 6
			default:
				n++
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if (r == utf8.RuneError && size == 1) || r == '\u2028' || r == '\u2029' {
			n += 6
		} else {
			n += size
		}
		i += size
	}
	return n
}

// TranscriptTurnByteLength returns the pre-bound JSON byte length of a turn:
// its content, its text parts, its image metadata, and every tool call's
// arguments and result, measured with the JSON encoder so escaping is counted
// as the wire carries it. It is the one measure of the cut and of a segment's
// byteLength. It is the size before any served-text bound, so an oversized turn
// stays detectable in the cut after its text has been shortened.
func TranscriptTurnByteLength(turn TurnDetail) int {
	return transcriptTurnByteLength(&turn)
}

func transcriptTurnByteLength(turn *TurnDetail) int {
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
// Preview segments are capture-local and need not match the later stored cut.
func CutTranscriptSegments(turns []TurnDetail, budget int) []TranscriptSegmentRef {
	return cutTranscript(len(turns), func(i int) transcriptCutItem {
		return transcriptCutItem{
			index:        turns[i].Index,
			size:         TranscriptTurnByteLength(turns[i]),
			conversation: IsConversationTurn(turns[i].Role, turns[i].EntryType, turns[i].Depth),
		}
	}, budget)
}

// transcriptCutItem is what the cut reads of one turn.
type transcriptCutItem struct {
	index, size  int
	conversation bool
}

// cutTranscript is the one cut rule, shared by CutTranscriptSegments and
// CutTranscriptCookTurns.
func cutTranscript(n int, item func(int) transcriptCutItem, budget int) []TranscriptSegmentRef {
	if n == 0 {
		return []TranscriptSegmentRef{}
	}
	refs := make([]TranscriptSegmentRef, 0, n)
	var current *TranscriptSegmentRef
	conversationTurns := 0
	for i := 0; i < n; i++ {
		turn := item(i)
		if current != nil && current.ByteLength+turn.size > budget {
			refs = append(refs, *current)
			current = nil
		}
		if current == nil {
			current = &TranscriptSegmentRef{
				Index:             len(refs),
				FirstTurn:         turn.index,
				LastTurn:          turn.index,
				ConversationTurns: &ConversationTurnSpan{Before: conversationTurns},
			}
		}
		current.LastTurn = turn.index
		current.ByteLength += turn.size
		if turn.conversation {
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

	firstPos, lastPos := assistantSpanBounds(func(i int) Role { return turns[i].Role }, len(turns), anchorPos, span)

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

// assistantSpanBounds collects up to span assistant rows on each side of an
// anchor. Both window resolvers use this positional rule.
func assistantSpanBounds(roleAt func(int) Role, n, anchor, span int) (first, last int) {
	first = anchor
	seen := 0
	for pos := anchor - 1; pos >= 0; pos-- {
		first = pos
		if roleAt(pos) == RoleAssistant {
			seen++
			if seen == span {
				break
			}
		}
	}
	last = anchor
	seen = 0
	for pos := anchor + 1; pos < n; pos++ {
		last = pos
		if roleAt(pos) == RoleAssistant {
			seen++
			if seen == span {
				break
			}
		}
	}
	return first, last
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
