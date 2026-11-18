package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

const (
	frozenRegistryPath = "cmd/schema-gen/testdata/frozen_specs.yaml"
	frozenManifestPath = "cmd/schema-gen/testdata/frozen_specs_manifest.yaml"
	versionMirrorPath  = "typescript/src/internal/generated/versions.gen.ts"
	exportsFixturePath = "testdata/typescript/public_exports.yaml"
)

type freezeLevel string

const (
	freezeNone  freezeLevel = "none"
	freezePatch freezeLevel = "patch"
	freezeMinor freezeLevel = "minor"
	freezeMajor freezeLevel = "major"
)

type freezeSurface struct {
	Name     string      `yaml:"name"`
	Label    string      `yaml:"label"`
	From     string      `yaml:"from"`
	To       string      `yaml:"to"`
	Level    freezeLevel `yaml:"level"`
	Breaking bool        `yaml:"breaking"`
	Freeze   bool        `yaml:"freeze"`
}

type freezePlan struct {
	Plan        int             `yaml:"plan"`
	FreezeFloor string          `yaml:"freeze_floor,omitempty"`
	Surfaces    []freezeSurface `yaml:"surfaces"`
}

type freezeVersion struct {
	major, minor, patch int
}

var bareFreezeVersion = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

func parseFreezeVersion(value string) (freezeVersion, error) {
	match := bareFreezeVersion.FindStringSubmatch(value)
	if match == nil {
		return freezeVersion{}, fmt.Errorf("version %q must be a bare X.Y.Z version", value)
	}
	var result freezeVersion
	var err error
	result.major, err = strconv.Atoi(match[1])
	if err != nil {
		return freezeVersion{}, err
	}
	result.minor, err = strconv.Atoi(match[2])
	if err != nil {
		return freezeVersion{}, err
	}
	result.patch, err = strconv.Atoi(match[3])
	if err != nil {
		return freezeVersion{}, err
	}
	return result, nil
}

func nextFreezeVersion(value string, level freezeLevel) (string, error) {
	v, err := parseFreezeVersion(value)
	if err != nil {
		return "", err
	}
	switch level {
	case freezeNone:
		return value, nil
	case freezePatch:
		v.patch++
	case freezeMinor:
		v.minor++
		v.patch = 0
	case freezeMajor:
		if v.major == 0 {
			v.minor++
		} else {
			v.major++
			v.minor = 0
		}
		v.patch = 0
	default:
		return "", fmt.Errorf("unknown freeze level %q", level)
	}
	if v.major < 0 || v.minor < 0 || v.patch < 0 {
		return "", errors.New("freeze version arithmetic overflowed")
	}
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch), nil
}

func strictFreezeYAML(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read freeze handoff %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("decode freeze handoff %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("freeze handoff %s must contain one YAML document", path)
	}
	return nil
}

func readFreezeState(path string) (surfaceStates, error) {
	var raw struct {
		Status   *int `yaml:"status"`
		Surfaces *[]struct {
			Name   *string           `yaml:"name"`
			Label  *string           `yaml:"label"`
			Frozen *string           `yaml:"frozen"`
			State  *surfaceStateKind `yaml:"state"`
		} `yaml:"surfaces"`
	}
	if err := strictFreezeYAML(path, &raw); err != nil {
		return surfaceStates{}, err
	}
	if raw.Status == nil || *raw.Status != 1 || raw.Surfaces == nil {
		return surfaceStates{}, errors.New("surface state must contain status: 1 and every surface")
	}
	out := surfaceStates{Status: *raw.Status}
	for _, row := range *raw.Surfaces {
		if row.Name == nil || row.Label == nil || row.Frozen == nil || row.State == nil {
			return surfaceStates{}, errors.New("surface state has a missing field")
		}
		switch *row.State {
		case surfaceClean, surfaceDocsAhead, surfaceAhead:
		default:
			return surfaceStates{}, fmt.Errorf("unknown surface state %q", *row.State)
		}
		out.Surfaces = append(out.Surfaces, surfaceState{*row.Name, *row.Label, *row.Frozen, *row.State})
	}
	return out, nil
}

