package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
)

var (
	canonicalOperationFields   = map[string]struct{}{"replacement": {}, "license": {}, "associations": {}, "contentHash": {}}
	publishResponseFields      = map[string]struct{}{"transcriptId": {}, "transcriptUrl": {}, "visibility": {}, "contentHash": {}, "requestOperationFingerprint": {}, "applied": {}, "blobKey": {}, "blobSizeBytes": {}, "publishedAt": {}, "updatedAt": {}, "created": {}}
	publishAppliedFields       = map[string]struct{}{"license": {}, "associations": {}, "normalizedValues": {}}
	publishNormalizedFields    = map[string]struct{}{"rootHarness": {}, "entryHarnesses": {}, "derivedTitle": {}, "visibility": {}, "schemaVersion": {}}
	canonicalReplacementFields = map[string]struct{}{"identity": {}, "model": {}, "timestamp": {}, "source": {}, "git": {}, "project": {}, "stats": {}, "quality": {}, "entries": {}, "subagents": {}, "diagnostics": {}}
	canonicalGitFields         = map[string]struct{}{"branch": {}, "remote": {}, "worktree": {}, "tracking": {}, "commits": {}}
	licenseOperationFields     = map[string]struct{}{"kind": {}, "license": {}}
	associationOperationFields = map[string]struct{}{"kind": {}, "associations": {}}
	publishedAssociationFields = map[string]struct{}{"id": {}, "observedCommitHash": {}}
	authoritativeRequestFields = map[string]struct{}{"identity": {}, "model": {}, "contentHash": {}, "visibilityIntent": {}, "timestamp": {}, "source": {}, "git": {}, "project": {}, "stats": {}, "quality": {}, "entries": {}, "subagents": {}, "diagnostics": {}, "license": {}}
	identityFields             = fieldSet("sessionId", "schemaVersion", "parentSessionId")
	modelFields                = fieldSet("harness", "model", "harnessVersion", "hostSlug")
	timestampFields            = fieldSet("start", "end", "ingested")
	sourceFields               = fieldSet("format", "filePath")
	commitFields               = fieldSet("hash", "message", "authorName", "authorEmail", "commitTime", "authorTime")
	requestGitFields           = fieldSet("branch", "remote", "worktree", "tracking", "commits", "associations")
	projectFields              = fieldSet("hash", "filePath", "name")
	statsFields                = fieldSet("turnCount", "toolCallCount", "subagentCount", "durationMs", "tokensIn", "tokensOut", "thoughtTokens", "cachedReadTokens", "cachedWriteTokens")
	qualityFields              = fieldSet("turnCount", "subagentCount", "totalTokens", "inputTokens", "outputTokens", "toolCalls", "titleGenerated", "outcome", "filesTouched", "linesChanged", "retryLoops", "retryTokensWasted", "withinSessionReverts", "signalDensity", "specQualityScore", "explorationRatio", "scopeBreadth", "discoveryTurns", "durationMinutes", "m2TokenOutcomeRatio", "m3UniqueToolCount", "m4ErrorRecoveryCount", "m4ConsecutiveErrorMax", "m5ContextUtilizationPct", "m5PeakContextTokens", "m5AvgMessageTokens", "m6OutputSurvivalPct", "m6LinesSurvived", "m6LinesTotal", "m7SpecWordCount", "m7SpecHasExamples", "m7SpecHasConstraints", "costInputUsd", "costOutputUsd", "costReasoningUsd", "costCacheReadUsd", "costCacheWriteUsd", "costTotalUsd", "costModelId", "scope", "computedAt", "computeVersion")
	entryFields                = fieldSet("sessionId", "entryIndex", "harness", "entryType", "role", "timestampMs", "contentPreview", "tokensIn", "tokensOut", "hasToolUse", "toolKind", "toolNamesCsv", "hasThinking", "isError", "stopReason", "rawByteLength", "toolCallId", "entryId", "parentEntryId", "depth", "parentIndex", "toolInput", "toolOutput", "extra", "partType")
	subagentFields             = fieldSet("sessionId", "parentUuid")
	diagnosticsFields          = fieldSet("warnings", "partial")
	diagnosticEntryFields      = fieldSet("errorType", "location", "message", "remediation")
)

func fieldSet(names ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, name := range names {
		out[name] = struct{}{}
	}
	return out
}

func DecodeAuthoritativePublishRequest(data []byte) (AuthoritativePublishRequest, error) {
	var value AuthoritativePublishRequest
	if err := json.Unmarshal(data, &value); err != nil {
		return AuthoritativePublishRequest{}, err
	}
	if err := validateAuthoritativePublishMetadata(value); err != nil {
		return AuthoritativePublishRequest{}, err
	}
	return value, nil
}

