package intent

import "fmt"

// This file defines every closed set the package trades in. Closed sets are
// typed, with no bare strings crossing the package boundary: callers compare
// against named constants, and membership is asserted exactly in fixtures.

// Type is a PR-title type: the verb before the optional scope list.
type Type string

const (
	TypeFeat     Type = "feat"
	TypeFix      Type = "fix"
	TypeRelease  Type = "release"
	TypeRefactor Type = "refactor"
	TypePerf     Type = "perf"
	TypeBuild    Type = "build"
	TypeRevert   Type = "revert"
	TypeDocs     Type = "docs"
	TypeTest     Type = "test"
	TypeCI       Type = "ci"
	TypeChore    Type = "chore"
	TypeStyle    Type = "style"
)

// AllTypes is the canonical closed set of title types, in grammar order.
var AllTypes = []Type{
	TypeFeat, TypeFix, TypeRelease,
	TypeRefactor, TypePerf, TypeBuild, TypeRevert,
	TypeDocs, TypeTest, TypeCI, TypeChore, TypeStyle,
}

// IsValid reports whether the type is one of the known members.
func (t Type) IsValid() bool {
	for _, known := range AllTypes {
		if t == known {
			return true
		}
	}
	return false
}

func (t Type) String() string { return string(t) }

// Class groups types by what a bump bracket means on them.
type Class int

const (
	// ClassMustRelease types always release and always carry a bracket.
	ClassMustRelease Class = iota + 1
	// ClassMayRelease types release only with a bracket.
	ClassMayRelease
	// ClassNeverRelease types never carry a bracket or a breaking mark.
	ClassNeverRelease
)

// AllClasses is the canonical closed set of type classes.
var AllClasses = []Class{ClassMustRelease, ClassMayRelease, ClassNeverRelease}

// Class reports the release class of the type.
func (t Type) Class() Class {
	switch t {
	case TypeFeat, TypeFix, TypeRelease:
		return ClassMustRelease
	case TypeRefactor, TypePerf, TypeBuild, TypeRevert:
		return ClassMayRelease
	default:
		return ClassNeverRelease
	}
}

// Level is a release bump level. The zero value means no release.
type Level int

const (
	// LevelNone marks a title that releases nothing.
	LevelNone Level = iota
	// LevelPatch advances the last number.
	LevelPatch
	// LevelMinor advances the middle number.
	LevelMinor
	// LevelMajor advances the first number.
	LevelMajor
)

// AllLevels is the canonical closed set of bump levels, in increasing order.
var AllLevels = []Level{LevelNone, LevelPatch, LevelMinor, LevelMajor}

func (l Level) String() string {
	switch l {
	case LevelPatch:
		return "patch"
	case LevelMinor:
		return "minor"
	case LevelMajor:
		return "major"
	default:
		return "none"
	}
}

// parseLevel maps a level word to its level. It is the one map both the
// title grammar and the fixture loader use.
func parseLevel(word string) (Level, bool) {
	switch word {
	case "patch":
		return LevelPatch, true
	case "minor":
		return LevelMinor, true
	case "major":
		return LevelMajor, true
	case "none", "":
		return LevelNone, true
	default:
		return LevelNone, false
	}
}

// UnmarshalYAML accepts the level names the fixtures use.
func (l *Level) UnmarshalYAML(unmarshal func(any) error) error {
	var word string
	if err := unmarshal(&word); err != nil {
		return err
	}
	level, ok := parseLevel(word)
	if !ok {
		return fmt.Errorf("unknown level %q: want one of none, patch, minor, major", word)
	}
	*l = level
	return nil
}

// Mode selects whether the release job may write. Anything but "tag" plans.
type Mode string

const (
	// ModePlanOnly prints the plan and writes nothing.
	ModePlanOnly Mode = "plan-only"
	// ModeTag arms the job to push a tag and keep the release PR current.
	ModeTag Mode = "tag"
)

// AllModes is the canonical closed set of release-job modes.
var AllModes = []Mode{ModePlanOnly, ModeTag}

// ParseMode maps "tag" to ModeTag and anything else to ModePlanOnly.
func ParseMode(raw string) Mode {
	if raw == string(ModeTag) {
		return ModeTag
	}
	return ModePlanOnly
}

// Action is the one thing a plan tells the release workflow to do.
type Action string

const (
	// ActionNone writes nothing; printed whenever tag mode is off.
	ActionNone Action = "none"
	// ActionVerify runs the pre-tag steps on the candidate.
	ActionVerify Action = "verify"
	// ActionTag pushes the printed tag at the candidate.
	ActionTag Action = "tag"
	// ActionOpenPR writes the release content and opens or updates the release PR.
	ActionOpenPR Action = "open-pr"
	// ActionClosePR closes the open release PR, if there is one.
	ActionClosePR Action = "close-pr"
)

// AllActions is the canonical closed set of plan actions.
var AllActions = []Action{ActionNone, ActionVerify, ActionTag, ActionOpenPR, ActionClosePR}

func (a Action) String() string { return string(a) }

// Verdict is the workflow's report on its pre-tag steps, and nothing else.
type Verdict string

