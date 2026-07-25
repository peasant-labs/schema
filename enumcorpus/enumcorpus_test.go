package enumcorpus_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/enumcorpus"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

// enumExhaustionCase is one registered committed corpus: the enum name, the
// path relative to this package directory (where go test runs, matching
// go:generate's cwd), the canonical member list, and a render thunk mirroring
// cmd/gen-enum-corpora's registration so the freshness gate exercises the
// same call the generator makes.
type enumExhaustionCase[T enumcorpus.ClosedEnum] struct {
	enumName string
	path     string
	all      []T
	negative T
	render   func() ([]byte, error)
}

// runExhaustionAndFreshness is the generic guard pair every registered enum
// corpus shares (mirrors licensecorpus_test.go's two tests, generalized over
// any ClosedEnum): exhaustive coverage over the committed corpus run against
// the real IsValid, and byte-for-byte freshness against a fresh render.
func runExhaustionAndFreshness[T enumcorpus.ClosedEnum](t *testing.T, c enumExhaustionCase[T]) {
	t.Helper()
	t.Run(c.enumName+"/ExhaustiveCoverage", func(t *testing.T) {
		data, err := os.ReadFile(c.path)
		if err != nil {
			t.Fatalf("read committed corpus %s: %v (regenerate with `go generate ./...`)", c.path, err)
		}
		corpus, err := testcase.LoadCorpus[T, bool](data)
		if err != nil {
			t.Fatalf("load committed corpus: %v", err)
		}
		assert.RequireMin(t, corpus, len(c.all)+1)
		assert.RequireValid(t, corpus)

		covered := map[T]bool{}
		for _, cs := range corpus.Cases {
			if got := cs.Input.IsValid(); got != cs.Expected {
				t.Errorf("case %q: %s(%q).IsValid() = %v, want %v", cs.Name, c.enumName, cs.Input, got, cs.Expected)
			}
			if cs.Classification == testcase.MustPass {
				covered[cs.Input] = true
			}
		}
		for _, member := range c.all {
			if !covered[member] {
				t.Errorf("%s member %q has no must-pass case in the committed corpus; regenerate with `go generate ./...` after widening the set", c.enumName, member)
			}
		}
		foundNegative := false
		for _, cs := range corpus.Cases {
			if cs.Classification == testcase.MustFail && cs.Input == c.negative {
				foundNegative = true
				break
			}
		}
		if !foundNegative {
			t.Errorf("missing must-fail negative for %q", c.negative)
		}
	})
	t.Run(c.enumName+"/Freshness", func(t *testing.T) {
		want, err := os.ReadFile(c.path)
		if err != nil {
			t.Fatalf("read committed corpus %s: %v", c.path, err)
		}
		got, err := c.render()
		if err != nil {
			t.Fatalf("render corpus: %v", err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("STALE committed %s corpus: %s drifted from a fresh render.\n"+
				"  what: the committed corpus does not match what the generator emits.\n"+
				"  why:  the set changed without regenerating, or the file was hand-edited.\n"+
				"  fix:  run `go generate ./...` and commit enumcorpus/%s.",
				c.enumName, c.path, c.path)
		}
	})
}

// TestEnumCorpora_ExhaustiveCoverageAndFreshness guards all ten Local API
// 0.5.0 closed sets registered by cmd/gen-enum-corpora: each reddens if its
// set widens without the corpus regenerated (ExhaustiveCoverage) or if the
// committed file drifts from the generator (Freshness).
func TestEnumCorpora_ExhaustiveCoverageAndFreshness(t *testing.T) {
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.AssociationKind]{
		enumName: "AssociationKind", path: "testdata/association_kind_corpus.yaml",
		all: schema.AllAssociationKinds, negative: schema.AssociationKind("unknown-kind"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("AssociationKind", "schema.AllAssociationKinds",
				schema.AllAssociationKinds, schema.AssociationKind("unknown-kind"),
				`accepting "unknown-kind" would silently widen the closed association-kind set beyond AllAssociationKinds.`)
		},
	})
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.AssociationEvidence]{
		enumName: "AssociationEvidence", path: "testdata/association_evidence_corpus.yaml",
		all: schema.AllAssociationEvidences, negative: schema.AssociationEvidence("unknown-evidence"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("AssociationEvidence", "schema.AllAssociationEvidences",
				schema.AllAssociationEvidences, schema.AssociationEvidence("unknown-evidence"),
				`accepting "unknown-evidence" would silently widen the closed association-evidence set beyond AllAssociationEvidences.`)
		},
	})
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.Confidence]{
		enumName: "Confidence", path: "testdata/confidence_corpus.yaml",
		all: schema.AllConfidences, negative: schema.Confidence("unknown-confidence"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("Confidence", "schema.AllConfidences",
				schema.AllConfidences, schema.Confidence("unknown-confidence"),
				`accepting "unknown-confidence" would silently widen the closed confidence set beyond AllConfidences.`)
		},
	})
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.RewriteResolution]{
		enumName: "RewriteResolution", path: "testdata/rewrite_resolution_corpus.yaml",
		all: schema.AllRewriteResolutions, negative: schema.RewriteResolution("unknown-resolution"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("RewriteResolution", "schema.AllRewriteResolutions",
				schema.AllRewriteResolutions, schema.RewriteResolution("unknown-resolution"),
				`accepting "unknown-resolution" would silently widen the closed rewrite-resolution set beyond AllRewriteResolutions.`)
		},
	})
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.RewriteMethod]{
		enumName: "RewriteMethod", path: "testdata/rewrite_method_corpus.yaml",
		all: schema.AllRewriteMethods, negative: schema.RewriteMethod("unknown-method"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("RewriteMethod", "schema.AllRewriteMethods",
				schema.AllRewriteMethods, schema.RewriteMethod("unknown-method"),
				`accepting "unknown-method" would silently widen the closed rewrite-method set beyond AllRewriteMethods.`)
		},
	})
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.InsightKind]{
		enumName: "InsightKind", path: "testdata/insight_kind_corpus.yaml",
		all: schema.AllInsightKinds, negative: schema.InsightKind("unknown-kind"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("InsightKind", "schema.AllInsightKinds",
				schema.AllInsightKinds, schema.InsightKind("unknown-kind"),
				`accepting "unknown-kind" would silently widen the closed insight-kind set beyond AllInsightKinds.`)
		},
	})
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.InsightProvenance]{
		enumName: "InsightProvenance", path: "testdata/insight_provenance_corpus.yaml",
		all: schema.AllInsightProvenances, negative: schema.InsightProvenance("unknown-provenance"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("InsightProvenance", "schema.AllInsightProvenances",
				schema.AllInsightProvenances, schema.InsightProvenance("unknown-provenance"),
				`accepting "unknown-provenance" would silently widen the closed insight-provenance set beyond AllInsightProvenances.`)
		},
	})
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.ReadAttributionState]{
		enumName: "ReadAttributionState", path: "testdata/read_attribution_state_corpus.yaml",
		all: schema.AllReadAttributionStates, negative: schema.ReadAttributionState("unknown-state"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("ReadAttributionState", "schema.AllReadAttributionStates",
				schema.AllReadAttributionStates, schema.ReadAttributionState("unknown-state"),
				`accepting "unknown-state" would silently widen the closed read-attribution-state set beyond AllReadAttributionStates.`)
		},
	})
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.ReadStateGrade]{
		enumName: "ReadStateGrade", path: "testdata/read_state_grade_corpus.yaml",
		all: schema.AllReadStateGrades, negative: schema.ReadStateGrade("unknown-grade"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("ReadStateGrade", "schema.AllReadStateGrades",
				schema.AllReadStateGrades, schema.ReadStateGrade("unknown-grade"),
				`accepting "unknown-grade" would silently widen the closed read-state-grade set beyond AllReadStateGrades.`)
		},
	})
	runExhaustionAndFreshness(t, enumExhaustionCase[schema.TargetKind]{
		enumName: "TargetKind", path: "testdata/target_kind_corpus.yaml",
		all: schema.AllTargetKinds, negative: schema.TargetKind("unknown-target"),
		render: func() ([]byte, error) {
			return enumcorpus.RenderCorpus("TargetKind", "schema.AllTargetKinds",
				schema.AllTargetKinds, schema.TargetKind("unknown-target"),
				`accepting "unknown-target" would silently widen the closed target-kind set beyond AllTargetKinds.`)
		},
	})
}

// TestReadStateGradeRegistrySeedCrossCheck pins this module's half of the
// the read-state grade / registry-seed cross-check: AllReadStateGrades with "none" removed
// byte-equals ReadStateGradeRegistrySeedPermissibleValues, in the same
// ascending order. The consumer-side read-state registry seed test
// pins the other half against ReadStateGradeRegistrySeedPermissibleValues.
func TestReadStateGradeRegistrySeedCrossCheck(t *testing.T) {
	want := make([]string, 0, len(schema.AllReadStateGrades))
	for _, grade := range schema.AllReadStateGrades {
		if grade == schema.ReadStateGradeNone {
			continue
		}
		want = append(want, grade.String())
	}
	got := schema.ReadStateGradeRegistrySeedPermissibleValues
	if len(got) != len(want) {
		t.Fatalf("ReadStateGradeRegistrySeedPermissibleValues has %d entries, want %d (AllReadStateGrades minus none)", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ReadStateGradeRegistrySeedPermissibleValues[%d] = %q, want %q; keep it byte-identical to AllReadStateGrades minus none so the peasant-side registry seed cannot drift", i, got[i], want[i])
		}
	}
}
