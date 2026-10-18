package schema

import (
	"bytes"
	"encoding/json"
	jsonschema "github.com/swaggest/jsonschema-go"
	"reflect"
	"time"
)

type CollectiveFetchDomain string

const (
	CollectiveFetchDomainCatalog      CollectiveFetchDomain = "catalog"
	CollectiveFetchDomainHistory      CollectiveFetchDomain = "history"
	CollectiveFetchDomainAssociations CollectiveFetchDomain = "associations"
)

var AllCollectiveFetchDomains = []CollectiveFetchDomain{CollectiveFetchDomainCatalog, CollectiveFetchDomainHistory, CollectiveFetchDomainAssociations}

func (v CollectiveFetchDomain) IsValid() bool { return observationMember(v, AllCollectiveFetchDomains) }
func (CollectiveFetchDomain) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Fetch Domain", "Independently installed complete metadata dataset", AllCollectiveFetchDomains), nil
}

type CollectiveFetchState string

const (
	CollectiveFetchStateComplete     CollectiveFetchState = "complete"
	CollectiveFetchStateFailed       CollectiveFetchState = "failed"
	CollectiveFetchStateCancelled    CollectiveFetchState = "cancelled"
	CollectiveFetchStateSuperseded   CollectiveFetchState = "superseded"
	CollectiveFetchStateNotRequested CollectiveFetchState = "not_requested"
)

var AllCollectiveFetchStates = []CollectiveFetchState{CollectiveFetchStateComplete, CollectiveFetchStateFailed, CollectiveFetchStateCancelled, CollectiveFetchStateSuperseded, CollectiveFetchStateNotRequested}

func (v CollectiveFetchState) IsValid() bool { return observationMember(v, AllCollectiveFetchStates) }
func (CollectiveFetchState) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Fetch State", "One invocation result; a failed domain retains prior completed dates", AllCollectiveFetchStates), nil
}

type CollectiveFetchCommit string

const (
	CollectiveFetchCommitUpdated   CollectiveFetchCommit = "updated"
	CollectiveFetchCommitUnchanged CollectiveFetchCommit = "unchanged"
	CollectiveFetchCommitRetained  CollectiveFetchCommit = "retained"
)

var AllCollectiveFetchCommits = []CollectiveFetchCommit{CollectiveFetchCommitUpdated, CollectiveFetchCommitUnchanged, CollectiveFetchCommitRetained}

func (v CollectiveFetchCommit) IsValid() bool { return observationMember(v, AllCollectiveFetchCommits) }
func (CollectiveFetchCommit) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Fetch Commit", "Committed effect, including earlier commits before cancellation", AllCollectiveFetchCommits), nil
}

type CollectiveFetchOutcome string

const (
	CollectiveFetchOutcomeComplete CollectiveFetchOutcome = "complete"
	CollectiveFetchOutcomePartial  CollectiveFetchOutcome = "partial"
	CollectiveFetchOutcomeFailed   CollectiveFetchOutcome = "failed"
)

var AllCollectiveFetchOutcomes = []CollectiveFetchOutcome{CollectiveFetchOutcomeComplete, CollectiveFetchOutcomePartial, CollectiveFetchOutcomeFailed}

func (v CollectiveFetchOutcome) IsValid() bool {
	return observationMember(v, AllCollectiveFetchOutcomes)
}
func (CollectiveFetchOutcome) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Fetch Outcome", "Complete only when every requested required scope completed", AllCollectiveFetchOutcomes), nil
}

type CollectiveObservationSource string

const (
	CollectiveObservationSourceMembership            CollectiveObservationSource = "membership"
	CollectiveObservationSourceContributionDiscovery CollectiveObservationSource = "contribution-discovery"
	CollectiveObservationSourceOrganization          CollectiveObservationSource = "organization"
	CollectiveObservationSourceExplicit              CollectiveObservationSource = "explicit"
	CollectiveObservationSourceSubmissions           CollectiveObservationSource = "submissions"
	CollectiveObservationSourceAttempts              CollectiveObservationSource = "attempts"
	CollectiveObservationSourceAssociations          CollectiveObservationSource = "associations"
)

var AllCollectiveObservationSources = []CollectiveObservationSource{CollectiveObservationSourceMembership, CollectiveObservationSourceContributionDiscovery, CollectiveObservationSourceOrganization, CollectiveObservationSourceExplicit, CollectiveObservationSourceSubmissions, CollectiveObservationSourceAttempts, CollectiveObservationSourceAssociations}

