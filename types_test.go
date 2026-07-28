package schema_test

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

// --- NewSessionID ---

//go:embed testdata/contract/session_ids.yaml
var sessionIDCasesYAML []byte

func loadSessionIDCases(t *testing.T) testcase.Corpus[string, bool] {
	t.Helper()
	corpus, err := testcase.LoadCorpus[string, bool](sessionIDCasesYAML)
	if err != nil {
		t.Fatalf("load session ID corpus: %v", err)
	}
	if got := len(corpus.Cases); got != 5 {
		t.Fatalf("session ID corpus has %d cases, want exactly 5", got)
	}
	caseassert.RequireValid(t, corpus)
	return corpus
}

func TestNewSessionID_StrikeGrammar(t *testing.T) {
	corpus := loadSessionIDCases(t)
	jsonSchema, err := schema.SessionID("").JSONSchema()
	if err != nil {
		t.Fatalf("SessionID.JSONSchema(): %v", err)
	}
	if jsonSchema.Pattern == nil {
		t.Fatal("SessionID.JSONSchema() returned no pattern")
	}
	pattern, err := regexp.Compile(*jsonSchema.Pattern)
	if err != nil {
		t.Fatalf("compile SessionID JSON Schema pattern: %v", err)
	}

	required := map[string]bool{
		"strike_timestamped":  false,
		"strike_bare":         false,
		"strike_lowercase":    false,
		"strike_wrong_length": false,
		"strike_path_unsafe":  false,
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Expected != (c.Classification == testcase.MustPass) {
				t.Fatalf("fixture expected=%v disagrees with classification=%q", c.Expected, c.Classification)
			}
			_, constructorErr := schema.NewSessionID(c.Input)
			if got := constructorErr == nil; got != c.Expected {
				t.Errorf("NewSessionID(%q) accepted=%v, want %v (err=%v)", c.Input, got, c.Expected, constructorErr)
			}
			if got := pattern.MatchString(c.Input); got != c.Expected {
				t.Errorf("SessionID JSON Schema pattern matched %q = %v, want %v", c.Input, got, c.Expected)
			}
		})
		if _, ok := required[c.Name]; ok {
			required[c.Name] = true
		}
	}
	for name, present := range required {
		if !present {
			t.Errorf("session ID corpus is missing required case %q", name)
		}
	}
}

func TestNewSessionID_ValidUUID(t *testing.T) {
	sid, err := schema.NewSessionID("99d59925-36bc-424c-a789-8be54d9702ba")
	if err != nil {
		t.Fatalf("expected valid UUID to succeed: %v", err)
	}
	if sid.String() != "99d59925-36bc-424c-a789-8be54d9702ba" {
		t.Errorf("unexpected value: %s", sid)
	}
}

func TestNewSessionID_ValidAgentHex(t *testing.T) {
	sid, err := schema.NewSessionID("agent-a3aee4f")
	if err != nil {
		t.Fatalf("expected agent-{hex} to succeed: %v", err)
	}
	if sid.String() != "agent-a3aee4f" {
		t.Errorf("unexpected value: %s", sid)
	}
}

func TestNewSessionID_ValidSesFormat(t *testing.T) {
	sid, err := schema.NewSessionID("ses_3cd91f52effeXd3QAJ54")
	if err != nil {
		t.Fatalf("expected ses_{id} to succeed: %v", err)
	}
	_ = sid
}

func TestNewSessionID_ValidSessFormat_ACP(t *testing.T) {
	// ACP uses double-s sess_ prefix
	sid, err := schema.NewSessionID("sess_3cd91f52effeXd3QAJ54")
	if err != nil {
		t.Fatalf("expected sess_{id} (ACP format) to succeed: %v", err)
	}
	_ = sid
}

func TestNewSessionID_ValidMsgFormat(t *testing.T) {
	sid, err := schema.NewSessionID("msg_001abc")
	if err != nil {
		t.Fatalf("expected msg_{id} to succeed: %v", err)
	}
	_ = sid
}

