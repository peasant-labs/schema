package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestContractGatesAllAndReleaseVerdict runs the shipped aggregate and report
// CLI on isolated two-commit ranges, then judges their reports with the pinned
// release tool. Only the fixture repository and its refs are written.
func TestContractGatesAllAndReleaseVerdict(t *testing.T) {
	requireReportOasdiff(t)
	for _, bin := range []string{"bash", "git", "go", "go-apidiff", "vacuum"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Fatalf("%s is required for the real contract gates; run in nix develop: %v", bin, err)
		}
	}
	var fixture reportFixture
	if err := yaml.Unmarshal(contractReportFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	var tool struct {
		Pin string `yaml:"tool_pin"`
	}
	if err := yaml.Unmarshal(releaseAdoptionFixture, &tool); err != nil || len(tool.Pin) != 40 {
		t.Fatalf("read the pinned release tool: pin %q, error %v", tool.Pin, err)
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	scratch := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	env := append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod",
		"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
		"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
	run := func(workdir, name string, args ...string) ([]byte, error) {
		t.Helper()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = workdir
		cmd.Env = env
		return cmd.CombinedOutput()
	}
	git := func(args ...string) string {
		t.Helper()
		out, err := run(dir, "git", args...)
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	write := func(path string, data []byte) {
		t.Helper()
		path = filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	copySource := func(path string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		write(path, data)
	}
	for _, path := range []string{"go.sum", ".vacuum.yaml", ".github/release-intent.yaml", "scripts/contract-gates.sh"} {
		copySource(path)
	}
	sources, err := filepath.Glob(filepath.Join(root, "cmd/schema-gen/*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range sources {
		if !strings.HasSuffix(source, "_test.go") {
			copySource(filepath.Join("cmd/schema-gen", filepath.Base(source)))
		}
	}
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	modText := strings.Replace(string(mod), "module github.com/peasant-labs/schema", "module fixture.invalid/contract-report", 1)
	modText += "\nrequire github.com/peasant-labs/schema v0.0.0\nreplace github.com/peasant-labs/schema => " + strconv.Quote(root) + "\n"
	write("go.mod", []byte(modText))
	write("CHANGELOG.md", []byte("# Changelog\n"))

	var breakCaseFound bool
	for _, c := range fixture.Cases {
		if !c.Breaking {
			continue
		}
		breakCaseFound = true
		versions := func(village string) []byte {
			return []byte(fmt.Sprintf("package schema\nconst (\nTypesVersion = %q\nVillageAPIVersion = %q\nPeasantLocalAPIVersion = %q\nSegmentedReadAPIVersion = %q\n)\n", c.BaseVersion, village, c.BaseVersion, c.BaseVersion))
		}
		write("versions.go", versions(c.BaseVersion))
		for _, surface := range reportSurfaces {
			for _, stem := range surface.stems {
				extension, content := ".json", c.BaseSpec
				if stem == "publish-request" || stem == "annotation-push-request" {
					extension, content = ".schema.json", c.RequestSchema
				}
				write("generated/"+stem+"-"+c.BaseVersion+extension, []byte(content))
			}
		}
		write("generated/village-api-"+c.BaseVersion+".yaml", []byte(c.BaseSpec))
		git("init", "-q")
		git("add", ".")
		git("commit", "-qm", "chore: establish fixture")
		base := git("rev-parse", "HEAD")
		git("tag", "v"+c.BaseVersion, base)
		write("versions.go", versions(c.HeadVersion))
		write("generated/village-api-"+c.HeadVersion+".json", []byte(c.HeadSpec))
		write("generated/village-api-"+c.HeadVersion+".yaml", []byte(c.HeadSpec))
		write("generated/publish-request-"+c.HeadVersion+".schema.json", []byte(c.RequestSchema))
		write("generated/annotation-push-request-"+c.HeadVersion+".schema.json", []byte(c.RequestSchema))
		git("add", "versions.go", "generated")
		tree := git("write-tree")
		heads := make(map[string]string)
		for _, subject := range fixture.Subjects {
			head := git("commit-tree", tree, "-p", base, "-m", subject.Title)
			git("branch", subject.Name, head)
			git("tag", subject.Tag, head)
			heads[subject.Name] = head
		}
		// The current checkout stays at the base. This forces --head to select
		// the fixture commit rather than accidentally using the worktree HEAD.
		write("generated/village-api-"+c.BaseVersion+".json", []byte(c.HeadSpec))
		git("add", "generated")
		badTree := git("write-tree")
		badHead := git("commit-tree", badTree, "-p", base, "-m", fixture.Subjects[0].Title)
		git("branch", "immutable-file-mutation", badHead)
		write("generated/village-api-"+c.BaseVersion+".json", []byte(c.BaseSpec))

		install := exec.CommandContext(ctx, "go", "install", "github.com/peasant-labs/infra/tools/release/cmd/release-intent@"+tool.Pin)
		install.Dir = root
		install.Env = append(env, "GOBIN="+scratch, "GOFLAGS=")
		if out, err := install.CombinedOutput(); err != nil {
			t.Fatalf("build pinned release-intent %s: %v\n%s", tool.Pin, err, out)
		}
		binary := filepath.Join(scratch, "release-intent")
		if out, err := run(root, binary, "check-config", "--config", ".github/release-intent.yaml"); err != nil {
			t.Fatalf("pinned check-config: %v\n%s", err, out)
		}
		if out, err := run(root, binary, "check-workflows", "--policy", ".github/release-queue.policy.yml", "--dir", ".github/workflows"); err != nil {
			t.Fatalf("pinned check-workflows: %v\n%s", err, out)
		}
		seen := make(map[string]bool)
		for _, subject := range fixture.Subjects {
			subject := subject
			seen[subject.Name] = true
			t.Run(subject.Name, func(t *testing.T) {
				reportPath := filepath.Join(scratch, subject.Name+".yaml")
				out, err := run(dir, "bash", filepath.Join(dir, "scripts/contract-gates.sh"), "all", base, "--head", subject.Tag, "--out", reportPath)
				if err != nil {
					t.Fatalf("contract gates all: %v\n%s", err, out)
				}
				data, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatalf("read requested report: %v\n%s", err, out)
				}
				var report contractReport
				if err := yaml.Unmarshal(data, &report); err != nil {
					t.Fatal(err)
				}
				if report.Base != base || report.Head != heads[subject.Name] {
					t.Fatalf("report range = %s..%s, want %s..%s", report.Base, report.Head, base, heads[subject.Name])
				}
				found := false
				for _, surface := range report.Surfaces {
					if surface.Name == c.ExpectedSurface {
						found = true
						if !surface.Breaking || string(surface.Change) != c.ExpectedChange {
							t.Fatalf("ERR-level break not recorded: %+v", surface)
						}
					}
				}
				if !found {
					t.Fatalf("surface %q missing from report", c.ExpectedSurface)
				}
				out, err = run(dir, binary, "check-contract", "--until", subject.Tag, "--floor", "", "--report", reportPath)
				exit := 0
				if err != nil {
					if failure, ok := err.(*exec.ExitError); ok {
						exit = failure.ExitCode()
					} else {
						t.Fatalf("run pinned check-contract: %v\n%s", err, out)
					}
				}
				if exit != subject.ExpectedExit || !strings.Contains(string(out), subject.ExpectedOutput) {
					t.Fatalf("check-contract exit %d, want %d and %q:\n%s", exit, subject.ExpectedExit, subject.ExpectedOutput, out)
				}
			})
		}
		seen["immutable-file-mutation"] = true
		t.Run("immutable-file-mutation", func(t *testing.T) {
			reportPath := filepath.Join(scratch, "immutable.yaml")
			out, err := run(dir, "bash", filepath.Join(dir, "scripts/contract-gates.sh"), "all", base, "--head", badHead, "--out", reportPath)
			if err == nil || !strings.Contains(string(out), "released numbered spec") {
				t.Fatalf("immutable mutation must fail the aggregate:\n%s", out)
			}
			if _, err := os.Stat(reportPath); err != nil {
				t.Fatalf("independent report missing after immutable failure: %v\n%s", err, out)
			}
		})
		for _, required := range fixture.ScriptRequired {
			if !seen[required] {
				t.Errorf("required contract-gates case %q missing", required)
			}
		}
		break
	}
	if !breakCaseFound {
		t.Fatal("contract report fixture has no breaking case for the real aggregate")
	}
}
