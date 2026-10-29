package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	schemaopenapi "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

type publicationVerdict struct {
	GoValid               *bool  `yaml:"go_valid"`
	ZodValid              *bool  `yaml:"zod_valid"`
	Fingerprint           string `yaml:"fingerprint,omitempty"`
	FingerprintApplicable *bool  `yaml:"fingerprint_applicable,omitempty"`
}
type publicationCorpus struct {
	Metadata               testcase.Corpus[string, publicationVerdict]                  `yaml:"metadata"`
	OwnerUpdates           testcase.Corpus[string, publicationVerdict]                  `yaml:"owner_updates"`
	OwnerUpdateResponses   testcase.Corpus[string, publicationVerdict]                  `yaml:"owner_update_responses"`
	VisibilityIntents      testcase.Corpus[schema.VisibilityIntent, publicationVerdict] `yaml:"visibility_intents"`
	ClientInfos            testcase.Corpus[string, publicationVerdict]                  `yaml:"client_infos"`
	Operations             testcase.Corpus[string, publicationVerdict]                  `yaml:"operations"`
	Responses              testcase.Corpus[string, publicationVerdict]                  `yaml:"responses"`
	FingerprintMutations   fingerprintMutationFixture                                   `yaml:"fingerprint_mutations"`
	NestedUnknowns         nestedUnknownFixture                                         `yaml:"nested_unknowns"`
	StrictComponents       []string                                                     `yaml:"strict_components"`
	StrictSchemaComponents []string                                                     `yaml:"strict_schema_components"`
	ParentIdentity         testcase.Corpus[string, parentIdentityVerdict]               `yaml:"parent_identity"`
	HistoricalIdentityJSON testcase.Corpus[string, string]                              `yaml:"historical_identity_json"`
}
type parentIdentityVerdict struct {
	Scenario    parentIdentityScenario `yaml:"scenario"`
	GoValid     bool                   `yaml:"go_valid"`
	ZodValid    *bool                  `yaml:"zod_valid,omitempty"`
	Parent      string                 `yaml:"parent,omitempty"`
	Fingerprint string                 `yaml:"fingerprint,omitempty"`
}
type parentIdentityScenario string

const (
	parentScenarioRoot       parentIdentityScenario = "root"
	parentScenarioValid      parentIdentityScenario = "valid-parent"
	parentScenarioNull       parentIdentityScenario = "null"
	parentScenarioHistorical parentIdentityScenario = "historical-key"
	parentScenarioWrongType  parentIdentityScenario = "wrong-type"
	parentScenarioDuplicate  parentIdentityScenario = "duplicate-key"
)

type nestedUnknownFixture struct {
	MetadataBase  string `yaml:"metadata_base"`
	OperationBase string `yaml:"operation_base"`
	// MetadataOnly names strict components that exist only on the metadata arm.
	MetadataOnly []string            `yaml:"metadata_only_components"`
	Cases        []nestedUnknownCase `yaml:"cases"`
}
type nestedUnknownCase struct {
	Name     string `yaml:"name"`
	Arm      string `yaml:"arm"`
	Path     string `yaml:"path"`
	Value    string `yaml:"value,omitempty"`
	GoValid  *bool  `yaml:"go_valid"`
	ZodValid *bool  `yaml:"zod_valid"`
}
type fingerprintMutationFixture struct {
	Base  string                    `yaml:"base"`
	Cases []fingerprintMutationCase `yaml:"cases"`
}
type fingerprintMutationCase struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	Value       string `yaml:"value"`
	Fingerprint string `yaml:"fingerprint"`
}

//go:embed testdata/publication/contract.yaml
var publicationCorpusYAML []byte

