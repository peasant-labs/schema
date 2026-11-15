package schema

import (
	"fmt"
	"regexp"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// TranscriptReadTransportKind identifies the host-owned route identity.
type TranscriptReadTransportKind string

const (
	TranscriptReadTransportLocalSession      TranscriptReadTransportKind = "local_session"
	TranscriptReadTransportVillageTranscript TranscriptReadTransportKind = "village_transcript"
)

var AllTranscriptReadTransportKinds = []TranscriptReadTransportKind{
	TranscriptReadTransportLocalSession, TranscriptReadTransportVillageTranscript,
}

func (v TranscriptReadTransportKind) IsValid() bool {
	return inSet(v, AllTranscriptReadTransportKinds)
}
func (TranscriptReadTransportKind) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Transcript Read Transport Kind", AllTranscriptReadTransportKinds)
}

// TranscriptReadTransport names exactly one route identity. A content session
// identifier is not a route identity and is carried separately in the context.
type TranscriptReadTransport struct {
	Kind         TranscriptReadTransportKind `json:"kind"`
	SessionID    *SessionID                  `json:"sessionId,omitempty"`
	TranscriptID *VillageUUID                `json:"transcriptId,omitempty"`
}

// Validate refuses a missing, ambiguous, mismatched, or malformed route identity.
func (t TranscriptReadTransport) Validate() error {
	const where = "transcript transport validation failed at schema.TranscriptReadTransport.Validate: "
	switch t.Kind {
	case TranscriptReadTransportLocalSession:
		if t.SessionID == nil || t.TranscriptID != nil {
			return fmt.Errorf("%slocal_session requires only sessionId; a route cannot bind an absent or mixed identity; emit exactly the matching identifier", where)
		}
		if _, err := NewSessionID(string(*t.SessionID)); err != nil {
			return fmt.Errorf("%sthe sessionId is not a valid local session identifier; the read cannot be routed; emit a canonical session identifier", where)
		}
	case TranscriptReadTransportVillageTranscript:
		if t.TranscriptID == nil || t.SessionID != nil {
			return fmt.Errorf("%svillage_transcript requires only transcriptId; a route cannot bind an absent or mixed identity; emit exactly the matching identifier", where)
		}
		if !sessionIDUUID.MatchString(string(*t.TranscriptID)) {
			return fmt.Errorf("%sthe transcriptId is not a canonical Village UUID; the read cannot be routed; emit the lowercase UUID", where)
		}
	default:
		return fmt.Errorf("%skind is outside the transport set; the read cannot be routed; emit a known transport kind", where)
	}
	return nil
}

// TranscriptReadValidation describes the boundary that accepted the served body.
type TranscriptReadValidation string

const (
	TranscriptReadPublishedStrict TranscriptReadValidation = "published_strict"
	TranscriptReadServedLegacy    TranscriptReadValidation = "served_legacy"
	TranscriptReadMigrated        TranscriptReadValidation = "migrated"
)

var AllTranscriptReadValidations = []TranscriptReadValidation{
	TranscriptReadPublishedStrict, TranscriptReadServedLegacy, TranscriptReadMigrated,
}

func (v TranscriptReadValidation) IsValid() bool { return inSet(v, AllTranscriptReadValidations) }
func (TranscriptReadValidation) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Transcript Read Validation", AllTranscriptReadValidations)
}

// TranscriptNativePresence states whether native records exist or are withheld.
type TranscriptNativePresence string

const (
	TranscriptNativeAbsent   TranscriptNativePresence = "absent"
	TranscriptNativePresent  TranscriptNativePresence = "present"
	TranscriptNativeWithheld TranscriptNativePresence = "withheld"
)

var AllTranscriptNativePresences = []TranscriptNativePresence{
	TranscriptNativeAbsent, TranscriptNativePresent, TranscriptNativeWithheld,
}

func (v TranscriptNativePresence) IsValid() bool { return inSet(v, AllTranscriptNativePresences) }
func (TranscriptNativePresence) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Transcript Native Presence", AllTranscriptNativePresences)
}

