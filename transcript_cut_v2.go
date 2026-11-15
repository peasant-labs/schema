package schema

import (
	"errors"
	"fmt"
)

// TranscriptWindowTurnV2 is the identity and role of one served row. A nil
// recorded index means the row has no comparable recorded identity.
type TranscriptWindowTurnV2 struct {
	Position      int
	Role          Role
	Depth         int
	RecordedIndex *int
}

// TranscriptWindowV2Input supplies one partition's served rows and a
// contiguous ascending interval of absolute segment ordinals. A server may
// supply its whole directory or adjacent pages concatenated in ordinal order.
type TranscriptWindowV2Input struct {
	Context         TranscriptReadContext
	Partition       TranscriptSegmentPartition
	RequestedAnchor *int
	Span            int
	Turns           []TranscriptWindowTurnV2
	Refs            []TranscriptSegmentRefV2
}

// ErrTranscriptWindowAnchorUnresolvable reports that user rows exist but none
// has a recorded index that an anchored request can compare.
var ErrTranscriptWindowAnchorUnresolvable = errors.New("the transcript has no user turn with a comparable recorded index")

// ErrTranscriptWindowEmptyPartition reports that a partition has no rows to anchor.
var ErrTranscriptWindowEmptyPartition = errors.New("the transcript partition contains no served rows")

// ResolveTranscriptWindowV2 resolves one partition's window in served order.
// Recorded indexes are comparison keys only; positions and ordinals locate the
// resulting run. The resolver never reads turn content.
func ResolveTranscriptWindowV2(in TranscriptWindowV2Input) (TranscriptWindowV2, error) {
	const where = "transcript window resolution failed at schema.ResolveTranscriptWindowV2: "
	span := in.Span
	if span == 0 {
		span = TranscriptWindowSpanDefault
	}
	if span < TranscriptWindowSpanMinimum || span > TranscriptWindowSpanMaximum {
		return TranscriptWindowV2{}, fmt.Errorf("%sspan %d is outside %d..%d; the window cannot be resolved; request a supported assistant-turn span", where, span, TranscriptWindowSpanMinimum, TranscriptWindowSpanMaximum)
	}
	if len(in.Turns) == 0 {
		return TranscriptWindowV2{}, fmt.Errorf("%sthe partition has no served rows; a window cannot be anchored; open a nonempty partition: %w", where, ErrTranscriptWindowEmptyPartition)
	}
	if err := validateTranscriptV2Refs(in.Refs, in.Context.TotalSegments, where); err != nil {
		return TranscriptWindowV2{}, err
	}
	for i, turn := range in.Turns {
		if turn.Position != i {
			return TranscriptWindowV2{}, fmt.Errorf("%srow %d carries position %d; served positions must be consecutive from zero; pass the ordered partition rows", where, i, turn.Position)
		}
	}
	if in.RequestedAnchor != nil && *in.RequestedAnchor < 0 {
		return TranscriptWindowV2{}, fmt.Errorf("%sthe requested anchor is negative; recorded indexes are nonnegative; request a valid recorded index", where)
	}
	anchor, err := transcriptWindowV2Anchor(in.Turns, in.RequestedAnchor)
	if err != nil {
		return TranscriptWindowV2{}, fmt.Errorf("%s%w", where, err)
	}
	first, last := assistantSpanBounds(func(i int) Role { return in.Turns[i].Role }, len(in.Turns), anchor, span)
	var requested *int
	if in.RequestedAnchor != nil {
		for i, turn := range in.Turns {
			if turn.RecordedIndex != nil && *turn.RecordedIndex == *in.RequestedAnchor {
				requested = new(int)
				*requested = i
				if i < first {
					first = i
				}
				if i > last {
					last = i
				}
				break
			}
		}
	}
	run, firstOrdinal, lastOrdinal, err := transcriptWindowCoveringRun(in, first, last)
	if err != nil {
		return TranscriptWindowV2{}, err
	}
	anchorRef, ok := transcriptRefContaining(in.Refs, anchor)
	if !ok {
		return TranscriptWindowV2{}, fmt.Errorf("%sno supplied ref covers the anchor position; the directory disagrees with the rows; pass the directory built from this partition", where)
	}
	firstRef, _ := transcriptRefByOrdinal(in.Refs, firstOrdinal)
	lastRef, _ := transcriptRefByOrdinal(in.Refs, lastOrdinal)
	window := TranscriptWindowV2{
		Context:           in.Context,
		Partition:         in.Partition,
		RequestedAnchor:   in.RequestedAnchor,
		AnchorTurn:        in.Turns[anchor].RecordedIndex,
		RequestedPosition: requested,
		AnchorPosition:    anchor,
		AnchorSegment:     anchorRef.Index,
		Span:              span,
		FirstPosition:     firstRef.SourceFirstPosition,
		LastPosition:      lastRef.SourceLastPosition,
		Segments:          run,
		HasOlder:          firstOrdinal > 0,
		HasNewer:          lastOrdinal < in.Context.TotalSegments-1,
	}
	if err := ValidateTranscriptWindowV2Coverage(window, in.Refs); err != nil {
		return TranscriptWindowV2{}, err
	}
	return window, nil
}

