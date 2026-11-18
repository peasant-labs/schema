package main

import (
	"strings"

	"github.com/peasant-labs/schema/openapi"
)

// workingArtifactName maps the generator's current contract documents to the
// files content changes update. The minimum segmented protocol stays numbered.
func workingArtifactName(name string) string {
	for _, surface := range reportSurfaces {
		var version string
		switch surface.version {
		case "TypesVersion":
			version = openapi.TypesVersion
		case "VillageAPIVersion":
			version = openapi.VillageAPIVersion
		case "PeasantLocalAPIVersion":
			version = openapi.PeasantLocalAPIVersion
		case "SegmentedReadSpecVersion":
			version = openapi.SegmentedReadSpecVersion
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
