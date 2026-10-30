package schema

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/publication/round_trip.yaml
var publicationRoundTripYAML []byte

type roundTripPatch struct {
	Path  string `yaml:"path"`
	Value string `yaml:"value"`
}

type roundTripInput struct {
	Value   string           `yaml:"value"`
	Context []roundTripPatch `yaml:"context,omitempty"`
}

type roundTripExpected struct {
	Survives bool `yaml:"survives"`
}

type roundTripDiscard struct {
	Name   string `yaml:"name"`
	Reason string `yaml:"reason"`
}

type roundTripFixture struct {
	Documents map[string]string                                  `yaml:"documents"`
	Keys      testcase.Corpus[roundTripInput, roundTripExpected] `yaml:"keys"`
	Discarded []roundTripDiscard                                 `yaml:"discarded"`
}

// roundTripFieldSet binds one decoder field set to the document whose decoder
// enforces it and to the location of its object inside that document.
type roundTripFieldSet struct {
	document string
	name     string
	variable string
	location string
	fields   map[string]struct{}
}

// roundTripFieldSets lists every field set the authoritative publication
// decoders accept. TestPublicationRoundTrip_FieldSetRegistryIsComplete keeps it
// equal to the field set variables declared in publication_decode.go.
var roundTripFieldSets = []roundTripFieldSet{
	{"request", "request", "authoritativeRequestFields", "", authoritativeRequestFields},
	{"request", "client", "clientFields", "client", clientFields},
	{"request", "identity", "identityFields", "identity", identityFields},
	{"request", "model", "modelFields", "model", modelFields},
	{"request", "timestamp", "timestampFields", "timestamp", timestampFields},
	{"request", "source", "sourceFields", "source", sourceFields},
	{"request", "git", "requestGitFields", "git", requestGitFields},
	{"request", "commit", "commitFields", "git.commits.0", commitFields},
	{"request", "project", "projectFields", "project", projectFields},
	{"request", "stats", "statsFields", "stats", statsFields},
	{"request", "quality", "qualityFields", "quality", qualityFields},
	{"request", "entry", "entryFields", "entries.0", entryFields},
	{"request", "subagent", "subagentFields", "subagents.0", subagentFields},
	{"request", "diagnostics", "diagnosticsFields", "diagnostics", diagnosticsFields},
	{"request", "diagnosticEntry", "diagnosticEntryFields", "diagnostics.warnings.0", diagnosticEntryFields},
	{"operation", "operation", "canonicalOperationFields", "", canonicalOperationFields},
	{"operation", "replacement", "canonicalReplacementFields", "replacement", canonicalReplacementFields},
	{"operation", "git", "canonicalGitFields", "replacement.git", canonicalGitFields},
	{"operation", "license", "licenseOperationFields", "license", licenseOperationFields},
	{"operation", "associations", "associationOperationFields", "associations", associationOperationFields},
	{"operation", "association", "publishedAssociationFields", "associations.associations.0", publishedAssociationFields},
	{"response", "response", "publishResponseFields", "", publishResponseFields},
	{"response", "applied", "publishAppliedFields", "applied", publishAppliedFields},
	{"response", "normalizedValues", "publishNormalizedFields", "applied.normalizedValues", publishNormalizedFields},
}

var roundTripDecoders = map[string]func([]byte) (any, error){
	"request": func(data []byte) (any, error) { return DecodeAuthoritativePublishRequest(data) },
	"operation": func(data []byte) (any, error) {
		return DecodeCanonicalPublishOperation(data)
	},
	"response": func(data []byte) (any, error) { return DecodePublishResponse(data) },
}

func loadPublicationRoundTrip(t *testing.T) roundTripFixture {
	t.Helper()
	var fixture roundTripFixture
	decoder := yaml.NewDecoder(bytes.NewReader(publicationRoundTripYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("load publication round-trip fixture: %v", err)
	}
	assert.RequireMin(t, fixture.Keys, 1)
	assert.RequireValid(t, fixture.Keys)
	return fixture
}

// TestPublicationRoundTrip_FieldSetRegistryIsComplete fails when a field set
// variable is declared in publication_decode.go without a registry row, so a
// new accepted object cannot escape the round-trip guard.
func TestPublicationRoundTrip_FieldSetRegistryIsComplete(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "publication_decode.go", nil, 0)
	if err != nil {
		t.Fatalf("parse publication_decode.go: %v", err)
	}
	declared := map[string]struct{}{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			for _, name := range spec.(*ast.ValueSpec).Names {
				if strings.HasSuffix(name.Name, "Fields") {
					declared[name.Name] = struct{}{}
				}
			}
		}
	}
	registered := map[string]struct{}{}
	for _, set := range roundTripFieldSets {
		registered[set.variable] = struct{}{}
		if _, ok := roundTripDecoders[set.document]; !ok {
			t.Fatalf("field set %s names unknown document %q", set.variable, set.document)
		}
	}
	for name := range declared {
		if _, ok := registered[name]; !ok {
			t.Errorf("decoder field set %s has no round-trip registry row", name)
		}
	}
	for name := range registered {
		if _, ok := declared[name]; !ok {
			t.Errorf("round-trip registry names %s, which publication_decode.go does not declare", name)
		}
	}
}