// TranscriptUsagePresence states a partition's usage evidence availability.
type TranscriptUsagePresence string

const (
	TranscriptUsagePresent TranscriptUsagePresence = "present"
	TranscriptUsageEmpty   TranscriptUsagePresence = "empty"
	TranscriptUsageUnknown TranscriptUsagePresence = "unknown"
)

var AllTranscriptUsagePresences = []TranscriptUsagePresence{
	TranscriptUsagePresent, TranscriptUsageEmpty, TranscriptUsageUnknown,
}

func (v TranscriptUsagePresence) IsValid() bool { return inSet(v, AllTranscriptUsagePresences) }
func (TranscriptUsagePresence) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Transcript Usage Presence", AllTranscriptUsagePresences)
}

// TranscriptReadContext carries the authorized facts repeated on each read.
type TranscriptReadContext struct {
	Transport              TranscriptReadTransport  `json:"transport"`
	ContentSessionID       *string                  `json:"contentSessionId,omitempty"`
	Revision               string                   `json:"revision"`
	Representation         string                   `json:"representation"`
	Harness                Harness                  `json:"harness,omitempty"`
	Model                  string                   `json:"model,omitempty"`
	ThinkingLevel          ThinkingLevel            `json:"thinkingLevel,omitempty"`
	ThinkingLevelRaw       ThinkingLevelRaw         `json:"thinkingLevelRaw,omitempty"`
	Validation             TranscriptReadValidation `json:"validation"`
	NativePresence         TranscriptNativePresence `json:"nativePresence"`
	TotalTurns             int                      `json:"totalTurns"`
	TotalSegments          int                      `json:"totalSegments"`
	TotalConversationTurns *int                     `json:"totalConversationTurns,omitempty"`
	ContentHash            string                   `json:"contentHash,omitempty"`
}

// Validate checks the context's route, identity, closed facts, and counts.
func (c TranscriptReadContext) Validate() error {
	const where = "transcript context validation failed at schema.TranscriptReadContext.Validate: "
	if err := c.Transport.Validate(); err != nil {
		return err
	}
	if !isDecimalRevision(c.Revision) || c.Representation == "" {
		return fmt.Errorf("%srevision must be decimal and representation must be present; responses cannot be joined across generations otherwise; emit both identities", where)
	}
	if !c.Validation.IsValid() || !c.NativePresence.IsValid() {
		return fmt.Errorf("%svalidation or nativePresence is outside its closed set; the read cannot be interpreted; emit a known value", where)
	}
	if c.TotalTurns < 0 || c.TotalSegments < 0 || (c.TotalConversationTurns != nil && *c.TotalConversationTurns < 0) {
		return fmt.Errorf("%sa total is negative; a partition cannot contain a negative number of rows; emit nonnegative totals", where)
	}
	return nil
}

// TranscriptSegmentPartitionKind distinguishes main and earlier history.
type TranscriptSegmentPartitionKind string

const (
	TranscriptPartitionMain    TranscriptSegmentPartitionKind = "main"
	TranscriptPartitionHistory TranscriptSegmentPartitionKind = "history"
)

var AllTranscriptSegmentPartitionKinds = []TranscriptSegmentPartitionKind{
	TranscriptPartitionMain, TranscriptPartitionHistory,
}

func (v TranscriptSegmentPartitionKind) IsValid() bool {
	return inSet(v, AllTranscriptSegmentPartitionKinds)
}
func (TranscriptSegmentPartitionKind) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Transcript Segment Partition Kind", AllTranscriptSegmentPartitionKinds)
}

// TranscriptSegmentPartition identifies a main partition or a history section.
type TranscriptSegmentPartition struct {
	Kind         TranscriptSegmentPartitionKind `json:"kind"`
	SectionIndex *int                           `json:"sectionIndex,omitempty"`
}

