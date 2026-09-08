package schema_test

import (
	"embed"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_grouped_contract.yaml
var villageGroupedFixtureFS embed.FS

type villageRowInput struct {
	Variant        string `yaml:"variant"`
	SessionCount   *int64 `yaml:"session_count"`
	VariantCount   *int64 `yaml:"variant_count"`
	MismatchID     bool   `yaml:"mismatch_id"`
	MismatchMetric bool   `yaml:"mismatch_metric"`
}
type villageItemInput struct {
	Kind       schema.SessionListItemKind `yaml:"kind"`
	Transcript bool                       `yaml:"transcript"`
	Context    bool                       `yaml:"context"`
}
type villageRouteInput struct {
	Path        string   `yaml:"path"`
	OperationID string   `yaml:"operation_id"`
	View        bool     `yaml:"view"`
	Scope       bool     `yaml:"scope"`
	Legacy      string   `yaml:"legacy"`
	LegacyItem  string   `yaml:"legacy_item"`
	Grouped     string   `yaml:"grouped"`
	Parameters  []string `yaml:"parameters"`
}
type villageExpected struct {
	Valid bool `yaml:"valid"`
}
type villageCollectionInput struct {
	Target      string `yaml:"target"`
	Initialized bool   `yaml:"initialized"`
	Page        int    `yaml:"page"`
	Limit       int    `yaml:"limit"`
}
type villageGroupedFixtures struct {
	Rows          testcase.Corpus[villageRowInput, villageExpected]        `yaml:"rows"`
	Items         testcase.Corpus[villageItemInput, villageExpected]       `yaml:"items"`
	Routes        testcase.Corpus[villageRouteInput, villageExpected]      `yaml:"routes"`
	Collections   testcase.Corpus[villageCollectionInput, villageExpected] `yaml:"collections"`
	RequiredNames struct {
		Rows        []string `yaml:"rows"`
		Items       []string `yaml:"items"`
		Routes      []string `yaml:"routes"`
		Collections []string `yaml:"collections"`
	} `yaml:"required_names"`
}

func loadVillageGroupedFixtures(t *testing.T) villageGroupedFixtures {
	t.Helper()
	b, err := villageGroupedFixtureFS.ReadFile("testdata/village_grouped_contract.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var f villageGroupedFixtures
	if err := yaml.Unmarshal(b, &f); err != nil {
		t.Fatalf("load Village grouped contract fixtures: %v", err)
	}
	assertcase.RequireMin(t, f.Rows, 9)
	assertcase.RequireValid(t, f.Rows)
	assertcase.RequireMin(t, f.Items, 3)
	assertcase.RequireValid(t, f.Items)
	assertcase.RequireMin(t, f.Routes, 8)
	assertcase.RequireValid(t, f.Routes)
	assertcase.RequireMin(t, f.Collections, 7)
	assertcase.RequireValid(t, f.Collections)
	requireNames(t, "rows", f.RequiredNames.Rows, namesOf(f.Rows))
	requireNames(t, "items", f.RequiredNames.Items, namesOf(f.Items))
	requireNames(t, "routes", f.RequiredNames.Routes, namesOf(f.Routes))
	requireNames(t, "collections", f.RequiredNames.Collections, namesOf(f.Collections))
	return f
}

func TestVillageGroupedCollectionsAndPagination(t *testing.T) {
	f := loadVillageGroupedFixtures(t)
	for _, c := range f.Collections.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var err error
			if c.Input.Target == "list" {
				p := schema.VillageSessionListPayload{Page: c.Input.Page, Limit: c.Input.Limit}
				if c.Input.Initialized {
					p.Items = []schema.VillageSessionListItem{}
				}
				err = p.Validate()
			} else if c.Input.Target == "members" {
				p := schema.VillageHelperMembersPayload{Page: c.Input.Page, Limit: c.Input.Limit}
				if c.Input.Initialized {
					p.Members = []schema.VillageSessionRow{}
				}
				err = p.Validate()
			} else {
				p := schema.VillageTranscriptListResponse{Page: c.Input.Page, Limit: c.Input.Limit}
				if c.Input.Initialized {
					p.Transcripts = []schema.VillageTranscriptListRow{{Tags: []schema.VillageTag{}, OwnerOrgs: nil, Shares: nil, Attestations: nil}}
				}
				err = p.Validate()
			}
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v error=%v", err == nil, err)
			}
		})
	}
}

