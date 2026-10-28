package schema_test

import (
	"reflect"
	"testing"

	schema "github.com/peasant-labs/schema"
)

func TestVillageIntegrityClosedSets(t *testing.T) {
	var want []string
	for _, enum := range loadClosedEnumFixture(t).Enums {
		if enum.Name == "VillageIntegrityCheckResult" {
			want = enum.AllValues
		}
	}
	if len(want) == 0 {
		t.Fatal("enums.yaml does not define VillageIntegrityCheckResult")
	}
	got := make([]string, 0, len(schema.AllVillageIntegrityCheckResults))
	for _, result := range schema.AllVillageIntegrityCheckResults {
		if !result.IsValid() {
			t.Fatalf("%q is listed but not valid", result)
		}
		got = append(got, result.String())
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AllVillageIntegrityCheckResults = %v, want %v", got, want)
	}
	if schema.VillageIntegrityCheckResult("clean").IsValid() || schema.VillageIntegrityCheckResult("").IsValid() {
		t.Fatal("a result outside the closed set is valid")
	}
}