func loadPublicationCorpus(t *testing.T) publicationCorpus {
	t.Helper()
	var corpus publicationCorpus
	decoder := yaml.NewDecoder(bytes.NewReader(publicationCorpusYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&corpus); err != nil {
		t.Fatalf("load publication corpus: %v", err)
	}
	assert.RequireMin(t, corpus.Operations, 2)
	assert.RequireValid(t, corpus.Operations)
	for _, row := range corpus.Operations.Cases {
		if row.Expected.FingerprintApplicable == nil {
			t.Fatalf("operations case %q must explicitly declare fingerprint_applicable", row.Name)
		}
		if *row.Expected.FingerprintApplicable != (row.Expected.Fingerprint != "") {
			t.Fatalf("operations case %q fingerprint applicability and known answer disagree", row.Name)
		}
	}
	assert.RequireMin(t, corpus.Responses, 4)
	assert.RequireValid(t, corpus.Responses)
	assert.RequireMin(t, corpus.Metadata, 5)
	assert.RequireValid(t, corpus.Metadata)
	assert.RequireMin(t, corpus.OwnerUpdates, 5)
	assert.RequireValid(t, corpus.OwnerUpdates)
	assert.RequireMin(t, corpus.OwnerUpdateResponses, 9)
	assert.RequireValid(t, corpus.OwnerUpdateResponses)
	assert.RequireMin(t, corpus.VisibilityIntents, 3)
	assert.RequireValid(t, corpus.VisibilityIntents)
	assert.RequireMin(t, corpus.ClientInfos, 3)
	assert.RequireValid(t, corpus.ClientInfos)
	assert.RequireMin(t, corpus.ParentIdentity, 6)
	assert.RequireValid(t, corpus.ParentIdentity)
	requireParentIdentityCoverage(t, corpus.ParentIdentity)
	assert.RequireMin(t, corpus.HistoricalIdentityJSON, 2)
	assert.RequireValid(t, corpus.HistoricalIdentityJSON)
	if corpus.FingerprintMutations.Base == "" || len(corpus.FingerprintMutations.Cases) < 12 {
		t.Fatal("fingerprint_mutations must provide a base and at least twelve classified semantic mutations")
	}
	seenMutationNames := map[string]struct{}{}
	for _, row := range corpus.FingerprintMutations.Cases {
		if row.Name == "" || row.Path == "" || row.Value == "" || row.Fingerprint == "" {
			t.Fatalf("fingerprint mutation must explicitly declare name, path, value, and fingerprint: %+v", row)
		}
		if _, exists := seenMutationNames[row.Name]; exists {
			t.Fatalf("duplicate fingerprint mutation %q", row.Name)
		}
		seenMutationNames[row.Name] = struct{}{}
	}
	for arm, rows := range map[string][]testcase.Case[string, publicationVerdict]{
		"metadata": corpus.Metadata.Cases, "owner_updates": corpus.OwnerUpdates.Cases,
		"owner_update_responses": corpus.OwnerUpdateResponses.Cases, "operations": corpus.Operations.Cases,
		"responses": corpus.Responses.Cases,
	} {
		for _, row := range rows {
			if row.Expected.GoValid == nil || row.Expected.ZodValid == nil {
				t.Fatalf("%s case %q must explicitly declare go_valid and zod_valid", arm, row.Name)
			}
		}
	}
	for _, row := range corpus.VisibilityIntents.Cases {
		if row.Expected.GoValid == nil || row.Expected.ZodValid == nil {
			t.Fatalf("visibility_intents case %q must explicitly declare go_valid and zod_valid", row.Name)
		}
	}
	if want := len(corpus.StrictComponents)*2 - len(corpus.NestedUnknowns.MetadataOnly); len(corpus.NestedUnknowns.Cases) != want {
		t.Fatalf("nested unknown inventory has %d rows, want %d (two per %d strict nested components, one for each of %d metadata-only components)", len(corpus.NestedUnknowns.Cases), want, len(corpus.StrictComponents), len(corpus.NestedUnknowns.MetadataOnly))
	}
	return corpus
}

func requireParentIdentityCoverage(t *testing.T, corpus testcase.Corpus[string, parentIdentityVerdict]) {
	t.Helper()
	wantCounts := map[parentIdentityScenario]int{
		parentScenarioRoot: 1, parentScenarioValid: 2, parentScenarioNull: 1,
		parentScenarioHistorical: 1, parentScenarioWrongType: 1, parentScenarioDuplicate: 1,
	}
	gotCounts := make(map[parentIdentityScenario]int, len(wantCounts))
	parents := map[string]string{}
	fingerprints := map[string]string{}
	for _, row := range corpus.Cases {
		if _, known := wantCounts[row.Expected.Scenario]; !known {
			t.Fatalf("parent_identity case %q has unknown scenario %q", row.Name, row.Expected.Scenario)
		}
		gotCounts[row.Expected.Scenario]++
		if row.Expected.Scenario == parentScenarioDuplicate {
			if row.Expected.ZodValid != nil {
				t.Fatalf("parent_identity duplicate case %q must be raw-Go-only", row.Name)
			}
		} else if row.Expected.ZodValid == nil {
			t.Fatalf("parent_identity case %q must declare zod_valid", row.Name)
		}
		if row.Expected.Scenario == parentScenarioValid {
			if row.Expected.Parent == "" || row.Expected.Fingerprint == "" {
				t.Fatalf("valid parent case %q must pin parent and fingerprint", row.Name)
			}
			if prior, exists := parents[row.Expected.Parent]; exists {
				t.Fatalf("valid parent cases %q and %q repeat parent %q", prior, row.Name, row.Expected.Parent)
			}
			if prior, exists := fingerprints[row.Expected.Fingerprint]; exists {
				t.Fatalf("valid parent cases %q and %q repeat fingerprint %q", prior, row.Name, row.Expected.Fingerprint)
			}
			parents[row.Expected.Parent], fingerprints[row.Expected.Fingerprint] = row.Name, row.Name
		}
	}
	for scenario, want := range wantCounts {
		if gotCounts[scenario] != want {
			t.Fatalf("parent_identity scenario %q count=%d want %d", scenario, gotCounts[scenario], want)
		}
	}
}

func TestPublicationContractCorpus_ParentIdentityDecodeAndFingerprint(t *testing.T) {
	corpus := loadPublicationCorpus(t)
	seenFingerprints := map[string]string{}
	for _, row := range corpus.ParentIdentity.Cases {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			request, err := schema.DecodeAuthoritativePublishRequest([]byte(row.Input))
			if (err == nil) != row.Expected.GoValid {
				t.Fatalf("Go validity=%v want %v: %v", err == nil, row.Expected.GoValid, err)
			}
			if err != nil {
				return
			}
			if row.Expected.Parent == "" {
				if request.Identity.ParentSessionID != nil {
					t.Fatalf("decoded parent=%q want nil", request.Identity.ParentSessionID.String())
				}
			} else if request.Identity.ParentSessionID == nil || request.Identity.ParentSessionID.String() != row.Expected.Parent {
				t.Fatalf("decoded parent=%v want %q", request.Identity.ParentSessionID, row.Expected.Parent)
			}
			operation, err := schema.CanonicalizePublishRequest(request)
			if err != nil {
				t.Fatalf("canonicalize decoded request: %v", err)
			}
			fingerprint, err := schema.FingerprintPublishOperation(operation)
			if err != nil {
				t.Fatalf("fingerprint decoded request: %v", err)
			}
			if fingerprint.String() != row.Expected.Fingerprint {
				t.Fatalf("fingerprint=%s want %s", fingerprint, row.Expected.Fingerprint)
			}
			second, err := schema.FingerprintPublishOperation(operation)
			if err != nil || second != fingerprint {
				t.Fatalf("repeat fingerprint=%s error=%v want deterministic %s", second, err, fingerprint)
			}
			if prior, exists := seenFingerprints[fingerprint.String()]; exists {
				t.Fatalf("fingerprint collides with %q", prior)
			}
			seenFingerprints[fingerprint.String()] = row.Name
		})
	}
}

