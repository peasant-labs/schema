package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// --- Settings ---

// LocalSettingKind is the JSON type of one setting's value. Peasant owns the
// key set, each key's description, and which keys are editable; this contract
// owns the value shapes.
type LocalSettingKind string

const (
	// LocalSettingBoolean: the value is a JSON boolean.
	LocalSettingBoolean LocalSettingKind = "boolean"
	// LocalSettingInteger: the value is a JSON integer.
	LocalSettingInteger LocalSettingKind = "integer"
	// LocalSettingString: the value is a JSON string.
	LocalSettingString LocalSettingKind = "string"
	// LocalSettingStringList: the value is a JSON array of strings.
	LocalSettingStringList LocalSettingKind = "string_list"
	// LocalSettingChoice: the value is one of the setting's options.
	LocalSettingChoice LocalSettingKind = "choice"
	// LocalSettingStructured: the value is the key's configuration value as a
	// JSON object or array, for example a list of custom redaction patterns.
	// It is only for a list or map of records; Peasant owns its shape. A key
	// whose value fits another kind uses that kind.
	LocalSettingStructured LocalSettingKind = "structured"
)

// AllLocalSettingKinds is the canonical setting value kind menu.
var AllLocalSettingKinds = []LocalSettingKind{LocalSettingBoolean, LocalSettingInteger, LocalSettingString, LocalSettingStringList, LocalSettingChoice, LocalSettingStructured}

func (k LocalSettingKind) IsValid() bool  { return inSet(k, AllLocalSettingKinds) }
func (k LocalSettingKind) String() string { return string(k) }

// JSONSchema implements jsonschema.Exposer.
func (LocalSettingKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Local Setting Kind", "JSON type of one local setting's value", AllLocalSettingKinds), nil
}

// LocalSettingValue is one setting's JSON value, held as its exact JSON text.
// The setting's kind chooses which JSON type it is. JSON null means the key is
// unset: the configuration file does not name it and the server's default
// applies. In an update, null unsets the key.
type LocalSettingValue []byte

// MarshalJSON emits the held JSON text.
func (v LocalSettingValue) MarshalJSON() ([]byte, error) {
	if len(v) == 0 {
		return []byte("null"), nil
	}
	return append([]byte(nil), v...), nil
}

// UnmarshalJSON keeps the JSON text as received, in fresh storage so decoding
// never writes into a value another variable shares.
func (v *LocalSettingValue) UnmarshalJSON(data []byte) error {
	*v = append(LocalSettingValue(nil), data...)
	return nil
}

// JSONSchema implements jsonschema.Exposer.
func (LocalSettingValue) JSONSchema() (jsonschema.Schema, error) {
	arms := make([]jsonschema.SchemaOrBool, 0, 6)
	for _, kind := range []jsonschema.SimpleType{jsonschema.Null, jsonschema.Boolean, jsonschema.Integer, jsonschema.String, jsonschema.Array, jsonschema.Object} {
		arm := jsonschema.Schema{}
		arm.AddType(kind)
		arms = append(arms, arm.ToSchemaOrBool())
	}
	s := jsonschema.Schema{}
	s.WithTitle("Local Setting Value")
	s.WithDescription("One setting's value, of the JSON type its kind names; null means the key is unset and the server's default applies")
	s.WithAnyOf(arms...)
	return s, nil
}

// LocalSettingNull returns a fresh JSON null, which a producer sets for an
// unset value or when no effective value applies.
func LocalSettingNull() LocalSettingValue { return LocalSettingValue("null") }

// IsUnset reports whether the value is JSON null. A nil value is unset too:
// it marshals as null.
func (v LocalSettingValue) IsUnset() bool {
	trimmed := bytes.TrimSpace(v)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

// ValidateFor checks that a set value has the JSON type kind names and, for a
// choice, is one of options. An unset value, nil or JSON null, is valid for
// every kind. A producer checks an update's value against the key's kind with
// this method.
func (v LocalSettingValue) ValidateFor(kind LocalSettingKind, options []string) error {
	return v.validate(kind, options, true)
}

// validate checks the JSON type kind names; menu reports whether a choice must
// also be one of options.
func (v LocalSettingValue) validate(kind LocalSettingKind, options []string, menu bool) error {
	if v.IsUnset() {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(v))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("the value is not JSON: %v", err)
	}
	switch kind {
	case LocalSettingBoolean:
		if _, ok := value.(bool); ok {
			return nil
		}
	case LocalSettingInteger:
		if number, ok := value.(json.Number); ok {
			if _, err := number.Int64(); err == nil {
				return nil
			}
		}
	case LocalSettingString:
		if _, ok := value.(string); ok {
			return nil
		}
	case LocalSettingStringList:
		if items, ok := value.([]any); ok {
			for _, item := range items {
				if _, isString := item.(string); !isString {
					return fmt.Errorf("the value holds a non-string item")
				}
			}
			return nil
		}
	case LocalSettingChoice:
		if text, ok := value.(string); ok && (!menu || inSet(text, options)) {
			return nil
		}
	case LocalSettingStructured:
		switch value.(type) {
		case map[string]any, []any:
			return nil
		}
	}
	return fmt.Errorf("the value is not a %s", kind)
}

