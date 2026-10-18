package schema_test

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/collective_canonical_fields.yaml
var collectiveCanonicalFieldsYAML []byte

type collectiveCanonicalInput struct {
	Target string `yaml:"target"`
	Raw    string `yaml:"raw"`
}
type collectiveCanonicalExpected struct {
	Valid       bool   `yaml:"valid"`
	Decoded     string `yaml:"decoded"`
	Disposition string `yaml:"disposition"`
	Status      int    `yaml:"status"`
}
type collectiveCanonicalFixtures struct {
	RequiredNames []string                                                               `yaml:"required_names"`
	Cases         []testcase.Case[collectiveCanonicalInput, collectiveCanonicalExpected] `yaml:"cases"`
}

func loadCollectiveCanonicalFixtures(t *testing.T) collectiveCanonicalFixtures {
	t.Helper()
	var f collectiveCanonicalFixtures
	if err := yaml.Unmarshal(collectiveCanonicalFieldsYAML, &f); err != nil {
		t.Fatal(err)
	}
	c := testcase.Corpus[collectiveCanonicalInput, collectiveCanonicalExpected]{Cases: f.Cases}
	assertcase.RequireValid(t, c)
	assertcase.RequireMin(t, c, 1)
	requireNames(t, "canonical-field manifest", f.RequiredNames, namesOf(c))
	requireNames(t, "required canonical-field behaviors", strings.Fields("decision-uppercase-after decision-uppercase-before recorded-uppercase-after actor-uppercase-after status-case-variant counter-uppercase-after counter-uppercase-before counter-missing-canonical decision-missing-canonical attempt-uppercase-object first-seen-case-variant populated-history-nested-alias strict-history-aliased-selected strict-history-aliased-other-arm strict-history-aliased-scope scope-case-variant organization-aliased-array discovery-aliased-array report-aliased-array source-aliased-date scoped-aliased-array open-unrelated-bad-time exact-duplicate-counter-last exact-duplicate-decision-last exact-duplicate-object-last exact-duplicate-invalid-time-last"), namesOf(c))
	return f
}
func canonicalDecodeTarget(name string) interface{} {
	switch name {
	case "VillageShareEvent":
		return new(schema.VillageShareEvent)
	case "VillageContributedCollective":
		return new(schema.VillageContributedCollective)
	case "VillageContributedCollectivesResponse":
		return new(schema.VillageContributedCollectivesResponse)
	case "LocalRemoteContributionAttempt":
		return new(schema.LocalRemoteContributionAttempt)
	case "LocalContributionHistoryResponse":
		return new(schema.LocalContributionHistoryResponse)
	case "CollectiveObservationScope":
		return new(schema.CollectiveObservationScope)
	case "VillageOrganizationCollectivesResponse":
		return new(schema.VillageOrganizationCollectivesResponse)
	case "CollectiveFetchReport":
		return new(schema.CollectiveFetchReport)
	case "CollectiveFetchSourceResult":
		return new(schema.CollectiveFetchSourceResult)
	case "CollectiveFetchAssociationResults":
		return new(schema.CollectiveFetchAssociationResults)
	default:
		panic("unknown canonical-field fixture target " + name)
	}
}
func TestCollectiveCanonicalFields(t *testing.T) {
	for _, c := range loadCollectiveCanonicalFixtures(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			decoded := canonicalDecodeTarget(c.Input.Target)
			err := json.Unmarshal([]byte(c.Input.Raw), decoded)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("public raw decode valid=%v error=%v", c.Expected.Valid, err)
			}
			if !c.Expected.Valid {
				return
			}
			if validator, ok := decoded.(interface{ Validate() error }); ok {
				if err := validator.Validate(); err != nil {
					t.Fatal(err)
				}
			}
			actualJSON, err := json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			var actual, want interface{}
			if err = json.Unmarshal(actualJSON, &actual); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal([]byte(c.Expected.Decoded), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, want) {
				t.Fatalf("decoded canonical evidence differs:\nactual %s\nwant %s", actualJSON, c.Expected.Decoded)
			}
			if report, ok := decoded.(*schema.CollectiveFetchReport); ok {
				if got := schema.CollectiveFetchHTTPStatus(*report); got != c.Expected.Status {
					t.Fatalf("status=%d want=%d", got, c.Expected.Status)
				}
			}
			if c.Expected.Disposition != "" {
				saved := canonicalDecodeTarget(c.Input.Target)
				if err := json.Unmarshal([]byte(c.Expected.Decoded), saved); err != nil {
					t.Fatal(err)
				}
				var old, next schema.LocalRemoteContributionAttempt
				switch value := decoded.(type) {
				case *schema.LocalRemoteContributionAttempt:
					old = *saved.(*schema.LocalRemoteContributionAttempt)
					next = *value
				case *schema.LocalContributionHistoryResponse:
					old = saved.(*schema.LocalContributionHistoryResponse).RemoteAttempts[0]
					next = value.RemoteAttempts[0]
				default:
					t.Fatal("fixture disposition requires attempt evidence")
				}
				result, disposition, err := schema.ReconcileVillageContributionAttempt(&old, next)
				if err != nil || string(disposition) != c.Expected.Disposition || !reflect.DeepEqual(result, old) {
					t.Fatalf("canonical replay disposition=%s error=%v", disposition, err)
				}
			}
		})
	}
}