func (v CollectiveObservationSource) IsValid() bool {
	return observationMember(v, AllCollectiveObservationSources)
}
func (CollectiveObservationSource) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Observation Source", "Narrow metadata source, no world directory or transcript content", AllCollectiveObservationSources), nil
}

type CollectiveSourceOutcome string

const (
	CollectiveSourceDisclosed    CollectiveSourceOutcome = "disclosed"
	CollectiveSourceNotDisclosed CollectiveSourceOutcome = "not_disclosed"
	CollectiveSourceFailed       CollectiveSourceOutcome = "failed"
)

var AllCollectiveSourceOutcomes = []CollectiveSourceOutcome{CollectiveSourceDisclosed, CollectiveSourceNotDisclosed, CollectiveSourceFailed}

func (v CollectiveSourceOutcome) IsValid() bool {
	return observationMember(v, AllCollectiveSourceOutcomes)
}
func (CollectiveSourceOutcome) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Collective Source Outcome", "Privacy-safe source outcome without hidden reasons", AllCollectiveSourceOutcomes), nil
}

type CollectiveFetchSourceResult struct {
	Source     CollectiveObservationSource `json:"source"`
	Outcome    CollectiveSourceOutcome     `json:"outcome"`
	ObservedAt *time.Time                  `json:"observedAt,omitempty"`
	Complete   bool                        `json:"complete"`
	Issue      *OperationIssue             `json:"issue,omitempty"`
}
type CollectiveFetchAssociationScopeResult struct {
	TranscriptID TranscriptID          `json:"transcriptId"`
	State        CollectiveFetchState  `json:"state"`
	Commit       CollectiveFetchCommit `json:"commit"`
	ObservedAt   *time.Time            `json:"observedAt,omitempty"`
	CompletedAt  *time.Time            `json:"completedAt,omitempty"`
	Issue        *OperationIssue       `json:"issue,omitempty"`
}
type CollectiveFetchAssociationResults struct {
	Requested int                                     `json:"requested" minimum:"0" maximum:"10000"`
	Completed int                                     `json:"completed" minimum:"0" maximum:"10000"`
	Failed    int                                     `json:"failed" minimum:"0" maximum:"10000"`
	Scopes    []CollectiveFetchAssociationScopeResult `json:"scopes" maxItems:"10000" nullable:"false"`
}
type CollectiveFetchDomainResult struct {
	Domain                CollectiveFetchDomain              `json:"domain"`
	Requested             bool                               `json:"requested"`
	State                 CollectiveFetchState               `json:"state"`
	Commit                CollectiveFetchCommit              `json:"commit"`
	StartedAt             *time.Time                         `json:"startedAt,omitempty"`
	LastAttemptAt         *time.Time                         `json:"lastAttemptAt,omitempty"`
	CompletedAt           *time.Time                         `json:"completedAt,omitempty"`
	LastCompletedAt       *time.Time                         `json:"lastCompletedAt,omitempty"`
	LastInstalledRevision *string                            `json:"lastInstalledRevision,omitempty"`
	Coverage              CollectiveEvidenceCoverage         `json:"coverage"`
	SourceOutcomes        []CollectiveFetchSourceResult      `json:"sourceOutcomes" maxItems:"10000" nullable:"false"`
	Associations          *CollectiveFetchAssociationResults `json:"associations,omitempty"`
	Issue                 *OperationIssue                    `json:"issue,omitempty"`
}
type CollectiveFetchReport struct {
	Scope         CollectiveObservationScope    `json:"scope"`
	InvocationID  VillageUUID                   `json:"invocationId"`
	CacheRevision string                        `json:"cacheRevision" pattern:"^(0|[1-9][0-9]{0,19})$"`
	Outcome       CollectiveFetchOutcome        `json:"outcome"`
	Domains       []CollectiveFetchDomainResult `json:"domains" minItems:"3" maxItems:"3" nullable:"false"`
	Issues        []OperationIssue              `json:"issues" maxItems:"10000" nullable:"false"`
}
type CollectiveFetchDomainSummary struct {
	LastAttemptAt         *time.Time                 `json:"lastAttemptAt,omitempty"`
	LastCompletedAt       *time.Time                 `json:"lastCompletedAt,omitempty"`
	LastInstalledRevision *string                    `json:"lastInstalledRevision,omitempty"`
	State                 CollectiveFetchState       `json:"state"`
	Coverage              CollectiveEvidenceCoverage `json:"coverage"`
	Issue                 *OperationIssue            `json:"issue,omitempty"`
}
type CollectiveFetchSummary struct {
	LastCompletedFetchAt *time.Time                   `json:"lastCompletedFetchAt,omitempty"`
	LastAttemptAt        *time.Time                   `json:"lastAttemptAt,omitempty"`
	LatestReport         *CollectiveFetchReport       `json:"latestReport,omitempty"`
	Catalog              CollectiveFetchDomainSummary `json:"catalog"`
	History              CollectiveFetchDomainSummary `json:"history"`
	Associations         CollectiveFetchDomainSummary `json:"associations"`
}
type CollectiveObservationResponse struct {
	Scope         CollectiveObservationScope `json:"scope"`
	CacheRevision string                     `json:"cacheRevision" pattern:"^(0|[1-9][0-9]{0,19})$"`
	Collectives   []CollectiveObservation    `json:"collectives" maxItems:"10000" nullable:"false"`
	Fetch         CollectiveFetchSummary     `json:"fetch"`
}

