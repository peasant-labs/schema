package schema

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
	"golang.org/x/crypto/sha3"
)

const publishFingerprintDomain = "peasant.publish-operation.v1"

// TranscriptContentHash is the SHA3-256 digest of the exact transcript file bytes.
type TranscriptContentHash string

func NewTranscriptContentHash(raw string) (TranscriptContentHash, error) {
	v := TranscriptContentHash(raw)
	if err := v.Validate(); err != nil {
		return "", err
	}
	return v, nil
}

func ComputeTranscriptContentHash(content []byte) TranscriptContentHash {
	return TranscriptContentHash(ComputeTranscriptHash(content))
}

func (v TranscriptContentHash) Validate() error {
	return validateDigest(string(v), "transcript content hash")
}
func (v TranscriptContentHash) String() string { return string(v) }
func (TranscriptContentHash) JSONSchema() (jsonschema.Schema, error) {
	return digestSchema("Transcript Content Hash", "SHA3-256 digest of the exact transcript file bytes"), nil
}

// PublishRequestFingerprint identifies one canonical accepted publish operation.
type PublishRequestFingerprint string

func NewPublishRequestFingerprint(raw string) (PublishRequestFingerprint, error) {
	v := PublishRequestFingerprint(raw)
	if err := v.Validate(); err != nil {
		return "", err
	}
	return v, nil
}

func (v PublishRequestFingerprint) Validate() error {
	return validateDigest(string(v), "publish operation fingerprint")
}
func (v PublishRequestFingerprint) String() string { return string(v) }
func (PublishRequestFingerprint) JSONSchema() (jsonschema.Schema, error) {
	return digestSchema("Publish Request Fingerprint", "SHA3-256 digest of the canonical domain-separated publish operation"), nil
}

func digestSchema(title, description string) jsonschema.Schema {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle(title)
	s.WithDescription(description)
	s.WithPattern(`^[0-9a-f]{64}$`)
	s.WithMinLength(64)
	s.WithMaxLength(64)
	return s
}

func validateDigest(raw, name string) error {
	if len(raw) != 64 {
		return publicationError(name, fmt.Sprintf("value has %d characters instead of 64", len(raw)), "compute and send the complete lowercase SHA3-256 hexadecimal digest")
	}
	if _, err := hex.DecodeString(raw); err != nil || raw != strings.ToLower(raw) {
		return publicationError(name, "value is not lowercase hexadecimal", "compute and send the complete lowercase SHA3-256 hexadecimal digest")
	}
	return nil
}

// VisibilityIntent is the private-before-replacement visibility intent.
// Public intent never widens access during publish; widening is a separate owner update.
type VisibilityIntent string

const (
	VisibilityIntentPrivate VisibilityIntent = "private"
	VisibilityIntentPublic  VisibilityIntent = "public"
)

var AllVisibilityIntents = []VisibilityIntent{VisibilityIntentPrivate, VisibilityIntentPublic}

func (v VisibilityIntent) IsValid() bool {
	return v == "" || v == VisibilityIntentPrivate || v == VisibilityIntentPublic
}
func (v VisibilityIntent) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Visibility Intent", "Optional desired final access for legacy compatibility; content replacement remains private and widening occurs separately", AllVisibilityIntents), nil
}

type PublishOperationKind string

const (
	PublishOperationPreserve PublishOperationKind = "preserve"
	PublishOperationReplace  PublishOperationKind = "replace"
	PublishOperationAppend   PublishOperationKind = "append"
)

var AllPublishOperationKinds = []PublishOperationKind{PublishOperationPreserve, PublishOperationReplace, PublishOperationAppend}

func (v PublishOperationKind) IsValid() bool {
	return v == PublishOperationPreserve || v == PublishOperationReplace || v == PublishOperationAppend
}
func (v PublishOperationKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Publish Operation Kind", "Mutation semantic for one publish operation arm", AllPublishOperationKinds), nil
}

