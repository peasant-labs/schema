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

func TestClassifyPR(t *testing.T) {
	t.Parallel()
	f := loadClassifyFixtures(t)
	for _, c := range f.Cases.Cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			cfg := Config{AutomationAuthor: "peasant-labs-releaser[bot]", BotBranches: c.Input.BotBranches}
			// The empty-author case still uses the bot author in config; the
			// input author is what classifies.
			if got := ClassifyPR(cfg, c.Input.Author, c.Input.HeadBranch); string(got) != c.Expected.Kind {
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
			_, rej := parseFullTitle(c.Input.Title, c.Input.Releases, c.Input.BelowOne)
			got := ""
			if rej != nil {
				got = string(rej.Code)
			}
			if got != c.Expected.Code {
				t.Fatalf("parseFullTitle(%q) code = %q, want %q", c.Input.Title, got, c.Expected.Code)
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