func readFreezePlan(path string) (freezePlan, error) {
	var raw struct {
		Plan        *int   `yaml:"plan"`
		FreezeFloor string `yaml:"freeze_floor"`
		Surfaces    *[]struct {
			Name     *string      `yaml:"name"`
			Label    *string      `yaml:"label"`
			From     *string      `yaml:"from"`
			To       *string      `yaml:"to"`
			Level    *freezeLevel `yaml:"level"`
			Breaking *bool        `yaml:"breaking"`
			Freeze   *bool        `yaml:"freeze"`
		} `yaml:"surfaces"`
	}
	if err := strictFreezeYAML(path, &raw); err != nil {
		return freezePlan{}, err
	}
	if raw.Plan == nil || *raw.Plan != 1 || raw.Surfaces == nil {
		return freezePlan{}, errors.New("freeze plan must contain plan: 1 and every surface")
	}
	out := freezePlan{Plan: *raw.Plan, FreezeFloor: raw.FreezeFloor}
	for _, row := range *raw.Surfaces {
		if row.Name == nil || row.Label == nil || row.From == nil || row.To == nil || row.Level == nil || row.Breaking == nil || row.Freeze == nil {
			return freezePlan{}, errors.New("freeze plan has a missing field")
		}
		out.Surfaces = append(out.Surfaces, freezeSurface{*row.Name, *row.Label, *row.From, *row.To, *row.Level, *row.Breaking, *row.Freeze})
	}
	return out, nil
}

func validateFreezeHandoff(surfaces []reportSurface, state surfaceStates, plan freezePlan) error {
	if state.Status != 1 || plan.Plan != 1 || len(state.Surfaces) != len(surfaces) || len(plan.Surfaces) != len(surfaces) {
		return errors.New("freeze handoffs must name every configured surface once, in config order")
	}
	for i, surface := range surfaces {
		s, p := state.Surfaces[i], plan.Surfaces[i]
		if s.Name != surface.name || p.Name != surface.name || s.Label != surface.label || p.Label != surface.label || s.Frozen != p.From {
			return fmt.Errorf("freeze handoffs disagree with config, label or frozen version for %s", surface.name)
		}
		if p.Freeze != (s.State == surfaceAhead) || (!workingSurface(surface) && p.Freeze) {
			return fmt.Errorf("freeze plan does not match the state of %s", surface.name)
		}
		next, err := nextFreezeVersion(p.From, p.Level)
		if err != nil || next != p.To || (p.Freeze && p.Level == freezeNone) || (!p.Freeze && (p.Level != freezeNone || p.Breaking)) {
			return fmt.Errorf("freeze plan has an incoherent version, level or flag for %s", surface.name)
		}
	}
	return nil
}

func runFreeze(args []string) error {
	fs := flag.NewFlagSet("freeze", flag.ContinueOnError)
	check := fs.Bool("check", false, "check the release tree without writing")
	planPath := fs.String("plan", "", "total release-intent freeze plan")
	statePath := fs.String("state", "", "surface state from the plan's source tree")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || (*planPath == "") != (*statePath == "") || (!*check && *planPath == "") {
		return errors.New("schema-gen freeze requires --plan FILE --state FILE, or --check with both files or neither")
	}
	root, err := findModuleRoot()
	if err != nil {
		return err
	}
	current, err := calculateSurfaceStates(root)
	if err != nil {
		return err
	}
	if *check && *planPath == "" {
		return checkReleaseTree(current)
	}
	state, err := readFreezeState(*statePath)
	if err != nil {
		return err
	}
	plan, err := readFreezePlan(*planPath)
	if err != nil {
		return err
	}
	surfaces, err := configuredSurfaces(root)
	if err != nil {
		return err
	}
	if err := validateFreezeHandoff(surfaces, state, plan); err != nil {
		return err
	}
	if *check {
		return checkFreezeCandidate(current, plan)
	}
	if !reflectSurfaceStates(state, current) {
		return errors.New("surface state is stale or does not describe the current tree; regenerate the state and plan")
	}
	return writeFreeze(root, surfaces, plan)
}

func checkFreezeCandidate(current surfaceStates, plan freezePlan) error {
	if err := checkReleaseTree(current); err != nil {
		return err
	}
	if len(current.Surfaces) != len(plan.Surfaces) {
		return errors.New("candidate surface set does not match the freeze plan")
	}
	for i, p := range plan.Surfaces {
		if current.Surfaces[i].Name != p.Name || current.Surfaces[i].Frozen != p.To {
			return fmt.Errorf("candidate version for %s does not equal planned %s", p.Name, p.To)
		}
	}
	return nil
}

func reflectSurfaceStates(a, b surfaceStates) bool {
	if a.Status != b.Status || len(a.Surfaces) != len(b.Surfaces) {
		return false
	}
	for i := range a.Surfaces {
		if a.Surfaces[i] != b.Surfaces[i] {
			return false
		}
	}
	return true
}

func checkReleaseTree(state surfaceStates) error {
	for _, surface := range state.Surfaces {
		if surface.State == surfaceAhead {
			return fmt.Errorf("release tree has an ahead surface %s; rebuild the release PR with a fresh freeze", surface.Name)
		}
	}
	return nil
}