func transcriptWindowV2Anchor(turns []TranscriptWindowTurnV2, requested *int) (int, error) {
	firstUser, firstRoot := -1, -1
	if requested == nil {
		for i, turn := range turns {
			if turn.Role == RoleUser {
				if firstUser < 0 {
					firstUser = i
				}
				if turn.Depth == 0 && firstRoot < 0 {
					firstRoot = i
				}
			}
		}
		if firstRoot >= 0 {
			return firstRoot, nil
		}
		if firstUser >= 0 {
			return firstUser, nil
		}
		return 0, nil
	}
	lower, higher, hasUser := -1, -1, false
	for i, turn := range turns {
		if turn.Role != RoleUser {
			continue
		}
		hasUser = true
		if turn.RecordedIndex == nil {
			continue
		}
		switch {
		case *turn.RecordedIndex == *requested:
			return i, nil
		case *turn.RecordedIndex < *requested:
			lower = i
		case higher < 0:
			higher = i
		}
	}
	if lower >= 0 {
		return lower, nil
	}
	if higher >= 0 {
		return higher, nil
	}
	if !hasUser {
		return 0, ErrTranscriptWindowNoUserTurn
	}
	return 0, ErrTranscriptWindowAnchorUnresolvable
}

func validateTranscriptV2Refs(refs []TranscriptSegmentRefV2, total int, where string) error {
	if len(refs) == 0 {
		return fmt.Errorf("%sno refs were supplied; a window cannot be resolved; pass the partition's directory", where)
	}
	if total < 1 {
		return fmt.Errorf("%sthe partition has rows but the context says no segments; the directory cannot be placed; emit the partition's true segment count", where)
	}
	return validateTranscriptRefChain(refs, refs[0].Index, total, where, transcriptRefChainWindow, false)
}

type transcriptRefChainContext uint8

const (
	transcriptRefChainDirectory transcriptRefChainContext = iota
	transcriptRefChainWindow
)

// validateTranscriptRefChain is the one adjacency rule for directory pages and
// resolver intervals. Their diagnostics differ because a page has a prescribed
// starting ordinal while a resolver can accept any adjacent page interval.
func validateTranscriptRefChain(refs []TranscriptSegmentRefV2, firstOrdinal, total int, where string, context transcriptRefChainContext, requirePartitionStart bool) error {
	for i, ref := range refs {
		if err := ref.Validate(); err != nil {
			return err
		}
		if ref.Index >= total || ref.Index != firstOrdinal+i {
			if context == transcriptRefChainDirectory {
				return fmt.Errorf("%sref %d has ordinal %d; refs must be consecutive from page ordinal %d and below totalSegments %d; rebuild the directory", where, i, ref.Index, firstOrdinal, total)
			}
			if ref.Index >= total {
				return fmt.Errorf("%ssegment %d is outside totalSegments %d; the ref names an ordinal the partition does not have; rebuild the directory", where, ref.Index, total)
			}
			return fmt.Errorf("%ssegment ordinals %d and %d are not consecutive; a host may have concatenated non-adjacent or reordered pages; pass adjacent pages in ordinal order", where, refs[i-1].Index, ref.Index)
		}
		if i == 0 {
			if requirePartitionStart && ref.SourceFirstPosition != 0 {
				return fmt.Errorf("%sthe first unpaged ref starts after position zero; the served partition would have a gap; rebuild the directory", where)
			}
			continue
		}
		prev := refs[i-1]
		if ref.SourceFirstPosition-prev.SourceLastPosition != 1 {
			if context == transcriptRefChainDirectory {
				return fmt.Errorf("%sref %d does not begin immediately after the preceding served range; rows would be skipped or repeated; rebuild the directory", where, ref.Index)
			}
			return fmt.Errorf("%ssegment %d starts at position %d, not after segment %d ending at %d; the refs and rows disagree; rebuild the directory from the rows this window was resolved on", where, ref.Index, ref.SourceFirstPosition, prev.Index, prev.SourceLastPosition)
		}
	}
	return nil
}

func transcriptRefByOrdinal(refs []TranscriptSegmentRefV2, ordinal int) (TranscriptSegmentRefV2, bool) {
	for i := range refs {
		if refs[i].Index == ordinal {
			return refs[i], true
		}
	}
	return TranscriptSegmentRefV2{}, false
}

func transcriptRefContaining(refs []TranscriptSegmentRefV2, position int) (TranscriptSegmentRefV2, bool) {
	for i := range refs {
		if refs[i].SourceFirstPosition <= position && position <= refs[i].SourceLastPosition {
			return refs[i], true
		}
	}
	return TranscriptSegmentRefV2{}, false
}