// LocalSetting is one Peasant setting with its metadata.
type LocalSetting struct {
	// Key is the setting's dotted configuration path, for example
	// push.concurrency.
	Key  string           `json:"key" minLength:"1"`
	Kind LocalSettingKind `json:"kind"`
	// Value is what the configuration file names for the key, or null when
	// the file does not name it. A choice value may lie outside Options when
	// the server still accepts it from an existing file but no longer offers
	// it; Effective then names what applies.
	Value LocalSettingValue `json:"value"`
	// Effective is the value that applies now: normally Value when the key is
	// set, otherwise the server's default, and it can differ from Value when
	// something outside the file, such as an environment variable, overrides
	// it. It is null only when no value applies. A client shows Effective, so
	// an unset key never reads as off or blank when its default is on or
	// filled.
	Effective LocalSettingValue `json:"effective"`
	// Options is the menu of a choice setting; an option may be the empty
	// string when that is a meaningful setting. Other kinds have none.
	Options []string `json:"options,omitempty" nullable:"false"`
	// Editable reports that PATCH /api/v1/settings accepts this key. A key
	// that changes another way, such as village.connected through sign-in and
	// sign-out, is shown but not editable.
	Editable bool `json:"editable"`
	// InPeasantConfig reports that the terminal editor `peasant config` can
	// also change this setting.
	InPeasantConfig bool `json:"inPeasantConfig"`
	// Description is Peasant's product text for the setting.
	Description string `json:"description,omitempty"`
}

// Validate checks the key, the kind, the options, and that the value matches
// the kind.
func (s LocalSetting) Validate() error {
	if strings.TrimSpace(s.Key) == "" {
		return fmt.Errorf("setting validation failed at schema.LocalSetting.Validate: key is empty; the settings page cannot address the setting; emit its dotted configuration path")
	}
	if !s.Kind.IsValid() {
		return fmt.Errorf("setting validation failed for %q at schema.LocalSetting.Validate: kind %q is outside the closed set; emit a member of schema.AllLocalSettingKinds", s.Key, s.Kind)
	}
	if (s.Kind == LocalSettingChoice) != (len(s.Options) > 0) {
		return fmt.Errorf("setting validation failed for %q at schema.LocalSetting.Validate: kind %q has the wrong options; a choice setting lists its options and no other kind has any", s.Key, s.Kind)
	}
	seen := make(map[string]struct{}, len(s.Options))
	for _, option := range s.Options {
		if _, duplicate := seen[option]; duplicate {
			return fmt.Errorf("setting validation failed for %q at schema.LocalSetting.Validate: option %q is repeated; list each option once", s.Key, option)
		}
		seen[option] = struct{}{}
	}
	if err := s.Value.validate(s.Kind, nil, false); err != nil {
		return fmt.Errorf("setting validation failed for %q at schema.LocalSetting.Validate: %v; emit a value of kind %q or null when the key is unset", s.Key, err, s.Kind)
	}
	if s.Effective == nil {
		return fmt.Errorf("setting validation failed for %q at schema.LocalSetting.Validate: effective is missing; the settings page shows what applies; emit it, or schema.LocalSettingNull() when no value applies", s.Key)
	}
	if err := s.Effective.ValidateFor(s.Kind, s.Options); err != nil {
		return fmt.Errorf("setting validation failed for %q at schema.LocalSetting.Validate: effective %v; emit the value that applies, of kind %q", s.Key, err, s.Kind)
	}
	if !s.Value.IsUnset() && s.Effective.IsUnset() {
		return fmt.Errorf("setting validation failed for %q at schema.LocalSetting.Validate: a set value has a null effective; a value the file names always applies or is replaced; emit the value that applies", s.Key)
	}
	return nil
}

