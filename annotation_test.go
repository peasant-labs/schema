package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/peasant-labs/schema"
)

// --- AnnotatorKind ---

func TestAnnotatorKind_IsValid_AllKnown(t *testing.T) {
	for _, k := range schema.AllAnnotatorKinds {
		if !k.IsValid() {
			t.Errorf("AllAnnotatorKinds[%s] should be valid", k)
		}
	}
}

func TestAnnotatorKind_IsValid_Unknown(t *testing.T) {
	if schema.AnnotatorKind("unknown").IsValid() {
		t.Error("unknown annotator kind should be invalid")
	}
	if schema.AnnotatorKind("").IsValid() {
		t.Error("empty annotator kind should be invalid")
	}
}

func TestAnnotatorKind_Priority(t *testing.T) {
	cases := []struct {
		kind     schema.AnnotatorKind
		wantPrio int
	}{
		{schema.AnnotatorHuman, 3},
		{schema.AnnotatorAgent, 2},
		{schema.AnnotatorRule, 1},
		{schema.AnnotatorKind("unknown"), 0},
	}
	for _, tc := range cases {
		if got := tc.kind.Priority(); got != tc.wantPrio {
			t.Errorf("AnnotatorKind(%q).Priority() = %d, want %d", tc.kind, got, tc.wantPrio)
		}
	}
}

func TestAnnotatorKind_Priority_HumanBeatsAgentBeatsRule(t *testing.T) {
	if schema.AnnotatorHuman.Priority() <= schema.AnnotatorAgent.Priority() {
		t.Error("human priority must be higher than agent priority")
	}
	if schema.AnnotatorAgent.Priority() <= schema.AnnotatorRule.Priority() {
		t.Error("agent priority must be higher than rule priority")
	}
}

func TestAnnotatorKind_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.AnnotatorKind("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("AnnotatorKind JSONSchema: expected enum values")
	}
}

func TestAnnotatorKind_RoundTrip_JSON(t *testing.T) {
	for _, k := range schema.AllAnnotatorKinds {
		data, err := json.Marshal(k)
		if err != nil {
			t.Fatalf("marshal %q: %v", k, err)
		}
		var got schema.AnnotatorKind
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %q: %v", k, err)
		}
		if got != k {
			t.Errorf("round-trip: got %q, want %q", got, k)
		}
	}
}

// --- AnnotationStatus ---

func TestAnnotationStatus_IsValid_AllKnown(t *testing.T) {
	for _, s := range schema.AllAnnotationStatuses {
		if !s.IsValid() {
			t.Errorf("AllAnnotationStatuses[%s] should be valid", s)
		}
	}
}

func TestAnnotationStatus_IsValid_Unknown(t *testing.T) {
	if schema.AnnotationStatus("unknown").IsValid() {
		t.Error("unknown annotation status should be invalid")
	}
}

func TestAnnotationStatus_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.AnnotationStatus("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("AnnotationStatus JSONSchema: expected enum values")
	}
}

func TestAnnotationStatus_RoundTrip_JSON(t *testing.T) {
	for _, status := range schema.AllAnnotationStatuses {
		data, err := json.Marshal(status)
		if err != nil {
			t.Fatalf("marshal %q: %v", status, err)
		}
		var got schema.AnnotationStatus
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %q: %v", status, err)
		}
		if got != status {
			t.Errorf("round-trip: got %q, want %q", got, status)
		}
	}
}

// --- ValueDomainKind ---

func TestValueDomainKind_IsValid_AllKnown(t *testing.T) {
	for _, k := range schema.AllValueDomainKinds {
		if !k.IsValid() {
			t.Errorf("AllValueDomainKinds[%s] should be valid", k)
		}
	}
}

func TestValueDomainKind_IsValid_Unknown(t *testing.T) {
	if schema.ValueDomainKind("unknown").IsValid() {
		t.Error("unknown value domain kind should be invalid")
	}
}

func TestValueDomainKind_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.ValueDomainKind("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("ValueDomainKind JSONSchema: expected enum values")
	}
}

func TestValueDomainKind_RoundTrip_JSON(t *testing.T) {
	for _, k := range schema.AllValueDomainKinds {
		data, err := json.Marshal(k)
		if err != nil {
			t.Fatalf("marshal %q: %v", k, err)
		}
		var got schema.ValueDomainKind
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %q: %v", k, err)
		}
		if got != k {
			t.Errorf("round-trip: got %q, want %q", got, k)
		}
	}
}

