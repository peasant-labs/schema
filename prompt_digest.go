package schema

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// DigestItemKind is the closed set of item kinds in a PromptDigest chain.
type DigestItemKind string

const (
	// DigestItemSession marks where one recorded session began.
	DigestItemSession DigestItemKind = "session"
	// DigestItemPrompt is one human-typed user turn.
	DigestItemPrompt DigestItemKind = "prompt"
	// DigestItemSkill marks a skill or user slash-command invocation at the
	// position it happened.
	DigestItemSkill DigestItemKind = "skill"
	// DigestItemCommit anchors the chain to a commit in the pull request.
	DigestItemCommit DigestItemKind = "commit"
)

// AllDigestItemKinds is the canonical inventory, in the order a renderer
// documents them.
var AllDigestItemKinds = []DigestItemKind{DigestItemSession, DigestItemPrompt, DigestItemSkill, DigestItemCommit}

// IsValid reports whether k is a known kind.
func (k DigestItemKind) IsValid() bool {
	switch k {
	case DigestItemSession, DigestItemPrompt, DigestItemSkill, DigestItemCommit:
		return true
	}
	return false
}

func (k DigestItemKind) String() string { return string(k) }

// JSONSchema implements jsonschema.Exposer.
func (DigestItemKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Digest Item Kind",
		"Kind of one item in the prompt digest chain: a session boundary, a human prompt, a skill invocation marker, or a commit anchor",
		AllDigestItemKinds,
	), nil
}

// PromptDigestHeader summarises the complete chain. Counts describe the whole
// digest, never a rendered tier.
type PromptDigestHeader struct {
	SessionCount   int     `json:"sessionCount"`
	PromptCount    int     `json:"promptCount"`
	CommitsCovered int     `json:"commitsCovered"`
	CommitsTotal   int     `json:"commitsTotal"`
	Harness        Harness `json:"harness"`
	RedactionLevel string  `json:"redactionLevel"`
	VillageURL     string  `json:"villageUrl"`
}

// PromptDigestSkill is one distinct skill or user slash command across every
// attached session, with how many times it was invoked. Name is a slash-prefixed
// skill or user command as recorded by the harness, or a bare plugin identifier
// for a tool provider observed through tool calls. Renderers distinguish the two
// by the leading slash.
type PromptDigestSkill struct {
	Name            string `json:"name"`
	InvocationCount int    `json:"invocationCount"`
}

// PromptDigestItem is one entry in the chronological chain. Which optional
// fields are set depends on Kind; Validate states the rule for each kind.
type PromptDigestItem struct {
	Kind         DigestItemKind `json:"kind"`
	TranscriptID TranscriptID   `json:"transcriptId"`
	Timestamp    time.Time      `json:"timestamp"`
	Text         string         `json:"text"`
	// TurnIndex deep-links a prompt or skill marker into the transcript viewer.
	TurnIndex *int `json:"turnIndex,omitempty"`
	// Ordinal is the 1-based prompt number a reviewer sees. Prompts only.
	Ordinal *int `json:"ordinal,omitempty"`
	// CommitSHA is the full or abbreviated SHA of a commit anchor. Commits only.
	CommitSHA string `json:"commitSha,omitempty"`
	// PromptCount and CommitCount summarise a session boundary. Sessions only.
	PromptCount *int `json:"promptCount,omitempty"`
	CommitCount *int `json:"commitCount,omitempty"`
	// Additions, Deletions, and FilesChanged are the lines added, the lines
	// deleted, and the number of files a commit touched, as the hosting
	// provider reports them. Commits only, and all three together or not at
	// all; they are absent when the counts were not read.
	Additions    *int `json:"additions,omitempty"`
	Deletions    *int `json:"deletions,omitempty"`
	FilesChanged *int `json:"filesChanged,omitempty"`
}

// PromptDigest is the reviewer-facing projection of the prompts behind a pull
// request. It is always the complete chain across every attached transcript;
// the comment and check-run budgets described in the design are applied by the
// renderer, never by dropping items from this type.
type PromptDigest struct {
	Header PromptDigestHeader  `json:"header"`
	Skills []PromptDigestSkill `json:"skills" nullable:"false"`
	Items  []PromptDigestItem  `json:"items" nullable:"false"`
}

var digestCommitSHAPattern = regexp.MustCompile(`^[0-9a-f]{7,40}$`)