func transcriptWindowCoveringRun(in TranscriptWindowV2Input, first, last int) ([]int, int, int, error) {
	const where = "transcript window resolution failed at schema.ResolveTranscriptWindowV2: "
	firstRef, ok := transcriptRefContaining(in.Refs, first)
	if !ok {
		return nil, 0, 0, fmt.Errorf("%sthe supplied refs hold no segment covering position %d, the window's first edge (supplied interval: ordinals %d..%d of %d segments); pass the partition's whole directory, or the adjacent pages that contain it; the window was not resolved", where, first, in.Refs[0].Index, in.Refs[len(in.Refs)-1].Index, in.Context.TotalSegments)
	}
	lastRef, ok := transcriptRefContaining(in.Refs, last)
	if !ok {
		return nil, 0, 0, fmt.Errorf("%sthe supplied refs hold no segment covering position %d, the window's last edge (supplied interval: ordinals %d..%d of %d segments); pass the partition's whole directory, or the adjacent pages that contain it; the window was not resolved", where, last, in.Refs[0].Index, in.Refs[len(in.Refs)-1].Index, in.Context.TotalSegments)
	}
	firstOrdinal, lastOrdinal := firstRef.Index, lastRef.Index
	run := make([]int, 0, lastOrdinal-firstOrdinal+1)
	for ordinal := firstOrdinal; ordinal <= lastOrdinal; ordinal++ {
		run = append(run, ordinal)
	}
	return run, firstOrdinal, lastOrdinal, nil
}

// Validate checks the window's own coordinates and edge flags.
func (w TranscriptWindowV2) Validate() error {
	return w.validateWindowShape(true)
}

// validateWindowShape validates the requested position numerically for the
// standalone wire check. Coverage instead checks its corresponding ref.
func (w TranscriptWindowV2) validateWindowShape(checkRequestedBounds bool) error {
	const where = "v2 transcript window validation failed at schema.TranscriptWindowV2.Validate: "
	if err := w.Context.Validate(); err != nil {
		return err
	}
	if err := w.Partition.Validate(); err != nil {
		return err
	}
	if w.Span < TranscriptWindowSpanMinimum || w.Span > TranscriptWindowSpanMaximum ||
		w.FirstPosition < 0 || w.AnchorPosition < w.FirstPosition || w.LastPosition < w.AnchorPosition ||
		(w.RequestedAnchor != nil && *w.RequestedAnchor < 0) || (w.AnchorTurn != nil && *w.AnchorTurn < 0) ||
		(w.RequestedPosition != nil && (w.RequestedAnchor == nil ||
			(checkRequestedBounds && (*w.RequestedPosition < w.FirstPosition || *w.RequestedPosition > w.LastPosition)))) {
		return fmt.Errorf("%sspan, anchor, requested position, or range is invalid; the window cannot be placed; emit a supported span and positions inside the run", where)
	}
	if len(w.Segments) == 0 {
		return fmt.Errorf("%sthe segment run is empty; no window content can be loaded; emit the covering ordinal run", where)
	}
	foundAnchor := false
	for i, ordinal := range w.Segments {
		if ordinal < 0 || ordinal >= w.Context.TotalSegments || (i > 0 && ordinal-w.Segments[i-1] != 1) {
			return fmt.Errorf("%ssegment ordinal %d is outside the total or not consecutive; the run cannot be loaded; emit consecutive valid ordinals", where, ordinal)
		}
		foundAnchor = foundAnchor || ordinal == w.AnchorSegment
	}
	if !foundAnchor || w.HasOlder != (w.Segments[0] > 0) ||
		w.HasNewer != (w.Segments[len(w.Segments)-1] < w.Context.TotalSegments-1) {
		return fmt.Errorf("%sthe anchor segment or edge flags disagree with the run; the viewer cannot place boundaries; emit flags from absolute ordinals and the partition total", where)
	}
	return nil
}

// ValidateTranscriptWindowV2Coverage checks the window against the refs it was
// resolved on, including the anchor and an optional requested-position widening.
func ValidateTranscriptWindowV2Coverage(window TranscriptWindowV2, refs []TranscriptSegmentRefV2) error {
	const where = "v2 transcript window coverage failed at schema.ValidateTranscriptWindowV2Coverage: "
	if err := window.validateWindowShape(false); err != nil {
		return err
	}
	if err := validateTranscriptV2Refs(refs, window.Context.TotalSegments, where); err != nil {
		return err
	}
	first, firstOK := transcriptRefContaining(refs, window.FirstPosition)
	last, lastOK := transcriptRefContaining(refs, window.LastPosition)
	anchor, anchorOK := transcriptRefContaining(refs, window.AnchorPosition)
	if !firstOK || !lastOK || !anchorOK || first.SourceFirstPosition != window.FirstPosition ||
		last.SourceLastPosition != window.LastPosition || anchor.Index != window.AnchorSegment ||
		window.Segments[0] != first.Index || window.Segments[len(window.Segments)-1] != last.Index ||
		len(window.Segments)-1 != last.Index-first.Index {
		return fmt.Errorf("%sthe refs do not cover the named bounds, anchor, and contiguous run; the directory and window disagree; resolve the window with the same partition refs", where)
	}
	if window.RequestedPosition != nil {
		ref, ok := transcriptRefContaining(refs, *window.RequestedPosition)
		if !ok || ref.Index < first.Index || ref.Index > last.Index {
			return fmt.Errorf("%sthe requested position is not covered by the widened run; the target cannot load; resolve with the refs covering the requested row", where)
		}
	}
	return nil
}
