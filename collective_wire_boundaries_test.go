package schema_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

//go:embed testdata/collective_wire_boundaries.yaml
var collectiveWireYAML []byte

type collectiveRawInput struct {
	Set         map[string]interface{} `yaml:"set"`
	IncomingSet map[string]interface{} `yaml:"incomingSet"`
	Target      string                 `yaml:"target"`
	Payload     map[string]interface{} `yaml:"payload"`
	Omit        []string               `yaml:"omit"`
}
type collectiveWireFixtures struct {
	Identity      collectiveTokenInput `yaml:"identity"`
	RequiredNames struct {
		Scopes   []string `yaml:"scopes"`
		Payloads []string `yaml:"payloads"`
		Statuses []string `yaml:"statuses"`
	} `yaml:"required_names"`
	Scopes   testcase.Corpus[collectiveTokenInput, collectiveExpected] `yaml:"scopes"`
	Payloads testcase.Corpus[collectiveRawInput, collectiveExpected]   `yaml:"payloads"`
	Statuses testcase.Corpus[collectiveRawInput, collectiveExpected]   `yaml:"statuses"`
}

func loadCollectiveWireFixtures(t *testing.T) collectiveWireFixtures {
	t.Helper()
	var f collectiveWireFixtures
	if err := yaml.Unmarshal(collectiveWireYAML, &f); err != nil {
		t.Fatal(err)
	}
	assertcase.RequireValid(t, f.Scopes)
	assertcase.RequireMin(t, f.Scopes, 1)
	assertcase.RequireValid(t, f.Payloads)
	assertcase.RequireMin(t, f.Payloads, 1)
	assertcase.RequireValid(t, f.Statuses)
	assertcase.RequireMin(t, f.Statuses, 1)
	requireNames(t, "scope boundary manifest", f.RequiredNames.Scopes, namesOf(f.Scopes))
	requireNames(t, "raw payload manifest", f.RequiredNames.Payloads, namesOf(f.Payloads))
	requireNames(t, "raw status manifest", f.RequiredNames.Statuses, namesOf(f.Statuses))
	requireNames(t, "required scope boundaries", strings.Fields("scope-port-zero scope-slash-dot-ambiguity scope-abbreviated-ipv4 scope-integer-ipv4 scope-hex-ipv4 scope-octal-ipv4 scope-mixed-hex-ipv4 scope-three-part-ipv4 scope-ipv4-out-of-range scope-ordinary-ipv4 scope-decimal-ipv4-port scope-ipv6-compressed scope-ipv6-mapped scope-ascii-idna scope-unicode-host scope-escaped-path scope-dot-escaped-path scope-repeated-slash scope-canonical-path scope-reserved-path"), namesOf(f.Scopes))
	requireNames(t, "required raw boundaries", strings.Fields("pending-explicit-null-decision auto-approved-explicit-null-decision approved-decision-nanoseconds legacy-rejected-explicit-null-decision retracted-explicit-null-decision revoked-explicit-null-decision pending-missing-actor pending-null-actor pending-missing-decision attempt-missing-ordinal attempt-null-ordinal attempt-missing-status attempt-null-status attempt-missing-recorded-time attempt-null-recorded-time history-valid-populated-attempt history-valid-legacy-rejected-null-then-newer history-missing-actor history-null-actor history-missing-decision discovery-zero-counters discovery-positive-counters discovery-approved-missing discovery-approved-null discovery-approved-negative discovery-pending-missing discovery-pending-null discovery-pending-negative discovery-rejected-missing discovery-rejected-null discovery-rejected-negative discovery-withdrawn-missing discovery-withdrawn-null discovery-withdrawn-negative discovery-empty-response discovery-null-response discovery-missing-response discovery-nested-missing-counter discovery-counter-overflow discovery-invalid-identity organization-zero-counters organization-missing-member-count organization-null-member-count organization-missing-transcript-count organization-null-transcript-count organization-negative-member-count organization-negative-transcript-count associations-complete-updated associations-complete-unchanged associations-complete-contradictory-commit associations-complete-empty"), namesOf(f.Payloads))
	requireNames(t, "required raw status boundaries", strings.Fields("status-valid-complete status-valid-partial status-invalid-domain status-invalid-coverage-complete status-empty-issue-message status-empty-issue-retry status-missing-domains status-null-domains status-unsafe-count status-contradictory-complete-association"), namesOf(f.Statuses))
	return f
}

