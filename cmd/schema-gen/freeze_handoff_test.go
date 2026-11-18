package main

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/freeze_handoff.yaml
var freezeHandoffCases []byte

//go:embed testdata/freeze_handoff_manifest.yaml
var freezeHandoffManifest []byte

type freezeHandoffInput struct {
	Edit    freezeFixtureEdit `yaml:"edit"`
	Sibling string            `yaml:"sibling"`
}

type freezeHandoffExpected struct {
	ErrorContains string `yaml:"error_contains"`
}

func TestFreezeHandoffValidation(t *testing.T) {
	corpus, err := testcase.LoadCorpus[freezeHandoffInput, freezeHandoffExpected](freezeHandoffCases)
	if err != nil {
		t.Fatal(err)
	}
	caseassert.RequireValid(t, corpus)
	var names []string
	accepts := make(map[string]bool)
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
		accepts[c.Name] = c.Classification == testcase.MustPass
	}
	caseassert.RequireManifestNames(t, "freeze handoff", freezeHandoffManifest, names)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Classification == testcase.MustFail && !accepts[c.Input.Sibling] {
				t.Fatal("rejection case has no accepting sibling")
			}
			state := surfaceStates{Status: 1}
			plan := freezePlan{Plan: 1}
			candidate := surfaceStates{Status: 1}
			for _, surface := range reportSurfaces {
				ahead := surface.name == "types"
				kind, level, to := surfaceClean, freezeNone, "1.2.3"
				if ahead {
					kind, level, to = surfaceAhead, freezeMinor, "1.3.0"
				}
				state.Surfaces = append(state.Surfaces, surfaceState{surface.name, surface.label, "1.2.3", kind})
				plan.Surfaces = append(plan.Surfaces, freezeSurface{surface.name, surface.label, "1.2.3", to, level, false, ahead})
				candidate.Surfaces = append(candidate.Surfaces, surfaceState{surface.name, surface.label, to, surfaceClean})
			}
			switch c.Input.Edit {
			case freezeFixtureUnchanged:
			case "stale-source":
				state.Surfaces[0].Frozen = "1.2.4"
			case "wrong-next":
				plan.Surfaces[0].To = "1.9.0"
			case "missing-surface":
				plan.Surfaces = plan.Surfaces[:len(plan.Surfaces)-1]
			case "unplanned-ahead":
				state.Surfaces[1].State = surfaceAhead
			case "segmented-freeze":
				i := len(plan.Surfaces) - 1
				state.Surfaces[i].State = surfaceAhead
				plan.Surfaces[i].Freeze = true
				plan.Surfaces[i].Level = freezeMinor
				plan.Surfaces[i].To = "1.3.0"
			case "wrong-candidate":
				candidate.Surfaces[0].Frozen = "1.4.0"
			default:
				t.Fatalf("unknown freeze handoff edit %q", c.Input.Edit)
			}
			err := validateFreezeHandoff(reportSurfaces, state, plan)
			if err == nil {
				err = checkFreezeCandidate(candidate, plan)
			}
			if c.Classification == testcase.MustPass && err != nil {
				t.Fatalf("valid handoff rejected: %v", err)
			}
			if c.Classification == testcase.MustFail && (err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains)) {
				t.Fatalf("handoff error = %v, want %q", err, c.Expected.ErrorContains)
			}
		})
	}
}
