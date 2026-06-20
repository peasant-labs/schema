package schema

import "errors"

// ErrTypeNotFound is returned when a requested annotation type does not exist in the registry.
var ErrTypeNotFound = errors.New("annotation type not found")

// ErrCycleDetected is returned when adding a dependency would create a cycle
// in the annotation_type_deps graph (V14: depth limit 20).
var ErrCycleDetected = errors.New("annotation type dependency cycle detected")
