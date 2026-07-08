package schema

import (
	"fmt"

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
	Sessions []QualitySessionFixture `yaml:"quality_sessions"`
	Sets     []QualityFixtureSet     `yaml:"quality_fixture_sets"`
}

// QualitySessionFixture is one named quality-session fixture row.
type QualitySessionFixture struct {
	Name                 QualityFixtureName `yaml:"name"`
	ID                   string             `yaml:"id"`
	Date                 string             `yaml:"date"`
	Project              string             `yaml:"project"`
	Scope                string             `yaml:"scope"`
	Title                string             `yaml:"title"`
	TotalTokens          int                `yaml:"totalTokens"`
	InputTokens          int                `yaml:"inputTokens"`
	OutputTokens         int                `yaml:"outputTokens"`
	TurnCount            int                `yaml:"turnCount"`
	ToolCalls            int                `yaml:"toolCalls"`
	Outcome              string             `yaml:"outcome"`
	FilesTouched         int                `yaml:"filesTouched"`
	LinesChanged         int                `yaml:"linesChanged"`
	DurationMinutes      float64            `yaml:"durationMinutes"`
	RetryLoops           int                `yaml:"retryLoops"`
	RetryTokensWasted    int                `yaml:"retryTokensWasted"`
	WithinSessionReverts int                `yaml:"withinSessionReverts"`
	SignalDensity        float64            `yaml:"signalDensity"`
	SpecQualityScore     float64            `yaml:"specQualityScore"`
	ExplorationRatio     float64            `yaml:"explorationRatio"`
	ScopeBreadth         int                `yaml:"scopeBreadth"`
	DiscoveryTurns       int                `yaml:"discoveryTurns"`
}

// QualityFixtureSet is a named reusable list of quality-session fixture rows.
type QualityFixtureSet struct {
	Name  QualityFixtureSetName `yaml:"name"`
	Cases []QualityFixtureName  `yaml:"cases"`
}

// LoadQualityFixtures parses QualitySessionsYAML into structured fixtures.
func LoadQualityFixtures() (*QualityFixtures, error) {
	var f QualityFixtures
	if err := yaml.Unmarshal(QualitySessionsYAML, &f); err != nil {
		return nil, fmt.Errorf("load quality fixtures: %w", err)
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
