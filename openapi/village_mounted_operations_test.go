package openapi

import (
	"embed"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

// villageMountedOperationFixtures holds one fixture per group of routes
// Village mounts. The pattern also matches older Village operation fixtures in
// another shape; only the files named in villageMountedOperationGroups are
// read.
//
//go:embed testdata/village_*_operations.yaml
var villageMountedOperationFixtures embed.FS

// villageMountedOperationGroups names the fixture of each group, in the order
// they are read. Each is a testcase corpus of must-pass cases.
var villageMountedOperationGroups = []string{
	"testdata/village_sign_in_operations.yaml",
	"testdata/village_account_operations.yaml",
	"testdata/village_browse_operations.yaml",
	"testdata/village_transcript_aux_operations.yaml",
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
// it answers a redirect, as the sign-in routes do.
var requiredCorrectedOperationIDs = []string{
	"cliLogin",
}

// villageMountedOperationInput locates one operation in the document.
type villageMountedOperationInput struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type villageMountedOperationCorpus = testcase.Corpus[villageMountedOperationInput, villageOperationExpectation]

// loadVillageMountedOperationCorpora reads every group fixture. A group whose
// fixture does not exist is reported by name rather than skipped, so a missing
// group cannot pass vacuously.
func loadVillageMountedOperationCorpora(t *testing.T) map[string]villageMountedOperationCorpus {
	t.Helper()
	corpora := map[string]villageMountedOperationCorpus{}
	var missing []string
	for _, name := range villageMountedOperationGroups {
		raw, err := villageMountedOperationFixtures.ReadFile(name)
		if errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, name)
			continue
		}
		if err != nil {
			t.Fatalf("read Village mounted operation fixture %s: %v", name, err)
		}
		corpus, err := testcase.LoadCorpus[villageMountedOperationInput, villageOperationExpectation](raw)
		if err != nil {
			t.Fatalf("load Village mounted operation fixture %s: %v", name, err)
		}
		assert.RequireMin(t, corpus, 1)
		assert.RequireValid(t, corpus)
		corpora[name] = corpus
	}
	if len(missing) > 0 {
		t.Fatalf("Village mounted operation fixtures %s do not exist; each group of declared routes keeps its rows in its own fixture under openapi/testdata; add the fixture with one row per operation of the group", strings.Join(missing, ", "))
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
	document := villageDocument(t, spec)

	var ids []string
	locations := map[string]string{}
	for _, group := range villageMountedOperationGroups {
		for _, c := range corpora[group].Cases {
			t.Run(c.Name, func(t *testing.T) {
				if c.Classification != testcase.MustPass {
					t.Fatalf("%s: every row is a declared operation, so it is must-pass", group)
				}
				if c.Name != c.Expected.OperationID {
					t.Fatalf("%s: case name %q differs from operation_id %q; name each row by its operation id", group, c.Name, c.Expected.OperationID)
				}
				method := strings.ToLower(c.Input.Method)
				for _, problem := range checkVillageOperationExpectation(document, method, c.Input.Path, c.Expected) {
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
