package schema

import (
	"fmt"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// --- Sync chooser status ---

// SyncStatus is the publication status the local sync chooser shows for one
// listed session. The server decides it; a client displays it and never infers
// a status from a row that is absent from the response.
type SyncStatus string

const (
	// SyncStatusNew: the session has no recorded publication.
	SyncStatusNew SyncStatus = "new"
	// SyncStatusUpdated: the session changed after its last publication.
	SyncStatusUpdated SyncStatus = "updated"
	// SyncStatusSynced: the last publication carries the current content.
	SyncStatusSynced SyncStatus = "synced"
	// SyncStatusHeld: the session cannot be published until ingest completes.
	// A held row always names its hold reason.
	SyncStatusHeld SyncStatus = "held"
)

// AllSyncStatuses is the canonical sync chooser status menu.
var AllSyncStatuses = []SyncStatus{SyncStatusNew, SyncStatusUpdated, SyncStatusSynced, SyncStatusHeld}

func (s SyncStatus) IsValid() bool  { return inSet(s, AllSyncStatuses) }
func (s SyncStatus) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (SyncStatus) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Sync Status", "Publication status of one listed local session; held rows always carry a hold reason", AllSyncStatuses), nil
}

// SyncHoldReason says why a held session cannot be published yet.
type SyncHoldReason string

const (
	// SyncHoldReasonMetricsMissing: the session has no computed metrics yet.
	SyncHoldReasonMetricsMissing SyncHoldReason = "metrics_missing"
	// SyncHoldReasonMetadataMissing: the session's publication metadata is not
	// ready, so the capture and the index do not agree yet.
	SyncHoldReasonMetadataMissing SyncHoldReason = "metadata_missing"
)

// AllSyncHoldReasons is the canonical hold reason menu.
var AllSyncHoldReasons = []SyncHoldReason{SyncHoldReasonMetricsMissing, SyncHoldReasonMetadataMissing}

func (r SyncHoldReason) IsValid() bool  { return inSet(r, AllSyncHoldReasons) }
func (r SyncHoldReason) String() string { return string(r) }

// JSONSchema implements jsonschema.Exposer.
func (SyncHoldReason) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Sync Hold Reason", "Why a held local session cannot be published yet", AllSyncHoldReasons), nil
}

// Validate checks the typed status, the hold reason pairing, and the count
// bounds of one sync chooser row. PreviouslyPushed is independent evidence: a
// held row can have been published before, so no status implies its value.
func (s LocalSyncSummary) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("local sync row validation failed at schema.LocalSyncSummary.Validate: id is empty; the chooser cannot select the session; emit the durable local session ID")
	}
	if !s.SyncStatus.IsValid() {
		return fmt.Errorf("local sync row validation failed for %q at schema.LocalSyncSummary.Validate: syncStatus %q is outside the closed set; the chooser cannot show a truthful status; emit a member of schema.AllSyncStatuses", s.ID, s.SyncStatus)
	}
	if s.SyncStatus == SyncStatusHeld && !s.HoldReason.IsValid() {
		return fmt.Errorf("local sync row validation failed for %q at schema.LocalSyncSummary.Validate: a held row has holdReason %q; the chooser cannot say why the session waits; emit a member of schema.AllSyncHoldReasons", s.ID, s.HoldReason)
	}
	if s.SyncStatus != SyncStatusHeld && s.HoldReason != "" {
		return fmt.Errorf("local sync row validation failed for %q at schema.LocalSyncSummary.Validate: syncStatus %q carries holdReason %q; only a held row has a hold reason; omit holdReason", s.ID, s.SyncStatus, s.HoldReason)
	}
	return ValidateInputSubmissionCount(s.InputSubmissionCount, "LocalSyncSummary.inputSubmissionCount")
}

// Validate checks that the flat sync list is an array of valid rows and lists
// each session once.
func (p LocalSyncSessionsPayload) Validate() error {
	if p.Sessions == nil {
		return fmt.Errorf("local sync list validation failed at schema.LocalSyncSessionsPayload.Validate: sessions is null; the chooser cannot tell an empty list from a failure; emit []")
	}
	seen := make(map[string]struct{}, len(p.Sessions))
	for _, row := range p.Sessions {
		if err := row.Validate(); err != nil {
			return err
		}
		if _, duplicate := seen[row.ID]; duplicate {
			return fmt.Errorf("local sync list validation failed at schema.LocalSyncSessionsPayload.Validate: session %q is listed twice; the chooser would count it twice; emit each session once", row.ID)
		}
		seen[row.ID] = struct{}{}
	}
	return nil
}

