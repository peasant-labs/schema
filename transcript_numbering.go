package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

// ErrTranscriptUncookable identifies a partition the whole-document reader's
// evidence gate or duplicate comparison would also refuse.
var ErrTranscriptUncookable = errors.New("transcript cooking refused this stored partition")

type transcriptMemberState uint8

const (
	transcriptMemberAbsent transcriptMemberState = iota
	transcriptMemberNull
	transcriptMemberValue
)

type transcriptStringMember struct {
	state transcriptMemberState
	value string
}

type transcriptCookFacts struct {
	role, content transcriptStringMember
	depth         int
	protected     bool
	control       bool
	image         bool
	tools         int
}

// CookTranscriptRows filters, deduplicates and numbers one partition using the
// same exact-name rule as the whole-document reader. Pi skips the legacy
// prefilter. A refusal returns no partial verdicts.
func CookTranscriptRows(turns []TranscriptCookTurn, harness Harness) ([]CookedTurn, error) {
	facts := make([]transcriptCookFacts, len(turns))
	for i := range turns {
		if turns[i].members == nil {
			return nil, fmt.Errorf("transcript cooking failed at schema.CookTranscriptRows: turn %d was not constructed from served bytes; no verdict was produced; call schema.NewTranscriptCookTurns", i)
		}
		var err error
		facts[i], err = transcriptFacts(turns[i].members, harness != HarnessPi)
		if err != nil {
			return nil, fmt.Errorf("%w: evidence validation failed at turn %d; no verdict was produced; repair the recorded evidence before retrying: %v", ErrTranscriptUncookable, i, err)
		}
	}
	result := make([]CookedTurn, len(turns))
	kept := make([]int, 0, len(turns))
	for i := range turns {
		result[i] = CookedTurn{Position: i, ServedPosition: -1}
		if harness != HarnessPi {
			if !keepTranscriptTurn(facts[i]) {
				continue
			}
			if len(kept) != 0 {
				prev := kept[len(kept)-1]
				duplicate, err := duplicateTranscriptTurns(facts[prev], facts[i])
				if err != nil {
					return nil, fmt.Errorf("%w: duplicate comparison failed at turns %d and %d because equal content is absent or null; no verdict was produced; repair the recorded turns before retrying", ErrTranscriptUncookable, prev, i)
				}
				if duplicate {
					if facts[i].tools > 0 && facts[prev].tools == 0 {
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
		turn := facts[stored]
		startsTask := turn.role == (transcriptStringMember{state: transcriptMemberValue, value: string(RoleUser)}) &&
			turn.depth == 0 && turn.content.state == transcriptMemberValue && trimTranscriptText(turn.content.value) != ""
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
	return result, nil
}

func keepTranscriptTurn(turn transcriptCookFacts) bool {
	if turn.protected || turn.control || turn.image {
		return true
	}
	content := trimTranscriptText(turn.content.value)
	if content == "" && turn.tools == 0 {
		return false
	}
	if turn.role == (transcriptStringMember{state: transcriptMemberValue, value: string(RoleSystem)}) && turn.tools == 0 && transcriptUTF16LengthBelowEight(content) {
		return false
	}
	return true
}

func duplicateTranscriptTurns(prev, curr transcriptCookFacts) (bool, error) {
	if prev.protected || curr.protected || prev.image || curr.image || prev.role != curr.role || prev.content != curr.content {
		return false, nil
	}
	if prev.content.state != transcriptMemberValue {
		return false, ErrTranscriptUncookable
	}
	return trimTranscriptText(prev.content.value) != "", nil
}

func transcriptMemberString(members map[string]json.RawMessage, name string) transcriptStringMember {
	raw, ok := members[name]
	if !ok {
		return transcriptStringMember{state: transcriptMemberAbsent}
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return transcriptStringMember{state: transcriptMemberNull}
	}
	var value string
	_ = json.Unmarshal(raw, &value) // The constructor's typed decode checked its type.
	return transcriptStringMember{state: transcriptMemberValue, value: value}
}

func transcriptMemberSet(members map[string]json.RawMessage, name string) bool {
	raw, ok := members[name]
	return ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func transcriptMemberArray(members map[string]json.RawMessage, name string) []json.RawMessage {
	var value []json.RawMessage
	_ = json.Unmarshal(members[name], &value) // The constructor checked known member types.
	return value
}

func transcriptFacts(members map[string]json.RawMessage, validateEvidence bool) (transcriptCookFacts, error) {
	facts := transcriptCookFacts{role: transcriptMemberString(members, "role"), content: transcriptMemberString(members, "content")}
	_ = json.Unmarshal(members["depth"], &facts.depth)
	facts.control = transcriptMemberString(members, "controlKind").value != ""
	facts.protected = transcriptMemberString(members, "sourceEntryRef").value != "" || transcriptMemberSet(members, "usage")
	if _, present := members["provenance"]; present {
		facts.protected = true
	}
	if validateEvidence {
		if err := transcriptValidateEvidence(members, "sourceEntryRef", "provenance"); err != nil {
			return facts, err
		}
	}
	for _, raw := range transcriptMemberArray(members, "contentParts") {
		part, err := transcriptRawMembers(raw)
		if err != nil { // A null part is not an image.
			continue
		}
		if transcriptMemberString(part, "type").value == "image" && transcriptMemberSet(part, "image") {
			facts.image = true
		}
	}
	tools := transcriptMemberArray(members, "toolCalls")
	facts.tools = len(tools)
	for _, raw := range tools {
		tool, err := transcriptRawMembers(raw)
		if err != nil {
			if validateEvidence {
				return facts, fmt.Errorf("a tool is not an object")
			}
			continue
		}
		if validateEvidence {
			if err := transcriptValidateEvidence(tool, "callEntryRef", "callProvenance"); err != nil {
				return facts, err
			}
			if err := transcriptValidateEvidence(tool, "resultEntryRef", "resultProvenance"); err != nil {
				return facts, err
			}
		}
		_, callProvenance := tool["callProvenance"]
		_, resultProvenance := tool["resultProvenance"]
		if transcriptMemberString(tool, "callEntryRef").value != "" || transcriptMemberString(tool, "resultEntryRef").value != "" ||
			transcriptMemberSet(tool, "usage") || callProvenance || resultProvenance {
			facts.protected = true
		}
	}
	return facts, nil
}

func transcriptValidateEvidence(members map[string]json.RawMessage, refName, provenanceName string) error {
	ref := transcriptMemberString(members, refName)
	if ref.state == transcriptMemberNull {
		return fmt.Errorf("%s is null", refName)
	}
	if ref.state == transcriptMemberValue && ref.value != "" {
		if err := SourceEntryRef(ref.value).Validate(); err != nil {
			return fmt.Errorf("%s is invalid", refName)
		}
	}
	raw, present := members[provenanceName]
	if !present {
		return nil
	}
	fields, err := transcriptRawMembers(raw)
	if err != nil {
		return fmt.Errorf("%s is not an evidence object", provenanceName)
	}
	for name := range fields {
		switch name {
		case "origin", "actor", "delivery", "ownership", "evidence", "inputModality", "submissionRef":
		default:
			return fmt.Errorf("%s has an unknown member", provenanceName)
		}
	}
	if rawRef, ok := fields["submissionRef"]; ok && bytes.Equal(bytes.TrimSpace(rawRef), []byte("null")) {
		return fmt.Errorf("%s has a null submission reference", provenanceName)
	}
	var provenance ContentProvenance
	if err := json.Unmarshal(raw, &provenance); err != nil {
		return fmt.Errorf("%s cannot be decoded", provenanceName)
	}
	if err := provenance.Validate(); err != nil {
		return fmt.Errorf("%s is invalid", provenanceName)
	}
	return nil
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
