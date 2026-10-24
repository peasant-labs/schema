package schema_test

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/village_public_owner.yaml
var villagePublicOwnerFixtureYAML []byte

type publicOwnerEncodeInput struct {
	Kind        schema.VillagePublicOwnerKind `yaml:"kind"`
	ID          schema.VillageUUID            `yaml:"id"`
	Username    string                        `yaml:"username"`
	DisplayName *string                       `yaml:"display_name"`
	AvatarURL   *string                       `yaml:"avatar_url"`
}

type publicOwnerEncodeExpected struct {
	Valid bool   `yaml:"valid"`
	JSON  string `yaml:"json"`
}

type publicOwnerDecodeInput struct {
	Viewer schema.VillageUUID `yaml:"viewer"`
	JSON   string             `yaml:"json"`
}

type publicOwnerDecodeExpected struct {
	Valid     bool `yaml:"valid"`
	Profile   bool `yaml:"profile"`
	IsAccount bool `yaml:"is_account"`
}

type publicOwnerResponseInput struct {
	Target    string `yaml:"target"`
	Anonymous bool   `yaml:"anonymous"`
	Orgs      int    `yaml:"orgs"`
}

type publicOwnerFixtures struct {
	Encode    testcase.Corpus[publicOwnerEncodeInput, publicOwnerEncodeExpected] `yaml:"encode"`
	Decode    testcase.Corpus[publicOwnerDecodeInput, publicOwnerDecodeExpected] `yaml:"decode"`
	Responses testcase.Corpus[publicOwnerResponseInput, villageExpected]         `yaml:"responses"`
	Surfaces  struct {
		PublicOwnerProperties []string `yaml:"public_owner_properties"`
		SelfOperations        []string `yaml:"self_operations"`
		ForbiddenProperties   []string `yaml:"forbidden_properties"`
	} `yaml:"surfaces"`
	RequiredNames struct {
		Encode    []string `yaml:"encode"`
		Decode    []string `yaml:"decode"`
		Responses []string `yaml:"responses"`
	} `yaml:"required_names"`
}

func loadPublicOwnerFixtures(t *testing.T) publicOwnerFixtures {
	t.Helper()
	var f publicOwnerFixtures
	if err := yaml.Unmarshal(villagePublicOwnerFixtureYAML, &f); err != nil {
		t.Fatalf("load public owner fixtures: %v", err)
	}
	assertcase.RequireMin(t, f.Encode, 10)
	assertcase.RequireValid(t, f.Encode)
	assertcase.RequireMin(t, f.Decode, 12)
	assertcase.RequireValid(t, f.Decode)
	requirePublicOwnerNames(t, "encode", f.RequiredNames.Encode, namesOf(f.Encode))
	requirePublicOwnerNames(t, "decode", f.RequiredNames.Decode, namesOf(f.Decode))
	assertcase.RequireMin(t, f.Responses, 9)
	assertcase.RequireValid(t, f.Responses)
	requirePublicOwnerNames(t, "responses", f.RequiredNames.Responses, namesOf(f.Responses))
	if len(f.Surfaces.PublicOwnerProperties) == 0 || len(f.Surfaces.SelfOperations) == 0 || len(f.Surfaces.ForbiddenProperties) == 0 {
		t.Fatal("public owner surfaces fixture is empty; the surface checks would assert nothing")
	}
	return f
}

func requirePublicOwnerNames(t *testing.T, arm string, required []string, actual map[string]bool) {
	t.Helper()
	assertSameStrings(t, "public owner "+arm+" case names", required, mapKeys(actual))
}

func TestVillagePublicOwnerEncodesExactFieldSets(t *testing.T) {
	f := loadPublicOwnerFixtures(t)
	for _, c := range f.Encode.Cases {
		t.Run(c.Name, func(t *testing.T) {
			owner := schema.VillagePublicOwner{Kind: c.Input.Kind, ID: c.Input.ID, Username: c.Input.Username, DisplayName: c.Input.DisplayName, AvatarURL: c.Input.AvatarURL}
			validateErr := owner.Validate()
			encoded, marshalErr := json.Marshal(owner)
			if (validateErr == nil) != c.Expected.Valid || (marshalErr == nil) != c.Expected.Valid {
				t.Fatalf("validate error=%v marshal error=%v, want valid=%v", validateErr, marshalErr, c.Expected.Valid)
			}
			if !c.Expected.Valid {
				return
			}
			if string(encoded) != c.Expected.JSON {
				t.Fatalf("wire bytes\n got %s\nwant %s", encoded, c.Expected.JSON)
			}
			var decoded schema.VillagePublicOwner
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("decode own encoding: %v", err)
			}
			if again, _ := json.Marshal(decoded); string(again) != c.Expected.JSON {
				t.Fatalf("round trip changed the wire bytes: %s", again)
			}
		})
	}
}

