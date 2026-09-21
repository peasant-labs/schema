package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/retained_unknown.yaml
var retainedUnknownYAML []byte

type retainedUnknownInput struct {
	PayloadUnit       string  `yaml:"payload_unit"`
	TransportBytes    int     `yaml:"transport_bytes"`
	SiblingBlocks     int     `yaml:"sibling_blocks"`
	DropRecordField   string  `yaml:"drop_record_field"`
	Harness           string  `yaml:"harness"`
	RecordPatch       string  `yaml:"record_patch"`
	SecondRecordPatch *string `yaml:"second_record_patch"`
	DetailPatch       string  `yaml:"detail_patch"`
	DropRetention     bool    `yaml:"drop_retention"`
	DropDiagnostics   bool    `yaml:"drop_diagnostics"`
	PayloadBytes      int     `yaml:"payload_bytes"`
	MetadataPartial   *bool   `yaml:"metadata_partial"`
	MetadataMissing   bool    `yaml:"metadata_missing"`
}
type retainedUnknownExpected struct {
	ValueValid            *bool `yaml:"value_valid"`
	EnvelopeValid         *bool `yaml:"envelope_valid"`
	HTMLEncodingOverLimit bool  `yaml:"html_encoding_over_limit"`
	Valid                 bool  `yaml:"valid"`
	ShapeValid            bool  `yaml:"shape_valid"`
	TypedValid            bool  `yaml:"typed_valid"`
	Capability            bool  `yaml:"capability"`
	MirrorInvalid         bool  `yaml:"mirror_invalid"`
}
type retainedUnknownFixtures struct {
	BaseDetail        string                                                         `yaml:"base_detail"`
	BaseRecord        string                                                         `yaml:"base_record"`
	RequiredHarnesses []string                                                       `yaml:"required_harnesses"`
	RequiredNames     []string                                                       `yaml:"required_names"`
	Cases             []testcase.Case[retainedUnknownInput, retainedUnknownExpected] `yaml:"cases"`
}

func decodeRetainedUnknownFixtures(data []byte) (retainedUnknownFixtures, error) {
	var f retainedUnknownFixtures
	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)
	if err := d.Decode(&f); err != nil {
		return f, err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return f, fmt.Errorf("expected one fixture document: %v", err)
	}
	if err := (testcase.Corpus[retainedUnknownInput, retainedUnknownExpected]{Cases: f.Cases}).Validate(); err != nil {
		return f, err
	}
	names := map[string]bool{}
	for _, name := range f.RequiredNames {
		if names[name] || name == "" {
			return f, fmt.Errorf("invalid required name %q", name)
		}
		names[name] = true
	}
	for _, row := range f.Cases {
		if !names[row.Name] {
			return f, fmt.Errorf("untracked case %s", row.Name)
		}
		delete(names, row.Name)
	}
	if len(names) != 0 || len(f.RequiredNames) == 0 {
		return f, fmt.Errorf("missing required cases: %v", names)
	}
	harnesses := map[string]bool{}
	for _, harness := range f.RequiredHarnesses {
		if harnesses[harness] {
			return f, fmt.Errorf("duplicate required harness %s", harness)
		}
		harnesses[harness] = true
	}
	for _, harness := range schema.AllHarnesses {
		if !harnesses[string(harness)] {
			return f, fmt.Errorf("missing harness %s", harness)
		}
		delete(harnesses, string(harness))
	}
	if len(harnesses) != 0 {
		return f, fmt.Errorf("unexpected harnesses: %v", harnesses)
	}
	for _, harness := range f.RequiredHarnesses {
		if !slices.ContainsFunc(f.Cases, func(row testcase.Case[retainedUnknownInput, retainedUnknownExpected]) bool {
			return row.Input.Harness == harness && row.Expected.Valid
		}) {
			return f, fmt.Errorf("no accepted case for %s", harness)
		}
	}
	return f, nil
}

