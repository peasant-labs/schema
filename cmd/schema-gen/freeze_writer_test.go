package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/freeze_writer.yaml
var freezeWriterCases []byte

//go:embed testdata/freeze_writer_manifest.yaml
var freezeWriterManifest []byte

type freezeWriterEdit struct {
	Path    string `yaml:"path"`
	Find    string `yaml:"find"`
	Replace string `yaml:"replace"`
}

type freezeWriterInput struct {
	Edits          []freezeWriterEdit `yaml:"edits"`
	Ahead          []string           `yaml:"ahead"`
	RequiredPaths  []string           `yaml:"required_paths"`
	PreservePaths  []string           `yaml:"preserve_paths"`
	VerifyProtocol bool               `yaml:"verify_protocol"`
}

type freezeWriterExpected struct {
	Check bool `yaml:"check"`
}

// TestFreezeWriterCoverage runs the actual writer after generating source
// changes in an isolated copy. Its write seam records every path, including
// writes whose bytes happen not to change.
func TestFreezeWriterCoverage(t *testing.T) {
	requireReportOasdiff(t)
	corpus, err := testcase.LoadCorpus[freezeWriterInput, freezeWriterExpected](freezeWriterCases)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	var names []string
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "freeze writer", freezeWriterManifest, names)
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "repository")
			copyFreezeSource(t, root, dir)
			baseline, err := calculateSurfaceStates(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, edit := range c.Input.Edits {
				filename := filepath.Join(dir, edit.Path)
				data, err := os.ReadFile(filename)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Count(string(data), edit.Find) != 1 {
					t.Fatalf("source edit %s does not have exactly one anchor", edit.Path)
				}
				data = []byte(strings.Replace(string(data), edit.Find, edit.Replace, 1))
				if err := os.WriteFile(filename, data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			runFreezeFixtureCommand(t, dir, "go", "run", "./cmd/schema-gen")
			state, err := calculateSurfaceStates(dir)
			if err != nil {
				t.Fatal(err)
			}
			surfaces, err := configuredSurfaces(dir)
			if err != nil {
				t.Fatal(err)
			}
			plan := freezePlan{Plan: 1, FreezeFloor: strings.Repeat("a", 40)}
			for _, row := range state.Surfaces {
				ahead := slices.Contains(c.Input.Ahead, row.Name)
				for _, initial := range baseline.Surfaces {
					if initial.Name == row.Name && initial.State == surfaceAhead {
						ahead = true
					}
				}
				if (row.State == surfaceAhead) != ahead {
					t.Fatalf("surface %s state = %s, expected ahead=%v", row.Name, row.State, ahead)
				}
				level := freezeNone
				if ahead {
					level = freezeMinor
				}
				to, err := nextFreezeVersion(row.Frozen, level)
				if err != nil {
					t.Fatal(err)
				}
				plan.Surfaces = append(plan.Surfaces, freezeSurface{row.Name, row.Label, row.Frozen, to, level, false, ahead})
			}
			if err := validateFreezeHandoff(surfaces, state, plan); err != nil {
				t.Fatal(err)
			}
			before := freezeFileSnapshot(t, dir)
			versionBytes, err := os.ReadFile(filepath.Join(dir, "versions.go"))
			if err != nil {
				t.Fatal(err)
			}
			beforeVersions, err := parseReportVersions(versionBytes, "before freeze", false)
			if err != nil {
				t.Fatal(err)
			}
			written := make(map[string]bool)
			writeFile := func(filename string, data []byte, mode os.FileMode) error {
				name, err := filepath.Rel(dir, filename)
				if err != nil {
					return err
				}
				written[filepath.ToSlash(name)] = true
				return os.WriteFile(filename, data, mode)
			}
			if err := writeFreezeWithWriter(dir, surfaces, plan, writeFile); err != nil {
				t.Fatalf("freeze writer failed: %v", err)
			}
			after := freezeFileSnapshot(t, dir)
			for _, name := range c.Input.PreservePaths {
				old, ok := before[name]
				if !ok || after[name] != old {
					t.Errorf("document freeze changed or removed preserved file %s", name)
				}
			}
			if c.Input.VerifyProtocol {
				versionBytes, err = os.ReadFile(filepath.Join(dir, "versions.go"))
				if err != nil {
					t.Fatal(err)
				}
				afterVersions, err := parseReportVersions(versionBytes, "after freeze", false)
				if err != nil {
					t.Fatal(err)
				}
				for _, name := range []string{"SegmentedReadAPIVersion", "SegmentedReadMinVersion"} {
					if afterVersions[name] != beforeVersions[name] {
						t.Errorf("document freeze changed protocol constant %s", name)
					}
				}
				if afterVersions["SegmentedReadSpecVersion"] == beforeVersions["SegmentedReadSpecVersion"] {
					t.Error("document freeze did not advance the segmented document version")
				}
			}
			var config struct {
				Changelog      string   `yaml:"changelog"`
				ProtectedPaths []string `yaml:"protected_paths"`
				FreezeOutputs  []string `yaml:"freeze_outputs"`
				VersionFiles   []struct {
					Path string `yaml:"path"`
				} `yaml:"version_files"`
			}
			configBytes, err := os.ReadFile(filepath.Join(dir, ".github/release-intent.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			if err := yaml.Unmarshal(configBytes, &config); err != nil {
				t.Fatal(err)
			}
			patterns := append(append([]string{config.Changelog}, config.ProtectedPaths...), config.FreezeOutputs...)
			for _, version := range config.VersionFiles {
				patterns = append(patterns, version.Path)
			}
			changed := make(map[string]bool)
			for name := range written {
				covered := false
				for _, pattern := range patterns {
					match, err := path.Match(pattern, name)
					if err != nil {
						t.Fatal(err)
					}
					covered = covered || match
				}
				if !covered {
					t.Errorf("actual freeze write %s is not authorized by release config", name)
				}
			}
			for name, sum := range after {
				if old, exists := before[name]; exists && old == sum {
					continue
				}
				changed[name] = true
				covered := false
				for _, pattern := range patterns {
					match, err := path.Match(pattern, name)
					if err != nil {
						t.Fatal(err)
					}
					covered = covered || match
				}
				if !covered {
					t.Errorf("actual freeze output %s is not authorized by release config", name)
				}
			}
			for _, required := range c.Input.RequiredPaths {
				if !changed[required] {
					t.Errorf("expected freeze output %s was not written", required)
				}
			}
			for i, p := range plan.Surfaces {
				if !p.Freeze {
					continue
				}
				for _, stem := range surfaces[i].stems {
					for _, ext := range specExtensions(stem) {
						working := filepath.Join(dir, "generated", stem+"-next"+ext)
						frozen := filepath.Join(dir, "generated", stem+"-"+p.To+ext)
						a, err := os.ReadFile(working)
						if err != nil {
							t.Fatal(err)
						}
						b, err := os.ReadFile(frozen)
						if err != nil || !bytes.Equal(a, b) {
							t.Fatalf("frozen file %s is not an exact working copy: %v", frozen, err)
						}
						if !changed[filepath.ToSlash(strings.TrimPrefix(frozen, dir+string(os.PathSeparator)))] {
							t.Errorf("freeze did not create %s", frozen)
						}
					}
				}
			}
			planFile, stateFile := filepath.Join(t.TempDir(), "plan.yaml"), filepath.Join(t.TempDir(), "state.yaml")
			for filename, value := range map[string]any{planFile: plan, stateFile: state} {
				data, err := yaml.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filename, data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			runFreezeFixtureCommand(t, dir, "go", "run", "./cmd/schema-gen", "freeze", "--check", "--plan", planFile, "--state", stateFile)
			if c.Input.VerifyProtocol {
				runFreezeFixtureCommand(t, dir, "go", "test", ".", "-run", "^(TestSegmentedReadProtocolVersion|TestTranscriptReadContextProtocolVersion|TestSegmentedReadSpecAccessors)$", "-count=1")
				runFreezeFixtureCommand(t, dir, "go", "test", "./openapi", "-run", "^(TestSegmentedReadSpec|TestSegmentedReadSpec_GenerationTableMatches|TestSegmentedReadHostGenerationParametersMatch)$", "-count=1")
			}
		})
	}
}

func runFreezeFixtureCommand(t *testing.T, dir string, command string, args ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", command, args, err, output)
	}
}