const (
	// VerdictNone means no pre-tag report was given.
	VerdictNone Verdict = ""
	// VerdictTaggable means every pre-tag step passed.
	VerdictTaggable Verdict = "taggable"
	// VerdictRejected means a pre-tag step failed.
	VerdictRejected Verdict = "rejected"
)

// AllVerdicts is the canonical closed set of pre-tag verdicts.
var AllVerdicts = []Verdict{VerdictNone, VerdictTaggable, VerdictRejected}

func (v Verdict) String() string { return string(v) }

// Gate is the merge-queue hold decision for the release PR.
type Gate string

const (
	// GateArm lets the job turn auto-merge on for the release PR.
	GateArm Gate = "arm"
	// GateHold keeps the release PR unarmed while a releasing PR is ahead.
	GateHold Gate = "hold"
)

// AllGates is the canonical closed set of queue-gate outcomes.
var AllGates = []Gate{GateArm, GateHold}

func (g Gate) String() string { return string(g) }

// Freshness is whether a landed release candidate still matches what the
// generator produces now. Unknown means there was nothing to compare, never
// that a read failed: a read failure is an error, not a stale verdict.
type Freshness string

const (
	// FreshnessUnknown marks the absence of a candidate.
	FreshnessUnknown Freshness = "unknown"
	// FreshnessFresh marks content equal to what the generator produces now.
	FreshnessFresh Freshness = "fresh"
	// FreshnessStale marks content that no longer matches the generator.
	FreshnessStale Freshness = "stale"
)

// AllFreshness is the canonical closed set of freshness outcomes.
var AllFreshness = []Freshness{FreshnessUnknown, FreshnessFresh, FreshnessStale}

func (f Freshness) String() string { return string(f) }

// PRKind tells the workflows which bot PR a head branch belongs to, if any.
type PRKind string

const (
	// PRKindOrdinary is any author PR or any bot PR on an unlisted branch.
	PRKindOrdinary PRKind = "ordinary"
	// PRKindRelease is the one standing release PR.
	PRKindRelease PRKind = "release"
	// PRKindVendorHash is the dependency-hash fix PR.
	PRKindVendorHash PRKind = "vendor-hash"
	// PRKindCutRelease is the maintainer-triggered forced-release PR.
	PRKindCutRelease PRKind = "cut-release"
)

// AllPRKinds is the canonical closed set of PR kinds.
var AllPRKinds = []PRKind{PRKindOrdinary, PRKindRelease, PRKindVendorHash, PRKindCutRelease}

func (k PRKind) String() string { return string(k) }

// botKinds lists the bot PR kinds in canonical order, derived once from
// AllPRKinds so the closed set has one home.
var botKinds = func() []PRKind {
	out := []PRKind{}
	for _, k := range AllPRKinds {
		if k != PRKindOrdinary {
			out = append(out, k)
		}
	}
	return out
}()

// ErrorCode is the closed set of typed rejections. Every member has a reject
// fixture whose message carries all six parts; codes used only by later
// work join this set with that work.
type ErrorCode string

const (
	ErrMalformed           ErrorCode = "malformed"
	ErrUnknownType         ErrorCode = "unknown-type"
	ErrMissingBump         ErrorCode = "missing-bump"
	ErrBumpOnNonReleasing  ErrorCode = "bump-on-non-releasing"
	ErrBangBumpMismatch    ErrorCode = "bang-bump-mismatch"
	ErrOldReleaseTitle     ErrorCode = "old-release-title"
	ErrReleasesDisabled    ErrorCode = "releases-disabled"
	ErrProtectedPath       ErrorCode = "protected-path"
	ErrNoFinalTag          ErrorCode = "no-final-tag"
	ErrEpochMissing        ErrorCode = "epoch-missing"
	ErrFloorNotOnChain     ErrorCode = "floor-not-on-chain"
	ErrReleaseContentStale ErrorCode = "release-content-stale"
	ErrConfigInvalid       ErrorCode = "config-invalid"
	ErrPolicyViolation     ErrorCode = "policy-violation"
)

// AllErrorCodes is the canonical closed set of typed error codes.
var AllErrorCodes = []ErrorCode{
	ErrMalformed,
	ErrUnknownType,
	ErrMissingBump,
	ErrBumpOnNonReleasing,
	ErrBangBumpMismatch,
	ErrOldReleaseTitle,
	ErrReleasesDisabled,
	ErrProtectedPath,
	ErrNoFinalTag,
	ErrEpochMissing,
	ErrFloorNotOnChain,
	ErrReleaseContentStale,
	ErrConfigInvalid,
	ErrPolicyViolation,
}

func (c ErrorCode) String() string { return string(c) }

// SemVer is a bare X.Y.Z version value, without any tag prefix. A tag is
// always TagPrefix plus a SemVer.
type SemVer string

func (s SemVer) String() string { return string(s) }

// Title is one parsed intent title: a PR title, or a landed subject with its
// PR number attached. PR is 0 when the text is a PR title, not a landed
// subject.
type Title struct {
	Type     Type     `yaml:"type"`
	Scopes   []string `yaml:"scopes"`
	Breaking bool     `yaml:"breaking"`
	Level    Level    `yaml:"level"`
	Subject  string   `yaml:"subject"`
	PR       int      `yaml:"pr"`
}

