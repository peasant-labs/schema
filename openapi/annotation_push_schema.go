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
		forbidden = append(forbidden, map[string]interface{}{"required": []interface{}{field}})
	}
	return map[string]interface{}{
		"properties": properties,
		"required":   []interface{}{"targetKind", requiredField},
		"not":        map[string]interface{}{"anyOf": forbidden},
	}
}

func componentRef(name string) map[string]interface{} {
	return map[string]interface{}{"$ref": "#/components/schemas/" + name}
}

func nonEmptyStringSchema() map[string]interface{} {
	return map[string]interface{}{"type": "string", "minLength": 1}
}
