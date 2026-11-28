package openapi

import (
	"strings"
	"testing"

	"github.com/peasant-labs/schema/route"
	jsonschema "github.com/swaggest/jsonschema-go"
)

func TestVillageDiscoverabilityMovedToSettings(t *testing.T) {
	spec, err := BuildVillageAPISpec()
	if err != nil {
		t.Fatalf("BuildVillageAPISpec: %v", err)
	}

	authItem, ok := declaredPathItem(spec, route.VillageSetMyDiscoverability.Pattern())
	if !ok || authItem.Patch == nil {
		t.Fatalf("missing %s", route.VillageSetMyDiscoverability.Pattern())
	}
	setMyDiscoverability := authItem.Patch
	if setMyDiscoverability.Deprecated == nil || !*setMyDiscoverability.Deprecated {
		t.Fatalf("setMyDiscoverability deprecated = %#v, want true", setMyDiscoverability.Deprecated)
	}
	if setMyDiscoverability.Description == nil {
		t.Fatal("setMyDiscoverability missing description")
	}
	desc := *setMyDiscoverability.Description
	if !strings.Contains(desc, route.VillageUpdateMySettings.Pattern()) || !strings.Contains(desc, "Sunset") {
		t.Fatalf("setMyDiscoverability description %q must name replacement and Sunset", desc)
	}

	if setMyDiscoverability.Responses == nil {
		t.Fatal("setMyDiscoverability missing responses")
	}
	for _, status := range []string{"200", "400", "401", "500", "503"} {
		respOrRef, ok := setMyDiscoverability.Responses.MapOfResponseOrReferenceValues[status]
		if !ok || respOrRef.Response == nil {
			t.Fatalf("setMyDiscoverability missing status %s", status)
		}
		for _, hdr := range []string{"Deprecation", "Sunset"} {
			h, ok := respOrRef.Response.Headers[hdr]
			if !ok || h.Header == nil || h.Header.Required == nil || !*h.Header.Required {
				t.Fatalf("setMyDiscoverability status %s missing required %s header: %#v", status, hdr, h)
			}
		}
	}

	settingsItem, ok := declaredPathItem(spec, route.VillageGetMySettings.Pattern())
	if !ok || settingsItem.Get == nil || settingsItem.Patch == nil {
		t.Fatalf("missing %s", route.VillageGetMySettings.Pattern())
	}
	if settingsItem.Get.Description == nil || !strings.Contains(*settingsItem.Get.Description, "discoverability") {
		t.Fatalf("getMySettings description must mention discoverability")
	}
	if settingsItem.Patch.Description == nil {
		t.Fatal("updateMySettings missing description")
	}
	updateDesc := *settingsItem.Patch.Description
	if !strings.Contains(updateDesc, "discoverability") || !strings.Contains(updateDesc, "absent field keeps the stored value") {
		t.Fatalf("updateMySettings description %q must mention discoverability and partial update", updateDesc)
	}
	if strings.Contains(updateDesc, "7396") {
		t.Fatalf("updateMySettings description %q must not cite RFC 7396", updateDesc)
	}

	schemas := spec.Components.Schemas
	settingsSchema := schemas["SchemaVillageUserSettings"]
	settingsProps, _ := settingsSchema["properties"].(map[string]interface{})  // unstring-allow: go-json-key-assert — inspects generated OpenAPI schema map
	isDiscProp, _ := settingsProps["is_discoverable"].(map[string]interface{}) // unstring-allow: go-json-key-assert — inspects generated OpenAPI schema map
	if isDiscProp["type"] != string(jsonschema.Boolean) {
		t.Fatalf("SchemaVillageUserSettings.is_discoverable type = %#v, want boolean", isDiscProp["type"])
	}
	updateSchema := schemas["SchemaVillageUpdateUserSettingsRequest"]
	updateProps, _ := updateSchema["properties"].(map[string]interface{})        // unstring-allow: go-json-key-assert — inspects generated OpenAPI schema map
	updateDiscProp, _ := updateProps["is_discoverable"].(map[string]interface{}) // unstring-allow: go-json-key-assert — inspects generated OpenAPI schema map
	if updateDiscProp["type"] != string(jsonschema.Boolean) {
		t.Fatalf("SchemaVillageUpdateUserSettingsRequest.is_discoverable type = %#v, want boolean", updateDiscProp["type"])
	}
	if updateSchema["additionalProperties"] != false {
		t.Fatalf("SchemaVillageUpdateUserSettingsRequest.additionalProperties = %#v, want false", updateSchema["additionalProperties"])
	}
}