func TestVillagePublicOwnerReadsEveryKindFailClosed(t *testing.T) {
	f := loadPublicOwnerFixtures(t)
	for _, c := range f.Decode.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var owner schema.VillagePublicOwner
			err := json.Unmarshal([]byte(c.Input.JSON), &owner)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("decode error=%v, want valid=%v", err, c.Expected.Valid)
			}
			if !c.Expected.Valid {
				return
			}
			if owner.IsProfile() != c.Expected.Profile || owner.IsAccount(c.Input.Viewer) != c.Expected.IsAccount {
				t.Fatalf("profile=%v is_account=%v, want %v and %v", owner.IsProfile(), owner.IsAccount(c.Input.Viewer), c.Expected.Profile, c.Expected.IsAccount)
			}
			if !owner.IsProfile() && owner != schema.AnonymousVillageOwner() {
				t.Fatalf("a person not read as a profile must decode to exactly the anonymous value, got %+v", owner)
			}
			if !owner.IsProfile() && owner.Kind != schema.VillagePublicOwnerAnonymous {
				t.Fatalf("kind=%q, want anonymous", owner.Kind)
			}
			if _, err := json.Marshal(owner); err != nil {
				t.Fatalf("re-encoding a decoded value failed: %v", err)
			}
		})
	}
}

// TestVillagePublicOwnerResponsesKeepAnonymousPeopleAnonymous checks the
// whole-response rules: no organization beside an anonymous person, and no
// unresolved person in a collective detail.
func TestVillagePublicOwnerResponsesKeepAnonymousPeopleAnonymous(t *testing.T) {
	f := loadPublicOwnerFixtures(t)
	for _, c := range f.Responses.Cases {
		t.Run(c.Name, func(t *testing.T) {
			person := schema.NewVillagePublicProfile("123e4567-e89b-12d3-a456-426614174001", "alice", nil, nil)
			if c.Input.Anonymous {
				person = schema.AnonymousVillageOwner()
			}
			orgs := make([]string, c.Input.Orgs)
			for i := range orgs {
				orgs[i] = "acme"
			}
			var err error
			switch c.Input.Target {
			case "list_row":
				listOrgs := make([]schema.VillageListUserOrganization, len(orgs))
				for i := range orgs {
					listOrgs[i] = schema.VillageListUserOrganization{OrgLogin: orgs[i]}
				}
				err = schema.VillageTranscriptListResponse{Transcripts: []schema.VillageTranscriptListRow{{Tags: []schema.VillageTag{}, Owner: person, OwnerOrgs: listOrgs, PullRequests: schema.VillagePullRequestsSummary{Recent: []schema.VillagePullRequestRef{}}}}, Page: 1, Limit: 1, Total: 1}.Validate()
			case "metadata":
				metadataOrgs := make([]schema.VillageMetadataUserOrganization, len(orgs))
				for i := range orgs {
					metadataOrgs[i] = schema.VillageMetadataUserOrganization{OrgLogin: orgs[i]}
				}
				err = schema.VillageTranscriptMetadataResponse{Tags: []schema.VillageTag{}, Shares: []schema.VillageTranscriptShare{}, EnrichedShares: []schema.VillageEnrichedTranscriptShare{}, Owner: person, OwnerOrgs: metadataOrgs, Attestations: []schema.VillageTranscriptAttestation{}}.Validate()
			case "grouped_detail":
				err = schema.VillageGroupedCollectiveDetailResponse{Group: schema.VillageCollectiveDetailRecord{Creator: schema.AnonymousVillageOwner()}, Members: []schema.VillageCollectiveMember{{Member: person, GithubOrgs: orgs}}, Models: []schema.VillageCollectiveModelBreakdown{}, Contributors: []schema.VillageCollectiveContributor{}, TranscriptList: schema.VillageSessionListPayload{Items: []schema.VillageSessionListItem{}, Page: 1, Limit: 1}}.Validate()
			case "flat_detail":
				err = schema.VillageCollectiveDetailResponse{Members: []schema.VillageCollectiveMember{{Member: person, GithubOrgs: orgs}}}.Validate()
			case "flat_detail_transcript":
				err = schema.VillageCollectiveDetailResponse{Members: []schema.VillageCollectiveMember{{Member: person, GithubOrgs: orgs}}, Transcripts: []schema.VillageCollectiveTranscript{{}}}.Validate()
			default:
				t.Fatalf("unknown response target %q", c.Input.Target)
			}
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v error=%v", err == nil, err)
			}
		})
	}
}

