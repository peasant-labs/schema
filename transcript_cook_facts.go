package schema

import "fmt"

// TranscriptCookTurnFacts are the per-turn values a partition index keeps. Each
// equals the value Detail() gives: Index, Role, EntryType and Depth come from
// the typed, case-insensitive decode, not from the exact-name members.
type TranscriptCookTurnFacts struct {
	Index     int
	Role      Role
	EntryType EntryType
	Depth     int
	// UsagePresent reports a usage record on the turn or on any tool call.
	UsagePresent bool
	// ConversationTurn is IsConversationTurn(Role, EntryType, Depth).
	ConversationTurn bool
	// ByteLength is TranscriptTurnByteLength(Detail()), the cut's measure.
	ByteLength int
}

// Facts returns the turn's index facts without copying its TurnDetail.
func (t TranscriptCookTurn) Facts() TranscriptCookTurnFacts {
	d := &t.detail
	usage := d.Usage != nil
	for i := range d.ToolCalls {
		usage = usage || d.ToolCalls[i].Usage != nil
	}
	return TranscriptCookTurnFacts{
		Index:            d.Index,
		Role:             d.Role,
		EntryType:        d.EntryType,
		Depth:            d.Depth,
		UsagePresent:     usage,
		ConversationTurn: IsConversationTurn(d.Role, d.EntryType, d.Depth),
		ByteLength:       transcriptTurnByteLength(d),
	}
}

// CutTranscriptCookTurns cuts the kept turns of one cooked partition, in
// stored order, with the rule of CutTranscriptSegments. verdicts is the
// CookTranscriptRows result for the same turns. The refs equal
// CutTranscriptSegments over the kept turns' Detail values.
func CutTranscriptCookTurns(turns []TranscriptCookTurn, verdicts []CookedTurn, budget int) ([]TranscriptSegmentRef, error) {
	const where = "transcript cut failed at schema.CutTranscriptCookTurns: "
	if len(verdicts) != len(turns) {
		return nil, fmt.Errorf("%s%d verdicts were supplied for %d turns; no segment was cut; pass the CookTranscriptRows result for the same turns", where, len(verdicts), len(turns))
	}
	kept := make([]int, 0, len(turns))
	for i := range turns {
		if !turns[i].constructed() {
			return nil, fmt.Errorf("%sturn %d was not constructed from served bytes; no segment was cut; call schema.NewTranscriptCookTurns", where, i)
		}
		if verdicts[i].Position != i {
			return nil, fmt.Errorf("%sverdict %d names stored position %d; no segment was cut; pass the CookTranscriptRows result unchanged", where, i, verdicts[i].Position)
		}
		if verdicts[i].Keep {
			kept = append(kept, i)
		}
	}
	return cutTranscript(len(kept), func(i int) transcriptCutItem {
		d := &turns[kept[i]].detail
		return transcriptCutItem{
			index:        d.Index,
			size:         transcriptTurnByteLength(d),
			conversation: IsConversationTurn(d.Role, d.EntryType, d.Depth),
		}
	}, budget), nil
}