func (s CollectiveFetchDomainSummary) Validate() error {
	if !s.State.IsValid() || !s.Coverage.IsValid() || !observationTimes(s.LastAttemptAt, s.LastCompletedAt) || (s.LastInstalledRevision != nil && !observationDecimal(*s.LastInstalledRevision)) {
		return observationError("domain summary requires known state/coverage and valid independent UTC dates/revision")
	}
	if s.Issue != nil {
		return s.Issue.Validate()
	}
	return nil
}
func (s CollectiveFetchSummary) Validate() error {
	if !observationTimes(s.LastCompletedFetchAt, s.LastAttemptAt) {
		return observationError("fetch summary dates must be actual UTC observations")
	}
	for _, d := range []CollectiveFetchDomainSummary{s.Catalog, s.History, s.Associations} {
		if err := d.Validate(); err != nil {
			return err
		}
	}
	if s.LatestReport != nil {
		return s.LatestReport.Validate()
	}
	return nil
}
func (r CollectiveObservationResponse) Validate() error {
	if err := r.Scope.Validate(); err != nil {
		return err
	}
	if !observationDecimal(r.CacheRevision) || r.Collectives == nil || len(r.Collectives) > MaxCollectiveFetchItems {
		return observationError("catalog requires decimal revision and bounded nonnull observations")
	}
	seen := map[VillageUUID]bool{}
	for _, o := range r.Collectives {
		if seen[o.ID] {
			return observationError("catalog repeats a collective identity")
		}
		seen[o.ID] = true
		if err := o.Validate(); err != nil {
			return err
		}
	}
	if r.Fetch.LatestReport != nil && r.Fetch.LatestReport.Scope != r.Scope {
		return observationError("catalog latest report belongs to a different scope")
	}
	return r.Fetch.Validate()
}

