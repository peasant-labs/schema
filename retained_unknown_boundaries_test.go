package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/retained_unknown_boundaries.yaml
var retainedBoundaryYAML []byte

type aliasInput struct {
	Path     []string `yaml:"path"`
	Key      string   `yaml:"key"`
	Value    string   `yaml:"value"`
	First    bool     `yaml:"first"`
	Envelope bool     `yaml:"envelope"`
}
type evidenceMutationInput struct {
	SourceCase  string `yaml:"source_case"`
	Replacement string `yaml:"replacement"`
}
type retainedBoundaryFixtures struct {
	BasePatch             string                                       `yaml:"base_patch"`
	RequiredAliasNames    []string                                     `yaml:"required_alias_names"`
	RequiredMutationNames []string                                     `yaml:"required_mutation_names"`
	Aliases               testcase.Corpus[aliasInput, bool]            `yaml:"aliases"`
	Mutations             testcase.Corpus[evidenceMutationInput, bool] `yaml:"preservation_mutations"`
}

func decodeRetainedBoundaryFixtures(raw []byte) (retainedBoundaryFixtures, error) {
	var f retainedBoundaryFixtures
	d := yaml.NewDecoder(bytes.NewReader(raw))
	d.KnownFields(true)
	if err := d.Decode(&f); err != nil {
		return f, err
	}
	var trailing any
	if err := d.Decode(&trailing); err != io.EOF {
		return f, fmt.Errorf("expected one boundary fixture document: %v", err)
	}
	if err := requiredBoundaryCases(f.Aliases, f.RequiredAliasNames); err != nil {
		return f, err
	}
	if err := requiredBoundaryCases(f.Mutations, f.RequiredMutationNames); err != nil {
		return f, err
	}
	return f, nil
}

func requiredBoundaryCases[I any](corpus testcase.Corpus[I, bool], required []string) error {
	if err := corpus.Validate(); err != nil {
		return err
	}
	names := map[string]bool{}
	for _, name := range required {
		if name == "" || names[name] {
			return fmt.Errorf("invalid required case %q", name)
		}
		names[name] = true
	}
	if len(names) == 0 {
		return fmt.Errorf("empty required-name manifest")
	}
	for _, c := range corpus.Cases {
		if !names[c.Name] {
			return fmt.Errorf("untracked case %s", c.Name)
		}
		delete(names, c.Name)
	}
	if len(names) > 0 {
		return fmt.Errorf("missing required cases %v", names)
	}
	return nil
}