// --- Push ---

// SyncPushCollectives names the collectives one push adds and removes. Add
// shares each published transcript with the collective; remove takes it back.
// Collectives named in neither list keep their current access.
type SyncPushCollectives struct {
	Add    []VillageUUID `json:"add,omitempty" nullable:"false" description:"Collectives to share the published transcripts with"`
	Remove []VillageUUID `json:"remove,omitempty" nullable:"false" description:"Collectives to take the published transcripts back from"`
}

// SyncPushRequest is the body of POST /api/v1/sync/push. Publishing is
// collectives only: the request carries no visibility and no license, and the
// server applies no default license to a publish that names collectives.
type SyncPushRequest struct {
	SessionIDs []string `json:"sessionIds" nullable:"false" minItems:"1" description:"Local sessions to publish or update"`
	// RedactionLevel is a request for a level. Omission lets the server resolve
	// its configured default; the server refuses a level it does not offer.
	RedactionLevel string               `json:"redactionLevel,omitempty" description:"Requested redaction level; omit to use the configured default"`
	Collectives    *SyncPushCollectives `json:"collectives,omitempty" nullable:"false" description:"Audience changes; omit to keep each transcript's current audience"`
}

// Validate checks the session list and that no collective is both added and
// removed.
func (r SyncPushRequest) Validate() error {
	if len(r.SessionIDs) == 0 {
		return fmt.Errorf("sync push validation failed at schema.SyncPushRequest.Validate: sessionIds is empty; nothing would be published; name at least one local session")
	}
	seen := make(map[string]struct{}, len(r.SessionIDs))
	for _, id := range r.SessionIDs {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("sync push validation failed at schema.SyncPushRequest.Validate: sessionIds holds an empty ID; the server cannot select a session for it; remove the empty entry")
		}
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("sync push validation failed at schema.SyncPushRequest.Validate: session %q is named twice; one push publishes each session once; remove the duplicate", id)
		}
		seen[id] = struct{}{}
	}
	if r.Collectives == nil {
		return nil
	}
	added := make(map[VillageUUID]struct{}, len(r.Collectives.Add))
	for _, id := range r.Collectives.Add {
		if _, duplicate := added[id]; duplicate {
			return fmt.Errorf("sync push validation failed at schema.SyncPushRequest.Validate: collectives.add names %q twice; remove the duplicate", id)
		}
		added[id] = struct{}{}
	}
	removed := make(map[VillageUUID]struct{}, len(r.Collectives.Remove))
	for _, id := range r.Collectives.Remove {
		if _, duplicate := removed[id]; duplicate {
			return fmt.Errorf("sync push validation failed at schema.SyncPushRequest.Validate: collectives.remove names %q twice; remove the duplicate", id)
		}
		if _, both := added[id]; both {
			return fmt.Errorf("sync push validation failed at schema.SyncPushRequest.Validate: collective %q is both added and removed; the resulting access is ambiguous; name it in one list only", id)
		}
		removed[id] = struct{}{}
	}
	return nil
}

// SyncPushSessionStatus is the outcome of one session in a push.
type SyncPushSessionStatus string

const (
	SyncPushSessionNew     SyncPushSessionStatus = "new"
	SyncPushSessionUpdated SyncPushSessionStatus = "updated"
	SyncPushSessionSkipped SyncPushSessionStatus = "skipped"
	SyncPushSessionError   SyncPushSessionStatus = "error"
	SyncPushSessionHeld    SyncPushSessionStatus = "held"
)

// AllSyncPushSessionStatuses is the canonical per-session push outcome menu.
var AllSyncPushSessionStatuses = []SyncPushSessionStatus{SyncPushSessionNew, SyncPushSessionUpdated, SyncPushSessionSkipped, SyncPushSessionError, SyncPushSessionHeld}

