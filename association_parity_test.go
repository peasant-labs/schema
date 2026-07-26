package schema_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/association_parity.yaml
var associationParityCasesYAML []byte

//go:embed testdata/local-api/association_parity_manifest.yaml
var associationParityManifestYAML []byte

type associationParityFixtureManifest struct {
	ExpectedCaseCount int
	RequiredCaseNames []string
}

type associationParityFixtureManifestYAML struct {
	ExpectedCaseCount associationParityManifestCount      `yaml:"expectedCaseCount"`
	RequiredCaseNames []associationParityManifestCaseName `yaml:"requiredCaseNames"`
}

type associationParityManifestCount int

func (count *associationParityManifestCount) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode || value.Tag != "!!int" {
		return fmt.Errorf("expectedCaseCount must be a YAML integer")
	}
	var decoded int
	if err := value.Decode(&decoded); err != nil {
		return fmt.Errorf("decode expectedCaseCount: %w", err)
	}
	*count = associationParityManifestCount(decoded)
	return nil
}

type associationParityManifestCaseName string

func (name *associationParityManifestCaseName) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
		return fmt.Errorf("required case name must be a YAML string")
	}
	*name = associationParityManifestCaseName(value.Value)
	return nil
}

func loadAssociationParityFixtures(t *testing.T) testcase.Corpus[schema.SessionAssociation, bool] {
	t.Helper()
	fixtures, _, err := loadAssociationParityFixturesFromYAML(associationParityCasesYAML, associationParityManifestYAML)
	if err != nil {
		t.Fatalf("load association parity fixtures and inventory: %v", err)
	}
	return fixtures
}

func loadAssociationParityFixturesFromYAML(corpusYAML, manifestYAML []byte) (testcase.Corpus[schema.SessionAssociation, bool], associationParityFixtureManifest, error) {
	fixtures, err := testcase.LoadCorpus[schema.SessionAssociation, bool](corpusYAML)
	if err != nil {
		return testcase.Corpus[schema.SessionAssociation, bool]{}, associationParityFixtureManifest{}, fmt.Errorf("decode association parity corpus: %w", err)
	}
	manifest, err := decodeAssociationParityFixtureManifest(manifestYAML)
	if err != nil {
		return testcase.Corpus[schema.SessionAssociation, bool]{}, associationParityFixtureManifest{}, err
	}
	if err := validateAssociationParityFixtureInventory(fixtures, manifest); err != nil {
		return testcase.Corpus[schema.SessionAssociation, bool]{}, associationParityFixtureManifest{}, err
	}
	return fixtures, manifest, nil
}

func decodeAssociationParityFixtureManifest(data []byte) (associationParityFixtureManifest, error) {
	var decoded associationParityFixtureManifestYAML
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&decoded); err != nil {
		return associationParityFixtureManifest{}, fmt.Errorf("decode association parity manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return associationParityFixtureManifest{}, fmt.Errorf("decode trailing association parity manifest document: %w", err)
		}
		return associationParityFixtureManifest{}, fmt.Errorf("decode association parity manifest: multiple YAML documents are not allowed")
	}
	manifest := associationParityFixtureManifest{
		ExpectedCaseCount: int(decoded.ExpectedCaseCount),
		RequiredCaseNames: make([]string, len(decoded.RequiredCaseNames)),
	}
	for index, name := range decoded.RequiredCaseNames {
		manifest.RequiredCaseNames[index] = string(name)
	}
	if manifest.ExpectedCaseCount <= 0 {
		return associationParityFixtureManifest{}, fmt.Errorf("association parity manifest expectedCaseCount must be positive, got %d", manifest.ExpectedCaseCount)
	}
	if len(manifest.RequiredCaseNames) != manifest.ExpectedCaseCount {
		return associationParityFixtureManifest{}, fmt.Errorf("association parity manifest has %d requiredCaseNames, want exactly %d", len(manifest.RequiredCaseNames), manifest.ExpectedCaseCount)
	}
	seen := make(map[string]struct{}, len(manifest.RequiredCaseNames))
	for index, name := range manifest.RequiredCaseNames {
		if strings.TrimSpace(name) == "" {
			return associationParityFixtureManifest{}, fmt.Errorf("association parity manifest requiredCaseNames[%d] is blank", index)
		}
		if _, exists := seen[name]; exists {
			return associationParityFixtureManifest{}, fmt.Errorf("association parity manifest repeats required case name %q", name)
		}
		seen[name] = struct{}{}
	}
	return manifest, nil
}