// Successor publication projections have distinct component identities so
// recursive unknown-field rejection does not tighten legacy uses of shared
// metadata components.
type AuthoritativeSessionIdentity struct {
	SessionID       SessionID  `json:"sessionId"`
	ParentSessionID *SessionID `json:"parentSessionId,omitempty" nullable:"false"`
	SchemaVersion   int        `json:"schemaVersion"`
}
type AuthoritativeModelInfo ModelInfo
type AuthoritativeTimestampInfo TimestampInfo
type AuthoritativeSourceInfo SourceInfo
type AuthoritativeCommitInfo CommitInfo
type AuthoritativeGitContext struct {
	Branch       *string                   `json:"branch,omitempty"`
	Remote       *string                   `json:"remote,omitempty"`
	Worktree     *string                   `json:"worktree,omitempty"`
	Tracking     *string                   `json:"tracking,omitempty"`
	Commits      []AuthoritativeCommitInfo `json:"commits,omitempty"`
	Associations []PublishedAssociation    `json:"associations,omitempty"`
}
type AuthoritativeProjectContext ProjectContext
type AuthoritativeSessionStats SessionStats
type AuthoritativeQualityMetrics QualityMetrics
type AuthoritativeSessionEntry SessionEntry
type AuthoritativeSubagentRef SubagentRef
type AuthoritativeDiagnosticEntry DiagnosticEntry
type AuthoritativeDiagnosticsInfo struct {
	Warnings []AuthoritativeDiagnosticEntry `json:"warnings"`
	Partial  *bool                          `json:"partial,omitempty"`
}

type CanonicalPublishGitContext struct {
	Branch   *string                   `json:"branch"`
	Remote   *string                   `json:"remote"`
	Worktree *string                   `json:"worktree"`
	Tracking *string                   `json:"tracking"`
	Commits  []AuthoritativeCommitInfo `json:"commits" nullable:"false"`
}

type CanonicalPublishReplacement struct {
	Identity    AuthoritativeSessionIdentity `json:"identity"`
	Model       AuthoritativeModelInfo       `json:"model"`
	Timestamp   AuthoritativeTimestampInfo   `json:"timestamp"`
	Source      AuthoritativeSourceInfo      `json:"source"`
	Git         CanonicalPublishGitContext   `json:"git"`
	Project     AuthoritativeProjectContext  `json:"project"`
	Stats       AuthoritativeSessionStats    `json:"stats"`
	Quality     *AuthoritativeQualityMetrics `json:"quality"`
	Entries     []AuthoritativeSessionEntry  `json:"entries" nullable:"false"`
	Subagents   []AuthoritativeSubagentRef   `json:"subagents" nullable:"false"`
	Diagnostics AuthoritativeDiagnosticsInfo `json:"diagnostics"`
}

type PublishLicenseOperation struct {
	Kind    PublishOperationKind `json:"kind"`
	License *License             `json:"license"`
}
type PublishAssociationOperation struct {
	Kind         PublishOperationKind   `json:"kind"`
	Associations []PublishedAssociation `json:"associations" nullable:"false"`
}
type CanonicalPublishOperation struct {
	Replacement  CanonicalPublishReplacement `json:"replacement"`
	License      PublishLicenseOperation     `json:"license"`
	Associations PublishAssociationOperation `json:"associations"`
	ContentHash  TranscriptContentHash       `json:"contentHash"`
}

// PublishNormalizedValues is the complete set of values Village normalizes.
type PublishNormalizedValues struct {
	RootHarness    Harness    `json:"rootHarness"`
	EntryHarnesses []Harness  `json:"entryHarnesses" nullable:"false"`
	DerivedTitle   *string    `json:"derivedTitle"`
	Visibility     Visibility `json:"visibility"`
	SchemaVersion  string     `json:"schemaVersion"`
}

type PublishAppliedState struct {
	License          *License                `json:"license"`
	Associations     []PublishedAssociation  `json:"associations" nullable:"false"`
	NormalizedValues PublishNormalizedValues `json:"normalizedValues"`
}

