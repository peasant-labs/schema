package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/swaggest/openapi-go/openapi31"
)

// ---------------------------------------------------------------------------
// Data types
// ---------------------------------------------------------------------------

type schemaInfo struct {
	Name        string
	Description string
	SchemaType  string // "object", "enum", "string", "integer", etc.
	Category    string // "composite", "enum", "identifier", "primitive"
	Title       string
	Enum        []string
	Properties  []propertyInfo
	Pattern     string
	Format      string
	Examples    []string
}

type propertyInfo struct {
	Name     string
	Type     string // display type: "string", "Provider", "[]SessionEntry"
	Required bool
	Ref      string // non-empty if linked to another schema component
	IsArray  bool
}

type typeRelationship struct {
	From       string
	To         string
	Field      string
	IsArray    bool
	IsOptional bool
}

type schemaCategory struct {
	Label   string
	ID      string
	Schemas []schemaInfo
}

// ---------------------------------------------------------------------------
// Extraction
// ---------------------------------------------------------------------------

func stripSchemaPrefix(name string) string {
	return strings.TrimPrefix(name, "Schema")
}

// extractFromSpec parses a marshalled OpenAPI spec and returns schema info and
// $ref-based relationships between component schemas.
func extractFromSpec(spec *openapi31.Spec) ([]schemaInfo, []typeRelationship, error) {
	jsonData, err := spec.MarshalJSON()
	if err != nil {
		return nil, nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return nil, nil, err
	}

	comps, _ := raw["components"].(map[string]any)
	schemas, _ := comps["schemas"].(map[string]any)

	var infos []schemaInfo
	var rels []typeRelationship

	for rawName, sRaw := range schemas {
		s, ok := sRaw.(map[string]any)
		if !ok {
			continue
		}

		name := stripSchemaPrefix(rawName)
		info := schemaInfo{Name: name}

		if title, ok := s["title"].(string); ok {
			info.Title = title
		}
		if desc, ok := s["description"].(string); ok {
			info.Description = desc
		}
		if pattern, ok := s["pattern"].(string); ok {
			info.Pattern = pattern
		}
		if format, ok := s["format"].(string); ok {
			info.Format = format
		}

		// Classify type and category.
		if enumVals, ok := s["enum"].([]any); ok {
			info.SchemaType = "enum"
			info.Category = "enum"
			for _, v := range enumVals {
				if str, ok := v.(string); ok {
					info.Enum = append(info.Enum, str)
				}
			}
		} else if _, ok := s["properties"]; ok {
			info.SchemaType = "object"
			info.Category = "composite"
		} else if t, ok := s["type"].(string); ok {
			info.SchemaType = t
			if t == "string" && (info.Pattern != "" || info.Format != "") {
				info.Category = "identifier"
			} else {
				info.Category = "primitive"
			}
		} else {
			info.SchemaType = "unknown"
			info.Category = "primitive"
		}

		// Examples.
		if examples, ok := s["examples"].([]any); ok {
			for _, ex := range examples {
				info.Examples = append(info.Examples, fmt.Sprintf("%v", ex))
			}
		}

		// Required fields set.
		requiredSet := make(map[string]bool)
		if reqArr, ok := s["required"].([]any); ok {
			for _, r := range reqArr {
				if rs, ok := r.(string); ok {
					requiredSet[rs] = true
				}
			}
		}

		// Properties and $ref relationships.
		if props, ok := s["properties"].(map[string]any); ok {
			for pName, pRaw := range props {
				p, ok := pRaw.(map[string]any)
				if !ok {
					continue
				}

				pi := propertyInfo{
					Name:     pName,
					Required: requiredSet[pName],
				}

				if ref, ok := p["$ref"].(string); ok {
					refName := stripSchemaPrefix(strings.TrimPrefix(ref, "#/components/schemas/"))
					pi.Ref = refName
					pi.Type = refName
					rels = append(rels, typeRelationship{
						From:       name,
						To:         refName,
						Field:      pName,
						IsOptional: !requiredSet[pName],
					})
				} else if items, ok := p["items"].(map[string]any); ok {
					if ref, ok := items["$ref"].(string); ok {
						refName := stripSchemaPrefix(strings.TrimPrefix(ref, "#/components/schemas/"))
						pi.Ref = refName
						pi.Type = refName
						pi.IsArray = true
						rels = append(rels, typeRelationship{
							From:       name,
							To:         refName,
							Field:      pName,
							IsArray:    true,
							IsOptional: !requiredSet[pName],
						})
					} else if t, ok := items["type"].(string); ok {
						pi.Type = t
						pi.IsArray = true
					}
				} else {
					pi.Type = extractPropertyType(p)
				}

				info.Properties = append(info.Properties, pi)
			}
			sort.Slice(info.Properties, func(i, j int) bool {
				// Required fields first, then alphabetical.
				if info.Properties[i].Required != info.Properties[j].Required {
					return info.Properties[i].Required
				}
				return info.Properties[i].Name < info.Properties[j].Name
			})
		}

		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})
	sort.Slice(rels, func(i, j int) bool {
		if rels[i].From != rels[j].From {
			return rels[i].From < rels[j].From
		}
		return rels[i].Field < rels[j].Field
	})

	return infos, rels, nil
}