// LocalSettingUpdateRequest is the body of PATCH /api/v1/settings. One request
// changes one key; a null value unsets it so the server's default applies.
type LocalSettingUpdateRequest struct {
	Key   string            `json:"key" minLength:"1"`
	Value LocalSettingValue `json:"value"`
}

// Validate checks that the request names a key and carries a value. The
// server checks the value against the key's kind with ValidateFor.
func (r LocalSettingUpdateRequest) Validate() error {
	if strings.TrimSpace(r.Key) == "" {
		return fmt.Errorf("setting update validation failed at schema.LocalSettingUpdateRequest.Validate: key is empty; nothing would change; name the setting's dotted configuration path")
	}
	if len(bytes.TrimSpace(r.Value)) == 0 {
		return fmt.Errorf("setting update validation failed for %q at schema.LocalSettingUpdateRequest.Validate: value is missing; send the new value, or null to unset the key", r.Key)
	}
	return nil
}

// LocalSettingRefusal is the body of a refused setting update: the key and
// why it was not changed. A refused update changes nothing.
type LocalSettingRefusal struct {
	Key   string `json:"key" minLength:"1"`
	Error string `json:"error" minLength:"1"`
}

// Validate checks that the refusal names the key and the reason.
func (r LocalSettingRefusal) Validate() error {
	if strings.TrimSpace(r.Key) == "" || strings.TrimSpace(r.Error) == "" {
		return fmt.Errorf("setting refusal validation failed at schema.LocalSettingRefusal.Validate: key or error is empty; the settings page restores the value and shows the reason; emit both")
	}
	return nil
}

// LocalSettingsResponse is the response of GET /api/v1/settings: every
// setting the settings page shows and every auto-publish rule.
type LocalSettingsResponse struct {
	Settings    []LocalSetting    `json:"settings" nullable:"false"`
	AutoPublish []AutoPublishRule `json:"autoPublish" nullable:"false"`
}

// Validate checks every setting and rule and that keys and identifiers do not
// repeat.
func (r LocalSettingsResponse) Validate() error {
	if r.Settings == nil || r.AutoPublish == nil {
		return fmt.Errorf("settings validation failed at schema.LocalSettingsResponse.Validate: settings or autoPublish is null; emit [] for an empty list")
	}
	keys := make(map[string]struct{}, len(r.Settings))
	for _, setting := range r.Settings {
		if err := setting.Validate(); err != nil {
			return err
		}
		if _, duplicate := keys[setting.Key]; duplicate {
			return fmt.Errorf("settings validation failed at schema.LocalSettingsResponse.Validate: key %q is listed twice; list each setting once", setting.Key)
		}
		keys[setting.Key] = struct{}{}
	}
	ids := make(map[string]struct{}, len(r.AutoPublish))
	for _, rule := range r.AutoPublish {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, duplicate := ids[rule.ID]; duplicate {
			return fmt.Errorf("settings validation failed at schema.LocalSettingsResponse.Validate: rule %q is listed twice; list each rule once", rule.ID)
		}
		ids[rule.ID] = struct{}{}
	}
	return nil
}

// --- Auto-publish ---

// AutoPublishEvent is the git hook that publishes a rule's sessions.
type AutoPublishEvent string

const (
	AutoPublishPrePush    AutoPublishEvent = "pre-push"
	AutoPublishPostCommit AutoPublishEvent = "post-commit"
)

// AllAutoPublishEvents is the canonical auto-publish hook menu.
var AllAutoPublishEvents = []AutoPublishEvent{AutoPublishPrePush, AutoPublishPostCommit}

func (e AutoPublishEvent) IsValid() bool  { return inSet(e, AllAutoPublishEvents) }
func (e AutoPublishEvent) String() string { return string(e) }

// JSONSchema implements jsonschema.Exposer.
func (AutoPublishEvent) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Auto Publish Event", "Git hook that publishes an auto-publish rule's sessions", AllAutoPublishEvents), nil
}

// AutoPublishHookStatus is the state of one repository's hook for one event.
type AutoPublishHookStatus string

