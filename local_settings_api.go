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
// key set and each key's description; this contract owns the value shapes.
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
)

// AllLocalSettingKinds is the canonical setting value kind menu.
var AllLocalSettingKinds = []LocalSettingKind{LocalSettingBoolean, LocalSettingInteger, LocalSettingString, LocalSettingStringList, LocalSettingChoice}

func (k LocalSettingKind) IsValid() bool  { return inSet(k, AllLocalSettingKinds) }
func (k LocalSettingKind) String() string { return string(k) }

// JSONSchema implements jsonschema.Exposer.
func (LocalSettingKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Local Setting Kind", "JSON type of one local setting's value", AllLocalSettingKinds), nil
}

// LocalSettingValue is one setting's JSON value, held as its exact JSON text.
// The setting's kind chooses which JSON type it is.
type LocalSettingValue []byte

// MarshalJSON emits the held JSON text.
func (v LocalSettingValue) MarshalJSON() ([]byte, error) {
	if len(v) == 0 {
		return []byte("null"), nil
	}
	return append([]byte(nil), v...), nil
}

// UnmarshalJSON keeps the JSON text as received.
func (v *LocalSettingValue) UnmarshalJSON(data []byte) error {
	*v = append((*v)[:0], data...)
	return nil
}

// JSONSchema implements jsonschema.Exposer.
func (LocalSettingValue) JSONSchema() (jsonschema.Schema, error) {
	boolean, integer, text, list := jsonschema.Schema{}, jsonschema.Schema{}, jsonschema.Schema{}, jsonschema.Schema{}
	boolean.AddType(jsonschema.Boolean)
	integer.AddType(jsonschema.Integer)
	text.AddType(jsonschema.String)
	item := jsonschema.Schema{}
	item.AddType(jsonschema.String)
	list.AddType(jsonschema.Array)
	list.WithItems(*(&jsonschema.Items{}).WithSchemaOrBool(item.ToSchemaOrBool()))
	s := jsonschema.Schema{}
	s.WithTitle("Local Setting Value")
	s.WithDescription("One setting's value: a boolean, an integer, a string, or an array of strings, as its kind says")
	s.WithAnyOf(boolean.ToSchemaOrBool(), integer.ToSchemaOrBool(), text.ToSchemaOrBool(), list.ToSchemaOrBool())
	return s, nil
}

// validateFor checks that the value has the JSON type kind names.
func (v LocalSettingValue) validateFor(kind LocalSettingKind, options []string) error {
	decoder := json.NewDecoder(bytes.NewReader(v))
	decoder.UseNumber()
	var value any
	if len(bytes.TrimSpace(v)) == 0 || decoder.Decode(&value) != nil || value == nil {
		return fmt.Errorf("the value is missing or null")
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
		if text, ok := value.(string); ok && inSet(text, options) {
			return nil
		}
	}
	return fmt.Errorf("the value is not a %s", kind)
}

// LocalSetting is one Peasant setting with its metadata.
type LocalSetting struct {
	// Key is the setting's dotted configuration path, for example
	// push.concurrency.
	Key   string            `json:"key" minLength:"1"`
	Kind  LocalSettingKind  `json:"kind"`
	Value LocalSettingValue `json:"value"`
	// Options is the menu of a choice setting. Other kinds have none.
	Options []string `json:"options,omitempty" nullable:"false"`
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
		if _, duplicate := seen[option]; duplicate || option == "" {
			return fmt.Errorf("setting validation failed for %q at schema.LocalSetting.Validate: option %q is empty or repeated; list each option once", s.Key, option)
		}
		seen[option] = struct{}{}
	}
	if err := s.Value.validateFor(s.Kind, s.Options); err != nil {
		return fmt.Errorf("setting validation failed for %q at schema.LocalSetting.Validate: %v; emit a value of kind %q", s.Key, err, s.Kind)
	}
	return nil
}

// LocalSettingUpdateRequest is the body of PATCH /api/v1/settings. One request
// changes one key.
type LocalSettingUpdateRequest struct {
	Key   string            `json:"key" minLength:"1"`
	Value LocalSettingValue `json:"value"`
}

