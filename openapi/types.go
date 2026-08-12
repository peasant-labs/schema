package openapi

import (
	"fmt"
	"reflect"
	"strings"

	schema "github.com/peasant-labs/schema"
	jsonschema "github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi31"
)

// TypeCatalogEntry is one canonical public Go contract type registered in the
// language-neutral Types document. It is the production source of truth used by
// both OpenAPI generation and the TypeScript facade generator.
type TypeCatalogEntry struct {
	Name  string
	Value interface{}
}

// TypeCatalogEntries returns the complete public wire/domain catalog. Test-only
// fixture structs and registry/service interfaces deliberately do not belong to
// this language-neutral contract surface.
func TypeCatalogEntries() []TypeCatalogEntry {
	return []TypeCatalogEntry{
		{"ActivityEdge", new(schema.ActivityEdge)}, {"AnnotationAxis", new(schema.AnnotationAxis)},
		{"AnnotationDatatype", new(schema.AnnotationDatatype)}, {"AnnotationEntryTarget", new(schema.AnnotationEntryTarget)},
		{"AnnotationManifestResponse", new(schema.AnnotationManifestResponse)}, {"AnnotationPushItem", new(schema.AnnotationPushItem)},
		{"AnnotationPushRequest", new(schema.AnnotationPushRequest)}, {"AnnotationPushResponse", new(schema.AnnotationPushResponse)},
		{"AnnotationPushResult", new(schema.AnnotationPushResult)}, {"AnnotationPushStatus", new(schema.AnnotationPushStatus)},
		{"AnnotationStatus", new(schema.AnnotationStatus)}, {"AnnotationSummary", new(schema.AnnotationSummary)},
		{"AnnotationTypeSummary", new(schema.AnnotationTypeSummary)}, {"AnnotationsPayload", new(schema.AnnotationsPayload)},
		{"AnnotatorKind", new(schema.AnnotatorKind)},
		{"AnnotatorSummary", new(schema.AnnotatorSummary)}, {"AssociationConclusion", new(schema.AssociationConclusion)},
		{"AssociationEvidenceKind", new(schema.AssociationEvidenceKind)}, {"AssociationEvidenceObservation", new(schema.AssociationEvidenceObservation)},
		{"AssociationID", new(schema.AssociationID)}, {"BatchCreateAnnotationsErrorResponse", new(schema.BatchCreateAnnotationsErrorResponse)},
		{"BatchCreateAnnotationsRequest", new(schema.BatchCreateAnnotationsRequest)}, {"BatchCreateAnnotationsResponse", new(schema.BatchCreateAnnotationsResponse)},
		{"BuiltinCommand", new(schema.BuiltinCommand)}, {"ChangeBinding", new(schema.ChangeBinding)},
		{"ChangeDetailPayload", new(schema.ChangeDetailPayload)}, {"ChangeDiffPayload", new(schema.ChangeDiffPayload)},
		{"ChangeSession", new(schema.ChangeSession)}, {"ChangeSummary", new(schema.ChangeSummary)},
		{"ChannelSubscription", new(schema.ChannelSubscription)}, {"ChannelTopic", new(schema.ChannelTopic)},
		{"ChildSessionRef", new(schema.ChildSessionRef)}, {"CLILoginQuery", new(schema.CLILoginQuery)},
		{"ClientMessage", new(schema.ClientMessage)}, {"CommitInfo", new(schema.CommitInfo)},
		{"CommitRef", new(schema.CommitRef)}, {"Confidence", new(schema.Confidence)},
		{"ContentKind", new(schema.ContentKind)},
		{"ContractVersion", new(schema.ContractVersion)}, {"CreateAnnotationRequest", new(schema.CreateAnnotationRequest)},
		{"CreateAnnotationResponse", new(schema.CreateAnnotationResponse)}, {"DashboardPayload", new(schema.DashboardPayload)},
		{"DayStats", new(schema.DayStats)}, {"DecayLevel", new(schema.DecayLevel)},
		{"DiagnosticEntry", new(schema.DiagnosticEntry)}, {"DiagnosticsInfo", new(schema.DiagnosticsInfo)},
		{"DiffHunk", new(schema.DiffHunk)}, {"DiffLine", new(schema.DiffLine)}, {"DiffLineKind", new(schema.DiffLineKind)},
		{"EdgeViolation", new(schema.EdgeViolation)}, {"EdgeViolationKind", new(schema.EdgeViolationKind)},
		{"EntryType", new(schema.EntryType)}, {"ExchangeCodeRequest", new(schema.ExchangeCodeRequest)},
		{"ExchangeCodeResponse", new(schema.ExchangeCodeResponse)}, {"FamiliarityPayload", new(schema.FamiliarityPayload)},
		{"FileChange", new(schema.FileChange)}, {"FileChangeStatus", new(schema.FileChangeStatus)},
		{"FileFamiliarity", new(schema.FileFamiliarity)}, {"FrictionCluster", new(schema.FrictionCluster)},
		{"GitContext", new(schema.GitContext)}, {"Harness", new(schema.Harness)},
		{"HealthResponse", new(schema.HealthResponse)}, {"HostSlug", new(schema.HostSlug)},
		{"InsightClassification", new(schema.InsightClassification)}, {"InsightEvidence", new(schema.InsightEvidence)},
		{"InsightKind", new(schema.InsightKind)}, {"InsightProvenance", new(schema.InsightProvenance)},
		{"InteractionType", new(schema.InteractionType)}, {"License", new(schema.License)},
		{"MapEdge", new(schema.MapEdge)}, {"MapGraphPayload", new(schema.MapGraphPayload)},
		{"MapNode", new(schema.MapNode)}, {"MapNodeDetailPayload", new(schema.MapNodeDetailPayload)},
		{"MapNodeKind", new(schema.MapNodeKind)}, {"MapSlice", new(schema.MapSlice)},
		{"MessageType", new(schema.MessageType)}, {"MockConfigResponse", new(schema.MockConfigResponse)},
		{"ModelID", new(schema.ModelID)}, {"ModelInfo", new(schema.ModelInfo)}, {"ObservedModelID", new(schema.ObservedModelID)},
		{"ProjectContext", new(schema.ProjectContext)}, {"ProjectHash", new(schema.ProjectHash)},
		{"CanonicalPublishGitContext", new(schema.CanonicalPublishGitContext)}, {"CanonicalPublishReplacement", new(schema.CanonicalPublishReplacement)},
		{"AuthoritativeSessionIdentity", new(schema.AuthoritativeSessionIdentity)}, {"AuthoritativeModelInfo", new(schema.AuthoritativeModelInfo)},
		{"AuthoritativeTimestampInfo", new(schema.AuthoritativeTimestampInfo)}, {"AuthoritativeSourceInfo", new(schema.AuthoritativeSourceInfo)},
		{"AuthoritativeCommitInfo", new(schema.AuthoritativeCommitInfo)}, {"AuthoritativeGitContext", new(schema.AuthoritativeGitContext)},
		{"AuthoritativeProjectContext", new(schema.AuthoritativeProjectContext)}, {"AuthoritativeSessionStats", new(schema.AuthoritativeSessionStats)},
		{"AuthoritativeQualityMetrics", new(schema.AuthoritativeQualityMetrics)}, {"AuthoritativeSessionEntry", new(schema.AuthoritativeSessionEntry)},
		{"AuthoritativeSubagentRef", new(schema.AuthoritativeSubagentRef)}, {"AuthoritativeDiagnosticEntry", new(schema.AuthoritativeDiagnosticEntry)}, {"AuthoritativeDiagnosticsInfo", new(schema.AuthoritativeDiagnosticsInfo)},
		{"AuthoritativePublishRequest", new(schema.AuthoritativePublishRequest)},
		{"CanonicalPublishOperation", new(schema.CanonicalPublishOperation)}, {"PublishLicenseOperation", new(schema.PublishLicenseOperation)},
		{"PublishAssociationOperation", new(schema.PublishAssociationOperation)}, {"PublishOperationKind", new(schema.PublishOperationKind)},
		{"VisibilityIntent", new(schema.VisibilityIntent)}, {"TranscriptContentHash", new(schema.TranscriptContentHash)},
		{"PublishRequestFingerprint", new(schema.PublishRequestFingerprint)}, {"PublishAppliedState", new(schema.PublishAppliedState)},
		{"PublishNormalizedValues", new(schema.PublishNormalizedValues)},
		{"AuthoritativePublishResponse", new(schema.AuthoritativePublishResponse)},
		{"ProjectResolutionPayload", new(schema.ProjectResolutionPayload)},
		{"ProjectSummariesPayload", new(schema.ProjectSummariesPayload)}, {"ProjectSummary", new(schema.ProjectSummary)},
		{"ProjectTasksPayload", new(schema.ProjectTasksPayload)}, {"Provenance", new(schema.Provenance)},
		{"PublishRequest", new(schema.PublishRequest)}, {"PublishResponse", new(schema.PublishResponse)},
		{"PublishedAssociation", new(schema.PublishedAssociation)}, {"PullAnnotation", new(schema.PullAnnotation)}, {"PullListResponse", new(schema.PullListResponse)},
		{"PullSkipGateItem", new(schema.PullSkipGateItem)}, {"PullSkipGateRequest", new(schema.PullSkipGateRequest)},
		{"PullSkipGateResponse", new(schema.PullSkipGateResponse)}, {"PullSkipGateResult", new(schema.PullSkipGateResult)},
		{"PullTranscriptInfo", new(schema.PullTranscriptInfo)}, {"QualityMetrics", new(schema.QualityMetrics)},
		{"QualityPayload", new(schema.QualityPayload)}, {"QualitySession", new(schema.QualitySession)},
		{"ReadAttributionState", new(schema.ReadAttributionState)}, {"ReadStateGrade", new(schema.ReadStateGrade)},
		{"RedactionInfo", new(schema.RedactionInfo)}, {"ReviewListPayload", new(schema.ReviewListPayload)},
		{"ReviewSuggestion", new(schema.ReviewSuggestion)}, {"RewriteMethod", new(schema.RewriteMethod)},
		{"RewriteResolution", new(schema.RewriteResolution)}, {"RewrittenCommit", new(schema.RewrittenCommit)},
		{"Role", new(schema.Role)},
		{"ScaleKind", new(schema.ScaleKind)}, {"SchemaVersionResponse", new(schema.SchemaVersionResponse)},
		{"SearchPayload", new(schema.SearchPayload)}, {"SearchResult", new(schema.SearchResult)},
		{"ServerMessage", new(schema.ServerMessage)}, {"SessionAssociation", new(schema.SessionAssociation)},
		{"SessionDetailPayload", new(schema.SessionDetailPayload)},
		{"SessionEntry", new(schema.SessionEntry)}, {"SessionID", new(schema.SessionID)},
		{"SessionIdentity", new(schema.SessionIdentity)}, {"SessionInsight", new(schema.SessionInsight)},
		{"SessionOutcome", new(schema.SessionOutcome)},
		{"SessionScorecard", new(schema.SessionScorecard)}, {"SessionStats", new(schema.SessionStats)},
		{"SessionSummary", new(schema.SessionSummary)}, {"SessionsPayload", new(schema.SessionsPayload)},
		{"ShutdownResponse", new(schema.ShutdownResponse)}, {"SourceFormat", new(schema.SourceFormat)},
		{"SourceInfo", new(schema.SourceInfo)},
		{"StopReason", new(schema.StopReason)}, {"SubagentRef", new(schema.SubagentRef)},
		{"TargetKind", new(schema.TargetKind)}, {"TaskSummary", new(schema.TaskSummary)},
		{"TaxonomyFamilyNode", new(schema.TaxonomyFamilyNode)}, {"TaxonomyNode", new(schema.TaxonomyNode)},
		{"TimestampInfo", new(schema.TimestampInfo)}, {"ToolCallDetail", new(schema.ToolCallDetail)},
		{"ToolCallKind", new(schema.ToolCallKind)}, {"TranscriptContent", new(schema.TranscriptContent)},
		{"TranscriptID", new(schema.TranscriptID)},
		{"TranscriptUpdateLicense", new(schema.TranscriptUpdateLicense)},
		{"TranscriptUpdateRequest", new(schema.TranscriptUpdateRequest)},
		{"TranscriptUpdateVisibility", new(schema.TranscriptUpdateVisibility)},
		{"OwnerUpdateLicenseIntent", new(schema.OwnerUpdateLicenseIntent)},
		{"OwnerTranscriptUpdateRequest", new(schema.OwnerTranscriptUpdateRequest)},
		{"OwnerTranscriptUpdateResponse", new(schema.OwnerTranscriptUpdateResponse)},
		{"TrendsPayload", new(schema.TrendsPayload)},
		{"TimelineSessionRef", new(schema.TimelineSessionRef)}, {"TurnDetail", new(schema.TurnDetail)},
		{"TypeOrigin", new(schema.TypeOrigin)},
		{"UnifiedMetadata", new(schema.UnifiedMetadata)}, {"UnusualSignal", new(schema.UnusualSignal)},
		{"ValueDomain", new(schema.ValueDomain)}, {"ValueDomainKind", new(schema.ValueDomainKind)},
		{"Visibility", new(schema.Visibility)}, {"WalkthroughStep", new(schema.WalkthroughStep)},
		{"WalkthroughTrail", new(schema.WalkthroughTrail)},
	}
}