func (p TranscriptSegmentPartition) Validate() error {
	const where = "transcript partition validation failed at schema.TranscriptSegmentPartition.Validate: "
	switch p.Kind {
	case TranscriptPartitionMain:
		if p.SectionIndex != nil {
			return fmt.Errorf("%smain carries sectionIndex; it cannot name a history section; omit sectionIndex for main", where)
		}
	case TranscriptPartitionHistory:
		if p.SectionIndex == nil || *p.SectionIndex < 0 {
			return fmt.Errorf("%shistory requires a nonnegative sectionIndex; the section cannot be selected otherwise; provide its index", where)
		}
	default:
		return fmt.Errorf("%skind is outside the partition set; the read cannot select a partition; emit main or history", where)
	}
	return nil
}

// TranscriptHistorySection holds one history section's manifest facts.
type TranscriptHistorySection struct {
	Index                  int                     `json:"index"`
	State                  string                  `json:"state,omitempty"`
	TotalTurns             int                     `json:"totalTurns"`
	TotalConversationTurns *int                    `json:"totalConversationTurns,omitempty"`
	UsagePresent           TranscriptUsagePresence `json:"usagePresent"`
}

// TranscriptSegmentRefV2 bounds one segment in the partition's served order.
// FirstTurn and LastTurn are recorded-index facts and are both present or absent.
type TranscriptSegmentRefV2 struct {
	Index               int                   `json:"index"`
	SourceFirstPosition int                   `json:"sourceFirstPosition" minimum:"0"`
	SourceLastPosition  int                   `json:"sourceLastPosition" minimum:"0"`
	FirstTurn           *int                  `json:"firstTurn,omitempty" minimum:"0" nullable:"false"`
	LastTurn            *int                  `json:"lastTurn,omitempty" minimum:"0" nullable:"false"`
	ByteLength          int                   `json:"byteLength"`
	ConversationTurns   *ConversationTurnSpan `json:"conversationTurns,omitempty"`
}

func (r TranscriptSegmentRefV2) Validate() error {
	const where = "v2 transcript ref validation failed at schema.TranscriptSegmentRefV2.Validate: "
	if r.Index < 0 || r.SourceFirstPosition < 0 || r.SourceLastPosition < r.SourceFirstPosition || r.ByteLength < 0 {
		return fmt.Errorf("%sthe ordinal, served-position range, or byte length is invalid; the segment cannot be located; emit nonnegative ascending bounds and byte length", where)
	}
	if (r.FirstTurn == nil) != (r.LastTurn == nil) ||
		(r.FirstTurn != nil && (*r.FirstTurn < 0 || *r.LastTurn < 0)) {
		return fmt.Errorf("%sfirstTurn and lastTurn must both be nonnegative or both absent; a half range cannot describe recorded indexes; emit both or neither", where)
	}
	return nil
}

var transcriptDisplayNumberPattern = regexp.MustCompile(`^[1-9][0-9]*[a-z]*$`)

// TranscriptSegmentRow holds one served row. Turn is the source JSON object;
// recordedIndexState and recordedIndex are its only recorded-index statement.
type TranscriptSegmentRow struct {
	Position           int                          `json:"position"`
	RecordedIndexState TranscriptRecordedIndexState `json:"recordedIndexState"`
	RecordedIndex      *int                         `json:"recordedIndex,omitempty"`
	DisplayNumber      string                       `json:"displayNumber,omitempty" minLength:"1" pattern:"^[1-9][0-9]*[a-z]*$"`
	Turn               TranscriptV2Turn             `json:"turn"`
}

