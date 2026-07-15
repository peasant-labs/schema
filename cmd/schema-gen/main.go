package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/peasant-labs/schema/openapi"

	"github.com/swaggest/openapi-go/openapi31"
)

func main() {
	moduleRoot, err := findModuleRoot()
	if err != nil {
		log.Fatal(err)
	}

	generatedDir := filepath.Join(moduleRoot, "generated")

	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		log.Fatal(err)
	}

	// --- Session-detail redaction fixture (testdata/session-detail/redactions.yaml) ---
	// Format-only serialisation of schema.RedactionExamples (never recomputes the
	// engine output). Regenerated here so one `go run ./cmd/schema-gen` covers it.
	if err := generateRedactions(moduleRoot); err != nil {
		log.Fatal(err)
	}
	log.Printf("Generated %s", filepath.Join(moduleRoot, redactionsYAMLRelPath))

	// --- OpenAPI JSON/YAML specs ---
	// Built once via the shared generator (also used by the codegen-freshness
	// test) and written to the in-module generated/ dir — this module root is the
	// single source of truth for the published specs.
	artifacts, err := openapi.GenerateSpecArtifacts()
	if err != nil {
		log.Fatal(err)
	}
	for filename, data := range artifacts {
		path := filepath.Join(generatedDir, filename)
		if err := os.WriteFile(path, data, 0644); err != nil {
			log.Fatal(err)
		}
		log.Printf("Generated %s", path)
	}

	// --- TypeScript package ---
	// Raw OpenAPI maps are generated with the pinned openapi-typescript tool,
	// then wrapped in collision-checked named exports. The quality fixture module
	// is rendered only after the canonical Go loader validates the source YAML.
	if err := generateTypeScript(moduleRoot); err != nil {
		log.Fatal(err)
	}
	log.Printf("Generated TypeScript package sources under %s", filepath.Join(moduleRoot, "typescript", "src"))

	// Specs still needed below for the HTML docs.
	villageSpec, err := openapi.BuildVillageAPISpec()
	if err != nil {
		log.Fatal(err)
	}
	peasantLocalSpec, err := openapi.BuildPeasantLocalAPISpec()
	if err != nil {
		log.Fatal(err)
	}
	sharedSpec, err := openapi.BuildTypesSpec()
	if err != nil {
		log.Fatal(err)
	}

	// --- HTML documentation (Redoc via CDN, no Node dependency) ---
	docsDir := filepath.Join(moduleRoot, "docs", "api")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Human-readable doc labels. The version suffix is DERIVED from the single-source
	// version consts in the schema root (re-exported via openapi; never retyped), so a
	// doc-surface semver bump flows automatically into the Redoc pages + index.
	villageLabel := "Village API v" + openapi.VillageAPIVersion
	peasantLocalLabel := "Local Dashboard API v" + openapi.PeasantLocalAPIVersion
	typesLabel := "Types v" + openapi.TypesVersion

	// Redoc pages for specs with operations.
	redocSpecs := []struct {
		spec  *openapi31.Spec
		name  string
		label string
	}{
		{villageSpec, "village-api", villageLabel},
		{peasantLocalSpec, "peasantlocal-api", peasantLocalLabel},
	}
	for _, ds := range redocSpecs {
		if err := writeRedocHTML(ds.spec, ds.name, ds.label, docsDir); err != nil {
			log.Fatal(err)
		}
	}

	// Custom pages for type catalog and relationship graph (Redoc shows nothing for zero-path specs).
	if err := writeTypesHTML(sharedSpec, docsDir); err != nil {
		log.Fatal(err)
	}
	if err := writeTypeGraphHTML(villageSpec, docsDir); err != nil {
		log.Fatal(err)
	}

	indexEntries := []struct {
		spec  *openapi31.Spec
		name  string
		label string
		desc  string
	}{
		{villageSpec, "village-api", villageLabel, "Transcript publishing + CLI auth"},
		{peasantLocalSpec, "peasantlocal-api", peasantLocalLabel, "REST + WebSocket"},
		{sharedSpec, "types", typesLabel, "Domain type catalog + relationship diagram"},
		{nil, "type-graph", "Type Relationship Graph", "Full dependency graph across all types"},
	}
	if err := writeIndexHTML(indexEntries, docsDir); err != nil {
		log.Fatal(err)
	}

	// NOTE: the peasant CLI-reference docgen block (`go run ./cmd/peasant docgen`)
	// was intentionally NOT ported (a7) — the schema repo ships no peasant binary.
	// The Redoc/HTML spec docs above are the schema repo's doc surface.
}

// writeRedocHTML generates a self-contained HTML file that renders an OpenAPI spec
// using Redoc loaded from CDN. The spec JSON is inlined so the file works with file:// URLs.
func writeRedocHTML(spec *openapi31.Spec, name, title, dir string) error {
	jsonData, err := spec.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshal %s JSON for HTML: %w", name, err)
	}

	tmpl := template.Must(template.New("redoc").Parse(`<!DOCTYPE html>
<html><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>body { margin: 0; padding: 0; }</style>
</head><body>
<div id="redoc-container"></div>
<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
<script>
Redoc.init({{.SpecJSON}}, {
  scrollYOffset: 0,
  hideDownloadButton: false,
  expandResponses: "200"
}, document.getElementById('redoc-container'));
</script>
</body></html>
`))

	outPath := filepath.Join(dir, name+".html")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, struct {
		Title    string
		SpecJSON template.JS
	}{
		Title:    title,
		SpecJSON: template.JS(jsonData),
	}); err != nil {
		return fmt.Errorf("execute template %s: %w", name, err)
	}

	log.Printf("Generated %s", outPath)
	return nil
}

// writeIndexHTML generates an index page linking to all generated API doc HTML files.
func writeIndexHTML(specs []struct {
	spec  *openapi31.Spec
	name  string
	label string
	desc  string
}, dir string) error {
	tmpl := template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Peasant API Documentation</title>
<style>
body { font-family: system-ui, sans-serif; max-width: 600px; margin: 80px auto; color: #1a1a1a; padding: 0 1em; }
a { color: #2563eb; text-decoration: none; }
a:hover { text-decoration: underline; }
li { margin: 0.5em 0; }
</style>
</head><body>
<h1>Peasant API Documentation</h1>
<ul>
{{range .}}<li><a href="{{.Name}}.html">{{.Label}}</a> &mdash; {{.Desc}}</li>
{{end}}</ul>
</body></html>
`))

	type entry struct{ Name, Label, Desc string }
	entries := make([]entry, len(specs))
	for i, s := range specs {
		entries[i] = entry{s.name, s.label, s.desc}
	}

	outPath := filepath.Join(dir, "index.html")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, entries); err != nil {
		return fmt.Errorf("execute index template: %w", err)
	}

	log.Printf("Generated %s", outPath)
	return nil
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
