package schema_test

import (
	_ "embed"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_directory_position_diagnostics.yaml
var directoryPositionDiagnosticsYAML []byte

//go:embed testdata/transcript_directory_position_diagnostics_manifest.yaml
var directoryPositionDiagnosticsManifestYAML []byte

type directoryPositionDiagnosticRef struct {
	Index  int `yaml:"index"`
	Before int `yaml:"before"`
	Count  int `yaml:"count"`
}

type directoryPositionDiagnosticInput struct {
	Version           string                           `yaml:"version"`
	Page              int                              `yaml:"page"`
	PerPage           int                              `yaml:"perPage"`
	TotalTurns        int                              `yaml:"totalTurns"`
	TotalSegments     int                              `yaml:"totalSegments"`
	TotalConversation int                              `yaml:"totalConversation"`
	Refs              []directoryPositionDiagnosticRef `yaml:"refs"`
}

func TestTranscriptDirectoryPositionDiagnostics(t *testing.T) {
	corpus := decodeStrictYAML[testcase.Corpus[directoryPositionDiagnosticInput, string]](
		t, directoryPositionDiagnosticsYAML, "directory position diagnostics",
	)
	caseassert.RequireValid(t, corpus)
	var names []string
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "directory position diagnostics", directoryPositionDiagnosticsManifestYAML, names)

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			in := c.Input
			var err error
			switch in.Version {
			case "v1":
				refs := make([]schema.TranscriptSegmentRef, 0, len(in.Refs))
				for _, ref := range in.Refs {
					refs = append(refs, schema.TranscriptSegmentRef{
						Index: ref.Index, FirstTurn: ref.Index, LastTurn: ref.Index,
						ConversationTurns: &schema.ConversationTurnSpan{Before: ref.Before, Count: ref.Count},
					})
				}
				err = (schema.TranscriptSegmentDirectory{
					Revision: "1", TotalTurns: in.TotalTurns, TotalSegments: in.TotalSegments,
					TotalConversationTurns: &in.TotalConversation, DirectoryPage: &in.Page,
					DirectoryPerPage: &in.PerPage, Segments: refs,
				}).Validate()
			case "v2":
				refs := make([]schema.PositionedSegmentRef, 0, len(in.Refs))
				for _, ref := range in.Refs {
					refs = append(refs, schema.PositionedSegmentRef{
						Index: ref.Index, SourceFirstPosition: ref.Index, SourceLastPosition: ref.Index,
						ConversationTurns: &schema.ConversationTurnSpan{Before: ref.Before, Count: ref.Count},
					})
				}
				session := schema.SessionID("ses_test")
				err = (schema.PositionedDirectory{
					Context: schema.TranscriptReadContext{
						ProtocolVersion: schema.SegmentedReadAPIVersion,
						Transport:       schema.TranscriptReadTransport{Kind: schema.TranscriptReadTransportLocalSession, SessionID: &session},
						Revision:        "1", Representation: "r1.test", Validation: schema.TranscriptReadServedLegacy,
						NativePresence: schema.TranscriptNativeAbsent, TotalTurns: in.TotalTurns,
						TotalSegments: in.TotalSegments, TotalConversationTurns: &in.TotalConversation,
					},
					Partition:    schema.TranscriptSegmentPartition{Kind: schema.TranscriptPartitionMain},
					UsagePresent: schema.TranscriptUsageUnknown, DirectoryPage: &in.Page,
					DirectoryPerPage: &in.PerPage, Segments: refs,
				}).Validate()
			default:
				t.Fatalf("unknown directory version %q", in.Version)
			}
			if err == nil || err.Error() != c.Expected {
				t.Fatalf("diagnostic = %v\nwant %s", err, c.Expected)
			}
		})
	}
}
