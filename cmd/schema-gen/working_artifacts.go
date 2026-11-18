package main

import (
	"strings"

	"github.com/peasant-labs/schema/openapi"
)

// workingArtifactName maps the generator's current contract documents to the
// files content changes update. Segmented read remains numbered until its
// protocol versioning is resolved.
func workingArtifactName(name string) string {
	for _, surface := range reportSurfaces {
		if surface.name == "segmented-read" {
			continue
		}
		var version string
		switch surface.version {
		case "TypesVersion":
			version = openapi.TypesVersion
		case "VillageAPIVersion":
			version = openapi.VillageAPIVersion
		case "PeasantLocalAPIVersion":
			version = openapi.PeasantLocalAPIVersion
		}
		for _, stem := range surface.stems {
			prefix := stem + "-" + version + "."
			if strings.HasPrefix(name, prefix) {
				return stem + "-next." + strings.TrimPrefix(name, prefix)
			}
		}
	}
	return name
}