func (value *AuthoritativePublishRequest) UnmarshalJSON(data []byte) error {
	raw, err := decodeStrictObject(data, authoritativeRequestFields, "authoritative publish request")
	if err != nil {
		return err
	}
	if err := validateSuccessorNestedJSON(raw, false); err != nil {
		return err
	}
	for _, required := range []string{"model", "contentHash"} {
		item, present := raw[required]
		if !present || bytes.Equal(bytes.TrimSpace(item), []byte("null")) {
			return publicationError("authoritative publish request", fmt.Sprintf("required field %q is missing or null", required), "send complete successor multipart metadata")
		}
	}
	if intent, present := raw["visibilityIntent"]; present && bytes.Equal(bytes.TrimSpace(intent), []byte("null")) {
		return publicationError("authoritative publish request", "optional field \"visibilityIntent\" is null", "omit it for legacy compatibility or send private/public")
	}
	type wire AuthoritativePublishRequest
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return publicationError("authoritative publish request", err.Error(), "send values matching the declared successor metadata types")
	}
	*value = AuthoritativePublishRequest(decoded)
	return validateAuthoritativePublishMetadata(*value)
}

func validateAuthoritativePublishMetadata(value AuthoritativePublishRequest) error {
	if err := value.ContentHash.Validate(); err != nil {
		return err
	}
	if !value.VisibilityIntent.IsValid() {
		return publicationError("authoritative publish request", fmt.Sprintf("visibilityIntent %q is invalid", value.VisibilityIntent), "omit it or send private/public")
	}
	if !value.Model.Harness.IsKnown() || value.Model.Model == "" {
		return publicationError("authoritative publish request", "model.harness or model.model is missing or invalid", "send a canonical harness and nonempty model identifier")
	}
	return nil
}

func DecodeCanonicalPublishOperation(data []byte) (CanonicalPublishOperation, error) {
	raw, err := decodeStrictObject(data, canonicalOperationFields, "canonical publish operation")
	if err != nil {
		return CanonicalPublishOperation{}, err
	}
	if err := validateCanonicalOperationJSON(raw); err != nil {
		return CanonicalPublishOperation{}, err
	}
	var value CanonicalPublishOperation
	if err := decodeRequiredStrict(data, canonicalOperationFields, nil, &value, "canonical publish operation"); err != nil {
		return CanonicalPublishOperation{}, err
	}
	if err := value.Validate(); err != nil {
		return CanonicalPublishOperation{}, err
	}
	return value, nil
}

func DecodePublishResponse(data []byte) (AuthoritativePublishResponse, error) {
	var value AuthoritativePublishResponse
	if err := decodeRequiredStrict(data, publishResponseFields, nil, &value, "publish response"); err != nil {
		return AuthoritativePublishResponse{}, err
	}
	if err := value.Validate(); err != nil {
		return AuthoritativePublishResponse{}, err
	}
	return value, nil
}

func (value *AuthoritativePublishResponse) UnmarshalJSON(data []byte) error {
	type wire AuthoritativePublishResponse
	var decoded wire
	if err := decodeRequiredStrict(data, publishResponseFields, nil, &decoded, "publish response"); err != nil {
		return err
	}
	*value = AuthoritativePublishResponse(decoded)
	return value.Validate()
}

func (value *PublishAppliedState) UnmarshalJSON(data []byte) error {
	raw, err := decodeStrictObject(data, publishAppliedFields, "publish applied state")
	if err != nil {
		return err
	}
	if associations, ok := raw["associations"]; ok {
		if err := validateAssociationArrayJSON(associations, "publish applied associations"); err != nil {
			return err
		}
	}
	type wire PublishAppliedState
	var decoded wire
	if err := decodeRequiredStrict(data, publishAppliedFields, map[string]struct{}{"license": {}}, &decoded, "publish applied state"); err != nil {
		return err
	}
	*value = PublishAppliedState(decoded)
	return value.Validate()
}

func validateCanonicalOperationJSON(raw map[string]json.RawMessage) error {
	replacement, err := decodeRequiredObject(raw, "replacement", canonicalReplacementFields, map[string]struct{}{"quality": {}}, "canonical replacement")
	if err != nil {
		return err
	}
	if err := validateSuccessorNestedJSON(replacement, true); err != nil {
		return err
	}
	if _, err := decodeRequiredObject(replacement, "git", canonicalGitFields, map[string]struct{}{"branch": {}, "remote": {}, "worktree": {}, "tracking": {}}, "canonical replacement git"); err != nil {
		return err
	}
	license, err := decodeRequiredObject(raw, "license", licenseOperationFields, map[string]struct{}{"license": {}}, "publish license operation")
	if err != nil {
		return err
	}
	_ = license
	associations, err := decodeRequiredObject(raw, "associations", associationOperationFields, nil, "publish association operation")
	if err != nil {
		return err
	}
	return validateAssociationArrayJSON(associations["associations"], "publish operation associations")
}