func TestNewSessionID_EmptyRejected(t *testing.T) {
	_, err := schema.NewSessionID("")
	if err == nil {
		t.Fatal("expected empty string to be rejected")
	}
}

func TestNewSessionID_PathSeparatorRejected(t *testing.T) {
	_, err := schema.NewSessionID("foo/bar")
	if err == nil {
		t.Fatal("expected path separator to be rejected")
	}
}

func TestNewSessionID_DotDotRejected(t *testing.T) {
	_, err := schema.NewSessionID("../etc/passwd")
	if err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}

func TestNewSessionID_ArbitraryStringRejected(t *testing.T) {
	_, err := schema.NewSessionID("not-a-valid-id")
	if err == nil {
		t.Fatal("expected arbitrary string to be rejected")
	}
}

// --- NewModelID ---

func TestNewModelID_Valid(t *testing.T) {
	m, err := schema.NewModelID("claude-opus-4-6")
	if err != nil {
		t.Fatalf("expected valid model ID to succeed: %v", err)
	}
	if m.String() != "claude-opus-4-6" {
		t.Errorf("unexpected value: %s", m)
	}
}

func TestNewModelID_EmptyRejected(t *testing.T) {
	_, err := schema.NewModelID("")
	if err == nil {
		t.Fatal("expected empty model ID to be rejected")
	}
}

// --- NewProjectHash ---

func TestNewProjectHash_ValidHex64(t *testing.T) {
	h := "a3aee4f0000000000000000000000000000000000000000000000000000abcde"
	if len(h) != 64 {
		t.Fatalf("test case is wrong length: %d", len(h))
	}
	ph, err := schema.NewProjectHash(h)
	if err != nil {
		t.Fatalf("expected valid hash to succeed: %v", err)
	}
	if ph.String() != h {
		t.Errorf("unexpected value: %s", ph)
	}
}

func TestNewProjectHash_TooShortRejected(t *testing.T) {
	_, err := schema.NewProjectHash("abc123")
	if err == nil {
		t.Fatal("expected short hash to be rejected")
	}
}

func TestNewProjectHash_UpperCaseRejected(t *testing.T) {
	h := "A3AEE4F0000000000000000000000000000000000000000000000000000ABCDE"
	_, err := schema.NewProjectHash(h)
	if err == nil {
		t.Fatal("expected uppercase hex to be rejected")
	}
}

// --- NewHostSlug ---

func TestNewHostSlug_Valid(t *testing.T) {
	hs, err := schema.NewHostSlug("github.com_acme-dev_peasant")
	if err != nil {
		t.Fatalf("expected valid host slug to succeed: %v", err)
	}
	if hs.String() != "github.com_acme-dev_peasant" {
		t.Errorf("unexpected value: %s", hs)
	}
}

func TestNewHostSlug_EmptyRejected(t *testing.T) {
	_, err := schema.NewHostSlug("")
	if err == nil {
		t.Fatal("expected empty host slug to be rejected")
	}
}

func TestNewHostSlug_DotDotRejected(t *testing.T) {
	_, err := schema.NewHostSlug("foo..bar")
	if err == nil {
		t.Fatal("expected host slug with '..' to be rejected")
	}
}

func TestNewHostSlug_InvalidCharRejected(t *testing.T) {
	_, err := schema.NewHostSlug("foo/bar")
	if err == nil {
		t.Fatal("expected host slug with '/' to be rejected")
	}
}

// --- SessionID.JSONSchema ---

func TestSessionID_JSONSchema_HasFormatAndPattern(t *testing.T) {
	s, err := schema.SessionID("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}

	if s.Format == nil {
		t.Fatal("JSONSchema: expected Format to be set (AC2: format:'session-id')")
	}
	if *s.Format != "session-id" {
		t.Errorf("JSONSchema: expected format 'session-id', got %q", *s.Format)
	}

	if s.Pattern == nil {
		t.Fatal("JSONSchema: expected Pattern to be set (AC2: regex pattern)")
	}
	if *s.Pattern == "" {
		t.Error("JSONSchema: expected non-empty pattern")
	}
}

// --- Enum type JSONSchema ---

func TestProvider_JSONSchema_HasEnum(t *testing.T) {
	// Harness is an alias for bestiary.Harness (external type) and cannot carry a
	// JSONSchema() Exposer method; the package-level HarnessJSONSchema() supplies it.
	s := schema.HarnessJSONSchema()
	if len(s.Enum) == 0 {
		t.Fatal("Provider JSONSchema: expected enum values")
	}
}

func TestVisibility_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.Visibility("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("Visibility JSONSchema: expected enum values")
	}
}

