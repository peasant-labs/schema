package openapi

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

// The fixtures of the groups of routes Village mounts, one testcase corpus of
// must-pass cases each.
var (
	//go:embed testdata/village_sign_in_operations.yaml
	villageSignInOperationsYAML []byte
	//go:embed testdata/village_account_operations.yaml
	villageAccountOperationsYAML []byte
	//go:embed testdata/village_browse_operations.yaml
	villageBrowseOperationsYAML []byte
	//go:embed testdata/village_transcript_aux_operations.yaml
	villageTranscriptAuxOperationsYAML []byte
)

// villageMountedOperationGroup is one group's fixture, named for messages.
type villageMountedOperationGroup struct {
	name    string
	fixture []byte
}

// villageMountedOperationGroups lists every group's fixture, in the order
// they are read.
var villageMountedOperationGroups = []villageMountedOperationGroup{
	{name: "sign-in and installation", fixture: villageSignInOperationsYAML},
	{name: "account", fixture: villageAccountOperationsYAML},
	{name: "browse", fixture: villageBrowseOperationsYAML},
	{name: "transcript auxiliary", fixture: villageTranscriptAuxOperationsYAML},
}

// requiredMountedOperationIDs are the routes Village mounts that the contract
// declares from this document version on: every route that had no contract
// owner before. Each must have exactly one fixture row and one operation in
// the document.
var requiredMountedOperationIDs = []string{
	// Browser sign-in: five provider entries and five callbacks.
	"startGitHubSignIn",
	"startGitLabSignIn",
	"startHuggingFaceSignIn",
	"startCodebergSignIn",
	"startSourceHutSignIn",
	"completeGitHubSignIn",
	"completeGitLabSignIn",
	"completeHuggingFaceSignIn",
	"completeCodebergSignIn",
	"completeSourceHutSignIn",
	// GitHub App installation.
	"startGitHubAppInstall",
	"completeGitHubAppInstall",
	// Session, account and API keys.
	"logOut",
	"deleteMyAccount",
	"updateMyDiscoverability",
	"setMyUsername",
	"createMyAPIKey",
	"listMyAPIKeys",
	"revokeMyAPIKey",
	"setMyOrganizationVisibility",
	// Organizations, tags, profiles and projects.
	"searchOrganizations",
	"listTags",
	"listPopularTags",
	"getUserProfile",
	"listUserOrganizations",
	"getUserProject",
	"setMyProjectDisplayName",
	"clearMyProjectDisplayName",
	// Transcripts.
	"deleteTranscript",
	"listTranscriptAttestations",
	"createTranscriptAttestation",
	"deleteTranscriptAttestation",
	"listTranscriptAnnotations",
	"createTranscriptAnnotation",
	"listTranscriptCommits",
	"publishTranscriptBatch",
	// Self-description and health.
	"getVillageAPIDocument",
	"getHealth",
}

// requiredCorrectedOperationIDs are operations declared before these routes
// whose declaration was corrected alongside them and is pinned by a row in
// the same fixtures. cliLogin was declared with a 204 its handler never sends;
// it answers a redirect, as the sign-in routes do. cliExchangeCode sits behind
// the sign-in limiter, whose headers and 429 it did not declare.
var requiredCorrectedOperationIDs = []string{
	"cliLogin",
	"cliExchangeCode",
}

// villageMountedOperationInput locates one operation in the document.
type villageMountedOperationInput struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type villageMountedOperationCorpus = testcase.Corpus[villageMountedOperationInput, villageOperationExpectation]

// loadVillageMountedOperationCorpora reads every group fixture and guards
// each against being empty or vacuous.
func loadVillageMountedOperationCorpora(t *testing.T) map[string]villageMountedOperationCorpus {
	t.Helper()
	corpora := map[string]villageMountedOperationCorpus{}
	for _, group := range villageMountedOperationGroups {
		corpus, err := testcase.LoadCorpus[villageMountedOperationInput, villageOperationExpectation](group.fixture)
		if err != nil {
			t.Fatalf("load the %s operation fixture: %v", group.name, err)
		}
		assert.RequireMin(t, corpus, 1)
		assert.RequireValid(t, corpus)
		corpora[group.name] = corpus
	}
	return corpora
}

// TestVillageMountedOperationsAreDeclared requires every route Village mounts
// without a former contract owner to be declared exactly as its fixture row
// states: statuses, media types, headers and parameters as exact sets.
func TestVillageMountedOperationsAreDeclared(t *testing.T) {
	corpora := loadVillageMountedOperationCorpora(t)
	spec, err := BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec: %v", err)
	}

	var ids []string
	locations := map[string]string{}
	for _, entry := range villageMountedOperationGroups {
		group := entry.name
		for _, c := range corpora[group].Cases {
			t.Run(c.Name, func(t *testing.T) {
				if c.Classification != testcase.MustPass {
					t.Fatalf("%s: every row is a declared operation, so it is must-pass", group)
				}
				if c.Name != c.Expected.OperationID {
					t.Fatalf("%s: case name %q differs from operation_id %q; name each row by its operation id", group, c.Name, c.Expected.OperationID)
				}
				for _, problem := range checkVillageOperationExpectation(spec, c.Input.Method, c.Input.Path, c.Expected) {
					t.Errorf("%s %s %s (%s): %s", group, c.Input.Method, c.Input.Path, c.Name, problem)
				}
			})
			key := strings.ToLower(c.Input.Method) + " " + c.Input.Path
			if previous, ok := locations[key]; ok {
				t.Errorf("%s and %s both describe %s", previous, c.Name, key)
			}
			locations[key] = c.Name
			ids = append(ids, c.Name)
		}
	}
	required := append(append([]string(nil), requiredMountedOperationIDs...), requiredCorrectedOperationIDs...)
	assertSameNames(t, "Village mounted operation rows", ids, required)
}