type AuthoritativePublishResponse struct {
	TranscriptID                TranscriptID              `json:"transcriptId"`
	TranscriptURL               string                    `json:"transcriptUrl"`
	Visibility                  Visibility                `json:"visibility"`
	ContentHash                 TranscriptContentHash     `json:"contentHash"`
	RequestOperationFingerprint PublishRequestFingerprint `json:"requestOperationFingerprint"`
	Applied                     PublishAppliedState       `json:"applied"`
	BlobKey                     string                    `json:"blobKey"`
	BlobSizeBytes               int64                     `json:"blobSizeBytes"`
	PublishedAt                 int64                     `json:"publishedAt"`
	UpdatedAt                   int64                     `json:"updatedAt"`
	Created                     bool                      `json:"created"`
}

func CanonicalizePublishRequest(request AuthoritativePublishRequest) (CanonicalPublishOperation, error) {
	if err := ValidatePublicationRequest(request); err != nil {
		return CanonicalPublishOperation{}, err
	}
	associations, err := canonicalAssociations(request.Git.Associations)
	if err != nil {
		return CanonicalPublishOperation{}, err
	}
	license := PublishLicenseOperation{Kind: PublishOperationPreserve}
	if request.License != "" {
		if !request.License.IsValid() {
			return CanonicalPublishOperation{}, publicationError("publish license", fmt.Sprintf("%q is not canonical", request.License), "send a canonical license or omit it")
		}
		value := request.License
		license = PublishLicenseOperation{Kind: PublishOperationReplace, License: &value}
	}
	return CanonicalPublishOperation{
		Replacement: CanonicalPublishReplacement{Identity: request.Identity, Model: request.Model, Timestamp: request.Timestamp, Source: request.Source, Git: CanonicalPublishGitContext{Branch: request.Git.Branch, Remote: request.Git.Remote, Worktree: request.Git.Worktree, Tracking: request.Git.Tracking, Commits: nonNil(request.Git.Commits)}, Project: request.Project, Stats: request.Stats, Quality: request.Quality, Entries: nonNil(request.Entries), Subagents: nonNil(request.Subagents), Diagnostics: request.Diagnostics},
		License:     license, Associations: PublishAssociationOperation{Kind: PublishOperationAppend, Associations: associations}, ContentHash: request.ContentHash,
	}, nil
}

func FingerprintPublishOperation(operation CanonicalPublishOperation) (PublishRequestFingerprint, error) {
	if err := operation.Validate(); err != nil {
		return "", err
	}
	var stream bytes.Buffer
	writeFrame(&stream, 0, []byte(publishFingerprintDomain))
	encodeCanonicalPublishOperation(&stream, operation)
	digest := sha3.Sum256(stream.Bytes())
	return PublishRequestFingerprint(hex.EncodeToString(digest[:])), nil
}

func writeFrame(dst *bytes.Buffer, id uint16, payload []byte) {
	var header [10]byte
	binary.BigEndian.PutUint16(header[:2], id)
	binary.BigEndian.PutUint64(header[2:], uint64(len(payload)))
	dst.Write(header[:])
	dst.Write(payload)
}

// encodeCanonicalPublishOperation is the protocol encoder for fingerprint
// currency. It is intentionally explicit: field identifiers and framing are
// static, so JSON naming, reflection order, and serializer upgrades cannot
// change a previously accepted operation's fingerprint.
func encodeCanonicalPublishOperation(dst *bytes.Buffer, operation CanonicalPublishOperation) {
	writeFrame(dst, 1, encodeReplacement(operation.Replacement))
	writeFrame(dst, 2, encodeLicenseOperation(operation.License))
	writeFrame(dst, 3, encodeAssociationOperation(operation.Associations))
	var content fingerprintEncoder
	content.text(1, string(operation.ContentHash))
	writeFrame(dst, 4, content.Bytes())
}