func (r TranscriptSegmentRow) Validate() error {
	const where = "v2 transcript row validation failed at schema.TranscriptSegmentRow.Validate: "
	if r.Position < 0 || !r.RecordedIndexState.IsValid() ||
		(r.RecordedIndex != nil) != (r.RecordedIndexState == TranscriptRecordedIndexPresent) {
		return fmt.Errorf("%sposition or recorded-index state/value is invalid; the row cannot be addressed safely; emit a nonnegative position and pair a present state with its recorded index", where)
	}
	if r.RecordedIndex != nil && (*r.RecordedIndex < 0 || int64(*r.RecordedIndex) > maxSafeJSONInteger) {
		return fmt.Errorf("%srecordedIndex is outside the safe nonnegative range; it cannot be used as an address; emit a valid index or its invalid state", where)
	}
	if r.DisplayNumber != "" && !transcriptDisplayNumberPattern.MatchString(r.DisplayNumber) {
		return fmt.Errorf("%sdisplayNumber is malformed; it cannot label a served row; emit a positive task number with an optional lowercase suffix", where)
	}
	if len(r.Turn.raw) == 0 {
		return fmt.Errorf("%sturn is empty; the row cannot be served; construct a turn from the served JSON object", where)
	}
	return nil
}

// TranscriptSegmentDirectoryV2 is one partition's bounded directory page.
type TranscriptSegmentDirectoryV2 struct {
	Context          TranscriptReadContext      `json:"context"`
	Partition        TranscriptSegmentPartition `json:"partition"`
	History          []TranscriptHistorySection `json:"history,omitempty"`
	UsagePresent     TranscriptUsagePresence    `json:"usagePresent"`
	Segments         []TranscriptSegmentRefV2   `json:"segments"`
	DirectoryPage    *int                       `json:"directoryPage,omitempty"`
	DirectoryPerPage *int                       `json:"directoryPerPage,omitempty"`
}

// Validate checks page bounds, contiguous served positions, and conversation spans.
func (d TranscriptSegmentDirectoryV2) Validate() error {
	const where = "v2 transcript directory validation failed at schema.TranscriptSegmentDirectoryV2.Validate: "
	if err := d.Context.Validate(); err != nil {
		return err
	}
	if err := d.Partition.Validate(); err != nil {
		return err
	}
	if !d.UsagePresent.IsValid() {
		return fmt.Errorf("%susagePresent is outside its closed set; availability is unknown; emit present, empty, or unknown", where)
	}
	if d.Partition.Kind != TranscriptPartitionMain && len(d.History) != 0 {
		return fmt.Errorf("%sa history partition carries a history manifest; the manifest belongs to main; omit it on section reads", where)
	}
	if d.Segments == nil || len(d.Segments) > MaxTranscriptDirectoryRefs {
		return fmt.Errorf("%ssegments is null or exceeds the directory cap; a page must be a bounded array; emit an array of at most %d refs", where, MaxTranscriptDirectoryRefs)
	}
	total := d.Context.TotalSegments
	paged := d.DirectoryPage != nil || d.DirectoryPerPage != nil
	start := 0
	if paged {
		if d.DirectoryPage == nil || d.DirectoryPerPage == nil {
			return fmt.Errorf("%sdirectoryPage and directoryPerPage must be set together; a partial page cannot be placed; emit both or neither", where)
		}
		page, size := *d.DirectoryPage, *d.DirectoryPerPage
		if page < 1 || size != MaxTranscriptDirectoryRefs || total <= MaxTranscriptDirectoryRefs ||
			len(d.Segments) > size || total < len(d.Segments) {
			return fmt.Errorf("%sthe page number, size, or total is invalid; the directory page cannot be placed; use a positive page and the fixed cap only above that cap", where)
		}
		if int64(page)-1 > maxSafeJSONInteger/int64(size) {
			return fmt.Errorf("%sthe page offset exceeds the safe integer range; the page cannot be addressed; request a reachable one-based page", where)
		}
		start = (page - 1) * size
	} else if total != len(d.Segments) {
		return fmt.Errorf("%stotalSegments differs from the ref count; an unpaged directory must hold every ref; emit the whole directory or page it", where)
	}
	if d.Context.TotalTurns < len(d.Segments) {
		return fmt.Errorf("%stotalTurns is smaller than the served ref count; each segment contains a row; emit the partition total", where)
	}
	spans := make([]*ConversationTurnSpan, len(d.Segments))
	for i, ref := range d.Segments {
		if err := ref.Validate(); err != nil {
			return err
		}
		if ref.Index != start+i || ref.Index >= total {
			return fmt.Errorf("%sref %d has ordinal %d; refs must be consecutive from page ordinal %d and below totalSegments %d; rebuild the directory", where, i, ref.Index, start, total)
		}
		if i == 0 && !paged && ref.SourceFirstPosition != 0 {
			return fmt.Errorf("%sthe first unpaged ref starts after position zero; the served partition would have a gap; rebuild the directory", where)
		}
		if i > 0 && ref.SourceFirstPosition-d.Segments[i-1].SourceLastPosition != 1 {
			return fmt.Errorf("%sref %d does not begin immediately after the preceding served range; rows would be skipped or repeated; rebuild the directory", where, ref.Index)
		}
		spans[i] = ref.ConversationTurns
	}
	return validateDirectoryConversationTurns(d.Context.TotalConversationTurns, spans, where)
}

