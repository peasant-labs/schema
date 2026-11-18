package main

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/freeze_status.yaml
var freezeStatusCases []byte

//go:embed testdata/freeze_status_manifest.yaml
var freezeStatusManifest []byte

//go:embed testdata/freeze_openapi.json
var freezeOpenAPI []byte

//go:embed testdata/freeze_request.json
var freezeRequest []byte

type freezeFixtureEdit string

const freezeFixtureUnchanged freezeFixtureEdit = "unchanged"

type freezeStatusInput struct {
	Edit    freezeFixtureEdit `yaml:"edit"`
	Sibling string            `yaml:"sibling"`
}

type freezeStatusExpected struct {
	Surface string           `yaml:"surface"`
	State   surfaceStateKind `yaml:"state"`
}

func TestFreezeSurfaceStates(t *testing.T) {
	requireReportOasdiff(t)
	corpus, err := testcase.LoadCorpus[freezeStatusInput, freezeStatusExpected](freezeStatusCases)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	names := make([]string, 0, len(corpus.Cases))
	accepts := make(map[string]bool)
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
		accepts[c.Name] = c.Classification == testcase.MustPass
	}
	caseassert.RequireManifestNames(t, "freeze state", freezeStatusManifest, names)
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(root, ".github/release-intent.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Classification == testcase.MustFail && !accepts[c.Input.Sibling] {
				t.Fatalf("refusal %s has no accepting sibling", c.Name)
			}
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, ".github"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(dir, "generated"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, ".github/release-intent.yaml"), config, 0o644); err != nil {
				t.Fatal(err)
			}
			versionSource := "package schema\nconst (\nTypesVersion = \"1.2.3\"\nVillageAPIVersion = \"1.2.3\"\nPeasantLocalAPIVersion = \"1.2.3\"\nSegmentedReadAPIVersion = \"1.2.3\"\n)\n"
			if err := os.WriteFile(filepath.Join(dir, "versions.go"), []byte(versionSource), 0o644); err != nil {
				t.Fatal(err)
			}
			for _, surface := range reportSurfaces {
				for _, stem := range surface.stems {
					original := freezeOpenAPI
					if requestStem(stem) {
						original = []byte(strings.ReplaceAll(string(freezeRequest), "STEM", stem))
					}
					for _, ext := range specExtensions(stem) {
						value := original
						if ext == ".yaml" {
							var object any
							if err := json.Unmarshal(original, &object); err != nil {
								t.Fatal(err)
							}
							value, err = yaml.Marshal(object)
							if err != nil {
								t.Fatal(err)
							}
						}
						frozen := filepath.Join(dir, "generated", stem+"-1.2.3"+ext)
						if err := os.WriteFile(frozen, value, 0o644); err != nil {
							t.Fatal(err)
						}
						if workingSurface(surface) {
							if err := os.WriteFile(filepath.Join(dir, "generated", stem+"-next"+ext), value, 0o644); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
			}
			git := func(args ...string) string {
				t.Helper()
				cmd := exec.Command("git", args...)
				cmd.Dir = dir
				cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
					"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
					"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, out)
				}
				return strings.TrimSpace(string(out))
			}
			git("init", "-q")
			git("add", ".")
			git("commit", "-qm", "ci: initial fixture")
			base := git("rev-parse", "HEAD")
			if c.Input.Edit != freezeFixtureUnchanged {
				stem := "peasantlocal-api"
				if strings.HasPrefix(string(c.Input.Edit), "request-") {
					stem = "publish-request"
				}
				ext := ".json"
				if requestStem(stem) {
					ext = ".schema.json"
				}
				path := filepath.Join(dir, "generated", stem+"-next"+ext)
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var doc map[string]any
				if err := json.Unmarshal(data, &doc); err != nil {
					t.Fatal(err)
				}
				mutateFreezeDocument(t, doc, c.Input.Edit)
				data, err = json.Marshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			git("add", ".")
			git("commit", "--allow-empty", "-qm", "fix: fixture change")
			head := git("rev-parse", "HEAD")
			state, err := calculateSurfaceStates(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range state.Surfaces {
				if row.Name == c.Expected.Surface && row.State != c.Expected.State {
					t.Fatalf("%s state = %s, want %s", row.Name, row.State, c.Expected.State)
				}
				if row.Name == "segmented-read" && row.State != surfaceClean {
					t.Fatalf("numbered segmented-read is not clean: %s", row.State)
				}
			}
			err = checkReleaseTree(state)
			if (err == nil) != (c.Classification == testcase.MustPass) {
				t.Fatalf("release tree check error = %v for %s", err, c.Classification)
			}
			apiSignal := filepath.Join(t.TempDir(), "apidiff.txt")
			if err := os.WriteFile(apiSignal, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			report, err := buildContractReport(dir, base, head, apiSignal)
			if err != nil {
				t.Fatal(err)
			}
			wantChange := reportNone
			if c.Expected.State == surfaceDocsAhead {
				wantChange = reportDocs
			} else if c.Expected.State == surfaceAhead {
				wantChange = reportFunctional
			}
			for _, row := range report.Surfaces {
				if row.Name == c.Expected.Surface && (row.Change != wantChange || row.AheadOnBase) {
					t.Fatalf("report surface = %+v, want change=%s and ahead_on_base=false", row, wantChange)
				}
			}
			if c.Expected.State == surfaceAhead {
				git("commit", "--allow-empty", "-qm", "chore: later change")
				later := git("rev-parse", "HEAD")
				report, err = buildContractReport(dir, head, later, apiSignal)
				if err != nil {
					t.Fatal(err)
				}
				for _, row := range report.Surfaces {
					if row.Name == c.Expected.Surface && (!row.AheadOnBase || row.Change != reportNone) {
						t.Fatalf("unchanged surface ahead on base = %+v", row)
					}
				}
			}
		})
	}
}

func mutateFreezeDocument(t *testing.T, doc map[string]any, edit freezeFixtureEdit) {
	t.Helper()
	if edit == "openapi-description" {
		doc["info"].(map[string]any)["description"] = "A clarified description." // unstring-allow: go-json-key-assert — fixture mutates arbitrary documents
		return
	}
	if edit == "openapi-removed-path" {
		delete(doc["paths"].(map[string]any), "/fixture") // unstring-allow: go-json-key-assert — fixture mutates arbitrary documents
		return
	}
	var properties map[string]any
	if strings.HasPrefix(string(edit), "request-") {
		properties = doc["properties"].(map[string]any) // unstring-allow: go-json-key-assert — fixture mutates arbitrary documents
	} else {
		properties = doc["components"].(map[string]any)["schemas"].(map[string]any)["Example"].(map[string]any)["properties"].(map[string]any) // unstring-allow: go-json-key-assert — fixture mutates arbitrary documents
	}
	switch edit {
	case "openapi-new-field", "request-new-field":
		properties["newValue"] = properties["value"]
	case "openapi-changed-type", "request-changed-type":
		properties["value"].(map[string]any)["type"] = "boolean" // unstring-allow: go-json-key-assert — fixture mutates arbitrary documents
	case "request-description":
		properties["value"].(map[string]any)["description"] = "A clarified value." // unstring-allow: go-json-key-assert — fixture mutates arbitrary documents
	case "request-removed-field":
		delete(properties, "value")
	default:
		t.Fatalf("unknown freeze fixture edit %q", edit)
	}
}