// runFreezeRender is re-executed with a versions.go overlay, so the generator
// sees freshly compiled constants. It writes only to the caller's scratch dir.
func runFreezeRender(args []string) error {
	fs := flag.NewFlagSet("freeze-render", flag.ContinueOnError)
	out := fs.String("out", "", "scratch directory for regenerated working specs")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *out == "" || fs.NArg() != 0 {
		return errors.New("schema-gen freeze-render requires --out DIR")
	}
	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	for filename, data := range artifacts {
		name := workingArtifactName(filename)
		if name == filename {
			continue
		}
		if err := os.WriteFile(filepath.Join(*out, name), data, 0o644); err != nil {
			return fmt.Errorf("write regenerated working document %s: %w", name, err)
		}
	}
	return nil
}

type freezeWrite struct {
	path string
	data []byte
}

func replaceVersionLine(data []byte, prefix, oldVersion, newVersion string) ([]byte, error) {
	pattern := regexp.MustCompile(`(?m)^(` + prefix + `)` + regexp.QuoteMeta(oldVersion) + `([^"\n]*"?[^\n]*)$`)
	matches := pattern.FindAllIndex(data, -1)
	if len(matches) != 1 {
		return nil, fmt.Errorf("version line for %q occurs %d times, want exactly one", oldVersion, len(matches))
	}
	return pattern.ReplaceAll(data, []byte("${1}"+newVersion+"${2}")), nil
}

func replaceDocumentVersion(data []byte, stem, extension, oldVersion, newVersion string) ([]byte, error) {
	switch {
	case requestStem(stem):
		return replaceVersionLine(data, `  "\$id": "urn:peasant:`+regexp.QuoteMeta(stem)+`:`, oldVersion, newVersion)
	case extension == ".yaml":
		return replaceVersionLine(data, `  version: `, oldVersion, newVersion)
	default:
		return replaceVersionLine(data, `    "version": "`, oldVersion, newVersion)
	}
}

func updatedVersionSources(root string, surfaces []reportSurface, plan freezePlan) ([]byte, []byte, []byte, error) {
	versions, err := os.ReadFile(filepath.Join(root, "versions.go"))
	if err != nil {
		return nil, nil, nil, err
	}
	mirror, err := os.ReadFile(filepath.Join(root, versionMirrorPath))
	if err != nil {
		return nil, nil, nil, err
	}
	fixture, err := os.ReadFile(filepath.Join(root, exportsFixturePath))
	if err != nil {
		return nil, nil, nil, err
	}
	for i, p := range plan.Surfaces {
		if !p.Freeze {
			continue
		}
		name := regexp.QuoteMeta(surfaces[i].version)
		versions, err = replaceVersionLine(versions, `\s*`+name+`\s*=\s*"`, p.From, p.To)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("update versions.go %s: %w", name, err)
		}
		mirror, err = replaceVersionLine(mirror, `export const `+name+` = "`, p.From, p.To)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("update TypeScript version mirror %s: %w", name, err)
		}
		pattern := regexp.MustCompile(`(?m)(  - name: ` + name + `\n    value: )` + regexp.QuoteMeta(p.From) + `(\n)`)
		if len(pattern.FindAllIndex(fixture, -1)) != 1 {
			return nil, nil, nil, fmt.Errorf("public exports fixture has no unique %s version value", name)
		}
		fixture = pattern.ReplaceAll(fixture, []byte("${1}"+p.To+"${2}"))
	}
	return versions, mirror, fixture, nil
}

// writeFreeze accepts only a state/plan pair validated against the current
// tree. It recompiles the generator before comparing and copying its output.
func writeFreeze(root string, surfaces []reportSurface, plan freezePlan) (resultErr error) {
	return writeFreezeWithWriter(root, surfaces, plan, os.WriteFile)
}