func loadRetainedBoundaryFixtures(t *testing.T) retainedBoundaryFixtures {
	t.Helper()
	f, err := decodeRetainedBoundaryFixtures(retainedBoundaryYAML)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// Preserve the deliberately selected order within the mutated object. Ordinary
// map serialization would sort the aliases and miss one overwrite direction.
func injectWireAlias(t *testing.T, raw []byte, in aliasInput) []byte {
	t.Helper()
	if len(in.Path) == 0 {
		key, _ := json.Marshal(in.Key)
		field := append(append(key, ':'), []byte(in.Value)...)
		body := bytes.TrimSpace(raw)[1 : len(bytes.TrimSpace(raw))-1]
		if in.First {
			return append(append(append(append([]byte{'{'}, field...), ','), body...), '}')
		}
		return append(append(append(append([]byte{'{'}, body...), ','), field...), '}')
	}
	part := in.Path[0]
	in.Path = in.Path[1:]
	var result any
	if bytes.TrimSpace(raw)[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			t.Fatal(err)
		}
		i, err := strconv.Atoi(part)
		if err != nil || i < 0 || i >= len(items) {
			t.Fatal("invalid fixture array path")
		}
		items[i] = injectWireAlias(t, items[i], in)
		result = items
	} else {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		if _, ok := fields[part]; !ok {
			t.Fatal("missing fixture path", part)
		}
		fields[part] = injectWireAlias(t, fields[part], in)
		result = fields
	}
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestRetainedUnknownCanonicalKeys(t *testing.T) {
	f, base := loadRetainedBoundaryFixtures(t), loadRetainedUnknownFixtures(t)
	raw := retainedUnknownWire(t, base, retainedUnknownInput{DetailPatch: f.BasePatch})
	baseline, err := schema.DecodeSessionDetailPayloadRaw(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(schema.RequiredContentCapabilities(baseline), schema.AllContentCapabilities) {
		t.Fatal("alias baseline does not carry every existing capability")
	}
	wrap := func(detail []byte) []byte {
		return []byte(`{"contractVersion":"1.0.0","kind":"session_detail","sessionDetail":` + string(detail) + `}`)
	}
	for _, c := range f.Aliases.Cases {
		t.Run(c.Name, func(t *testing.T) {
			input := raw
			if c.Input.Envelope {
				input = wrap(input)
			}
			mutated := injectWireAlias(t, input, c.Input)
			if !c.Input.Envelope {
				value, err := schema.DecodeSessionDetailPayloadRaw(mutated)
				if (err == nil) != c.Expected {
					t.Fatalf("standalone alias accepted=%v: %v", err == nil, err)
				}
				if err == nil && !slices.Equal(schema.RequiredContentCapabilities(value), schema.AllContentCapabilities) {
					t.Fatal("additive field removed evidence requirements")
				}
				mutated = wrap(mutated)
			}
			_, err := schema.DecodeTranscriptContentRaw(mutated)
			if (err == nil) != c.Expected {
				t.Fatalf("envelope alias accepted=%v: %v", err == nil, err)
			}
			if err != nil && !strings.Contains(err.Error(), "alias") {
				t.Fatalf("fixture failed for unrelated reason: %v", err)
			}
		})
	}
}

func TestRetainedUnknownPreservationOracle(t *testing.T) {
	f, base := loadRetainedBoundaryFixtures(t), loadRetainedUnknownFixtures(t)
	for _, c := range f.Mutations.Cases {
		t.Run(c.Name, func(t *testing.T) {
			index := slices.IndexFunc(base.Cases, func(row testcase.Case[retainedUnknownInput, retainedUnknownExpected]) bool {
				return row.Name == c.Input.SourceCase
			})
			if index < 0 {
				t.Fatal("missing source case", c.Input.SourceCase)
			}
			raw := retainedUnknownWire(t, base, base.Cases[index].Input)
			var original struct {
				Records     []originalRetainedRecord           `json:"retainedUnknown"`
				Diagnostics *originalInterpretationDiagnostics `json:"diagnostics"`
			}
			if err := json.Unmarshal(raw, &original); err != nil {
				t.Fatal(err)
			}
			actual, err := schema.DecodeSessionDetailPayloadRaw(raw)
			if err != nil {
				t.Fatal(err)
			}
			if !retainedEvidenceEqual(actual, original.Records, original.Diagnostics) {
				t.Fatal("unmodified decoder changed evidence")
			}
			actual.RetainedUnknown[0].Payload = c.Input.Replacement
			if retainedEvidenceEqual(actual, original.Records, original.Diagnostics) != c.Expected {
				t.Fatal("preservation oracle missed payload mutation")
			}
		})
	}
}

func TestRetainedUnknownBoundaryDeletionProtection(t *testing.T) {
	f := loadRetainedBoundaryFixtures(t)
	for i, c := range f.Aliases.Cases {
		t.Run(c.Name, func(t *testing.T) {
			copy := f
			copy.Aliases.Cases = slices.Delete(slices.Clone(f.Aliases.Cases), i, i+1)
			raw, _ := yaml.Marshal(copy)
			if _, err := decodeRetainedBoundaryFixtures(raw); err == nil {
				t.Fatal("deleted alias case accepted")
			}
		})
	}
	for i, c := range f.Mutations.Cases {
		t.Run(c.Name, func(t *testing.T) {
			copy := f
			copy.Mutations.Cases = slices.Delete(slices.Clone(f.Mutations.Cases), i, i+1)
			raw, _ := yaml.Marshal(copy)
			if _, err := decodeRetainedBoundaryFixtures(raw); err == nil {
				t.Fatal("deleted mutation case accepted")
			}
		})
	}
	if _, err := decodeRetainedBoundaryFixtures(append(slices.Clone(retainedBoundaryYAML), []byte("\nunknown: true\n")...)); err == nil {
		t.Fatal("unknown fixture key accepted")
	}
	if _, err := decodeRetainedBoundaryFixtures(append(slices.Clone(retainedBoundaryYAML), []byte("\n---\n{}\n")...)); err == nil {
		t.Fatal("trailing fixture document accepted")
	}
}

func TestRetainedUnknownDuplicateRecipes(t *testing.T) {
	f := loadRetainedUnknownFixtures(t)
	payload := func(name string) string {
		index := slices.IndexFunc(f.Cases, func(c testcase.Case[retainedUnknownInput, retainedUnknownExpected]) bool { return c.Name == name })
		if index < 0 {
			t.Fatal("missing duplicate recipe", name)
		}
		var wire struct {
			Records []originalRetainedRecord `json:"retainedUnknown"`
		}
		if err := json.Unmarshal(retainedUnknownWire(t, f, f.Cases[index].Input), &wire); err != nil {
			t.Fatal(err)
		}
		return wire.Records[0].Payload
	}
	escaped, plain := payload("duplicate_payload_keys"), payload("plain_duplicate_payload_keys")
	if !strings.Contains(escaped, `\u0061`) || strings.Contains(plain, `\u0061`) || escaped == plain {
		t.Fatal("duplicate recipes lost their escaped/plain distinction before the production scanner")
	}
}