// Validate checks that the request names a key and carries a value. The
// server checks the value against the key's kind.
func (r LocalSettingUpdateRequest) Validate() error {
	if strings.TrimSpace(r.Key) == "" {
		return fmt.Errorf("setting update validation failed at schema.LocalSettingUpdateRequest.Validate: key is empty; nothing would change; name the setting's dotted configuration path")
	}
	trimmed := bytes.TrimSpace(r.Value)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return fmt.Errorf("setting update validation failed for %q at schema.LocalSettingUpdateRequest.Validate: value is missing or null; one request sets one value; send the new value", r.Key)
	}
	return nil
}

// --- Auto-publish ---

// AutoPublishEvent is the git hook that publishes a binding's sessions.
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
	return closedStringEnumSchema("Auto Publish Event", "Git hook that publishes an auto-publish binding's sessions", AllAutoPublishEvents), nil
}

// AutoPublishHookStatus is the state of a binding's git hook.
type AutoPublishHookStatus string

const (
	// AutoPublishHookActive: Peasant's hook is installed and publishes.
	AutoPublishHookActive AutoPublishHookStatus = "active"
	// AutoPublishHookBlocked: a hook Peasant does not manage is in the way, and
	// Peasant never overwrites one. The hook carries a remedy.
	AutoPublishHookBlocked AutoPublishHookStatus = "blocked"
	// AutoPublishHookOff: no hook publishes for this binding.
	AutoPublishHookOff AutoPublishHookStatus = "off"
)

// AllAutoPublishHookStatuses is the canonical hook state menu.
var AllAutoPublishHookStatuses = []AutoPublishHookStatus{AutoPublishHookActive, AutoPublishHookBlocked, AutoPublishHookOff}

func (s AutoPublishHookStatus) IsValid() bool  { return inSet(s, AllAutoPublishHookStatuses) }
func (s AutoPublishHookStatus) String() string { return string(s) }

// JSONSchema implements jsonschema.Exposer.
func (AutoPublishHookStatus) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Auto Publish Hook Status", "State of an auto-publish binding's git hook", AllAutoPublishHookStatuses), nil
}

// AutoPublishHookRemedy tells the user how to unblock a hook. Command, when
// present, is the exact line to add to the existing hook.
type AutoPublishHookRemedy struct {
	Message string `json:"message" minLength:"1"`
	Command string `json:"command,omitempty"`
}

// AutoPublishHook is the state of a binding's git hook. Only a blocked hook
// carries a remedy.
type AutoPublishHook struct {
	Status AutoPublishHookStatus  `json:"status"`
	Remedy *AutoPublishHookRemedy `json:"remedy,omitempty" nullable:"false"`
}

// Validate checks the status and the remedy pairing.
func (h AutoPublishHook) Validate() error {
	if !h.Status.IsValid() {
		return fmt.Errorf("auto-publish hook validation failed at schema.AutoPublishHook.Validate: status %q is outside the closed set; emit a member of schema.AllAutoPublishHookStatuses", h.Status)
	}
	if (h.Status == AutoPublishHookBlocked) != (h.Remedy != nil) {
		return fmt.Errorf("auto-publish hook validation failed at schema.AutoPublishHook.Validate: status %q has the wrong remedy presence; a blocked hook says how to unblock it and no other status carries a remedy", h.Status)
	}
	if h.Remedy != nil && strings.TrimSpace(h.Remedy.Message) == "" {
		return fmt.Errorf("auto-publish hook validation failed at schema.AutoPublishHook.Validate: the remedy has no message; the user cannot unblock the hook; say what to do")
	}
	return nil
}

// AutoPublishBindingRequest is the body of PUT
// /api/v1/settings/auto-publish/{id}. Match is a folder pattern or a
// repository label pattern. An empty Events list keeps the binding and
// publishes nothing.
type AutoPublishBindingRequest struct {
	Match       string             `json:"match" minLength:"1"`
	Events      []AutoPublishEvent `json:"events" nullable:"false"`
	Collectives []VillageUUID      `json:"collectives" nullable:"false"`
}

// Validate checks the match and that no event or collective repeats.
func (r AutoPublishBindingRequest) Validate() error {
	return validateAutoPublishBinding(r.Match, r.Events, r.Collectives)
}