func (s SyncPushSessionStatus) IsValid() bool  { return inSet(s, AllSyncPushSessionStatuses) }
func (s SyncPushSessionStatus) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (SyncPushSessionStatus) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Sync Push Session Status", "Outcome of one session in a push", AllSyncPushSessionStatuses), nil
}

// SyncPushStep is one step a push runs for a session: the content step, then
// one step per collective added or removed.
type SyncPushStep string

const (
	// SyncPushStepContent sends the redacted transcript to Village.
	SyncPushStepContent SyncPushStep = "content"
	// SyncPushStepAddCollective shares the transcript with one collective.
	SyncPushStepAddCollective SyncPushStep = "add_collective"
	// SyncPushStepRemoveCollective takes the transcript back from one collective.
	SyncPushStepRemoveCollective SyncPushStep = "remove_collective"
)

// AllSyncPushSteps is the canonical push step menu.
var AllSyncPushSteps = []SyncPushStep{SyncPushStepContent, SyncPushStepAddCollective, SyncPushStepRemoveCollective}

func (s SyncPushStep) IsValid() bool  { return inSet(s, AllSyncPushSteps) }
func (s SyncPushStep) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (SyncPushStep) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Sync Push Step", "One step a push runs for a session", AllSyncPushSteps), nil
}

// SyncPushStepOutcome is the result of one push step.
type SyncPushStepOutcome string

const (
	// SyncPushStepSucceeded: Village applied the step.
	SyncPushStepSucceeded SyncPushStepOutcome = "succeeded"
	// SyncPushStepPendingApproval: the collective holds the share until its
	// owner approves it. Only an add_collective step has this outcome.
	SyncPushStepPendingApproval SyncPushStepOutcome = "pending_approval"
	// SyncPushStepSkipped: the step was not needed or not allowed, for example
	// a collective that already holds the transcript. It is not a failure and
	// carries its reason.
	SyncPushStepSkipped SyncPushStepOutcome = "skipped"
	// SyncPushStepFailed: the step failed, Village keeps what it had for it, and
	// the step carries its reason.
	SyncPushStepFailed SyncPushStepOutcome = "failed"
	// SyncPushStepNotAttempted: the step did not run because an earlier step of
	// the same session failed.
	SyncPushStepNotAttempted SyncPushStepOutcome = "not_attempted"
)

// AllSyncPushStepOutcomes is the canonical push step outcome menu.
var AllSyncPushStepOutcomes = []SyncPushStepOutcome{SyncPushStepSucceeded, SyncPushStepPendingApproval, SyncPushStepSkipped, SyncPushStepFailed, SyncPushStepNotAttempted}

func (o SyncPushStepOutcome) IsValid() bool  { return inSet(o, AllSyncPushStepOutcomes) }
func (o SyncPushStepOutcome) String() string { return string(o) }

// JSONSchema implements jsonschema.Exposer.
func (SyncPushStepOutcome) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Sync Push Step Outcome", "Result of one push step", AllSyncPushStepOutcomes), nil
}

// SyncPushStepResult is the result of one step for one session.
type SyncPushStepResult struct {
	Step SyncPushStep `json:"step"`
	// CollectiveID names the collective of an add_collective or
	// remove_collective step. The content step has none.
	CollectiveID *VillageUUID        `json:"collectiveId,omitempty" nullable:"false"`
	Outcome      SyncPushStepOutcome `json:"outcome"`
	// Reason says why a skipped or failed step did not apply. No other outcome
	// carries one.
	Reason string `json:"reason,omitempty"`
}

