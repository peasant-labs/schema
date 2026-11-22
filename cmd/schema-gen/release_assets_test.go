package main

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/release_assets.yaml
var releaseAssetsYAML []byte

//go:embed testdata/release_assets_manifest.yaml
var releaseAssetsManifestYAML []byte

type releaseAssetsInput struct {
	Source   []string `yaml:"source"`
	Existing []string `yaml:"existing"`
	Sibling  string   `yaml:"sibling"`
}

type releaseAssetsExpected struct {
	Files         []string `yaml:"files"`
	ErrorContains string   `yaml:"error_contains"`
}

func TestReleaseAssetSelection(t *testing.T) {
	corpus, err := testcase.LoadCorpus[releaseAssetsInput, releaseAssetsExpected](releaseAssetsYAML)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	names := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "release asset", releaseAssetsManifestYAML, names)
	for _, c := range corpus.Cases {
		if c.Classification != testcase.MustFail {
			continue
		}
		found := false
		for _, sibling := range corpus.Cases {
			if sibling.Name == c.Input.Sibling && sibling.Classification == testcase.MustPass {
				found = true
			}
		}
		if !found {
			t.Fatalf("refusal case %q has no accepting sibling", c.Name)
		}
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			source, dist := filepath.Join(dir, "generated"), filepath.Join(dir, "dist")
			for _, path := range []string{source, dist} {
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			write := func(dir string, names []string) {
				t.Helper()
				for _, name := range names {
					if filepath.Base(name) != name {
						t.Fatalf("fixture filename is not a basename: %q", name)
					}
					if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
						t.Fatal(err)
					}
				}
			}
			write(source, c.Input.Source)
			write(dist, c.Input.Existing)
			cmd := exec.Command("bash", filepath.Join(root, "scripts", "collect-release-assets.sh"), source, dist)
			output, err := cmd.CombinedOutput()
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(string(output), c.Expected.ErrorContains) {
					t.Fatalf("expected refusal containing %q; err=%v output=%s", c.Expected.ErrorContains, err, output)
				}
				return
			}
			if err != nil {
				t.Fatalf("collect release assets: %v\n%s", err, output)
			}
			entries, err := os.ReadDir(dist)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, entry := range entries {
				got = append(got, entry.Name())
			}
			sort.Strings(got)
			want := append([]string(nil), c.Expected.Files...)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("published files = %v, want %v", got, want)
			}
		})
	}
}
