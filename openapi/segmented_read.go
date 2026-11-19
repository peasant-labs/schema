package openapi

import (
	"fmt"
	"strings"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/route"
	"github.com/swaggest/openapi-go/openapi31"
)

// BuildSegmentedReadSpec builds one declared protocol's component catalog and
// stamps it with the independently versioned document. The host owns routes.
func BuildSegmentedReadSpec(protocolVersion, documentVersion string) (*openapi31.Spec, error) {
	var roots []string
	var description string
	switch protocolVersion {
	case schema.SegmentedReadMinVersion:
		roots = []string{"TranscriptSegmentDirectory", "TranscriptSegment", "TranscriptWindow"}
		description = historicalSegmentedReadGenerationDescription
	case schema.SegmentedReadAPIVersion:
		roots = []string{"PositionedDirectory", "PositionedSegment", "PositionedWindow"}
		description = segmentedReadGenerationDescription()
	default:
		return nil, fmt.Errorf("segmented read spec generation failed at openapi.BuildSegmentedReadSpec: protocol version %q is not declared; no protocol document can be generated; use %q or %q from versions.go", protocolVersion, schema.SegmentedReadMinVersion, schema.SegmentedReadAPIVersion)
	}

	spec, err := BuildTypesSpec()
	if err != nil {
		return nil, fmt.Errorf("build segmented read component source: %w", err)
	}
	all := spec.Components.Schemas
	selected := make(map[string]map[string]interface{})
	var add func(string) error
	var visit func(interface{}) error
	add = func(name string) error {
		if _, ok := selected[name]; ok {
			return nil
		}
		component, ok := all[name]
		if !ok {
			return fmt.Errorf("segmented read spec generation failed at component %q: a referenced type is absent from the canonical catalog; the document cannot describe its wire shape; restore the catalog declaration", name)
		}
		selected[name] = component
		return visit(component)
	}
	visit = func(value interface{}) error {
		switch node := value.(type) {
		case map[string]interface{}:
			for key, child := range node {
				if key == "$ref" {
					ref, ok := child.(string)
					if !ok || !strings.HasPrefix(ref, "#/components/schemas/") {
						return fmt.Errorf("segmented read spec generation failed at $ref %v: reference is not a local component; the catalog cannot close its dependency graph; restore a canonical component reference", child)
					}
					if err := add(strings.TrimPrefix(ref, "#/components/schemas/")); err != nil {
						return err
					}
				} else if err := visit(child); err != nil {
					return err
				}
			}
		case []interface{}:
			for _, child := range node {
				if err := visit(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, name := range roots {
		if err := add(name); err != nil {
			return nil, err
		}
	}
	spec.Components.Schemas = selected
	spec.Info.WithTitle("Peasant Segmented Read Protocol v" + documentVersion).
		WithVersion(documentVersion).
		WithDescription(segmentedReadDescription(description))
	return spec, nil
}

// The minimum document is an already-shipped artifact; later selector changes
// must not rewrite its historical description.
const historicalSegmentedReadGenerationDescription = "Protocol generation selectors (omission is v=0):\n- v=0 -> 0.1.0\n- v=1 -> 0.1.0\n- v=2 -> 0.2.0"

func segmentedReadDescription(generationDescription string) string {
	var description strings.Builder
	description.WriteString("This document owns segmented read shapes; the host API document owns the routes. Routes:")
	for _, member := range []route.Route{route.VillageGetTranscriptSegments, route.VillageGetTranscriptSegment, route.VillageGetTranscriptWindow} {
		fmt.Fprintf(&description, "\n- %s %s (%s)", member.Method(), member.Pattern(), member.OperationID())
	}
	description.WriteString("\n")
	description.WriteString(generationDescription)
	return description.String()
}

func segmentedReadGenerationDescription() string {
	var description strings.Builder
	description.WriteString("Protocol generation selectors (omission is v=0):")
	for _, generation := range schema.SegmentedReadGenerations() {
		version, ok := schema.SegmentedReadProtocolVersion(generation)
		if !ok {
			panic("segmented read generation table returned an unsupported selector")
		}
		fmt.Fprintf(&description, "\n- v=%d -> %s", generation, version)
	}
	return description.String()
}
