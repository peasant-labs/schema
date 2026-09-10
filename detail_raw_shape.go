package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// validateWireShape checks presence before struct decoding erases it. Required
// fields follow the same Go JSON tags as contract generation. Slices and pointer
// fields retain the contract's nullability, including legacy turns:null.
func validateWireShape(raw json.RawMessage, typ reflect.Type, path string) error {
	fail := func(reason string) error {
		return fmt.Errorf("wire shape validation failed at schema raw detail decoder during pre-decode validation of %s: %s; typed decoding would erase required evidence; send the published field shape and enum value", path, reason)
	}
	if typ == reflect.TypeFor[json.RawMessage]() {
		return nil
	}
	if isNullRaw(raw) {
		if typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
			return nil
		}
		return fail("non-null value required")
	}
	if typ.Kind() == reflect.Pointer {
		return validateWireShape(raw, typ.Elem(), path)
	}
	decoded := reflect.New(typ)
	if err := json.Unmarshal(raw, decoded.Interface()); err != nil {
		return fail(err.Error())
	}
	if typ == reflect.TypeFor[Harness]() {
		found := false
		for _, h := range Harnesses() {
			if string(h) == decoded.Elem().String() {
				found = true
				break
			}
		}
		if !found {
			return fail("harness is outside its closed set")
		}
	}
	if enum, ok := decoded.Elem().Interface().(interface{ IsValid() bool }); ok && !enum.IsValid() {
		return fail("enum is outside its closed set")
	}
	if typ == reflect.TypeFor[ObservedModelID]() {
		if _, err := NewObservedModelID(decoded.Elem().String()); err != nil {
			return fail(err.Error())
		}
	}
	// Scalar codecs such as time.Time own their wire representation.
	if _, ok := decoded.Interface().(json.Unmarshaler); ok {
		return nil
	}
	switch typ.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return fail("object required")
		}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			tag := strings.Split(field.Tag.Get("json"), ",")
			if field.PkgPath != "" || tag[0] == "-" || tag[0] == "" {
				continue
			}
			value, exists := fields[tag[0]]
			if !exists {
				if !strings.Contains(field.Tag.Get("json"), ",omitempty") {
					return fail("required field " + tag[0] + " is missing")
				}
				continue
			}
			if field.Tag.Get("nullable") == "false" && isNullRaw(value) {
				return fail("field " + tag[0] + " must not be null")
			}
			if err := validateWireShape(value, field.Type, path+"/"+tag[0]); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return fail("array required")
		}
		for i, item := range items {
			if err := validateWireShape(item, typ.Elem(), fmt.Sprintf("%s/%d", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}