// extractPropertyType returns a display string for a JSON Schema property.
func extractPropertyType(p map[string]any) string {
	if t, ok := p["type"].(string); ok {
		return t
	}
	if types, ok := p["type"].([]any); ok {
		for _, t := range types {
			if ts, ok := t.(string); ok && ts != "null" {
				return ts
			}
		}
	}
	return "any"
}

// ---------------------------------------------------------------------------
// Mermaid ER diagram generation
// ---------------------------------------------------------------------------

func generateMermaidER(rels []typeRelationship) string {
	var b strings.Builder
	b.WriteString("erDiagram\n")

	seen := make(map[string]bool)
	for _, r := range rels {
		key := r.From + "|" + r.To + "|" + r.Field
		if seen[key] {
			continue
		}
		seen[key] = true

		var notation string
		switch {
		case r.IsArray && r.IsOptional:
			notation = "||--o{"
		case r.IsArray:
			notation = "||--|{"
		case r.IsOptional:
			notation = "||--o|"
		default:
			notation = "||--||"
		}

		b.WriteString(fmt.Sprintf("    %s %s %s : %s\n", r.From, notation, r.To, r.Field))
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// Categorisation
// ---------------------------------------------------------------------------

func categorizeSchemas(schemas []schemaInfo) []schemaCategory {
	groups := map[string][]schemaInfo{}
	for _, s := range schemas {
		groups[s.Category] = append(groups[s.Category], s)
	}

	order := []struct{ key, label, id string }{
		{"composite", "Composites", "composites"},
		{"enum", "Enums", "enums"},
		{"identifier", "Identifiers", "identifiers"},
		{"primitive", "Primitives", "primitives"},
	}

	var cats []schemaCategory
	for _, o := range order {
		if g, ok := groups[o.key]; ok && len(g) > 0 {
			cats = append(cats, schemaCategory{Label: o.label, ID: o.id, Schemas: g})
		}
	}
	return cats
}

// ---------------------------------------------------------------------------
// HTML writers
// ---------------------------------------------------------------------------

// typesData is the template context for the types page.
type typesData struct {
	Title         string
	Version       string
	Description   string
	MermaidSource template.HTML
	Categories    []schemaCategory
	TotalTypes    int
	TotalRels     int
}

// typeGraphData is the template context for the combined type-graph page.
type typeGraphData struct {
	Title         string
	Subtitle      string
	MermaidSource template.HTML
	TotalTypes    int
	TotalRels     int
}

var templateFuncs = template.FuncMap{
	"add": func(a, b int) int { return a + b },
}

func writeTypesHTML(spec *openapi31.Spec, dir string) error {
	schemas, rels, err := extractFromSpec(spec)
	if err != nil {
		return fmt.Errorf("extract shared types: %w", err)
	}

	data := typesData{
		Title:         "Peasant Types",
		Version:       "1.0.0",
		Description:   "Foundational domain types shared across all Peasant APIs. These components define the vocabulary for sessions, providers, quality metrics, and content entries.",
		MermaidSource: template.HTML(generateMermaidER(rels)),
		Categories:    categorizeSchemas(schemas),
		TotalTypes:    len(schemas),
		TotalRels:     len(rels),
	}

	tmpl := template.Must(template.New("types").Funcs(templateFuncs).Parse(typesHTMLTemplate))

	outPath := filepath.Join(dir, "types.html")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute types template: %w", err)
	}

	log.Printf("Generated %s", outPath)
	return nil
}

func writeTypeGraphHTML(spec *openapi31.Spec, dir string) error {
	schemas, rels, err := extractFromSpec(spec)
	if err != nil {
		return fmt.Errorf("extract type graph: %w", err)
	}
	_ = schemas // used only for count

	data := typeGraphData{
		Title:         "Peasant Type Relationship Graph",
		Subtitle:      "Complete type dependency graph across the Village API, showing how PublishRequest composes all domain types.",
		MermaidSource: template.HTML(generateMermaidER(rels)),
		TotalTypes:    len(schemas),
		TotalRels:     len(rels),
	}

	tmpl := template.Must(template.New("type-graph").Parse(typeGraphHTMLTemplate))

	outPath := filepath.Join(dir, "type-graph.html")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute type-graph template: %w", err)
	}

	log.Printf("Generated %s", outPath)
	return nil
}