func TestVillageGroupedWrappersRoundTripCompleteValues(t *testing.T) {
	now := time.Date(2026, 9, 8, 1, 2, 3, 0, time.UTC)
	id := schema.VillageUUID("123e4567-e89b-12d3-a456-426614174001")
	list := schema.VillageSessionListPayload{Items: []schema.VillageSessionListItem{}, Page: 1, Limit: 20, TotalItems: 3, OrdinarySessionTotal: 2, HelperThreadTotal: 1}
	group := schema.VillageGroupedGroupDetailResponse{Group: schema.VillageGroup{ID: id, Name: "collective", CreatedBy: id, CreatedAt: now, UpdatedAt: now, AcceptanceMode: schema.VillageGroupAcceptanceOpen, DataAccess: schema.VillageGroupDataAccessPublic, TranscriptDeletionPolicy: schema.VillageTranscriptDeletionUserChoice}, Members: []schema.VillageGroupMember{{ID: id, GithubUsername: "owner", GithubOrgs: []string{"org"}, JoinedAt: now, Role: schema.VillageGroupRoleOwner}}, Stats: schema.VillageGroupTranscriptStats{TotalTranscripts: 3, ContributorCount: 1, TotalTurns: 5, TotalDurationMs: 8, TotalTokens: 13}, Models: []schema.VillageGroupModelBreakdown{{ModelProvider: "provider", TranscriptCount: 3}}, Contributors: []schema.VillageGroupContributor{{ID: id, GithubUsername: "owner", TranscriptCount: 3}}, CanRead: true, YourRole: schema.VillageGroupViewerRoleOwner, TranscriptList: list, PendingMembers: []schema.VillageGroupMember{{ID: id, GithubUsername: "pending", GithubOrgs: []string{}, JoinedAt: now, Role: schema.VillageGroupRolePending}}}
	contribution := schema.VillageGroupedContributableResponse{GroupID: id, TranscriptList: list}
	encoded, err := json.Marshal(group)
	if err != nil {
		t.Fatal(err)
	}
	var gotGroup schema.VillageGroupedGroupDetailResponse
	if err := json.Unmarshal(encoded, &gotGroup); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotGroup, group) {
		t.Fatalf("round trip changed grouped collective: %#v", gotGroup)
	}
	if err := gotGroup.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err = json.Marshal(contribution)
	if err != nil {
		t.Fatal(err)
	}
	var gotContribution schema.VillageGroupedContributableResponse
	if err := json.Unmarshal(encoded, &gotContribution); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotContribution, contribution) {
		t.Fatalf("round trip changed grouped contribution: %#v", gotContribution)
	}
	if err := gotContribution.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestVillageExistingTranscriptEnvelopesRoundTripCompleteValues(t *testing.T) {
	now := time.Date(2026, 9, 8, 4, 5, 6, 0, time.UTC)
	transcriptID := schema.TranscriptID("123e4567-e89b-12d3-a456-426614174000")
	ownerID := schema.VillageUUID("123e4567-e89b-12d3-a456-426614174001")
	localID := schema.SessionID("123e4567-e89b-12d3-a456-426614174002")
	title := "title"
	display := "display"
	avatar := "avatar"
	note := "note"
	transcript := schema.VillageTranscript{ID: transcriptID, OwnerID: ownerID, LocalID: localID, Title: &title, Visibility: schema.VillageTranscriptVisibilityPublic, ModelProvider: "provider", SchemaVersion: "1", PublishedAt: now, UpdatedAt: now, ProjectHash: schema.ProjectHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), ProjectDisplayName: "project", ProjectNameSource: schema.VillageProjectNameSourceOverride, ProjectRemoteLabel: "owner/repo", SessionOrigin: schema.SessionOriginUser}
	tags := []schema.VillageTag{{ID: ownerID, Name: "tag"}}
	owner := schema.VillageUser{ID: ownerID, GithubID: 42, GithubUsername: "owner", DisplayName: &display, AvatarURL: &avatar, CreatedAt: now, UpdatedAt: now, IsDiscoverable: true, Provider: "github", ProviderUserID: "42", UsernameChosen: true}
	listOrgs := []schema.VillageListUserOrganization{{UserID: ownerID, OrgLogin: "org", AvatarURL: &avatar}}
	metadataOrgs := []schema.VillageMetadataUserOrganization{{OrgLogin: "org", OrgID: 7, AvatarURL: &avatar, Visible: true, FetchedAt: now}}
	shares := []schema.VillageEnrichedTranscriptShare{{TranscriptID: transcriptID, GroupID: ownerID, GroupName: "group", AcceptanceMode: schema.VillageGroupAcceptanceOpen, Status: schema.VillageShareStatusApproved, SharedAt: now}}
	list := schema.VillageTranscriptListResponse{Transcripts: []schema.VillageTranscriptListRow{{Transcript: transcript, Tags: tags, Owner: owner, OwnerOrgs: listOrgs, Shares: shares, Attestations: []schema.VillageListTranscriptAttestation{{TranscriptID: transcriptID, OrgLogin: "org", AttestationType: "member", CreatedAt: now}}}}, Total: 1, AgentTotal: 0, Page: 1, Limit: 20}
	if err := list.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	var gotList schema.VillageTranscriptListResponse
	if err := json.Unmarshal(encoded, &gotList); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotList, list) {
		t.Fatalf("existing list envelope changed values: %#v", gotList)
	}
	metadata := schema.VillageTranscriptMetadataResponse{Transcript: transcript, Tags: tags, Shares: []schema.VillageTranscriptShare{{GroupID: ownerID, GroupName: "group", SharedAt: now}}, EnrichedShares: shares, Owner: owner, OwnerOrgs: metadataOrgs, Attestations: []schema.VillageTranscriptAttestation{{ID: ownerID, TranscriptID: transcriptID, OrgLogin: "org", AttestationType: "member", Note: &note, CreatedAt: now, AttesterUsername: "attester", AttesterAvatar: &avatar}}, RelationshipNavigation: []schema.SessionRelationshipNavigation{{Kind: schema.SessionRelationshipStartedBy, Status: schema.RelationshipNavigationKnownUnavailable}}}
	if err := metadata.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err = json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	var gotMetadata schema.VillageTranscriptMetadataResponse
	if err := json.Unmarshal(encoded, &gotMetadata); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotMetadata, metadata) {
		t.Fatalf("existing metadata envelope changed values: %#v", gotMetadata)
	}
}

