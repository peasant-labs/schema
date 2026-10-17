package schema

import (
	"fmt"
	"math"
	"strings"
	"time"

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

type LocalHomeStatsQuery struct {
	ProjectHash *ProjectHash `query:"projectHash" nullable:"false" description:"Narrow rootFacets only, not global roots, projects, or activity"`
}

func (q LocalHomeStatsQuery) Validate() error {
	if q.ProjectHash != nil && q.ProjectHash.Validate() != nil {
		return homeError("LocalHomeStatsQuery", "projectHash is malformed; supply a canonical project hash")
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
// ScopeRevision hashes effective selection and signed-in origin/owner, not data;
// offset pages are live and do not promise snapshot consistency.
type LocalHomeSessionsPage struct {
	Rows          []LocalHomeSessionRow `json:"rows" nullable:"false"`
	Offset        int                   `json:"offset" minimum:"0"`
	Limit         int                   `json:"limit" minimum:"1" maximum:"100"`
	NextOffset    *int                  `json:"nextOffset,omitempty" nullable:"false" minimum:"0"`
	ScopeRevision string                `json:"scopeRevision" minLength:"1"`
}

func (p LocalHomeSessionsPage) Validate() error {
	if p.Rows == nil || p.Offset < 0 || p.Limit < 1 || p.Limit > 100 || len(p.Rows) > p.Limit || strings.TrimSpace(p.ScopeRevision) == "" {
		return homeError("LocalHomeSessionsPage", "page bounds, rows, or scopeRevision is invalid; emit an array bounded by limit 1..100, offset >=0, and a nonempty scope revision")
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

// LocalHomeRootCounts contains account-scoped receipt counts, never auto-hook counts.
type LocalHomeRootCounts struct {
	All          int `json:"all" minimum:"0"`
	Published    int `json:"published" minimum:"0"`
	NotPublished int `json:"notPublished" minimum:"0"`
}

func (c LocalHomeRootCounts) Validate() error {
	if c.All < 0 || c.Published < 0 || c.NotPublished < 0 || c.Published > c.All || c.NotPublished != c.All-c.Published {
		return homeError("LocalHomeRootCounts", "counts disagree; emit nonnegative counts with all = published + notPublished")
	}
	return nil
}

type LocalHomeProject struct {
	ProjectHash ProjectHash `json:"projectHash"`
	ProjectName string      `json:"projectName"`
}

// LocalHomeWeek starts on a UTC Monday.
type LocalHomeWeek struct {
	StartDate string `json:"startDate" pattern:"^[0-9]{4}-[0-9]{2}-[0-9]{2}$"`
	Sessions  int    `json:"sessions" minimum:"0"`
}

// LocalHomeActivity reduces the selection-only activity cohort, including children
// and nonpushable sessions. Empty median is null; zero is a real observation.
type LocalHomeActivity struct {
	ThisWeek              int             `json:"thisWeek" minimum:"0"`
	LongestStreakWeeks    int             `json:"longestStreakWeeks" minimum:"0"`
	MedianDurationMinutes *float64        `json:"medianDurationMinutes" minimum:"0"`
	Weeks                 []LocalHomeWeek `json:"weeks" nullable:"false" minItems:"8" maxItems:"8"`
}

func (a LocalHomeActivity) Validate() error {
	if a.ThisWeek < 0 || a.LongestStreakWeeks < 0 || len(a.Weeks) != 8 || a.MedianDurationMinutes != nil && (math.IsNaN(*a.MedianDurationMinutes) || math.IsInf(*a.MedianDurationMinutes, 0) || *a.MedianDurationMinutes < 0) {
		return homeError("LocalHomeActivity", "activity bounds are invalid; emit nonnegative counts, eight weeks, and a finite nonnegative median or null")
	}
	var previous time.Time
	for i, w := range a.Weeks {
		d, err := time.Parse("2006-01-02", w.StartDate)
		if err != nil || d.Weekday() != time.Monday || w.Sessions < 0 || i > 0 && !d.Equal(previous.AddDate(0, 0, 7)) {
			return homeError("LocalHomeActivity", "weeks are invalid; emit eight consecutive UTC Mondays in ascending order with nonnegative sessions")
		}
		previous = d
	}
	return nil
}

// LocalHomeStats is independent of page hydration. ProjectHash narrows only
// RootFacets; Projects is the complete selected root picker, sorted by name/hash.
// No transcript, Git, live hook, audience, or annotation reads are needed.
type LocalHomeStats struct {
	ScopeRevision string              `json:"scopeRevision" minLength:"1"`
	Roots         LocalHomeRootCounts `json:"roots"`
	RootFacets    LocalHomeRootCounts `json:"rootFacets"`
	Projects      []LocalHomeProject  `json:"projects" nullable:"false"`
	Activity      LocalHomeActivity   `json:"activity"`
}

func (s LocalHomeStats) Validate() error {
	if strings.TrimSpace(s.ScopeRevision) == "" || s.Projects == nil {
		return homeError("LocalHomeStats", "scopeRevision is empty or projects is null; emit a nonempty scope revision and an array")
	}
	if err := s.Roots.Validate(); err != nil {
		return err
	}
	if err := s.RootFacets.Validate(); err != nil {
		return err
	}
	if s.RootFacets.Published > s.Roots.Published || s.RootFacets.NotPublished > s.Roots.NotPublished {
		return homeError("LocalHomeStats", "rootFacets exceeds global roots; narrow only the root cohort by project")
	}
	seen := map[ProjectHash]bool{}
	for i, p := range s.Projects {
		if p.ProjectHash.Validate() != nil || seen[p.ProjectHash] {
			return homeError("LocalHomeStats", "projects contains a malformed or duplicate hash; emit each canonical project once")
		}
		if i > 0 {
			prev := s.Projects[i-1]
			if prev.ProjectName > p.ProjectName || prev.ProjectName == p.ProjectName && prev.ProjectHash > p.ProjectHash {
				return homeError("LocalHomeStats", "projects is unsorted; order by projectName then projectHash")
			}
		}
		seen[p.ProjectHash] = true
	}
	return s.Activity.Validate()
}

func homeError(location, reason string) error {
	return fmt.Errorf("Home contract validation failed at schema.%s.Validate: %s; the caller must correct the request or response before using it", location, reason)
}