func copyFreezeSource(t *testing.T, root, dest string) {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	cmd.Dir = root
	data, err := cmd.Output()
	var names []string
	if err == nil {
		names = strings.Split(string(data), "\x00")
	} else {
		// Nix builds test an exported source tree with no Git metadata.
		err = filepath.WalkDir(root, func(filename string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			name, err := filepath.Rel(root, filename)
			if err != nil {
				return err
			}
			name = filepath.ToSlash(name)
			if entry.IsDir() {
				switch name {
				case ".git", ".direnv", "typescript/node_modules", "typescript/dist", "typescript/.test-dist", "docs/api":
					return filepath.SkipDir
				}
				return nil
			}
			if name != ".git" {
				names = append(names, name)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("list exported source files for freeze fixture: %v", err)
		}
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		from, to := filepath.Join(root, name), filepath.Join(dest, name)
		info, err := os.Lstat(from)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatalf("inspect freeze fixture source %s: %v", name, err)
		}
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(from)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, to); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if !info.Mode().IsRegular() {
			t.Fatalf("freeze fixture source %s has unsupported mode %v", name, info.Mode())
		}
		in, err := os.Open(from)
		if err != nil {
			t.Fatal(err)
		}
		out, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm()|0o200)
		if err != nil {
			_ = in.Close()
			t.Fatal(err)
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		_ = in.Close()
		if copyErr != nil || closeErr != nil {
			t.Fatalf("copy fixture source %s: %v %v", name, copyErr, closeErr)
		}
	}
}

func freezeFileSnapshot(t *testing.T, root string) map[string][32]byte {
	t.Helper()
	snapshot := make(map[string][32]byte)
	err := filepath.WalkDir(root, func(filename string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(filename)
			if err != nil {
				return err
			}
			name, err := filepath.Rel(root, filename)
			if err != nil {
				return err
			}
			snapshot[filepath.ToSlash(name)] = sha256.Sum256([]byte("symlink:" + target))
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("freeze fixture contains non-regular file %s", filename)
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		name, err := filepath.Rel(root, filename)
		if err != nil {
			return err
		}
		snapshot[filepath.ToSlash(name)] = sha256.Sum256(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