type fingerprintEncoder struct{ bytes.Buffer }

func (e *fingerprintEncoder) text(id uint16, value string) {
	writeFrame(&e.Buffer, id, append([]byte{1}, []byte(value)...))
}
func (e *fingerprintEncoder) integer(id uint16, value int64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(value))
	writeFrame(&e.Buffer, id, append([]byte{2}, b[:]...))
}
func (e *fingerprintEncoder) boolean(id uint16, value bool) {
	if value {
		writeFrame(&e.Buffer, id, []byte{3, 1})
	} else {
		writeFrame(&e.Buffer, id, []byte{3, 0})
	}
}
func (e *fingerprintEncoder) optionalText(id uint16, value *string) {
	if value == nil {
		writeFrame(&e.Buffer, id, []byte{4, 0})
		return
	}
	writeFrame(&e.Buffer, id, append([]byte{4, 1, 1}, []byte(*value)...))
}
func (e *fingerprintEncoder) optionalInt(id uint16, value *int) {
	if value == nil {
		writeFrame(&e.Buffer, id, []byte{4, 0})
		return
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(*value))
	writeFrame(&e.Buffer, id, append([]byte{4, 1, 2}, b[:]...))
}
func (e *fingerprintEncoder) optionalInt64(id uint16, value *int64) {
	if value == nil {
		writeFrame(&e.Buffer, id, []byte{4, 0})
		return
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(*value))
	writeFrame(&e.Buffer, id, append([]byte{4, 1, 2}, b[:]...))
}
func (e *fingerprintEncoder) optionalFloat(id uint16, value *float64) {
	if value == nil {
		writeFrame(&e.Buffer, id, []byte{4, 0})
		return
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], math.Float64bits(*value))
	writeFrame(&e.Buffer, id, append([]byte{4, 1, 5}, b[:]...))
}
func (e *fingerprintEncoder) optionalBool(id uint16, value *bool) {
	if value == nil {
		writeFrame(&e.Buffer, id, []byte{4, 0})
		return
	}
	encoded := byte(0)
	if *value {
		encoded = 1
	}
	writeFrame(&e.Buffer, id, []byte{4, 1, 3, encoded})
}

func encodeReplacement(v CanonicalPublishReplacement) []byte {
	var e fingerprintEncoder
	e.text(1, string(v.Identity.SessionID))
	e.integer(2, int64(v.Identity.SchemaVersion))
	if v.Identity.ParentSessionID == nil {
		writeFrame(&e.Buffer, 3, nil)
	} else {
		e.text(3, string(*v.Identity.ParentSessionID))
	}
	e.text(4, string(v.Model.Harness))
	e.text(5, string(v.Model.Model))
	e.text(6, v.Model.HarnessVersion)
	e.text(7, string(v.Model.HostSlug))
	e.integer(8, v.Timestamp.Start)
	e.integer(9, v.Timestamp.End)
	e.optionalInt64(10, v.Timestamp.Ingested)
	e.text(11, string(v.Source.Format))
	e.text(12, v.Source.FilePath)
	e.optionalText(13, v.Git.Branch)
	e.optionalText(14, v.Git.Remote)
	e.optionalText(15, v.Git.Worktree)
	e.optionalText(16, v.Git.Tracking)
	for i, c := range v.Git.Commits {
		var n fingerprintEncoder
		n.text(1, c.Hash)
		n.text(2, c.Message)
		n.text(3, c.AuthorName)
		n.text(4, c.AuthorEmail)
		n.integer(5, c.CommitTime)
		n.integer(6, c.AuthorTime)
		writeFrame(&e.Buffer, uint16(1000+i), n.Bytes())
	}
	e.text(17, string(v.Project.Hash))
	e.text(18, v.Project.FilePath)
	e.text(19, v.Project.Name)
	e.integer(20, int64(v.Stats.TurnCount))
	e.integer(21, int64(v.Stats.ToolCallCount))
	e.integer(22, int64(v.Stats.SubagentCount))
	e.integer(23, v.Stats.DurationMs)
	e.integer(24, int64(v.Stats.TokensIn))
	e.integer(25, int64(v.Stats.TokensOut))
	e.optionalInt(26, v.Stats.ThoughtTokens)
	e.optionalInt(27, v.Stats.CachedReadTokens)
	e.optionalInt(28, v.Stats.CachedWriteTokens)
	if v.Quality == nil {
		writeFrame(&e.Buffer, 29, nil)
	} else {
		writeFrame(&e.Buffer, 29, encodeQuality(QualityMetrics(*v.Quality)))
	}
	for i, entry := range v.Entries {
		writeFrame(&e.Buffer, uint16(2000+i), encodeEntry(SessionEntry(entry)))
	}
	for i, sub := range v.Subagents {
		var n fingerprintEncoder
		n.text(1, string(sub.SessionID))
		n.text(2, string(sub.ParentUUID))
		writeFrame(&e.Buffer, uint16(3000+i), n.Bytes())
	}
	for i, warning := range v.Diagnostics.Warnings {
		var n fingerprintEncoder
		n.text(1, warning.ErrorType)
		n.text(2, warning.Location)
		n.text(3, warning.Message)
		n.text(4, warning.Remediation)
		writeFrame(&e.Buffer, uint16(4000+i), n.Bytes())
	}
	e.optionalBool(30, v.Diagnostics.Partial)
	return e.Bytes()
}

