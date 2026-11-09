package schema

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// LocalSyncPublicationFilter narrows the stored root cohort before paging.
type LocalSyncPublicationFilter string

const (
	LocalSyncPublicationAll          LocalSyncPublicationFilter = "all"
	LocalSyncPublicationNotPublished LocalSyncPublicationFilter = "not-published"
	LocalSyncPublicationPublished    LocalSyncPublicationFilter = "published"
	LocalSyncPublicationAuto         LocalSyncPublicationFilter = "auto"
)

var AllLocalSyncPublicationFilters = []LocalSyncPublicationFilter{LocalSyncPublicationAll, LocalSyncPublicationNotPublished, LocalSyncPublicationPublished, LocalSyncPublicationAuto}

func (f LocalSyncPublicationFilter) IsValid() bool  { return inSet(f, AllLocalSyncPublicationFilters) }
func (f LocalSyncPublicationFilter) String() string { return string(f) }
func (LocalSyncPublicationFilter) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Local Sync Publication Filter", "Root publication filter; auto uses the canonical live hook decision", AllLocalSyncPublicationFilters), nil
}

// LocalSyncSessionsQuery is the sync resource's query. The omitted view is
// Home: it serves the same bounded offset page as view=home. An explicit
// grouped or project-options read keeps its legacy envelope and refuses paging
// fields. pageToken is retained only so an old client receives an actionable
// refusal; no reader consumes a token.
type LocalSyncSessionsQuery struct {
	View            string                      `query:"view" enum:"grouped,home,project-options" description:"Omit for the Home offset page; grouped retains helper groups; home selects forward pages; project-options selects metadata-only root projects"`
	MaxItems        *int                        `query:"maxItems" nullable:"false" minimum:"0" default:"25" description:"Home only: omitted or zero defaults to 25; 25, 50, and 100 are accepted; 26, 101, and negative values are refused"`
	Page            *int                        `query:"page" nullable:"false" minimum:"1" description:"Home only: one-based page number; omitted defaults to 1"`
	PageToken       *string                     `query:"pageToken" nullable:"false" maxLength:"4096" description:"Home only: retired continuation; omitted or empty is absent; a nonempty value is refused so an old client can recover"`
	ParentSessionID *string                     `query:"parentSessionId" nullable:"false" minLength:"1"`
	ProjectHash     *ProjectHash                `query:"projectHash" nullable:"false"`
	Publication     *LocalSyncPublicationFilter `query:"publication" nullable:"false" default:"all" description:"Root publication filter; auto uses the canonical live hook decision. This is an explicit publication filter, not selection scope"`
}

// Home reports whether the query selects the offset Home read. The omitted view
// is Home; grouped and project-options are not.
func (q LocalSyncSessionsQuery) Home() bool { return q.View == "" || q.View == "home" }

// HasPageToken reports whether a non-empty retired continuation was supplied.
// An empty pageToken is absent, exactly as an omitted one is.
func (q LocalSyncSessionsQuery) HasPageToken() bool { return q.PageToken != nil && *q.PageToken != "" }

