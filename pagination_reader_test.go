package schema_test

import (
	"context"

	schema "github.com/peasant-labs/schema"
)

// Compile the concrete domain reader against the shared generic interface; no
// consumer-private page/request shape or runtime iterator is introduced.
type syncPageReaderSignature func(context.Context, schema.PageRequest) (schema.Page[schema.LocalSyncSessionRow], error)

func (read syncPageReaderSignature) ReadPage(ctx context.Context, request schema.PageRequest) (schema.Page[schema.LocalSyncSessionRow], error) {
	return read(ctx, request)
}

var _ schema.PageReader[schema.LocalSyncSessionRow] = syncPageReaderSignature(nil)
