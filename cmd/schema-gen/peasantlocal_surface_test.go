package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

type localAPISurfaceFixture struct {
	Operations []localAPIOperationIdentity                    `yaml:"operations"`
	Mutations  testcase.Corpus[localAPISurfaceMutation, bool] `yaml:"mutations"`
}

// localAPIDescriptionAnchor names one clause a published Local API operation
// description must state. Prose is the product in a contract module, so the
// clauses live in a closed set: the fixture declares every member exactly once,
// and removing a member, or a fixture row that carries one, fails here instead
// of quietly shrinking the guarded prose.
type localAPIDescriptionAnchor string

const (
	localAPIAnchorNoIngestionBrowseFilter            localAPIDescriptionAnchor = "no_ingestion_browse_filter"
	localAPIAnchorListIndependentBounds              localAPIDescriptionAnchor = "list_independent_bounds"
	localAPIAnchorIndexedSearchMatch                 localAPIDescriptionAnchor = "indexed_search_match"
	localAPIAnchorEligibleSessionCount               localAPIDescriptionAnchor = "eligible_session_count"
	localAPIAnchorHomeNoIngestionBrowseFilter        localAPIDescriptionAnchor = "home_no_ingestion_browse_filter"
	localAPIAnchorHomeIndependentBounds              localAPIDescriptionAnchor = "home_independent_bounds"
	localAPIAnchorUnknownParentOnly404               localAPIDescriptionAnchor = "unknown_parent_only_404"
	localAPIAnchorTokenScopeExcludesSelection        localAPIDescriptionAnchor = "token_scope_excludes_ingestion_selection"
	localAPIAnchorTokenScopeExcludesPageSize         localAPIDescriptionAnchor = "token_scope_excludes_page_size"
	localAPIAnchorMetadataFailsClosed                localAPIDescriptionAnchor = "publication_metadata_fails_closed"
	localAPIAnchorMemberScopeExcludesSelection       localAPIDescriptionAnchor = "member_scope_excludes_ingestion_selection"
	localAPIAnchorPublicationsNotSelectionFiltered   localAPIDescriptionAnchor = "publications_not_selection_filtered"
	localAPIAnchorSessionByIDNotSelectionFiltered    localAPIDescriptionAnchor = "session_by_id_not_selection_filtered"
	localAPIAnchorSummariesReadNotNarrowed           localAPIDescriptionAnchor = "summaries_stored_read_not_narrowed"
	localAPIAnchorTranscriptByIDNotSelectionFiltered localAPIDescriptionAnchor = "transcript_by_id_not_selection_filtered"
)

// allLocalAPIDescriptionAnchors is the canonical set. Its order is
// alphabetical by constant name so a reader can compare it with the fixture.
var allLocalAPIDescriptionAnchors = []localAPIDescriptionAnchor{
	localAPIAnchorEligibleSessionCount,
	localAPIAnchorHomeIndependentBounds,
	localAPIAnchorHomeNoIngestionBrowseFilter,
	localAPIAnchorIndexedSearchMatch,
	localAPIAnchorListIndependentBounds,
	localAPIAnchorMemberScopeExcludesSelection,
	localAPIAnchorMetadataFailsClosed,
	localAPIAnchorNoIngestionBrowseFilter,
	localAPIAnchorPublicationsNotSelectionFiltered,
	localAPIAnchorSessionByIDNotSelectionFiltered,
	localAPIAnchorSummariesReadNotNarrowed,
	localAPIAnchorTokenScopeExcludesPageSize,
	localAPIAnchorTokenScopeExcludesSelection,
	localAPIAnchorTranscriptByIDNotSelectionFiltered,
	localAPIAnchorUnknownParentOnly404,
}

func (a localAPIDescriptionAnchor) isValid() bool {
	for _, known := range allLocalAPIDescriptionAnchors {
		if a == known {
			return true
		}
	}
	return false
}

// localAPIOperationDescriptionAnchor is one declared clause: its closed-set
// name and the literal text the published description must contain.
type localAPIOperationDescriptionAnchor struct {
	Name localAPIDescriptionAnchor `yaml:"name"`
	Text string                    `yaml:"text"`
}

type localAPIOperationIdentity struct {
	Path        string `yaml:"path"`
	Method      string `yaml:"method"`
	OperationID string `yaml:"operation_id"`
	// DescriptionAnchors are the named clauses this operation's published
	// description must still state. The identity check proves the route is
	// declared; these prove the rule its prose states did not quietly go away
	// in a regeneration.
	DescriptionAnchors []localAPIOperationDescriptionAnchor `yaml:"description_anchors,omitempty"`
}

type localAPISurfaceMutation struct {
	Kind        string `yaml:"kind"`
	Path        string `yaml:"path"`
	Method      string `yaml:"method"`
	OperationID string `yaml:"operation_id"`
}

