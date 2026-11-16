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
			"/sessionDetail/nativeMetadata/*/data",
			"/sessionDetail/earlierHistory/*/nativeMetadata/*/data",
			"/nativeMetadata/*/data",
			"/earlierHistory/*/nativeMetadata/*/data",
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
	if json.Unmarshal(object["harness"], &harness) == nil {
		for _, value := range Harnesses() {
			if harness == string(value) {
				return true
			}
		}
	}
	for key := range object {
		if strings.EqualFold(key, "retainedUnknown") || strings.EqualFold(key, "diagnostics") ||
			strings.EqualFold(key, "nativeMetadata") || transcriptThinkingMember(key) {
			return true
		}
	}
	var turns []map[string]json.RawMessage
	_ = json.Unmarshal(object["turns"], &turns)
	for _, turn := range turns {
		for key := range turn {
			if transcriptThinkingMember(key) {
				return true
			}
		}
		if _, ok := turn["usage"]; ok {
			return true
		}
		if _, ok := turn["sourceEntryRef"]; ok {
			return true
		}
		var tools []map[string]json.RawMessage
		_ = json.Unmarshal(turn["toolCalls"], &tools)
		for _, tool := range tools {
			for _, key := range []string{"usage", "callEntryRef", "resultEntryRef", "namespace"} {
				if _, ok := tool[key]; ok {
					return true
				}
			}
		}
	}
	return false
}

func transcriptDownloadPiIdentity(object map[string]json.RawMessage) bool {
	for _, name := range []string{"harness", "provider", "modelHarness"} {
		var value string
		if json.Unmarshal(object[name], &value) == nil && value == string(HarnessPi) {
			return true
		}
	}
	return false
}

func transcriptThinkingMember(key string) bool {
	return strings.EqualFold(key, "thinkingLevel") || strings.EqualFold(key, "thinkingLevelRaw") ||
		strings.EqualFold(key, "thinkingLevelHistory")
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
		_, hasDetail := object["sessionDetail"]
		_, hasVersion := object["contractVersion"]
		envelope = hasDetail || hasVersion
		if envelope {
			var root map[string]json.RawMessage
			_ = json.Unmarshal(raw, &root)
			detail = root["sessionDetail"]
			if _, ok := object["sessionDetail"].(map[string]any); !ok {
				return TranscriptDownloadRead{}, fmt.Errorf("%sthe envelope has no object sessionDetail; no export was produced; supply the complete served detail", where)
			}
		}
	}
	if TranscriptContentStrict(detail) {
		var detailMembers map[string]json.RawMessage
		_ = json.Unmarshal(detail, &detailMembers)
		if transcriptDownloadPiIdentity(detailMembers) {
			var harness string
			if json.Unmarshal(detailMembers["harness"], &harness) != nil || harness != string(HarnessPi) {
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
		kind, ok := object["kind"].(string)
		if !ok || kind != string(ContentKindSessionDetail) {
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
	_ = json.Unmarshal(object["turns"], &turns)
	for i, turn := range turns {
		raw, present := turn["observedModel"]
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
			if transcriptEmbeddedPublicRoot(child, allowRoot && key == "sessionDetail") {
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
		if strings.EqualFold(key, "retainedUnknown") || strings.EqualFold(key, "diagnostics") {
			return true
		}
		if strings.EqualFold(key, "sessionDetail") {
			if key != "sessionDetail" {
				return true
			}
			if detail, ok := value.(map[string]any); ok && transcriptPublicRootShape(detail) {
				return true
			}
		}
		if strings.EqualFold(key, "kind") && value == string(ContentKindSessionDetail) {
			return true
		}
	}
	if value, present := node["turns"]; present {
		if value == nil {
			return true
		}
		if _, ok := value.([]any); ok {
			return true
		}
	}
	if node["harness"] == string(HarnessPi) || node["kind"] == string(ContentKindSessionDetail) {
		return true
	}
	if detail, ok := node["sessionDetail"].(map[string]any); ok {
		if _, ok := node["contractVersion"].(string); ok {
			return true
		}
		if _, ok := detail["nativeMetadata"]; ok {
			return true
		}
		return transcriptPublicRootShape(detail)
	}
	return false
}

func transcriptOpaqueProviderChild(node map[string]any, key string, child any) bool {
	legacyMessage := func(value map[string]any) bool {
		_, role := value["role"].(string)
		_, content := value["content"]
		_, parts := value["parts"]
		return role && (content || parts)
	}
	if legacyMessage(node) && (key == "content" || key == "parts" || key == "toolCalls" || key == "tool_calls") {
		return true
	}
	if message, ok := child.(map[string]any); key == "message" && ok && legacyMessage(message) {
		return true
	}
	switch node["type"] {
	case "tool_use", "tool_result", "toolCall", "function_call", "function_call_output":
		switch key {
		case "input", "arguments", "result", "output", "content":
			return true
		}
	}
	return false
}