func encodeEntry(v SessionEntry) []byte {
	var e fingerprintEncoder
	e.text(1, string(v.SessionID))
	e.integer(2, int64(v.EntryIndex))
	e.text(3, string(v.Harness))
	e.text(4, string(v.EntryType))
	e.text(5, string(v.Role))
	e.optionalInt64(6, v.TimestampMs)
	e.optionalText(7, v.ContentPreview)
	e.optionalInt(8, v.TokensIn)
	e.optionalInt(9, v.TokensOut)
	e.boolean(10, v.HasToolUse)
	if v.ToolKind == nil {
		writeFrame(&e.Buffer, 11, nil)
	} else {
		e.text(11, string(*v.ToolKind))
	}
	e.optionalText(12, v.ToolNamesCSV)
	e.boolean(13, v.HasThinking)
	e.boolean(14, v.IsError)
	if v.StopReason == nil {
		writeFrame(&e.Buffer, 15, nil)
	} else {
		e.text(15, string(*v.StopReason))
	}
	e.optionalInt(16, v.RawByteLength)
	e.optionalText(17, v.ToolCallID)
	e.optionalText(18, v.EntryID)
	e.optionalText(19, v.ParentEntryID)
	e.integer(20, int64(v.Depth))
	e.optionalInt(21, v.ParentIndex)
	e.optionalText(22, v.ToolInput)
	e.optionalText(23, v.ToolOutput)
	e.optionalText(24, v.Extra)
	e.optionalText(25, v.PartType)
	return e.Bytes()
}

