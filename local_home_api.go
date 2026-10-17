package schema

import (
	"fmt"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// LocalHomePublicationFilter narrows the visible root cohort before pagination.
type LocalHomePublicationFilter string

const (
	LocalHomePublicationAll          LocalHomePublicationFilter = "all"
	LocalHomePublicationNotPublished LocalHomePublicationFilter = "not-published"
	LocalHomePublicationPublished    LocalHomePublicationFilter = "published"
	LocalHomePublicationAuto         LocalHomePublicationFilter = "auto"
)

var AllLocalHomePublicationFilters = []LocalHomePublicationFilter{LocalHomePublicationAll, LocalHomePublicationNotPublished, LocalHomePublicationPublished, LocalHomePublicationAuto}

func (f LocalHomePublicationFilter) IsValid() bool  { return inSet(f, AllLocalHomePublicationFilters) }
func (f LocalHomePublicationFilter) String() string { return string(f) }
func (LocalHomePublicationFilter) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Local Home Publication Filter", "Root publication filter; auto uses the canonical live hook decision", AllLocalHomePublicationFilters), nil
}

// LocalHomeSessionsQuery selects roots, or immediate children of a visible parent.
// Pointers distinguish omission from an explicitly invalid zero or empty value.
type LocalHomeSessionsQuery struct {
	Limit           *int                        `query:"limit" nullable:"false" minimum:"1" maximum:"100" default:"25"`
	Offset          *int                        `query:"offset" nullable:"false" minimum:"0" default:"0"`
	ParentSessionID *string                     `query:"parentSessionId" nullable:"false" minLength:"1"`
	ProjectHash     *ProjectHash                `query:"projectHash" nullable:"false"`
	Publication     *LocalHomePublicationFilter `query:"publication" nullable:"false" default:"all"`
}

func (q LocalHomeSessionsQuery) Validate() error {
	if q.Limit != nil && (*q.Limit < 1 || *q.Limit > 100) || q.Offset != nil && *q.Offset < 0 {
		return homeError("LocalHomeSessionsQuery", "limit or offset is out of bounds; request limit 1..100 and offset >=0")
	}
	if q.ParentSessionID != nil {
		if strings.TrimSpace(*q.ParentSessionID) == "" {
			return homeError("LocalHomeSessionsQuery", "parentSessionId is empty; supply a nonempty local session ID")
		}
		if q.ProjectHash != nil || q.Publication != nil {
			return homeError("LocalHomeSessionsQuery", "child request carries root filters; omit projectHash and publication to avoid hiding visible children")
		}
	}
	if q.ProjectHash != nil && q.ProjectHash.Validate() != nil {
		return homeError("LocalHomeSessionsQuery", "projectHash is malformed; supply a canonical project hash")
	}
	if q.Publication != nil && !q.Publication.IsValid() {
		return homeError("LocalHomeSessionsQuery", "publication is unknown; use all, not-published, published, or auto")
	}
	return nil
}

// LocalHomeSessionRow shares publication and sync projections with their existing reads.
// Title is the stored generated title; preview is a bounded redaction-safe first-user preview.
type LocalHomeSessionRow struct {
	Sync        LocalSyncSummary `json:"sync"`
	Publication LocalPublication `json:"publication"`
	Title       string           `json:"title"`
	Preview     string           `json:"preview"`
	ChildCount  int              `json:"childCount" minimum:"0"`
}

func (r LocalHomeSessionRow) Validate() error {
	if err := r.Sync.Validate(); err != nil {
		return err
	}
	if err := r.Publication.Validate(); err != nil {
		return err
	}
	if r.Sync.ID != r.Publication.SessionID || r.Publication.OutsideSelection || r.Publication.Audience != nil || r.ChildCount < 0 {
		return homeError("LocalHomeSessionRow", "sync/publication identity, selection, audience, or childCount is inconsistent; emit the same visible session ID, no audience, and a nonnegative immediate-child count")
	}
	return nil
}

// LocalHomeSessionsPage is ordered by startTime descending, then local ID ascending.
// Offset pages are live and do not promise snapshot consistency.
type LocalHomeSessionsPage struct {
	Rows       []LocalHomeSessionRow `json:"rows" nullable:"false"`
	Offset     int                   `json:"offset" minimum:"0"`
	Limit      int                   `json:"limit" minimum:"1" maximum:"100"`
	NextOffset *int                  `json:"nextOffset,omitempty" nullable:"false" minimum:"0"`
}

func (p LocalHomeSessionsPage) Validate() error {
	if p.Rows == nil || p.Offset < 0 || p.Limit < 1 || p.Limit > 100 || len(p.Rows) > p.Limit {
		return homeError("LocalHomeSessionsPage", "page bounds or rows is invalid; emit an array bounded by limit 1..100 and offset >=0")
	}
	if p.NextOffset != nil && (len(p.Rows) != p.Limit || p.Offset > int(^uint(0)>>1)-p.Limit || *p.NextOffset != p.Offset+p.Limit) {
		return homeError("LocalHomeSessionsPage", "nextOffset is inconsistent; emit offset+limit only for a full page with another matching metadata row")
	}
	seen := make(map[string]bool, len(p.Rows))
	for _, r := range p.Rows {
		if err := r.Validate(); err != nil {
			return err
		}
		if seen[r.Sync.ID] {
			return homeError("LocalHomeSessionsPage", "rows repeat a session; emit each ID once per page")
		}
		seen[r.Sync.ID] = true
	}
	return nil
}

func homeError(location, reason string) error {
	return fmt.Errorf("Home contract validation failed at schema.%s.Validate: %s; the caller must correct the request or response before using it", location, reason)
}
