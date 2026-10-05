package schema

import (
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/swaggest/jsonschema-go"
)

// ThinkingLevel is the canonical reasoning-effort level recorded for assistant
// output. The set is closed. Absence means unknown; "off" is the single
// disabled state; "ultra" is above "max", and "ultracode" is a distinct level. Native harness
// spellings are mapping inputs, never wire values: a native spelling that
// differs from the canonical value travels as ThinkingLevelRaw.
type ThinkingLevel string

const (
	ThinkingLevelOff       ThinkingLevel = "off"
	ThinkingLevelMinimal   ThinkingLevel = "minimal"
	ThinkingLevelLow       ThinkingLevel = "low"
	ThinkingLevelMedium    ThinkingLevel = "medium"
	ThinkingLevelHigh      ThinkingLevel = "high"
	ThinkingLevelXHigh     ThinkingLevel = "xhigh"
	ThinkingLevelMax       ThinkingLevel = "max"
	ThinkingLevelUltra     ThinkingLevel = "ultra"
	ThinkingLevelUltracode ThinkingLevel = "ultracode"
)

// AllThinkingLevels lists the nine levels in canonical order, off..ultracode.
var AllThinkingLevels = []ThinkingLevel{
	ThinkingLevelOff,
	ThinkingLevelMinimal,
	ThinkingLevelLow,
	ThinkingLevelMedium,
	ThinkingLevelHigh,
	ThinkingLevelXHigh,
	ThinkingLevelMax,
	ThinkingLevelUltra,
	ThinkingLevelUltracode,
}

// IsValid reports whether v is absent ("") or a member of AllThinkingLevels.
func (v ThinkingLevel) IsValid() bool {
	return v == "" || slices.Contains(AllThinkingLevels, v)
}

// String returns the canonical level spelling.
func (v ThinkingLevel) String() string { return string(v) }

// JSONSchema implements jsonschema.Exposer.
func (ThinkingLevel) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Thinking Level",
		"Canonical reasoning-effort level of assistant-generated output: off, minimal, low, medium, high, xhigh, max, ultra, ultracode. off is the single disabled state, ultra is above max, and ultracode is a distinct level. Omit the field when the level is unknown; never infer it from token budgets.",
		AllThinkingLevels,
	), nil
}

// ThinkingLevelRaw is the exact native level spelling observed when it differs
// from the emitted canonical value (including when none could be mapped).
// Source evidence, never a canonical value; non-level evidence (e.g. a numeric
// budget) is never raw. Non-empty, valid UTF-8, <=128 encoded bytes, no edge
// Unicode White_Space.
type ThinkingLevelRaw string

// ThinkingLevelRawMaxBytes is the encoded UTF-8 byte bound of ThinkingLevelRaw.
const ThinkingLevelRawMaxBytes = 128

// ThinkingLevelRawUTF8ByteFormat identifies the JSON Schema format assertion
// that counts encoded UTF-8 bytes of a ThinkingLevelRaw. maxLength counts code
// points, so this assertion carries the byte rule.
const ThinkingLevelRawUTF8ByteFormat = "thinking-level-raw-utf8-128-bytes"

// NewThinkingLevelRaw validates and constructs a ThinkingLevelRaw. Accepted
// bytes are preserved exactly; nothing is trimmed or normalized.
func NewThinkingLevelRaw(raw string) (ThinkingLevelRaw, error) {
	const where = "thinking level raw validation failed at schema.NewThinkingLevelRaw while constructing thinkingLevelRaw source evidence"
	if raw == "" {
		return "", fmt.Errorf("%s: the value is empty, so no native spelling can be represented; omit thinkingLevelRaw when the native spelling equals the canonical level or nothing was observed", where)
	}
	if !utf8.ValidString(raw) {
		return "", fmt.Errorf("%s: value %q is not valid UTF-8, so it cannot be emitted without changing source bytes; drop thinkingLevelRaw for this observation (keep any canonical level) or supply the original spelling as valid UTF-8", where, raw)
	}
	if len(raw) > ThinkingLevelRawMaxBytes {
		return "", fmt.Errorf("%s: value is %d encoded UTF-8 bytes, over the %d-byte bound, and truncation would fabricate evidence; drop thinkingLevelRaw for this observation (keep any canonical level) rather than truncating", where, len(raw), ThinkingLevelRawMaxBytes)
	}
	first, _ := utf8.DecodeRuneInString(raw)
	last, _ := utf8.DecodeLastRuneInString(raw)
	if observedModelEdgeWhitespace(first) || observedModelEdgeWhitespace(last) {
		return "", fmt.Errorf("%s: value %q has Unicode whitespace at an edge, but the wire preserves exact unpadded native spellings; drop thinkingLevelRaw for this observation or remove only the edge whitespace at the producing boundary", where, raw)
	}
	return ThinkingLevelRaw(raw), nil
}

