package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestPreparedTagSnapshot runs the actual verification and push-ancestry
// steps against local Git repositories and real bundles. Git's bundle and
// ancestry semantics, including a branch advancing, are the behavior at risk.
func TestPreparedTagSnapshot(t *testing.T) {
	var fixture struct {
		Required []string `yaml:"prepared_tag_required"`
		Cases    []struct {
			Name               string `yaml:"name"`
			AdvanceBeforePlan  bool   `yaml:"advance_before_plan"`
			AdvanceAfterPlan   bool   `yaml:"advance_after_plan"`
			TagAtBase          bool   `yaml:"tag_at_base"`
			Lightweight        bool   `yaml:"lightweight"`
			UnexpectedRef      bool   `yaml:"unexpected_ref"`
			ExtraRef           bool   `yaml:"extra_ref"`
			SourceAtBase       bool   `yaml:"source_at_base"`
			CandidateOffBranch bool   `yaml:"candidate_off_branch"`
			CorruptBundle      bool   `yaml:"corrupt_bundle"`
			InvalidCandidate   bool   `yaml:"invalid_candidate"`
			InvalidSource      bool   `yaml:"invalid_source"`
			UnavailableSource  bool   `yaml:"unavailable_source"`
			Success            bool   `yaml:"success"`
			ErrorContains      string `yaml:"error_contains"`
			AncestryAtBase     bool   `yaml:"ancestry_at_base"`
			AncestryExit       int    `yaml:"expected_ancestry_exit"`
		} `yaml:"prepared_tag_cases"`
	}
	if err := yaml.Unmarshal(releaseAdoptionFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release-batch.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	jobs := workflowMappingChild(doc.Content[0], "jobs")
	script := func(jobName, stepName string) string {
		t.Helper()
		job := workflowMappingChild(jobs, jobName)
		steps := workflowMappingChild(job, "steps")
		if steps == nil {
			t.Fatalf("%s job has no steps", jobName)
		}
		for _, step := range steps.Content {
			name := workflowMappingChild(step, "name")
			run := workflowMappingChild(step, "run")
			if name != nil && name.Value == stepName && run != nil {
				return run.Value
			}
		}
		t.Fatalf("%s job has no %q script", jobName, stepName)
		return ""
	}
	verifyScript := script("verify", "Verify the prepared tag")
	ancestryScript := script("release", "Check the release tag is on the default branch")
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("empty or duplicate prepared tag case %q", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			origin := filepath.Join(dir, "origin")
			verify := filepath.Join(dir, "verify")
			env := isolatedGitFixtureEnv()
			git := func(cwd string, args ...string) string {
				t.Helper()
				cmd := exec.Command("git", args...)
				cmd.Dir = cwd
				cmd.Env = env
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, out)
				}
				return strings.TrimSpace(string(out))
			}
			git(dir, "init", "-q", "-b", "develop", origin)
			git(origin, "commit", "-q", "--allow-empty", "-m", "initial")
			base := git(origin, "rev-parse", "HEAD")
			git(origin, "commit", "-q", "--allow-empty", "-m", "release[patch]: v1.2.3")
			runCommit := git(origin, "rev-parse", "HEAD")
			git(dir, "clone", "-q", "--no-local", origin, verify)
			candidate, source := runCommit, runCommit
			if c.AdvanceBeforePlan || c.AdvanceAfterPlan {
				git(origin, "commit", "-q", "--allow-empty", "-m", "fix(release)[patch]: later change")
				if c.AdvanceBeforePlan {
					source = git(origin, "rev-parse", "HEAD")
				}
			}
			if c.CandidateOffBranch {
				git(origin, "checkout", "-q", "-b", "other", base)
				git(origin, "commit", "-q", "--allow-empty", "-m", "release[patch]: other")
				candidate = git(origin, "rev-parse", "HEAD")
				source = candidate
				git(origin, "checkout", "-q", "develop")
			}
			if c.SourceAtBase {
				source = base
			}
			tagTarget := candidate
			if c.TagAtBase {
				tagTarget = base
			}
			tag := "v1.2.3"
			bundleTag := tag
			if c.UnexpectedRef {
				bundleTag = "v1.2.4"
			}
			if c.Lightweight {
				git(origin, "tag", bundleTag, tagTarget)
			} else {
				git(origin, "tag", "-a", bundleTag, tagTarget, "-m", "release tag")
			}
			if c.ExtraRef {
				git(origin, "tag", "-a", "v1.2.4", tagTarget, "-m", "other tag")
			}
			if err := os.Mkdir(filepath.Join(verify, "bundles"), 0o755); err != nil {
				t.Fatal(err)
			}
			bundle := filepath.Join(verify, "bundles", "tag.bundle")
			excluded := runCommit
			if c.Lightweight {
				// A lightweight tag at the excluded commit has no tag object
				// to bundle; include that commit to test the verifier's refusal.
				excluded = base
			}
			bundleRefs := []string{"bundle", "create", bundle, "refs/tags/" + bundleTag}
			if c.ExtraRef {
				bundleRefs = append(bundleRefs, "refs/tags/v1.2.4")
			}
			git(origin, append(bundleRefs, "^"+excluded)...)
			if c.CorruptBundle {
				if err := os.WriteFile(bundle, []byte("not a git bundle\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(dir, "release-intent"), []byte("#!/bin/sh\nprintf 'unexpected re-plan\\n' >&2\nexit 93\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			if c.InvalidCandidate {
				candidate = "not-a-commit"
			}
			if c.InvalidSource {
				source = "not-a-commit"
			}
			if c.UnavailableSource {
				source = strings.Repeat("a", len(runCommit))
			}
			env = append(env, "TAG="+tag, "CANDIDATE="+candidate, "PLAN_SOURCE="+source, "GITHUB_SHA="+runCommit, "RUNNER_TEMP="+dir, "MODE=tag", "FLOOR=")
			run := func(step string, overrides ...string) (string, error) {
				t.Helper()
				cmd := exec.Command("bash", "-e", "-c", step)
				cmd.Dir = verify
				cmd.Env = append(append([]string(nil), env...), overrides...)
				output, err := cmd.CombinedOutput()
				return string(output), err
			}
			output, runErr := run(verifyScript)
			if c.Success {
				if runErr != nil {
					t.Fatalf("snapshot verification failed: %v\n%s", runErr, output)
				}
				var overrides []string
				if c.AncestryAtBase {
					// The jobs normally share GITHUB_SHA. Override it only
					// here to exercise the release guard independently.
					overrides = append(overrides, "GITHUB_SHA="+base)
				}
				guardOutput, guardErr := run(ancestryScript, overrides...)
				guardExit := 0
				if guardErr != nil {
					if failure, ok := guardErr.(*exec.ExitError); ok {
						guardExit = failure.ExitCode()
					} else {
						t.Fatalf("execute release job's branch-ancestry check: %v\n%s", guardErr, guardOutput)
					}
				}
				if guardExit != c.AncestryExit {
					t.Fatalf("release job's branch-ancestry check exit %d, want %d:\n%s", guardExit, c.AncestryExit, guardOutput)
				}
			} else if runErr == nil || c.ErrorContains == "" || !strings.Contains(output, c.ErrorContains) {
				t.Fatalf("snapshot verification error = %v, want failure containing %q:\n%s", runErr, c.ErrorContains, output)
			}
		})
	}
	for _, name := range fixture.Required {
		if !seen[name] {
			t.Errorf("required prepared tag case %q missing", name)
		}
	}
}

// isolatedGitFixtureEnv returns an environment for Git fixture child commands
// that isolates host Git config and disables background maintenance/gc so no
// detached Git process can write into t.TempDir() after a test step exits.
func isolatedGitFixtureEnv() []string {
	var env []string
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "GIT_CONFIG_") {
			continue
		}
		env = append(env, entry)
	}
	return append(env,
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_COUNT=4",
		"GIT_CONFIG_KEY_0=maintenance.auto", "GIT_CONFIG_VALUE_0=false",
		"GIT_CONFIG_KEY_1=maintenance.autoDetach", "GIT_CONFIG_VALUE_1=false",
		"GIT_CONFIG_KEY_2=gc.auto", "GIT_CONFIG_VALUE_2=0",
		"GIT_CONFIG_KEY_3=gc.autoDetach", "GIT_CONFIG_VALUE_3=false",
		"GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
		"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid",
	)
}