// BuildTypesSpec builds the comprehensive OpenAPI 3.1 type catalog. It has no
// paths: the named components are the canonical cross-language contract.
func BuildTypesSpec() (*openapi31.Spec, error) {
	r := openapi31.NewReflector()
	registerHarnessSchema(r)
	r.Spec.Info.
		WithTitle("Peasant Types").
		// Derived from TypesVersion (single source; see package doc).
		WithVersion(TypesVersion).
		WithDescription("Shared domain type catalog for Peasant APIs. No operations; register these " +
			"components for SDK code generation and cross-API type sharing.")

	jsr := r.JSONSchemaReflector()
	directSchemas := make(map[string]map[string]interface{}, len(TypeCatalogEntries()))
	for _, t := range TypeCatalogEntries() {
		var definitionErr error
		s, err := jsr.Reflect(t.Value, jsonschema.CollectDefinitions(
			func(name string, defSchema jsonschema.Schema) {
				// Strip the "Schema" prefix that swaggest derives from the package alias,
				// so transitive types are stored as "Provider", "Role", etc.
				plainName := canonicalTypeName(name)
				sm, smErr := defSchema.ToSchemaOrBool().ToSimpleMap()
				if smErr != nil {
					definitionErr = fmt.Errorf("marshal transitive shared type %s while reflecting %s: %w", plainName, t.Name, smErr)
					return
				}
				r.SpecEns().ComponentsEns().WithSchemasItem(plainName, sm)
			},
		))
		if err != nil {
			return nil, fmt.Errorf("reflect shared type %s: %w", t.Name, err)
		}
		if definitionErr != nil {
			return nil, definitionErr
		}
		sm, err := s.ToSchemaOrBool().ToSimpleMap()
		if err != nil {
			return nil, fmt.Errorf("marshal shared type %s: %w", t.Name, err)
		}
		applyGoRequiredFields(sm, reflect.TypeOf(t.Value))
		directSchemas[t.Name] = sm
	}
	// Reflection discovers transitive definitions repeatedly. Re-apply every
	// explicitly catalogued Go type last so discovery order can never overwrite
	// its canonical name or Go-requiredness metadata.
	for name, schemaMap := range directSchemas {
		r.SpecEns().ComponentsEns().WithSchemasItem(name, schemaMap)
	}
	if err := closeStrictComponents(r.SpecEns().ComponentsEns().Schemas); err != nil {
		return nil, err
	}
	delete(r.SpecEns().ComponentsEns().Schemas, "BestiaryHarness")
	delete(r.SpecEns().ComponentsEns().Schemas, "Provider")

	// Fix $ref values: JSON Schema Draft 4 CollectDefinitions uses "#/definitions/"
	// but OAS 3.1 components live under "#/components/schemas/". Walk all component
	// schemas and rewrite any "#/definitions/SchemaX" ref to "#/components/schemas/X"
	// (stripping the "Schema" prefix to match the plain component names above).
	if comps := r.SpecEns().Components; comps != nil {
		for _, schemaMap := range comps.Schemas {
			fixSharedDefinitionRefs(schemaMap)
		}
	}
	if err := applyAnnotationPushIngressConstraints(r.Spec); err != nil {
		return nil, err
	}

	return r.Spec, nil
}

