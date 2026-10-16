package schema

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	jsonschema "github.com/swaggest/jsonschema-go"
)

const (
	MaxCollectiveObservationURLBytes   = 1024
	MaxContributionTraversalTokenBytes = 4096
	MaxContributionPageLimit           = 200
	MaxCollectiveFetchItems            = 10000
	MaxCollectiveOperationIssueBytes   = 2048
)

// CollectiveObservationScope identifies private evidence, never credentials or
// authority. The caller selects its authenticated scope before validating tokens.
type CollectiveObservationScope struct {
	VillageURL string      `json:"villageUrl" maxLength:"1024"`
	UserID     VillageUUID `json:"userId"`
}

func NormalizeCollectiveObservationScope(raw string, userID VillageUUID) (CollectiveObservationScope, error) {
	if len(raw) > MaxCollectiveObservationURLBytes || !transcriptIDUUID.MatchString(string(userID)) || strings.Contains(raw, "\\") || strings.TrimSpace(raw) != raw {
		return CollectiveObservationScope{}, observationError("scope URL exceeds its bound or userId is not a canonical UUID")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Opaque != "" || u.Host == "" || (strings.ToLower(u.Scheme) != "http" && strings.ToLower(u.Scheme) != "https") {
		return CollectiveObservationScope{}, observationError("scope requires an absolute HTTP(S) Village base URL")
	}
	u.Scheme = strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	if ip := net.ParseIP(host); ip != nil {
		host = ip.String()
	}
	if strings.Contains(host, "%") {
		return CollectiveObservationScope{}, observationError("scope host cannot contain a zone or ambiguous escape")
	}
	for _, r := range host {
		if r > 127 {
			return CollectiveObservationScope{}, observationError("scope host must use its ASCII DNS form")
		}
	}
	port := u.Port()
	if port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return CollectiveObservationScope{}, observationError("scope port is outside 1..65535")
		}
		port = strconv.Itoa(n)
	}
	if (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		port = ""
	}
	if port != "" {
		u.Host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		u.Host = "[" + host + "]"
	} else {
		u.Host = host
	}
	if host == "" || strings.Contains(u.EscapedPath(), "%") {
		return CollectiveObservationScope{}, observationError("scope host is empty or base path contains ambiguous escaped components")
	}
	u.User = nil
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	u.RawFragment = ""
	u.RawPath = ""
	u.Path = strings.TrimSuffix(path.Clean("/"+strings.TrimPrefix(u.Path, "/")), "/")
	if u.Path == "." {
		u.Path = ""
	}
	return CollectiveObservationScope{VillageURL: u.String(), UserID: userID}, nil
}

func (s CollectiveObservationScope) Validate() error {
	normalized, err := NormalizeCollectiveObservationScope(s.VillageURL, s.UserID)
	if err != nil {
		return err
	}
	if normalized != s {
		return observationError("scope URL is not normalized; normalize the saved Village base URL first")
	}
	return nil
}

type CollectiveObservationReason string

const (
	CollectiveObservationReasonMembership   CollectiveObservationReason = "membership"
	CollectiveObservationReasonContribution CollectiveObservationReason = "contribution"
	CollectiveObservationReasonOrganization CollectiveObservationReason = "organization"
	CollectiveObservationReasonExplicit     CollectiveObservationReason = "explicit"
)

var AllCollectiveObservationReasons = []CollectiveObservationReason{CollectiveObservationReasonMembership, CollectiveObservationReasonContribution, CollectiveObservationReasonOrganization, CollectiveObservationReasonExplicit}

func (v CollectiveObservationReason) IsValid() bool {
	return observationMember(v, AllCollectiveObservationReasons)
}
func (CollectiveObservationReason) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Observation Reason", "Independent dated relevance reason, not authority", AllCollectiveObservationReasons), nil
}

type CollectiveObservationState string

const (
	CollectiveObservationStateObserved    CollectiveObservationState = "observed"
	CollectiveObservationStateUnavailable CollectiveObservationState = "unavailable"
	CollectiveObservationStateInvalidated CollectiveObservationState = "invalidated"
)

