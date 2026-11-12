package main

import (
	"fmt"
	"regexp"
	"strings"
)

// PRKind is the closed set of pull-request classifications.
type PRKind string

const (
	PRKindOrdinary   PRKind = "ordinary"
	PRKindRelease    PRKind = "release"
	PRKindVendorHash PRKind = "vendor-hash"
	PRKindCutRelease PRKind = "cut-release"
)

// AllPRKinds is the canonical closed set.
var AllPRKinds = []PRKind{PRKindOrdinary, PRKindRelease, PRKindVendorHash, PRKindCutRelease}

func (k PRKind) String() string { return string(k) }

// ClassifyPR maps an author and head branch to a PR kind using the release
// config's bot branch map. The protected-path exemption needs the automation
// author AND the release head branch; every other bot branch is an ordinary
// PR judged by its title.
func ClassifyPR(cfg Config, author, headBranch string) PRKind {
	if author != cfg.AutomationAuthor || author == "" {
		return PRKindOrdinary
	}
	for kind, branch := range cfg.BotBranches {
		if branch != "" && branch == headBranch {
			switch PRKind(kind) {
			case PRKindRelease:
				return PRKindRelease
			case PRKindVendorHash:
				return PRKindVendorHash
			case PRKindCutRelease:
				return PRKindCutRelease
			}
		}
	}
	// Compatibility with the policy file's release_job.branch: when the config
	// carries no bot_branches map, only the automation author matters and the
	// branch check happens in the workflow. The CLI still reports ordinary
	// here so workflows never compare branch literals themselves.
	return PRKindOrdinary
}

// titleType is the closed set of title types.
type titleType string

const (
	typeFeat     titleType = "feat"
	typeFix      titleType = "fix"
	typeRelease  titleType = "release"
	typeRefactor titleType = "refactor"
	typePerf     titleType = "perf"
	typeBuild    titleType = "build"
	typeRevert   titleType = "revert"
	typeDocs     titleType = "docs"
	typeTest     titleType = "test"
	typeCI       titleType = "ci"
	typeChore    titleType = "chore"
	typeStyle    titleType = "style"
)

// AllTypes is the canonical closed set.
var AllTypes = []titleType{
	typeFeat, typeFix, typeRelease,
	typeRefactor, typePerf, typeBuild, typeRevert,
	typeDocs, typeTest, typeCI, typeChore, typeStyle,
}

var titlePattern = regexp.MustCompile(`^([a-z]+)(\([^)]*\))?(!)?(\[(major|minor|patch)\])?: (\S.*)$`)
var oldReleasePattern = regexp.MustCompile(`^release\(v[^)]*\):\s*\S`)

// parseTitleKind classifies a raw title for queue-gate hold decisions without
// importing the core intent package. It returns releasing=true when the title
// carries a bump bracket on a releasing type.
func parseTitleKind(raw string) (releasing bool, parsable bool) {
	if oldReleasePattern.MatchString(raw) {
		return false, false
	}
	if strings.HasPrefix(raw, "Revert \"") {
		return false, false
	}
	m := titlePattern.FindStringSubmatch(raw)
	if m == nil {
		return false, false
	}
	typ := titleType(m[1])
	bump := m[5]
	hasBang := m[3] == "!"
	switch typ {
	case typeFeat, typeFix, typeRelease:
		if bump == "" {
			return false, false
		}
		return true, true
	case typeRefactor, typePerf, typeBuild, typeRevert:
		return bump != "", true
	case typeDocs, typeTest, typeCI, typeChore, typeStyle:
		if bump != "" || hasBang {
			return false, false
		}
		return false, true
	default:
		return false, false
	}
}

// runClassify implements: release-intent classify --author A --head-branch B.
func runClassify(args []string) error {
	var author, headBranch, configPath string
	configPath = DefaultConfigPath
	var haveAuthor, haveBranch bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--author":
			val, next, err := requireFlagValue(args, i, "--author")
			if err != nil {
				return err
			}
			author = val
			haveAuthor = true
			i = next
		case "--head-branch":
			val, next, err := requireFlagValue(args, i, "--head-branch")
			if err != nil {
				return err
			}
			headBranch = val
			haveBranch = true
			i = next
		case "--config":
			val, next, err := requireFlagValue(args, i, "--config")
			if err != nil {
				return err
			}
			configPath = val
			i = next
		default:
			return usageErr("unknown flag %q for classify: expected --author LOGIN --head-branch BRANCH [--config F]", args[i])
		}
	}
	if !haveAuthor {
		return usageErr("usage: release-intent classify --author LOGIN --head-branch BRANCH [--config F] (missing required --author): pass the PR author login")
	}
	if !haveBranch {
		return usageErr("usage: release-intent classify --author LOGIN --head-branch BRANCH [--config F] (missing required --head-branch): pass the PR head branch name")
	}
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	kind := ClassifyPR(cfg, author, headBranch)
	if err := emitLines([]string{"kind=" + kind.String()}); err != nil {
		return err
	}
	return nil
}

var _ = fmt.Sprint
