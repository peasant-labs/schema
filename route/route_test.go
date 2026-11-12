package route_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/route"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"

	"gopkg.in/yaml.v3"
)

// documentOperations reads every operation of one built document as
// "<METHOD> <pattern> <operationId>" keys.
func documentOperations(t *testing.T, specs openapi.SpecArtifacts, artifact string) []string {
	t.Helper()
	raw, ok := specs[artifact]
	if !ok {
		t.Fatalf("document %s was not built", artifact)
	}
	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode %s: %v", artifact, err)
	}
	var keys []string
	for pattern, item := range doc.Paths {
		for method, rawOp := range item {
			var op struct {
				OperationID string `json:"operationId"`
			}
			if err := json.Unmarshal(rawOp, &op); err != nil {
				t.Fatalf("decode %s %s in %s: %v", method, pattern, artifact, err)
			}
			keys = append(keys, strings.ToUpper(method)+" "+pattern+" "+op.OperationID)
		}
	}
	sort.Strings(keys)
	return keys
}

func memberKeys(t *testing.T, label string, routes []route.Route) []string {
	t.Helper()
	keys := make([]string, 0, len(routes))
	seen := map[string]bool{}
	for _, r := range routes {
		if !r.Method().IsValid() {
			t.Errorf("%s member %s has method %q outside route.AllMethods", label, r.OperationID(), r.Method())
		}
		key := r.Method().String() + " " + r.Pattern() + " " + r.OperationID()
		if seen[key] {
			t.Errorf("%s repeats member %s", label, key)
		}
		seen[key] = true
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// TestAllRoutesAreTheDocumentOperations holds AllVillage and AllLocal equal,
// by exact set membership of method, pattern and operationId, to the
// operations of the built documents: no operation without a member, and no
// member without an operation.
func TestAllRoutesAreTheDocumentOperations(t *testing.T) {
	specs, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("build the documents: %v", err)
	}
	sets := []struct {
		label    string
		artifact string
		members  []route.Route
	}{
		{"AllVillage", "village-api-" + openapi.VillageAPIVersion + ".json", route.AllVillage},
		{"AllLocal", "peasantlocal-api-" + openapi.PeasantLocalAPIVersion + ".json", route.AllLocal},
	}
	for _, set := range sets {
		t.Run(set.label, func(t *testing.T) {
			want := documentOperations(t, specs, set.artifact)
			got := memberKeys(t, set.label, set.members)
			wantSet := map[string]bool{}
			for _, key := range want {
				wantSet[key] = true
			}
			gotSet := map[string]bool{}
			for _, key := range got {
				gotSet[key] = true
				if !wantSet[key] {
					t.Errorf("%s has %s, which %s does not declare", set.label, key, set.artifact)
				}
			}
			for _, key := range want {
				if !gotSet[key] {
					t.Errorf("%s declares %s, which has no member in %s; run `go run ./cmd/schema-gen`", set.artifact, key, set.label)
				}
			}
		})
	}
}

// TestRoutesAreOrderedByPatternThenMethod holds the documented order of the
// complete sets.
func TestRoutesAreOrderedByPatternThenMethod(t *testing.T) {
	rank := map[route.Method]int{}
	for i, method := range route.AllMethods {
		rank[method] = i
	}
	for label, routes := range map[string][]route.Route{"AllVillage": route.AllVillage, "AllLocal": route.AllLocal} {
		ordered := sort.SliceIsSorted(routes, func(i, j int) bool {
			if routes[i].Pattern() != routes[j].Pattern() {
				return routes[i].Pattern() < routes[j].Pattern()
			}
			return rank[routes[i].Method()] < rank[routes[j].Method()]
		})
		if !ordered {
			t.Errorf("%s is not ordered by pattern and then by AllMethods", label)
		}
	}
}

// TestEveryMuxPatternMountsOnOneServeMux registers every member of each API on
// one net/http ServeMux, as a server mounts them. A malformed pattern, or two
// members the mux cannot tell apart, panics the registration.
func TestEveryMuxPatternMountsOnOneServeMux(t *testing.T) {
	for label, routes := range map[string][]route.Route{"AllVillage": route.AllVillage, "AllLocal": route.AllLocal} {
		mux := http.NewServeMux()
		for _, r := range routes {
			if r.MuxPattern() != r.Method().String()+" "+r.Pattern() {
				t.Errorf("%s %s: MuxPattern %q is not the method, a space and the pattern", label, r.OperationID(), r.MuxPattern())
			}
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Errorf("%s %s: ServeMux refused %q: %v", label, r.OperationID(), r.MuxPattern(), recovered)
					}
				}()
				mux.HandleFunc(r.MuxPattern(), func(http.ResponseWriter, *http.Request) {})
			}()
		}
	}
}

