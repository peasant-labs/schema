package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// TranscriptContentReadPolicy is the raw scanner policy used by content reads.
type TranscriptContentReadPolicy = RawJSONPathPolicy

var (
	// TranscriptSourceReadPolicy is the unchanged envelope source boundary.
	TranscriptSourceReadPolicy = transcriptRawPolicy()
	// TranscriptSourceDetailReadPolicy is the unchanged bare-detail boundary.
	TranscriptSourceDetailReadPolicy = sessionDetailRawPolicy()
	// TranscriptDownloadPolicy preserves source validation without an additional
	// outer byte cap. It accepts envelope and bare-detail opaque metadata paths.
	TranscriptDownloadPolicy = TranscriptContentReadPolicy{
		MaxDocumentDepth: 64,
		OpaqueMetadataPointers: []string{
			"/" + transcriptWireSessionDetail + "/" + transcriptWireNativeMetadata + "/*/" + transcriptWireData,
			"/" + transcriptWireSessionDetail + "/" + transcriptWireEarlierHistory + "/*/" + transcriptWireNativeMetadata + "/*/" + transcriptWireData,
			"/" + transcriptWireNativeMetadata + "/*/" + transcriptWireData,
			"/" + transcriptWireEarlierHistory + "/*/" + transcriptWireNativeMetadata + "/*/" + transcriptWireData,
		},
	}
)

// TranscriptDownloadKind identifies the dispatch of one authorized body.
type TranscriptDownloadKind string

const (
	TranscriptDownloadStrict  TranscriptDownloadKind = "strict"
	TranscriptDownloadLegacy  TranscriptDownloadKind = "legacy"
	TranscriptDownloadRecords TranscriptDownloadKind = "records"
)

var AllTranscriptDownloadKinds = []TranscriptDownloadKind{
	TranscriptDownloadStrict, TranscriptDownloadLegacy, TranscriptDownloadRecords,
}

// TranscriptDownloadRead contains the served detail bytes or raw JSONL records.
type TranscriptDownloadRead struct {
	Kind    TranscriptDownloadKind
	Detail  json.RawMessage
	Records []json.RawMessage
}

// TranscriptContentStrict reports the in-band strict core of the content
// boundary. It reads the selected served detail, not an outer envelope or a
// trusted caller's harness attestation.
func TranscriptContentStrict(detail json.RawMessage) bool {
	var object map[string]json.RawMessage
	if json.Unmarshal(detail, &object) != nil {
		return false
	}
	if transcriptDownloadPiIdentity(object) {
		return true
	}
	var harness string
	if json.Unmarshal(object[transcriptWireHarness], &harness) == nil {
		harnessValue := Harness(harness)
		for _, value := range Harnesses() {
			if harnessValue == value {
				return true
			}
		}
	}
	for key := range object {
		if strings.EqualFold(key, transcriptWireRetainedUnknown) || strings.EqualFold(key, transcriptWireDiagnostics) ||
			strings.EqualFold(key, transcriptWireNativeMetadata) || transcriptThinkingMember(key) {
			return true
		}
	}
	var turns []map[string]json.RawMessage
	_ = json.Unmarshal(object[transcriptWireTurns], &turns)
	for _, turn := range turns {
		for key := range turn {
			if transcriptThinkingMember(key) {
				return true
			}
		}
		if _, ok := turn[transcriptWireUsage]; ok {
			return true
		}
		if _, ok := turn[transcriptWireSourceEntryRef]; ok {
			return true
		}
		var tools []map[string]json.RawMessage
		_ = json.Unmarshal(turn[transcriptWireToolCalls], &tools)
		for _, tool := range tools {
			for _, key := range []string{transcriptWireUsage, transcriptWireCallEntryRef, transcriptWireResultEntryRef, transcriptWireNamespace} {
				if _, ok := tool[key]; ok {
					return true
				}
			}
		}
	}
	return false
}

func transcriptDownloadPiIdentity(object map[string]json.RawMessage) bool {
	for _, name := range []string{transcriptWireHarness, transcriptWireProvider, transcriptWireModelHarness} {
		var value string
		if json.Unmarshal(object[name], &value) == nil && Harness(value) == HarnessPi {
			return true
		}
	}
	return false
}

func transcriptThinkingMember(key string) bool {
	return strings.EqualFold(key, transcriptWireThinkingLevel) || strings.EqualFold(key, transcriptWireThinkingLevelRaw) ||
		strings.EqualFold(key, transcriptWireThinkingLevelHistory)
}

