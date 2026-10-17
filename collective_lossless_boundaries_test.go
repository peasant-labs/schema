package schema_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

//go:embed testdata/collective_lossless_boundaries.yaml
var collectiveLosslessYAML []byte

type collectiveLosslessFixtures struct {
	Identity      collectiveTokenInput `yaml:"identity"`
	RequiredNames struct {
		Timestamps []string `yaml:"timestamps"`
		Scopes     []string `yaml:"scopes"`
		Statuses   []string `yaml:"statuses"`
	} `yaml:"required_names"`
	Timestamps testcase.Corpus[collectiveRawInput, collectiveExpected]   `yaml:"timestamps"`
	Scopes     testcase.Corpus[collectiveTokenInput, collectiveExpected] `yaml:"scopes"`
	Statuses   testcase.Corpus[collectiveRawInput, collectiveExpected]   `yaml:"statuses"`
}

func loadCollectiveLosslessFixtures(t *testing.T) collectiveLosslessFixtures {
	t.Helper()
	var f collectiveLosslessFixtures
	if err := yaml.Unmarshal(collectiveLosslessYAML, &f); err != nil {
		t.Fatal(err)
	}
	assertcase.RequireValid(t, f.Timestamps)
	assertcase.RequireMin(t, f.Timestamps, 1)
	assertcase.RequireValid(t, f.Scopes)
	assertcase.RequireMin(t, f.Scopes, 1)
	assertcase.RequireValid(t, f.Statuses)
	assertcase.RequireMin(t, f.Statuses, 1)
	requireNames(t, "timestamp manifest", f.RequiredNames.Timestamps, namesOf(f.Timestamps))
	requireNames(t, "authority manifest", f.RequiredNames.Scopes, namesOf(f.Scopes))
	requireNames(t, "cancellation manifest", f.RequiredNames.Statuses, namesOf(f.Statuses))
	requireNames(t, "required lossless cases", strings.Fields("recorded-overprecision decision-overprecision first-seen-overprecision recorded-minute-only decision-minute-only first-seen-minute-only recorded-nine-zero decision-nine-digits recorded-zero-fraction decision-zero-fraction first-seen-zero-offset decision-equivalent-trailing-zero recorded-equivalent-trailing-zero decision-distinct-nanosecond original-overprecision-replay original-minute-replay populated-history-overprecision fetch-time-overprecision fetch-source-minute-only observation-overprecision"), namesOf(f.Timestamps))
	requireNames(t, "required raw authority cases", strings.Fields("unbracketed-ipv6-default-port unbracketed-ipv6-eight-groups bracketed-ipv6-default-port bracketed-ipv6-eight-groups"), namesOf(f.Scopes))
	requireNames(t, "required cancelled cases", strings.Fields("domain-cancel-remote scope-cancel-remote cancel-auth cancel-persistence cancel-multiple-codes"), namesOf(f.Statuses))
	return f
}

func TestCollectiveLosslessRawTimestamps(t *testing.T) {
	f := loadCollectiveLosslessFixtures(t)
	for _, c := range f.Timestamps.Cases {
		t.Run(c.Name, func(t *testing.T) {
			raw := collectiveRawBytes(t, c.Input)
			_, err := validateCollectiveFixture(c.Input.Target, raw)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("raw valid=%v error=%v", c.Expected.Valid, err)
			}
			if c.Input.Target != "LocalRemoteContributionAttempt" {
				return
			}
			var saved schema.LocalRemoteContributionAttempt
			decodeErr := json.Unmarshal(raw, &saved)
			if !c.Expected.Valid {
				if decodeErr == nil {
					t.Fatal("public raw decoder accepted unsupported evidence before reconciliation")
				}
				return
			}
			if decodeErr != nil {
				t.Fatal(decodeErr)
			}
			if c.Input.IncomingSet == nil {
				return
			}
			input := c.Input
			input.Set = map[string]interface{}{}
			for k, v := range c.Input.Set {
				input.Set[k] = v
			}
			for k, v := range c.Input.IncomingSet {
				input.Set[k] = v
			}
			var incoming schema.LocalRemoteContributionAttempt
			if err := json.Unmarshal(collectiveRawBytes(t, input), &incoming); err != nil {
				t.Fatal(err)
			}
			_, disposition, err := schema.ReconcileVillageContributionAttempt(&saved, incoming)
			if (err != nil) != c.Expected.Conflict || string(disposition) != c.Expected.Disposition {
				t.Fatalf("public reconciliation disposition=%s error=%v", disposition, err)
			}
		})
	}
}

func TestCollectiveRawAuthority(t *testing.T) {
	f := loadCollectiveLosslessFixtures(t)
	for _, c := range f.Scopes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			scope, err := schema.NormalizeCollectiveObservationScope(c.Input.URL, f.Identity.UserID)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("raw authority valid=%v scope=%+v error=%v", c.Expected.Valid, scope, err)
			}
			if !c.Expected.Valid {
				raw, _ := json.Marshal([]string{"1", c.Input.URL, string(f.Identity.UserID), string(f.Identity.Incarnation), string(f.Identity.Kind), f.Identity.Generation})
				if _, err := schema.ParseContributionTraversalRevision(schema.ContributionTraversalRevision(base64.RawURLEncoding.EncodeToString(raw))); err == nil {
					t.Fatal("token accepted unsupported raw authority")
				}
				return
			}
			if scope.VillageURL != c.Expected.Normalized {
				t.Fatalf("scope=%+v", scope)
			}
			token, err := schema.EncodeContributionTraversalRevision(schema.ContributionTraversalIdentity{Scope: scope, Incarnation: f.Identity.Incarnation, Kind: f.Identity.Kind, Generation: f.Identity.Generation})
			if err != nil || string(token) != c.Expected.Token || c.Expected.Token == "" {
				t.Fatalf("token=%s error=%v", token, err)
			}
			if _, err := schema.ParseContributionTraversalRevision(schema.ContributionTraversalRevision(c.Expected.Token)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCollectiveCancellationStatus(t *testing.T) {
	f := loadCollectiveLosslessFixtures(t)
	for _, c := range f.Statuses.Cases {
		t.Run(c.Name, func(t *testing.T) {
			raw := collectiveRawBytes(t, c.Input)
			var report schema.CollectiveFetchReport
			if err := json.Unmarshal(raw, &report); err != nil {
				t.Fatal(err)
			}
			if got := schema.CollectiveFetchHTTPStatus(report); got != c.Expected.Status {
				t.Fatalf("status=%d want=%d", got, c.Expected.Status)
			}
			if err := report.Validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