func validateSuccessorNestedJSON(raw map[string]json.RawMessage, canonical bool) error {
	objects := []struct {
		name   string
		fields map[string]struct{}
	}{
		{"identity", identityFields}, {"model", modelFields}, {"timestamp", timestampFields}, {"source", sourceFields},
		{"project", projectFields}, {"stats", statsFields}, {"quality", qualityFields}, {"diagnostics", diagnosticsFields},
	}
	for _, object := range objects {
		value, present := raw[object.name]
		if !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			continue
		}
		if _, err := decodeStrictObject(value, object.fields, "successor publication "+object.name); err != nil {
			return err
		}
	}
	gitFields := requestGitFields
	if canonical {
		gitFields = canonicalGitFields
	}
	if value, present := raw["git"]; present && !bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		git, err := decodeStrictObject(value, gitFields, "successor publication git")
		if err != nil {
			return err
		}
		if commits, present := git["commits"]; present {
			if err := validateStrictObjectArray(commits, commitFields, "successor publication commits"); err != nil {
				return err
			}
		}
	}
	for _, array := range []struct {
		name   string
		fields map[string]struct{}
	}{{"entries", entryFields}, {"subagents", subagentFields}} {
		if value, present := raw[array.name]; present {
			if err := validateStrictObjectArray(value, array.fields, "successor publication "+array.name); err != nil {
				return err
			}
		}
	}
	if value, present := raw["diagnostics"]; present && !bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		diagnostics, err := decodeStrictObject(value, diagnosticsFields, "successor publication diagnostics")
		if err != nil {
			return err
		}
		if warnings, present := diagnostics["warnings"]; present {
			if err := validateStrictObjectArray(warnings, diagnosticEntryFields, "successor publication diagnostic warnings"); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateStrictObjectArray(data json.RawMessage, fields map[string]struct{}, what string) error {
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return publicationError(what, err.Error(), "send an array of declared objects")
	}
	for index, item := range items {
		if _, err := decodeStrictObject(item, fields, fmt.Sprintf("%s[%d]", what, index)); err != nil {
			return err
		}
	}
	return nil
}

func decodeRequiredObject(parent map[string]json.RawMessage, field string, fields, nullable map[string]struct{}, what string) (map[string]json.RawMessage, error) {
	value, ok := parent[field]
	if !ok {
		return nil, publicationError(what, fmt.Sprintf("required field %q is missing", field), "return the complete object")
	}
	raw, err := decodeStrictObject(value, fields, what)
	if err != nil {
		return nil, err
	}
	for name := range fields {
		item, present := raw[name]
		if !present {
			return nil, publicationError(what, fmt.Sprintf("required field %q is missing", name), "return the complete object")
		}
		if bytes.Equal(bytes.TrimSpace(item), []byte("null")) {
			if _, allowed := nullable[name]; !allowed {
				return nil, publicationError(what, fmt.Sprintf("required field %q is null", name), "return the required non-null value")
			}
		}
	}
	return raw, nil
}

func validateAssociationArrayJSON(data json.RawMessage, what string) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return publicationError(what, "association array is null", "return a non-null complete array")
	}
	var items []json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&items); err != nil {
		return publicationError(what, err.Error(), "return an array of complete association objects")
	}
	for index, item := range items {
		if _, err := decodeRequiredObject(map[string]json.RawMessage{"association": item}, "association", publishedAssociationFields, nil, fmt.Sprintf("%s[%d]", what, index)); err != nil {
			return err
		}
	}
	return nil
}

func (value *PublishNormalizedValues) UnmarshalJSON(data []byte) error {
	type wire PublishNormalizedValues
	var decoded wire
	if err := decodeRequiredStrict(data, publishNormalizedFields, map[string]struct{}{"derivedTitle": {}}, &decoded, "publish normalized values"); err != nil {
		return err
	}
	*value = PublishNormalizedValues(decoded)
	return nil
}

func decodeRequiredStrict(data []byte, fields, nullable map[string]struct{}, destination any, what string) error {
	raw, err := decodeStrictObject(data, fields, what)
	if err != nil {
		return err
	}
	for field := range fields {
		value, present := raw[field]
		if !present {
			return publicationError(what, fmt.Sprintf("required field %q is missing", field), "return the complete authoritative object")
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			if _, ok := nullable[field]; !ok {
				return publicationError(what, fmt.Sprintf("required field %q is null", field), "return the required non-null value")
			}
		}
	}
	type noMethods struct{ Value any }
	_ = noMethods{}
	if err := json.Unmarshal(data, destination); err != nil {
		return publicationError(what, err.Error(), "send values matching the declared field types")
	}
	return nil
}