// applyGoRequiredFields preserves Go's JSON presence contract in the canonical
// Types document. API operation documents may intentionally have distinct
// validation policy; the language binding catalog mirrors the Go wire shape:
// every exported JSON field without omitempty is present.
func applyGoRequiredFields(schemaMap map[string]interface{}, valueType reflect.Type) {
	for valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}
	if valueType.Kind() != reflect.Struct {
		return
	}
	var required []interface{}
	properties, _ := schemaMap["properties"].(map[string]interface{})
	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("json")
		parts := strings.Split(tag, ",")
		name := parts[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		optional := false
		for _, option := range parts[1:] {
			optional = optional || option == "omitempty"
		}
		if !optional {
			required = append(required, name)
		}
		if field.Type.Kind() == reflect.Pointer {
			if property, ok := properties[name].(map[string]interface{}); ok {
				properties[name] = nullableSchema(property)
			}
		}
	}
	if len(required) > 0 {
		schemaMap["required"] = required
	} else {
		delete(schemaMap, "required")
	}
}

func nullableSchema(property map[string]interface{}) map[string]interface{} {
	if propertyType, ok := property["type"].(string); ok {
		copy := make(map[string]interface{}, len(property))
		for key, value := range property {
			copy[key] = value
		}
		copy["type"] = []interface{}{propertyType, "null"}
		return copy
	}
	return map[string]interface{}{"anyOf": []interface{}{property, map[string]interface{}{"type": "null"}}}
}