func encodeQuality(v QualityMetrics) []byte {
	var e fingerprintEncoder
	ints := []*int{v.TurnCount, v.SubagentCount, v.TotalTokens, v.InputTokens, v.OutputTokens, v.ToolCalls, v.FilesTouched, v.LinesChanged, v.RetryLoops, v.RetryTokensWasted, v.WithinSessionReverts, v.ScopeBreadth, v.DiscoveryTurns, v.M3UniqueToolCount, v.M4ErrorRecoveryCount, v.M4ConsecutiveErrorMax, v.M5PeakContextTokens, v.M5AvgMessageTokens, v.M6LinesSurvived, v.M6LinesTotal, v.M7SpecWordCount, v.ComputeVersion}
	for i, p := range ints {
		e.optionalInt(uint16(i+1), p)
	}
	e.optionalText(30, v.TitleGenerated)
	if v.Outcome == nil {
		writeFrame(&e.Buffer, 31, nil)
	} else {
		e.text(31, string(*v.Outcome))
	}
	floats := []*float64{v.SignalDensity, v.SpecQualityScore, v.ExplorationRatio, v.DurationMinutes, v.M2TokenOutcomeRatio, v.M5ContextUtilizationPct, v.M6OutputSurvivalPct, v.CostInputUSD, v.CostOutputUSD, v.CostReasoningUSD, v.CostCacheReadUSD, v.CostCacheWriteUSD, v.CostTotalUSD}
	for i, p := range floats {
		e.optionalFloat(uint16(40+i), p)
	}
	e.optionalBool(60, v.M7SpecHasExamples)
	e.optionalBool(61, v.M7SpecHasConstraints)
	e.optionalText(62, v.CostModelID)
	e.optionalText(63, v.Scope)
	e.optionalInt64(64, v.ComputedAt)
	return e.Bytes()
}
func encodeLicenseOperation(v PublishLicenseOperation) []byte {
	var e fingerprintEncoder
	e.text(1, string(v.Kind))
	if v.License == nil {
		writeFrame(&e.Buffer, 2, nil)
	} else {
		e.text(2, string(*v.License))
	}
	return e.Bytes()
}
func encodeAssociationOperation(v PublishAssociationOperation) []byte {
	var e fingerprintEncoder
	e.text(1, string(v.Kind))
	for i, a := range v.Associations {
		var n fingerprintEncoder
		n.text(1, string(a.ID))
		n.text(2, a.ObservedCommitHash)
		writeFrame(&e.Buffer, uint16(100+i), n.Bytes())
	}
	return e.Bytes()
}
func isDecimalSchemaVersion(value string) bool {
	if value == "" {
		return false
	}
	n, err := strconv.ParseUint(value, 10, 64)
	return err == nil && n > 0 && strconv.FormatUint(n, 10) == value
}

func (operation CanonicalPublishOperation) Validate() error {
	if operation.License.Kind == PublishOperationPreserve {
		if operation.License.License != nil {
			return publicationError("publish license operation", "preserve carries a value", "remove the value")
		}
	} else if operation.License.Kind == PublishOperationReplace {
		if operation.License.License == nil || !operation.License.License.IsValid() {
			return publicationError("publish license operation", "replace lacks a canonical value", "provide one canonical license")
		}
	} else {
		return publicationError("publish license operation", "kind is neither preserve nor replace", "use preserve or replace")
	}
	if operation.Associations.Kind != PublishOperationAppend {
		return publicationError("publish association operation", "kind is not append", "use append")
	}
	if operation.Associations.Associations == nil {
		return publicationError("publish association operation", "associations is null", "send an empty array for no associations")
	}
	if err := validateCanonicalAssociations(operation.Associations.Associations); err != nil {
		return err
	}
	request := AuthoritativePublishRequest{
		Identity: operation.Replacement.Identity, Model: operation.Replacement.Model, Timestamp: operation.Replacement.Timestamp,
		Source: operation.Replacement.Source, Git: AuthoritativeGitContext{Branch: operation.Replacement.Git.Branch, Remote: operation.Replacement.Git.Remote, Worktree: operation.Replacement.Git.Worktree, Tracking: operation.Replacement.Git.Tracking, Commits: operation.Replacement.Git.Commits, Associations: operation.Associations.Associations},
		Project: operation.Replacement.Project, Stats: operation.Replacement.Stats, Quality: operation.Replacement.Quality,
		Entries: operation.Replacement.Entries, Subagents: operation.Replacement.Subagents, Diagnostics: operation.Replacement.Diagnostics,
		ContentHash: operation.ContentHash, VisibilityIntent: VisibilityIntentPrivate,
	}
	if operation.License.License != nil {
		request.License = *operation.License.License
	}
	return ValidatePublicationRequest(request)
}