func (r CollectiveFetchReport) Validate() error {
	if err := r.Scope.Validate(); err != nil {
		return err
	}
	if !transcriptIDUUID.MatchString(string(r.InvocationID)) || !observationDecimal(r.CacheRevision) || !r.Outcome.IsValid() || len(r.Domains) != len(AllCollectiveFetchDomains) || r.Issues == nil || len(r.Issues) > MaxCollectiveFetchItems {
		return observationError("fetch report requires valid identity/revision, all independent domains and a bounded nonnull issue list")
	}
	seen := map[CollectiveFetchDomain]bool{}
	complete, progress, requested := 0, false, 0
	for _, d := range r.Domains {
		if seen[d.Domain] {
			return observationError("fetch report repeats a domain")
		}
		seen[d.Domain] = true
		if err := d.Validate(); err != nil {
			return err
		}
		if d.Requested {
			requested++
			if d.State == CollectiveFetchStateComplete {
				complete++
			}
			if d.State == CollectiveFetchStateComplete || d.Commit != CollectiveFetchCommitRetained {
				progress = true
			}
		}
	}
	want := CollectiveFetchOutcomeFailed
	if requested > 0 && complete == requested {
		want = CollectiveFetchOutcomeComplete
	} else if progress {
		want = CollectiveFetchOutcomePartial
	}
	if r.Outcome != want {
		return observationError("overall outcome disagrees with independently completed or committed domains")
	}
	if r.Outcome == CollectiveFetchOutcomeComplete && len(r.Issues) != 0 {
		return observationError("complete report cannot contain required failure issues")
	}
	for _, i := range r.Issues {
		if err := i.Validate(); err != nil {
			return err
		}
	}
	return nil
}
func (d CollectiveFetchDomainResult) Validate() error {
	if !d.Domain.IsValid() || !d.State.IsValid() || !d.Commit.IsValid() || !d.Coverage.IsValid() || d.SourceOutcomes == nil || len(d.SourceOutcomes) > MaxCollectiveFetchItems {
		return observationError("domain result has an unknown menu value or missing/bounded source outcomes")
	}
	if d.Requested == (d.State == CollectiveFetchStateNotRequested) {
		return observationError("requested must agree with not_requested state")
	}
	if !observationTimes(d.StartedAt, d.LastAttemptAt, d.CompletedAt, d.LastCompletedAt) || (d.LastInstalledRevision != nil && !observationDecimal(*d.LastInstalledRevision)) {
		return observationError("domain dates/revision must be actual UTC observations and decimal saved revisions")
	}
	if d.State == CollectiveFetchStateComplete {
		if d.CompletedAt == nil || d.LastCompletedAt == nil || d.LastInstalledRevision == nil || d.Commit == CollectiveFetchCommitRetained || d.Issue != nil {
			return observationError("complete domain requires installed revision/completion times, updated or unchanged effect and no issue")
		}
	} else {
		if d.CompletedAt != nil {
			return observationError("incomplete domain cannot claim this invocation completed")
		}
		if d.Domain != CollectiveFetchDomainAssociations && d.Commit != CollectiveFetchCommitRetained {
			return observationError("incomplete catalog/history must retain the prior complete dataset")
		}
		if d.State != CollectiveFetchStateNotRequested && d.Issue == nil {
			return observationError("incomplete requested domain requires an actionable issue")
		}
	}
	if d.Issue != nil {
		if err := d.Issue.Validate(); err != nil {
			return err
		}
	}
	for _, s := range d.SourceOutcomes {
		if d.State == CollectiveFetchStateComplete && !s.Complete {
			return observationError("complete domain cannot depend on an incomplete source outcome")
		}
		if !s.Source.IsValid() || !s.Outcome.IsValid() || !observationTimes(s.ObservedAt) || (s.Outcome == CollectiveSourceFailed) == s.Complete || (s.Complete && s.ObservedAt == nil) || (s.Outcome == CollectiveSourceFailed) != (s.Issue != nil) {
			return observationError("source outcome must pair completeness, observation time and failure issue truthfully")
		}
		if s.Issue != nil {
			if err := s.Issue.Validate(); err != nil {
				return err
			}
		}
	}
	if d.Domain != CollectiveFetchDomainAssociations && d.Associations != nil {
		return observationError("association scoped results belong only to the associations domain")
	}
	if d.Domain == CollectiveFetchDomainAssociations && d.Requested && d.Associations == nil {
		return observationError("requested associations require bounded per-transcript results and counts")
	}
	if !d.Requested && (d.Commit != CollectiveFetchCommitRetained || d.Issue != nil || len(d.SourceOutcomes) != 0 || d.Associations != nil) {
		return observationError("not-requested domain cannot claim new effects or issues")
	}
	if d.Associations != nil {
		a := d.Associations
		if a.Scopes == nil || a.Requested < 0 || a.Requested > MaxCollectiveFetchItems || len(a.Scopes) != a.Requested || a.Completed < 0 || a.Failed < 0 || a.Completed+a.Failed != a.Requested {
			return observationError("association scoped counts must describe exactly the bounded requested sets")
		}
		seen := map[TranscriptID]bool{}
		n := 0
		updated := false
		for _, s := range a.Scopes {
			if seen[s.TranscriptID] || !transcriptIDUUID.MatchString(string(s.TranscriptID)) || !s.State.IsValid() || s.State == CollectiveFetchStateNotRequested || !s.Commit.IsValid() || !observationTimes(s.ObservedAt, s.CompletedAt) {
				return observationError("association scoped result has duplicate/invalid identity, state or dates")
			}
			seen[s.TranscriptID] = true
			if s.State == CollectiveFetchStateComplete {
				if s.Commit == CollectiveFetchCommitUpdated {
					updated = true
				}
				n++
				if s.CompletedAt == nil || s.ObservedAt == nil || s.Commit == CollectiveFetchCommitRetained || s.Issue != nil {
					return observationError("completed association set requires observation/completion times and committed effect")
				}
			} else {
				if s.Issue == nil || s.Commit != CollectiveFetchCommitRetained || s.CompletedAt != nil {
					return observationError("failed association set retains old evidence and requires an issue")
				}
				if err := s.Issue.Validate(); err != nil {
					return err
				}
			}
		}
		if n != a.Completed || (d.State == CollectiveFetchStateComplete && a.Failed != 0) {
			return observationError("association completion counts disagree with scoped effects")
		}
		if updated && d.Commit != CollectiveFetchCommitUpdated {
			return observationError("an updated association scope requires an updated aggregate committed effect")
		}
		if d.State != CollectiveFetchStateComplete {
			want := CollectiveFetchCommitRetained
			if updated {
				want = CollectiveFetchCommitUpdated
			} else if n > 0 {
				want = CollectiveFetchCommitUnchanged
			}
			if d.Commit != want {
				return observationError("partial association domain must preserve its actual per-set committed effects")
			}
		}
	}
	return nil
}

