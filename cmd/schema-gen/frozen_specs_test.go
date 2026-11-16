package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/peasant-labs/schema/openapi"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

func loadServedFrozenSpecRegistry(t *testing.T, root string) []retiredSpec {
	t.Helper()
	path := filepath.Join(root, "cmd", "schema-gen", "testdata", "frozen_specs.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read served frozen spec registry %s: %v", path, err)
	}
	specs, err := decodeRetiredSpecRegistry(data)
	if err != nil {
		t.Fatalf("decode served frozen spec registry %s: %v", path, err)
	}
	manifest, err := os.ReadFile(filepath.Join(root, "cmd", "schema-gen", "testdata", "frozen_specs_manifest.yaml"))
	if err != nil {
		t.Fatalf("read served frozen spec manifest: %v", err)
	}
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.Name)
	}
	caseassert.RequireManifestNames(t, "served frozen spec", manifest, names)
	return specs
}

func TestServedSpecArtifactsFrozen(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		t.Fatal(err)
	}
	retired := make(map[string]struct{})
	for _, spec := range loadRetiredSpecRegistry(t, root) {
		retired[spec.Name] = struct{}{}
	}
	for _, spec := range loadServedFrozenSpecRegistry(t, root) {
		if _, ok := retired[spec.Name]; ok {
			t.Errorf("served frozen spec %q is also retired; a served version must not be retired", spec.Name)
		}
		for filename, expected := range map[string]string{
			spec.Name + ".json": spec.JSONSHA256,
			spec.Name + ".yaml": spec.YAMLSHA256,
		} {
			generated, ok := artifacts[filename]
			if !ok {
				t.Errorf("served frozen spec %q is no longer generated; retain it until retirement", filename)
				continue
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(generated)); got != expected {
				t.Errorf("served frozen spec %q regenerates to sha256 %s, want %s; a shared declaration changed the still-served protocol; preserve the frozen shape before generating", filename, got, expected)
			}
		}
		for _, dir := range committedSpecDirs(root) {
			assertFrozen(t, filepath.Join(dir, spec.Name+".json"), spec.JSONSHA256)
			assertFrozen(t, filepath.Join(dir, spec.Name+".yaml"), spec.YAMLSHA256)
		}
	}
}