func TestPublicationContractCorpus_HistoricalParentUUIDJSONRemainsStable(t *testing.T) {
	corpus := loadPublicationCorpus(t)
	for _, row := range corpus.HistoricalIdentityJSON.Cases {
		parent, err := schema.NewSessionID(row.Input)
		if err != nil {
			t.Fatalf("%s fixture parent: %v", row.Name, err)
		}
		encoded, err := json.Marshal(schema.SessionIdentity{ParentSessionID: &parent})
		if err != nil {
			t.Fatalf("%s marshal historical identity: %v", row.Name, err)
		}
		if string(encoded) != row.Expected {
			t.Fatalf("%s JSON=%s want %s", row.Name, encoded, row.Expected)
		}
	}
}

func TestPublicationContract_GeneratedIdentityKeysRemainSeparated(t *testing.T) {
	typesSpec, err := schemaopenapi.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	villageSpec, err := schemaopenapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("build Village spec: %v", err)
	}
	assertIdentityKeys(t, typesSpec, "AuthoritativeSessionIdentity", "SessionIdentity")
	assertIdentityKeys(t, villageSpec, "SchemaAuthoritativeSessionIdentity", "")
}

func assertIdentityKeys(t *testing.T, spec any, authoritativeName, historicalName string) {
	t.Helper()
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal generated spec: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode generated spec: %v", err)
	}
	schemas := document["components"].(map[string]any)["schemas"].(map[string]any)
	authoritative := schemas[authoritativeName].(map[string]any)["properties"].(map[string]any)
	if _, ok := authoritative["parentSessionId"]; !ok {
		t.Fatalf("%s omits parentSessionId", authoritativeName)
	}
	if _, ok := authoritative["parentUuid"]; ok {
		t.Fatalf("%s exposes historical parentUuid", authoritativeName)
	}
	if historicalName != "" {
		historical := schemas[historicalName].(map[string]any)["properties"].(map[string]any)
		if _, ok := historical["parentUuid"]; !ok {
			t.Fatalf("%s omits historical parentUuid", historicalName)
		}
	}
}