func TestPeasantLocalGeneratedSurfaceHasExactIdentity(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	fixture := loadLocalAPISurfaceFixture(t, root)
	specPath := filepath.Join(root, "generated", "peasantlocal-api-"+schema.PeasantLocalAPIVersion+".json")
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := localAPIOperations(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateLocalAPIOperations(actual, fixture.Operations); err != nil {
		t.Fatal(err)
	}
	if err := validateLocalAPIOperationAnchors(spec, fixture.Operations); err != nil {
		t.Fatal(err)
	}
}

func TestPeasantLocalGeneratedSurfaceMutationProof(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	fixture := loadLocalAPISurfaceFixture(t, root)
	canonical := make(map[string]string, len(fixture.Operations))
	for _, operation := range fixture.Operations {
		canonical[operation.Path+"#"+operation.Method] = operation.OperationID
	}
	for _, mutation := range fixture.Mutations.Cases {
		t.Run(mutation.Name, func(t *testing.T) {
			mutated := make(map[string]string, len(canonical)+1)
			for key, value := range canonical {
				mutated[key] = value
			}
			key := mutation.Input.Path + "#" + mutation.Input.Method
			switch mutation.Input.Kind {
			case "remove":
				delete(mutated, key)
			case "add", "redirect":
				mutated[key] = mutation.Input.OperationID
			default:
				t.Fatalf("unknown local API surface mutation %q", mutation.Input.Kind)
			}
			accepted := validateLocalAPIOperations(mutated, fixture.Operations) == nil
			if accepted != mutation.Expected {
				t.Fatalf("accepted=%v, want %v", accepted, mutation.Expected)
			}
		})
	}
}

func loadLocalAPISurfaceFixture(t *testing.T, root string) localAPISurfaceFixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "testdata", "typescript", "local_api_surface.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture localAPISurfaceFixture
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if err := fixture.Mutations.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := fixture.Mutations.CheckMin(3); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func localAPIOperations(spec []byte) (map[string]string, error) {
	var document struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(spec, &document); err != nil {
		return nil, fmt.Errorf("decode local API spec: %w", err)
	}
	operations := make(map[string]string)
	for path, methods := range document.Paths {
		for method, operation := range methods {
			if operation.OperationID == "" {
				continue
			}
			operations[path+"#"+strings.ToLower(method)] = operation.OperationID
		}
	}
	return operations, nil
}

func validateLocalAPIOperations(actual map[string]string, expected []localAPIOperationIdentity) error {
	want := make(map[string]string, len(expected))
	for _, operation := range expected {
		key := operation.Path + "#" + operation.Method
		if _, duplicate := want[key]; duplicate {
			return fmt.Errorf("local API fixture repeats operation %s", key)
		}
		want[key] = operation.OperationID
	}
	keys := func(values map[string]string) []string {
		out := make([]string, 0, len(values))
		for key := range values {
			out = append(out, key)
		}
		sort.Strings(out)
		return out
	}
	for _, key := range keys(want) {
		if actual[key] != want[key] {
			return fmt.Errorf("local API operation %s=%q, want %q; restore the fixture-accounted path and operation identity", key, actual[key], want[key])
		}
	}
	for _, key := range keys(actual) {
		if _, expected := want[key]; !expected {
			return fmt.Errorf("local API exposes unaccounted operation %s=%q; add an intentional fixture row or remove the stray operation", key, actual[key])
		}
	}
	return nil
}

// validateLocalAPIOperationAnchors checks each declared clause against the
// published description of its operation, and checks the declaration itself
// against the closed set: every member appears exactly once, in one operation.
// The identity map proves the route exists; this proves the prose stating its
// stored-browsing rule survives a regeneration, which an operation-only
// manifest cannot see.
func validateLocalAPIOperationAnchors(spec []byte, expected []localAPIOperationIdentity) error {
	var document struct {
		Paths map[string]map[string]struct {
			Description string `json:"description"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(spec, &document); err != nil {
		return fmt.Errorf("decode local API spec for description anchors: %w", err)
	}
	descriptions := make(map[string]string, len(document.Paths))
	for path, methods := range document.Paths {
		for method, operation := range methods {
			if operation.Description != "" {
				descriptions[path+"#"+strings.ToLower(method)] = operation.Description
			}
		}
	}
	declared := make(map[localAPIDescriptionAnchor]string, len(allLocalAPIDescriptionAnchors))
	for _, operation := range expected {
		key := operation.Path + "#" + strings.ToLower(operation.Method)
		for _, anchor := range operation.DescriptionAnchors {
			if !anchor.Name.isValid() {
				return fmt.Errorf("local API operation %s declares description anchor %q, which is not a member of the closed set", key, anchor.Name)
			}
			if strings.TrimSpace(anchor.Text) == "" {
				return fmt.Errorf("local API operation %s declares description anchor %q with no text; an empty needle is satisfied by any description", key, anchor.Name)
			}
			if owner, duplicate := declared[anchor.Name]; duplicate {
				return fmt.Errorf("description anchor %q is declared twice, on %s and %s; each named clause has exactly one owner", anchor.Name, owner, key)
			}
			declared[anchor.Name] = key
			description, ok := descriptions[key]
			if !ok {
				return fmt.Errorf("local API operation %s publishes no description to anchor", key)
			}
			if !strings.Contains(description, anchor.Text) {
				return fmt.Errorf("local API operation %s no longer states %q (anchor %q); the published rule was weakened or deleted", key, anchor.Text, anchor.Name)
			}
		}
	}
	for _, name := range allLocalAPIDescriptionAnchors {
		if _, ok := declared[name]; !ok {
			return fmt.Errorf("the fixture declares no description anchor %q; deleting a named clause must fail here", name)
		}
	}
	return nil
}
