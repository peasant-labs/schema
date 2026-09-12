package schema_test

import (
	_ "embed"
	"strings"
	"testing"
	"time"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
)

//go:embed testdata/pulls/prompt_digest_items.yaml
var digestItemFixtureYAML []byte

//go:embed testdata/pulls/prompt_digest_items_manifest.yaml
var digestItemManifestYAML []byte

//go:embed testdata/pulls/prompt_digests.yaml
var digestFixtureYAML []byte

//go:embed testdata/pulls/prompt_digests_manifest.yaml
var digestManifestYAML []byte

type digestItemFixtureInput struct {
	Kind         string    `yaml:"kind"`
	TranscriptID string    `yaml:"transcriptId"`
	Timestamp    time.Time `yaml:"timestamp"`
	Text         string    `yaml:"text,omitempty"`
	TurnIndex    *int      `yaml:"turnIndex,omitempty"`
	Ordinal      *int      `yaml:"ordinal,omitempty"`
	CommitSha    string    `yaml:"commitSha,omitempty"`
	PromptCount  *int      `yaml:"promptCount,omitempty"`
	CommitCount  *int      `yaml:"commitCount,omitempty"`
	Additions    *int      `yaml:"additions,omitempty"`
	Deletions    *int      `yaml:"deletions,omitempty"`
	FilesChanged *int      `yaml:"filesChanged,omitempty"`
}

func (in digestItemFixtureInput) toItem() schema.PromptDigestItem {
	return schema.PromptDigestItem{
		Kind:         schema.DigestItemKind(in.Kind),
		TranscriptID: schema.TranscriptID(in.TranscriptID),
		Timestamp:    in.Timestamp,
		Text:         in.Text,
		TurnIndex:    in.TurnIndex,
		Ordinal:      in.Ordinal,
		CommitSHA:    in.CommitSha,
		PromptCount:  in.PromptCount,
		CommitCount:  in.CommitCount,
		Additions:    in.Additions,
		Deletions:    in.Deletions,
		FilesChanged: in.FilesChanged,
	}
}

type digestHeaderFixtureInput struct {
	SessionCount   int    `yaml:"sessionCount"`
	PromptCount    int    `yaml:"promptCount"`
	CommitsCovered int    `yaml:"commitsCovered"`
	CommitsTotal   int    `yaml:"commitsTotal"`
	Harness        string `yaml:"harness"`
	RedactionLevel string `yaml:"redactionLevel"`
	VillageURL     string `yaml:"villageUrl"`
}

type digestSkillFixtureInput struct {
	Name            string `yaml:"name"`
	InvocationCount int    `yaml:"invocationCount"`
}

type digestFixtureInput struct {
	Header digestHeaderFixtureInput  `yaml:"header"`
	Skills []digestSkillFixtureInput `yaml:"skills"`
	Items  []digestItemFixtureInput  `yaml:"items"`
}

func (in digestFixtureInput) toDigest() schema.PromptDigest {
	digest := schema.PromptDigest{
		Header: schema.PromptDigestHeader{
			SessionCount:   in.Header.SessionCount,
			PromptCount:    in.Header.PromptCount,
			CommitsCovered: in.Header.CommitsCovered,
			CommitsTotal:   in.Header.CommitsTotal,
			Harness:        schema.Harness(in.Header.Harness),
			RedactionLevel: in.Header.RedactionLevel,
			VillageURL:     in.Header.VillageURL,
		},
		Skills: make([]schema.PromptDigestSkill, 0, len(in.Skills)),
		Items:  make([]schema.PromptDigestItem, 0, len(in.Items)),
	}
	for _, skill := range in.Skills {
		digest.Skills = append(digest.Skills, schema.PromptDigestSkill{Name: skill.Name, InvocationCount: skill.InvocationCount})
	}
	for _, item := range in.Items {
		digest.Items = append(digest.Items, item.toItem())
	}
	return digest
}

type digestFixtureExpected struct {
	Accepted      bool   `yaml:"accepted"`
	ErrorContains string `yaml:"errorContains,omitempty"`
}

func TestPromptDigestItemFixture(t *testing.T) {
	corpus, err := testcase.LoadCorpus[digestItemFixtureInput, digestFixtureExpected](digestItemFixtureYAML)
	if err != nil {
		t.Fatalf("load digest item corpus: %v", err)
	}
	manifest, err := decodeTurnModelFixtureManifest(digestItemManifestYAML)
	if err != nil {
		t.Fatalf("load digest item manifest: %v", err)
	}
	if err := validateCorpusInventory("digest item", corpus, manifest); err != nil {
		t.Fatalf("validate digest item inventory: %v", err)
	}
	for _, fixtureCase := range corpus.Cases {
		fixtureCase := fixtureCase
		t.Run(fixtureCase.Name, func(t *testing.T) {
			assertValidation(t, fixtureCase.Input.toItem().Validate(), fixtureCase.Expected)
		})
	}
}

func TestPromptDigestFixture(t *testing.T) {
	corpus, err := testcase.LoadCorpus[digestFixtureInput, digestFixtureExpected](digestFixtureYAML)
	if err != nil {
		t.Fatalf("load digest corpus: %v", err)
	}
	manifest, err := decodeTurnModelFixtureManifest(digestManifestYAML)
	if err != nil {
		t.Fatalf("load digest manifest: %v", err)
	}
	if err := validateCorpusInventory("digest", corpus, manifest); err != nil {
		t.Fatalf("validate digest inventory: %v", err)
	}
	for _, fixtureCase := range corpus.Cases {
		fixtureCase := fixtureCase
		t.Run(fixtureCase.Name, func(t *testing.T) {
			assertValidation(t, fixtureCase.Input.toDigest().Validate(), fixtureCase.Expected)
		})
	}
}

func TestDigestItemKindClosedSet(t *testing.T) {
	for _, kind := range schema.AllDigestItemKinds {
		if !kind.IsValid() {
			t.Errorf("AllDigestItemKinds member %q is not valid", kind)
		}
	}
	if schema.DigestItemKind("paragraph").IsValid() {
		t.Error("an unknown kind must not be valid")
	}
}

func assertValidation(t *testing.T, err error, expected digestFixtureExpected) {
	t.Helper()
	if (err == nil) != expected.Accepted {
		t.Fatalf("Validate err=%v, want accepted=%v", err, expected.Accepted)
	}
	if expected.Accepted {
		return
	}
	if expected.ErrorContains == "" {
		t.Fatal("must-fail case declares no errorContains needle")
	}
	if !strings.Contains(err.Error(), expected.ErrorContains) {
		t.Fatalf("error %q does not contain %q", err, expected.ErrorContains)
	}
}
