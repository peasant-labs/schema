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
	corpus, err := testcase.LoadCorpus[schema.RewrittenCommit, bool](rewritesCasesYAML)
	if err != nil {
		t.Fatalf("load rewrites corpus: %v", err)
	}
	assert.RequireMin(t, corpus, len(schema.AllRewriteResolutions)+1)
	assert.RequireValid(t, corpus)

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
