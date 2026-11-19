package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

type surfaceStateKind string

const (
	surfaceClean     surfaceStateKind = "clean"
	surfaceDocsAhead surfaceStateKind = "docs-ahead"
	surfaceAhead     surfaceStateKind = "ahead"
)

type surfaceState struct {
	Name   string           `yaml:"name"`
	Label  string           `yaml:"label"`
	Frozen string           `yaml:"frozen"`
	State  surfaceStateKind `yaml:"state"`
}

type surfaceStates struct {
	Status   int            `yaml:"status"`
	Surfaces []surfaceState `yaml:"surfaces"`
}

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	write := fs.String("write", "", "write the surface state to a YAML file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("schema-gen status accepts only --write FILE")
	}
	root, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("find module root for status: %w", err)
	}
	state, err := calculateSurfaceStates(root)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode surface states: %w", err)
	}
	if *write != "" {
		if err := os.WriteFile(*write, data, 0o644); err != nil {
			return fmt.Errorf("write surface states %s: %w", *write, err)
		}
	}
	_, err = os.Stdout.Write(data)
	return err
}

// configuredSurfaces verifies the generator's document inventory against the
// release config and retains the config's order for the handoff.
func configuredSurfaces(root string) ([]reportSurface, error) {
	data, err := os.ReadFile(filepath.Join(root, ".github", "release-intent.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read release config for surface state: %w", err)
	}
	var config struct {
		SpecFreeze bool      `yaml:"spec_freeze"`
		FreezeDir  string    `yaml:"freeze_dir"`
		Surfaces   yaml.Node `yaml:"surfaces"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("decode release config for surface state: %w", err)
	}
	if !config.SpecFreeze || config.FreezeDir != "generated" || config.Surfaces.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("release config must enable spec_freeze with freeze_dir generated and configured surfaces")
	}
	byName := make(map[string]reportSurface, len(reportSurfaces))
	for _, surface := range reportSurfaces {
		byName[surface.name] = surface
	}
	var result []reportSurface
	seen := make(map[string]bool)
	for i := 0; i+1 < len(config.Surfaces.Content); i += 2 {
		name := config.Surfaces.Content[i].Value
		surface, ok := byName[name]
		if !ok || seen[name] {
			return nil, fmt.Errorf("release config has unknown or duplicate surface %q", name)
		}
		var stems []string
		if err := config.Surfaces.Content[i+1].Decode(&stems); err != nil {
			return nil, fmt.Errorf("decode release config surface %q: %w", name, err)
		}
		if !reflect.DeepEqual(stems, surface.stems) {
			return nil, fmt.Errorf("release config stems for %q do not match the generator: %v versus %v", name, stems, surface.stems)
		}
		seen[name] = true
		result = append(result, surface)
	}
	if len(result) != len(reportSurfaces) {
		return nil, fmt.Errorf("release config must name every generator surface exactly once")
	}
	return result, nil
}

func requestStem(stem string) bool {
	return stem == "publish-request" || stem == "annotation-push-request"
}

func specExtensions(stem string) []string {
	if requestStem(stem) {
		return []string{".schema.json"}
	}
	return []string{".json", ".yaml"}
}

func jsonDocument(data []byte, extension string) ([]byte, error) {
	if !strings.HasSuffix(extension, ".yaml") {
		return data, nil
	}
	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return json.Marshal(doc)
}

func validateDocumentStamp(data []byte, stem, extension, version string) error {
	encoded, err := jsonDocument(data, extension)
	if err != nil {
		return err
	}
	var document struct {
		ID   string `json:"$id"`
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	if err := json.Unmarshal(encoded, &document); err != nil {
		return err
	}
	if requestStem(stem) {
		if document.ID != "urn:peasant:"+stem+":"+version {
			return fmt.Errorf("request schema %s has $id %q, expected version %s", stem, document.ID, version)
		}
	} else if document.Info.Version != version {
		return fmt.Errorf("document %s has info.version %q, expected %s", stem, document.Info.Version, version)
	}
	return nil
}

func calculateSurfaceStates(root string) (surfaceStates, error) {
	surfaces, err := configuredSurfaces(root)
	if err != nil {
		return surfaceStates{}, err
	}
	data, err := os.ReadFile(filepath.Join(root, "versions.go"))
	if err != nil {
		return surfaceStates{}, fmt.Errorf("read versions.go for surface state: %w", err)
	}
	versions, err := parseReportVersions(data, "working tree", false)
	if err != nil {
		return surfaceStates{}, err
	}
	temp, err := os.MkdirTemp("", "schema-surface-status-")
	if err != nil {
		return surfaceStates{}, err
	}
	defer os.RemoveAll(temp)
	out := surfaceStates{Status: 1, Surfaces: make([]surfaceState, 0, len(surfaces))}
	for _, surface := range surfaces {
		row := surfaceState{Name: surface.name, Label: surface.label, Frozen: versions[surface.version], State: surfaceClean}
		if _, err := parseFreezeVersion(row.Frozen); err != nil {
			return surfaceStates{}, fmt.Errorf("surface %s has invalid frozen version: %w", surface.name, err)
		}
		for _, stem := range surface.stems {
			for _, extension := range specExtensions(stem) {
				frozenPath := filepath.Join(root, "generated", stem+"-"+row.Frozen+extension)
				frozen, err := os.ReadFile(frozenPath)
				if err != nil {
					return surfaceStates{}, fmt.Errorf("read frozen document %s: %w", frozenPath, err)
				}
				if err := validateDocumentStamp(frozen, stem, extension, row.Frozen); err != nil {
					return surfaceStates{}, fmt.Errorf("invalid frozen document %s: %w", frozenPath, err)
				}
				workingPath := filepath.Join(root, "generated", stem+"-next"+extension)
				working, err := os.ReadFile(workingPath)
				if err != nil {
					return surfaceStates{}, fmt.Errorf("read working document %s: %w", workingPath, err)
				}
				if err := validateDocumentStamp(working, stem, extension, row.Frozen); err != nil {
					return surfaceStates{}, fmt.Errorf("invalid working document %s: %w", workingPath, err)
				}
				if bytes.Equal(frozen, working) {
					continue
				}
				oldJSON, err := jsonDocument(frozen, extension)
				if err != nil {
					return surfaceStates{}, fmt.Errorf("decode frozen %s: %w", frozenPath, err)
				}
				newJSON, err := jsonDocument(working, extension)
				if err != nil {
					return surfaceStates{}, fmt.Errorf("decode working %s: %w", workingPath, err)
				}
				change, _, err := compareReportDocument(temp, stem, oldJSON, newJSON, requestStem(stem))
				if err != nil {
					return surfaceStates{}, fmt.Errorf("compare %s and %s: %w", frozenPath, workingPath, err)
				}
				if change == reportFunctional {
					row.State = surfaceAhead
				} else if row.State == surfaceClean {
					row.State = surfaceDocsAhead
				}
			}
		}
		out.Surfaces = append(out.Surfaces, row)
	}
	return out, nil
}