func TestPublicationContractCorpus_RejectsEveryNestedUnknownField(t *testing.T) {
	corpus := loadPublicationCorpus(t)
	seen := map[string]map[string]struct{}{"metadata": {}, "operation": {}}
	for _, row := range corpus.NestedUnknowns.Cases {
		if row.GoValid == nil || row.ZodValid == nil {
			t.Fatalf("nested unknown %q must declare both verdicts", row.Name)
		}
		base := corpus.NestedUnknowns.MetadataBase
		if row.Arm == "operation" {
			base = corpus.NestedUnknowns.OperationBase
		} else if row.Arm != "metadata" {
			t.Fatalf("nested unknown %q has invalid arm %q", row.Name, row.Arm)
		}
		value := row.Value
		if value == "" {
			value = "true"
		}
		data := mutatePublicationJSON(t, base, row.Path, value)
		var err error
		if row.Arm == "metadata" {
			_, err = schema.DecodeAuthoritativePublishRequest(data)
		} else {
			_, err = schema.DecodeCanonicalPublishOperation(data)
		}
		if (err == nil) != *row.GoValid {
			t.Fatalf("nested unknown %q Go validity=%v want %v: %v", row.Name, err == nil, *row.GoValid, err)
		}
		component := strings.Split(row.Name, "/")[0]
		seen[row.Arm][component] = struct{}{}
	}
	metadataOnly := map[string]struct{}{}
	for _, component := range corpus.NestedUnknowns.MetadataOnly {
		metadataOnly[component] = struct{}{}
	}
	for _, component := range corpus.StrictComponents {
		for arm, covered := range seen {
			if _, only := metadataOnly[component]; only && arm == "operation" {
				if _, ok := covered[component]; ok {
					t.Fatalf("metadata-only component %q has operation nested-unknown evidence", component)
				}
				continue
			}
			if _, ok := covered[component]; !ok {
				t.Fatalf("strict component %q lacks %s nested-unknown evidence", component, arm)
			}
		}
	}
}

func TestPublicationContractCorpus_OpenAPISuccessorNestedComponentsAreClosed(t *testing.T) {
	corpus := loadPublicationCorpus(t)
	spec, err := schemaopenapi.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal Types spec: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode Types spec: %v", err)
	}
	schemas := document["components"].(map[string]any)["schemas"].(map[string]any)
	seen := map[string]struct{}{}
	for _, name := range corpus.StrictSchemaComponents {
		if _, duplicate := seen[name]; duplicate {
			t.Fatalf("strict schema component %q is duplicated", name)
		}
		seen[name] = struct{}{}
		component, ok := schemas[name].(map[string]any)
		if !ok {
			t.Fatalf("strict schema component %q is absent", name)
		}
		if closed, ok := component["additionalProperties"].(bool); !ok || closed {
			t.Fatalf("strict schema component %q additionalProperties=%v, want false", name, component["additionalProperties"])
		}
	}
}

func TestPublicationContractCorpus_FingerprintFieldsAreLoadBearing(t *testing.T) {
	corpus := loadPublicationCorpus(t)
	seen := map[string]string{}
	for _, row := range corpus.FingerprintMutations.Cases {
		data := mutatePublicationJSON(t, corpus.FingerprintMutations.Base, row.Path, row.Value)
		operation, err := schema.DecodeCanonicalPublishOperation(data)
		if err != nil {
			t.Fatalf("%s decode mutated operation: %v", row.Name, err)
		}
		fingerprint, err := schema.FingerprintPublishOperation(operation)
		if err != nil {
			t.Fatalf("%s fingerprint: %v", row.Name, err)
		}
		if fingerprint.String() != row.Fingerprint {
			t.Fatalf("%s fingerprint=%s want %s", row.Name, fingerprint, row.Fingerprint)
		}
		if prior, exists := seen[fingerprint.String()]; exists {
			t.Fatalf("mutations %q and %q collide at %s", prior, row.Name, fingerprint)
		}
		seen[fingerprint.String()] = row.Name
	}
}