// AutoPublishBinding is one saved auto-publish binding with its hook state.
type AutoPublishBinding struct {
	ID          string             `json:"id" minLength:"1"`
	Match       string             `json:"match" minLength:"1"`
	Events      []AutoPublishEvent `json:"events" nullable:"false"`
	Collectives []VillageUUID      `json:"collectives" nullable:"false"`
	Hook        AutoPublishHook    `json:"hook"`
}

// Validate checks the binding and its hook.
func (b AutoPublishBinding) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return fmt.Errorf("auto-publish binding validation failed at schema.AutoPublishBinding.Validate: id is empty; the settings page cannot address the binding; emit its identifier")
	}
	if err := validateAutoPublishBinding(b.Match, b.Events, b.Collectives); err != nil {
		return err
	}
	if len(b.Events) == 0 && b.Hook.Status == AutoPublishHookActive {
		return fmt.Errorf("auto-publish binding validation failed for %q at schema.AutoPublishBinding.Validate: a binding with no events reports an active hook; nothing publishes for it; report the hook off", b.ID)
	}
	return b.Hook.Validate()
}

func validateAutoPublishBinding(match string, events []AutoPublishEvent, collectives []VillageUUID) error {
	if strings.TrimSpace(match) == "" {
		return fmt.Errorf("auto-publish binding validation failed at schema.AutoPublishBinding: match is empty; the binding would cover nothing; name a folder or repository pattern")
	}
	if events == nil || collectives == nil {
		return fmt.Errorf("auto-publish binding validation failed for %q: events or collectives is null; emit [] for an empty list", match)
	}
	seenEvents := make(map[AutoPublishEvent]struct{}, len(events))
	for _, event := range events {
		if !event.IsValid() {
			return fmt.Errorf("auto-publish binding validation failed for %q: event %q is outside the closed set; use a member of schema.AllAutoPublishEvents", match, event)
		}
		if _, duplicate := seenEvents[event]; duplicate {
			return fmt.Errorf("auto-publish binding validation failed for %q: event %q is listed twice; list each event once", match, event)
		}
		seenEvents[event] = struct{}{}
	}
	seenCollectives := make(map[VillageUUID]struct{}, len(collectives))
	for _, collective := range collectives {
		if _, duplicate := seenCollectives[collective]; duplicate {
			return fmt.Errorf("auto-publish binding validation failed for %q: collective %q is listed twice; list each collective once", match, collective)
		}
		seenCollectives[collective] = struct{}{}
	}
	return nil
}

// AutoPublishRemovalResponse is the response of DELETE
// /api/v1/settings/auto-publish/{id}: the removed binding's identifier and its
// hook state after removal.
type AutoPublishRemovalResponse struct {
	ID   string          `json:"id" minLength:"1"`
	Hook AutoPublishHook `json:"hook"`
}

// Validate checks that a removed binding leaves no active hook.
func (r AutoPublishRemovalResponse) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("auto-publish removal validation failed at schema.AutoPublishRemovalResponse.Validate: id is empty; emit the removed binding's identifier")
	}
	if r.Hook.Status == AutoPublishHookActive {
		return fmt.Errorf("auto-publish removal validation failed for %q at schema.AutoPublishRemovalResponse.Validate: the hook is still active after removal; report it off, or blocked with a remedy when a line must be removed by hand", r.ID)
	}
	return r.Hook.Validate()
}

// LocalSettingsResponse is the response of GET /api/v1/settings: every
// setting the settings page edits and every auto-publish binding.
type LocalSettingsResponse struct {
	Settings    []LocalSetting       `json:"settings" nullable:"false"`
	AutoPublish []AutoPublishBinding `json:"autoPublish" nullable:"false"`
}

// Validate checks every setting and binding and that keys and identifiers do
// not repeat.
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
	for _, binding := range r.AutoPublish {
		if err := binding.Validate(); err != nil {
			return err
		}
		if _, duplicate := ids[binding.ID]; duplicate {
			return fmt.Errorf("settings validation failed at schema.LocalSettingsResponse.Validate: binding %q is listed twice; list each binding once", binding.ID)
		}
		ids[binding.ID] = struct{}{}
	}
	return nil
}