// CollectiveFetchHTTPStatus chooses the transport status without dropping any
// domain effects. Validate the report first; malformed reports fail closed.
func CollectiveFetchHTTPStatus(report CollectiveFetchReport) int {
	if report.Validate() != nil {
		return 500
	}
	if report.Outcome == CollectiveFetchOutcomeComplete {
		return 200
	}
	rank := 0
	visit := func(i OperationIssue) {
		n := 0
		switch i.Code {
		case OperationIssuePersistence:
			n = 5
		case OperationIssueAuth:
			n = 4
		case OperationIssueDeadline:
			n = 3
		case OperationIssueRemote, OperationIssueContract:
			n = 2
		case OperationIssueSuperseded:
			n = 1
		}
		if n > rank {
			rank = n
		}
	}
	for _, i := range report.Issues {
		visit(i)
	}
	for _, d := range report.Domains {
		if d.State == CollectiveFetchStateCancelled && rank < 3 {
			rank = 3
		}
		if d.Issue != nil {
			visit(*d.Issue)
		}
		for _, s := range d.SourceOutcomes {
			if s.Issue != nil {
				visit(*s.Issue)
			}
		}
		if d.Associations != nil {
			for _, s := range d.Associations.Scopes {
				if s.State == CollectiveFetchStateCancelled && rank < 3 {
					rank = 3
				}
				if s.Issue != nil {
					visit(*s.Issue)
				}
			}
		}
	}
	switch rank {
	case 5:
		return 500
	case 4:
		return 401
	case 3:
		return 504
	case 2:
		return 502
	case 1:
		return 409
	default:
		return 500
	}
}
func observationTimes(values ...*time.Time) bool {
	for _, v := range values {
		if v != nil && !attemptUTCTime(*v) {
			return false
		}
	}
	return true
}

type LocalContributionHistoryResponse struct {
	Scope                CollectiveObservationScope          `json:"scope"`
	Kind                 LocalContributionEvidenceKind       `json:"kind"`
	TraversalRevision    ContributionTraversalRevision       `json:"traversalRevision"`
	Page                 int                                 `json:"page" minimum:"1" maximum:"9007199254740991"`
	Limit                int                                 `json:"limit" minimum:"1" maximum:"200"`
	Total                int64                               `json:"total" minimum:"0" maximum:"9007199254740991"`
	Coverage             CollectiveEvidenceCoverage          `json:"coverage"`
	RemoteAttempts       []LocalRemoteContributionAttempt    `json:"remoteAttempts,omitempty" maxItems:"200" nullable:"false"`
	Receipts             []LocalConfirmedContributionReceipt `json:"receipts,omitempty" maxItems:"200" nullable:"false"`
	OperationDiagnostics []LocalOperationDiagnostic          `json:"operationDiagnostics,omitempty" maxItems:"200" nullable:"false"`
}

