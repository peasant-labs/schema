package schema

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type QualityFixtureName string

const (
	QualityFixtureResolvedTypical    QualityFixtureName = "resolved_typical"
	QualityFixtureResolvedHighTokens QualityFixtureName = "resolved_high_tokens"
	QualityFixturePartialMedium      QualityFixtureName = "partial_medium"
	QualityFixtureFailedComplex      QualityFixtureName = "failed_complex"
	QualityFixtureResolvedMinimal    QualityFixtureName = "resolved_minimal"
)

var allQualityFixtureNames = []QualityFixtureName{
	QualityFixtureResolvedTypical,
	QualityFixtureResolvedHighTokens,
	QualityFixturePartialMedium,
	QualityFixtureFailedComplex,
	QualityFixtureResolvedMinimal,
}

type QualityFixtureSetName string

const (
	QualityFixtureSetProjectMix QualityFixtureSetName = "project_mix"
)

var allQualityFixtureSetNames = []QualityFixtureSetName{
	QualityFixtureSetProjectMix,
}

// QualityFixtures is the parsed testdata/quality/sessions.yaml corpus.
type QualityFixtures struct {
	Sessions   []QualitySessionFixture `yaml:"quality_sessions"`
	Sets       []QualityFixtureSet     `yaml:"quality_fixture_sets"`
	Variations QualityVariations       `yaml:"quality_variations"`
}

// QualityVariations is the reusable combinatorial input catalog carried by the
// canonical quality fixture document.
type QualityVariations struct {
	Outcomes    []QualityStringVariation `json:"outcomes" yaml:"outcomes"`
	Projects    []QualityStringVariation `json:"projects" yaml:"projects"`
	Scopes      []QualityStringVariation `json:"scopes" yaml:"scopes"`
	TaskTitles  []QualityStringVariation `json:"taskTitles" yaml:"task_titles"`
	TokenRatios []QualityRatioVariation  `json:"tokenRatios" yaml:"token_ratios"`
	Metrics     QualityMetricVariations  `json:"metrics" yaml:"metrics"`
}

type QualityStringVariation struct {
	Value string `json:"value" yaml:"value"`
}

type QualityRatioVariation struct {
	Name       string  `json:"name" yaml:"name"`
	InputRatio float64 `json:"inputRatio" yaml:"inputRatio"`
}

type QualityMetricVariation struct {
	Name  string  `json:"name" yaml:"name"`
	Value float64 `json:"value" yaml:"value"`
}

type QualityMetricVariations struct {
	RetryLoops       []QualityMetricVariation `json:"retryLoops" yaml:"retry_loops"`
	SignalDensity    []QualityMetricVariation `json:"signalDensity" yaml:"signal_density"`
	SpecQualityScore []QualityMetricVariation `json:"specQualityScore" yaml:"spec_quality_score"`
	FilesTouched     []QualityMetricVariation `json:"filesTouched" yaml:"files_touched"`
	LinesChanged     []QualityMetricVariation `json:"linesChanged" yaml:"lines_changed"`
}

// QualitySessionFixture is one named quality-session fixture row.
type QualitySessionFixture struct {
	Name                 QualityFixtureName `json:"name" yaml:"name"`
	ID                   string             `json:"id" yaml:"id"`
	Date                 string             `json:"date" yaml:"date"`
	Project              string             `json:"project" yaml:"project"`
	Scope                string             `json:"scope" yaml:"scope"`
	Title                string             `json:"title" yaml:"title"`
	TotalTokens          int                `json:"totalTokens" yaml:"totalTokens"`
	InputTokens          int                `json:"inputTokens" yaml:"inputTokens"`
	OutputTokens         int                `json:"outputTokens" yaml:"outputTokens"`
	TurnCount            int                `json:"turnCount" yaml:"turnCount"`
	ToolCalls            int                `json:"toolCalls" yaml:"toolCalls"`
	Outcome              string             `json:"outcome" yaml:"outcome"`
	FilesTouched         int                `json:"filesTouched" yaml:"filesTouched"`
	LinesChanged         int                `json:"linesChanged" yaml:"linesChanged"`
	DurationMinutes      float64            `json:"durationMinutes" yaml:"durationMinutes"`
	RetryLoops           int                `json:"retryLoops" yaml:"retryLoops"`
	RetryTokensWasted    int                `json:"retryTokensWasted" yaml:"retryTokensWasted"`
	WithinSessionReverts int                `json:"withinSessionReverts" yaml:"withinSessionReverts"`
	SignalDensity        float64            `json:"signalDensity" yaml:"signalDensity"`
	SpecQualityScore     float64            `json:"specQualityScore" yaml:"specQualityScore"`
	ExplorationRatio     float64            `json:"explorationRatio" yaml:"explorationRatio"`
	ScopeBreadth         int                `json:"scopeBreadth" yaml:"scopeBreadth"`
	DiscoveryTurns       int                `json:"discoveryTurns" yaml:"discoveryTurns"`
}

// QualityFixtureSet is a named reusable list of quality-session fixture rows.
type QualityFixtureSet struct {
	Name  QualityFixtureSetName `json:"name" yaml:"name"`
	Cases []QualityFixtureName  `json:"cases" yaml:"cases"`
}

