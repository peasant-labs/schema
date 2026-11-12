package main

import (
	"testing"
)

func TestParseGroupRef(t *testing.T) {
	t.Parallel()
	f := loadGroupPRFixtures(t)
	for _, c := range f.Cases.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			if got := parseGroupRef(c.Input.Ref); got != c.Expected.PR {
				t.Fatalf("parseGroupRef(%q) = %q, want %q", c.Input.Ref, got, c.Expected.PR)
			}
		})
	}
}
