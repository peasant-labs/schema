package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/route"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/route_members.yaml
var routeMembersYAML []byte

//go:embed testdata/route_members_manifest.yaml
var routeMembersManifestYAML []byte

// routeMembersAPI is the closed set of documents a fixture case reads as.
type routeMembersAPI string

const (
	routeMembersVillage routeMembersAPI = "village"
	routeMembersLocal   routeMembersAPI = "local"
)

func (a routeMembersAPI) api(t *testing.T) routeAPI {
	t.Helper()
	switch a {
	case routeMembersVillage:
		return villageRouteAPI
	case routeMembersLocal:
		return localRouteAPI
	}
	t.Fatalf("route_members.yaml names api %q; use %q or %q", a, routeMembersVillage, routeMembersLocal)
	return routeAPI{}
}

type routeMembersInput struct {
	API      routeMembersAPI `yaml:"api"`
	Document yaml.Node       `yaml:"document"`
}

type routeMembersExpected struct {
	Operations    []routeOperation `yaml:"operations"`
	ErrorContains string           `yaml:"error_contains"`
}

// TestRouteMemberCollectionFixtures runs the collector on each fixture
// document: a must-pass case yields exactly the listed operations, in order,
// and a must-fail case is refused with an actionable message naming the
// phrase.
func TestRouteMemberCollectionFixtures(t *testing.T) {
	corpus, err := testcase.LoadCorpus[routeMembersInput, routeMembersExpected](routeMembersYAML)
	if err != nil {
		t.Fatalf("load testdata/route_members.yaml: %v", err)
	}
	assert.RequireValid(t, corpus)
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	requireRouteManifestNames(t, "route members", routeMembersManifestYAML, names)

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var decoded any
			if err := c.Input.Document.Decode(&decoded); err != nil {
				t.Fatalf("decode the fixture document: %v", err)
			}
			document, err := json.Marshal(decoded)
			if err != nil {
				t.Fatalf("encode the fixture document as JSON: %v", err)
			}
			got, err := collectRoutes(c.Input.API.api(t), document)
			switch c.Classification {
			case testcase.MustPass:
				if err != nil {
					t.Fatalf("collectRoutes refused a valid document: %v", err)
				}
				if c.Expected.ErrorContains != "" || len(c.Expected.Operations) == 0 {
					t.Fatal("a must-pass case lists operations and no error")
				}
				if !reflect.DeepEqual(exportedRouteOperations(got), c.Expected.Operations) {
					t.Fatalf("collected operations differ\n got: %+v\nwant: %+v", exportedRouteOperations(got), c.Expected.Operations)
				}
			case testcase.MustFail:
				if c.Expected.ErrorContains == "" || len(c.Expected.Operations) != 0 {
					t.Fatal("a must-fail case names an error phrase and no operations")
				}
				if err == nil {
					t.Fatalf("collectRoutes accepted the document and collected %+v", exportedRouteOperations(got))
				}
				for _, want := range []string{c.Expected.ErrorContains, "no route members were written", "go run ./cmd/schema-gen"} {
					if !strings.Contains(err.Error(), want) {
						t.Fatalf("refusal %q does not contain %q", err, want)
					}
				}
			}
		})
	}
}

// exportedRouteOperations drops the renderer-only fields so a fixture compares
// what the document said.
func exportedRouteOperations(ops []routeOperation) []routeOperation {
	out := make([]routeOperation, 0, len(ops))
	for _, op := range ops {
		out = append(out, routeOperation{Name: op.Name, Method: op.Method, Pattern: op.Pattern, OperationID: op.OperationID, Deprecated: op.Deprecated, Params: op.Params})
	}
	return out
}

// TestRouteMembersFreshness_MatchSource regenerates the Go and TypeScript route
// member files from the built documents (through the same function main()
// calls) and requires the committed copies to be byte-identical.
func TestRouteMembersFreshness_MatchSource(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}
	specs, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("generate spec artifacts: %v", err)
	}
	artifacts, err := buildRouteArtifacts(specs)
	if err != nil {
		t.Fatalf("build route members: %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("route member generation wrote %d files, want the Go and the TypeScript file", len(artifacts))
	}
	for rel, want := range artifacts {
		got, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Errorf("committed route member file missing: %s\nRegenerate with `%s` and commit.", rel, regenCmd)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf(
				"STALE committed route member file: %s drifted from the built documents.\n"+
					"  what: the committed file does not match what the generator emits.\n"+
					"  why:  an operation changed without regeneration, or the file was hand-edited.\n"+
					"  fix:  run `%s` and commit the result.",
				rel, regenCmd)
		}
	}
}

// TestRouteMethodTableIsTheRouteMethodSet holds the generator's method table,
// which also renders the TypeScript HttpMethod set, equal to route.AllMethods
// in order.
func TestRouteMethodTableIsTheRouteMethodSet(t *testing.T) {
	var table []route.Method
	for _, method := range routeMethods {
		if method.docKey != strings.ToLower(method.wire) {
			t.Errorf("method table entry %s reads document key %q; want %q", method.wire, method.docKey, strings.ToLower(method.wire))
		}
		table = append(table, route.Method(method.wire))
	}
	if !reflect.DeepEqual(table, route.AllMethods) {
		t.Fatalf("generator method table %v differs from route.AllMethods %v; change both together", table, route.AllMethods)
	}
}

type routeRequiredNames struct {
	RequiredNames []string `yaml:"requiredNames"`
}

// requireRouteManifestNames fails when a required case is missing, a case is
// not in the manifest, or a name repeats.
func requireRouteManifestNames(t *testing.T, label string, manifestRaw []byte, names []string) {
	t.Helper()
	var manifest routeRequiredNames
	decoder := yaml.NewDecoder(bytes.NewReader(manifestRaw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("decode %s manifest: %v", label, err)
	}
	if len(manifest.RequiredNames) == 0 {
		t.Fatalf("%s manifest names no cases", label)
	}
	seen := map[string]bool{}
	for _, name := range names {
		seen[name] = true
	}
	required := map[string]bool{}
	for _, name := range manifest.RequiredNames {
		if required[name] {
			t.Fatalf("%s manifest repeats %q", label, name)
		}
		required[name] = true
		if !seen[name] {
			t.Fatalf("%s fixture is missing required case %q", label, name)
		}
	}
	for _, name := range names {
		if !required[name] {
			t.Fatalf("%s case %q is not in the manifest; add it there", label, name)
		}
	}
}
