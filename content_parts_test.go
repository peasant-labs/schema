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

func TestContentPartDetailFixtures(t *testing.T) {
	fixtures, err := schema.LoadContentPartDetailFixtures()
	if err != nil {
		t.Fatalf("load content part detail fixtures: %v", err)
	}
	for _, fixtureCase := range fixtures.Cases.Cases {
		fixtureCase := fixtureCase
		t.Run(fixtureCase.Name, func(t *testing.T) {
			var detail schema.SessionDetailPayload
			if err := json.Unmarshal([]byte(fixtureCase.Input.Detail), &detail); err != nil {
				t.Fatalf("decode fixture detail: %v", err)
			}
			mode := schema.ImagePartMode(fixtureCase.Input.Mode)
			var err error
			switch mode {
			case schema.ImagePartModeLocal:
				err = schema.ValidateSessionDetailPayload(detail)
			case schema.ImagePartModePublished:
				err = schema.ValidatePublishedSessionDetailPayload(detail)
			default:
				t.Fatalf("fixture mode %q is not local or published", mode)
			}
			if (err == nil) != fixtureCase.Expected.Accepted {
				t.Fatalf("detail validator for mode %q accepted=%v, want %v; err=%v", mode, err == nil, fixtureCase.Expected.Accepted, err)
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
			if mode == schema.ImagePartModePublished {
				envelopeErr := schema.ValidateTranscriptContent(schema.TranscriptContent{ContractVersion: "1.0.0", Kind: schema.ContentKindSessionDetail, SessionDetail: &detail})
				if (envelopeErr == nil) != (err == nil) {
					t.Fatalf("publication envelope outcome %v disagrees with the published detail outcome %v", envelopeErr == nil, err == nil)
				}
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

	localEnvelope := schema.TranscriptContent{ContractVersion: "1.0.0", Kind: schema.ContentKindSessionDetail, SessionDetail: &local}
	if err := schema.ValidateTranscriptContent(localEnvelope); err == nil {
		t.Fatal("publication envelope accepted a local inline image part")
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
	if err := schema.ValidatePublishedSessionDetailPayload(published); err != nil {
		t.Fatalf("published detail validator rejected a not_published image: %v", err)
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
	if err := schema.ValidatePublishedSessionDetailPayload(leaked); err == nil {
		t.Fatal("published detail validator accepted a not_published image carrying a local ref")
	}
	leakedEnvelope := schema.TranscriptContent{ContractVersion: "1.0.0", Kind: schema.ContentKindSessionDetail, SessionDetail: &leaked}
	if err := schema.ValidateTranscriptContent(leakedEnvelope); err == nil {
		t.Fatal("transcript envelope accepted a published image part carrying a local ref")
	}
}
