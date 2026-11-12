package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// queueFile is the raw GraphQL merge-queue response the workflow writes to a
// file. The tool reads the file and nothing else, so it keeps no network
// access.
type queueFile struct {
	Data struct {
		Repository struct {
			MergeQueue *struct {
				Entries struct {
					TotalCount int `json:"totalCount"`
					PageInfo   *struct {
						HasNextPage bool `json:"hasNextPage"`
					} `json:"pageInfo"`
					Nodes []queueNode `json:"nodes"`
				} `json:"entries"`
			} `json:"mergeQueue"`
		} `json:"repository"`
	} `json:"data"`
	Errors []any `json:"errors"`
	// Alternate flat shape some writers emit: {"entries": {...}, "errors": ...}.
	Entries *queueEntries `json:"entries"`
}

type queueEntries struct {
	TotalCount int         `json:"totalCount"`
	HasMore    *bool       `json:"hasMore"`
	PageInfo   *pageInfo   `json:"pageInfo"`
	Nodes      []queueNode `json:"nodes"`
}

type pageInfo struct {
	HasNextPage bool `json:"hasNextPage"`
}

type queueNode struct {
	Position    int `json:"position"`
	PullRequest *struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	} `json:"pullRequest"`
	// Flat alternate: {"number": N, "title": "..."}.
	Number int    `json:"number"`
	Title  string `json:"title"`
}

func (n queueNode) prNumber() int {
	if n.PullRequest != nil && n.PullRequest.Number != 0 {
		return n.PullRequest.Number
	}
	return n.Number
}

func (n queueNode) prTitle() string {
	if n.PullRequest != nil && n.PullRequest.Title != "" {
		return n.PullRequest.Title
	}
	return n.Title
}

// decodeQueueFile parses the raw GraphQL body. It returns an error (exit 1,
// no gate line) for unreadable or malformed JSON, a non-empty errors array, a
// null merge queue, or a missing totalCount. A complete page whose count does
// not match its nodes is also malformed.
func decodeQueueFile(data []byte) (nodes []queueNode, total int, hasMore bool, err error) {
	var q queueFile
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, 0, false, fmt.Errorf("queue-gate: cannot parse %d bytes as JSON: %v. The workflow must write the raw GraphQL mergeQueue response body to the file", len(data), err)
	}
	if len(q.Errors) > 0 {
		return nil, 0, false, fmt.Errorf("queue-gate: the GraphQL response carries %d error(s): the queue read failed. Fix the query or permissions and retry; never treat this as an empty queue", len(q.Errors))
	}
	// Flat shape takes precedence when present.
	if q.Entries != nil {
		if q.Entries.Nodes == nil {
			return nil, 0, false, fmt.Errorf("queue-gate: the queue file has entries but no nodes list: missing entries.nodes. The workflow must write the raw GraphQL response with entries { totalCount nodes }")
		}
		more := false
		if q.Entries.HasMore != nil {
			more = *q.Entries.HasMore
		}
		if q.Entries.PageInfo != nil {
			more = q.Entries.PageInfo.HasNextPage
		}
		if !more && q.Entries.TotalCount != len(q.Entries.Nodes) {
			return nil, 0, false, fmt.Errorf("queue-gate: entries.totalCount is %d but %d node(s) are listed on a complete page: the queue snapshot is inconsistent. Re-read the queue and retry", q.Entries.TotalCount, len(q.Entries.Nodes))
		}
		return q.Entries.Nodes, q.Entries.TotalCount, more, nil
	}
	mq := q.Data.Repository.MergeQueue
	if mq == nil {
		return nil, 0, false, fmt.Errorf("queue-gate: mergeQueue is null in the GraphQL response: the queue read failed (wrong branch, missing permission, or API error). Never treat this as an empty queue")
	}
	more := false
	if mq.Entries.PageInfo != nil {
		more = mq.Entries.PageInfo.HasNextPage
	}
	// Missing totalCount decodes as 0 with nodes present: distinguish by raw
	// presence check.
	if !hasTotalCount(data) {
		return nil, 0, false, fmt.Errorf("queue-gate: entries.totalCount is missing in the GraphQL response: the queue read is incomplete. Re-read the queue with totalCount and retry")
	}
	if !more && mq.Entries.TotalCount != len(mq.Entries.Nodes) {
		return nil, 0, false, fmt.Errorf("queue-gate: entries.totalCount is %d but %d node(s) are listed on a complete page: the queue snapshot is inconsistent. Re-read the queue and retry", mq.Entries.TotalCount, len(mq.Entries.Nodes))
	}
	return mq.Entries.Nodes, mq.Entries.TotalCount, more, nil
}