const (
	// AutoPublishHookAbsent: no hook file exists; installing is offered.
	AutoPublishHookAbsent AutoPublishHookStatus = "absent"
	// AutoPublishHookInstalled: Peasant's own hook is installed.
	AutoPublishHookInstalled AutoPublishHookStatus = "installed"
	// AutoPublishHookBlocked: Peasant will not manage the hook, because a file
	// it did not write is in the way or the hook path is shared outside the
	// repository. Peasant never overwrites such a file. It carries a remedy.
	AutoPublishHookBlocked AutoPublishHookStatus = "blocked"
	// AutoPublishHookFailed: installing the hook failed. Only an install
	// response reports it, and it carries a remedy naming the failure. A rule
	// or a removal response never reports it.
	AutoPublishHookFailed AutoPublishHookStatus = "failed"
)

// AllAutoPublishHookStatuses is the canonical hook state menu.
var AllAutoPublishHookStatuses = []AutoPublishHookStatus{AutoPublishHookAbsent, AutoPublishHookInstalled, AutoPublishHookBlocked, AutoPublishHookFailed}

func (s AutoPublishHookStatus) IsValid() bool  { return inSet(s, AllAutoPublishHookStatuses) }
func (s AutoPublishHookStatus) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (AutoPublishHookStatus) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Auto Publish Hook Status", "State of one repository's auto-publish hook for one event", AllAutoPublishHookStatuses), nil
}

// AutoPublishHookRemedy tells the user what to do about a blocked or failed
// hook. Snippet, present only for a blocked hook Peasant could otherwise
// manage, is the exact section of hook script to add to the existing hook by
// hand.
type AutoPublishHookRemedy struct {
	Message string `json:"message" minLength:"1"`
	Snippet string `json:"snippet,omitempty"`
}

// AutoPublishHook is the state of one repository's hook for one event. Only a
// blocked or failed hook carries a remedy.
type AutoPublishHook struct {
	Event  AutoPublishEvent       `json:"event"`
	Status AutoPublishHookStatus  `json:"status"`
	Remedy *AutoPublishHookRemedy `json:"remedy,omitempty" nullable:"false"`
}

// Validate checks the event, the status, and the remedy pairing.
func (h AutoPublishHook) Validate() error {
	if !h.Event.IsValid() {
		return fmt.Errorf("auto-publish hook validation failed at schema.AutoPublishHook.Validate: event %q is outside the closed set; emit a member of schema.AllAutoPublishEvents", h.Event)
	}
	if !h.Status.IsValid() {
		return fmt.Errorf("auto-publish hook validation failed at schema.AutoPublishHook.Validate: status %q is outside the closed set; emit a member of schema.AllAutoPublishHookStatuses", h.Status)
	}
	needsRemedy := h.Status == AutoPublishHookBlocked || h.Status == AutoPublishHookFailed
	if needsRemedy != (h.Remedy != nil) {
		return fmt.Errorf("auto-publish hook validation failed for %s at schema.AutoPublishHook.Validate: status %q has the wrong remedy presence; a blocked or failed hook says what to do and no other status carries a remedy", h.Event, h.Status)
	}
	if h.Remedy == nil {
		return nil
	}
	if strings.TrimSpace(h.Remedy.Message) == "" {
		return fmt.Errorf("auto-publish hook validation failed for %s at schema.AutoPublishHook.Validate: the remedy has no message; the user cannot act on the hook; say what to do", h.Event)
	}
	if h.Remedy.Snippet != "" && h.Status != AutoPublishHookBlocked {
		return fmt.Errorf("auto-publish hook validation failed for %s at schema.AutoPublishHook.Validate: a %s hook carries a snippet to paste; only a blocked hook has a section to add by hand", h.Event, h.Status)
	}
	return nil
}

// AutoPublishRepository is one recorded repository a rule matches, with the
// state of its hook for each of the rule's events.
type AutoPublishRepository struct {
	// Path is the repository's root on this computer.
	Path string `json:"path" minLength:"1"`
	// Label is the repository's schema.RemoteLabel, present when it has a
	// remote.
	Label string            `json:"label,omitempty"`
	Hooks []AutoPublishHook `json:"hooks" nullable:"false"`
}

// Validate checks the repository and that each event appears once.
func (r AutoPublishRepository) Validate() error {
	if strings.TrimSpace(r.Path) == "" || r.Hooks == nil {
		return fmt.Errorf("auto-publish repository validation failed at schema.AutoPublishRepository.Validate: path is empty or hooks is null; emit the repository root and [] when the rule names no event")
	}
	events := make(map[AutoPublishEvent]struct{}, len(r.Hooks))
	for _, hook := range r.Hooks {
		if err := hook.Validate(); err != nil {
			return fmt.Errorf("auto-publish repository validation failed for %q: %w", r.Path, err)
		}
		if _, duplicate := events[hook.Event]; duplicate {
			return fmt.Errorf("auto-publish repository validation failed for %q at schema.AutoPublishRepository.Validate: event %q has two hooks; report each event once", r.Path, hook.Event)
		}
		events[hook.Event] = struct{}{}
	}
	return nil
}