// --- AnnotationDatatype ---

func TestAnnotationDatatype_IsValid_AllKnown(t *testing.T) {
	for _, d := range schema.AllAnnotationDatatypes {
		if !d.IsValid() {
			t.Errorf("AllAnnotationDatatypes[%s] should be valid", d)
		}
	}
}

func TestAnnotationDatatype_IsValid_Unknown(t *testing.T) {
	if schema.AnnotationDatatype("unknown").IsValid() {
		t.Error("unknown annotation datatype should be invalid")
	}
}

func TestAnnotationDatatype_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.AnnotationDatatype("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("AnnotationDatatype JSONSchema: expected enum values")
	}
}

func TestAnnotationDatatype_RoundTrip_JSON(t *testing.T) {
	for _, d := range schema.AllAnnotationDatatypes {
		data, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("marshal %q: %v", d, err)
		}
		var got schema.AnnotationDatatype
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %q: %v", d, err)
		}
		if got != d {
			t.Errorf("round-trip: got %q, want %q", got, d)
		}
	}
}

// --- TypeOrigin ---

func TestTypeOrigin_IsValid_AllKnown(t *testing.T) {
	for _, o := range schema.AllTypeOrigins {
		if !o.IsValid() {
			t.Errorf("AllTypeOrigins[%s] should be valid", o)
		}
	}
}

func TestTypeOrigin_IsValid_Unknown(t *testing.T) {
	if schema.TypeOrigin("unknown").IsValid() {
		t.Error("unknown type origin should be invalid")
	}
}

func TestTypeOrigin_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.TypeOrigin("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("TypeOrigin JSONSchema: expected enum values")
	}
}

func TestTypeOrigin_RoundTrip_JSON(t *testing.T) {
	for _, o := range schema.AllTypeOrigins {
		data, err := json.Marshal(o)
		if err != nil {
			t.Fatalf("marshal %q: %v", o, err)
		}
		var got schema.TypeOrigin
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %q: %v", o, err)
		}
		if got != o {
			t.Errorf("round-trip: got %q, want %q", got, o)
		}
	}
}

// --- TargetKind ---

func TestTargetKind_IsValid_AllKnown(t *testing.T) {
	for _, k := range schema.AllTargetKinds {
		if !k.IsValid() {
			t.Errorf("AllTargetKinds[%s] should be valid", k)
		}
	}
}

func TestTargetKind_IsValid_Unknown(t *testing.T) {
	if schema.TargetKind("unknown").IsValid() {
		t.Error("unknown target kind should be invalid")
	}
}

func TestTargetKind_JSONSchema_HasEnum(t *testing.T) {
	s, err := schema.TargetKind("").JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() returned error: %v", err)
	}
	if len(s.Enum) == 0 {
		t.Fatal("TargetKind JSONSchema: expected enum values")
	}
}

func TestTargetKind_RoundTrip_JSON(t *testing.T) {
	for _, k := range schema.AllTargetKinds {
		data, err := json.Marshal(k)
		if err != nil {
			t.Fatalf("marshal %q: %v", k, err)
		}
		var got schema.TargetKind
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %q: %v", k, err)
		}
		if got != k {
			t.Errorf("round-trip: got %q, want %q", got, k)
		}
	}
}

// --- ValueDomain and Provenance marshaling ---

func TestValueDomain_JSONRoundTrip_Enumerated(t *testing.T) {
	vd := schema.ValueDomain{
		Kind:              schema.DomainEnumerated,
		Datatype:          schema.DatatypeText,
		PermissibleValues: []string{"approve", "deny"},
	}
	data, err := json.Marshal(vd)
	if err != nil {
		t.Fatalf("marshal ValueDomain: %v", err)
	}
	var got schema.ValueDomain
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal ValueDomain: %v", err)
	}
	if got.Kind != schema.DomainEnumerated {
		t.Errorf("Kind: got %q, want %q", got.Kind, schema.DomainEnumerated)
	}
	if got.Datatype != schema.DatatypeText {
		t.Errorf("Datatype: got %q, want %q", got.Datatype, schema.DatatypeText)
	}
	if len(got.PermissibleValues) != 2 {
		t.Errorf("PermissibleValues: got %v, want 2 elements", got.PermissibleValues)
	}
}