// TestAttachmentConstantsAreTheGeneratedMember holds the hand-written
// attachment constants equal to the generated Local member until they are
// removed.
func TestAttachmentConstantsAreTheGeneratedMember(t *testing.T) {
	if route.LocalGetAttachment.Pattern() != schema.AttachmentRoutePath {
		t.Errorf("LocalGetAttachment.Pattern() = %q, AttachmentRoutePath = %q", route.LocalGetAttachment.Pattern(), schema.AttachmentRoutePath)
	}
	if route.LocalGetAttachment.OperationID() != schema.AttachmentRouteOperationID {
		t.Errorf("LocalGetAttachment.OperationID() = %q, AttachmentRouteOperationID = %q", route.LocalGetAttachment.OperationID(), schema.AttachmentRouteOperationID)
	}
	if route.LocalGetAttachment.Method() != route.MethodGet {
		t.Errorf("LocalGetAttachment.Method() = %q, want GET", route.LocalGetAttachment.Method())
	}
}

// TestMethodSet holds the closed method set's membership test.
func TestMethodSet(t *testing.T) {
	for _, method := range route.AllMethods {
		if !method.IsValid() || method.String() != string(method) || method.String() != strings.ToUpper(method.String()) {
			t.Errorf("method %q is not a valid upper-case member", method)
		}
	}
	if route.Method("HEAD").IsValid() || route.Method("get").IsValid() {
		t.Error("IsValid accepted a method outside AllMethods")
	}
}

type routePathInput struct {
	API    string          `yaml:"api"`
	Member string          `yaml:"member"`
	Params routePathParams `yaml:"params"`
}

// routePathParams is every parameter the fixture's members take.
type routePathParams struct {
	ID     string `yaml:"id"`
	Index  int    `yaml:"index"`
	Owner  string `yaml:"owner"`
	Name   string `yaml:"name"`
	Number int    `yaml:"number"`
	Ref    string `yaml:"ref"`
}

type routePathExpected struct {
	Go         string `yaml:"go"`
	TypeScript string `yaml:"typescript"`
}

// routePathBuilders calls the generated builder of each member the fixture
// names. A member without parameters has no builder; its path is its pattern.
var routePathBuilders = map[string]func(routePathParams) string{
	"village/GetPullRequestAttachment": func(p routePathParams) string {
		return route.VillageGetPullRequestAttachmentPath(p.Owner, p.Name, p.Number)
	},
	"village/GetTranscriptSegment": func(p routePathParams) string {
		return route.VillageGetTranscriptSegmentPath(p.ID, p.Index)
	},
	"village/GetTranscriptSegments": func(p routePathParams) string {
		return route.VillageGetTranscriptSegmentsPath(p.ID)
	},
	"local/GetHealth":     func(routePathParams) string { return route.LocalGetHealth.Pattern() },
	"local/GetAttachment": func(p routePathParams) string { return route.LocalGetAttachmentPath(p.Ref) },
	"local/GetSession":    func(p routePathParams) string { return route.LocalGetSessionPath(p.ID) },
}

// TestPathBuildersFixtures builds each fixture path with the generated Go
// builder. The TypeScript members read the same file.
func TestPathBuildersFixtures(t *testing.T) {
	root := filepath.Join("..", "testdata")
	raw, err := os.ReadFile(filepath.Join(root, "route_paths.yaml"))
	if err != nil {
		t.Fatalf("read testdata/route_paths.yaml: %v", err)
	}
	corpus, err := testcase.LoadCorpus[routePathInput, routePathExpected](raw)
	if err != nil {
		t.Fatalf("load testdata/route_paths.yaml: %v", err)
	}
	assert.RequireValid(t, corpus)
	manifest, err := os.ReadFile(filepath.Join(root, "route_paths_manifest.yaml"))
	if err != nil {
		t.Fatalf("read testdata/route_paths_manifest.yaml: %v", err)
	}
	requireManifestNames(t, manifest, corpus)

	used := map[string]bool{}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Classification != testcase.MustPass {
				t.Fatalf("route_paths.yaml holds built paths only; move a refusal to testdata/typescript/route_path_refusals.yaml")
			}
			key := c.Input.API + "/" + c.Input.Member
			build, ok := routePathBuilders[key]
			if !ok {
				t.Fatalf("no Go builder is wired for %s; add it to routePathBuilders", key)
			}
			used[key] = true
			if got := build(c.Input.Params); got != c.Expected.Go {
				t.Fatalf("%s built %q, want %q", key, got, c.Expected.Go)
			}
		})
	}
	for key := range routePathBuilders {
		if !used[key] {
			t.Errorf("routePathBuilders wires %s, which no fixture case uses", key)
		}
	}
}

func requireManifestNames(t *testing.T, manifestRaw []byte, corpus testcase.Corpus[routePathInput, routePathExpected]) {
	t.Helper()
	var manifest struct {
		RequiredNames []string `yaml:"requiredNames"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(manifestRaw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("decode route_paths_manifest.yaml: %v", err)
	}
	var names []string
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	sort.Strings(names)
	required := append([]string(nil), manifest.RequiredNames...)
	sort.Strings(required)
	if fmt.Sprint(names) != fmt.Sprint(required) {
		t.Fatalf("route_paths.yaml cases %v differ from route_paths_manifest.yaml %v; a case was added or removed without its manifest entry", names, required)
	}
}