// ---------------------------------------------------------------------------
// HTML templates
// ---------------------------------------------------------------------------

const typesHTMLTemplate = `<!DOCTYPE html>
<html lang="en"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} ({{.Version}})</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Instrument+Serif:ital@0;1&family=DM+Sans:opsz,wght@9..40,300;9..40,400;9..40,500;9..40,600&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
<style>
:root {
  --bg: #faf9f7;
  --text: #2a2521;
  --text-dim: #857a6b;
  --text-faint: #b8b0a1;
  --primary: #00a288;
  --primary-dark: #00816e;
  --primary-light: #effefa;
  --primary-border: #c7fff2;
  --accent-amber: #b8860b;
  --accent-amber-bg: #fdf8ef;
  --accent-amber-border: #f0dca0;
  --accent-slate: #4a6fa5;
  --accent-slate-bg: #eef3fa;
  --accent-slate-border: #b8cfe6;
  --card-bg: #fff;
  --card-border: #e8e4dd;
  --card-shadow: 0 1px 3px rgba(0,0,0,0.03), 0 6px 16px rgba(4,200,165,0.04);
  --card-shadow-hover: 0 2px 6px rgba(0,0,0,0.05), 0 12px 28px rgba(4,200,165,0.08);
  --code-bg: #f5f2ee;
  --table-stripe: #faf8f5;
  --radius: 10px;
  --radius-sm: 6px;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: "DM Sans", system-ui, sans-serif;
  font-size: 15px;
  color: var(--text);
  background-color: var(--bg);
  background-image:
    linear-gradient(rgba(0,51,48,0.028) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0,51,48,0.028) 1px, transparent 1px);
  background-size: 48px 48px;
  -webkit-font-smoothing: antialiased;
  line-height: 1.6;
}

::selection { background: #90ffe6; color: #003330; }

a { color: var(--primary-dark); text-decoration: none; }
a:hover { text-decoration: underline; }

.page { max-width: 1100px; margin: 0 auto; padding: 0 2rem; }

/* ------- Header ------- */
.header {
  padding: 4rem 0 2.5rem;
  border-bottom: 1px solid var(--card-border);
  margin-bottom: 3rem;
}
.header-top {
  display: flex;
  align-items: baseline;
  gap: 1rem;
  margin-bottom: 0.75rem;
}
.header h1 {
  font-family: "Instrument Serif", Georgia, serif;
  font-size: 2.5rem;
  font-weight: 400;
  line-height: 1.15;
  color: var(--text);
}
.version-badge {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.7rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--primary-dark);
  background: var(--primary-light);
  border: 1px solid var(--primary-border);
  padding: 3px 10px;
  border-radius: 4px;
}
.header p {
  color: var(--text-dim);
  max-width: 640px;
  line-height: 1.7;
}
.header-stats {
  display: flex;
  gap: 2rem;
  margin-top: 1.25rem;
}
.stat {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.75rem;
  color: var(--text-faint);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
.stat strong {
  color: var(--text-dim);
  font-weight: 500;
}

/* ------- Nav chips ------- */
.type-nav {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 2.5rem;
}
.type-nav a {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.72rem;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 4px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  color: var(--text-dim);
  transition: all 0.15s ease;
}
.type-nav a:hover {
  text-decoration: none;
  border-color: var(--primary-border);
  background: var(--primary-light);
  color: var(--primary-dark);
}
.type-nav .nav-sep {
  width: 1px;
  background: var(--card-border);
  margin: 0 0.25rem;
  align-self: stretch;
}

/* ------- Diagram section ------- */
.diagram-section {
  margin-bottom: 3.5rem;
}
.diagram-section h2 {
  font-family: "DM Sans", sans-serif;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--primary-dark);
  margin-bottom: 1rem;
}
.diagram-card {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: var(--radius);
  padding: 2.5rem 2rem;
  box-shadow: var(--card-shadow);
  overflow-x: auto;
}
.diagram-card .mermaid {
  display: flex;
  justify-content: center;
}
.diagram-card .mermaid svg {
  max-width: 100%;
  height: auto;
}

/* ------- Category sections ------- */
.category {
  margin-bottom: 3rem;
}
.category-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--primary-border);
}
.category-header h2 {
  font-family: "DM Sans", sans-serif;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--primary-dark);
}
.category-count {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.65rem;
  color: var(--text-faint);
}

/* ------- Schema cards ------- */
.schema-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.25rem;
}
.schema-card {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: var(--radius);
  padding: 1.5rem 1.75rem;
  box-shadow: var(--card-shadow);
  transition: box-shadow 0.2s ease, transform 0.2s ease;
  border-left: 3px solid var(--primary);
  animation: card-in 0.35s ease both;
}
.schema-card:hover {
  box-shadow: var(--card-shadow-hover);
  transform: translateY(-1px);
}
.schema-card.cat-enum { border-left-color: var(--accent-amber); }
.schema-card.cat-identifier { border-left-color: var(--accent-slate); }

.schema-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}
.schema-header h3 {
  font-family: "JetBrains Mono", monospace;
  font-size: 1.1rem;
  font-weight: 500;
  color: var(--text);
}
.type-badge {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.6rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 2px 8px;
  border-radius: 3px;
}
.badge-composite {
  color: var(--primary-dark);
  background: var(--primary-light);
  border: 1px solid var(--primary-border);
}
.badge-enum {
  color: var(--accent-amber);
  background: var(--accent-amber-bg);
  border: 1px solid var(--accent-amber-border);
}
.badge-identifier {
  color: var(--accent-slate);
  background: var(--accent-slate-bg);
  border: 1px solid var(--accent-slate-border);
}
.badge-primitive {
  color: var(--text-dim);
  background: var(--code-bg);
  border: 1px solid var(--card-border);
}

.schema-desc {
  color: var(--text-dim);
  font-size: 0.9rem;
  line-height: 1.6;
  margin-bottom: 1rem;
}

/* Property table */
.prop-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
  margin-bottom: 0.75rem;
}
.prop-table th {
  text-align: left;
  font-family: "JetBrains Mono", monospace;
  font-size: 0.65rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-faint);
  padding: 0.4rem 0.75rem;
  border-bottom: 1px solid var(--card-border);
}
.prop-table td {
  padding: 0.45rem 0.75rem;
  border-bottom: 1px solid #f0ede8;
  vertical-align: top;
}
.prop-table tr:nth-child(even) td { background: var(--table-stripe); }
.prop-table tr:last-child td { border-bottom: none; }
.prop-name {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--text);
}
.prop-type {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.78rem;
  color: var(--text-dim);
}
.prop-type a { color: var(--primary-dark); }
.required-dot {
  display: inline-block;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--primary);
  margin-left: 4px;
  vertical-align: middle;
}

/* Enum chips */
.enum-values {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-bottom: 0.75rem;
}
.enum-chip {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.75rem;
  font-weight: 400;
  color: var(--accent-amber);
  background: var(--accent-amber-bg);
  border: 1px solid var(--accent-amber-border);
  padding: 2px 10px;
  border-radius: 3px;
}

/* Constraints and examples */
.constraint, .examples {
  font-size: 0.82rem;
  color: var(--text-dim);
  margin-bottom: 0.4rem;
}
.constraint .label, .examples .label {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.68rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-faint);
  margin-right: 0.5rem;
}
.constraint code, .examples code {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.78rem;
  background: var(--code-bg);
  padding: 1px 6px;
  border-radius: 3px;
}
.example-val {
  color: var(--primary-dark);
}
.pattern-code {
  font-size: 0.72rem;
  word-break: break-all;
}

/* ------- Footer ------- */
.footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 3rem 0;
  border-top: 1px solid var(--card-border);
  margin-top: 2rem;
}
.footer-text {
  font-family: "Instrument Serif", Georgia, serif;
  font-size: 1rem;
  color: var(--text-faint);
  letter-spacing: 0.5px;
}
.footer-logo {
  position: relative;
  width: 20px;
  height: 20px;
}
.footer-logo span {
  position: absolute;
  border-radius: 50%;
  opacity: 0.4;
}
.footer-logo .c1 { width: 12px; height: 12px; top: 0; left: 0; background: rgba(29,228,190,0.6); }
.footer-logo .c2 { width: 10px; height: 10px; top: 4px; left: 4px; background: rgba(0,162,136,0.5); }
.footer-logo .c3 { width: 9px; height: 9px; top: 2px; left: 9px; background: rgba(80,245,212,0.45); }

.back-link {
  display: inline-block;
  font-family: "JetBrains Mono", monospace;
  font-size: 0.72rem;
  color: var(--text-faint);
  margin-bottom: 2rem;
}
.back-link:hover { color: var(--primary-dark); }

@keyframes card-in {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

html { scroll-behavior: smooth; }

@media (max-width: 640px) {
  .page { padding: 0 1rem; }
  .header h1 { font-size: 1.75rem; }
  .diagram-card { padding: 1rem; }
  .schema-card { padding: 1rem 1.25rem; }
}
</style>
</head>
<body>
<div class="page">
  <a href="index.html" class="back-link">&larr; All docs</a>

  <div class="header">
    <div class="header-top">
      <h1>{{.Title}}</h1>
      <span class="version-badge">v{{.Version}}</span>
    </div>
    <p>{{.Description}}</p>
    <div class="header-stats">
      <span class="stat"><strong>{{.TotalTypes}}</strong> types</span>
      <span class="stat"><strong>{{.TotalRels}}</strong> relationships</span>
    </div>
  </div>

  <nav class="type-nav">
    {{- range .Categories}}
    {{- range .Schemas}}
    <a href="#schema-{{.Name}}">{{.Name}}</a>
    {{- end}}
    {{- if not (eq .ID "primitives")}}<div class="nav-sep"></div>{{end}}
    {{- end}}
  </nav>

  <section class="diagram-section">
    <h2>Type Relationships</h2>
    <div class="diagram-card">
      <pre class="mermaid">{{.MermaidSource}}</pre>
    </div>
  </section>

  {{range $ci, $cat := .Categories}}
  <section class="category" id="cat-{{$cat.ID}}">
    <div class="category-header">
      <h2>{{$cat.Label}}</h2>
      <span class="category-count">{{len $cat.Schemas}}</span>
    </div>
    <div class="schema-grid">
      {{range $si, $s := $cat.Schemas}}
      <div class="schema-card cat-{{$cat.ID}}" id="schema-{{$s.Name}}" style="animation-delay: {{$si}}00ms">
        <div class="schema-header">
          <h3>{{$s.Name}}</h3>
          <span class="type-badge badge-{{$cat.ID}}">{{$s.SchemaType}}</span>
        </div>
        {{if $s.Description}}<p class="schema-desc">{{$s.Description}}</p>{{end}}

        {{if $s.Enum}}
        <div class="enum-values">
          {{range $s.Enum}}<code class="enum-chip">{{.}}</code>{{end}}
        </div>
        {{end}}

        {{if $s.Properties}}
        <table class="prop-table">
          <thead><tr><th>Field</th><th>Type</th><th></th></tr></thead>
          <tbody>
            {{range $s.Properties}}
            <tr>
              <td><span class="prop-name">{{.Name}}</span></td>
              <td><span class="prop-type">{{if .IsArray}}[]{{end}}{{if .Ref}}<a href="#schema-{{.Ref}}">{{.Ref}}</a>{{else}}{{.Type}}{{end}}</span></td>
              <td>{{if .Required}}<span class="required-dot" title="required"></span>{{end}}</td>
            </tr>
            {{end}}
          </tbody>
        </table>
        {{end}}

        {{if $s.Pattern}}<div class="constraint"><span class="label">Pattern</span><code class="pattern-code">{{$s.Pattern}}</code></div>{{end}}
        {{if $s.Format}}<div class="constraint"><span class="label">Format</span><code>{{$s.Format}}</code></div>{{end}}
        {{if $s.Examples}}
        <div class="examples"><span class="label">Examples</span>{{range $s.Examples}} <code class="example-val">{{.}}</code>{{end}}</div>
        {{end}}
      </div>
      {{end}}
    </div>
  </section>
  {{end}}

  <div class="footer">
    <div class="footer-logo"><span class="c1"></span><span class="c2"></span><span class="c3"></span></div>
    <span class="footer-text">peasant</span>
  </div>
</div>

<script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
<script>
mermaid.initialize({
  startOnLoad: true,
  theme: 'base',
  themeVariables: {
    primaryColor: '#effefa',
    primaryBorderColor: '#00a288',
    primaryTextColor: '#2a2521',
    lineColor: '#857a6b',
    secondaryColor: '#faf9f7',
    tertiaryColor: '#fff',
    fontFamily: '"JetBrains Mono", monospace',
    fontSize: '13px'
  },
  er: {
    layoutDirection: 'TB',
    entityPadding: 12,
    minEntityWidth: 80
  }
});
</script>
</body></html>
`