func loadRetainedUnknownFixtures(t *testing.T) retainedUnknownFixtures {
	t.Helper()
	f, err := decodeRetainedUnknownFixtures(retainedUnknownYAML)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func retainedUnknownWire(t *testing.T, f retainedUnknownFixtures, in retainedUnknownInput) []byte {
	t.Helper()
	decode := func(raw string) map[string]any {
		var value map[string]any
		if err := json.Unmarshal([]byte(raw), &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	patch := func(value map[string]any, raw string) {
		if raw != "" {
			for k, v := range decode(raw) {
				value[k] = v
			}
		}
	}
	detail, record := decode(f.BaseDetail), decode(f.BaseRecord)
	patch(record, in.RecordPatch)
	if in.DropRecordField != "" {
		delete(record, in.DropRecordField)
	}
	if in.PayloadBytes > 0 {
		unit := in.PayloadUnit
		if unit == "" {
			unit = "x"
		}
		record["payload"] = `"` + strings.Repeat(unit, in.PayloadBytes) + `"`
	}
	detail["retainedUnknown"] = []any{record}
	if in.SiblingBlocks > 0 {
		records := make([]any, in.SiblingBlocks)
		for i := range records {
			sibling := decode(f.BaseRecord)
			sibling["pointer"] = fmt.Sprintf("/content/%d", i)
			sibling["position"] = i + 4
			records[i] = sibling
		}
		detail["retainedUnknown"] = records
	}
	if in.SecondRecordPatch != nil {
		second := decode(f.BaseRecord)
		patch(second, *in.SecondRecordPatch)
		detail["retainedUnknown"] = []any{record, second}
	}
	if in.Harness != "" {
		detail["harness"] = in.Harness
	}
	if in.DropRetention {
		delete(detail, "retainedUnknown")
	}
	if in.DropDiagnostics {
		delete(detail, "diagnostics")
	}
	patch(detail, in.DetailPatch)
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(detail); err != nil {
		t.Fatal(err)
	}
	raw := bytes.TrimSuffix(buffer.Bytes(), []byte("\n"))
	if in.TransportBytes > 0 {
		if len(raw) > in.TransportBytes {
			t.Fatal("fixture transport bytes smaller than document")
		}
		raw = append(raw, bytes.Repeat([]byte(" "), in.TransportBytes-len(raw))...)
	}
	return raw
}

func TestRetainedUnknownBoundaries(t *testing.T) {
	f := loadRetainedUnknownFixtures(t)
	publication := loadPublicationCorpus(t)
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	compiler := schema.NewJSONSchemaCompiler()
	if err := compiler.AddResource("types.json", bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	}
	contract, err := compiler.Compile("types.json#/components/schemas/SessionDetailPayload")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range f.Cases {
		t.Run(row.Name, func(t *testing.T) {
			raw := retainedUnknownWire(t, f, row.Input)
			// Capture the independent original evidence BEFORE either production
			// decoder can normalize, truncate, reorder, or omit it.
			var original struct {
				Records     []originalRetainedRecord           `json:"retainedUnknown"`
				Diagnostics *originalInterpretationDiagnostics `json:"diagnostics"`
			}
			if row.Expected.Valid {
				if err := json.Unmarshal(raw, &original); err != nil {
					t.Fatal(err)
				}
			}
			var value any
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			if err := contract.Validate(value); (err == nil) != row.Expected.ShapeValid {
				t.Errorf("generated schema=%v want valid=%v", err, row.Expected.ShapeValid)
			}
			var typed schema.SessionDetailPayload
			err := json.Unmarshal(raw, &typed)
			if err == nil {
				err = schema.ValidateSessionDetailPayload(typed)
			}
			if (err == nil) != row.Expected.TypedValid {
				t.Errorf("typed=%v want valid=%v", err, row.Expected.TypedValid)
			}
			detail, err := schema.DecodeSessionDetailPayloadRaw(raw)
			if (err == nil) != row.Expected.Valid {
				t.Fatalf("raw=%v want valid=%v", err, row.Expected.Valid)
			}
			envelopeRaw := append([]byte(`{"contractVersion":"1.0.0","kind":"session_detail","sessionDetail":`), raw...)
			envelopeRaw = append(envelopeRaw, '}')
			content, err := schema.DecodeTranscriptContentRaw(envelopeRaw)
			envelopeValid := row.Expected.Valid
			if row.Expected.EnvelopeValid != nil {
				envelopeValid = *row.Expected.EnvelopeValid
			}
			if (err == nil) != envelopeValid {
				t.Fatalf("envelope=%v want valid=%v", err, row.Expected.Valid)
			}
			if !row.Expected.Valid {
				return
			}
			if !retainedEvidenceEqual(detail, original.Records, original.Diagnostics) {
				t.Fatal("first detail decode changed original fixture evidence")
			}
			if !envelopeValid {
				return
			}
			if !retainedEvidenceEqual(*content.SessionDetail, original.Records, original.Diagnostics) {
				t.Fatal("first envelope decode changed original fixture evidence")
			}
			if row.Input.PayloadBytes > 0 {
				unit := row.Input.PayloadUnit
				if unit == "" {
					unit = "x"
				}
				want := `"` + strings.Repeat(unit, row.Input.PayloadBytes) + `"`
				if len(detail.RetainedUnknown[0].Payload) != len(want) || detail.RetainedUnknown[0].Payload != want {
					t.Fatal("expanded payload length/content differs from original recipe")
				}
			}
			if !reflect.DeepEqual(content.SessionDetail, &detail) {
				t.Fatal("envelope lost retained evidence")
			}
			encoded, err := json.Marshal(detail)
			if err != nil {
				t.Fatal(err)
			}
			if row.Expected.HTMLEncodingOverLimit {
				if len(encoded) <= 8<<20 {
					t.Fatal("HTML-escaped serializer did not cross transport boundary")
				}
				if _, err := schema.DecodeSessionDetailPayloadRaw(encoded); err == nil {
					t.Fatal("actual oversized outgoing encoding accepted")
				}
				var buffer bytes.Buffer
				encoder := json.NewEncoder(&buffer)
				encoder.SetEscapeHTML(false)
				if err := encoder.Encode(detail); err != nil {
					t.Fatal(err)
				}
				encoded = buffer.Bytes()
			}
			roundTrip, err := schema.DecodeSessionDetailPayloadRaw(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(roundTrip.RetainedUnknown, detail.RetainedUnknown) || !reflect.DeepEqual(roundTrip.Diagnostics, detail.Diagnostics) {
				t.Fatal("standalone export lost evidence or partial state")
			}
			if len(detail.RetainedUnknown) > 0 {
				var operation schema.CanonicalPublishOperation
				if err := json.Unmarshal([]byte(publication.FingerprintMutations.Base), &operation); err != nil {
					t.Fatal(err)
				}
				canonicalContent, err := json.Marshal(content)
				if err != nil {
					t.Fatal(err)
				}
				operation.ContentHash = schema.ComputeTranscriptContentHash(canonicalContent)
				before, err := schema.FingerprintPublishOperation(operation)
				if err != nil {
					t.Fatal(err)
				}
				stripped := detail
				stripped.RetainedUnknown = nil
				stripped.Diagnostics = nil
				strippedBytes, err := json.Marshal(schema.TranscriptContent{ContractVersion: "1.0.0", Kind: schema.ContentKindSessionDetail, SessionDetail: &stripped})
				if err != nil {
					t.Fatal(err)
				}
				operation.ContentHash = schema.ComputeTranscriptContentHash(strippedBytes)
				after, err := schema.FingerprintPublishOperation(operation)
				if err != nil {
					t.Fatal(err)
				}
				if before == after {
					t.Fatal("canonical publish fingerprint failed to bind retained content and partial state")
				}
				payloadOnly := detail
				payloadOnly.RetainedUnknown = append([]schema.RetainedUnknownRecord(nil), detail.RetainedUnknown...)
				payloadOnly.RetainedUnknown[0].Payload += " "
				mutatedBytes, err := json.Marshal(schema.TranscriptContent{ContractVersion: "1.0.0", Kind: schema.ContentKindSessionDetail, SessionDetail: &payloadOnly})
				if err != nil {
					t.Fatal(err)
				}
				operation.ContentHash = schema.ComputeTranscriptContentHash(mutatedBytes)
				mutatedHash, err := schema.FingerprintPublishOperation(operation)
				if err != nil {
					t.Fatal(err)
				}
				if before == mutatedHash {
					t.Fatal("payload-only byte mutation did not change publication fingerprint")
				}
			}
			required := schema.RequiredContentCapabilities(detail)
			if slices.Contains(required, schema.ContentCapabilityRetainedUnknownV1) != row.Expected.Capability {
				t.Fatalf("required=%v", required)
			}
			if row.Expected.Capability {
				if !slices.Contains(schema.MissingContentCapabilities(nil, required), schema.ContentCapabilityRetainedUnknownV1) {
					t.Fatal("missing receiver support did not refuse")
				}
				if len(schema.MissingContentCapabilities(schema.AllContentCapabilities, required)) != 0 {
					t.Fatal("supported receiver refused")
				}
			}
			partial := true
			if row.Input.MetadataPartial != nil {
				partial = *row.Input.MetadataPartial
			}
			metadata := schema.UnifiedMetadata{SessionID: schema.SessionID(detail.ID), ModelHarness: detail.Harness, Diagnostics: schema.DiagnosticsInfo{Partial: &partial}}
			if detail.Diagnostics != nil && row.Input.MetadataPartial == nil {
				partial = detail.Diagnostics.Partial
			}
			if row.Input.MetadataMissing {
				metadata.Diagnostics.Partial = nil
			}
			_, _, _, err = schema.BuildAuthoritativePublicationProjections(detail, metadata, 11, false)
			if (err != nil) != row.Expected.MirrorInvalid {
				t.Errorf("metadata mirror=%v want invalid=%v", err, row.Expected.MirrorInvalid)
			}
		})
	}
}

// These independent fixture wire structs intentionally have no production
// decoding methods. A future custom contract decoder cannot normalize both
// the value under test and its expected original evidence through one codec.
type originalRetainedRecord struct {
	SourceRef   string `json:"sourceRef"`
	RecordIndex int64  `json:"recordIndex"`
	Position    int64  `json:"position"`
	Pointer     string `json:"pointer"`
	Namespace   string `json:"namespace"`
	Kind        string `json:"kind"`
	Payload     string `json:"payload"`
}
type originalInterpretationDiagnostics struct {
	Partial bool `json:"partial"`
}

func retainedEvidenceEqual(actual schema.SessionDetailPayload, records []originalRetainedRecord, diagnostics *originalInterpretationDiagnostics) bool {
	if len(actual.RetainedUnknown) != len(records) || (actual.Diagnostics == nil) != (diagnostics == nil) {
		return false
	}
	if diagnostics != nil && actual.Diagnostics.Partial != diagnostics.Partial {
		return false
	}
	for i, expected := range records {
		record := actual.RetainedUnknown[i]
		if record.SourceRef != expected.SourceRef || record.RecordIndex != expected.RecordIndex || record.Position != expected.Position || record.Pointer != expected.Pointer || record.Namespace != expected.Namespace || record.Kind != expected.Kind || record.Payload != expected.Payload {
			return false
		}
	}
	return true
}

func TestRetainedUnknownFixtureDeletionProtection(t *testing.T) {
	f := loadRetainedUnknownFixtures(t)
	for i, row := range f.Cases {
		t.Run(row.Name, func(t *testing.T) {
			copy := f
			copy.Cases = append(append([]testcase.Case[retainedUnknownInput, retainedUnknownExpected]{}, f.Cases[:i]...), f.Cases[i+1:]...)
			data, err := yaml.Marshal(copy)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := decodeRetainedUnknownFixtures(data); err == nil {
				t.Fatal("deleted required fixture accepted")
			}
		})
	}
	if _, err := decodeRetainedUnknownFixtures(append(append([]byte{}, retainedUnknownYAML...), []byte("\nunknown_field: true\n")...)); err == nil {
		t.Fatal("unknown fixture field accepted")
	}
	if _, err := decodeRetainedUnknownFixtures(append(append([]byte{}, retainedUnknownYAML...), []byte("\n---\n{}\n")...)); err == nil {
		t.Fatal("trailing fixture document accepted")
	}
}