// TranscriptSegmentV2 is the segment transport unit. Encode it only with
// MarshalTranscriptSegmentV2 to preserve each turn's served bytes.
type TranscriptSegmentV2 struct {
	Context        TranscriptReadContext      `json:"context"`
	Partition      TranscriptSegmentPartition `json:"partition"`
	Index          int                        `json:"index"`
	Rows           []TranscriptSegmentRow     `json:"rows"`
	ByteLength     int                        `json:"byteLength"`
	HasOlder       bool                       `json:"hasOlder"`
	HasNewer       bool                       `json:"hasNewer"`
	NativeMetadata []NativeMetadataRecord     `json:"nativeMetadata,omitempty"`
}

// Validate checks the segment's own facts and the contiguous served rows.
func (s TranscriptSegmentV2) Validate() error {
	const where = "v2 transcript segment validation failed at schema.TranscriptSegmentV2.Validate: "
	if err := s.Context.Validate(); err != nil {
		return err
	}
	if err := s.Partition.Validate(); err != nil {
		return err
	}
	if s.Index < 0 || s.Index >= s.Context.TotalSegments || s.ByteLength < 0 ||
		s.HasOlder != (s.Index > 0) || s.HasNewer != (s.Index < s.Context.TotalSegments-1) {
		return fmt.Errorf("%sthe segment index, byte length, or edge flags disagree with its context; the segment cannot be placed; emit the correct ordinal, count, and flags", where)
	}
	if len(s.Rows) == 0 || s.Rows[0].Position < 0 {
		return fmt.Errorf("%srows are empty or start at a negative position; a segment needs contiguous served rows; emit its nonempty row range", where)
	}
	for i, row := range s.Rows {
		if row.Position-s.Rows[0].Position != i {
			return fmt.Errorf("%srow %d is at position %d rather than %d; rows must be contiguous in served order; emit the segment's consecutive row range", where, i, row.Position, s.Rows[0].Position+i)
		}
		if err := row.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// TranscriptWindowV2 describes a contiguous segment run without turn content.
type TranscriptWindowV2 struct {
	Context           TranscriptReadContext      `json:"context"`
	Partition         TranscriptSegmentPartition `json:"partition"`
	RequestedAnchor   *int                       `json:"requestedAnchor,omitempty" minimum:"0" nullable:"false"`
	AnchorTurn        *int                       `json:"anchorTurn,omitempty" minimum:"0" nullable:"false"`
	RequestedPosition *int                       `json:"requestedPosition,omitempty" minimum:"0" nullable:"false"`
	AnchorPosition    int                        `json:"anchorPosition"`
	AnchorSegment     int                        `json:"anchorSegment"`
	Span              int                        `json:"span" minimum:"1" maximum:"10"`
	FirstPosition     int                        `json:"firstPosition"`
	LastPosition      int                        `json:"lastPosition"`
	Segments          []int                      `json:"segments"`
	HasOlder          bool                       `json:"hasOlder"`
	HasNewer          bool                       `json:"hasNewer"`
}

// TranscriptFlaggedTurn marks a recorded turn at a served main-partition position.
type TranscriptFlaggedTurn struct {
	Position int `json:"position" minimum:"0"`
	Turn     int `json:"turn"`
}