// IsValid reports whether v satisfies the raw rule. The empty string is not a
// valid raw value; omit the field instead.
func (v ThinkingLevelRaw) IsValid() bool {
	_, err := NewThinkingLevelRaw(string(v))
	return err == nil
}

// Validate returns the actionable constructor error for an invalid raw value.
func (v ThinkingLevelRaw) Validate() error {
	_, err := NewThinkingLevelRaw(string(v))
	return err
}

// String returns the exact native spelling.
func (v ThinkingLevelRaw) String() string { return string(v) }

// JSONSchema implements jsonschema.Exposer.
func (ThinkingLevelRaw) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Thinking Level Raw")
	s.WithDescription("Exact native thinking-level spelling observed when it differs from the emitted canonical thinkingLevel, including when no canonical level could be mapped. Source evidence, never a canonical value; numeric budgets are never raw. Non-empty, valid UTF-8, at most 128 encoded UTF-8 bytes, and no Unicode White_Space code point at either edge. Omit when nothing was observed or the native spelling is canonical.")
	s.WithMinLength(1)
	s.WithMaxLength(ThinkingLevelRawMaxBytes)
	s.WithFormat(ThinkingLevelRawUTF8ByteFormat)
	s.WithPattern(observedModelPattern)
	s.WithExamples("none", "custom")
	return s, nil
}

// ValidateThinkingLevelRawJSONSchemaFormat implements
// ThinkingLevelRawUTF8ByteFormat for JSON Schema runtimes. Non-string values
// pass so type assertions remain the responsibility of the "type" keyword.
func ValidateThinkingLevelRawJSONSchemaFormat(value any) bool {
	raw, ok := value.(string)
	if !ok {
		return true
	}
	return raw != "" && utf8.ValidString(raw) && len(raw) <= ThinkingLevelRawMaxBytes
}

// ValidateThinkingLevelEvidence enforces the producer rule that a thinking
// level or raw spelling may appear only on assistant-generated output (root
// assistants and subagents both use RoleAssistant). Generated shape validators
// cannot express this role condition. Omitted evidence is valid.
func ValidateThinkingLevelEvidence(role Role, level ThinkingLevel, raw ThinkingLevelRaw) error {
	if level == "" && raw == "" {
		return nil
	}
	if !level.IsValid() {
		return fmt.Errorf("thinking level evidence validation failed at schema.ValidateThinkingLevelEvidence: thinkingLevel %q is outside the canonical set %v, so consumers cannot compare it; map the native spelling to a canonical level, or omit thinkingLevel and carry the spelling in thinkingLevelRaw", level, AllThinkingLevels)
	}
	if raw != "" {
		if err := raw.Validate(); err != nil {
			return fmt.Errorf("thinking level evidence validation failed at schema.ValidateThinkingLevelEvidence: thinkingLevelRaw is invalid and must not be emitted; drop the raw spelling for this observation: %w", err)
		}
	}
	if role != RoleAssistant {
		return fmt.Errorf("thinking level evidence validation failed at schema.ValidateThinkingLevelEvidence: role %q is not assistant, so thinkingLevel/thinkingLevelRaw cannot describe assistant or subagent output and the payload must be rejected; omit both fields from user, system, and tool turns", role)
	}
	return nil
}

// validateThinkingLevelPair validates a session-level or metadata pair without
// the role rule. where names the calling boundary for actionable errors.
func validateThinkingLevelPair(level ThinkingLevel, raw ThinkingLevelRaw, where string) error {
	if !level.IsValid() {
		return fmt.Errorf("thinking level validation failed at %s: thinkingLevel %q is outside the canonical set %v, so consumers cannot compare it; use a canonical level, or omit thinkingLevel and carry the native spelling in thinkingLevelRaw", where, level, AllThinkingLevels)
	}
	if raw != "" {
		if err := raw.Validate(); err != nil {
			return fmt.Errorf("thinking level validation failed at %s: thinkingLevelRaw is invalid; omit it or supply the exact bounded native spelling: %w", where, err)
		}
	}
	return nil
}
