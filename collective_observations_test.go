package schema_test

import (
	_ "embed"
	"encoding/json"
	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	assertcase "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
	"testing"
)

//go:embed testdata/collective_observations.yaml
var collectiveObservationFixtureYAML []byte

type collectiveValidationInput struct {
	Target  string                 `yaml:"target"`
	Payload map[string]interface{} `yaml:"payload"`
}
type collectiveExpected struct {
	Valid      bool   `yaml:"valid"`
	Status     int    `yaml:"status"`
	Normalized string `yaml:"normalized"`
	Token      string `yaml:"token"`
}
type collectiveTokenInput struct {
	URL         string                               `yaml:"url"`
	UserID      schema.VillageUUID                   `yaml:"userId"`
	Incarnation schema.VillageUUID                   `yaml:"incarnation"`
	Kind        schema.LocalContributionEvidenceKind `yaml:"kind"`
	QueryKind   schema.LocalContributionEvidenceKind `yaml:"queryKind"`
	Generation  string                               `yaml:"generation"`
	Page        int                                  `yaml:"page"`
	Limit       int                                  `yaml:"limit"`
	Revision    string                               `yaml:"revision"`
}
type collectiveObservationFixtures struct {
	Menus             map[string][]string `yaml:"menus"`
	InvalidMenuValues []string            `yaml:"invalid_menu_values"`
	RequiredNames     struct {
		Validation    []string `yaml:"validation"`
		Tokens        []string `yaml:"tokens"`
		History       []string `yaml:"history"`
		Organizations []string `yaml:"organizations"`
	} `yaml:"required_names"`
	Validation    testcase.Corpus[collectiveValidationInput, collectiveExpected] `yaml:"validation"`
	Tokens        testcase.Corpus[collectiveTokenInput, collectiveExpected]      `yaml:"tokens"`
	History       testcase.Corpus[collectiveValidationInput, collectiveExpected] `yaml:"history"`
	Organizations testcase.Corpus[collectiveValidationInput, collectiveExpected] `yaml:"organizations"`
}

func TestCollectiveClosedMenus(t *testing.T) {
	f := loadCollectiveObservationFixtures(t)
	required := strings.Fields("CollectiveObservationReason CollectiveObservationState CollectiveOperation OperationIssueCode LocalContributionEvidenceKind CollectiveEvidenceCoverage LocalConfirmedContributionOutcome LocalOperationDiagnosticOutcome PullAssociationMetadataState CollectiveFetchDomain CollectiveFetchState CollectiveFetchCommit CollectiveFetchOutcome CollectiveObservationSource CollectiveSourceOutcome")
	actual := map[string]bool{}
	for name := range f.Menus {
		actual[name] = true
	}
	requireNames(t, "closed collective menus", required, actual)
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	for name, expected := range f.Menus {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(spec.Components.Schemas[name]["enum"])
			if err != nil {
				t.Fatal(err)
			}
			var values []string
			if err = json.Unmarshal(data, &values); err != nil {
				t.Fatal(err)
			}
			got := map[string]bool{}
			for _, v := range values {
				if got[v] {
					t.Fatal("duplicate enum member")
				}
				got[v] = true
			}
			requireNames(t, name, expected, got)
			for _, v := range f.InvalidMenuValues {
				if got[v] {
					t.Fatalf("invalid closed enum member %s", v)
				}
			}
		})
	}
}

