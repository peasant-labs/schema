// Command gen-enum-corpora writes the Local API 0.5.0 enum-exhaustion
// corpora to their committed artifacts. It is invoked by the go:generate
// directive in the enumcorpus package (which sets the working directory to
// that package), so it writes each CommittedPath relative to the package
// directory.
//
// From the module root, regenerate with `go generate ./...` whenever one of
// the ten registered closed sets changes, then commit the result. The
// freshness gate fails if a committed file drifts from a fresh render.
package main

import (
	"log"
	"os"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/enumcorpus"
)

// registration pairs one closed enum's committed corpus path with a thunk
// that renders it, so main can loop over one table instead of repeating the
// render-and-write boilerplate ten times.
type registration struct {
	path   string
	render func() ([]byte, error)
}

func registrations() []registration {
	return []registration{
		{
			path: "testdata/association_kind_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"AssociationKind", "schema.AllAssociationKinds",
					schema.AllAssociationKinds, schema.AssociationKind("unknown-kind"),
					`accepting "unknown-kind" would silently widen the closed association-kind set beyond AllAssociationKinds.`,
				)
			},
		},
		{
			path: "testdata/association_evidence_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"AssociationEvidence", "schema.AllAssociationEvidences",
					schema.AllAssociationEvidences, schema.AssociationEvidence("unknown-evidence"),
					`accepting "unknown-evidence" would silently widen the closed association-evidence set beyond AllAssociationEvidences.`,
				)
			},
		},
		{
			path: "testdata/confidence_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"Confidence", "schema.AllConfidences",
					schema.AllConfidences, schema.Confidence("unknown-confidence"),
					`accepting "unknown-confidence" would silently widen the closed confidence set beyond AllConfidences.`,
				)
			},
		},
		{
			path: "testdata/rewrite_resolution_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"RewriteResolution", "schema.AllRewriteResolutions",
					schema.AllRewriteResolutions, schema.RewriteResolution("unknown-resolution"),
					`accepting "unknown-resolution" would silently widen the closed rewrite-resolution set beyond AllRewriteResolutions.`,
				)
			},
		},
		{
			path: "testdata/rewrite_method_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"RewriteMethod", "schema.AllRewriteMethods",
					schema.AllRewriteMethods, schema.RewriteMethod("unknown-method"),
					`accepting "unknown-method" would silently widen the closed rewrite-method set beyond AllRewriteMethods.`,
				)
			},
		},
		{
			path: "testdata/insight_kind_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"InsightKind", "schema.AllInsightKinds",
					schema.AllInsightKinds, schema.InsightKind("unknown-kind"),
					`accepting "unknown-kind" would silently widen the closed insight-kind set beyond AllInsightKinds.`,
				)
			},
		},
		{
			path: "testdata/insight_provenance_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"InsightProvenance", "schema.AllInsightProvenances",
					schema.AllInsightProvenances, schema.InsightProvenance("unknown-provenance"),
					`accepting "unknown-provenance" would silently widen the closed insight-provenance set beyond AllInsightProvenances.`,
				)
			},
		},
		{
			path: "testdata/read_attribution_state_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"ReadAttributionState", "schema.AllReadAttributionStates",
					schema.AllReadAttributionStates, schema.ReadAttributionState("unknown-state"),
					`accepting "unknown-state" would silently widen the closed read-attribution-state set beyond AllReadAttributionStates.`,
				)
			},
		},
		{
			path: "testdata/read_state_grade_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"ReadStateGrade", "schema.AllReadStateGrades",
					schema.AllReadStateGrades, schema.ReadStateGrade("unknown-grade"),
					`accepting "unknown-grade" would silently widen the closed read-state-grade set beyond AllReadStateGrades.`,
				)
			},
		},
		{
			path: "testdata/target_kind_corpus.yaml",
			render: func() ([]byte, error) {
				return enumcorpus.RenderCorpus(
					"TargetKind", "schema.AllTargetKinds",
					schema.AllTargetKinds, schema.TargetKind("unknown-target"),
					`accepting "unknown-target" would silently widen the closed target-kind set beyond AllTargetKinds.`,
				)
			},
		},
	}
}

func main() {
	for _, reg := range registrations() {
		data, err := reg.render()
		if err != nil {
			log.Fatalf("gen-enum-corpora: render %s: %v", reg.path, err)
		}
		if err := os.WriteFile(reg.path, data, 0o644); err != nil {
			log.Fatalf("gen-enum-corpora: write %s: %v", reg.path, err)
		}
	}
}
