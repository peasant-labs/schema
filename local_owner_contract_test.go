package schema_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	schema "github.com/peasant-labs/schema"
	specpkg "github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	testassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/local_owner_contract.yaml
var localOwnerContractYAML []byte

type localOwnerOperation struct {
	Name          string `yaml:"name"`
	Method        string `yaml:"method"`
	Path          string `yaml:"path"`
	OperationID   string `yaml:"operation_id"`
	ErrorStatuses []int  `yaml:"error_statuses"`
}

type localOwnerFixtures struct {
	RequiredNames    []string                      `yaml:"required_names"`
	LocalOwner       testcase.Corpus[string, bool] `yaml:"local_owner"`
	OwnerStatus      testcase.Corpus[string, bool] `yaml:"owner_status"`
	ClaimEntry       testcase.Corpus[string, bool] `yaml:"claim_entry"`
	ClaimEntryList   testcase.Corpus[string, bool] `yaml:"claim_entry_list"`
	ClaimResult      testcase.Corpus[string, bool] `yaml:"claim_result"`
	ArtifactIdentity testcase.Corpus[string, bool] `yaml:"artifact_identity"`
	AnnotationRow    testcase.Corpus[string, bool] `yaml:"annotation_row"`
	Prompt           testcase.Corpus[string, bool] `yaml:"prompt"`
	PromptStates     testcase.Corpus[string, bool] `yaml:"prompt_states"`
	ClaimSources     testcase.Corpus[string, bool] `yaml:"claim_sources"`
	Operations       []localOwnerOperation         `yaml:"operations"`
}

func loadLocalOwnerFixtures(t *testing.T) localOwnerFixtures {
	t.Helper()
	var fixtures localOwnerFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(localOwnerContractYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		t.Fatalf("load local owner contract: %v", err)
	}
	arms := map[string]testcase.Corpus[string, bool]{
		"local_owner":       fixtures.LocalOwner,
		"owner_status":      fixtures.OwnerStatus,
		"claim_entry":       fixtures.ClaimEntry,
		"claim_entry_list":  fixtures.ClaimEntryList,
		"claim_result":      fixtures.ClaimResult,
		"artifact_identity": fixtures.ArtifactIdentity,
		"annotation_row":    fixtures.AnnotationRow,
		"prompt":            fixtures.Prompt,
		"prompt_states":     fixtures.PromptStates,
		"claim_sources":     fixtures.ClaimSources,
	}
	seen := map[string]bool{}
	for name, corpus := range arms {
		minimum := 2
		if name == "claim_entry_list" {
			// The list arm holds the one two-account reuse case; every other
			// claim entry loads as a single object.
			minimum = 1
		}
		testassert.RequireMin(t, corpus, minimum)
		testassert.RequireValid(t, corpus)
		for _, c := range corpus.Cases {
			if seen[c.Name] {
				t.Fatalf("local owner case %q appears in more than one arm", c.Name)
			}
			seen[c.Name] = true
		}
	}
	for _, name := range fixtures.RequiredNames {
		if !seen[name] {
			t.Fatalf("required local owner case %q is missing", name)
		}
	}
	if len(seen) != len(fixtures.RequiredNames) {
		t.Fatalf("local owner required-name manifest has %d names for %d cases", len(fixtures.RequiredNames), len(seen))
	}
	wantOperations := map[string]bool{"read": false, "claim": false, "decline": false}
	for _, operation := range fixtures.Operations {
		if _, ok := wantOperations[operation.Name]; !ok {
			t.Fatalf("local owner operations fixture has unexpected row %q; want exactly read, claim, and decline", operation.Name)
		}
		if wantOperations[operation.Name] {
			t.Fatalf("local owner operations fixture names %q more than once", operation.Name)
		}
		wantOperations[operation.Name] = true
	}
	for name, seen := range wantOperations {
		if !seen {
			t.Fatalf("local owner operations fixture is missing the %q route", name)
		}
	}
	return fixtures
}

func TestLocalOwnerContractValidation(t *testing.T) {
	fixtures := loadLocalOwnerFixtures(t)
	for _, c := range fixtures.LocalOwner.Cases {
		t.Run(c.Name, func(t *testing.T) {
			requireLocalOwnerVerdict[schema.LocalOwner](t, c.Input, c.Expected)
		})
	}
	for _, c := range fixtures.OwnerStatus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			requireLocalOwnerVerdict[schema.LocalOwnerStatus](t, c.Input, c.Expected)
		})
	}
	for _, c := range fixtures.ClaimEntry.Cases {
		t.Run(c.Name, func(t *testing.T) {
			requireLocalOwnerVerdict[schema.LocalClaimEntry](t, c.Input, c.Expected)
		})
	}
	for _, c := range fixtures.ClaimEntryList.Cases {
		t.Run(c.Name, func(t *testing.T) {
			requireLocalOwnerListVerdict[schema.LocalClaimEntry](t, c.Input, c.Expected)
		})
	}
	for _, c := range fixtures.ClaimResult.Cases {
		t.Run(c.Name, func(t *testing.T) {
			requireLocalOwnerVerdict[schema.LocalOwnerClaimResult](t, c.Input, c.Expected)
		})
	}
	for _, c := range fixtures.ArtifactIdentity.Cases {
		t.Run(c.Name, func(t *testing.T) {
			requireLocalOwnerVerdict[schema.LocalArtifactIdentity](t, c.Input, c.Expected)
		})
	}
	for _, c := range fixtures.AnnotationRow.Cases {
		t.Run(c.Name, func(t *testing.T) {
			requireLocalOwnerVerdict[schema.LocalAnnotationRow](t, c.Input, c.Expected)
		})
	}
	for _, c := range fixtures.Prompt.Cases {
		t.Run(c.Name, func(t *testing.T) {
			requireLocalOwnerVerdict[schema.LocalOwnerClaimPrompt](t, c.Input, c.Expected)
		})
	}
}