func loadCollectiveObservationFixtures(t *testing.T) collectiveObservationFixtures {
	t.Helper()
	var f collectiveObservationFixtures
	if err := yaml.Unmarshal(collectiveObservationFixtureYAML, &f); err != nil {
		t.Fatal(err)
	}
	assertcase.RequireValid(t, f.Validation)
	assertcase.RequireMin(t, f.Validation, 1)
	assertcase.RequireValid(t, f.Tokens)
	assertcase.RequireMin(t, f.Tokens, 1)
	requireNames(t, "collective validation", f.RequiredNames.Validation, namesOf(f.Validation))
	requireNames(t, "collective tokens", f.RequiredNames.Tokens, namesOf(f.Tokens))
	assertcase.RequireValid(t, f.History)
	assertcase.RequireValid(t, f.Organizations)
	assertcase.RequireMin(t, f.Organizations, 1)
	requireNames(t, "organization projection manifest", f.RequiredNames.Organizations, namesOf(f.Organizations))
	requireNames(t, "required organization projections", strings.Fields("organization-projection-known-empty organization-projection-duplicate-refused organization-projection-query-not-empty"), namesOf(f.Organizations))
	assertcase.RequireMin(t, f.History, 1)
	requireNames(t, "collective history", f.RequiredNames.History, namesOf(f.History))
	requireNames(t, "required validation behaviors", strings.Fields("metadata-observed metadata-stale metadata-unknown metadata-invalidated metadata-no-invented-time metadata-warning-json metadata-observed-no-warning metadata-stale-needs-time receipt-published publication-not-share-approval receipt-submitted receipt-no-fetch-success diagnostic-uncertain diagnostic-not-success independent-fetch-domain-times catalog-complete-history-timeout catalog-committed-before-cancel-reported persistence-status-precedence auth-status-precedence deadline-status-precedence superseded-status complete-not-partial incomplete-domain-no-install no-incomplete-source-install association-one-complete-one-timeout association-partial-effect-not-lost association-count-mismatch"), namesOf(f.Validation))
	requireNames(t, "required token behaviors", strings.Fields("scope-normalized scope-base-path scope-html-path same-host-other-principal evidence-token-other-kind-409 evidence-recreated-scope-old-token-409 evidence-selected-change-409 evidence-missing-revision-page-two-400 evidence-malformed-token-400 evidence-wrong-kind-400 evidence-token-overflow evidence-token-version evidence-empty-beyond-total"), namesOf(f.Tokens))
	requireNames(t, "required history behaviors", strings.Fields("evidence-remote-only evidence-receipts-only evidence-diagnostics-only evidence-no-extra-category evidence-no-null-extra-category evidence-selected-array-nonnull evidence-token-kind-mismatch evidence-page-limit evidence-count-safe-integer evidence-total-required evidence-total-not-null evidence-no-cache-revision"), namesOf(f.History))
	return f
}
func validateCollectiveFixture(target string, data []byte) (int, error) {
	var v interface{ Validate() error }
	switch target {
	case "PullAssociationMetadataResult":
		v = new(schema.PullAssociationMetadataResult)
	case "LocalConfirmedContributionReceipt":
		v = new(schema.LocalConfirmedContributionReceipt)
	case "LocalOperationDiagnostic":
		v = new(schema.LocalOperationDiagnostic)
	case "CollectiveFetchReport":
		v = new(schema.CollectiveFetchReport)
	case "CollectiveFetchDomainResult":
		v = new(schema.CollectiveFetchDomainResult)
	case "CollectiveObservationResponse":
		v = new(schema.CollectiveObservationResponse)
	case "VillageOrganizationCollectivesResponse":
		v = new(schema.VillageOrganizationCollectivesResponse)
	case "LocalContributionHistoryResponse":
		v = new(schema.LocalContributionHistoryResponse)
	default:
		panic("unknown fixture target " + target)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return 0, err
	}
	if err := v.Validate(); err != nil {
		return 0, err
	}
	if report, ok := v.(*schema.CollectiveFetchReport); ok {
		return schema.CollectiveFetchHTTPStatus(*report), nil
	}
	return 0, nil
}
func TestCollectiveObservationContract(t *testing.T) {
	f := loadCollectiveObservationFixtures(t)
	for _, c := range append(f.Validation.Cases, f.Organizations.Cases...) {
		t.Run(c.Name, func(t *testing.T) {
			data, err := json.Marshal(c.Input.Payload)
			if err != nil {
				t.Fatal(err)
			}
			before := string(data)
			status, err := validateCollectiveFixture(c.Input.Target, data)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v error=%v", c.Expected.Valid, err)
			}
			if err == nil && status != c.Expected.Status {
				t.Fatalf("status=%d want %d", status, c.Expected.Status)
			}
			if string(data) != before {
				t.Fatal("validation mutated source")
			}
		})
	}
}