// Validate checks the step, its collective, and the outcome pairing.
func (r SyncPushStepResult) Validate() error {
	if !r.Step.IsValid() {
		return fmt.Errorf("sync push step validation failed at schema.SyncPushStepResult.Validate: step %q is outside the closed set; emit a member of schema.AllSyncPushSteps", r.Step)
	}
	if !r.Outcome.IsValid() {
		return fmt.Errorf("sync push step validation failed at schema.SyncPushStepResult.Validate: outcome %q is outside the closed set; emit a member of schema.AllSyncPushStepOutcomes", r.Outcome)
	}
	if (r.Step == SyncPushStepContent) != (r.CollectiveID == nil) {
		return fmt.Errorf("sync push step validation failed at schema.SyncPushStepResult.Validate: step %q has the wrong collectiveId presence; a collective step names its collective and the content step names none", r.Step)
	}
	if r.Outcome == SyncPushStepPendingApproval && r.Step != SyncPushStepAddCollective {
		return fmt.Errorf("sync push step validation failed at schema.SyncPushStepResult.Validate: step %q is pending_approval; only adding a collective can wait for its owner", r.Step)
	}
	explained := r.Outcome == SyncPushStepSkipped || r.Outcome == SyncPushStepFailed
	if explained != (strings.TrimSpace(r.Reason) != "") {
		return fmt.Errorf("sync push step validation failed at schema.SyncPushStepResult.Validate: outcome %q has the wrong reason presence; a skipped or failed step says why and no other step carries a reason", r.Outcome)
	}
	return nil
}

// SyncPushSessionResult is the result for one session in a push.
type SyncPushSessionResult struct {
	SessionID string `json:"sessionId"`
	// Status is the session's outcome. It describes the content: new or
	// updated when Village accepted it, skipped when it was already current
	// (also when only collectives changed), held when it cannot be published
	// yet, and error when any step failed or the session failed before a step
	// ran. The response counts tally sessions by this status.
	Status SyncPushSessionStatus `json:"status"`
	Error  string                `json:"error,omitempty"`
	Title  string                `json:"title,omitempty"`
	// TranscriptURL is the Village page of the session's transcript. A new or
	// updated session always carries it.
	TranscriptURL string `json:"transcriptUrl,omitempty"`
	// Steps lists every step the push planned for this session, in the order
	// it ran them. When present, the content step comes first: skipped when
	// the content was already current or the session is held.
	Steps []SyncPushStepResult `json:"steps,omitempty" nullable:"false"`
	// WaitingPullRequests are the caller's Village prompt requests that wait
	// for a transcript from this session's repository.
	WaitingPullRequests []VillagePromptRequest `json:"waitingPullRequests,omitempty" nullable:"false"`
}

// Validate checks one session result, its steps, and how the status follows
// from them. Steps start with the content step. A failed step makes the
// session an error, and an error with steps names the step that failed. New
// and updated sessions have a succeeded content step and a transcript URL;
// skipped and held sessions have a skipped content step, and a held session
// can only take a transcript back from collectives. A step is not_attempted
// only after an earlier step failed. Whether a later step still runs after a
// failure is the producer's choice, so taking a collective back is never
// forced to wait on an unrelated failure.
func (r SyncPushSessionResult) Validate() error {
	if strings.TrimSpace(r.SessionID) == "" {
		return fmt.Errorf("sync push result validation failed at schema.SyncPushSessionResult.Validate: sessionId is empty; the caller cannot match the result to a session; emit the requested session ID")
	}
	if !r.Status.IsValid() {
		return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: status %q is outside the closed set; emit a member of schema.AllSyncPushSessionStatuses", r.SessionID, r.Status)
	}
	if r.Status == SyncPushSessionError && strings.TrimSpace(r.Error) == "" {
		return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: status error has no error text; the caller cannot say what failed; emit the reason", r.SessionID)
	}
	collectives := make(map[VillageUUID]struct{}, len(r.Steps))
	failed := false
	for i, step := range r.Steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("sync push result validation failed for %q at steps[%d]: %w", r.SessionID, i, err)
		}
		if step.Step == SyncPushStepContent && i != 0 {
			return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: steps[%d] is a content step after another step; content runs once and first", r.SessionID, i)
		}
		if step.CollectiveID != nil {
			if _, duplicate := collectives[*step.CollectiveID]; duplicate {
				return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: collective %q has two steps; each collective is added or removed once", r.SessionID, *step.CollectiveID)
			}
			collectives[*step.CollectiveID] = struct{}{}
		}
		if step.Outcome == SyncPushStepNotAttempted && !failed {
			return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: steps[%d] is not_attempted with no earlier failed step; a step only goes unattempted because an earlier step failed", r.SessionID, i)
		}
		failed = failed || step.Outcome == SyncPushStepFailed
	}
	if failed && r.Status != SyncPushSessionError {
		return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: a step failed but status is %q; a session with a failed step is an error, so the counts cannot report success", r.SessionID, r.Status)
	}
	if r.Status == SyncPushSessionError && len(r.Steps) > 0 && !failed {
		return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: status error lists steps but none failed; name the step that failed, or omit steps when the session failed before any step ran", r.SessionID)
	}
	if (r.Status == SyncPushSessionNew || r.Status == SyncPushSessionUpdated) && strings.TrimSpace(r.TranscriptURL) == "" {
		return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: status %q has no transcriptUrl; Village accepted the content, so the caller can open it; emit the transcript URL", r.SessionID, r.Status)
	}
	if len(r.Steps) == 0 {
		if r.Status == SyncPushSessionNew || r.Status == SyncPushSessionUpdated {
			return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: status %q lists no steps; list the content step that Village accepted", r.SessionID, r.Status)
		}
		return nil
	}
	content := r.Steps[0]
	if content.Step != SyncPushStepContent {
		return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: steps start with %q; a session that lists steps lists the content step first, skipped when the content was already current", r.SessionID, content.Step)
	}
	switch r.Status {
	case SyncPushSessionNew, SyncPushSessionUpdated:
		if content.Outcome != SyncPushStepSucceeded {
			return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: status %q has a %s content step; new and updated mean Village accepted the content", r.SessionID, r.Status, content.Outcome)
		}
	case SyncPushSessionSkipped, SyncPushSessionHeld:
		if content.Outcome != SyncPushStepSkipped {
			return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: status %q has a %s content step; skipped and held mean the content was not sent", r.SessionID, r.Status, content.Outcome)
		}
	}
	if r.Status == SyncPushSessionHeld {
		for _, step := range r.Steps[1:] {
			if step.Step != SyncPushStepRemoveCollective {
				return fmt.Errorf("sync push result validation failed for %q at schema.SyncPushSessionResult.Validate: a held session has a %s step; a held session can only be taken back from collectives", r.SessionID, step.Step)
			}
		}
	}
	return nil
}

