package schema_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/local-api/rewrites.yaml
var rewritesCasesYAML []byte

// TestRewrittenCommit_FixtureContract drives every
// RewrittenCommit case in testdata/local-api/rewrites.yaml against the real
// RewrittenCommit.Validate, and confirms every member of the three closed
// sets it composes (RewriteResolution, RewriteMethod, Confidence) has at
// least one must-pass covering case.
func TestRewrittenCommit_FixtureContract(t *testing.T) {
	corpus := loadRewritesCorpus(t)

	coveredResolutions := map[schema.RewriteResolution]bool{}
	coveredMethods := map[schema.RewriteMethod]bool{}
	coveredConfidences := map[schema.Confidence]bool{}

	for _, c := range corpus.Cases {
		err := c.Input.Validate()
		valid := err == nil
		if valid != c.Expected {
			t.Errorf("%s: Validate() error=%v (valid=%v), want valid=%v", c.Name, err, valid, c.Expected)
			continue
		}
		if c.Classification == testcase.MustFail {
			for _, required := range []string{" at schema.", "during wire-boundary validation"} {
				if !strings.Contains(err.Error(), required) {
					t.Errorf("%s: validation error %q is not actionable; missing %q", c.Name, err, required)
				}
			}
			continue
		}
		coveredResolutions[c.Input.Resolution] = true
		coveredMethods[c.Input.Method] = true
		coveredConfidences[c.Input.Confidence] = true
	}

	for _, resolution := range schema.AllRewriteResolutions {
		if !coveredResolutions[resolution] {
			t.Errorf("RewriteResolution member %q has no must-pass fixture case", resolution)
		}
	}
	for _, method := range schema.AllRewriteMethods {
		if !coveredMethods[method] {
			t.Errorf("RewriteMethod member %q has no must-pass fixture case", method)
		}
	}
	for _, confidence := range schema.AllConfidences {
		if !coveredConfidences[confidence] {
			t.Errorf("Confidence member %q has no must-pass fixture case", confidence)
		}
	}
}

func loadRewritesCorpus(t *testing.T) testcase.Corpus[schema.RewrittenCommit, bool] {
	t.Helper()
	corpus, err := testcase.LoadCorpus[schema.RewrittenCommit, bool](rewritesCasesYAML)
	if err != nil {
		t.Fatalf("load rewrites corpus: %v", err)
	}
	assert.RequireMin(t, corpus, 17)
	assert.RequireValid(t, corpus)
	return corpus
}

func rewrittenCommitFixture(t *testing.T, name string) schema.RewrittenCommit {
	t.Helper()
	for _, fixture := range loadRewritesCorpus(t).Cases {
		if fixture.Name == name {
			if fixture.Classification != testcase.MustFail || fixture.Expected {
				t.Fatalf("fixture %q must describe a rejected input", name)
			}
			return fixture.Input
		}
	}
	t.Fatalf("rewrites corpus is missing required case %q", name)
	return schema.RewrittenCommit{}
}

// TestRewrittenCommit_RegressionMutationsRestoreValidity proves the two
// regression fixtures fail for the named invariant rather than an unrelated
// malformed field, then repairs that one mutation and requires validation to
// pass.
func TestRewrittenCommit_RegressionMutationsRestoreValidity(t *testing.T) {
	t.Run("live_with_successor_hash_is_rejected", func(t *testing.T) {
		fixture := rewrittenCommitFixture(t, "live_with_successor_hash_is_rejected")
		err := fixture.Validate()
		if err == nil || !strings.Contains(err.Error(), `resolution is "live" but successorHash`) {
			t.Fatalf("Validate() error=%v, want the actionable live-successor rejection", err)
		}
		fixture.SuccessorHash = nil
		if err := fixture.Validate(); err != nil {
			t.Fatalf("clearing successorHash did not repair the fixture: %v", err)
		}
	})

	t.Run("empty_session_ids_is_rejected", func(t *testing.T) {
		fixture := rewrittenCommitFixture(t, "empty_session_ids_is_rejected")
		err := fixture.Validate()
		if err == nil || !strings.Contains(err.Error(), "sessionIds is empty") {
			t.Fatalf("Validate() error=%v, want the actionable empty-sessionIds rejection", err)
		}
		fixture.SessionIDs = []schema.SessionID{"session-a"}
		if err := fixture.Validate(); err != nil {
			t.Fatalf("adding an originating session ID did not repair the fixture: %v", err)
		}
	})
}