func TestToolCallKind_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.ToolCallKind("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("ToolCallKind JSONSchema: expected enum values")
	}
}

func TestStopReason_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.StopReason("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("StopReason JSONSchema: expected enum values")
	}
}

func TestRole_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.Role("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("Role JSONSchema: expected enum values")
	}
}

func TestEntryType_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.EntryType("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("EntryType JSONSchema: expected enum values")
	}
}

func TestSourceFormat_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.SourceFormat("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("SourceFormat JSONSchema: expected enum values")
	}
}

func TestSessionOutcome_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.SessionOutcome("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("SessionOutcome JSONSchema: expected enum values")
	}
}

// --- ModelID.JSONSchema ---

func TestModelID_JSONSchema_HasMinLength(t *testing.T) {
	s, err := schema.ModelID("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if s.MinLength == 0 {
		t.Fatal("ModelID JSONSchema: expected MinLength >= 1")
	}
}

// --- ProjectHash.JSONSchema ---

func TestProjectHash_JSONSchema_HasPattern(t *testing.T) {
	s, err := schema.ProjectHash("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if s.Pattern == nil || *s.Pattern == "" {
		t.Fatal("ProjectHash JSONSchema: expected Pattern to be set")
	}
}

// --- IsValid methods ---

func TestProvider_IsValid(t *testing.T) {
	// bestiary.Harness exposes IsKnown() (the post-migration replacement for the
	// old schema.Provider.IsValid()).
	for _, p := range schema.AllHarnesses {
		if !p.IsKnown() {
			t.Errorf("AllProviders[%s] should be valid", p)
		}
	}
	if schema.Harness("unknown").IsKnown() {
		t.Error("unknown provider should be invalid")
	}
}

func TestHarness_StrikeContract(t *testing.T) {
	if !schema.HarnessStrike.IsKnown() {
		t.Fatal("HarnessStrike is not recognized by the canonical Bestiary harness set")
	}
	if got := schema.HarnessDisplayName(schema.HarnessStrike); got != "Strike" {
		t.Errorf("HarnessDisplayName(HarnessStrike) = %q, want %q", got, "Strike")
	}
	count := 0
	for _, harness := range schema.AllHarnesses {
		if harness == schema.HarnessStrike {
			count++
		}
	}
	if count != 1 {
		t.Errorf("AllHarnesses contains HarnessStrike %d times, want exactly once", count)
	}
}

func TestRole_IsValid(t *testing.T) {
	for _, r := range schema.AllRoles {
		if !r.IsValid() {
			t.Errorf("AllRoles[%s] should be valid", r)
		}
	}
	if schema.Role("unknown").IsValid() {
		t.Error("unknown role should be invalid")
	}
}

func TestEntryType_IsValid(t *testing.T) {
	for _, e := range schema.AllEntryTypes {
		if !e.IsValid() {
			t.Errorf("AllEntryTypes[%s] should be valid", e)
		}
	}
	if schema.EntryType("unknown").IsValid() {
		t.Error("unknown entry type should be invalid")
	}
}

func TestToolCallKind_IsValid(t *testing.T) {
	for _, k := range schema.AllToolCallKinds {
		if !k.IsValid() {
			t.Errorf("AllToolCallKinds[%s] should be valid", k)
		}
	}
	if schema.ToolCallKind("unknown").IsValid() {
		t.Error("unknown tool call kind should be invalid")
	}
}

