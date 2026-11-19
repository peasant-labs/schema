package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/tests_workflow_wiring.yaml
var testsWorkflowWiringYAML []byte

type testsWorkflowWiring struct {
	VerdictMarkers struct {
		Version int      `yaml:"version"`
		Names   []string `yaml:"names"`
	} `yaml:"verdict_markers"`
	AggregatorJob   string   `yaml:"aggregator_job"`
	AggregatorNeeds []string `yaml:"aggregator_needs"`
	MarkerSteps     []struct {
		Name string `yaml:"name"`
		If   string `yaml:"if"`
	} `yaml:"marker_steps"`
	MarkerCases []struct {
		Name, Event, Check, Race, Inherit, Expected string
	} `yaml:"marker_cases"`
	Cache struct {
		Jobs          []string `yaml:"jobs"`
		RestoreAction string   `yaml:"restore_action"`
		SaveAction    string   `yaml:"save_action"`
		Path          string   `yaml:"path"`
		Key           string   `yaml:"key"`
		SaveIf        string   `yaml:"save_if"`
		SaveKey       string   `yaml:"save_key"`
	} `yaml:"cache"`
	CheckJob      string `yaml:"check_job"`
	CheckIf       string `yaml:"check_if"`
	RaceSelective struct {
		Job  string `yaml:"job"`
		Step string `yaml:"step"`
		If   string `yaml:"if"`
		Run  string `yaml:"run"`
	} `yaml:"race_selective"`
	InheritStep struct {
		Job string `yaml:"job"`
		ID  string `yaml:"id"`
	} `yaml:"inherit_step"`
}

// evalMarker evaluates the small boolean expression subset used by marker
// conditions, failing closed if a workflow introduces a different expression.
func evalMarker(expr string, values map[string]string) (bool, error) {
	quoted := regexp.MustCompile(`'([^']*)'`).ReplaceAllString(expr, `"$1"`)
	node, err := parser.ParseExpr(quoted)
	if err != nil {
		return false, err
	}
	var eval func(ast.Expr) (string, error)
	eval = func(n ast.Expr) (string, error) {
		switch v := n.(type) {
		case *ast.ParenExpr:
			return eval(v.X)
		case *ast.BasicLit:
			if v.Kind == token.STRING {
				return strconv.Unquote(v.Value)
			}
		case *ast.SelectorExpr:
			var name func(ast.Expr) string
			name = func(e ast.Expr) string {
				switch x := e.(type) {
				case *ast.Ident:
					return x.Name
				case *ast.SelectorExpr:
					return name(x.X) + "." + x.Sel.Name
				}
				return ""
			}
			if value, ok := values[name(v)]; ok {
				return value, nil
			}
		case *ast.BinaryExpr:
			left, err := eval(v.X)
			if err != nil {
				return "", err
			}
			right, err := eval(v.Y)
			if err != nil {
				return "", err
			}
			switch v.Op {
			case token.EQL:
				return strconv.FormatBool(left == right), nil
			case token.NEQ:
				return strconv.FormatBool(left != right), nil
			case token.LAND, token.LOR:
				a, e1 := strconv.ParseBool(left)
				b, e2 := strconv.ParseBool(right)
				if e1 != nil || e2 != nil {
					return "", fmt.Errorf("non-boolean operand")
				}
				if v.Op == token.LAND {
					return strconv.FormatBool(a && b), nil
				}
				return strconv.FormatBool(a || b), nil
			}
		}
		return "", fmt.Errorf("unsupported marker expression %T", n)
	}
	value, err := eval(node)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value)
}

func workflowScalar(node *yaml.Node, key string) string {
	if v := workflowMappingChild(node, key); v != nil {
		return v.Value
	}
	return ""
}

