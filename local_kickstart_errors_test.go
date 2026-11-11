package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/kickstart_errors.yaml
var kickstartErrorsYAML []byte

type kickstartErrorsFixture struct {
	RequiredRoutes []string              `yaml:"required_routes"`
	Routes         []kickstartErrorRoute `yaml:"routes"`
	CodeStatus     map[string]int        `yaml:"code_status"`
}

type kickstartErrorRoute struct {
	OperationID   string              `yaml:"operation_id"`
	Method        string              `yaml:"method"`
	Path          string              `yaml:"path"`
	Errors        map[string][]string `yaml:"errors"`
	OtherStatuses []string            `yaml:"other_statuses"`
}

func loadKickstartErrors(t *testing.T) kickstartErrorsFixture {
	t.Helper()
	var fixture kickstartErrorsFixture
	decoder := yaml.NewDecoder(bytes.NewReader(kickstartErrorsYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode kickstart error fixture: %v", err)
	}
	if len(fixture.Routes) == 0 || len(fixture.CodeStatus) == 0 {
		t.Fatal("kickstart error fixture names no route or no code status")
	}
	return fixture
}

// The closed code set, the status each code answers with, and the routes are
// three views of one fact. Each is pinned independently so none can drift from
// the others.
func TestKickstartErrorCodeSetMatchesFixtureStatuses(t *testing.T) {
	fixture := loadKickstartErrors(t)
	set := make([]string, 0, len(schema.AllLocalKickstartErrorCodes))
	for _, code := range schema.AllLocalKickstartErrorCodes {
		set = append(set, code.String())
	}
	declared := make([]string, 0, len(fixture.CodeStatus))
	for code := range fixture.CodeStatus {
		declared = append(declared, code)
	}
	requireSameSet(t, "kickstart error codes and the fixture's code_status", set, declared)

	used := map[string]bool{}
	for _, route := range fixture.Routes {
		for status, codes := range route.Errors {
			want, err := strconv.Atoi(status)
			if err != nil {
				t.Fatalf("%s names status %q, which is not a number", route.OperationID, status)
			}
			for _, code := range codes {
				if _, err := schema.NewLocalKickstartErrorCode(code); err != nil {
					t.Errorf("%s status %s names code %q outside the closed set: %v", route.OperationID, status, code, err)
				}
				if got := fixture.CodeStatus[code]; got != want {
					t.Errorf("%s answers %q with %d, but the code is pinned to %d; one code answers one status", route.OperationID, code, want, got)
				}
				used[code] = true
			}
		}
	}
	for _, code := range set {
		if !used[code] {
			t.Errorf("code %q is in the closed set but no kickstart route can answer it", code)
		}
	}
}

func TestKickstartRoutesDeclareTheirErrorStatusesAndCodes(t *testing.T) {
	fixture := loadKickstartErrors(t)
	spec, err := openapi.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}

	names := make([]string, 0, len(fixture.Routes))
	pinned := map[string]bool{}
	for _, route := range fixture.Routes {
		names = append(names, route.OperationID)
		pinned[route.Method+" "+route.Path] = true
	}
	requireExactNames(t, fixture.RequiredRoutes, names)

	// Every kickstart operation in the document is pinned, so a route added
	// without its error table fails here.
	var declared []string
	for path, rawItem := range document["paths"].(map[string]any) {
		if !strings.HasPrefix(path, "/api/v1/kickstart/") {
			continue
		}
		for method := range rawItem.(map[string]any) {
			declared = append(declared, method+" "+path)
		}
	}
	var pinnedKeys []string
	for key := range pinned {
		pinnedKeys = append(pinnedKeys, key)
	}
	requireSameSet(t, "kickstart operations in the Local API document and the error fixture", declared, pinnedKeys)

	for _, route := range fixture.Routes {
		t.Run(route.OperationID, func(t *testing.T) {
			item, _ := document["paths"].(map[string]any)[route.Path].(map[string]any)
			operation, ok := item[route.Method].(map[string]any)
			if !ok {
				t.Fatalf("%s %s is not declared", route.Method, route.Path)
			}
			if got := fmt.Sprint(operation["operationId"]); got != route.OperationID {
				t.Fatalf("operationId = %s, want %s", got, route.OperationID)
			}
			responses, _ := operation["responses"].(map[string]any)
			var failures []string
			for status := range responses {
				if !strings.HasPrefix(status, "2") {
					failures = append(failures, status)
				}
			}
			var want []string
			for status := range route.Errors {
				want = append(want, status)
			}
			want = append(want, route.OtherStatuses...)
			requireSameSet(t, route.OperationID+" non-2xx statuses", want, failures)

			for status, codes := range route.Errors {
				response, _ := responses[status].(map[string]any)
				body := mediaSchema(t, route.OperationID+" "+status, response)
				ref, _ := body["$ref"].(string)
				name := strings.TrimPrefix(ref, "#/components/schemas/")
				if name != "LocalKickstartErrorResponse" && name != "SchemaLocalKickstartErrorResponse" {
					t.Errorf("%s status %s body references %q, want the closed LocalKickstartErrorResponse", route.OperationID, status, ref)
				}
				description, _ := response["description"].(string)
				listed := description[strings.LastIndex(description, "one of: ")+len("one of: "):]
				if got := strings.Split(strings.TrimSuffix(listed, "."), ", "); !reflect.DeepEqual(got, codes) {
					t.Errorf("%s status %s describes codes %v, fixture wants %v", route.OperationID, status, got, codes)
				}
			}
		})
	}
}