func canonicalTypeName(name string) string {
	name = strings.TrimPrefix(name, "Schema")
	if name == "BestiaryHarness" || name == "Provider" {
		return "Harness"
	}
	return name
}

// fixSharedDefinitionRefs is like fixDefinitionRefs but also strips the "Schema"
// prefix from definition names, normalizing "#/definitions/SchemaX" to
// "#/components/schemas/X". Used by BuildTypesSpec where all components
// are stored under plain names (no "Schema" prefix).
func fixSharedDefinitionRefs(m map[string]interface{}) {
	const oldPrefix = "#/definitions/"
	const newPrefix = "#/components/schemas/"
	for k, v := range m {
		switch val := v.(type) {
		case string:
			if k == "$ref" && strings.HasPrefix(val, oldPrefix) {
				name := canonicalTypeName(strings.TrimPrefix(val, oldPrefix))
				m[k] = newPrefix + name
			}
		case map[string]interface{}:
			fixSharedDefinitionRefs(val)
		case []interface{}:
			for _, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					fixSharedDefinitionRefs(itemMap)
				}
			}
		}
	}
}

// fixDefinitionRefs recursively walks a schema map and rewrites any $ref values
// that use the JSON Schema Draft 4 "#/definitions/" prefix to the OAS 3.1
// "#/components/schemas/" prefix.
func fixDefinitionRefs(m map[string]interface{}) {
	const oldPrefix = "#/definitions/"
	const newPrefix = "#/components/schemas/"
	for k, v := range m {
		switch val := v.(type) {
		case string:
			if k == "$ref" && strings.HasPrefix(val, oldPrefix) {
				m[k] = newPrefix + strings.TrimPrefix(val, oldPrefix)
			}
		case map[string]interface{}:
			fixDefinitionRefs(val)
		case []interface{}:
			for _, item := range val {
				if itemMap, ok := item.(map[string]interface{}); ok {
					fixDefinitionRefs(itemMap)
				}
			}
		}
	}
}