func TestValueDomain_JSONRoundTrip_Described(t *testing.T) {
	vd := schema.ValueDomain{
		Kind:           schema.DomainDescribed,
		Datatype:       schema.DatatypeReal,
		ConstraintSpec: `{"min":0.0,"max":1.0}`,
	}
	data, err := json.Marshal(vd)
	if err != nil {
		t.Fatalf("marshal ValueDomain: %v", err)
	}
	var got schema.ValueDomain
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal ValueDomain: %v", err)
	}
	if got.Kind != schema.DomainDescribed {
		t.Errorf("Kind: got %q, want %q", got.Kind, schema.DomainDescribed)
	}
	if got.ConstraintSpec != `{"min":0.0,"max":1.0}` {
		t.Errorf("ConstraintSpec: got %q, want %q", got.ConstraintSpec, `{"min":0.0,"max":1.0}`)
	}
	// PermissibleValues should be omitted (empty slice → nil)
	if len(got.PermissibleValues) != 0 {
		t.Errorf("PermissibleValues: expected empty/nil for described domain, got %v", got.PermissibleValues)
	}
}

func TestProvenance_JSONRoundTrip(t *testing.T) {
	p := schema.Provenance{
		Method:   "heuristic",
		Function: "classifyOutcome",
		Version:  "v1.2",
		Details:  map[string]string{"threshold": "0.5"},
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal Provenance: %v", err)
	}
	var got schema.Provenance
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal Provenance: %v", err)
	}
	if got.Method != "heuristic" {
		t.Errorf("Method: got %q, want %q", got.Method, "heuristic")
	}
	if got.Details["threshold"] != "0.5" {
		t.Errorf("Details[threshold]: got %q, want %q", got.Details["threshold"], "0.5")
	}
}

// --- AnnotationSummary marshaling ---

func TestAnnotationSummary_JSONRoundTrip_SessionTarget(t *testing.T) {
	sessionID := "99d59925-36bc-424c-a789-8be54d9702ba"
	summary := schema.AnnotationSummary{
		ID:              "42",
		TargetKind:      schema.TargetSession,
		TargetSessionID: &sessionID,
		AnnotatorKind:   schema.AnnotatorRule,
		AnnotatorName:   "outcome-classifier",
		TypeID:          "quality.session_outcome",
		TypeName:        "Session Outcome",
		Value:           "resolved",
		CreatedAt:       1700000000000,
	}
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal AnnotationSummary: %v", err)
	}
	var got schema.AnnotationSummary
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal AnnotationSummary: %v", err)
	}
	if got.ID != "42" {
		t.Errorf("ID: got %q, want 42", got.ID)
	}
	if got.TargetKind != schema.TargetSession {
		t.Errorf("TargetKind: got %q, want %q", got.TargetKind, schema.TargetSession)
	}
	if got.TargetSessionID == nil || *got.TargetSessionID != sessionID {
		t.Errorf("TargetSessionID: got %v, want %q", got.TargetSessionID, sessionID)
	}
	if got.Value != "resolved" {
		t.Errorf("Value: got %q, want %q", got.Value, "resolved")
	}
}

// TestAnnotationSummary_OmitEmpty verifies that nil pointer fields are omitted from JSON.
func TestAnnotationSummary_OmitEmpty(t *testing.T) {
	sessionID := "99d59925-36bc-424c-a789-8be54d9702ba"
	summary := schema.AnnotationSummary{
		ID:              "1",
		TargetKind:      schema.TargetSession,
		TargetSessionID: &sessionID,
		AnnotatorKind:   schema.AnnotatorRule,
		AnnotatorName:   "scope-classifier",
		TypeID:          "metadata.session_scope",
		TypeName:        "Session Scope",
		Value:           "feature",
		CreatedAt:       1700000000000,
		// Confidence, Reason, Provenance, SupersededBy, TargetAnnotID, TargetEntryIndex: nil
	}
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Verify nil pointer fields are absent from JSON.
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	for _, absent := range []string{"confidence", "reason", "provenance", "supersededBy", "targetAnnotationId", "targetEntryIndex"} {
		if _, ok := m[absent]; ok {
			t.Errorf("field %q should be omitted when nil, but was present in JSON", absent)
		}
	}
}
