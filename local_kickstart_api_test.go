package schema_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

type kickstartSetFixture struct {
	RequiredNames      []string          `yaml:"required_names"`
	Sets               []kickstartSetRow `yaml:"sets"`
	StatusFields       []string          `yaml:"status_fields"`
	ForbiddenJSONNames []string          `yaml:"forbidden_json_names"`
	StrictObjects      []string          `yaml:"strict_objects"`
}

type kickstartSetRow struct {
	Name   string   `yaml:"name"`
	Values []string `yaml:"values"`
}

func TestKickstartClosedSetsMatchNeutralInventory(t *testing.T) {
	fixture := loadKickstartSetFixture(t)
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	names := make([]string, 0, len(fixture.Sets))
	for _, row := range fixture.Sets {
		names = append(names, row.Name)
		all, parse, ok := kickstartInventory(row.Name)
		if !ok {
			t.Fatalf("set %s has no production inventory", row.Name)
			continue
		}
		if !reflect.DeepEqual(all, row.Values) {
			t.Errorf("%s All* = %v, fixture wants %v", row.Name, all, row.Values)
		}
		component, ok := spec.Components.Schemas[row.Name]
		if !ok {
			t.Errorf("Types spec has no %s component", row.Name)
			continue
		}
		got := schemaEnumJSONValues(t, row.Name, component["enum"])
		want := make([]string, len(row.Values))
		for i, value := range row.Values {
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			want[i] = string(encoded)
			if err := parse(value); err != nil {
				t.Errorf("New%s(%q) refused a published value: %v", row.Name, value, err)
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s schema enum = %v, want %v", row.Name, got, want)
		}
		if err := parse("not-a-member"); err == nil {
			t.Errorf("New%s accepted not-a-member", row.Name)
		}
	}
	requireExactNames(t, fixture.RequiredNames, names)
}

func TestKickstartStatusAndPagingFields(t *testing.T) {
	fixture := loadKickstartSetFixture(t)
	status := jsonNames(reflect.TypeOf(schema.LocalKickstartStatus{}))
	wantStatus := append([]string(nil), fixture.StatusFields...)
	sort.Strings(wantStatus)
	if !reflect.DeepEqual(status, wantStatus) {
		t.Fatalf("status fields = %v, want %v", status, wantStatus)
	}
	forbidden := map[string]struct{}{}
	for _, name := range fixture.ForbiddenJSONNames {
		forbidden[name] = struct{}{}
	}
	for _, name := range fixture.StrictObjects {
		value, ok := kickstartStruct(name)
		if !ok {
			t.Fatalf("strict object %s is not a kickstart struct", name)
		}
		for _, field := range jsonNames(reflect.TypeOf(value)) {
			if _, banned := forbidden[field]; banned {
				t.Errorf("%s declares forbidden JSON field %q", name, field)
			}
		}
		raw, err := json.Marshal(map[string]bool{"notAField": true})
		if err != nil {
			t.Fatal(err)
		}
		pointer := reflect.New(reflect.TypeOf(value))
		if err := json.Unmarshal(raw, pointer.Interface()); err == nil {
			t.Errorf("%s accepted an unknown field", name)
		}
	}
}

func loadKickstartSetFixture(t *testing.T) kickstartSetFixture {
	t.Helper()
	raw, err := os.ReadFile("testdata/local-api/kickstart_sets.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var fixture kickstartSetFixture
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode kickstart set fixture: %v", err)
	}
	if len(fixture.Sets) == 0 || len(fixture.StatusFields) == 0 || len(fixture.ForbiddenJSONNames) == 0 || len(fixture.StrictObjects) == 0 {
		t.Fatal("kickstart set fixture is missing sets, status fields, forbidden names, or strict objects")
	}
	return fixture
}

func jsonNames(typ reflect.Type) []string {
	var names []string
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	sorted := append([]string(nil), names...)
	return sorted
}

func kickstartInventory(name string) ([]string, func(string) error, bool) {
	switch name {
	case "LocalKickstartScope":
		return stringTokens(schema.AllLocalKickstartScopes), func(raw string) error {
			_, err := schema.NewLocalKickstartScope(raw)
			return err
		}, true
	case "LocalKickstartStopKind":
		return stringTokens(schema.AllLocalKickstartStopKinds), func(raw string) error {
			_, err := schema.NewLocalKickstartStopKind(raw)
			return err
		}, true
	case "LocalKickstartFieldKind":
		return stringTokens(schema.AllLocalKickstartFieldKinds), func(raw string) error {
			_, err := schema.NewLocalKickstartFieldKind(raw)
			return err
		}, true
	case "LocalKickstartLineKind":
		return stringTokens(schema.AllLocalKickstartLineKinds), func(raw string) error {
			_, err := schema.NewLocalKickstartLineKind(raw)
			return err
		}, true
	case "LocalKickstartNodeKind":
		return stringTokens(schema.AllLocalKickstartNodeKinds), func(raw string) error {
			_, err := schema.NewLocalKickstartNodeKind(raw)
			return err
		}, true
	case "LocalKickstartNodeState":
		return stringTokens(schema.AllLocalKickstartNodeStates), func(raw string) error {
			_, err := schema.NewLocalKickstartNodeState(raw)
			return err
		}, true
	case "LocalKickstartScanState":
		return stringTokens(schema.AllLocalKickstartScanStates), func(raw string) error {
			_, err := schema.NewLocalKickstartScanState(raw)
			return err
		}, true
	case "LocalKickstartIngestState":
		return stringTokens(schema.AllLocalKickstartIngestStates), func(raw string) error {
			_, err := schema.NewLocalKickstartIngestState(raw)
			return err
		}, true
	case "LocalKickstartResolution":
		return stringTokens(schema.AllLocalKickstartResolutions), func(raw string) error {
			_, err := schema.NewLocalKickstartResolution(raw)
			return err
		}, true
	case "LocalKickstartNextStep":
		return stringTokens(schema.AllLocalKickstartNextSteps), func(raw string) error {
			_, err := schema.NewLocalKickstartNextStep(raw)
			return err
		}, true
	case "LocalKickstartDiffLineType":
		return stringTokens(schema.AllLocalKickstartDiffLineTypes), func(raw string) error {
			_, err := schema.NewLocalKickstartDiffLineType(raw)
			return err
		}, true
	case "LocalKickstartEventKind":
		return stringTokens(schema.AllLocalKickstartEventKinds), func(raw string) error {
			_, err := schema.NewLocalKickstartEventKind(raw)
			return err
		}, true
	case "LocalKickstartCloseReason":
		return stringTokens(schema.AllLocalKickstartCloseReasons), func(raw string) error {
			_, err := schema.NewLocalKickstartCloseReason(raw)
			return err
		}, true
	case "LocalKickstartConnectAnswer":
		return stringTokens(schema.AllLocalKickstartConnectAnswers), func(raw string) error {
			_, err := schema.NewLocalKickstartConnectAnswer(raw)
			return err
		}, true
	default:
		return nil, nil, false
	}
}

func kickstartStruct(name string) (any, bool) {
	switch name {
	case "LocalKickstartStatus":
		return schema.LocalKickstartStatus{}, true
	case "LocalKickstartOpenRequest":
		return schema.LocalKickstartOpenRequest{}, true
	case "LocalKickstartGuide":
		return schema.LocalKickstartGuide{}, true
	case "LocalKickstartOption":
		return schema.LocalKickstartOption{}, true
	case "LocalKickstartLine":
		return schema.LocalKickstartLine{}, true
	case "LocalKickstartSummary":
		return schema.LocalKickstartSummary{}, true
	case "LocalKickstartDiffLine":
		return schema.LocalKickstartDiffLine{}, true
	case "LocalKickstartHunk":
		return schema.LocalKickstartHunk{}, true
	case "LocalKickstartDiff":
		return schema.LocalKickstartDiff{}, true
	case "LocalKickstartConnect":
		return schema.LocalKickstartConnect{}, true
	case "LocalKickstartField":
		return schema.LocalKickstartField{}, true
	case "LocalKickstartStop":
		return schema.LocalKickstartStop{}, true
	case "LocalKickstartConflict":
		return schema.LocalKickstartConflict{}, true
	case "LocalKickstartScan":
		return schema.LocalKickstartScan{}, true
	case "LocalKickstartVillage":
		return schema.LocalKickstartVillage{}, true
	case "LocalKickstartIngest":
		return schema.LocalKickstartIngest{}, true
	case "LocalKickstartSession":
		return schema.LocalKickstartSession{}, true
	case "LocalKickstartNode":
		return schema.LocalKickstartNode{}, true
	case "LocalKickstartTree":
		return schema.LocalKickstartTree{}, true
	case "LocalKickstartPreview":
		return schema.LocalKickstartPreview{}, true
	case "LocalKickstartAnswers":
		return schema.LocalKickstartAnswers{}, true
	case "LocalKickstartOverwrite":
		return schema.LocalKickstartOverwrite{}, true
	case "LocalKickstartPatchResult":
		return schema.LocalKickstartPatchResult{}, true
	case "LocalKickstartCommitted":
		return schema.LocalKickstartCommitted{}, true
	case "LocalKickstartCommitRefusal":
		return schema.LocalKickstartCommitRefusal{}, true
	case "LocalKickstartEvent":
		return schema.LocalKickstartEvent{}, true
	default:
		return nil, false
	}
}

func stringTokens[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = string(value)
	}
	return out
}