// addComponentSchema reflects a Go type and registers it as a named component schema
// in the OpenAPI spec. Use this for types not naturally referenced by any operation.
func addComponentSchema(r *openapi31.Reflector, name string, val interface{}) error {
	jsr := r.JSONSchemaReflector()
	s, err := jsr.Reflect(val)
	if err != nil {
		return fmt.Errorf("reflect component schema %s: %w", name, err)
	}
	sm, err := s.ToSchemaOrBool().ToSimpleMap()
	if err != nil {
		return fmt.Errorf("marshal component schema %s: %w", name, err)
	}
	r.SpecEns().ComponentsEns().WithSchemasItem(name, sm)
	return nil
}

// addRESTOp adds a REST operation to the reflector.
func addRESTOp(r *openapi31.Reflector, method, path, opID, desc string, tags []string, req, resp interface{}) error {
	oc, err := r.NewOperationContext(method, path)
	if err != nil {
		return fmt.Errorf("new operation %s %s: %w", method, path, err)
	}
	if req != nil {
		oc.AddReqStructure(req)
	}
	if resp != nil {
		oc.AddRespStructure(resp)
	}
	oc.SetDescription(desc)
	oc.SetID(opID)
	for _, tag := range tags {
		oc.SetTags(tag)
	}
	if err := r.AddOperation(oc); err != nil {
		return fmt.Errorf("add operation %s %s: %w", method, path, err)
	}
	return nil
}