// TestVillagePublicOwnerSurfaces walks the served Village API: every property
// that holds a person is a VillagePublicOwner, the complete VillageUser is
// reachable only from the caller's own record, and no response another viewer
// can receive declares an owner id, handle, discoverable flag, or sign-in
// account field.
func TestVillagePublicOwnerSurfaces(t *testing.T) {
	f := loadPublicOwnerFixtures(t)
	spec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	components := doc["components"].(map[string]any)["schemas"].(map[string]any)

	var holders []string
	for name, component := range components {
		for property, value := range schemaProperties(component.(map[string]any)) {
			if refs := schemaRefs(value); refs["SchemaVillagePublicOwner"] {
				holders = append(holders, name+"."+property)
			}
		}
	}
	assertSameStrings(t, "properties holding a VillagePublicOwner", f.Surfaces.PublicOwnerProperties, holders)

	self := map[string]bool{}
	for _, id := range f.Surfaces.SelfOperations {
		self[id] = true
	}
	forbidden := map[string]bool{}
	for _, name := range f.Surfaces.ForbiddenProperties {
		forbidden[name] = true
	}
	selfReachesUser := false
	for path, item := range doc["paths"].(map[string]any) {
		for method, raw := range item.(map[string]any) {
			operation, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			id, _ := operation["operationId"].(string)
			reached := reachableComponents(components, schemaRefs(operation["responses"]))
			if self[id] {
				selfReachesUser = selfReachesUser || reached["SchemaVillageUser"]
				continue
			}
			if reached["SchemaVillageUser"] {
				t.Errorf("%s %s (%s) reaches the complete VillageUser; another viewer must receive a VillagePublicOwner", strings.ToUpper(method), path, id)
			}
			for name := range reached {
				for property := range declaredPropertyNames(components[name]) {
					if forbidden[property] {
						t.Errorf("%s %s (%s) reaches %s.%s, which identifies a person outside VillagePublicOwner", strings.ToUpper(method), path, id, name, property)
					}
				}
			}
		}
	}
	if !selfReachesUser {
		t.Fatal("no self operation reaches VillageUser; the self exemption is stale and the surface check is vacuous")
	}
}

func reachableComponents(components map[string]any, start map[string]bool) map[string]bool {
	reached := map[string]bool{}
	queue := mapKeys(start)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if reached[name] {
			continue
		}
		reached[name] = true
		if component, ok := components[name]; ok {
			queue = append(queue, mapKeys(schemaRefs(component))...)
		}
	}
	return reached
}

// declaredPropertyNames lists every property a component declares, including
// those inside allOf, oneOf, anyOf, items, and inline objects; referenced
// components are walked separately.
func declaredPropertyNames(value any) map[string]bool {
	names := map[string]bool{}
	var walk func(any)
	walk = func(v any) {
		switch typed := v.(type) {
		case map[string]any:
			for key, child := range typed {
				if properties, ok := child.(map[string]any); ok && key == "properties" {
					for name, property := range properties {
						names[name] = true
						walk(property)
					}
					continue
				}
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(value)
	return names
}

func schemaRefs(value any) map[string]bool {
	refs := map[string]bool{}
	var walk func(any)
	walk = func(v any) {
		switch typed := v.(type) {
		case map[string]any:
			for key, child := range typed {
				if ref, ok := child.(string); ok && key == "$ref" {
					refs[strings.TrimPrefix(ref, "#/components/schemas/")] = true
					continue
				}
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(value)
	return refs
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func assertSameStrings(t *testing.T, label string, want, got []string) {
	t.Helper()
	wantSet, gotSet := map[string]bool{}, map[string]bool{}
	for _, value := range want {
		wantSet[value] = true
	}
	for _, value := range got {
		gotSet[value] = true
	}
	for value := range wantSet {
		if !gotSet[value] {
			t.Errorf("%s: %q is required but missing", label, value)
		}
	}
	for value := range gotSet {
		if !wantSet[value] {
			t.Errorf("%s: %q is present but not in the fixture", label, value)
		}
	}
}