func TestCollectiveHistoryStrictUnion(t *testing.T) {
	f := loadCollectiveObservationFixtures(t)
	for _, c := range f.History.Cases {
		t.Run(c.Name, func(t *testing.T) {
			data, err := json.Marshal(c.Input.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var object map[string]json.RawMessage
			if err = json.Unmarshal(data, &object); err != nil {
				t.Fatal(err)
			}
			var scope schema.CollectiveObservationScope
			var kind schema.LocalContributionEvidenceKind
			var placeholder string
			if err = json.Unmarshal(object["scope"], &scope); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(object["kind"], &kind); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(object["traversalRevision"], &placeholder); err != nil {
				t.Fatal(err)
			}
			if placeholder == "$other-token" {
				kind = schema.LocalContributionEvidenceReceipts
			}
			token, err := schema.EncodeContributionTraversalRevision(schema.ContributionTraversalIdentity{Scope: scope, Kind: kind, Incarnation: f.Tokens.Cases[0].Input.Incarnation, Generation: "1"})
			if err != nil {
				t.Fatal(err)
			}
			object["traversalRevision"], err = json.Marshal(token)
			if err != nil {
				t.Fatal(err)
			}
			data, err = json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			_, err = validateCollectiveFixture(c.Input.Target, data)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v error=%v", c.Expected.Valid, err)
			}
			if err == nil {
				var response schema.LocalContributionHistoryResponse
				if err = json.Unmarshal(data, &response); err != nil {
					t.Fatal(err)
				}
				roundtrip, err := json.Marshal(response)
				if err != nil {
					t.Fatal(err)
				}
				var result map[string]json.RawMessage
				if err = json.Unmarshal(roundtrip, &result); err != nil {
					t.Fatal(err)
				}
				var decoded, expected interface{}
				if err = json.Unmarshal(roundtrip, &decoded); err != nil {
					t.Fatal(err)
				}
				if err = json.Unmarshal(data, &expected); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(decoded, expected) {
					t.Fatalf("selected empty arrays did not survive JSON: %s", roundtrip)
				}
			}
		})
	}
}
func TestCollectiveTraversalContract(t *testing.T) {
	for _, c := range loadCollectiveObservationFixtures(t).Tokens.Cases {
		t.Run(c.Name, func(t *testing.T) {
			scope, err := schema.NormalizeCollectiveObservationScope(c.Input.URL, c.Input.UserID)
			if err != nil {
				if c.Expected.Valid {
					t.Fatal(err)
				}
				return
			}
			identity := schema.ContributionTraversalIdentity{Scope: scope, Incarnation: c.Input.Incarnation, Kind: c.Input.Kind, Generation: c.Input.Generation}
			current, err := schema.EncodeContributionTraversalRevision(identity)
			if (err == nil) != c.Expected.Valid {
				t.Fatalf("valid=%v error=%v", c.Expected.Valid, err)
			}
			if err != nil {
				return
			}
			if scope.VillageURL != c.Expected.Normalized {
				t.Fatalf("normalized=%s", scope.VillageURL)
			}
			if c.Expected.Token != "" && string(current) != c.Expected.Token {
				t.Fatalf("token=%s", current)
			}
			parsed, err := schema.ParseContributionTraversalRevision(current)
			if err != nil || !reflect.DeepEqual(parsed, identity) {
				t.Fatalf("codec roundtrip: %+v %v", parsed, err)
			}
			query := schema.LocalContributionHistoryQuery{Kind: c.Input.Kind, Page: c.Input.Page, Limit: c.Input.Limit, Revision: &current}
			if c.Input.QueryKind != "" {
				query.Kind = c.Input.QueryKind
			}
			other := identity
			switch c.Input.Revision {
			case "none":
				query.Revision = nil
			case "malformed":
				v := schema.ContributionTraversalRevision("not-a-token")
				query.Revision = &v
			case "other-kind":
				other.Kind = schema.LocalContributionEvidenceReceipts
			case "other-user":
				other.Scope.UserID = c.Input.Incarnation
			case "other-incarnation":
				other.Incarnation = c.Input.UserID
			case "older-generation":
				other.Generation = "0"
			case "other-scope":
				other.Scope.VillageURL += "/other"
			}
			if other != identity {
				token, err := schema.EncodeContributionTraversalRevision(other)
				if err != nil {
					t.Fatal(err)
				}
				query.Revision = &token
			}
			if got := schema.ValidateContributionHistoryQuery(query, current); got != c.Expected.Status {
				t.Fatalf("status=%d want %d", got, c.Expected.Status)
			}
		})
	}
}