func writeFreezeWithWriter(root string, surfaces []reportSurface, plan freezePlan, writeFile func(string, []byte, os.FileMode) error) error {
	var some bool
	for _, p := range plan.Surfaces {
		some = some || p.Freeze
	}
	if !some {
		return nil
	}
	versions, mirror, fixture, err := updatedVersionSources(root, surfaces, plan)
	if err != nil {
		return err
	}
	registry, err := os.ReadFile(filepath.Join(root, frozenRegistryPath))
	if err != nil {
		return err
	}
	manifest, err := os.ReadFile(filepath.Join(root, frozenManifestPath))
	if err != nil {
		return err
	}
	for i, p := range plan.Surfaces {
		if !p.Freeze {
			continue
		}
		for _, stem := range surfaces[i].stems {
			for _, extension := range specExtensions(stem) {
				target := filepath.Join(root, "generated", stem+"-"+p.To+extension)
				if _, err := os.Lstat(target); err == nil {
					return fmt.Errorf("refusing to overwrite frozen document %s", target)
				} else if !os.IsNotExist(err) {
					return err
				}
			}
		}
	}
	temp, err := os.MkdirTemp("", "schema-freeze-render-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	overlayVersions := filepath.Join(temp, "versions.go")
	if err := os.WriteFile(overlayVersions, versions, 0o600); err != nil {
		return err
	}
	overlay, err := json.Marshal(struct {
		Replace map[string]string `json:"Replace"`
	}{Replace: map[string]string{filepath.Join(root, "versions.go"): overlayVersions}})
	if err != nil {
		return err
	}
	overlayPath := filepath.Join(temp, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o600); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "-mod=readonly", "-overlay", overlayPath, "./cmd/schema-gen", "freeze-render", "--out", temp)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rebuild schema-gen with the planned versions: %w: %s", err, output)
	}
	writes := []freezeWrite{{"versions.go", versions}, {versionMirrorPath, mirror}, {exportsFixturePath, fixture}}
	for i, p := range plan.Surfaces {
		if !p.Freeze {
			continue
		}
		for _, stem := range surfaces[i].stems {
			var jsonHash, yamlHash string
			for _, extension := range specExtensions(stem) {
				name := stem + "-next" + extension
				before, err := os.ReadFile(filepath.Join(root, "generated", name))
				if err != nil {
					return err
				}
				rendered, err := os.ReadFile(filepath.Join(temp, name))
				if err != nil {
					return err
				}
				expected, err := replaceDocumentVersion(before, stem, extension, p.From, p.To)
				if err != nil {
					return fmt.Errorf("update working document %s: %w", name, err)
				}
				if !bytes.Equal(rendered, expected) {
					return fmt.Errorf("regenerating %s changes more than its version line; regenerate and review the content before freezing", name)
				}
				numbered := stem + "-" + p.To + extension
				writes = append(writes, freezeWrite{filepath.Join("generated", name), rendered}, freezeWrite{filepath.Join("generated", numbered), rendered})
				sum := fmt.Sprintf("%x", sha256.Sum256(rendered))
				if strings.HasSuffix(extension, ".json") {
					jsonHash = sum
				} else {
					yamlHash = sum
				}
			}
			registryName := stem + "-" + p.To
			if requestStem(stem) {
				registryName += ".schema"
			}
			registry = append(registry, []byte(fmt.Sprintf("  - name: %s\n    json_sha256: %s\n", registryName, jsonHash))...)
			if requestStem(stem) {
				registry = append(registry, []byte("    json_only: true\n")...)
			} else {
				registry = append(registry, []byte(fmt.Sprintf("    yaml_sha256: %s\n", yamlHash))...)
			}
			manifest = append(manifest, []byte("  - "+registryName+"\n")...)
		}
	}
	writes = append(writes, freezeWrite{frozenRegistryPath, registry}, freezeWrite{frozenManifestPath, manifest})
	type priorOutput struct {
		path   string
		data   []byte
		mode   os.FileMode
		exists bool
	}
	var prior []priorOutput
	for _, write := range writes {
		filename := filepath.Join(root, write.path)
		entry := priorOutput{path: filename}
		info, err := os.Lstat(filename)
		if err == nil {
			if !info.Mode().IsRegular() {
				return fmt.Errorf("freeze output %s is not a regular file", write.path)
			}
			entry.data, err = os.ReadFile(filename)
			if err != nil {
				return err
			}
			entry.exists, entry.mode = true, info.Mode().Perm()
		} else if !os.IsNotExist(err) {
			return err
		}
		prior = append(prior, entry)
	}
	for i, write := range writes {
		if err := writeFile(prior[i].path, write.data, 0o644); err != nil {
			var rollbackErr error
			for j := i; j >= 0; j-- {
				entry := prior[j]
				if entry.exists {
					if restoreErr := os.WriteFile(entry.path, entry.data, entry.mode); restoreErr != nil {
						rollbackErr = errors.Join(rollbackErr, restoreErr)
					}
				} else if removeErr := os.Remove(entry.path); removeErr != nil && !os.IsNotExist(removeErr) {
					rollbackErr = errors.Join(rollbackErr, removeErr)
				}
			}
			return fmt.Errorf("write freeze output %s: %w; rollback error: %v", write.path, err, rollbackErr)
		}
	}
	return nil
}