var AllCollectiveObservationStates = []CollectiveObservationState{CollectiveObservationStateObserved, CollectiveObservationStateUnavailable, CollectiveObservationStateInvalidated}

func (v CollectiveObservationState) IsValid() bool {
	return observationMember(v, AllCollectiveObservationStates)
}
func (CollectiveObservationState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Observation State", "Dated disclosure state, never current authorization", AllCollectiveObservationStates), nil
}

type CollectiveObservation struct {
	ID             VillageUUID                   `json:"id"`
	Name           string                        `json:"name"`
	Description    *string                       `json:"description,omitempty"`
	AcceptanceMode *VillageGroupAcceptanceMode   `json:"acceptanceMode,omitempty"`
	DataAccess     *VillageGroupDataAccess       `json:"dataAccess,omitempty"`
	Reasons        []CollectiveObservationReason `json:"reasons" nullable:"false" maxItems:"4"`
	ObservedAt     time.Time                     `json:"observedAt"`
	State          CollectiveObservationState    `json:"state"`
}

func (o CollectiveObservation) Validate() error {
	if !transcriptIDUUID.MatchString(string(o.ID)) || !o.State.IsValid() || !attemptUTCTime(o.ObservedAt) || o.Reasons == nil || len(o.Reasons) > 4 {
		return observationError("observation requires canonical identity, UTC time, closed state and nonnull bounded reasons")
	}
	seen := map[CollectiveObservationReason]bool{}
	for _, r := range o.Reasons {
		if !r.IsValid() || seen[r] {
			return observationError("relevance reasons must be known and unique")
		}
		seen[r] = true
	}
	if o.AcceptanceMode != nil && !observationMember(*o.AcceptanceMode, AllVillageCollectiveAcceptanceModes) {
		return observationError("unknown collective acceptance mode")
	}
	if o.DataAccess != nil && !observationMember(*o.DataAccess, AllVillageCollectiveDataAccessPolicies) {
		return observationError("unknown collective data-access policy")
	}
	return nil
}

type CollectiveOperation string

const (
	CollectiveOperationPublish CollectiveOperation = "publish"
	CollectiveOperationShare   CollectiveOperation = "share"
	CollectiveOperationFetch   CollectiveOperation = "fetch"
	CollectiveOperationPull    CollectiveOperation = "pull"
)

var AllCollectiveOperations = []CollectiveOperation{CollectiveOperationPublish, CollectiveOperationShare, CollectiveOperationFetch, CollectiveOperationPull}

func (v CollectiveOperation) IsValid() bool { return observationMember(v, AllCollectiveOperations) }
func (CollectiveOperation) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Operation", "Operation whose actual outcome is recorded", AllCollectiveOperations), nil
}

type OperationIssueCode string

const (
	OperationIssueAuth        OperationIssueCode = "auth"
	OperationIssueRemote      OperationIssueCode = "remote"
	OperationIssueContract    OperationIssueCode = "contract"
	OperationIssueDeadline    OperationIssueCode = "deadline"
	OperationIssuePersistence OperationIssueCode = "persistence"
	OperationIssueSuperseded  OperationIssueCode = "superseded"
)

var AllOperationIssueCodes = []OperationIssueCode{OperationIssueAuth, OperationIssueRemote, OperationIssueContract, OperationIssueDeadline, OperationIssuePersistence, OperationIssueSuperseded}

func (v OperationIssueCode) IsValid() bool { return observationMember(v, AllOperationIssueCodes) }
func (OperationIssueCode) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Operation Issue Code", "Bounded failure category; retry is corrective text only", AllOperationIssueCodes), nil
}

type OperationIssue struct {
	Code      OperationIssueCode  `json:"code"`
	Operation CollectiveOperation `json:"operation"`
	Message   string              `json:"message" minLength:"1" maxLength:"2048"`
	Retry     string              `json:"retry" minLength:"1" maxLength:"2048"`
}

