package main

import (
	"testing"

	"github.com/peasant-labs/redact"
)

func TestProjectionUsesReleasedEngine(t *testing.T) {
	p, err := project()
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[redact.Category]bool)
	for _, pair := range p.Pairs {
		category := redact.Category(pair[0])
		if err := category.Validate(); err != nil {
			t.Fatal(err)
		}
		if string(category.String()) != pair[1] || seen[category] {
			t.Fatalf("projection pair differs from released engine or repeats a category: %v", pair)
		}
		seen[category] = true
	}
	for _, category := range redact.AllCategories() {
		if !seen[category] {
			t.Fatalf("released category missing from projection: %s", category)
		}
	}
}