func validateAssociationParityFixtureInventory(fixtures testcase.Corpus[schema.SessionAssociation, bool], manifest associationParityFixtureManifest) error {
	if len(fixtures.Cases) != manifest.ExpectedCaseCount {
		return fmt.Errorf("association parity corpus has %d cases, want exactly %d from its manifest", len(fixtures.Cases), manifest.ExpectedCaseCount)
	}
	required := make(map[string]struct{}, len(manifest.RequiredCaseNames))
	for _, name := range manifest.RequiredCaseNames {
		required[name] = struct{}{}
	}
	actual := make(map[string]struct{}, len(fixtures.Cases))
	for index, fixture := range fixtures.Cases {
		if strings.TrimSpace(fixture.Name) == "" {
			return fmt.Errorf("association parity corpus case %d has a blank name", index)
		}
		if _, exists := actual[fixture.Name]; exists {
			return fmt.Errorf("association parity corpus repeats case name %q", fixture.Name)
		}
		actual[fixture.Name] = struct{}{}
		if _, exists := required[fixture.Name]; !exists {
			return fmt.Errorf("association parity corpus contains unregistered case %q", fixture.Name)
		}
	}
	for name := range required {
		if _, exists := actual[name]; !exists {
			return fmt.Errorf("association parity corpus is missing required case %q", name)
		}
	}
	return nil
}

func TestSessionAssociation_SharedParityCorpus(t *testing.T) {
	for _, fixture := range loadAssociationParityFixtures(t).Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			valid := fixture.Input.Validate() == nil
			if valid != fixture.Expected {
				t.Errorf("Validate() valid=%t, want %t", valid, fixture.Expected)
			}
			wantClassification := testcase.MustFail
			if fixture.Expected {
				wantClassification = testcase.MustPass
			}
			if fixture.Classification != wantClassification {
				t.Errorf("classification=%q, want %q for expected=%t", fixture.Classification, wantClassification, fixture.Expected)
			}
		})
	}
}

func TestAssociationParityFixtureManifestIsStrict(t *testing.T) {
	manifest, err := decodeAssociationParityFixtureManifest(associationParityManifestYAML)
	if err != nil {
		t.Fatalf("load association parity manifest: %v", err)
	}
	manifestSource := string(associationParityManifestYAML)
	firstLine := strings.SplitN(manifestSource, "\n", 2)[0]
	if firstLine == "" {
		t.Fatal("association parity manifest is missing its expectedCaseCount line")
	}
	if len(manifest.RequiredCaseNames) < 2 {
		t.Fatal("association parity manifest needs two names for duplicate-name mutation coverage")
	}

	t.Run("unknown field", func(t *testing.T) {
		requireAssociationParityFixtureLoadFailure(t, associationParityCasesYAML, []byte(manifestSource+"\nunknown: value\n"))
	})
	t.Run("duplicate YAML key", func(t *testing.T) {
		requireAssociationParityFixtureLoadFailure(t, associationParityCasesYAML, []byte(manifestSource+"\n"+firstLine+"\n"))
	})
	t.Run("second YAML document", func(t *testing.T) {
		requireAssociationParityFixtureLoadFailure(t, associationParityCasesYAML, []byte(manifestSource+"\n---\n"+firstLine+"\n"))
	})
	t.Run("duplicate required case name", func(t *testing.T) {
		requireAssociationParityFixtureLoadFailure(t, associationParityCasesYAML, []byte(replaceAssociationParityFixtureText(t, manifestSource, "  - "+manifest.RequiredCaseNames[1], "  - "+manifest.RequiredCaseNames[0])))
	})
	t.Run("blank required case name", func(t *testing.T) {
		requireAssociationParityFixtureLoadFailure(t, associationParityCasesYAML, []byte(replaceAssociationParityFixtureText(t, manifestSource, "  - "+manifest.RequiredCaseNames[0], "  - \"\"")))
	})
	t.Run("malformed required case name", func(t *testing.T) {
		requireAssociationParityFixtureLoadFailure(t, associationParityCasesYAML, []byte(fmt.Sprintf("expectedCaseCount: %d\nrequiredCaseNames: [1]\n", manifest.ExpectedCaseCount)))
	})
	t.Run("count and name mismatch", func(t *testing.T) {
		countLine := fmt.Sprintf("expectedCaseCount: %d", manifest.ExpectedCaseCount)
		requireAssociationParityFixtureLoadFailure(t, associationParityCasesYAML, []byte(replaceAssociationParityFixtureText(t, manifestSource, countLine, fmt.Sprintf("expectedCaseCount: %d", manifest.ExpectedCaseCount+1))))
	})
	t.Run("count-preserving unregistered corpus name", func(t *testing.T) {
		corpusSource := replaceAssociationParityFixtureText(t, string(associationParityCasesYAML), "name: "+manifest.RequiredCaseNames[0], "name: unregistered_association_parity_case")
		requireAssociationParityFixtureLoadFailure(t, []byte(corpusSource), associationParityManifestYAML)
	})
}

func requireAssociationParityFixtureLoadFailure(t *testing.T, corpusYAML, manifestYAML []byte) {
	t.Helper()
	if _, _, err := loadAssociationParityFixturesFromYAML(corpusYAML, manifestYAML); err == nil {
		t.Fatal("load association parity fixture corpus and manifest succeeded, want an actionable inventory error")
	}
}

func replaceAssociationParityFixtureText(t *testing.T, source, old, replacement string) string {
	t.Helper()
	if count := strings.Count(source, old); count != 1 {
		t.Fatalf("fixture mutation target %q occurs %d times, want exactly once", old, count)
	}
	return strings.Replace(source, old, replacement, 1)
}
