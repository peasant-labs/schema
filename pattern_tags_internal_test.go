package schema

import (
	"reflect"
	"testing"
)

// Struct tags cannot reference constants, so each pattern is written twice.
// These tests keep the tag and the constant that Validate uses equal.
func TestPatternTagsEqualTheirConstants(t *testing.T) {
	for _, tc := range []struct {
		typ      reflect.Type
		field    string
		constant string
	}{
		{reflect.TypeOf(AuthoritativeClientInfo{}), "Name", authoritativeClientNamePattern},
		{reflect.TypeOf(AuthoritativeClientInfo{}), "Version", authoritativeClientVersionPattern},
		{reflect.TypeOf(ScryCheckResponse{}), "ScryVersion", scrySemVerCorePattern},
		{reflect.TypeOf(ScryHealthResponse{}), "ScryVersion", scrySemVerCorePattern},
	} {
		field, ok := tc.typ.FieldByName(tc.field)
		if !ok {
			t.Fatalf("%s has no field %s", tc.typ, tc.field)
		}
		if got := field.Tag.Get("pattern"); got != tc.constant {
			t.Errorf("%s.%s pattern tag %q differs from its constant %q; update the tag", tc.typ, tc.field, got, tc.constant)
		}
	}
}
