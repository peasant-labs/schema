package schema

import (
	"fmt"
	"strconv"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// PageSize is the closed set of page sizes a local list may request. It is a
// real Go enum with a numeric wire form: the members are 25, 50, and 100, and no
// layer carries a string spelling.
type PageSize int

const (
	PageSize25  PageSize = 25
	PageSize50  PageSize = 50
	PageSize100 PageSize = 100
)

// AllPageSizes is the canonical inventory, in ascending order. The TypeScript
// set and the OpenAPI enum both derive from it.
var AllPageSizes = []PageSize{PageSize25, PageSize50, PageSize100}

// IsValid reports whether v is a member of AllPageSizes. Zero is not a member:
// an omitted maximum maps to 25 before NewPageSize is called.
func (p PageSize) IsValid() bool { return inSet(p, AllPageSizes) }

// String returns the decimal spelling of the member.
func (p PageSize) String() string { return strconv.Itoa(int(p)) }

// NewPageSize accepts 25, 50, or 100. Any other value, including zero and a
// negative number, is an error; a boundary never casts an integer into a member.
func NewPageSize(raw int) (PageSize, error) {
	size := PageSize(raw)
	if !size.IsValid() {
		return 0, fmt.Errorf("page size validation failed at schema.NewPageSize: %d is outside the closed set; a list accepts only 25, 50, or 100, so map an omitted or zero maximum to 25 before calling this constructor", raw)
	}
	return size, nil
}

// JSONSchema emits the numeric enum the generated bindings consume.
func (PageSize) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.Integer)
	s.WithTitle("Page Size")
	s.WithDescription("Closed page size set for local list reads: 25, 50, or 100. The value is a number on the wire; an omitted maximum maps to 25 before validation.")
	enum := make([]any, len(AllPageSizes))
	for i, size := range AllPageSizes {
		enum[i] = int(size)
	}
	s.WithEnum(enum...)
	s.WithExamples(enum...)
	return s, nil
}

// OffsetPage is the one page shape every offset reader returns. Items is the
// requested page; NextItems carries up to perPage rows from the next page, so a
// client can pre-compute the following page without a second read. Page is
// one-based. Revision is the store revision rendered as a decimal string, never
// an opaque revision name. A short or empty Items array may still continue: only
// an empty NextItems means the traversal is done.
type OffsetPage[T any] struct {
	Items     []T      `json:"items" nullable:"false"`
	NextItems []T      `json:"nextItems" nullable:"false"`
	Page      int      `json:"page" minimum:"1" maximum:"9007199254740991"`
	PerPage   PageSize `json:"perPage"`
	Total     int      `json:"total" minimum:"0" maximum:"9007199254740991"`
	Revision  string   `json:"revision" minLength:"1"`
}

// Validate checks the common offset-page contract and any item-owned
// validations. Domain readers own how the rows and the total are selected.
func (p OffsetPage[T]) Validate() error {
	if p.Items == nil || p.NextItems == nil {
		return fmt.Errorf("offset page validation failed at schema.OffsetPage.Validate: items or nextItems is null; callers cannot tell a usable empty page from a missing array; emit [] for either array and inspect nextItems for continuation")
	}
	if p.Page < 1 || int64(p.Page) > maxSafeJSONInteger {
		return fmt.Errorf("offset page validation failed at schema.OffsetPage.Validate: page %d is outside 1..%d; a one-based page number cannot be paged exactly; emit the requested page or 1", p.Page, maxSafeJSONInteger)
	}
	if !p.PerPage.IsValid() {
		return fmt.Errorf("offset page validation failed at schema.OffsetPage.Validate: perPage %d is outside the closed set; clients can only request 25, 50, or 100, so map an omitted maximum to 25 before emitting the page", int(p.PerPage))
	}
	if len(p.Items) > int(p.PerPage) || len(p.NextItems) > int(p.PerPage) {
		return fmt.Errorf("offset page validation failed at schema.OffsetPage.Validate: items (%d) or nextItems (%d) exceeds perPage %d; a page larger than its declared size cannot be paged exactly; serve at most perPage rows in either array", len(p.Items), len(p.NextItems), int(p.PerPage))
	}
	if p.Total < 0 || int64(p.Total) > maxSafeJSONInteger {
		return fmt.Errorf("offset page validation failed at schema.OffsetPage.Validate: total %d is outside 0..%d; clients cannot render real page numbers from an out-of-range count; emit the counted total", p.Total, maxSafeJSONInteger)
	}
	if !isDecimalRevision(p.Revision) {
		return fmt.Errorf("offset page validation failed at schema.OffsetPage.Validate: revision %q is not a decimal string; the revision is the store revision rendered as decimal digits, not an opaque name; emit store_state.revision", p.Revision)
	}
	for _, item := range p.Items {
		if v, ok := any(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	}
	for _, item := range p.NextItems {
		if v, ok := any(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// isDecimalRevision reports whether raw is one or more ASCII decimal digits.
func isDecimalRevision(raw string) bool {
	if raw == "" {
		return false
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