// TestPublicationRoundTrip_EveryAcceptedKeySurvives decodes one document per
// accepted key, re-encodes the decoded value, and requires the key's value back.
// The required case names are derived from the decoder field sets.
func TestPublicationRoundTrip_EveryAcceptedKeySurvives(t *testing.T) {
	fixture := loadPublicationRoundTrip(t)
	required := map[string]roundTripFieldSet{}
	for _, set := range roundTripFieldSets {
		for key := range set.fields {
			required[set.document+"/"+set.name+"/"+key] = set
		}
	}
	covered := map[string]struct{}{}
	for _, discard := range fixture.Discarded {
		if _, ok := required[discard.Name]; !ok {
			t.Errorf("discarded key %q is not accepted by any decoder field set", discard.Name)
		}
		if strings.TrimSpace(discard.Reason) == "" {
			t.Errorf("discarded key %q must state why the contract discards it", discard.Name)
		}
		if _, dup := covered[discard.Name]; dup {
			t.Errorf("discarded key %q is listed twice", discard.Name)
		}
		covered[discard.Name] = struct{}{}
	}
	for _, row := range fixture.Keys.Cases {
		set, ok := required[row.Name]
		if !ok {
			t.Errorf("round-trip case %q is not an accepted key of any decoder field set", row.Name)
			continue
		}
		if _, dup := covered[row.Name]; dup {
			t.Errorf("round-trip key %q is covered twice", row.Name)
		}
		covered[row.Name] = struct{}{}
		if !row.Expected.Survives || row.Classification != testcase.MustPass {
			t.Errorf("round-trip case %q must be a must-pass case that survives", row.Name)
		}
		key := row.Name[strings.LastIndex(row.Name, "/")+1:]
		path := key
		if set.location != "" {
			path = set.location + "." + key
		}
		t.Run(row.Name, func(t *testing.T) {
			requireRoundTripSurvives(t, fixture.Documents[set.document], set.document, path, row.Input)
		})
	}
	var missing []string
	for name := range required {
		if _, ok := covered[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	for _, name := range missing {
		t.Errorf("accepted key %s has no round-trip case and no discard reason", name)
	}
}

func requireRoundTripSurvives(t *testing.T, base, document, path string, input roundTripInput) {
	t.Helper()
	if base == "" {
		t.Fatalf("no base document %q", document)
	}
	var doc any
	if err := json.Unmarshal([]byte(base), &doc); err != nil {
		t.Fatalf("decode base document %q: %v", document, err)
	}
	for _, patch := range input.Context {
		doc = setRoundTripPath(t, doc, patch.Path, patch.Value)
	}
	doc = setRoundTripPath(t, doc, path, input.Value)
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("encode mutated document: %v", err)
	}
	decoded, err := roundTripDecoders[document](data)
	if err != nil {
		t.Fatalf("the %s decoder refused an accepted key at %s: %v", document, path, err)
	}
	encoded, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("re-encode decoded %s: %v", document, err)
	}
	var reread any
	if err := json.Unmarshal(encoded, &reread); err != nil {
		t.Fatalf("decode re-encoded %s: %v", document, err)
	}
	var want any
	if err := json.Unmarshal([]byte(input.Value), &want); err != nil {
		t.Fatalf("decode case value: %v", err)
	}
	got, found := getRoundTripPath(reread, path)
	if !found {
		t.Fatalf("%s was dropped: the decoder accepted it but the re-encoded %s has no such key\nre-encoded: %s", path, document, encoded)
	}
	if !reflect.DeepEqual(got, want) {
		gotJSON, _ := json.Marshal(got)
		t.Fatalf("%s changed across decode and re-encode: sent %s, got %s", path, input.Value, gotJSON)
	}
}

func setRoundTripPath(t *testing.T, doc any, path, raw string) any {
	t.Helper()
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatalf("decode value for %s: %v", path, err)
	}
	parts := strings.Split(path, ".")
	cursor := doc
	for i, part := range parts {
		last := i == len(parts)-1
		switch node := cursor.(type) {
		case map[string]any:
			if last {
				node[part] = value
				return doc
			}
			cursor = node[part]
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(node) {
				t.Fatalf("path %s has no array element %q in the base document", path, part)
			}
			if last {
				node[index] = value
				return doc
			}
			cursor = node[index]
		default:
			t.Fatalf("path %s has no container at %q in the base document", path, part)
		}
	}
	panic(fmt.Sprintf("unreachable: empty path %q", path))
}

func getRoundTripPath(doc any, path string) (any, bool) {
	cursor := doc
	for _, part := range strings.Split(path, ".") {
		switch node := cursor.(type) {
		case map[string]any:
			next, ok := node[part]
			if !ok {
				return nil, false
			}
			cursor = next
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(node) {
				return nil, false
			}
			cursor = node[index]
		default:
			return nil, false
		}
	}
	return cursor, true
}