// strictComponents names the components whose reflected schema is deliberately
// TIGHTER than Go reflection produces on its own: unknown properties are
// rejected rather than ignored, and nullability is kept only for fields whose
// wire semantics explicitly distinguish null from a missing or concrete value.
//
// This is an explicit, source-owned list rather than a blanket rule because
// closing a component is a contract promise, not a style preference: most of
// this catalog is intentionally open so a producer can add a field without
// breaking every consumer. Only a request body whose whole purpose is to say
// exactly what changed belongs here, where an unrecognized field means the
// caller asked for something the server will silently drop.
var strictComponents = []string{"TranscriptUpdateRequest", "AuthoritativePublishRequest", "AuthoritativeSessionIdentity", "AuthoritativeModelInfo", "AuthoritativeTimestampInfo", "AuthoritativeSourceInfo", "AuthoritativeCommitInfo", "AuthoritativeGitContext", "AuthoritativeProjectContext", "AuthoritativeSessionStats", "AuthoritativeQualityMetrics", "AuthoritativeSessionEntry", "AuthoritativeSubagentRef", "AuthoritativeDiagnosticEntry", "AuthoritativeDiagnosticsInfo", "CanonicalPublishGitContext", "CanonicalPublishReplacement", "CanonicalPublishOperation", "PublishLicenseOperation", "PublishAssociationOperation", "PublishAppliedState", "PublishNormalizedValues", "AuthoritativePublishResponse", "PublishedAssociation", "OwnerTranscriptUpdateRequest", "OwnerTranscriptUpdateResponse"}

