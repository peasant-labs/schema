// Command redaction-categories projects the released engine's public bindings.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/peasant-labs/redact"
)

type member struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type projection struct {
	Module     string      `json:"module"`
	Version    string      `json:"version"`
	Sum        string      `json:"sum"`
	Categories []member    `json:"categories"`
	Labels     []member    `json:"labels"`
	Pairs      [][2]string `json:"pairs"`
}

func project() (projection, error) {
	p := projection{Module: "github.com/peasant-labs/redact"}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return p, fmt.Errorf("redaction projection: build provenance is unavailable; run go run with the pinned module")
	}
	for _, dep := range info.Deps {
		if dep.Path == p.Module {
			if dep.Replace != nil || dep.Version != "v0.1.6" {
				return p, fmt.Errorf("redaction projection: engine pin changed or was replaced; restore the released v0.1.6 module before generation")
			}
			p.Version, p.Sum = dep.Version, dep.Sum
		}
	}
	if p.Sum == "" {
		return p, fmt.Errorf("redaction projection: engine checksum is missing; download the pinned released module before generation")
	}
	categoryNames := map[redact.Category]string{
		redact.CategorySecrets: "CategorySecrets", redact.CategoryPII: "CategoryPII",
		redact.CategoryPaths: "CategoryPaths", redact.CategoryProject: "CategoryProject",
	}
	labelNames := map[redact.CategoryString]string{
		redact.CategoryStringCredential: "CategoryStringCredential", redact.CategoryStringPII: "CategoryStringPII",
		redact.CategoryStringPath: "CategoryStringPath", redact.CategoryStringInternal: "CategoryStringInternal",
	}
	seen := map[redact.Category]bool{}
	labels := map[redact.CategoryString]bool{}
	for _, category := range redact.AllCategories() {
		name, known := categoryNames[category]
		if !known || seen[category] {
			return p, fmt.Errorf("redaction projection: engine category inventory changed; review the shared contract before generation")
		}
		if err := category.Validate(); err != nil {
			return p, err
		}
		label := category.String()
		labelName, known := labelNames[label]
		if !known || label == "" || labels[label] {
			return p, fmt.Errorf("redaction projection: engine labels are empty, repeated or unknown; review the engine mapping before generation")
		}
		seen[category], labels[label] = true, true
		p.Categories = append(p.Categories, member{strings.TrimPrefix(name, "Category"), string(category)})
		p.Labels = append(p.Labels, member{strings.TrimPrefix(labelName, "CategoryString"), string(label)})
		p.Pairs = append(p.Pairs, [2]string{string(category), string(label)})
	}
	for category := range categoryNames {
		if !seen[category] {
			return p, fmt.Errorf("redaction projection: engine category is missing; review the shared contract before generation")
		}
	}
	return p, nil
}

func main() {
	p, err := project()
	if err == nil {
		err = json.NewEncoder(os.Stdout).Encode(p)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