// DecodeTranscriptDownloadRaw scans and dispatches an authorized content body.
// Typed validation gates strict content but never replaces the served bytes.
func DecodeTranscriptDownloadRaw(raw []byte) (TranscriptDownloadRead, error) {
	const where = "transcript download failed at schema.DecodeTranscriptDownloadRaw: "
	if !json.Valid(raw) {
		var records []json.RawMessage
		for _, line := range bytes.Split(raw, []byte("\n")) {
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}
			if err := ScanRawJSONDocument(line, TranscriptDownloadPolicy); err != nil {
				return TranscriptDownloadRead{}, fmt.Errorf("%sJSONL record is invalid: %w; repair the content before downloading", where, err)
			}
			var value any
			decoder := json.NewDecoder(bytes.NewReader(line))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				return TranscriptDownloadRead{}, fmt.Errorf("%sJSONL record cannot be decoded: %w; repair the content before downloading", where, err)
			}
			if transcriptEmbeddedPublicRoot(value, false) {
				return TranscriptDownloadRead{}, fmt.Errorf("%sJSONL embeds a public transcript root; the records cannot be exported safely; republish one public envelope", where)
			}
			records = append(records, append(json.RawMessage(nil), line...))
		}
		if len(records) == 0 {
			return TranscriptDownloadRead{}, fmt.Errorf("%sthe body has no JSON document or records; no export was produced; provide a complete content body", where)
		}
		return TranscriptDownloadRead{Kind: TranscriptDownloadRecords, Records: records}, nil
	}
	if err := ScanRawJSONDocument(raw, TranscriptDownloadPolicy); err != nil {
		return TranscriptDownloadRead{}, fmt.Errorf("%scontent scan refused the body: %w; no export was produced", where, err)
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return TranscriptDownloadRead{}, fmt.Errorf("%sthe scanned document cannot be decoded: %w; no export was produced", where, err)
	}
	detail := json.RawMessage(raw)
	object, isObject := value.(map[string]any)
	envelope := false
	if isObject {
		_, hasDetail := object[transcriptWireSessionDetail]
		_, hasVersion := object[transcriptWireContractVersion]
		envelope = hasDetail || hasVersion
		if envelope {
			var root map[string]json.RawMessage
			_ = json.Unmarshal(raw, &root)
			detail = root[transcriptWireSessionDetail]
			if _, ok := object[transcriptWireSessionDetail].(map[string]any); !ok {
				return TranscriptDownloadRead{}, fmt.Errorf("%sthe envelope has no object sessionDetail; no export was produced; supply the complete served detail", where)
			}
		}
	}
	if TranscriptContentStrict(detail) {
		var detailMembers map[string]json.RawMessage
		_ = json.Unmarshal(detail, &detailMembers)
		if transcriptDownloadPiIdentity(detailMembers) {
			var harness string
			if json.Unmarshal(detailMembers[transcriptWireHarness], &harness) != nil || Harness(harness) != HarnessPi {
				return TranscriptDownloadRead{}, fmt.Errorf("%sPi identity conflicts with the session harness; no export was produced; republish a complete Pi detail with its original harness", where)
			}
		}
		var err error
		if envelope {
			_, err = DecodeTranscriptContentRawWithPolicy(raw, TranscriptDownloadPolicy)
		} else {
			_, err = DecodeSessionDetailPayloadRawWithPolicy(raw, TranscriptDownloadPolicy)
		}
		if err != nil {
			return TranscriptDownloadRead{}, fmt.Errorf("%sthe typed content gate refused the body: %w; no export was produced", where, err)
		}
		return TranscriptDownloadRead{Kind: TranscriptDownloadStrict, Detail: append(json.RawMessage(nil), detail...)}, nil
	}
	if transcriptEmbeddedPublicRoot(value, isObject) {
		return TranscriptDownloadRead{}, fmt.Errorf("%sthe legacy body embeds a public transcript root; no export was produced; republish one supported envelope", where)
	}
	if envelope {
		kind, ok := object[transcriptWireKind].(string)
		if !ok || ContentKind(kind) != ContentKindSessionDetail {
			return TranscriptDownloadRead{}, fmt.Errorf("%sthe legacy envelope has an unsupported kind; no export was produced; supply a session_detail envelope", where)
		}
	}
	if err := transcriptDownloadObservedModels(detail); err != nil {
		return TranscriptDownloadRead{}, fmt.Errorf("%slegacy observed-model validation failed: %w; no export was produced", where, err)
	}
	return TranscriptDownloadRead{Kind: TranscriptDownloadLegacy, Detail: append(json.RawMessage(nil), detail...)}, nil
}

