package main

import (
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
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

func TestClassifyPR(t *testing.T) {
	t.Parallel()
	f := loadClassifyFixtures(t)
	for _, c := range f.Cases.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			cfg := intent.Config{AutomationAuthor: "peasant-labs-releaser[bot]", BotBranches: c.Input.BotBranches}
			if got := intent.ClassifyPR(cfg, c.Input.Author, c.Input.HeadBranch); string(got) != c.Expected.Kind {
				t.Fatalf("ClassifyPR(author=%q branch=%q) = %q, want %q", c.Input.Author, c.Input.HeadBranch, got, c.Expected.Kind)
			}
		})
	}
}

func TestParseFullTitle(t *testing.T) {
	t.Parallel()
	f := loadTitleFixtures(t)
	for _, c := range f.Cases.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			cfg := intent.Config{Releases: c.Input.Releases, TagPrefix: "v"}
			current := intent.SemVer("0.1.0")
			if !c.Input.BelowOne {
				current = intent.SemVer("1.0.0")
			}
			_, perr := intent.ParseTitle(cfg, current, c.Input.Title)
			got := ""
			if perr != nil {
				if ie, ok := perr.(*intent.Error); ok {
					got = string(ie.Code)
				} else {
					t.Fatalf("ParseTitle(%q) error type = %T, want *intent.Error", c.Input.Title, perr)
				}
			}
			if got != c.Expected.Code {
				t.Fatalf("ParseTitle(%q) code = %q, want %q", c.Input.Title, got, c.Expected.Code)
			}
		})
	}
}

func TestQueueGateFixtures(t *testing.T) {
	t.Parallel()
	f := loadQueueGateFixtures(t)
	for _, c := range f.Cases.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			data, err := queueJSONFS.ReadFile("testdata/queue/" + c.Input.File)
			if err != nil {
				t.Fatalf("read queue file %q: %v", c.Input.File, err)
			}
			nodes, total, hasMore, derr := decodeQueueFile(data)
			if c.Expected.Exit == 1 {
				if derr == nil {
					t.Fatalf("decodeQueueFile(%q) = nil error, want a failed-read error", c.Input.File)
				}
				return
			}
			if derr != nil {
				t.Fatalf("decodeQueueFile(%q): %v", c.Input.File, derr)
			}
			if got := gateDecision(nodes, total, hasMore, c.Input.ReleasePR); got != c.Expected.Gate {
				t.Fatalf("gateDecision(%q) = %q, want %q", c.Input.File, got, c.Expected.Gate)
			}
		})
	}
}