func (r LocalContributionHistoryResponse) Validate() error {
	if err := r.Scope.Validate(); err != nil {
		return err
	}
	identity, err := ParseContributionTraversalRevision(r.TraversalRevision)
	if err != nil {
		return err
	}
	if identity.Scope != r.Scope || identity.Kind != r.Kind || !r.Coverage.IsValid() || r.Page < 1 || int64(r.Page) > 9007199254740991 || r.Limit < 1 || r.Limit > MaxContributionPageLimit || r.Total < 0 || r.Total > 9007199254740991 {
		return observationError("history requires matching scope/kind token, safe page/counts, limit 1..200 and known coverage")
	}
	n := 0
	switch r.Kind {
	case LocalContributionEvidenceRemoteAttempts:
		if r.RemoteAttempts == nil || r.Receipts != nil || r.OperationDiagnostics != nil {
			return observationError("remote-attempts must contain exactly remoteAttempts")
		}
		n = len(r.RemoteAttempts)
		for _, v := range r.RemoteAttempts {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	case LocalContributionEvidenceReceipts:
		if r.Receipts == nil || r.RemoteAttempts != nil || r.OperationDiagnostics != nil {
			return observationError("receipts must contain exactly receipts")
		}
		n = len(r.Receipts)
		for _, v := range r.Receipts {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	case LocalContributionEvidenceOperationDiagnostics:
		if r.OperationDiagnostics == nil || r.RemoteAttempts != nil || r.Receipts != nil {
			return observationError("operation-diagnostics must contain exactly operationDiagnostics")
		}
		n = len(r.OperationDiagnostics)
		for _, v := range r.OperationDiagnostics {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	default:
		return observationError("unknown evidence kind")
	}
	if n > r.Limit || int64(n) > r.Total {
		return observationError("selected evidence array exceeds page limit or filtered total")
	}
	return nil
}
func (r LocalContributionHistoryResponse) MarshalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	type plain LocalContributionHistoryResponse
	data, err := json.Marshal(plain(r))
	if err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err = json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	switch r.Kind {
	case LocalContributionEvidenceRemoteAttempts:
		object["remoteAttempts"], err = json.Marshal(r.RemoteAttempts)
	case LocalContributionEvidenceReceipts:
		object["receipts"], err = json.Marshal(r.Receipts)
	case LocalContributionEvidenceOperationDiagnostics:
		object["operationDiagnostics"], err = json.Marshal(r.OperationDiagnostics)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(object)
}
func (r *LocalContributionHistoryResponse) UnmarshalJSON(data []byte) error {
	type presence LocalContributionHistoryResponse
	var required presence
	if err := requireCollectiveFields(data, &required); err != nil {
		return err
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(data, &object) != nil {
		return observationError("history contains invalid JSON framing")
	}
	var kind LocalContributionEvidenceKind
	if json.Unmarshal(object["kind"], &kind) != nil {
		return observationError("history kind is missing or invalid")
	}
	selected := "remoteAttempts"
	if kind == LocalContributionEvidenceReceipts {
		selected = "receipts"
	} else if kind == LocalContributionEvidenceOperationDiagnostics {
		selected = "operationDiagnostics"
	}
	for _, key := range []string{"remoteAttempts", "receipts", "operationDiagnostics"} {
		if key != selected {
			if _, present := object[key]; present {
				return observationError("history contains a nonselected category array")
			}
		}
	}
	type plain LocalContributionHistoryResponse
	var v plain
	canonical, err := projectCollectiveJSON(data, reflect.TypeOf(required), true)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(canonical))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return observationError("history contains invalid fields or framing")
	}
	result := LocalContributionHistoryResponse(v)
	if err := result.Validate(); err != nil {
		return err
	}
	*r = result
	return nil
}

type LocalContributionHistoryQuery struct {
	Kind         LocalContributionEvidenceKind  `json:"kind"`
	CollectiveID *VillageUUID                   `json:"collectiveId,omitempty"`
	Page         int                            `json:"page"`
	Limit        int                            `json:"limit"`
	Revision     *ContributionTraversalRevision `json:"revision,omitempty"`
}

// ValidateContributionHistoryQuery returns 400 for malformed requests, 409 for
// a valid but different selected category token, and 200 otherwise. Tokens never
// select scope and no overall cache revision participates in this comparison.
func ValidateContributionHistoryQuery(q LocalContributionHistoryQuery, current ContributionTraversalRevision) int {
	if !q.Kind.IsValid() || q.Page < 1 || int64(q.Page) > 9007199254740991 || q.Limit < 1 || q.Limit > MaxContributionPageLimit || (q.CollectiveID != nil && !transcriptIDUUID.MatchString(string(*q.CollectiveID))) {
		return 400
	}
	if q.Revision == nil {
		if q.Page > 1 {
			return 400
		}
		return 200
	}
	if _, err := ParseContributionTraversalRevision(*q.Revision); err != nil {
		return 400
	}
	if *q.Revision != current {
		return 409
	}
	return 200
}

type LocalContributionTraversalConflict struct {
	CurrentTraversalRevision ContributionTraversalRevision `json:"currentTraversalRevision"`
	Issue                    OperationIssue                `json:"issue"`
}
