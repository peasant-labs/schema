package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// UnverifiedPullRequestSource locates the original call in the main transcript.
type UnverifiedPullRequestSource struct {
	TurnIndex     int    `json:"turnIndex" minimum:"0" maximum:"9007199254740991"`
	ToolCallIndex int    `json:"toolCallIndex" minimum:"0" maximum:"9007199254740991"`
	ToolCallID    string `json:"toolCallId,omitempty" minLength:"1"`
}

// UnverifiedPullRequest is evidence from a recorded command, not a verified attachment.
type UnverifiedPullRequest struct {
	Owner  string                      `json:"owner" minLength:"1"`
	Repo   string                      `json:"repo" minLength:"1"`
	Number int                         `json:"number" minimum:"1" maximum:"9007199254740991"`
	URL    string                      `json:"url" minLength:"1"`
	Source UnverifiedPullRequestSource `json:"source"`
}

var githubPRSegment = regexp.MustCompile(`^[A-Za-z0-9_-][A-Za-z0-9._-]*$`)

// Validate checks the canonical URL and exact, safe source coordinates.
func (p UnverifiedPullRequest) Validate() error {
	if !githubPRSegment.MatchString(p.Owner) || !githubPRSegment.MatchString(p.Repo) || p.Owner == "." || p.Owner == ".." || p.Repo == "." || p.Repo == ".." || p.Number < 1 || int64(p.Number) > maxSafeJSONInteger || p.URL != "https://github.com/"+p.Owner+"/"+p.Repo+"/pull/"+strconv.Itoa(p.Number) {
		return fmt.Errorf("unverified pull request validation failed at schema.UnverifiedPullRequest.Validate: invalid GitHub identity or canonical URL; emit the exact owner, repository and positive pull number from the recorded result")
	}
	if p.Source.TurnIndex < 0 || int64(p.Source.TurnIndex) > maxSafeJSONInteger || p.Source.ToolCallIndex < 0 || int64(p.Source.ToolCallIndex) > maxSafeJSONInteger || !utf8.ValidString(p.Source.ToolCallID) {
		return fmt.Errorf("unverified pull request validation failed at schema.UnverifiedPullRequest.Validate: invalid source locator; emit safe nonnegative turn and call indices and a valid UTF-8 call ID")
	}
	return nil
}

func validateUnverifiedPullRequests(rows *[]UnverifiedPullRequest) error {
	if rows == nil {
		return nil
	}
	if *rows == nil {
		return fmt.Errorf("unverified pull requests validation failed: null array; omit unsupported projections or emit an initialized array")
	}
	for _, row := range *rows {
		if err := row.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalJSON rejects explicit null on an optional source ID.
func (s *UnverifiedPullRequestSource) UnmarshalJSON(data []byte) error {
	type plain UnverifiedPullRequestSource
	if err := rejectOptionalNull(data, "toolCallId"); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, key := range []string{"turnIndex", "toolCallIndex"} {
		value, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("source locator decoding failed at schema.UnverifiedPullRequestSource.UnmarshalJSON: %s is missing or null; provide both exact nonnegative coordinates", key)
		}
	}
	if id, ok := fields["toolCallId"]; ok && bytes.Equal(bytes.TrimSpace(id), []byte(`""`)) {
		return fmt.Errorf("source locator decoding failed at schema.UnverifiedPullRequestSource.UnmarshalJSON: toolCallId is empty; omit an absent ID or provide the exact recorded ID")
	}
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = UnverifiedPullRequestSource(value)
	return nil
}

func rejectOptionalNull(data []byte, key string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for name, value := range fields {
		if strings.EqualFold(name, key) && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("read validation failed at schema JSON decoding: %s is null; omit it when unsupported or provide its nonnull value", key)
		}
	}
	return nil
}

// UnmarshalJSON preserves omission while rejecting an explicit null array.
func (r *SessionPullRequestsResponse) UnmarshalJSON(data []byte) error {
	type plain SessionPullRequestsResponse
	if err := rejectOptionalNull(data, "unverified_pull_requests"); err != nil {
		return err
	}
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = SessionPullRequestsResponse(value)
	return nil
}

// UnmarshalJSON preserves omission while rejecting an explicit null array.
func (r *VillageTranscriptPullRequestsResponse) UnmarshalJSON(data []byte) error {
	type plain VillageTranscriptPullRequestsResponse
	if err := rejectOptionalNull(data, "unverified_pull_requests"); err != nil {
		return err
	}
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = VillageTranscriptPullRequestsResponse(value)
	return nil
}

// UnmarshalJSON requires a concrete locator instead of treating absent source as
// an invented zero-index call.
func (p *UnverifiedPullRequest) UnmarshalJSON(data []byte) error {
	type plain UnverifiedPullRequest
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	source, ok := fields["source"]
	if !ok || bytes.Equal(bytes.TrimSpace(source), []byte("null")) {
		return fmt.Errorf("unverified pull request decoding failed at schema.UnverifiedPullRequest.UnmarshalJSON: source is missing or null; callers cannot locate recorded evidence; provide the source turn and tool call indices")
	}
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = UnverifiedPullRequest(value)
	return nil
}