// TestTestsWorkflowWiring pins the tests.yml wiring that verdict inheritance
// and the selective race pass depend on: the aggregator's marker steps (their
// names equal infra's versioned set, mirrored in the fixture) and conditions,
// the full-suite job's condition, the proof step's shape, and the race pass
// running on merge groups only.
func TestTestsWorkflowWiring(t *testing.T) {
	var want testsWorkflowWiring
	decoder := yaml.NewDecoder(bytes.NewReader(testsWorkflowWiringYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&want); err != nil {
		t.Fatalf("decode the tests.yml wiring fixture: %v", err)
	}
	if want.VerdictMarkers.Version < 1 || len(want.VerdictMarkers.Names) == 0 {
		t.Fatal("the wiring fixture mirrors no verdict marker set")
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "tests.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil || len(doc.Content) == 0 {
		t.Fatalf("parse tests.yml: %v", err)
	}
	jobs := workflowMappingChild(doc.Content[0], "jobs")

	var names []string
	for _, m := range want.MarkerSteps {
		names = append(names, m.Name)
	}
	if !slices.Equal(names, want.VerdictMarkers.Names) {
		t.Fatalf("marker steps %q differ from infra's marker set version %d %q", names, want.VerdictMarkers.Version, want.VerdictMarkers.Names)
	}
	steps := workflowMappingChild(workflowMappingChild(jobs, want.AggregatorJob), "steps")
	needs := workflowMappingChild(workflowMappingChild(jobs, want.AggregatorJob), "needs")
	for _, expected := range want.AggregatorNeeds {
		found := false
		if needs != nil {
			for _, need := range needs.Content {
				found = found || need.Value == expected
			}
		}
		if !found {
			t.Errorf("jobs.%s does not depend on %s", want.AggregatorJob, expected)
		}
	}
	if steps == nil || len(steps.Content) < len(want.MarkerSteps) {
		t.Fatalf("jobs.%s has fewer steps than the %d marker steps", want.AggregatorJob, len(want.MarkerSteps))
	}
	tail := steps.Content[len(steps.Content)-len(want.MarkerSteps):]
	for i, m := range want.MarkerSteps {
		if got := workflowScalar(tail[i], "name"); got != m.Name {
			t.Errorf("jobs.%s final step %d is %q, want %q", want.AggregatorJob, i, got, m.Name)
		}
		if got := workflowScalar(tail[i], "if"); got != m.If {
			t.Errorf("marker step %q if = %q, want %q", m.Name, got, m.If)
		}
	}
	for _, c := range want.MarkerCases {
		t.Run(c.Name, func(t *testing.T) {
			values := map[string]string{
				"github.event_name":             c.Event,
				"needs.check.result":            c.Check,
				"needs.race.result":             c.Race,
				"needs.changes.outputs.inherit": c.Inherit,
			}
			var selected []string
			for _, step := range tail {
				ok, err := evalMarker(workflowScalar(step, "if"), values)
				if err != nil {
					t.Fatal(err)
				}
				if ok {
					selected = append(selected, workflowScalar(step, "name"))
				}
			}
			if len(selected) != 1 || selected[0] != c.Expected {
				t.Errorf("selected markers = %q, want %q", selected, c.Expected)
			}
		})
	}

	saveCount := 0
	for _, jobName := range want.Cache.Jobs {
		jobSteps := workflowMappingChild(workflowMappingChild(jobs, jobName), "steps")
		restores := 0
		for _, step := range jobSteps.Content {
			action := workflowScalar(step, "uses")
			if action == want.Cache.RestoreAction {
				restores++
				with := workflowMappingChild(step, "with")
				if got := workflowScalar(with, "path"); got != want.Cache.Path {
					t.Errorf("jobs.%s cache path = %q, want %q", jobName, got, want.Cache.Path)
				}
				if got := workflowScalar(with, "key"); got != want.Cache.Key {
					t.Errorf("jobs.%s cache key = %q, want %q", jobName, got, want.Cache.Key)
				}
			}
			if action == want.Cache.SaveAction {
				saveCount++
				if jobName != want.CheckJob || workflowScalar(step, "if") != want.Cache.SaveIf {
					t.Errorf("unexpected cache save in jobs.%s with if %q", jobName, workflowScalar(step, "if"))
				}
				if got := workflowScalar(workflowMappingChild(step, "with"), "path"); got != want.Cache.Path {
					t.Errorf("jobs.%s save path = %q, want %q", jobName, got, want.Cache.Path)
				}
				if got := workflowScalar(workflowMappingChild(step, "with"), "key"); got != want.Cache.SaveKey {
					t.Errorf("jobs.%s save key = %q, want %q", jobName, got, want.Cache.SaveKey)
				}
			}
		}
		if restores != 1 {
			t.Errorf("jobs.%s has %d cache restores, want one", jobName, restores)
		}
	}
	if saveCount != 1 {
		t.Errorf("check and race have %d cache saves, want one", saveCount)
	}

	check := workflowMappingChild(jobs, want.CheckJob)
	if got := workflowScalar(check, "if"); got != want.CheckIf {
		t.Errorf("jobs.%s if = %q, want %q", want.CheckJob, got, want.CheckIf)
	}

	raceJob := workflowMappingChild(jobs, want.RaceSelective.Job)
	if got := workflowScalar(raceJob, "if"); got != want.RaceSelective.If {
		t.Errorf("jobs.%s if = %q, want %q", want.RaceSelective.Job, got, want.RaceSelective.If)
	}

	raceSteps := 0
	for i := 0; i+1 < len(jobs.Content); i += 2 {
		job := jobs.Content[i].Value
		jobSteps := workflowMappingChild(jobs.Content[i+1], "steps")
		if jobSteps == nil {
			continue // a reusable-workflow call has no steps of its own
		}
		for _, step := range jobSteps.Content {
			if !strings.Contains(workflowScalar(step, "run"), "race-selective") {
				continue
			}
			raceSteps++
			if job != want.RaceSelective.Job || workflowScalar(step, "name") != want.RaceSelective.Step {
				t.Errorf("race-selective runs in jobs.%s step %q, want only jobs.%s step %q", job, workflowScalar(step, "name"), want.RaceSelective.Job, want.RaceSelective.Step)
			}
			if got := workflowScalar(step, "if"); got != "" {
				t.Errorf("race-selective step unexpectedly has if = %q", got)
			}
			if got := strings.TrimSpace(workflowScalar(step, "run")); got != want.RaceSelective.Run {
				t.Errorf("race-selective step runs %q, want %q", got, want.RaceSelective.Run)
			}
		}
	}
	if raceSteps != 1 {
		t.Errorf("tests.yml runs race-selective in %d steps, want exactly one", raceSteps)
	}

	found := false
	for _, step := range workflowMappingChild(workflowMappingChild(jobs, want.InheritStep.Job), "steps").Content {
		if workflowScalar(step, "id") != want.InheritStep.ID {
			continue
		}
		found = true
		if workflowMappingChild(step, "if") != nil || workflowMappingChild(step, "continue-on-error") != nil {
			t.Errorf("jobs.%s step %s must have no if and no continue-on-error", want.InheritStep.Job, want.InheritStep.ID)
		}
		if !strings.Contains(workflowScalar(step, "run"), "inherit-verdict") {
			t.Errorf("jobs.%s step %s does not run inherit-verdict", want.InheritStep.Job, want.InheritStep.ID)
		}
	}
	if !found {
		t.Errorf("jobs.%s has no step %s", want.InheritStep.Job, want.InheritStep.ID)
	}
}