func namesOf[I, E any](c testcase.Corpus[I, E]) map[string]bool {
	m := map[string]bool{}
	for _, row := range c.Cases {
		m[row.Name] = true
	}
	return m
}
func requireNames(t *testing.T, arm string, required []string, actual map[string]bool) {
	t.Helper()
	for _, name := range required {
		if !actual[name] {
			t.Errorf("Village grouped %s corpus is missing required case %q", arm, name)
		}
	}
	if len(actual) != len(required) {
		t.Errorf("Village grouped %s exact case set differs: got %d names, required %d", arm, len(actual), len(required))
	}
}

func TestVillageSessionRowsPreserveRouteIdentityAndCounts(t *testing.T) {
	f := loadVillageGroupedFixtures(t)
	for _, c := range f.Rows.Cases {
		t.Run(c.Name, func(t *testing.T) {
			row := villageRow(c.Input)
			err := row.Validate()
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v, error=%v", err == nil, err)
			}
		})
	}
}

func villageRow(in villageRowInput) schema.VillageSessionRow {
	id := schema.TranscriptID("123e4567-e89b-12d3-a456-426614174000")
	other := schema.TranscriptID("223e4567-e89b-12d3-a456-426614174000")
	owner := schema.VillageUUID("123e4567-e89b-12d3-a456-426614174001")
	local := schema.SessionID("123e4567-e89b-12d3-a456-426614174002")
	parent := schema.SessionID("123e4567-e89b-12d3-a456-426614174003")
	model := "model"
	turns := int32(5)
	tokens := int32(8)
	tokensIn := int64(3)
	tokensOut := int64(5)
	project := schema.ProjectHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	r := schema.VillageSessionRow{Session: schema.VillageTranscript{ID: id, OwnerID: owner, LocalID: local, ModelProvider: "provider", ModelName: &model, TurnCount: &turns, TokenCount: &tokens, TokensIn: &tokensIn, TokensOut: &tokensOut, ProjectHash: project, ParentSessionID: &parent, Purpose: schema.SessionPurposeInteraction, SessionOrigin: schema.SessionOriginUser, InputSubmissionCount: in.SessionCount}}
	variantID := id
	if in.MismatchID {
		variantID = other
	}
	switch in.Variant {
	case "collective":
		variantTurns := turns
		if in.MismatchMetric {
			variantTurns++
		}
		r.Collective = &schema.VillageGroupTranscript{ID: variantID, OwnerID: owner, LocalID: local, ModelProvider: "provider", ModelName: &model, TurnCount: &variantTurns, TokenCount: &tokens, TokensIn: &tokensIn, TokensOut: &tokensOut, ProjectHash: project, ParentSessionID: &parent, Purpose: schema.SessionPurposeInteraction, SessionOrigin: schema.SessionOriginUser, InputSubmissionCount: in.VariantCount}
	case "pending":
		r.Pending = &schema.VillagePendingShare{TranscriptID: variantID, OwnerID: owner, LocalID: local, ModelProvider: "provider", ProjectHash: project, ParentSessionID: &parent, Purpose: schema.SessionPurposeInteraction, InputSubmissionCount: in.VariantCount}
	case "myShare":
		r.MyShare = &schema.VillageUserGroupShare{ID: variantID, OwnerID: owner, LocalID: local, ModelProvider: "provider", ModelName: &model, TurnCount: &turns, TokensIn: &tokensIn, TokensOut: &tokensOut, ParentSessionID: &parent, Purpose: schema.SessionPurposeInteraction, InputSubmissionCount: in.VariantCount}
	case "contributable":
		r.Contributable = &schema.VillageContributableTranscript{ID: variantID, LocalID: local, ModelProvider: "provider", ProjectHash: project, ParentSessionID: &parent, Purpose: schema.SessionPurposeInteraction, SessionOrigin: schema.SessionOriginUser, InputSubmissionCount: in.VariantCount}
	}
	return r
}