func transcriptDownloadObservedModels(detail json.RawMessage) error {
	var object map[string]json.RawMessage
	if json.Unmarshal(detail, &object) != nil {
		return nil
	}
	var turns []map[string]json.RawMessage
	_ = json.Unmarshal(object[transcriptWireTurns], &turns)
	for i, turn := range turns {
		raw, present := turn[transcriptWireObservedModel]
		if !present {
			continue
		}
		var value *string
		if json.Unmarshal(raw, &value) != nil {
			return fmt.Errorf("turn %d has a non-string observedModel; use a valid model identifier", i)
		}
		// Stored reads decode null and empty values as an unrecorded model.
		// Leave the original bytes intact rather than inventing evidence.
		if value == nil || *value == "" {
			continue
		}
		if _, err := NewObservedModelID(*value); err != nil {
			return fmt.Errorf("turn %d has an invalid observedModel: %w", i, err)
		}
	}
	return nil
}

func transcriptEmbeddedPublicRoot(value any, allowRoot bool) bool {
	switch node := value.(type) {
	case map[string]any:
		if !allowRoot && transcriptPublicRootShape(node) {
			return true
		}
		for key, child := range node {
			if transcriptOpaqueProviderChild(node, key, child) {
				continue
			}
			if transcriptEmbeddedPublicRoot(child, allowRoot && key == transcriptWireSessionDetail) {
				return true
			}
		}
	case []any:
		for _, child := range node {
			if transcriptEmbeddedPublicRoot(child, false) {
				return true
			}
		}
	}
	return false
}

func transcriptPublicRootShape(node map[string]any) bool {
	for key, value := range node {
		if strings.EqualFold(key, transcriptWireRetainedUnknown) || strings.EqualFold(key, transcriptWireDiagnostics) {
			return true
		}
		if strings.EqualFold(key, transcriptWireSessionDetail) {
			if key != transcriptWireSessionDetail {
				return true
			}
			if detail, ok := value.(map[string]any); ok && transcriptPublicRootShape(detail) {
				return true
			}
		}
		kind, isString := value.(string)
		if strings.EqualFold(key, transcriptWireKind) && isString && ContentKind(kind) == ContentKindSessionDetail {
			return true
		}
	}
	if value, present := node[transcriptWireTurns]; present {
		if value == nil {
			return true
		}
		if _, ok := value.([]any); ok {
			return true
		}
	}
	harness, _ := node[transcriptWireHarness].(string)
	kind, _ := node[transcriptWireKind].(string)
	if Harness(harness) == HarnessPi || ContentKind(kind) == ContentKindSessionDetail {
		return true
	}
	if detail, ok := node[transcriptWireSessionDetail].(map[string]any); ok {
		if _, ok := node[transcriptWireContractVersion].(string); ok {
			return true
		}
		if _, ok := detail[transcriptWireNativeMetadata]; ok {
			return true
		}
		return transcriptPublicRootShape(detail)
	}
	return false
}

// These are the recognized tool shapes in an open provider-defined type field.
// Other provider types are not treated as opaque tool payloads.
type transcriptProviderToolKind string

const (
	transcriptProviderToolUse            transcriptProviderToolKind = transcriptProviderToolKind(EntryTypeToolUse)
	transcriptProviderToolResult         transcriptProviderToolKind = transcriptProviderToolKind(EntryTypeToolResult)
	transcriptProviderToolCall           transcriptProviderToolKind = "toolCall"
	transcriptProviderFunctionCall       transcriptProviderToolKind = "function_call"
	transcriptProviderFunctionCallOutput transcriptProviderToolKind = "function_call_output"
)

func transcriptOpaqueProviderChild(node map[string]any, key string, child any) bool {
	legacyMessage := func(value map[string]any) bool {
		_, role := value[transcriptWireRole].(string)
		_, content := value[transcriptWireContent]
		_, parts := value[transcriptWireParts]
		return role && (content || parts)
	}
	if legacyMessage(node) && (key == transcriptWireContent || key == transcriptWireParts || key == transcriptWireToolCalls || key == transcriptWireToolCallsLegacy) {
		return true
	}
	if message, ok := child.(map[string]any); key == transcriptWireMessage && ok && legacyMessage(message) {
		return true
	}
	kind, _ := node[transcriptWireType].(string)
	switch transcriptProviderToolKind(kind) {
	case transcriptProviderToolUse, transcriptProviderToolResult, transcriptProviderToolCall, transcriptProviderFunctionCall, transcriptProviderFunctionCallOutput:
		switch key {
		case transcriptWireInput, transcriptWireArguments, transcriptWireResult, transcriptWireOutput, transcriptWireContent:
			return true
		}
	}
	return false
}
