package openapi

import (
	"fmt"

	"github.com/swaggest/openapi-go/openapi31"
)

// applyAnnotationPushIngressConstraints adds the cross-field target
// discriminator the reflector cannot infer from Go struct tags alone. The
// canonical Types schema owns these constraints; Village receives byte-equivalent
// shared components through harmonizeSharedTypeComponents.
func applyAnnotationPushIngressConstraints(spec *openapi31.Spec) error {
	schemas := spec.ComponentsEns().Schemas
	item, ok := schemas["AnnotationPushItem"]
	if !ok {
		return fmt.Errorf("annotation push schema generation cannot find AnnotationPushItem in the canonical Types catalog")
	}
	properties, ok := item["properties"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("annotation push schema generation found AnnotationPushItem without properties")
	}
	for _, name := range annotationPushTargetFields {
		if _, exists := properties[name]; !exists {
			return fmt.Errorf("annotation push schema generation found AnnotationPushItem without target property %q", name)
		}
	}

	item["oneOf"] = []interface{}{
		annotationPushTargetArm("association", "targetAssociationId", map[string]interface{}{
			"targetAssociationId": componentRef("AssociationID"),
		}),
		annotationPushTargetArm("session", "sessionId", map[string]interface{}{
			"sessionId": nonEmptyStringSchema(),
		}),
		annotationPushTargetArm("entry", "entryTarget", map[string]interface{}{
			"entryTarget": componentRef("AnnotationEntryTarget"),
		}),
		annotationPushTargetArm("annotation", "annotationId", map[string]interface{}{
			"annotationId": nonEmptyStringSchema(),
		}),
		annotationPushTargetArm("project", "projectHash", map[string]interface{}{
			"projectHash": componentRef("ProjectHash"),
		}),
	}
	return nil
}

var annotationPushTargetFields = []string{
	"targetAssociationId",
	"sessionId",
	"entryTarget",
	"annotationId",
	"projectHash",
}

func annotationPushTargetArm(kind, requiredField string, constrainedProperties map[string]interface{}) map[string]interface{} {
	properties := map[string]interface{}{"targetKind": map[string]interface{}{"const": kind}}
	for name, value := range constrainedProperties {
		properties[name] = value
	}
	forbidden := make([]interface{}, 0, len(annotationPushTargetFields)-1)
	for _, field := range annotationPushTargetFields {
		if field == requiredField {
			continue
		}
		forbidden = append(forbidden, requiredNonNullProperty(field))
	}
	return map[string]interface{}{
		"properties": properties,
		"required":   []interface{}{"targetKind", requiredField},
		"not":        map[string]interface{}{"anyOf": forbidden},
	}
}

// requiredNonNullProperty matches an inactive target only when its wire value
// is present and non-null. Go pointer decoding and the generated root Zod
// contract both treat an explicit JSON null as absent, so the served schema
// must preserve that acceptance policy rather than treating key presence alone
// as a competing target arm.
func requiredNonNullProperty(name string) map[string]interface{} {
	return map[string]interface{}{
		"required": []interface{}{name},
		"properties": map[string]interface{}{
			name: map[string]interface{}{
				"not": map[string]interface{}{"type": "null"},
			},
		},
	}
}

func componentRef(name string) map[string]interface{} {
	return map[string]interface{}{"$ref": "#/components/schemas/" + name}
}

func nonEmptyStringSchema() map[string]interface{} {
	return map[string]interface{}{"type": "string", "minLength": 1}
}
