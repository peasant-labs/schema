package intent

import (
	"encoding/json"
	"strconv"
	"strings"
)

// This file holds the merge-queue hold rule: the raw GraphQL response
// decoder and the gate over the decoded entries.
//
// The decoder reads the queue file the workflow wrote and nothing else, so
// the tool keeps no network access. It fails closed: an unreadable answer,
// a service-side error, a null queue, or a count the nodes do not match is
// an error with no gate line, never an empty queue.

// queueNode is one decoded queue entry; pointers tell a missing field apart
// from a zero value so either fails the read.
type queueNode struct {
	Position    *int `json:"position"`
	PullRequest *struct {
		Number *int    `json:"number"`
		Title  *string `json:"title"`
	} `json:"pullRequest"`
}

// queueResponse is the GraphQL answer shape the workflow queries.
type queueResponse struct {
	Data *struct {
		Repository *struct {
			MergeQueue *struct {
				Entries *struct {
					TotalCount *int        `json:"totalCount"`
					Nodes      []queueNode `json:"nodes"`
				} `json:"entries"`
			} `json:"mergeQueue"`
		} `json:"repository"`
	} `json:"data"`
	Errors []any `json:"errors"`
}

// queueFailure builds a fail-closed decoder error: no gate value exists, so
// the job must neither arm nor disarm on this read.
func queueFailure(what, why string) error {
	return inputError(
		what, why,
		"the merge-queue response file, decoded by intent.DecodeQueueResponse",
		"reading the merge queue before arming the release PR",
		"no gate value exists, so the release PR is neither armed nor disarmed on this run",
		"retry the run; a later run reads the queue again",
	)
}

// DecodeQueueResponse decodes a raw GraphQL merge-queue response body into
// its entries and the total the service reported.
func DecodeQueueResponse(data []byte) (QueueSnapshot, error) {
	var raw queueResponse
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	if err := decoder.Decode(&raw); err != nil {
		return QueueSnapshot{}, queueFailure(
			"the queue response is not valid JSON: "+err.Error(),
			"the gate judges the queue file the workflow wrote, and a malformed file judges nothing",
		)
	}
	if len(raw.Errors) > 0 {
		return QueueSnapshot{}, queueFailure(
			"the queue response carries "+strconv.Itoa(len(raw.Errors))+" service error(s)",
			"a response with errors answers nothing about the queue",
		)
	}
	if raw.Data == nil || raw.Data.Repository == nil || raw.Data.Repository.MergeQueue == nil {
		return QueueSnapshot{}, queueFailure(
			"the queue response has no merge queue",
			"a null queue is a failed read, not an empty queue",
		)
	}
	entries := raw.Data.Repository.MergeQueue.Entries
	if entries == nil || entries.TotalCount == nil {
		return QueueSnapshot{}, queueFailure(
			"the queue response has no entry total",
			"without the total the gate cannot tell a complete page from a cut one",
		)
	}
	if entries.Nodes == nil {
		return QueueSnapshot{}, queueFailure(
			"the queue response has no entry list",
			"without the nodes the gate cannot judge what is ahead",
		)
	}
	if *entries.TotalCount != len(entries.Nodes) {
		return QueueSnapshot{}, queueFailure(
			"the queue response lists "+strconv.Itoa(len(entries.Nodes))+" entries for a total of "+strconv.Itoa(*entries.TotalCount),
			"a count the nodes do not match is a cut page, and arming on a cut page can skip a releasing PR",
		)
	}
	out := QueueSnapshot{Entries: make([]QueueEntry, 0, len(entries.Nodes)), TotalCount: *entries.TotalCount}
	for i, node := range entries.Nodes {
		if node.Position == nil || node.PullRequest == nil || node.PullRequest.Number == nil || node.PullRequest.Title == nil {
			return QueueSnapshot{}, queueFailure(
				"queue entry "+strconv.Itoa(i)+" is missing its position, number, or title",
				"every entry needs all three before the gate can place the release PR",
			)
		}
		out.Entries = append(out.Entries, QueueEntry{
			Position: *node.Position,
			Number:   *node.PullRequest.Number,
			Title:    *node.PullRequest.Title,
		})
	}
	return out, nil
}

// QueueGate applies the hold rule to a decoded queue read. The release PR is
// held while a PR with a releasing title, or a title that does not parse, is
// ahead of it. An entry is ahead when its position is lower than the release
// PR's; when the release PR is not in the queue, every entry is ahead.
// Otherwise the gate arms.
func QueueGate(cfg Config, current SemVer, entries []QueueEntry, releasePR int) (Gate, error) {
	if _, err := ParseSemVer(string(current)); err != nil {
		return "", err
	}
	releasePos := -1
	for _, entry := range entries {
		if entry.Number == releasePR {
			releasePos = entry.Position
			break
		}
	}
	for _, entry := range entries {
		if releasePos >= 0 && entry.Position >= releasePos {
			continue
		}
		title, err := ParseTitle(cfg, current, entry.Title)
		if err != nil {
			if _, bad := err.(*Error); bad {
				return GateHold, nil
			}
			return "", err
		}
		if title.Level != LevelNone {
			return GateHold, nil
		}
	}
	return GateArm, nil
}