// SyncPushResponse is the response of POST /api/v1/sync/push: the counts the
// push already returned plus one result per session. New, Updated, Skipped,
// and Errors count the sessions with that status; held sessions are not
// counted.
type SyncPushResponse struct {
	New      int                     `json:"new" minimum:"0"`
	Updated  int                     `json:"updated" minimum:"0"`
	Skipped  int                     `json:"skipped" minimum:"0"`
	Errors   int                     `json:"errors" minimum:"0"`
	Sessions []SyncPushSessionResult `json:"sessions" nullable:"false"`
}

// Validate checks the counts and every session result.
func (r SyncPushResponse) Validate() error {
	if r.New < 0 || r.Updated < 0 || r.Skipped < 0 || r.Errors < 0 {
		return fmt.Errorf("sync push response validation failed at schema.SyncPushResponse.Validate: a count is negative; emit nonnegative counts")
	}
	if r.Sessions == nil {
		return fmt.Errorf("sync push response validation failed at schema.SyncPushResponse.Validate: sessions is null; emit [] when no session ran")
	}
	seen := make(map[string]struct{}, len(r.Sessions))
	tally := map[SyncPushSessionStatus]int{}
	for _, session := range r.Sessions {
		if err := session.Validate(); err != nil {
			return err
		}
		if _, duplicate := seen[session.SessionID]; duplicate {
			return fmt.Errorf("sync push response validation failed at schema.SyncPushResponse.Validate: session %q has two results; emit one result per session", session.SessionID)
		}
		seen[session.SessionID] = struct{}{}
		tally[session.Status]++
	}
	if r.New != tally[SyncPushSessionNew] || r.Updated != tally[SyncPushSessionUpdated] || r.Skipped != tally[SyncPushSessionSkipped] || r.Errors != tally[SyncPushSessionError] {
		return fmt.Errorf("sync push response validation failed at schema.SyncPushResponse.Validate: counts new=%d updated=%d skipped=%d errors=%d disagree with the session statuses; a banner built from the counts would misreport the push; count sessions by status after every step ran (held sessions are not counted)", r.New, r.Updated, r.Skipped, r.Errors)
	}
	return nil
}