func mutatePublicationJSON(t *testing.T, base, path, value string) []byte {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal([]byte(base), &document); err != nil {
		t.Fatalf("decode mutation base: %v", err)
	}
	var replacement any
	if err := json.Unmarshal([]byte(value), &replacement); err != nil {
		t.Fatalf("decode mutation %s: %v", path, err)
	}
	parts := strings.Split(path, ".")
	cursor := document
	for _, part := range parts[:len(parts)-1] {
		next, ok := cursor[part].(map[string]any)
		if !ok {
			t.Fatalf("mutation path %q has no object at %q", path, part)
		}
		cursor = next
	}
	cursor[parts[len(parts)-1]] = replacement
	out, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("encode mutation %s: %v", path, err)
	}
	return out
}

func TestPublicationContractCorpus_VisibilityIntentExcludedFromFingerprint(t *testing.T) {
	corpus := loadPublicationCorpus(t)
	operation, err := schema.DecodeCanonicalPublishOperation([]byte(corpus.Operations.Cases[0].Input))
	if err != nil {
		t.Fatalf("decode baseline operation: %v", err)
	}
	for _, row := range corpus.VisibilityIntents.Cases {
		request := schema.AuthoritativePublishRequest{
			Identity: operation.Replacement.Identity, Model: operation.Replacement.Model, Timestamp: operation.Replacement.Timestamp,
			Source: operation.Replacement.Source, Git: schema.AuthoritativeGitContext{Branch: operation.Replacement.Git.Branch, Remote: operation.Replacement.Git.Remote, Worktree: operation.Replacement.Git.Worktree, Tracking: operation.Replacement.Git.Tracking, Commits: operation.Replacement.Git.Commits, Associations: operation.Associations.Associations},
			Project: operation.Replacement.Project, Stats: operation.Replacement.Stats, Quality: operation.Replacement.Quality, Entries: operation.Replacement.Entries, Subagents: operation.Replacement.Subagents, Diagnostics: operation.Replacement.Diagnostics,
			ContentHash: operation.ContentHash, VisibilityIntent: row.Input}
		canonical, canonicalErr := schema.CanonicalizePublishRequest(request)
		if canonicalErr != nil {
			t.Fatalf("canonicalize visibility intent %q: %v", row.Input, canonicalErr)
		}
		fingerprint, fingerprintErr := schema.FingerprintPublishOperation(canonical)
		if fingerprintErr != nil {
			t.Fatalf("fingerprint visibility intent %q: %v", row.Input, fingerprintErr)
		}
		if fingerprint.String() != row.Expected.Fingerprint {
			t.Fatalf("visibility intent %q fingerprint=%s want %s", row.Input, fingerprint, row.Expected.Fingerprint)
		}
	}
}

