package schema

import (
	"context"
	"fmt"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// PageToken is an opaque continuation. Empty starts traversal or marks its end.
// Tokens are not credentials, authorization, or a snapshot-consistency promise.
type PageToken string

func (PageToken) JSONSchema() (jsonschema.Schema, error) {
	s := jsonschema.Schema{}
	s.AddType(jsonschema.String)
	s.WithTitle("Page Token")
	s.WithDescription("Opaque continuation; empty starts traversal or marks its end. Not a credential or snapshot token.")
	return s, nil
}

// PageRequest is Go-only reader input. Domain adapters own defaults and caps.
type PageRequest struct {
	MaxItems int
	Token    PageToken
}

// Page has at most the requested maximum items. Only empty Next means done:
// a short or empty Items array may still carry continuation.
type Page[T any] struct {
	Items []T       `json:"items" nullable:"false"`
	Next  PageToken `json:"next"`
}

// Validate checks the common array contract and any item-owned validations.
// Request-specific bounds and token scope belong to the domain reader.
func (p Page[T]) Validate() error {
	if p.Items == nil {
		return fmt.Errorf("page validation failed at schema.Page.Validate: items is null; callers cannot distinguish a usable empty page; emit [] and inspect next for continuation")
	}
	for _, item := range p.Items {
		if v, ok := any(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// PageReader reads forward within constructor-configured domain scope.
type PageReader[T any] interface {
	ReadPage(context.Context, PageRequest) (Page[T], error)
}