func (state PublishAppliedState) Validate() error {
	if state.License != nil && !state.License.IsValid() {
		return publicationError("applied license", "license is not canonical", "return null or a canonical license")
	}
	if state.Associations == nil {
		return publicationError("applied associations", "associations is null", "return the complete canonical array")
	}
	if err := validateCanonicalAssociations(state.Associations); err != nil {
		return err
	}
	if state.NormalizedValues.EntryHarnesses == nil {
		return publicationError("normalized values", "entryHarnesses is null", "return an ordered array, including empty")
	}
	if !slices.Contains(Harnesses(), state.NormalizedValues.RootHarness) || !state.NormalizedValues.Visibility.IsValid() || !isDecimalSchemaVersion(state.NormalizedValues.SchemaVersion) {
		return publicationError("normalized values", "a required normalized value is invalid", "return the persisted root harness, visibility, and positive decimal schema version")
	}
	for _, harness := range state.NormalizedValues.EntryHarnesses {
		if !slices.Contains(Harnesses(), harness) {
			return publicationError("normalized values", fmt.Sprintf("entry harness %q is invalid", harness), "return canonical harness values in entry order")
		}
	}
	return nil
}

func (response AuthoritativePublishResponse) Validate() error {
	if _, err := NewTranscriptID(response.TranscriptID.String()); err != nil {
		return err
	}
	if err := validatePublicationURL(response.TranscriptURL, response.TranscriptID); err != nil {
		return err
	}
	if !response.Visibility.IsValid() {
		return publicationError("publish visibility", "actual visibility is invalid", "return authoritative visibility")
	}
	if err := response.ContentHash.Validate(); err != nil {
		return err
	}
	if err := response.RequestOperationFingerprint.Validate(); err != nil {
		return err
	}
	if err := response.Applied.Validate(); err != nil {
		return err
	}
	if response.Applied.NormalizedValues.Visibility != response.Visibility {
		return publicationError("publish visibility", "top-level visibility differs from applied normalized visibility", "return the same authoritative persisted visibility in both receipt locations")
	}
	if strings.TrimSpace(response.BlobKey) == "" || response.BlobSizeBytes < 0 {
		return publicationError("publish blob facts", "blob key is empty or size is negative", "return a nonempty key and nonnegative exact size")
	}
	if response.PublishedAt <= 0 || response.UpdatedAt < response.PublishedAt {
		return publicationError("publish chronology", "timestamps are not positive and ordered", "return positive timestamps with updatedAt at or after publishedAt")
	}
	return nil
}

func canonicalAssociations(input []PublishedAssociation) ([]PublishedAssociation, error) {
	git := GitContext{Associations: input}
	if err := git.Validate(); err != nil {
		return nil, publicationError("published associations", err.Error(), "send unique nonempty IDs and bindings")
	}
	out := append([]PublishedAssociation(nil), input...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		return out[i].ObservedCommitHash < out[j].ObservedCommitHash
	})
	return nonNil(out), nil
}
func validateCanonicalAssociations(input []PublishedAssociation) error {
	canonical, err := canonicalAssociations(input)
	if err != nil {
		return err
	}
	for i := range input {
		if input[i] != canonical[i] {
			return publicationError("published associations", "array is not in canonical ID and binding order", "return the complete sorted array")
		}
	}
	return nil
}
func nonNil[T any](input []T) []T {
	if input == nil {
		return []T{}
	}
	return input
}
func validatePublicationURL(raw string, id TranscriptID) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return publicationError("transcript URL", "URL is not absolute HTTPS", "return the canonical frontend HTTPS URL")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" || strings.TrimSuffix(parsed.Path, "/") != "/transcripts/"+id.String() {
		return publicationError("transcript URL", "URL is not exactly /transcripts/{transcriptId}", "build it from the authoritative transcript ID without query or fragment")
	}
	return nil
}
func publicationError(what, why, fix string) error {
	return fmt.Errorf("%s validation failed at schema publication boundary while decoding or validating authoritative publication state: %s; the caller must not treat the operation as successful; %s", what, why, fix)
}