// AutoPublishRuleKind says what a rule's match pattern names.
type AutoPublishRuleKind string

const (
	// AutoPublishRuleFolder: Match is a folder glob on this computer.
	AutoPublishRuleFolder AutoPublishRuleKind = "folder"
	// AutoPublishRuleRemote: Match is a git remote pattern.
	AutoPublishRuleRemote AutoPublishRuleKind = "remote"
)

// AllAutoPublishRuleKinds is the canonical rule kind menu.
var AllAutoPublishRuleKinds = []AutoPublishRuleKind{AutoPublishRuleFolder, AutoPublishRuleRemote}

func (k AutoPublishRuleKind) IsValid() bool  { return inSet(k, AllAutoPublishRuleKinds) }
func (k AutoPublishRuleKind) String() string { return string(k) }

// JSONSchema implements jsonschema.Exposer.
func (AutoPublishRuleKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Auto Publish Rule Kind", "What an auto-publish rule's match pattern names", AllAutoPublishRuleKinds), nil
}

// AutoPublishRuleRequest is the body of PUT /api/v1/settings/auto-publish/{id}.
// Kind says whether Match is a folder glob or a git remote pattern. An empty
// Events list keeps the rule and publishes nothing.
type AutoPublishRuleRequest struct {
	Kind        AutoPublishRuleKind `json:"kind"`
	Match       string              `json:"match" minLength:"1"`
	Events      []AutoPublishEvent  `json:"events" nullable:"false"`
	Collectives []VillageUUID       `json:"collectives" nullable:"false"`
}

// Validate checks the kind, the match, and that no event or collective
// repeats.
func (r AutoPublishRuleRequest) Validate() error {
	return validateAutoPublishRule(r.Kind, r.Match, r.Events, r.Collectives)
}

// AutoPublishRule is one saved auto-publish rule with the recorded
// repositories it matches. Saving a rule installs nothing: each repository
// reports its hooks as they are, and installing is a separate call per
// repository.
type AutoPublishRule struct {
	ID           string                  `json:"id" minLength:"1"`
	Kind         AutoPublishRuleKind     `json:"kind"`
	Match        string                  `json:"match" minLength:"1"`
	Events       []AutoPublishEvent      `json:"events" nullable:"false"`
	Collectives  []VillageUUID           `json:"collectives" nullable:"false"`
	Repositories []AutoPublishRepository `json:"repositories" nullable:"false"`
}

// Validate checks the rule, each repository, and that every repository
// reports exactly the rule's events.
func (r AutoPublishRule) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("auto-publish rule validation failed at schema.AutoPublishRule.Validate: id is empty; the settings page cannot address the rule; emit its identifier")
	}
	if err := validateAutoPublishRule(r.Kind, r.Match, r.Events, r.Collectives); err != nil {
		return err
	}
	if r.Repositories == nil {
		return fmt.Errorf("auto-publish rule validation failed for %q at schema.AutoPublishRule.Validate: repositories is null; emit [] when no recorded repository matches", r.ID)
	}
	paths := make(map[string]struct{}, len(r.Repositories))
	for _, repository := range r.Repositories {
		if err := repository.Validate(); err != nil {
			return err
		}
		if _, duplicate := paths[repository.Path]; duplicate {
			return fmt.Errorf("auto-publish rule validation failed for %q at schema.AutoPublishRule.Validate: repository %q is listed twice; list each repository once", r.ID, repository.Path)
		}
		paths[repository.Path] = struct{}{}
		if len(repository.Hooks) != len(r.Events) {
			return fmt.Errorf("auto-publish rule validation failed for %q at schema.AutoPublishRule.Validate: repository %q reports %d hooks for %d events; report one hook per rule event", r.ID, repository.Path, len(repository.Hooks), len(r.Events))
		}
		for _, hook := range repository.Hooks {
			if !inSet(hook.Event, r.Events) {
				return fmt.Errorf("auto-publish rule validation failed for %q at schema.AutoPublishRule.Validate: repository %q reports event %q, which the rule does not name", r.ID, repository.Path, hook.Event)
			}
			if hook.Status == AutoPublishHookFailed {
				return fmt.Errorf("auto-publish rule validation failed for %q at schema.AutoPublishRule.Validate: repository %q reports a failed hook; failed is an install outcome; report the hook as it is now", r.ID, repository.Path)
			}
		}
	}
	return nil
}