func hasTotalCount(data []byte) bool {
	return strings.Contains(string(data), "totalCount")
}

// gateDecision applies the hold rule: hold when an entry ahead of release PR N
// has a releasing title or a title that does not parse, or when the file says
// the queue holds more entries than it lists.
func gateDecision(nodes []queueNode, total int, hasMore bool, releasePR int) string {
	if hasMore || total > len(nodes) {
		return "hold"
	}
	releasePos := -1
	for _, n := range nodes {
		if n.prNumber() == releasePR {
			releasePos = n.Position
			break
		}
	}
	// When the release PR carries no position, fall back to list order: every
	// entry ahead in the file is ahead in the queue.
	for idx, n := range nodes {
		pos := n.Position
		if releasePos == -1 {
			if n.prNumber() == releasePR {
				continue
			}
			// Entries before the release PR in file order are ahead when
			// positions are absent (position 0 for all).
			isAhead := idx < releaseIndex(nodes, releasePR)
			if releaseIndex(nodes, releasePR) == -1 {
				isAhead = true
			}
			if !isAhead {
				_ = pos
				continue
			}
		} else {
			if n.prNumber() == releasePR {
				continue
			}
			if pos == 0 && n.Position == 0 {
				// No positions: use file order.
				if idx >= releaseIndex(nodes, releasePR) && releaseIndex(nodes, releasePR) != -1 {
					continue
				}
			} else if pos >= releasePos {
				continue
			}
		}
		releasing, parsable := parseTitleKind(n.prTitle())
		if !parsable || releasing {
			return "hold"
		}
	}
	return "arm"
}

func releaseIndex(nodes []queueNode, releasePR int) int {
	for i, n := range nodes {
		if n.prNumber() == releasePR {
			return i
		}
	}
	return -1
}

// runQueueGate implements: release-intent queue-gate --queue FILE --release-pr N.
func runQueueGate(args []string) error {
	var queuePath string
	var releasePR int
	var haveQueue, havePR bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--queue":
			val, next, err := requireFlagValue(args, i, "--queue")
			if err != nil {
				return err
			}
			queuePath = val
			haveQueue = true
			i = next
		case "--release-pr":
			val, next, err := requireFlagValue(args, i, "--release-pr")
			if err != nil {
				return err
			}
			var n int
			if _, err := fmt.Sscanf(val, "%d", &n); err != nil || n <= 0 {
				return usageErr("usage: release-intent queue-gate --queue FILE --release-pr N (bad --release-pr %q): pass the release PR number as a positive integer", val)
			}
			releasePR = n
			havePR = true
			i = next
		default:
			return usageErr("unknown flag %q for queue-gate: expected --queue FILE --release-pr N", args[i])
		}
	}
	if !haveQueue {
		return usageErr("usage: release-intent queue-gate --queue FILE --release-pr N (missing required --queue): pass the file holding the raw GraphQL mergeQueue response")
	}
	if !havePR {
		return usageErr("usage: release-intent queue-gate --queue FILE --release-pr N (missing required --release-pr): pass the release PR number")
	}
	data, err := os.ReadFile(queuePath)
	if err != nil {
		return fmt.Errorf("queue-gate: cannot read queue file %q: %v. The workflow must write the raw GraphQL response to this path before calling queue-gate", queuePath, err)
	}
	nodes, total, hasMore, err := decodeQueueFile(data)
	if err != nil {
		// Exit 1 with no gate line: nothing is armed and nothing is disarmed.
		return err
	}
	return emitLines([]string{"gate=" + gateDecision(nodes, total, hasMore, releasePR)})
}
