package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/peasant-labs/schema/openapi"
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
		for _, ext := range []string{".json", ".yaml"} {
			filename := spec.Name + ext
			if _, ok := artifacts[filename]; !ok {
				t.Errorf("served frozen spec %q is no longer generated; retain it until retirement", filename)
			}
		}
		for _, dir := range committedSpecDirs(root) {
			assertFrozen(t, filepath.Join(dir, spec.Name+".json"), spec.JSONSHA256)
			assertFrozen(t, filepath.Join(dir, spec.Name+".yaml"), spec.YAMLSHA256)
		}
	}
}