func (i OperationIssue) Validate() error {
	if !i.Code.IsValid() || !i.Operation.IsValid() || strings.TrimSpace(i.Message) == "" || strings.TrimSpace(i.Retry) == "" || len(i.Message) > MaxCollectiveOperationIssueBytes || len(i.Retry) > MaxCollectiveOperationIssueBytes {
		return observationError("issue requires a known code/operation and bounded actionable message/retry text")
	}
	return nil
}

type LocalContributionEvidenceKind string

const (
	LocalContributionEvidenceRemoteAttempts       LocalContributionEvidenceKind = "remote-attempts"
	LocalContributionEvidenceReceipts             LocalContributionEvidenceKind = "receipts"
	LocalContributionEvidenceOperationDiagnostics LocalContributionEvidenceKind = "operation-diagnostics"
)

var AllLocalContributionEvidenceKinds = []LocalContributionEvidenceKind{LocalContributionEvidenceRemoteAttempts, LocalContributionEvidenceReceipts, LocalContributionEvidenceOperationDiagnostics}

func (v LocalContributionEvidenceKind) IsValid() bool {
	return observationMember(v, AllLocalContributionEvidenceKinds)
}
func (LocalContributionEvidenceKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Local Contribution Evidence Kind", "One independently bounded evidence category", AllLocalContributionEvidenceKinds), nil
}

// ContributionTraversalIdentity is codec input, not an authorization token or a
// server revision. Generation belongs ONLY to the selected category payload and
// coverage, never a cache revision, poll timestamp, or joined catalog label.
type ContributionTraversalIdentity struct {
	Scope       CollectiveObservationScope    `json:"scope"`
	Incarnation VillageUUID                   `json:"incarnation"`
	Kind        LocalContributionEvidenceKind `json:"kind"`
	Generation  string                        `json:"generation"`
}
type ContributionTraversalRevision string

func (ContributionTraversalRevision) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithMaxLength(MaxContributionTraversalTokenBytes)
	s.WithDescription("Opaque versioned selected-category traversal token; not cacheRevision or authority")
	return s, nil
}

func EncodeContributionTraversalRevision(identity ContributionTraversalIdentity) (ContributionTraversalRevision, error) {
	if err := identity.Scope.Validate(); err != nil {
		return "", err
	}
	if !transcriptIDUUID.MatchString(string(identity.Incarnation)) || !identity.Kind.IsValid() || !observationDecimal(identity.Generation) {
		return "", observationError("traversal identity requires canonical incarnation, evidence kind and uint64 decimal generation")
	}
	fields := []string{"1", identity.Scope.VillageURL, string(identity.Scope.UserID), string(identity.Incarnation), string(identity.Kind), identity.Generation}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(fields); err != nil {
		return "", err
	}
	data := bytes.TrimSuffix(buffer.Bytes(), []byte{'\n'})
	token := base64.RawURLEncoding.EncodeToString(data)
	if len(token) > MaxContributionTraversalTokenBytes {
		return "", observationError("traversal token exceeds its encoded byte bound")
	}
	return ContributionTraversalRevision(token), nil
}
func ParseContributionTraversalRevision(token ContributionTraversalRevision) (ContributionTraversalIdentity, error) {
	var identity ContributionTraversalIdentity
	if len(token) == 0 || len(token) > MaxContributionTraversalTokenBytes {
		return identity, observationError("traversal token is missing or oversized")
	}
	data, err := base64.RawURLEncoding.Strict().DecodeString(string(token))
	if err != nil {
		return identity, observationError("traversal token is not canonical base64url")
	}
	var fields []string
	if json.Unmarshal(data, &fields) != nil || len(fields) != 6 || fields[0] != "1" {
		return identity, observationError("traversal token has an unsupported version or field framing")
	}
	identity = ContributionTraversalIdentity{Scope: CollectiveObservationScope{VillageURL: fields[1], UserID: VillageUUID(fields[2])}, Incarnation: VillageUUID(fields[3]), Kind: LocalContributionEvidenceKind(fields[4]), Generation: fields[5]}
	canonical, err := EncodeContributionTraversalRevision(identity)
	if err != nil {
		return ContributionTraversalIdentity{}, err
	}
	if canonical != token {
		return ContributionTraversalIdentity{}, observationError("traversal token is not canonically encoded")
	}
	return identity, nil
}

