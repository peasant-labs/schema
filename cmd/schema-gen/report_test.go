package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/openapi"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/contract_report.yaml
var contractReportFixture []byte

type reportFixture struct {
	Required       []string `yaml:"required"`
	ScriptRequired []string `yaml:"script_required"`
	StemRequired   []string `yaml:"stem_required"`
	Subjects       []struct {
		Name           string `yaml:"name"`
		Tag            string `yaml:"tag"`
		Title          string `yaml:"title"`
		ExpectedExit   int    `yaml:"expected_exit"`
		ExpectedOutput string `yaml:"expected_output"`
	} `yaml:"subjects"`
	StemCases []struct {
		Name     string `yaml:"name"`
		Path     string `yaml:"path"`
		Expected string `yaml:"expected"`
	} `yaml:"stem_cases"`
	Cases []struct {
		Name            string `yaml:"name"`
		BaseVersion     string `yaml:"base_version"`
		HeadVersion     string `yaml:"head_version"`
		BaseSpec        string `yaml:"base_spec"`
		HeadSpec        string `yaml:"head_spec"`
		RequestSchema   string `yaml:"request_schema"`
		ExpectedSurface string `yaml:"expected_surface"`
		ExpectedChange  string `yaml:"expected_change"`
		Breaking        bool   `yaml:"expected_breaking"`
		OasdiffID       string `yaml:"expected_oasdiff_id"`
		OasdiffLevel    int    `yaml:"expected_oasdiff_level"`
	} `yaml:"cases"`
}

func TestReportSurfacesMatchConfigAndGenerator(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github/release-intent.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Surfaces map[string][]string `yaml:"surfaces"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	actual := make(map[string][]string)
	stems := make(map[string]bool)
	for _, surface := range reportSurfaces {
		actual[surface.name] = surface.stems
		for _, stem := range surface.stems {
			stems[stem] = true
		}
	}
	if !reflect.DeepEqual(cfg.Surfaces, actual) {
		t.Fatalf("report surfaces differ from release config: report=%v config=%v", actual, cfg.Surfaces)
	}
	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	generated := make(map[string]bool)
	for path := range artifacts {
		stem := generatedStem(path, stems)
		if stem == "" {
			t.Errorf("generated contract %s has no report surface", path)
		} else {
			generated[stem] = true
		}
	}
	if !reflect.DeepEqual(stems, generated) {
		t.Fatalf("report stems differ from generator: report=%v generator=%v", stems, generated)
	}
	var fixture reportFixture
	if err := yaml.Unmarshal(contractReportFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, c := range fixture.StemCases {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("empty or duplicate stem case %q", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			if got := generatedStem(c.Path, stems); got != c.Expected {
				t.Errorf("stem for %q = %q, want %q", c.Path, got, c.Expected)
			}
		})
	}
	for _, name := range fixture.StemRequired {
		if !seen[name] {
			t.Errorf("required stem case %q missing", name)
		}
	}
}

var generatedVersionedSuffix = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(\.schema)?\.(json|yaml)$`)

func generatedStem(path string, stems map[string]bool) string {
	for stem := range stems {
		if suffix, ok := strings.CutPrefix(path, stem+"-"); ok && generatedVersionedSuffix.MatchString(suffix) {
			return stem
		}
	}
	return ""
}

func TestReportFindsBreakInNewNumberedFileAcrossRange(t *testing.T) {
	requireReportOasdiff(t)
	var fixture reportFixture
	if err := yaml.Unmarshal(contractReportFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			git := func(args ...string) string {
				t.Helper()
				cmd := exec.Command("git", args...)
				cmd.Dir = dir
				cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, out)
				}
				return strings.TrimSpace(string(out))
			}
			write := func(path, content string) {
				t.Helper()
				path = filepath.Join(dir, path)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			git("init", "-q")
			versions := func(village string) string {
				return fmt.Sprintf("package schema\nconst (\nTypesVersion = %q\nVillageAPIVersion = %q\nPeasantLocalAPIVersion = %q\nSegmentedReadAPIVersion = %q\n)\n", c.BaseVersion, village, c.BaseVersion, c.BaseVersion)
			}
			write("versions.go", versions(c.BaseVersion))
			for _, surface := range reportSurfaces {
				for _, stem := range surface.stems {
					extension, content := ".json", c.BaseSpec
					if stem == "publish-request" || stem == "annotation-push-request" {
						extension, content = ".schema.json", c.RequestSchema
					}
					write("generated/"+stem+"-"+c.BaseVersion+extension, content)
				}
			}
			git("add", "versions.go", "generated")
			git("commit", "-qm", "base")
			base := git("rev-parse", "HEAD")
			write("versions.go", versions(c.HeadVersion))
			write("generated/village-api-"+c.HeadVersion+".json", c.HeadSpec)
			write("generated/publish-request-"+c.HeadVersion+".schema.json", c.RequestSchema)
			write("generated/annotation-push-request-"+c.HeadVersion+".schema.json", c.RequestSchema)
			git("add", "versions.go", "generated")
			git("commit", "-qm", "remove endpoint in new version")
			write("later.txt", "a later change does not hide the break")
			git("add", "later.txt")
			git("commit", "-qm", "later change")
			head := git("rev-parse", "HEAD")
			write("apidiff.txt", "")
			if c.OasdiffID != "" {
				write("warning-base.json", c.BaseSpec)
				write("warning-head.json", c.HeadSpec)
				cmd := exec.Command("oasdiff", "breaking", filepath.Join(dir, "warning-base.json"),
					filepath.Join(dir, "warning-head.json"), "--fail-on", "ERR", "--format", "json")
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("non-ERR fixture must not fail oasdiff: %v\n%s", err, out)
				}
				var changes []struct {
					ID    string `json:"id"`
					Level int    `json:"level"`
				}
				if err := json.Unmarshal(out, &changes); err != nil {
					t.Fatalf("decode oasdiff warnings: %v\n%s", err, out)
				}
				found := false
				for _, change := range changes {
					if change.ID == c.OasdiffID && change.Level == c.OasdiffLevel {
						found = true
					}
				}
				if !found {
					t.Fatalf("fixture did not produce oasdiff %q at level %d:\n%s", c.OasdiffID, c.OasdiffLevel, out)
				}
			}
			report, err := buildContractReport(dir, base, head, filepath.Join(dir, "apidiff.txt"))
			if err != nil {
				t.Fatal(err)
			}
			if report.Base != base || report.Head != head {
				t.Fatalf("report endpoints = %s..%s, want %s..%s", report.Base, report.Head, base, head)
			}
			for _, surface := range report.Surfaces {
				if surface.Name == c.ExpectedSurface {
					if string(surface.Change) != c.ExpectedChange || surface.Breaking != c.Breaking {
						t.Fatalf("surface = %+v, expected change=%s breaking=%v", surface, c.ExpectedChange, c.Breaking)
					}
					return
				}
			}
			t.Fatalf("surface %q missing from report", c.ExpectedSurface)
		})
	}
	for _, name := range fixture.Required {
		if !seen[name] {
			t.Errorf("required report case %q missing", name)
		}
	}
}

func requireReportOasdiff(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("oasdiff"); err != nil {
		if _, ci := os.LookupEnv("CI"); ci {
			t.Fatal("oasdiff is required in CI; run the report tests in nix develop")
		}
		t.Skip("oasdiff is supplied by nix develop")
	}
}