func TestStopReason_IsValid(t *testing.T) {
	for _, r := range schema.AllStopReasons {
		if !r.IsValid() {
			t.Errorf("AllStopReasons[%s] should be valid", r)
		}
	}
	if schema.StopReason("unknown").IsValid() {
		t.Error("unknown stop reason should be invalid")
	}
}

func TestVisibility_IsValid(t *testing.T) {
	for _, v := range schema.AllVisibilities {
		if !v.IsValid() {
			t.Errorf("AllVisibilities[%s] should be valid", v)
		}
	}
	if schema.Visibility("unknown").IsValid() {
		t.Error("unknown visibility should be invalid")
	}
}

// --- License ---

func TestLicense_IsValid(t *testing.T) {
	for _, l := range schema.AllLicenses {
		if !l.IsValid() {
			t.Errorf("AllLicenses contains %q but IsValid() = false", l)
		}
	}
	if schema.License("MIT").IsValid() {
		t.Error(`License("MIT").IsValid() = true, want false`)
	}
	if schema.License("").IsValid() {
		t.Error(`License("").IsValid() = true, want false`)
	}
}

// TestLicenseMenu pins the human-readable menu directly (de-circularizing the
// CLI/error-message tests that assert against LicenseMenu()'s own output): it
// catches a separator or ordering regression in the AllLicenses join.
func TestLicenseMenu(t *testing.T) {
	const want = "CC0-1.0, CC-BY-4.0, CC-BY-SA-4.0"
	if got := schema.LicenseMenu(); got != want {
		t.Errorf("LicenseMenu() = %q, want %q", got, want)
	}
}

// TestLicense_WireRoundTrip pins the omitempty + round-trip behaviour on the two
// wire fields that carry a license: an absent license must NOT serialize the key
// (omitempty), and a set license must survive a JSON marshal→unmarshal round-trip.
func TestLicense_WireRoundTrip(t *testing.T) {
	// PublishRequest.License
	var pr schema.PublishRequest
	b, err := json.Marshal(pr)
	if err != nil {
		t.Fatalf("marshal empty PublishRequest: %v", err)
	}
	if strings.Contains(string(b), "license") {
		t.Errorf("PublishRequest with no license must omit the key (omitempty); got %s", b)
	}
	pr.License = schema.LicenseCCBY
	b, err = json.Marshal(pr)
	if err != nil {
		t.Fatalf("marshal PublishRequest with license: %v", err)
	}
	if !strings.Contains(string(b), `"license":"CC-BY-4.0"`) {
		t.Errorf("PublishRequest license not serialized; got %s", b)
	}
	var prBack schema.PublishRequest
	if err := json.Unmarshal(b, &prBack); err != nil {
		t.Fatalf("unmarshal PublishRequest: %v", err)
	}
	if prBack.License != schema.LicenseCCBY {
		t.Errorf("PublishRequest license round-trip = %q, want %q", prBack.License, schema.LicenseCCBY)
	}

	// PullTranscriptInfo.License
	var pti schema.PullTranscriptInfo
	b, err = json.Marshal(pti)
	if err != nil {
		t.Fatalf("marshal empty PullTranscriptInfo: %v", err)
	}
	if strings.Contains(string(b), "license") {
		t.Errorf("PullTranscriptInfo with no license must omit the key (omitempty); got %s", b)
	}
	pti.License = schema.LicenseCC0
	b, err = json.Marshal(pti)
	if err != nil {
		t.Fatalf("marshal PullTranscriptInfo with license: %v", err)
	}
	if !strings.Contains(string(b), `"license":"CC0-1.0"`) {
		t.Errorf("PullTranscriptInfo license not serialized; got %s", b)
	}
	var ptiBack schema.PullTranscriptInfo
	if err := json.Unmarshal(b, &ptiBack); err != nil {
		t.Fatalf("unmarshal PullTranscriptInfo: %v", err)
	}
	if ptiBack.License != schema.LicenseCC0 {
		t.Errorf("PullTranscriptInfo license round-trip = %q, want %q", ptiBack.License, schema.LicenseCC0)
	}
}
