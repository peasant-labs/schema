package schema

import (
	"bytes"
	"encoding"
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// The typed decode at capture assigns each top-level member of a turn to its
// TurnDetail field from the member table the scan built, instead of running
// encoding/json over the whole turn again. It is equivalent to
// json.Unmarshal(turn, &TurnDetail{}) because, for an object whose members
// have distinct names that each match a field name exactly or no field name
// at all, that call decodes each matched member into its zero-valued field
// independently and ignores the rest. A member whose name matches a field
// only case-insensitively, or a field whose tag changes its own decoding,
// makes the turn take json.Unmarshal itself.

type turnFieldKind uint8

const (
	turnFieldGeneric turnFieldKind = iota // json.Unmarshal of the member value into the field
	turnFieldString                       // a string kind without a text or JSON unmarshaler
	turnFieldInt                          // a signed integer kind without an unmarshaler
	turnFieldBool                         // a bool kind without an unmarshaler
	turnFieldWhole                        // the member forces a whole-turn json.Unmarshal
)

type turnField struct {
	index int
	kind  turnFieldKind
}

var turnFieldTable = buildTurnFieldTable()

type turnFieldTableValue struct {
	byName map[string]turnField
	names  []string
	usable bool
}

func buildTurnFieldTable() turnFieldTableValue {
	typ := reflect.TypeFor[TurnDetail]()
	table := turnFieldTableValue{byName: map[string]turnField{}, usable: true}
	jsonUnmarshaler := reflect.TypeFor[json.Unmarshaler]()
	textUnmarshaler := reflect.TypeFor[encoding.TextUnmarshaler]()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			table.usable = false
			continue
		}
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, options, _ := strings.Cut(tag, ",")
		if name == "" {
			name = field.Name
		}
		kind := turnFieldGeneric
		custom := reflect.PointerTo(field.Type).Implements(jsonUnmarshaler) || reflect.PointerTo(field.Type).Implements(textUnmarshaler)
		switch {
		case strings.Contains(","+options+",", ",string,"):
			kind = turnFieldWhole
		case custom:
		case field.Type.Kind() == reflect.String:
			kind = turnFieldString
		case field.Type.Kind() >= reflect.Int && field.Type.Kind() <= reflect.Int64:
			kind = turnFieldInt
		case field.Type.Kind() == reflect.Bool:
			kind = turnFieldBool
		}
		if _, exists := table.byName[name]; exists {
			table.usable = false
		}
		table.byName[name] = turnField{index: i, kind: kind}
		table.names = append(table.names, name)
	}
	return table
}

// decodeCapturedTurn decodes one validated turn object into d, which must be
// the zero value.
func decodeCapturedTurn(raw []byte, members []transcriptMember, d *TurnDetail) error {
	if !turnFieldTable.usable || uint64(len(raw)) > math.MaxUint32 {
		return json.Unmarshal(raw, d)
	}
	for _, e := range members {
		name := raw[e.nameStart:e.nameEnd]
		key := string(name)
		if bytes.IndexByte(name, '\\') >= 0 {
			key = transcriptDecodeName(raw[e.nameStart-1 : e.nameEnd+1])
		}
		if f, ok := turnFieldTable.byName[key]; ok {
			if f.kind == turnFieldWhole {
				return json.Unmarshal(raw, d)
			}
			continue
		}
		for _, field := range turnFieldTable.names {
			if strings.EqualFold(key, field) {
				return json.Unmarshal(raw, d)
			}
		}
	}
	v := reflect.ValueOf(d).Elem()
	for _, e := range members {
		name := raw[e.nameStart:e.nameEnd]
		key := string(name)
		if bytes.IndexByte(name, '\\') >= 0 {
			key = transcriptDecodeName(raw[e.nameStart-1 : e.nameEnd+1])
		}
		f, ok := turnFieldTable.byName[key]
		if !ok {
			continue
		}
		value := raw[e.valueStart:e.valueEnd]
		fv := v.Field(f.index)
		if f.kind != turnFieldGeneric && string(value) == "null" {
			continue // encoding/json leaves a scalar field unchanged for null.
		}
		switch f.kind {
		case turnFieldString:
			if value[0] == '"' && bytes.IndexByte(value, '\\') < 0 {
				fv.SetString(string(value[1 : len(value)-1]))
				continue
			}
		case turnFieldInt:
			if value[0] == '-' || ('0' <= value[0] && value[0] <= '9') {
				if n, err := strconv.ParseInt(string(value), 10, 64); err == nil && !fv.OverflowInt(n) {
					fv.SetInt(n)
					continue
				}
			}
		case turnFieldBool:
			if string(value) == "true" || string(value) == "false" {
				fv.SetBool(value[0] == 't')
				continue
			}
		}
		if err := json.Unmarshal(value, fv.Addr().Interface()); err != nil {
			return err
		}
	}
	return nil
}