type CollectiveEvidenceCoverage string

const (
	CollectiveEvidenceCoverageAvailable CollectiveEvidenceCoverage = "available"
	CollectiveEvidenceCoverageUnknown   CollectiveEvidenceCoverage = "unknown"
)

var AllCollectiveEvidenceCoverages = []CollectiveEvidenceCoverage{CollectiveEvidenceCoverageAvailable, CollectiveEvidenceCoverageUnknown}

func (v CollectiveEvidenceCoverage) IsValid() bool {
	return observationMember(v, AllCollectiveEvidenceCoverages)
}
func (CollectiveEvidenceCoverage) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Evidence Coverage", "Available saved evidence or explicitly unknown source ranges", AllCollectiveEvidenceCoverages), nil
}

type LocalConfirmedContributionOutcome string

const (
	LocalConfirmedContributionPublished LocalConfirmedContributionOutcome = "published"
	LocalConfirmedContributionSubmitted LocalConfirmedContributionOutcome = "submitted"
)

var AllLocalConfirmedContributionOutcomes = []LocalConfirmedContributionOutcome{LocalConfirmedContributionPublished, LocalConfirmedContributionSubmitted}

func (v LocalConfirmedContributionOutcome) IsValid() bool {
	return observationMember(v, AllLocalConfirmedContributionOutcomes)
}
func (LocalConfirmedContributionOutcome) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Confirmed Contribution Outcome", "Actual request effect, never collective approval", AllLocalConfirmedContributionOutcomes), nil
}

type LocalConfirmedContributionReceipt struct {
	ReceiptID       string                            `json:"receiptId" minLength:"1" maxLength:"256"`
	Operation       CollectiveOperation               `json:"operation"`
	TranscriptID    TranscriptID                      `json:"transcriptId"`
	CollectiveID    *VillageUUID                      `json:"collectiveId,omitempty"`
	ConfirmedAt     time.Time                         `json:"confirmedAt"`
	Outcome         LocalConfirmedContributionOutcome `json:"outcome"`
	SourceReceiptID *string                           `json:"sourceReceiptId,omitempty" maxLength:"256"`
}

func (r LocalConfirmedContributionReceipt) Validate() error {
	if r.ReceiptID == "" || len(r.ReceiptID) > 256 || !transcriptIDUUID.MatchString(string(r.TranscriptID)) || !attemptUTCTime(r.ConfirmedAt) || !r.Outcome.IsValid() {
		return observationError("receipt requires bounded stable identity, transcript UUID, confirmed UTC time and actual effect")
	}
	if r.Operation == CollectiveOperationPublish {
		if r.Outcome != LocalConfirmedContributionPublished {
			return observationError("publish receipt must describe a published effect")
		}
	} else if r.Operation == CollectiveOperationShare {
		if r.Outcome != LocalConfirmedContributionSubmitted {
			return observationError("share receipt must describe a submitted request effect")
		}
	} else {
		return observationError("only publish/share can produce confirmed contribution receipts")
	}
	if r.CollectiveID != nil && !transcriptIDUUID.MatchString(string(*r.CollectiveID)) {
		return observationError("target-specific receipt collectiveId must be a canonical UUID")
	}
	if r.SourceReceiptID != nil && (*r.SourceReceiptID == "" || len(*r.SourceReceiptID) > 256) {
		return observationError("sourceReceiptId must be nonempty and bounded when present")
	}
	return nil
}

type LocalOperationDiagnosticOutcome string

const (
	LocalOperationDiagnosticFailed    LocalOperationDiagnosticOutcome = "failed"
	LocalOperationDiagnosticUncertain LocalOperationDiagnosticOutcome = "uncertain"
)

var AllLocalOperationDiagnosticOutcomes = []LocalOperationDiagnosticOutcome{LocalOperationDiagnosticFailed, LocalOperationDiagnosticUncertain}

func (v LocalOperationDiagnosticOutcome) IsValid() bool {
	return observationMember(v, AllLocalOperationDiagnosticOutcomes)
}
func (LocalOperationDiagnosticOutcome) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Local Operation Diagnostic Outcome", "Failure or uncertain remote effect, never successful contribution", AllLocalOperationDiagnosticOutcomes), nil
}