func (q LocalSyncSessionsQuery) Validate() error {
	if q.View != "" && q.View != "grouped" && q.View != "home" && q.View != "project-options" {
		return syncPageError("LocalSyncSessionsQuery", "view is unknown; omit it or use grouped, home, or project-options")
	}
	// The byte cap is checked before the retired-token refusal so an oversized
	// old token reports the input bound it broke, not the retirement.
	if q.PageToken != nil && len(*q.PageToken) > 4096 {
		return syncPageError("LocalSyncSessionsQuery", "pageToken exceeds 4096 bytes; no page can be used; discard continuation and restart")
	}
	if q.HasPageToken() {
		return syncPageError("LocalSyncSessionsQuery", "pageToken is retired; no reader consumes a continuation token; omit pageToken and request the one-based page instead")
	}
	if !q.Home() && (q.MaxItems != nil || q.Page != nil || q.ParentSessionID != nil || q.ProjectHash != nil || q.Publication != nil) {
		return syncPageError("LocalSyncSessionsQuery", "paging or root filters require view=home or an omitted view; remove them for grouped or project-options reads")
	}
	if q.MaxItems != nil {
		if *q.MaxItems < 0 {
			return syncPageError("LocalSyncSessionsQuery", "maxItems is negative; use zero/default or one of 25, 50, or 100")
		}
		if *q.MaxItems != 0 {
			if _, err := NewPageSize(*q.MaxItems); err != nil {
				return syncPageError("LocalSyncSessionsQuery", "maxItems is outside the closed set; use 25, 50, or 100, or omit it for the default of 25")
			}
		}
	}
	if q.Page != nil && *q.Page < 1 {
		return syncPageError("LocalSyncSessionsQuery", "page is below one; supply a one-based page number or omit it for the first page")
	}
	if q.ParentSessionID != nil {
		if strings.TrimSpace(*q.ParentSessionID) == "" {
			return syncPageError("LocalSyncSessionsQuery", "parentSessionId is empty; supply a nonempty local session ID")
		}
		if q.ProjectHash != nil || q.Publication != nil {
			return syncPageError("LocalSyncSessionsQuery", "child request carries root filters; omit projectHash and publication to avoid hiding visible children")
		}
	}
	if q.ProjectHash != nil && q.ProjectHash.Validate() != nil {
		return syncPageError("LocalSyncSessionsQuery", "projectHash is malformed; supply a canonical project hash")
	}
	if q.Publication != nil && !q.Publication.IsValid() {
		return syncPageError("LocalSyncSessionsQuery", "publication is unknown; use all, not-published, published, or auto")
	}
	return nil
}

// EffectivePageSize applies the Home transport's default after validation. An
// omitted or zero maxItems maps to 25; any other value is re-parsed through the
// constructor and its error is returned, so an unvalidated non-member can never
// silently become page size zero.
func (q LocalSyncSessionsQuery) EffectivePageSize() (PageSize, error) {
	if q.MaxItems == nil || *q.MaxItems == 0 {
		return PageSize25, nil
	}
	return NewPageSize(*q.MaxItems)
}

// EffectivePage applies the one-based default after validation.
func (q LocalSyncSessionsQuery) EffectivePage() int {
	if q.Page == nil {
		return 1
	}
	return *q.Page
}

// ParseLocalSyncSessionsQuery preserves parameter presence and rejects ambiguous
// repeated values before handlers dispatch to a domain reader.
func ParseLocalSyncSessionsQuery(values url.Values) (LocalSyncSessionsQuery, error) {
	var q LocalSyncSessionsQuery
	for name, entries := range values {
		switch name {
		case "view", "maxItems", "page", "pageToken", "parentSessionId", "projectHash", "publication":
			if len(entries) != 1 {
				return q, syncPageError("ParseLocalSyncSessionsQuery", name+" is repeated; supply it once so no page is chosen ambiguously")
			}
		}
	}
	q.View = values.Get("view")
	if raw, ok := values["maxItems"]; ok {
		n, err := strconv.Atoi(raw[0])
		if err != nil {
			return q, syncPageError("ParseLocalSyncSessionsQuery", "maxItems is malformed; supply an integer, or omit it for the default")
		}
		q.MaxItems = &n
	}
	if raw, ok := values["page"]; ok {
		n, err := strconv.Atoi(raw[0])
		if err != nil {
			return q, syncPageError("ParseLocalSyncSessionsQuery", "page is malformed; supply a one-based integer, or omit it for the first page")
		}
		q.Page = &n
	}
	if raw, ok := values["pageToken"]; ok {
		token := raw[0]
		q.PageToken = &token
	}
	if raw, ok := values["parentSessionId"]; ok {
		q.ParentSessionID = &raw[0]
	}
	if raw, ok := values["projectHash"]; ok {
		hash := ProjectHash(raw[0])
		q.ProjectHash = &hash
	}
	if raw, ok := values["publication"]; ok {
		filter := LocalSyncPublicationFilter(raw[0])
		q.Publication = &filter
	}
	return q, q.Validate()
}