func TestVillageSessionListItemDiscriminator(t *testing.T) {
	f := loadVillageGroupedFixtures(t)
	for _, c := range f.Items.Cases {
		t.Run(c.Name, func(t *testing.T) {
			item := schema.VillageSessionListItem{Kind: c.Input.Kind}
			if c.Input.Transcript {
				row := villageRow(villageRowInput{})
				item.Transcript = &row
			}
			if c.Input.Context {
				item.Context = &schema.HelperContextSummary{GroupID: "hg_fixture", OwnerStatus: schema.RelationshipNavigationUnknown}
			}
			err := item.Validate()
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v, error=%v", err == nil, err)
			}
		})
	}
}

func TestVillageGroupedRoutesPreserveLegacyAndSelectedShapes(t *testing.T) {
	f := loadVillageGroupedFixtures(t)
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(spec)
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	paths := doc["paths"].(map[string]any)
	components := doc["components"].(map[string]any)["schemas"].(map[string]any)
	for _, c := range f.Routes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			op := paths[c.Input.Path].(map[string]any)["get"].(map[string]any)
			if op["operationId"] != c.Input.OperationID {
				t.Fatalf("operationId=%v", op["operationId"])
			}
			params := parameterMap(t, op)
			if len(params) != len(c.Input.Parameters) {
				t.Fatalf("parameter count=%d, want %d: %v", len(params), len(c.Input.Parameters), params)
			}
			for _, name := range c.Input.Parameters {
				p := params[name]
				if p == nil {
					t.Errorf("missing parameter %s", name)
					continue
				}
				if p["in"] == "path" && p["required"] != true {
					t.Errorf("path parameter %s is not required", name)
				}
				typ := parameterType(p, components)
				if name == "page" || name == "limit" || name == "offset" {
					if typ != "integer" {
						t.Errorf("%s type=%s", name, typ)
					}
				} else if typ != "string" {
					t.Errorf("%s type=%s", name, typ)
				}
			}
			if view := params["view"]; view != nil {
				if view["in"] != "query" || view["required"] == true {
					t.Error("view must be an optional query parameter")
				}
				resolved := resolveParameterSchema(t, view, components)
				enum := resolved["enum"].([]any)
				if len(enum) != 1 || enum[0] != "grouped" {
					t.Errorf("view enum=%v, want [grouped]", enum)
				}
			}
			if c.Input.View && params["view"] == nil {
				t.Error("missing view query")
			}
			if c.Input.Scope && (params["scope"] == nil || params["scope"]["in"] != "query" || params["scope"]["required"] != true || parameterType(params["scope"], components) != "string") {
				t.Error("missing scope query")
			}
			responseSchema := op["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
			refs := responseRefs(responseSchema)
			if _, union := responseSchema["oneOf"]; union && len(responseSchema["oneOf"].([]any)) != 2 {
				t.Errorf("response union has %d arms, want exactly 2", len(responseSchema["oneOf"].([]any)))
			}
			if c.Input.Legacy != "" && !refs[c.Input.Legacy] {
				t.Errorf("missing legacy response arm %s in %v", c.Input.Legacy, refs)
			}
			if !refs[c.Input.Grouped] {
				t.Errorf("missing selected response arm %s in %v", c.Input.Grouped, refs)
			}
			if c.Input.LegacyItem != "" && !refs[c.Input.LegacyItem] {
				t.Errorf("legacy array item=%v, want %s", refs, c.Input.LegacyItem)
			}
			if c.Name == "metadata-navigation" {
				requireExactProperties(t, components[c.Input.Grouped].(map[string]any), []string{"transcript", "tags", "shares", "enriched_shares", "owner", "owner_orgs", "attestations", "relationshipNavigation"})
			}
			if c.Name == "durable-content" {
				if _, ok := schemaProperties(components[c.Input.Grouped].(map[string]any))["relationshipNavigation"]; ok {
					t.Error("durable content schema exposes read navigation")
				}
			}
		})
	}
	requireExactProperties(t, components["SchemaVillageGroupedGroupDetailResponse"].(map[string]any), []string{"group", "members", "stats", "models", "contributors", "can_read", "your_role", "transcriptList", "pending_members"})
	requireExactProperties(t, components["SchemaVillageGroupedContributableResponse"].(map[string]any), []string{"groupId", "transcriptList"})
	requireExactProperties(t, components["SchemaVillageMetadataUserOrganization"].(map[string]any), []string{"org_login", "org_id", "avatar_url", "visible", "fetched_at"})
	requireExactProperties(t, components["SchemaVillageListUserOrganization"].(map[string]any), []string{"user_id", "org_login", "avatar_url"})
	legacyRow := schemaProperties(components["SchemaVillageTranscriptListRow"].(map[string]any))
	for _, name := range []string{"owner_orgs", "shares", "attestations"} {
		field := legacyRow[name].(map[string]any)
		types := field["type"].([]any)
		if len(types) != 2 || types[0] != "array" || types[1] != "null" {
			t.Errorf("legacy %s type=%v, want [array null]", name, types)
		}
	}
}