func validateAutoPublishRule(kind AutoPublishRuleKind, match string, events []AutoPublishEvent, collectives []VillageUUID) error {
	if !kind.IsValid() {
		return fmt.Errorf("auto-publish rule validation failed for %q: kind %q is outside the closed set; say whether the match names a folder or a remote with a member of schema.AllAutoPublishRuleKinds", match, kind)
	}
	if strings.TrimSpace(match) == "" {
		return fmt.Errorf("auto-publish rule validation failed: match is empty; the rule would cover nothing; name a folder or remote pattern")
	}
	if events == nil || collectives == nil {
		return fmt.Errorf("auto-publish rule validation failed for %q: events or collectives is null; emit [] for an empty list", match)
	}
	seenEvents := make(map[AutoPublishEvent]struct{}, len(events))
	for _, event := range events {
		if !event.IsValid() {
			return fmt.Errorf("auto-publish rule validation failed for %q: event %q is outside the closed set; use a member of schema.AllAutoPublishEvents", match, event)
		}
		if _, duplicate := seenEvents[event]; duplicate {
			return fmt.Errorf("auto-publish rule validation failed for %q: event %q is listed twice; list each event once", match, event)
		}
		seenEvents[event] = struct{}{}
	}
	seenCollectives := make(map[VillageUUID]struct{}, len(collectives))
	for _, collective := range collectives {
		if _, duplicate := seenCollectives[collective]; duplicate {
			return fmt.Errorf("auto-publish rule validation failed for %q: collective %q is listed twice; list each collective once", match, collective)
		}
		seenCollectives[collective] = struct{}{}
	}
	return nil
}

// AutoPublishInstallRequest is the body of POST
// /api/v1/settings/auto-publish/{id}/install: the one recorded repository to
// install the rule's hooks in.
type AutoPublishInstallRequest struct {
	Path string `json:"path" minLength:"1"`
}

// Validate checks that the request names a repository.
func (r AutoPublishInstallRequest) Validate() error {
	if strings.TrimSpace(r.Path) == "" {
		return fmt.Errorf("auto-publish install validation failed at schema.AutoPublishInstallRequest.Validate: path is empty; name the recorded repository root to install in")
	}
	return nil
}

// AutoPublishRemovalResponse is the response of DELETE
// /api/v1/settings/auto-publish/{id}: the removed rule's identifier and the
// hooks of the repositories it matched, as they are. Removing a rule keeps
// hook files but stops their rule-required uploads unless another active rule
// covers the repository and event. Paused rules grant no publishing consent.
// Hooks installed separately from the terminal keep their independent consent.
type AutoPublishRemovalResponse struct {
	ID           string                  `json:"id" minLength:"1"`
	Repositories []AutoPublishRepository `json:"repositories" nullable:"false"`
}

// Validate checks every repository and that each appears once.
func (r AutoPublishRemovalResponse) Validate() error {
	if strings.TrimSpace(r.ID) == "" || r.Repositories == nil {
		return fmt.Errorf("auto-publish removal validation failed at schema.AutoPublishRemovalResponse.Validate: id is empty or repositories is null; emit the removed rule's identifier and [] when it matched no repository")
	}
	paths := make(map[string]struct{}, len(r.Repositories))
	for _, repository := range r.Repositories {
		if err := repository.Validate(); err != nil {
			return err
		}
		if _, duplicate := paths[repository.Path]; duplicate {
			return fmt.Errorf("auto-publish removal validation failed for %q at schema.AutoPublishRemovalResponse.Validate: repository %q is listed twice; list each repository once", r.ID, repository.Path)
		}
		paths[repository.Path] = struct{}{}
		for _, hook := range repository.Hooks {
			if hook.Status == AutoPublishHookFailed {
				return fmt.Errorf("auto-publish removal validation failed for %q at schema.AutoPublishRemovalResponse.Validate: repository %q reports a failed hook; removing a rule changes no hook; report the hook as it is", r.ID, repository.Path)
			}
		}
	}
	return nil
}
