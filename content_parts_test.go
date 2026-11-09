package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
)

func TestContentPartsFixtures(t *testing.T) {
	fixtures, err := schema.LoadContentPartsFixtures()
	if err != nil {
		t.Fatalf("load content parts fixtures: %v", err)
	}
	for _, fixtureCase := range fixtures.Cases.Cases {
		fixtureCase := fixtureCase
		t.Run(fixtureCase.Name, func(t *testing.T) {
			var part schema.TurnContentPart
			if err := json.Unmarshal([]byte(fixtureCase.Input.Part), &part); err != nil {
				t.Fatalf("decode fixture part: %v", err)
			}
			err := schema.ValidateTurnContentPart(schema.ImagePartMode(fixtureCase.Input.Mode), part)
			if (err == nil) != fixtureCase.Expected.Accepted {
				t.Fatalf("ValidateTurnContentPart accepted=%v, want %v; err=%v", err == nil, fixtureCase.Expected.Accepted, err)
			}
			if fixtureCase.Expected.Accepted && fixtureCase.Classification != testcase.MustPass {
				t.Fatalf("case classification %q disagrees with an accepted expectation", fixtureCase.Classification)
			}
			if !fixtureCase.Expected.Accepted && fixtureCase.Classification != testcase.MustFail {
				t.Fatalf("case classification %q disagrees with a rejected expectation", fixtureCase.Classification)
			}
			if err != nil && !strings.Contains(err.Error(), fixtureCase.Expected.ErrorContains) {
				t.Fatalf("validation error %q does not contain %q", err, fixtureCase.Expected.ErrorContains)
			}
		})
	}
}

// TestContentPartsValidatesOnTheSessionDetailPath proves the local detail
// validator and the transcript envelope validator both run the ordered parts,
// so the fixture matrix is not only exercised through the direct validator.
func TestContentPartsValidatesOnTheSessionDetailPath(t *testing.T) {
	local := schema.SessionDetailPayload{
		Harness: schema.HarnessClaudeCode,
		Turns: []schema.TurnDetail{{
			Index:   0,
			Role:    schema.RoleUser,
			Content: "look",
			ContentParts: []schema.TurnContentPart{
				{Type: schema.TurnPartTypeText, Text: "look"},
				{Type: schema.TurnPartTypeImage, Image: &schema.ImagePart{
					Ref:       strings.Repeat("a", 64),
					MediaType: schema.ImageMediaTypePNG,
					ByteSize:  2048,
				}},
			},
		}},
	}
	if err := schema.ValidateSessionDetailPayload(local); err != nil {
		t.Fatalf("local detail with an inline image part was rejected: %v", err)
	}

	published := schema.SessionDetailPayload{
		Harness: schema.HarnessClaudeCode,
		Turns: []schema.TurnDetail{{
			Index: 0,
			Role:  schema.RoleUser,
			ContentParts: []schema.TurnContentPart{
				{Type: schema.TurnPartTypeImage, Image: &schema.ImagePart{OmittedReason: schema.OmittedReasonNotPublished}},
			},
		}},
	}
	if err := schema.ValidateSessionDetailPayload(published); err == nil {
		t.Fatal("local detail validator accepted the published image shape")
	}
	envelope := schema.TranscriptContent{ContractVersion: "1.0.0", Kind: schema.ContentKindSessionDetail, SessionDetail: &published}
	if err := schema.ValidateTranscriptContent(envelope); err != nil {
		t.Fatalf("published transcript envelope with a not_published image was rejected: %v", err)
	}

	leaked := schema.SessionDetailPayload{
		Harness: schema.HarnessClaudeCode,
		Turns: []schema.TurnDetail{{
			Index: 0,
			Role:  schema.RoleUser,
			ContentParts: []schema.TurnContentPart{
				{Type: schema.TurnPartTypeImage, Image: &schema.ImagePart{OmittedReason: schema.OmittedReasonNotPublished, Ref: strings.Repeat("b", 64)}},
			},
		}},
	}
	leakedEnvelope := schema.TranscriptContent{ContractVersion: "1.0.0", Kind: schema.ContentKindSessionDetail, SessionDetail: &leaked}
	if err := schema.ValidateTranscriptContent(leakedEnvelope); err == nil {
		t.Fatal("transcript envelope accepted a published image part carrying a local ref")
	}
}
