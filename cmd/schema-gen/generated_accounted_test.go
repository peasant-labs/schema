package main

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/openapi"
)

// TestGeneratedDirFullyAccounted is the PARTITION-COMPLETENESS guard over the
// committed generated/ spec dir. Every spec file in it MUST be accounted for as
// exactly one of:
//   - CURRENT-generated — a key of openapi.GenerateSpecArtifacts(), the set the
//     generator emits today (derived from the version consts in versions.go); or
//   - RETIRED-registered — a retiredSpecRegistry entry, byte-frozen under
//     TestRetiredSpecsImmutable.
//
// It closes the silent-gap class the two existing guards leave open BETWEEN them:
// TestCodegenFreshness_SpecsMatchSource only iterates the CURRENT artifacts map (a
// retired file it no longer emits is invisible to it), and TestRetiredSpecsImmutable
// only iterates the registry (a file that is neither current nor registered is
// invisible to it). So a version bump that RETIRES a spec without REGISTERING it
// leaves that file mutable-and-unguarded, and NEITHER existing gate fails. This guard
// fails loudly on exactly that: an unaccounted spec in generated/ — a
// retired-but-unregistered version, or a stray committed file.
func TestGeneratedDirFullyAccounted(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}

	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatalf("GenerateSpecArtifacts: %v", err)
	}
	retiredSpecRegistry := loadRetiredSpecRegistry(t, root)

	// accounted = current-generated ∪ retired-registered (both keyed by filename).
	accounted := make(map[string]bool, len(artifacts)+2*len(retiredSpecRegistry))
	for filename := range artifacts {
		accounted[filename] = true
	}
	for _, rs := range retiredSpecRegistry {
		accounted[rs.Name+".json"] = true
		if !rs.JSONOnly {
			accounted[rs.Name+".yaml"] = true
		}
	}

	for _, dir := range committedSpecDirs(root) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read committed spec dir %s: %v", dir, err)
		}
		var unaccounted []string
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			// generated/ holds only spec artifacts (.json / .yaml); guard those.
			if !strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".yaml") {
				continue
			}
			if !accounted[name] {
				unaccounted = append(unaccounted, name)
			}
		}
		if len(unaccounted) > 0 {
			sort.Strings(unaccounted)
			t.Errorf(
				"UNACCOUNTED spec file(s) in %s: %v\n"+
					"  what: a committed spec is neither CURRENT-generated (a key of openapi.GenerateSpecArtifacts(), derived from versions.go) nor RETIRED-registered (retiredSpecRegistry).\n"+
					"  why:  a version bump retired a spec without registering it (register-at-freeze-time was skipped), or a stray file was committed.\n"+
					"  fix:  if it is a newly-retired version, add it to retiredSpecRegistry pinned to its frozen sha256; if it is stray, remove it.",
				dir, unaccounted)
		}
	}
}