// The error envelope names its code through the closed set, so a client that
// reads the generated component parses the token and never compares strings.
func TestKickstartErrorResponseComponentBindsTheClosedCode(t *testing.T) {
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	component, ok := spec.Components.Schemas["LocalKickstartErrorResponse"]
	if !ok {
		t.Fatal("Types spec has no LocalKickstartErrorResponse component")
	}
	properties, _ := component["properties"].(map[string]any)
	code, _ := properties["code"].(map[string]any)
	if ref, _ := code["$ref"].(string); !strings.HasSuffix(ref, "/LocalKickstartErrorCode") {
		t.Fatalf("code property = %v, want a $ref to LocalKickstartErrorCode", code)
	}
	required := anyStrings(component["required"])
	sort.Strings(required)
	if want := []string{"code", "error"}; !reflect.DeepEqual(required, want) {
		t.Fatalf("required = %v, want %v", required, want)
	}
	if component["additionalProperties"] != false {
		t.Fatalf("additionalProperties = %v, want false: kickstart bodies are closed", component["additionalProperties"])
	}
}

// LocalKickstartDroppedLabelsMax is the exported bound; the struct tag and the
// generated schema must be the same number, because a Go struct tag cannot read
// a constant.
func TestKickstartDroppedLabelsMaxIsTheDeclaredBound(t *testing.T) {
	fixture := loadKickstartSetFixture(t)
	if fixture.DroppedLabelsMax != schema.LocalKickstartDroppedLabelsMax {
		t.Fatalf("schema.LocalKickstartDroppedLabelsMax = %d, fixture pins %d", schema.LocalKickstartDroppedLabelsMax, fixture.DroppedLabelsMax)
	}
	field, ok := reflect.TypeOf(schema.LocalKickstartScan{}).FieldByName("DroppedLabels")
	if !ok {
		t.Fatal("LocalKickstartScan has no DroppedLabels field")
	}
	if got := field.Tag.Get("maxItems"); got != strconv.Itoa(schema.LocalKickstartDroppedLabelsMax) {
		t.Fatalf("DroppedLabels maxItems tag = %q, want %d", got, schema.LocalKickstartDroppedLabelsMax)
	}
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	properties, _ := spec.Components.Schemas["LocalKickstartScan"]["properties"].(map[string]any)
	labels, _ := properties["droppedLabels"].(map[string]any)
	if got := fmt.Sprint(labels["maxItems"]); got != strconv.Itoa(schema.LocalKickstartDroppedLabelsMax) {
		t.Fatalf("generated droppedLabels maxItems = %s, want %d", got, schema.LocalKickstartDroppedLabelsMax)
	}
	over := schema.LocalKickstartScan{State: schema.LocalKickstartScanDone, DroppedLabels: make([]string, schema.LocalKickstartDroppedLabelsMax+1)}
	if err := over.Validate(); err == nil {
		t.Fatal("Validate accepted one label more than the exported maximum")
	}
	at := schema.LocalKickstartScan{State: schema.LocalKickstartScanDone, DroppedLabels: make([]string, schema.LocalKickstartDroppedLabelsMax)}
	if err := at.Validate(); err != nil {
		t.Fatalf("Validate refused exactly the exported maximum: %v", err)
	}
}

// The partial-scan notice is its own declared field, never smuggled into phase.
func TestKickstartScanDeclaresTheNoticeBesidePhase(t *testing.T) {
	fixture := loadKickstartSetFixture(t)
	names := jsonNames(reflect.TypeOf(schema.LocalKickstartScan{}))
	found := false
	for _, name := range names {
		if name == fixture.ScanNoticeField {
			found = true
		}
	}
	if !found {
		t.Fatalf("LocalKickstartScan JSON fields %v do not include the notice field %q", names, fixture.ScanNoticeField)
	}
}