type localOwnerValidatable interface {
	Validate() error
}

func requireLocalOwnerVerdict[T any, P interface {
	*T
	localOwnerValidatable
}](t *testing.T, input string, want bool) {
	t.Helper()
	var value T
	if err := json.Unmarshal([]byte(input), P(&value)); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	err := P(&value).Validate()
	if (err == nil) != want {
		t.Fatalf("valid=%v, want %v: %v", err == nil, want, err)
	}
}

// requireLocalOwnerListVerdict loads one list-shaped claim-entry case as a
// slice, fails on a decode error, and validates every entry. The two-account
// reuse case is the only list; everything else loads as one entry.
func requireLocalOwnerListVerdict[T any, P interface {
	*T
	localOwnerValidatable
}](t *testing.T, input string, want bool) {
	t.Helper()
	var values []T
	if err := json.Unmarshal([]byte(input), &values); err != nil {
		t.Fatalf("decode fixture as a claim-entry list: %v", err)
	}
	var err error
	for i := range values {
		if err = P(&values[i]).Validate(); err != nil {
			break
		}
	}
	if (err == nil) != want {
		t.Fatalf("valid=%v, want %v: %v", err == nil, want, err)
	}
}

func TestLocalOwnerPromptStateAndClaimSourceSets(t *testing.T) {
	fixtures := loadLocalOwnerFixtures(t)
	for _, c := range fixtures.PromptStates.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := schema.LocalOwnerPromptState(c.Input).IsValid(); got != c.Expected {
				t.Fatalf("IsValid(%q)=%v, want %v", c.Input, got, c.Expected)
			}
		})
	}
	for _, c := range fixtures.ClaimSources.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := schema.LocalClaimSource(c.Input).IsValid(); got != c.Expected {
				t.Fatalf("IsValid(%q)=%v, want %v", c.Input, got, c.Expected)
			}
		})
	}
}

// TestLocalOwnerRouteStatuses pins the non-2xx statuses each local-owner route
// declares. The read can fail in the store; claim can be malformed,
// unauthenticated, or fail in the store; decline can be unauthenticated or fail
// in the store. None of these routes contacts Village.
func TestLocalOwnerRouteStatuses(t *testing.T) {
	fixtures := loadLocalOwnerFixtures(t)
	spec, err := specpkg.BuildPeasantLocalAPISpec()
	if err != nil {
		t.Fatalf("build Local API spec: %v", err)
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal Local API spec: %v", err)
	}
	var document struct {
		Paths map[string]map[string]struct {
			OperationID string                     `json:"operationId"`
			Responses   map[string]json.RawMessage `json:"responses"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("decode Local API spec: %v", err)
	}
	for _, want := range fixtures.Operations {
		t.Run(want.Name, func(t *testing.T) {
			path, ok := document.Paths[want.Path]
			if !ok {
				t.Fatalf("path %s is absent from the Local API spec", want.Path)
			}
			operation, ok := path[want.Method]
			if !ok {
				t.Fatalf("%s %s is absent from the Local API spec", want.Method, want.Path)
			}
			if operation.OperationID != want.OperationID {
				t.Fatalf("%s %s operationId=%q, want %q", want.Method, want.Path, operation.OperationID, want.OperationID)
			}
			var got []int
			for rawStatus := range operation.Responses {
				var status int
				if _, err := fmt.Sscanf(rawStatus, "%d", &status); err != nil {
					t.Fatalf("%s %s has non-numeric response status %q", want.Method, want.Path, rawStatus)
				}
				if status >= 400 {
					got = append(got, status)
				}
			}
			sort.Ints(got)
			wantStatuses := append([]int(nil), want.ErrorStatuses...)
			sort.Ints(wantStatuses)
			if fmt.Sprint(got) != fmt.Sprint(wantStatuses) {
				t.Fatalf("%s %s error statuses=%v, want %v", want.Method, want.Path, got, wantStatuses)
			}
		})
	}
}

// TestTypesComponentsRequireOnlyDeclaredProperties guards the generated Types
// document against a required entry that names no property. The two frozen
// legacy envelopes keep the Go field name of their embedded summary; every
// other component must require only properties it declares.
func TestTypesComponentsRequireOnlyDeclaredProperties(t *testing.T) {
	spec, err := specpkg.BuildTypesSpec()
	if err != nil {
		t.Fatalf("build Types spec: %v", err)
	}
	// These two released envelopes required the Go field name of the embedded
	// summary before the expansion hook existed. Their versions are frozen, so
	// they cannot be corrected; new envelopes opt in to expansion instead.
	frozen := map[string]string{
		"PullAnnotation":       "frozen released envelope; carries the Go embedded-field name",
		"LegacyPullAnnotation": "frozen legacy envelope; carries the Go embedded-field name",
	}
	for component, schemaMap := range spec.Components.Schemas {
		properties, _ := schemaMap["properties"].(map[string]interface{}) // unstring-allow: go-json-key-assert — inspect the generated component map directly
		required, _ := schemaMap["required"].([]interface{})              // unstring-allow: go-json-key-assert — inspect the generated component map directly
		for _, raw := range required {
			name, ok := raw.(string)
			if !ok {
				t.Fatalf("component %s has a non-string required entry %v", component, raw)
			}
			if _, declared := properties[name]; declared {
				continue
			}
			if _, exempt := frozen[component]; exempt {
				continue
			}
			t.Errorf("component %s requires %q but declares no such property; add the embedded struct to expandedEmbeddedRequired or correct the required list", component, name)
		}
	}
}