// --- Village sign-in ---

// SyncAuthResponse is the response of GET /api/v1/sync/auth: whether this
// computer holds a valid Village credential.
type SyncAuthResponse struct {
	Authenticated bool `json:"authenticated"`
	// Username, VillageURL, and VillageConfigured are present only when
	// Authenticated is true.
	Username          string `json:"username,omitempty"`
	VillageURL        string `json:"villageUrl,omitempty"`
	VillageConfigured bool   `json:"villageConfigured,omitempty"`
}

// Validate checks that a signed-out response names no account.
func (r SyncAuthResponse) Validate() error {
	if !r.Authenticated && (r.Username != "" || r.VillageURL != "" || r.VillageConfigured) {
		return fmt.Errorf("sync auth validation failed at schema.SyncAuthResponse.Validate: authenticated is false but the response names an account or Village; a signed-out computer has neither; omit username, villageUrl, and villageConfigured")
	}
	return nil
}

// SyncLoginStatus is the result of POST /api/v1/sync/login.
type SyncLoginStatus string

const (
	// SyncLoginPending: the browser sign-in started; poll /api/v1/sync/auth.
	SyncLoginPending SyncLoginStatus = "pending"
	// SyncLoginAlreadyAuthenticated: this computer is already signed in.
	SyncLoginAlreadyAuthenticated SyncLoginStatus = "already_authenticated"
)

// AllSyncLoginStatuses is the canonical sign-in result menu.
var AllSyncLoginStatuses = []SyncLoginStatus{SyncLoginPending, SyncLoginAlreadyAuthenticated}

func (s SyncLoginStatus) IsValid() bool  { return inSet(s, AllSyncLoginStatuses) }
func (s SyncLoginStatus) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (SyncLoginStatus) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Sync Login Status", "Result of starting the Village sign-in from the local server", AllSyncLoginStatuses), nil
}

// SyncLoginResponse is the response of POST /api/v1/sync/login.
type SyncLoginResponse struct {
	Status SyncLoginStatus `json:"status"`
}

// Validate checks the sign-in result.
func (r SyncLoginResponse) Validate() error {
	if !r.Status.IsValid() {
		return fmt.Errorf("sync login validation failed at schema.SyncLoginResponse.Validate: status %q is outside the closed set; emit a member of schema.AllSyncLoginStatuses", r.Status)
	}
	return nil
}

// SyncLogoutStatus is the result of POST /api/v1/sync/logout.
type SyncLogoutStatus string

const (
	// SyncLogoutLoggedOut: the stored Village credential was removed.
	SyncLogoutLoggedOut SyncLogoutStatus = "logged_out"
	// SyncLogoutAlreadyLoggedOut: this computer held no Village credential.
	SyncLogoutAlreadyLoggedOut SyncLogoutStatus = "already_logged_out"
)

// AllSyncLogoutStatuses is the canonical sign-out result menu.
var AllSyncLogoutStatuses = []SyncLogoutStatus{SyncLogoutLoggedOut, SyncLogoutAlreadyLoggedOut}

func (s SyncLogoutStatus) IsValid() bool  { return inSet(s, AllSyncLogoutStatuses) }
func (s SyncLogoutStatus) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (SyncLogoutStatus) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Sync Logout Status", "Result of ending this computer's Village sign-in", AllSyncLogoutStatuses), nil
}

// SyncLogoutResponse is the response of POST /api/v1/sync/logout.
type SyncLogoutResponse struct {
	Status SyncLogoutStatus `json:"status"`
}

// Validate checks the sign-out result.
func (r SyncLogoutResponse) Validate() error {
	if !r.Status.IsValid() {
		return fmt.Errorf("sync logout validation failed at schema.SyncLogoutResponse.Validate: status %q is outside the closed set; emit a member of schema.AllSyncLogoutStatuses", r.Status)
	}
	return nil
}

// --- Redaction preview ---