func collectiveRawBytes(t *testing.T, input collectiveRawInput) []byte {
	t.Helper()
	data, err := json.Marshal(input.Payload)
	if err != nil {
		t.Fatal(err)
	}
	var object interface{}
	if err = json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	for _, path := range input.Omit {
		parts := strings.Split(path, ".")
		node := object
		for _, part := range parts[:len(parts)-1] {
			switch v := node.(type) {
			case map[string]interface{}:
				node = v[part]
			case []interface{}:
				if part != "0" || len(v) == 0 {
					t.Fatal("unsupported fixture index")
				}
				node = v[0]
			default:
				t.Fatal("fixture omission path is absent")
			}
		}
		delete(node.(map[string]interface{}), parts[len(parts)-1])
	}
	for path, value := range input.Set {
		parts := strings.Split(path, ".")
		node := object
		for _, part := range parts[:len(parts)-1] {
			switch v := node.(type) {
			case map[string]interface{}:
				node = v[part]
			case []interface{}:
				i, err := strconv.Atoi(part)
				if err != nil || i < 0 || i >= len(v) {
					t.Fatal("invalid fixture index")
				}
				node = v[i]
			default:
				t.Fatal("absent fixture set path")
			}
		}
		node.(map[string]interface{})[parts[len(parts)-1]] = value
	}
	data, err = json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCollectiveRawWireBoundaries(t *testing.T) {
	f := loadCollectiveWireFixtures(t)
	for _, c := range f.Payloads.Cases {
		t.Run(c.Name, func(t *testing.T) {
			data := collectiveRawBytes(t, c.Input)
			_, err := validateCollectiveFixture(c.Input.Target, data)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v error=%v", c.Expected.Valid, err)
			}
			if err == nil && (c.Input.Target == "VillageShareEvent" || c.Input.Target == "LocalContributionHistoryResponse") {
				var decoded interface{}
				if c.Input.Target == "VillageShareEvent" {
					decoded = new(schema.VillageShareEvent)
				} else {
					decoded = new(schema.LocalContributionHistoryResponse)
				}
				if err := json.Unmarshal(data, decoded); err != nil {
					t.Fatal(err)
				}
				roundtrip, err := json.Marshal(decoded)
				if err != nil {
					t.Fatal(err)
				}
				var before, after interface{}
				if err = json.Unmarshal(data, &before); err != nil {
					t.Fatal(err)
				}
				if err = json.Unmarshal(roundtrip, &after); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(before, after) {
					t.Fatal("raw attempt history lost decision presence or nanosecond precision")
				}
			}
		})
	}
	for _, c := range f.Statuses.Cases {
		t.Run(c.Name, func(t *testing.T) {
			data := collectiveRawBytes(t, c.Input)
			var report schema.CollectiveFetchReport
			err := json.Unmarshal(data, &report)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v error=%v", c.Expected.Valid, err)
			}
			if got := schema.CollectiveFetchHTTPStatus(report); got != c.Expected.Status {
				t.Fatalf("raw report status=%d want %d", got, c.Expected.Status)
			}
		})
	}
}

func TestCollectiveScopeWireIdentity(t *testing.T) {
	f := loadCollectiveWireFixtures(t)
	for _, c := range f.Scopes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			scope, err := schema.NormalizeCollectiveObservationScope(c.Input.URL, f.Identity.UserID)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v normalized=%+v error=%v", c.Expected.Valid, scope, err)
			}
			if !c.Expected.Valid {
				raw, err := json.Marshal([]string{"1", c.Input.URL, string(f.Identity.UserID), string(f.Identity.Incarnation), string(f.Identity.Kind), f.Identity.Generation})
				if err != nil {
					t.Fatal(err)
				}
				if _, err = schema.ParseContributionTraversalRevision(schema.ContributionTraversalRevision(base64.RawURLEncoding.EncodeToString(raw))); err == nil {
					t.Fatal("unsupported scope accepted through framed token")
				}
				return
			}
			if scope.VillageURL != c.Expected.Normalized || c.Expected.Token == "" {
				t.Fatalf("normalized=%s expected=%s missing golden=%v", scope.VillageURL, c.Expected.Normalized, c.Expected.Token == "")
			}
			identity := schema.ContributionTraversalIdentity{Scope: scope, Incarnation: f.Identity.Incarnation, Kind: f.Identity.Kind, Generation: f.Identity.Generation}
			token, err := schema.EncodeContributionTraversalRevision(identity)
			if err != nil || string(token) != c.Expected.Token {
				t.Fatalf("exact token bytes mismatch: %s error=%v", token, err)
			}
			parsed, err := schema.ParseContributionTraversalRevision(schema.ContributionTraversalRevision(c.Expected.Token))
			if err != nil || !reflect.DeepEqual(parsed, identity) {
				t.Fatalf("golden token cross-acceptance failed: %+v %v", parsed, err)
			}
		})
	}
}