const typeGraphHTMLTemplate = `<!DOCTYPE html>
<html lang="en"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Instrument+Serif:ital@0;1&family=DM+Sans:opsz,wght@9..40,300;9..40,400;9..40,500;9..40,600&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
<style>
:root {
  --bg: #faf9f7;
  --text: #2a2521;
  --text-dim: #857a6b;
  --text-faint: #b8b0a1;
  --primary: #00a288;
  --primary-dark: #00816e;
  --primary-light: #effefa;
  --primary-border: #c7fff2;
  --card-bg: #fff;
  --card-border: #e8e4dd;
  --card-shadow: 0 1px 3px rgba(0,0,0,0.03), 0 6px 16px rgba(4,200,165,0.04);
}

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: "DM Sans", system-ui, sans-serif;
  font-size: 15px;
  color: var(--text);
  background-color: var(--bg);
  background-image:
    linear-gradient(rgba(0,51,48,0.028) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0,51,48,0.028) 1px, transparent 1px);
  background-size: 48px 48px;
  -webkit-font-smoothing: antialiased;
}

::selection { background: #90ffe6; color: #003330; }

a { color: var(--primary-dark); text-decoration: none; }
a:hover { text-decoration: underline; }

.page { max-width: 1200px; margin: 0 auto; padding: 0 2rem; }

.back-link {
  display: inline-block;
  font-family: "JetBrains Mono", monospace;
  font-size: 0.72rem;
  color: var(--text-faint);
  margin-bottom: 2rem;
  padding-top: 2rem;
}
.back-link:hover { color: var(--primary-dark); }

.header {
  padding-bottom: 2rem;
  border-bottom: 1px solid var(--card-border);
  margin-bottom: 2.5rem;
}
.header h1 {
  font-family: "Instrument Serif", Georgia, serif;
  font-size: 2.25rem;
  font-weight: 400;
  line-height: 1.15;
  margin-bottom: 0.75rem;
}
.header p {
  color: var(--text-dim);
  max-width: 700px;
  line-height: 1.7;
}
.header-stats {
  display: flex;
  gap: 2rem;
  margin-top: 1rem;
}
.stat {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.75rem;
  color: var(--text-faint);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
.stat strong {
  color: var(--text-dim);
  font-weight: 500;
}

.diagram-card {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  padding: 2.5rem 2rem;
  box-shadow: var(--card-shadow);
  overflow-x: auto;
  margin-bottom: 3rem;
}
.diagram-card .mermaid {
  display: flex;
  justify-content: center;
}
.diagram-card .mermaid svg {
  max-width: 100%;
  height: auto;
}

.legend {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  padding: 1.5rem 2rem;
  margin-bottom: 3rem;
}
.legend h3 {
  font-family: "DM Sans", sans-serif;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-faint);
  margin-bottom: 0.75rem;
}
.legend-items {
  display: flex;
  flex-wrap: wrap;
  gap: 1.5rem;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: "JetBrains Mono", monospace;
  font-size: 0.75rem;
  color: var(--text-dim);
}
.legend-line {
  width: 40px;
  height: 2px;
  background: var(--text-dim);
  position: relative;
}
.legend-line.dashed { background: none; border-top: 2px dashed var(--text-dim); }

.footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 3rem 0;
  border-top: 1px solid var(--card-border);
}
.footer-text {
  font-family: "Instrument Serif", Georgia, serif;
  font-size: 1rem;
  color: var(--text-faint);
  letter-spacing: 0.5px;
}
.footer-logo {
  position: relative;
  width: 20px;
  height: 20px;
}
.footer-logo span {
  position: absolute;
  border-radius: 50%;
  opacity: 0.4;
}
.footer-logo .c1 { width: 12px; height: 12px; top: 0; left: 0; background: rgba(29,228,190,0.6); }
.footer-logo .c2 { width: 10px; height: 10px; top: 4px; left: 4px; background: rgba(0,162,136,0.5); }
.footer-logo .c3 { width: 9px; height: 9px; top: 2px; left: 9px; background: rgba(80,245,212,0.45); }

@media (max-width: 640px) {
  .page { padding: 0 1rem; }
  .header h1 { font-size: 1.5rem; }
  .diagram-card { padding: 1rem; }
}
</style>
</head>
<body>
<div class="page">
  <a href="index.html" class="back-link">&larr; All docs</a>

  <div class="header">
    <h1>{{.Title}}</h1>
    <p>{{.Subtitle}}</p>
    <div class="header-stats">
      <span class="stat"><strong>{{.TotalTypes}}</strong> types</span>
      <span class="stat"><strong>{{.TotalRels}}</strong> relationships</span>
    </div>
  </div>

  <div class="diagram-card">
    <pre class="mermaid">{{.MermaidSource}}</pre>
  </div>

  <div class="legend">
    <h3>Notation</h3>
    <div class="legend-items">
      <div class="legend-item"><div class="legend-line"></div> required</div>
      <div class="legend-item"><div class="legend-line dashed"></div> optional</div>
      <div class="legend-item"><code>||--||</code> one-to-one</div>
      <div class="legend-item"><code>||--o|</code> one-to-optional</div>
      <div class="legend-item"><code>||--|{</code> one-to-many</div>
      <div class="legend-item"><code>||--o{</code> one-to-optional-many</div>
    </div>
  </div>

  <div class="footer">
    <div class="footer-logo"><span class="c1"></span><span class="c2"></span><span class="c3"></span></div>
    <span class="footer-text">peasant</span>
  </div>
</div>

<script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
<script>
mermaid.initialize({
  startOnLoad: true,
  theme: 'base',
  themeVariables: {
    primaryColor: '#effefa',
    primaryBorderColor: '#00a288',
    primaryTextColor: '#2a2521',
    lineColor: '#857a6b',
    secondaryColor: '#faf9f7',
    tertiaryColor: '#fff',
    fontFamily: '"JetBrains Mono", monospace',
    fontSize: '12px'
  },
  er: {
    layoutDirection: 'TB',
    entityPadding: 12,
    minEntityWidth: 80
  }
});
</script>
</body></html>
`