// VersionFile names one version file and the single top-level field the
// release PR writes and protects. The rest of the file stays ordinary.
type VersionFile struct {
	Path  string `yaml:"path"`
	Field string `yaml:"field"`
}

// Config is one repo's release configuration, read from its release-intent
// file. BotBranches maps a bot PR kind to the one head branch that PR uses,
// so every workflow compares against the same names.
type Config struct {
	Releases         bool              `yaml:"releases"`
	TagPrefix        string            `yaml:"tag_prefix"`
	AutomationAuthor string            `yaml:"automation_author"`
	Changelog        string            `yaml:"changelog"`
	VersionFiles     []VersionFile     `yaml:"version_files"`
	ProtectedPaths   []string          `yaml:"protected_paths"`
	Surfaces         map[string]string `yaml:"surfaces"`
	BotBranches      map[PRKind]string `yaml:"bot_branches"`
}

// Error is an actionable rejection: what went wrong, why, where it failed,
// when it failed, what it means for the caller, and how to fix it.
type Error struct {
	Code        ErrorCode
	What, Why   string
	Where, When string
	Means, Fix  string
}

func (e *Error) Error() string {
	return "release-intent: " + e.Code.String() + ": " + e.What +
		" [why: " + e.Why +
		"; where: " + e.Where +
		"; when: " + e.When +
		"; means: " + e.Means +
		"; fix: " + e.Fix + "]"
}

// Commit is one first-parent commit on the release branch: its SHA, its
// subject, whether it touches the changelog file, and its committer date as
// YYYY-MM-DD for a dated changelog heading.
type Commit struct {
	SHA              string `yaml:"sha"`
	Subject          string `yaml:"subject"`
	ChangesChangelog bool   `yaml:"changes_changelog"`
	Date             string `yaml:"date"`
}

// Facts is everything Plan reads: the branch under plan, its tip's
// first-parent history, every tag in the repository, the grammar epoch and
// the recovery floor, the job mode and verdict, and per-commit freshness for
// the changelog-changing commits in range. The command layer fills Fresh by
// comparing each candidate through Generate; Plan only reads the entry for
// the commit it picks.
type Facts struct {
	Branch  string            `yaml:"branch"`
	Commits []Commit          `yaml:"commits"`
	Tags    map[string]string `yaml:"tags"`
	Epoch   string            `yaml:"epoch"`
	Floor   string            `yaml:"floor"`
	Mode    Mode              `yaml:"mode"`
	Verdict Verdict           `yaml:"verdict"`
	Fresh   map[string]bool   `yaml:"fresh"`
}

// BeforeCandidate is the sub-range a candidate release commit was built
// from: the releasing subjects older than the candidate, the version and tag
// they compute, and the committer date of their newest commit for a dated
// heading. The freshness check renders exactly this set, so the caller never
// re-plans a sub-range.
type BeforeCandidate struct {
	Pending []Title `yaml:"pending"`
	Version SemVer  `yaml:"version"`
	Tag     string  `yaml:"tag"`
	Date    string  `yaml:"date"`
}

// Decision is the whole plan: the one action plus every value the workflow
// needs to carry it out. For verify and tag, Pending, Level, Breaking,
// Version and Tag come from the subjects older than the candidate, and Before
// repeats that same set with its date; the subjects after the candidate stay
// pending for the next release. Full holds the whole range's set in the same
// shape, always filled when anything is pending, whatever the verdict; the
// release-pr write path renders Full. Date is the committer date of the
// newest commit in Pending, for the dated heading of the set being rendered.
type Decision struct {
	Action         Action          `yaml:"action"`
	RangeBase      string          `yaml:"range_base"`
	VersionBase    string          `yaml:"version_base"`
	Candidate      string          `yaml:"candidate"`
	CandidateFresh Freshness       `yaml:"candidate_fresh"`
	Pending        []Title         `yaml:"pending"`
	Level          Level           `yaml:"level"`
	Breaking       bool            `yaml:"breaking"`
	Version        SemVer          `yaml:"version"`
	Tag            string          `yaml:"tag"`
	Before         BeforeCandidate `yaml:"before"`
	Full           BeforeCandidate `yaml:"full"`
	Date           string          `yaml:"date"`
}

// ContentInput is what the release-content generator renders.
type ContentInput struct {
	Pending       []Title `yaml:"pending"`
	BaseChangelog string  `yaml:"base_changelog"`
	Version       SemVer  `yaml:"version"`
	Tag           string  `yaml:"tag"`
	Date          string  `yaml:"date"`
}

// Content is the generated release content: the full new changelog text, the
// new section on its own, and one value per version file path.
type Content struct {
	Changelog string            `yaml:"changelog"`
	Section   string            `yaml:"section"`
	Versions  map[string]string `yaml:"versions"`
}

// QueueEntry is one merge-queue entry the gate judges.
type QueueEntry struct {
	Position int    `yaml:"position"`
	Number   int    `yaml:"number"`
	Title    string `yaml:"title"`
}