type LocalOperationDiagnostic struct {
	OperationID  VillageUUID                     `json:"operationId"`
	Operation    CollectiveOperation             `json:"operation"`
	TranscriptID *TranscriptID                   `json:"transcriptId,omitempty"`
	CollectiveID *VillageUUID                    `json:"collectiveId,omitempty"`
	AttemptedAt  time.Time                       `json:"attemptedAt"`
	Outcome      LocalOperationDiagnosticOutcome `json:"outcome"`
	Issue        OperationIssue                  `json:"issue"`
}

func (d LocalOperationDiagnostic) Validate() error {
	if !transcriptIDUUID.MatchString(string(d.OperationID)) || !d.Operation.IsValid() || !d.Outcome.IsValid() || !attemptUTCTime(d.AttemptedAt) || d.Issue.Operation != d.Operation {
		return observationError("diagnostic requires local UUID, known operation/outcome, UTC attempt time and matching issue operation")
	}
	if d.TranscriptID != nil && !transcriptIDUUID.MatchString(string(*d.TranscriptID)) {
		return observationError("diagnostic transcriptId must be a canonical UUID")
	}
	if d.CollectiveID != nil && !transcriptIDUUID.MatchString(string(*d.CollectiveID)) {
		return observationError("diagnostic collectiveId must be a canonical UUID")
	}
	return d.Issue.Validate()
}

type PullAssociationMetadataState string

const (
	PullAssociationMetadataObserved    PullAssociationMetadataState = "observed"
	PullAssociationMetadataStale       PullAssociationMetadataState = "stale"
	PullAssociationMetadataUnknown     PullAssociationMetadataState = "unknown"
	PullAssociationMetadataInvalidated PullAssociationMetadataState = "invalidated"
)

var AllPullAssociationMetadataStates = []PullAssociationMetadataState{PullAssociationMetadataObserved, PullAssociationMetadataStale, PullAssociationMetadataUnknown, PullAssociationMetadataInvalidated}

func (v PullAssociationMetadataState) IsValid() bool {
	return observationMember(v, AllPullAssociationMetadataStates)
}
func (PullAssociationMetadataState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Pull Association Metadata State", "Optional metadata result, independent of required content success", AllPullAssociationMetadataStates), nil
}

type PullAssociationMetadataResult struct {
	State      PullAssociationMetadataState `json:"state"`
	ObservedAt *time.Time                   `json:"observedAt,omitempty"`
	Warning    *OperationIssue              `json:"warning,omitempty"`
}

func (r PullAssociationMetadataResult) Validate() error {
	if !r.State.IsValid() {
		return observationError("pull association metadata state is unknown")
	}
	needsTime := r.State != PullAssociationMetadataUnknown
	needsWarning := r.State == PullAssociationMetadataStale || r.State == PullAssociationMetadataUnknown
	if (r.ObservedAt != nil) != needsTime || (r.ObservedAt != nil && !attemptUTCTime(*r.ObservedAt)) || (r.Warning != nil) != needsWarning {
		return observationError("observed/invalidated require time without warning; stale requires prior time and warning; unknown requires warning without invented time")
	}
	if r.Warning != nil {
		if r.Warning.Operation != CollectiveOperationPull {
			return observationError("optional pull metadata warning must name pull")
		}
		return r.Warning.Validate()
	}
	return nil
}

func observationMember[T ~string](v T, values []T) bool {
	for _, known := range values {
		if v == known {
			return true
		}
	}
	return false
}
func observationDecimal(v string) bool {
	if v == "" || len(v) > 20 || (len(v) > 1 && v[0] == '0') {
		return false
	}
	_, err := strconv.ParseUint(v, 10, 64)
	return err == nil && !strings.HasPrefix(v, "+")
}
func observationError(reason string) error {
	return fmt.Errorf("collective observation contract validation failed at schema during metadata processing: %s; retain prior evidence and correct the request or fetch validated metadata before retrying", reason)
}