// LoadQualityFixtures parses QualitySessionsYAML into structured fixtures.
func LoadQualityFixtures() (*QualityFixtures, error) {
	var f QualityFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(QualitySessionsYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&f); err != nil {
		return nil, fmt.Errorf("load quality fixtures: decode document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("load quality fixtures: decode trailing document: %w", err)
		}
		return nil, fmt.Errorf("load quality fixtures: multiple YAML documents are not allowed")
	}
	if err := f.validate(); err != nil {
		return nil, err
	}
	return &f, nil
}

// SessionByName returns the named quality-session fixture row.
func (f *QualityFixtures) SessionByName(name QualityFixtureName) (QualitySessionFixture, bool) {
	for _, s := range f.Sessions {
		if s.Name == name {
			return s, true
		}
	}
	return QualitySessionFixture{}, false
}

// SetByName returns the named quality fixture set.
func (f *QualityFixtures) SetByName(name QualityFixtureSetName) (QualityFixtureSet, bool) {
	for _, s := range f.Sets {
		if s.Name == name {
			return s, true
		}
	}
	return QualityFixtureSet{}, false
}

// QualitySessions returns all quality-session fixtures as wire payload rows.
func (f *QualityFixtures) QualitySessions() []QualitySession {
	out := make([]QualitySession, len(f.Sessions))
	for i, s := range f.Sessions {
		out[i] = s.ToQualitySession()
	}
	return out
}

// QualitySessionsForSet returns a named fixture set as wire payload rows.
func (f *QualityFixtures) QualitySessionsForSet(name QualityFixtureSetName) ([]QualitySession, error) {
	set, ok := f.SetByName(name)
	if !ok {
		return nil, fmt.Errorf("unknown quality fixture set %q", name)
	}

	out := make([]QualitySession, 0, len(set.Cases))
	for _, caseName := range set.Cases {
		session, ok := f.SessionByName(caseName)
		if !ok {
			return nil, fmt.Errorf("quality fixture set %q references unknown case %q", name, caseName)
		}
		out = append(out, session.ToQualitySession())
	}
	return out, nil
}

// ToQualitySession converts a named fixture row to the wire payload shape.
func (f QualitySessionFixture) ToQualitySession() QualitySession {
	return QualitySession{
		ID:                   f.ID,
		Date:                 f.Date,
		Project:              f.Project,
		Scope:                f.Scope,
		Title:                f.Title,
		TotalTokens:          f.TotalTokens,
		InputTokens:          f.InputTokens,
		OutputTokens:         f.OutputTokens,
		TurnCount:            f.TurnCount,
		ToolCalls:            f.ToolCalls,
		Outcome:              f.Outcome,
		FilesTouched:         f.FilesTouched,
		LinesChanged:         f.LinesChanged,
		DurationMinutes:      f.DurationMinutes,
		RetryLoops:           f.RetryLoops,
		RetryTokensWasted:    f.RetryTokensWasted,
		WithinSessionReverts: f.WithinSessionReverts,
		SignalDensity:        f.SignalDensity,
		SpecQualityScore:     f.SpecQualityScore,
		ExplorationRatio:     f.ExplorationRatio,
		ScopeBreadth:         f.ScopeBreadth,
		DiscoveryTurns:       f.DiscoveryTurns,
	}
}

func (f *QualityFixtures) validate() error {
	sessionNames := make(map[QualityFixtureName]struct{}, len(f.Sessions))
	for _, session := range f.Sessions {
		if _, exists := sessionNames[session.Name]; exists {
			return fmt.Errorf("quality fixture corpus has duplicate case %q", session.Name)
		}
		sessionNames[session.Name] = struct{}{}
	}
	knownSessionNames := make(map[QualityFixtureName]struct{}, len(allQualityFixtureNames))
	for _, name := range allQualityFixtureNames {
		knownSessionNames[name] = struct{}{}
		if _, ok := sessionNames[name]; !ok {
			return fmt.Errorf("quality fixture corpus missing case %q", name)
		}
	}
	for _, session := range f.Sessions {
		if _, ok := knownSessionNames[session.Name]; !ok {
			return fmt.Errorf("quality fixture corpus has unregistered case %q", session.Name)
		}
	}

	setNames := make(map[QualityFixtureSetName]struct{}, len(f.Sets))
	knownSetNames := make(map[QualityFixtureSetName]struct{}, len(allQualityFixtureSetNames))
	for _, name := range allQualityFixtureSetNames {
		knownSetNames[name] = struct{}{}
	}
	for _, set := range f.Sets {
		if _, exists := setNames[set.Name]; exists {
			return fmt.Errorf("quality fixture corpus has duplicate set %q", set.Name)
		}
		if _, ok := knownSetNames[set.Name]; !ok {
			return fmt.Errorf("quality fixture corpus has unregistered set %q", set.Name)
		}
		setNames[set.Name] = struct{}{}
		for _, caseName := range set.Cases {
			if _, ok := sessionNames[caseName]; !ok {
				return fmt.Errorf("quality fixture set %q references unknown case %q", set.Name, caseName)
			}
		}
	}
	for _, name := range allQualityFixtureSetNames {
		if _, ok := setNames[name]; !ok {
			return fmt.Errorf("quality fixture corpus missing set %q", name)
		}
	}
	return nil
}