func TestPublicationContractCorpus_ClientExcludedFromFingerprint(t *testing.T) {
	corpus := loadPublicationCorpus(t)
	operation, err := schema.DecodeCanonicalPublishOperation([]byte(corpus.Operations.Cases[0].Input))
	if err != nil {
		t.Fatalf("decode baseline operation: %v", err)
	}
	var baseline *schema.CanonicalPublishOperation
	var baselineJSON []byte
	for _, row := range corpus.ClientInfos.Cases {
		request := schema.AuthoritativePublishRequest{
			Identity: operation.Replacement.Identity, Model: operation.Replacement.Model, Timestamp: operation.Replacement.Timestamp,
			Source: operation.Replacement.Source, Git: schema.AuthoritativeGitContext{Branch: operation.Replacement.Git.Branch, Remote: operation.Replacement.Git.Remote, Worktree: operation.Replacement.Git.Worktree, Tracking: operation.Replacement.Git.Tracking, Commits: operation.Replacement.Git.Commits, Associations: operation.Associations.Associations},
			Project: operation.Replacement.Project, Stats: operation.Replacement.Stats, Quality: operation.Replacement.Quality, Entries: operation.Replacement.Entries, Subagents: operation.Replacement.Subagents, Diagnostics: operation.Replacement.Diagnostics,
			ContentHash: operation.ContentHash}
		if row.Input != "" {
			var client schema.AuthoritativeClientInfo
			if err := json.Unmarshal([]byte(row.Input), &client); err != nil {
				t.Fatalf("client_infos %q: decode client: %v", row.Name, err)
			}
			if err := client.Validate(); err != nil {
				t.Fatalf("client_infos %q: client is invalid: %v", row.Name, err)
			}
			request.Client = &client
		}
		canonical, canonicalErr := schema.CanonicalizePublishRequest(request)
		if canonicalErr != nil {
			t.Fatalf("canonicalize client %q: %v", row.Name, canonicalErr)
		}
		encoded, marshalErr := json.Marshal(canonical)
		if marshalErr != nil {
			t.Fatalf("marshal canonical operation for %q: %v", row.Name, marshalErr)
		}
		if bytes.Contains(encoded, []byte(`"client"`)) {
			t.Fatalf("client %q leaked into the canonical operation: %s", row.Name, encoded)
		}
		if baseline == nil {
			baseline, baselineJSON = &canonical, encoded
		} else {
			if !reflect.DeepEqual(*baseline, canonical) {
				t.Fatalf("client %q changed the canonical operation", row.Name)
			}
			if !bytes.Equal(baselineJSON, encoded) {
				t.Fatalf("client %q changed the canonical operation encoding", row.Name)
			}
		}
		fingerprint, fingerprintErr := schema.FingerprintPublishOperation(canonical)
		if fingerprintErr != nil {
			t.Fatalf("fingerprint client %q: %v", row.Name, fingerprintErr)
		}
		if fingerprint.String() != row.Expected.Fingerprint {
			t.Fatalf("client %q fingerprint=%s want %s", row.Name, fingerprint, row.Expected.Fingerprint)
		}
	}
	if corpus.ClientInfos.Cases[0].Input != "" {
		t.Fatal("client_infos must begin with the omitted-client baseline row")
	}
}

func TestPublicationContractCorpus_GoVerdicts(t *testing.T) {
	corpus := loadPublicationCorpus(t)
	for _, row := range corpus.Metadata.Cases {
		row := row
		t.Run("metadata/"+row.Name, func(t *testing.T) {
			_, err := schema.DecodeAuthoritativePublishRequest([]byte(row.Input))
			if (err == nil) != *row.Expected.GoValid {
				t.Fatalf("Go verdict=%v want %v: %v", err == nil, *row.Expected.GoValid, err)
			}
		})
	}
	for _, row := range corpus.OwnerUpdates.Cases {
		row := row
		t.Run("owner-update/"+row.Name, func(t *testing.T) {
			_, err := schema.DecodeOwnerTranscriptUpdateRequest([]byte(row.Input))
			if (err == nil) != *row.Expected.GoValid {
				t.Fatalf("Go verdict=%v want %v: %v", err == nil, row.Expected.GoValid, err)
			}
		})
	}
	for _, row := range corpus.OwnerUpdateResponses.Cases {
		row := row
		t.Run("owner-update-response/"+row.Name, func(t *testing.T) {
			var response schema.OwnerTranscriptUpdateResponse
			err := response.UnmarshalJSON([]byte(row.Input))
			if (err == nil) != *row.Expected.GoValid {
				t.Fatalf("Go verdict=%v want %v: %v", err == nil, *row.Expected.GoValid, err)
			}
		})
	}
	for _, row := range corpus.Operations.Cases {
		row := row
		t.Run("operation/"+row.Name, func(t *testing.T) {
			operation, err := schema.DecodeCanonicalPublishOperation([]byte(row.Input))
			if (err == nil) != *row.Expected.GoValid {
				t.Fatalf("Go verdict=%v want %v: %v", err == nil, row.Expected.GoValid, err)
			}
			if err == nil && row.Expected.Fingerprint != "" {
				fingerprint, fingerprintErr := schema.FingerprintPublishOperation(operation)
				if fingerprintErr != nil {
					t.Fatalf("fingerprint canonical operation: %v", fingerprintErr)
				}
				if fingerprint.String() != row.Expected.Fingerprint {
					t.Fatalf("fingerprint=%s want %s", fingerprint, row.Expected.Fingerprint)
				}
			}
		})
	}
	for _, row := range corpus.Responses.Cases {
		row := row
		t.Run("response/"+row.Name, func(t *testing.T) {
			_, err := schema.DecodePublishResponse([]byte(row.Input))
			if (err == nil) != *row.Expected.GoValid {
				t.Fatalf("Go verdict=%v want %v: %v", err == nil, row.Expected.GoValid, err)
			}
		})
	}
}