var strictNullableProperties = map[string]map[string]struct{}{
	"AuthoritativeTimestampInfo":    {"ingested": {}},
	"AuthoritativeGitContext":       {"branch": {}, "remote": {}, "worktree": {}, "tracking": {}},
	"AuthoritativeSessionStats":     {"thoughtTokens": {}, "cachedReadTokens": {}, "cachedWriteTokens": {}},
	"AuthoritativeQualityMetrics":   {"turnCount": {}, "subagentCount": {}, "totalTokens": {}, "inputTokens": {}, "outputTokens": {}, "toolCalls": {}, "titleGenerated": {}, "outcome": {}, "filesTouched": {}, "linesChanged": {}, "retryLoops": {}, "retryTokensWasted": {}, "withinSessionReverts": {}, "signalDensity": {}, "specQualityScore": {}, "explorationRatio": {}, "scopeBreadth": {}, "discoveryTurns": {}, "durationMinutes": {}, "m2TokenOutcomeRatio": {}, "m3UniqueToolCount": {}, "m4ErrorRecoveryCount": {}, "m4ConsecutiveErrorMax": {}, "m5ContextUtilizationPct": {}, "m5PeakContextTokens": {}, "m5AvgMessageTokens": {}, "m6OutputSurvivalPct": {}, "m6LinesSurvived": {}, "m6LinesTotal": {}, "m7SpecWordCount": {}, "m7SpecHasExamples": {}, "m7SpecHasConstraints": {}, "costInputUsd": {}, "costOutputUsd": {}, "costReasoningUsd": {}, "costCacheReadUsd": {}, "costCacheWriteUsd": {}, "costTotalUsd": {}, "costModelId": {}, "scope": {}, "computedAt": {}, "computeVersion": {}},
	"AuthoritativeSessionEntry":     {"timestampMs": {}, "contentPreview": {}, "tokensIn": {}, "tokensOut": {}, "toolKind": {}, "toolNamesCsv": {}, "stopReason": {}, "rawByteLength": {}, "toolCallId": {}, "entryId": {}, "parentEntryId": {}, "parentIndex": {}, "toolInput": {}, "toolOutput": {}, "extra": {}, "partType": {}},
	"AuthoritativeDiagnosticsInfo":  {"partial": {}},
	"CanonicalPublishGitContext":    {"branch": {}, "remote": {}, "worktree": {}, "tracking": {}},
	"CanonicalPublishReplacement":   {"quality": {}},
	"PublishLicenseOperation":       {"license": {}},
	"PublishAppliedState":           {"license": {}},
	"PublishNormalizedValues":       {"derivedTitle": {}},
	"OwnerTranscriptUpdateRequest":  {"license": {}},
	"OwnerTranscriptUpdateResponse": {"title": {}, "description": {}, "license": {}},
}

// closeStrictComponents applies that tightening after reflection and
// requiredness. Reflection maps a Go pointer to a nullable schema, which is the
// right default for a payload where null and absent mean the same thing; on
// these components they do not, so the null arm is removed and the object is
// closed. Component harmonization then carries the tightened schema into every
// API document that references the same canonical type, so the Types catalog and
// the operation cannot disagree about what the body accepts.
func closeStrictComponents(components map[string]map[string]interface{}) error {
	for _, name := range strictComponents {
		component, ok := components[name]
		if !ok {
			return fmt.Errorf("close strict component %s: it is named as strict but absent from the Types catalog; either add it to TypeCatalogEntries or remove it from strictComponents, because a strict name that matches nothing would silently stop closing anything", name)
		}
		component["additionalProperties"] = false
		properties, ok := component["properties"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("close strict component %s: it declares no properties map, so the non-nullable tightening below would silently apply to nothing", name)
		}
		for propertyName, raw := range properties {
			property, ok := raw.(map[string]interface{})
			if !ok {
				return fmt.Errorf("close strict component %s property %s: the property schema is not an object and cannot be tightened", name, propertyName)
			}
			if _, nullable := strictNullableProperties[name][propertyName]; !nullable {
				properties[propertyName] = dropNullArm(property)
			}
		}
	}
	return nil
}

// dropNullArm removes the null alternative reflection adds for a Go pointer,
// collapsing anyOf[X, null] back to X and ["null","string"] back to "string".
// A property that carries no null arm is returned unchanged.
func dropNullArm(property map[string]interface{}) map[string]interface{} {
	if arms, ok := property["anyOf"].([]interface{}); ok {
		var kept []interface{}
		for _, arm := range arms {
			if armMap, ok := arm.(map[string]interface{}); ok && isNullOnly(armMap) {
				continue
			}
			kept = append(kept, arm)
		}
		if len(kept) == 1 {
			if only, ok := kept[0].(map[string]interface{}); ok {
				return dropNullArm(only)
			}
		}
		property["anyOf"] = kept
		return property
	}
	if types, ok := property["type"].([]interface{}); ok {
		var kept []interface{}
		for _, t := range types {
			if t != "null" {
				kept = append(kept, t)
			}
		}
		if len(kept) == 1 {
			property["type"] = kept[0]
		} else {
			property["type"] = kept
		}
	}
	return property
}

// isNullOnly reports whether a schema arm is the bare null type reflection emits
// as the second alternative for a pointer.
func isNullOnly(arm map[string]interface{}) bool {
	if len(arm) != 1 {
		return false
	}
	return arm["type"] == "null"
}
