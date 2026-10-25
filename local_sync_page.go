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

// LocalSyncSessionsQuery keeps forward paging opt-in on the existing resource.
// Validate before applying defaults: explicit root filters are invalid on children.
type LocalSyncSessionsQuery struct {
	View            string                      `query:"view" enum:"grouped,home,project-options" description:"Omit for legacy flat rows; grouped retains helper groups; home selects forward pages; project-options selects metadata-only root projects"`
	MaxItems        *int                        `query:"maxItems" nullable:"false" minimum:"0" default:"25" description:"Home only: omitted or zero defaults to 25; positive values clamp to 100; negative, malformed or repeated values are rejected"`
	PageToken       *PageToken                  `query:"pageToken" nullable:"false" maxLength:"4096" description:"Home only: opaque continuation, omitted or empty starts; at most 4096 bytes; repeated values are rejected"`
	ParentSessionID *string                     `query:"parentSessionId" nullable:"false" minLength:"1"`
	ProjectHash     *ProjectHash                `query:"projectHash" nullable:"false"`
	Publication     *LocalSyncPublicationFilter `query:"publication" nullable:"false" default:"all"`
}

func (q LocalSyncSessionsQuery) Validate() error {
	if q.View != "" && q.View != "grouped" && q.View != "home" && q.View != "project-options" {
		return syncPageError("LocalSyncSessionsQuery", "view is unknown; omit it or use grouped, home, or project-options")
	}
	if q.View != "home" && (q.MaxItems != nil || q.PageToken != nil || q.ParentSessionID != nil || q.ProjectHash != nil || q.Publication != nil) {
		return syncPageError("LocalSyncSessionsQuery", "paging or root filters require view=home; remove them for legacy/grouped/project-options reads")
	}
	if q.MaxItems != nil && *q.MaxItems < 0 {
		return syncPageError("LocalSyncSessionsQuery", "maxItems is negative; use zero/default or a positive maximum (clamped to 100)")
	}
	if q.PageToken != nil && len(*q.PageToken) > 4096 {
		return syncPageError("LocalSyncSessionsQuery", "pageToken exceeds 4096 bytes; no page can be used; discard continuation and restart")
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

// PageRequest applies the Home transport's default and cap after validation.
func (q LocalSyncSessionsQuery) PageRequest() PageRequest {
	maxItems := 25
	if q.MaxItems != nil && *q.MaxItems > 0 {
		maxItems = min(*q.MaxItems, 100)
	}
	var token PageToken
	if q.PageToken != nil {
		token = *q.PageToken
	}
	return PageRequest{MaxItems: maxItems, Token: token}
}

// ParseLocalSyncSessionsQuery preserves parameter presence and rejects ambiguous
// repeated values before handlers dispatch to a domain reader.
func ParseLocalSyncSessionsQuery(values url.Values) (LocalSyncSessionsQuery, error) {
	var q LocalSyncSessionsQuery
	for name, entries := range values {
		switch name {
		case "view", "maxItems", "pageToken", "parentSessionId", "projectHash", "publication":
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
	if raw, ok := values["pageToken"]; ok {
		token := PageToken(raw[0])
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

// LocalSyncSessionPage is the concrete wire instantiation, not a second shape.
type LocalSyncSessionPage = Page[LocalSyncSessionRow]

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