// SyncRedactionItem is one redaction match in the preview of what leaves the
// machine.
type SyncRedactionItem struct {
	// Category is the redaction engine's category label, rendered verbatim.
	Category            string `json:"category"`
	RuleID              string `json:"ruleId"`
	RuleDisplayName     string `json:"ruleDisplayName"`
	OriginalText        string `json:"originalText"`
	RedactedReplacement string `json:"redactedReplacement"`
	Description         string `json:"description"`
	LineNumber          int    `json:"lineNumber" minimum:"0"`
	// EntryIndex is the TurnDetail.index of the turn that shows the match, so
	// a client can name and open the turn without projecting entries itself. A
	// match in a tool call's arguments or result names the turn that owns the
	// tool call, and ToolCallID names the call. One item stands for every
	// occurrence of the same text under the same rule, and EntryIndex names the
	// first. It is absent when the match lies outside every turn, for example
	// in session metadata.
	EntryIndex *int `json:"entryIndex,omitempty" minimum:"0" nullable:"false"`
	// ToolCallID is the ToolCallDetail.id of the tool call that holds the
	// match, present only when the match lies in a tool call.
	ToolCallID    string   `json:"toolCallId,omitempty"`
	ContextBefore []string `json:"contextBefore" nullable:"false"`
	ContextAfter  []string `json:"contextAfter" nullable:"false"`
}

// SyncRedactionRuleGroup holds the deduplicated matches of one rule. Items is
// capped by the server, so it can hold fewer entries than Count.
type SyncRedactionRuleGroup struct {
	RuleID      string              `json:"ruleId"`
	DisplayName string              `json:"displayName"`
	Count       int                 `json:"count" minimum:"0"`
	Items       []SyncRedactionItem `json:"items" nullable:"false"`
}

// SyncRedactionCategoryGroup holds the rule groups of one category.
type SyncRedactionCategoryGroup struct {
	Category   string                   `json:"category"`
	TotalCount int                      `json:"totalCount" minimum:"0"`
	Rules      []SyncRedactionRuleGroup `json:"rules" nullable:"false"`
}

// SyncRedactionsResponse is the response of GET /api/v1/sync/redactions. Total
// counts every raw match; the groups count deduplicated matches.
type SyncRedactionsResponse struct {
	Total      int                          `json:"total" minimum:"0"`
	Categories []SyncRedactionCategoryGroup `json:"categories" nullable:"false"`
}

// Validate checks the counts, the item cap, and that each item sits in the
// category and rule group that holds it.
func (r SyncRedactionsResponse) Validate() error {
	if r.Total < 0 || r.Categories == nil {
		return fmt.Errorf("redaction preview validation failed at schema.SyncRedactionsResponse.Validate: total is negative or categories is null; emit a nonnegative total and [] when nothing matched")
	}
	for _, category := range r.Categories {
		if category.TotalCount < 0 || category.Rules == nil {
			return fmt.Errorf("redaction preview validation failed at schema.SyncRedactionsResponse.Validate: category %q has a negative totalCount or null rules; emit a nonnegative count and an array", category.Category)
		}
		for _, rule := range category.Rules {
			if rule.Count < 0 || rule.Items == nil || len(rule.Items) > rule.Count {
				return fmt.Errorf("redaction preview validation failed at schema.SyncRedactionsResponse.Validate: rule %q has a negative count, null items, or more items than its count; the preview would overstate or hide matches", rule.RuleID)
			}
			for _, item := range rule.Items {
				if item.Category != category.Category || item.RuleID != rule.RuleID {
					return fmt.Errorf("redaction preview validation failed at schema.SyncRedactionsResponse.Validate: an item of rule %q sits in the wrong group; group each match under its own category and rule", item.RuleID)
				}
				if item.ToolCallID != "" && item.EntryIndex == nil {
					return fmt.Errorf("redaction preview validation failed at schema.SyncRedactionsResponse.Validate: an item of rule %q names tool call %q but no turn; a tool call match names the turn that owns the call", item.RuleID, item.ToolCallID)
				}
				if item.LineNumber < 0 || (item.EntryIndex != nil && *item.EntryIndex < 0) || item.ContextBefore == nil || item.ContextAfter == nil {
					return fmt.Errorf("redaction preview validation failed at schema.SyncRedactionsResponse.Validate: an item of rule %q has a negative position or null context; emit nonnegative positions and arrays", item.RuleID)
				}
			}
		}
	}
	return nil
}