// LocalSyncSessionRow hydrates only a returned session, sharing the canonical
// sync/publication projections. Preview is bounded and redaction-safe. A row may
// carry OutsideSelection: that field reports ingestion and upload-hook selection
// membership, so it never removes a stored row from a page.
type LocalSyncSessionRow struct {
	Sync        LocalSyncSummary `json:"sync"`
	Publication LocalPublication `json:"publication"`
	Title       string           `json:"title"`
	Preview     string           `json:"preview"`
	ChildCount  int              `json:"childCount" minimum:"0"`
}

func (r LocalSyncSessionRow) Validate() error {
	if err := r.Sync.Validate(); err != nil {
		return err
	}
	if err := r.Publication.Validate(); err != nil {
		return err
	}
	if r.Sync.ID != r.Publication.SessionID || r.Publication.Audience != nil || r.ChildCount < 0 {
		return syncPageError("LocalSyncSessionRow", "sync/publication identity, audience, or childCount is inconsistent; emit the same visible session ID, no audience, and a nonnegative immediate-child count")
	}
	return nil
}

// LocalSyncSessionTotals is the pushable-session count beside one offset page.
// It counts the same request filter on both branches: the filtered session set
// on a root page and the same filtered session set (not the paged children) on
// a child page. It is independent of the page total, which counts the paged
// roots or children.
type LocalSyncSessionTotals struct {
	Sessions int `json:"sessions" minimum:"0" maximum:"9007199254740991"`
}

func (t LocalSyncSessionTotals) Validate() error {
	if t.Sessions < 0 || int64(t.Sessions) > maxSafeJSONInteger {
		return syncPageError("LocalSyncSessionTotals", "sessions is outside 0..9007199254740991; clients cannot render the count exactly; emit the counted pushable sessions")
	}
	return nil
}

// LocalSyncSessionPage is the concrete offset page for the sync resource, not a
// second shape: it embeds OffsetPage[LocalSyncSessionRow] and adds the grouped
// root total. Totals is optional, so a page answers with six or seven fields:
// a reader that has no root total omits it rather than fabricating a zero.
type LocalSyncSessionPage struct {
	OffsetPage[LocalSyncSessionRow]
	Totals *LocalSyncSessionTotals `json:"totals,omitempty" nullable:"false"`
}

func (p LocalSyncSessionPage) Validate() error {
	if err := p.OffsetPage.Validate(); err != nil {
		return err
	}
	if p.Totals == nil {
		return nil
	}
	return p.Totals.Validate()
}

type LocalSyncProjectOption struct {
	ProjectHash ProjectHash `json:"projectHash"`
	ProjectName string      `json:"projectName"`
}

// LocalSyncProjectOptions contains the complete stored root picker, independent
// of the chosen root project/publication filter. No counts or enrichment.
type LocalSyncProjectOptions struct {
	Projects []LocalSyncProjectOption `json:"projects" nullable:"false"`
}

func (o LocalSyncProjectOptions) Validate() error {
	if o.Projects == nil {
		return syncPageError("LocalSyncProjectOptions", "projects is null; emit [] when no stored root project exists")
	}
	seen := map[ProjectHash]bool{}
	for i, p := range o.Projects {
		if p.ProjectHash.Validate() != nil || seen[p.ProjectHash] {
			return syncPageError("LocalSyncProjectOptions", "projects contains a malformed or duplicate hash; emit each canonical stored root project once")
		}
		if i > 0 {
			previous := o.Projects[i-1]
			if previous.ProjectName > p.ProjectName || previous.ProjectName == p.ProjectName && previous.ProjectHash > p.ProjectHash {
				return syncPageError("LocalSyncProjectOptions", "projects is unsorted; order by projectName then projectHash")
			}
		}
		seen[p.ProjectHash] = true
	}
	return nil
}

func syncPageError(location, reason string) error {
	return fmt.Errorf("sync page contract validation failed at schema.%s: %s; no page is usable until the caller corrects the request or response", location, reason)
}