// Validate enforces the per-kind field rules.
func (i PromptDigestItem) Validate() error {
	const where = "prompt digest item validation failed at schema.PromptDigestItem.Validate: "
	if !i.Kind.IsValid() {
		return fmt.Errorf(where+"the kind %q is outside the closed set %v, so a renderer cannot place the item; emit one of the known kinds", i.Kind, AllDigestItemKinds)
	}
	if _, err := NewTranscriptID(i.TranscriptID.String()); err != nil {
		return fmt.Errorf(where+"the item cannot link to its transcript: %w", err)
	}
	if i.Timestamp.IsZero() {
		return fmt.Errorf(where + "a zero timestamp cannot be ordered in the chain; supply the turn or commit time")
	}
	if i.Kind != DigestItemCommit && (i.Additions != nil || i.Deletions != nil || i.FilesChanged != nil) {
		return fmt.Errorf(where+"only a commit anchor carries additions, deletions, and filesChanged; a %s item describes no commit", i.Kind)
	}
	switch i.Kind {
	case DigestItemPrompt:
		if i.TurnIndex == nil || *i.TurnIndex < 0 {
			return fmt.Errorf(where + "a prompt needs a non-negative turnIndex to deep-link into the transcript")
		}
		if i.Ordinal == nil || *i.Ordinal < 1 {
			return fmt.Errorf(where + "a prompt needs a positive ordinal, the 1-based number shown to reviewers")
		}
		if strings.TrimSpace(i.Text) == "" {
			return fmt.Errorf(where + "a prompt needs non-empty text")
		}
		if i.CommitSHA != "" || i.PromptCount != nil || i.CommitCount != nil {
			return fmt.Errorf(where + "a prompt carries no commitSha, promptCount, or commitCount")
		}
	case DigestItemSkill:
		if i.TurnIndex == nil || *i.TurnIndex < 0 {
			return fmt.Errorf(where + "a skill marker needs a non-negative turnIndex to deep-link into the transcript")
		}
		if len(i.Text) < 2 || i.Text[0] != '/' {
			return fmt.Errorf(where + "a skill marker's text is the slash-prefixed invocation as recorded")
		}
		if i.Ordinal != nil {
			return fmt.Errorf(where + "a skill marker carries no ordinal; only prompts are numbered")
		}
		if i.CommitSHA != "" || i.PromptCount != nil || i.CommitCount != nil {
			return fmt.Errorf(where + "a skill marker carries no commitSha, promptCount, or commitCount")
		}
	case DigestItemCommit:
		if !digestCommitSHAPattern.MatchString(i.CommitSHA) {
			return fmt.Errorf(where+"commitSha %q must be 7 to 40 lowercase hex characters", i.CommitSHA)
		}
		if i.TurnIndex != nil {
			return fmt.Errorf(where + "a commit anchor carries no turnIndex; it links to the commit, not to a turn")
		}
		if i.Ordinal != nil || i.PromptCount != nil || i.CommitCount != nil {
			return fmt.Errorf(where + "a commit anchor carries no ordinal, promptCount, or commitCount")
		}
		supplied := 0
		for _, count := range []*int{i.Additions, i.Deletions, i.FilesChanged} {
			if count == nil {
				continue
			}
			if *count < 0 {
				return fmt.Errorf(where+"a change count cannot be negative; got %d for additions, deletions, or filesChanged", *count)
			}
			supplied++
		}
		if supplied != 0 && supplied != 3 {
			return fmt.Errorf(where + "a commit anchor carries additions, deletions, and filesChanged together or not at all; a partial set would read as a zero the commit does not have")
		}
	case DigestItemSession:
		if i.PromptCount == nil || *i.PromptCount < 0 || i.CommitCount == nil || *i.CommitCount < 0 {
			return fmt.Errorf(where + "a session boundary needs a non-negative promptCount and commitCount")
		}
		if strings.TrimSpace(i.Text) == "" {
			return fmt.Errorf(where + "a session boundary needs non-empty text")
		}
		if i.TurnIndex != nil || i.Ordinal != nil || i.CommitSHA != "" {
			return fmt.Errorf(where + "a session boundary carries no turnIndex, ordinal, or commitSha")
		}
	}
	return nil
}