func parameterMap(t *testing.T, op map[string]any) map[string]map[string]any {
	t.Helper()
	result := map[string]map[string]any{}
	for _, raw := range op["parameters"].([]any) {
		p := raw.(map[string]any)
		result[p["name"].(string)] = p
	}
	return result
}
func parameterType(p map[string]any, components map[string]any) string {
	s := resolveParameterSchema(nil, p, components)
	if typ, ok := s["type"].(string); ok {
		return typ
	}
	return ""
}
func resolveParameterSchema(t *testing.T, p map[string]any, components map[string]any) map[string]any {
	s := p["schema"].(map[string]any)
	if ref, ok := s["$ref"].(string); ok {
		name := ref[len("#/components/schemas/"):]
		resolved, ok := components[name].(map[string]any)
		if !ok && t != nil {
			t.Fatalf("parameter component %s missing", name)
		}
		return resolved
	}
	return s
}
func responseRefs(s map[string]any) map[string]bool {
	r := map[string]bool{}
	arms := []any{s}
	if one, ok := s["oneOf"].([]any); ok {
		arms = one
	}
	for _, raw := range arms {
		arm := raw.(map[string]any)
		if ref, ok := arm["$ref"].(string); ok {
			r[ref[len("#/components/schemas/"):]] = true
		}
		if items, ok := arm["items"].(map[string]any); ok {
			if ref, ok := items["$ref"].(string); ok {
				r["array"] = true
				r[ref[len("#/components/schemas/"):]] = true
			}
		}
	}
	return r
}
func schemaProperties(s map[string]any) map[string]any {
	if p, ok := s["properties"].(map[string]any); ok {
		return p
	}
	return map[string]any{}
}
func requireExactProperties(t *testing.T, schemaMap map[string]any, want []string) {
	t.Helper()
	got := schemaProperties(schemaMap)
	expected := map[string]bool{}
	for _, name := range want {
		expected[name] = true
	}
	if len(got) != len(expected) {
		t.Fatalf("property count=%d, want %d: %v", len(got), len(expected), got)
	}
	for name := range expected {
		if _, ok := got[name]; !ok {
			t.Errorf("missing property %s", name)
		}
	}
}
