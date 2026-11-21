package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	schema "github.com/peasant-labs/schema"
)

// observation is everything one capture verdict exposes through the public
// API. Its YAML form is the expected value of the root fixture; its JSON form
// is one differential line.
type observation struct {
	Refusal string             `json:"refusal,omitempty" yaml:"refusal,omitempty"`
	Turn    *int               `json:"turn,omitempty" yaml:"turn,omitempty"`
	Turns   []turnObservation  `json:"turns,omitempty" yaml:"turns,omitempty"`
	Cook    *cookObservation   `json:"cook,omitempty" yaml:"cook,omitempty"`
	Cuts    *cutObservation    `json:"cuts,omitempty" yaml:"cuts,omitempty"`
	Digest  *digestObservation `json:"digest,omitempty" yaml:"-"`
}

type turnObservation struct {
	Raw           string `json:"raw" yaml:"raw"`
	RecordedIndex string `json:"recordedIndex" yaml:"recordedIndex"`
	ByteLength    int    `json:"byteLength" yaml:"byteLength"`
	Detail        string `json:"detail" yaml:"detail"`
	Role          string `json:"role" yaml:"role"`
	EntryType     string `json:"entryType" yaml:"entryType"`
	Depth         int    `json:"depth" yaml:"depth"`
	Usage         bool   `json:"usage" yaml:"usage"`
	Conversation  bool   `json:"conversation" yaml:"conversation"`
}

type cookObservation struct {
	ClaudeCode []string `json:"claudeCode" yaml:"claudeCode,flow"`
	Pi         []string `json:"pi" yaml:"pi,flow"`
}

type cutObservation struct {
	Budget   []string `json:"budget" yaml:"budget,flow"`
	One      []string `json:"one" yaml:"one,flow"`
	Fit      []string `json:"fit,omitempty" yaml:"fit,omitempty,flow"`
	FitBelow []string `json:"fitBelow,omitempty" yaml:"fitBelow,omitempty,flow"`
}

// digestObservation replaces the per-turn parts of a large input's line.
type digestObservation struct {
	TurnCount int    `json:"turnCount"`
	Turns     string `json:"turns"`
	Cook      string `json:"cook"`
	Cuts      string `json:"cuts"`
}

// digestAbove is the turn count above which a line carries digests.
const digestAbove = 64

func sha(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("capture-differential: marshal observation: %v", err))
	}
	return b
}

// observe captures doc and records every public consequence. The adapter
// supplies the two version-specific parts: refusal classification and the
// recorded index.
func observe(doc []byte) (observation, error) {
	turns, err := schema.NewTranscriptCookTurns(doc)
	if err != nil {
		refusal, turn, cerr := classify(err)
		if cerr != nil {
			return observation{}, cerr
		}
		return observation{Refusal: refusal, Turn: &turn}, nil
	}
	o := observation{Turns: []turnObservation{}}
	details := make([]schema.TurnDetail, len(turns))
	for i, t := range turns {
		d := t.Detail()
		details[i] = d
		raw, merr := t.Raw().MarshalJSON()
		if merr != nil {
			return observation{}, merr
		}
		state, index := recordedIndex(t)
		ri := string(state)
		if index != nil {
			ri = fmt.Sprintf("%s:%d", state, *index)
		}
		usage := d.Usage != nil
		for _, tool := range d.ToolCalls {
			usage = usage || tool.Usage != nil
		}
		o.Turns = append(o.Turns, turnObservation{
			Raw: sha(raw), RecordedIndex: ri, ByteLength: schema.TranscriptTurnByteLength(d),
			Detail: sha(mustJSON(d)), Role: string(d.Role), EntryType: string(d.EntryType), Depth: d.Depth,
			Usage: usage, Conversation: schema.IsConversationTurn(d.Role, d.EntryType, d.Depth),
		})
	}
	o.Cook = &cookObservation{ClaudeCode: cook(turns, schema.HarnessClaudeCode), Pi: cook(turns, schema.HarnessPi)}
	o.Cuts = &cutObservation{
		Budget: cuts(details, schema.TranscriptSegmentBudgetBytes),
		One:    cuts(details, 1),
	}
	if len(details) >= 2 {
		fit := schema.TranscriptTurnByteLength(details[0]) + schema.TranscriptTurnByteLength(details[1])
		o.Cuts.Fit = cuts(details, fit)
		o.Cuts.FitBelow = cuts(details, fit-1)
	}
	if len(turns) > digestAbove {
		o.Digest = &digestObservation{TurnCount: len(turns), Turns: sha(mustJSON(o.Turns)), Cook: sha(mustJSON(o.Cook)), Cuts: sha(mustJSON(o.Cuts))}
		o.Turns, o.Cook, o.Cuts = nil, nil, nil
	}
	return o, nil
}

func cook(turns []schema.TranscriptCookTurn, harness schema.Harness) []string {
	rows, err := schema.CookTranscriptRows(turns, harness)
	if err != nil {
		return []string{"uncookable"}
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		if !r.Keep {
			out[i] = fmt.Sprintf("drop %d", r.ServedPosition)
			continue
		}
		out[i] = fmt.Sprintf("keep %d %s", r.ServedPosition, r.DisplayNumber)
	}
	return out
}

func cuts(details []schema.TurnDetail, budget int) []string {
	refs := schema.CutTranscriptSegments(details, budget)
	out := make([]string, len(refs))
	for i, r := range refs {
		span := "none"
		if r.ConversationTurns != nil {
			span = fmt.Sprintf("%d+%d", r.ConversationTurns.Before, r.ConversationTurns.Count)
		}
		out[i] = fmt.Sprintf("%d %d-%d %dB %s", r.Index, r.FirstTurn, r.LastTurn, r.ByteLength, span)
	}
	return out
}