// Validate checks every item, the chronological order of the chain, the header
// skills, and that the header counts describe the complete chain.
func (d PromptDigest) Validate() error {
	const where = "prompt digest validation failed at schema.PromptDigest.Validate: "
	seen := make(map[string]struct{}, len(d.Skills))
	for index, skill := range d.Skills {
		if skill.Name == "" {
			return fmt.Errorf(where+"skills[%d] has an empty name; a header entry names a slash-prefixed skill or user command, or a bare plugin identifier", index)
		}
		for _, r := range skill.Name {
			if unicode.IsSpace(r) || unicode.IsControl(r) {
				return fmt.Errorf(where+"skills[%d] name %q contains whitespace or a control character; a header entry is one token", index, skill.Name)
			}
		}
		if skill.Name == "/" {
			return fmt.Errorf(where+"skills[%d] name is a bare slash and names no command; supply the command token after the slash", index)
		}
		if skill.InvocationCount < 1 {
			return fmt.Errorf(where+"skills[%d] %q has invocationCount %d; a listed skill was invoked at least once", index, skill.Name, skill.InvocationCount)
		}
		if _, repeated := seen[skill.Name]; repeated {
			return fmt.Errorf(where+"skills[%d] is a duplicate header entry for %q; the header names each skill once and carries its complete invocationCount", index, skill.Name)
		}
		seen[skill.Name] = struct{}{}
	}
	prompts, sessions := 0, 0
	skillItemCounts := map[string]int{}
	distinctCommitSHAs := map[string]struct{}{}
	var previous time.Time
	for index, item := range d.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf(where+"items[%d]: %w", index, err)
		}
		if index > 0 && item.Timestamp.Before(previous) {
			return fmt.Errorf(where+"items[%d] at %s precedes items[%d] at %s; the chain must be chronological", index, item.Timestamp.Format(time.RFC3339), index-1, previous.Format(time.RFC3339))
		}
		previous = item.Timestamp
		switch item.Kind {
		case DigestItemPrompt:
			prompts++
			if *item.Ordinal != prompts {
				return fmt.Errorf(where+"items[%d] is prompt %d in chain order but carries ordinal %d; ordinals run 1..promptCount in chain order", index, prompts, *item.Ordinal)
			}
		case DigestItemSession:
			sessions++
		case DigestItemSkill:
			skillItemCounts[item.Text]++
		case DigestItemCommit:
			distinctCommitSHAs[item.CommitSHA] = struct{}{}
		}
	}
	headerSkillNames := make(map[string]struct{}, len(d.Skills))
	for _, skill := range d.Skills {
		headerSkillNames[skill.Name] = struct{}{}
	}
	for index, item := range d.Items {
		if item.Kind != DigestItemSkill {
			continue
		}
		if _, listed := headerSkillNames[item.Text]; !listed {
			return fmt.Errorf(where+"items[%d] is a skill item %q that no header entry names; the header lists every invocation in the chain", index, item.Text)
		}
	}
	for index, skill := range d.Skills {
		if skill.Name[0] != '/' {
			continue
		}
		if skill.InvocationCount != skillItemCounts[skill.Name] {
			return fmt.Errorf(where+"skills[%d] %q declares invocationCount %d but the chain carries %d matching skill items; the header counts the complete chain", index, skill.Name, skill.InvocationCount, skillItemCounts[skill.Name])
		}
	}
	if d.Header.PromptCount != prompts {
		return fmt.Errorf(where+"header promptCount %d does not match the %d prompt items; the header describes the complete chain", d.Header.PromptCount, prompts)
	}
	if d.Header.SessionCount != sessions {
		return fmt.Errorf(where+"header sessionCount %d does not match the %d session items; the header describes the complete chain", d.Header.SessionCount, sessions)
	}
	if d.Header.CommitsCovered != len(distinctCommitSHAs) {
		return fmt.Errorf(where+"header commitsCovered %d does not match the %d distinct commit anchors in the chain; the header describes the complete chain", d.Header.CommitsCovered, len(distinctCommitSHAs))
	}
	if d.Header.CommitsCovered < 0 || d.Header.CommitsTotal < 0 || d.Header.CommitsCovered > d.Header.CommitsTotal {
		return fmt.Errorf(where+"header commitsCovered %d must be between 0 and commitsTotal %d", d.Header.CommitsCovered, d.Header.CommitsTotal)
	}
	return nil
}
