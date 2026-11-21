package schema

import (
	"bytes"
	"encoding"
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
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
	turnFieldGeneric     turnFieldKind = iota // json.Unmarshal of the member value into the field
	turnFieldString                           // a string kind without a text or JSON unmarshaler
	turnFieldInt                              // a signed integer kind without an unmarshaler
	turnFieldBool                             // a bool kind without an unmarshaler
	turnFieldUnmarshaler                      // a non-pointer type whose pointer is a json.Unmarshaler
	turnFieldWhole                            // the member forces a whole-turn json.Unmarshal
)

type turnField struct {
	index int
	kind  turnFieldKind
}

var turnFieldTable = buildTurnFieldTable()

type turnFieldTableValue struct {
	byName map[string]turnField
	kinds  []turnFieldKind // by field index
	names  []string
	usable bool
}

func buildTurnFieldTable() turnFieldTableValue {
	typ := reflect.TypeFor[TurnDetail]()
	table := turnFieldTableValue{byName: map[string]turnField{}, kinds: make([]turnFieldKind, typ.NumField()), usable: true}
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
		case reflect.PointerTo(field.Type).Implements(jsonUnmarshaler) && field.Type.Kind() != reflect.Pointer && field.Type.Kind() != reflect.Interface:
			kind = turnFieldUnmarshaler
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
		table.kinds[i] = kind
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
	const stackFields = 32
	var buffer [stackFields]int16
	fields := buffer[:0]
	if len(members) > stackFields {
		fields = make([]int16, 0, len(members))
	}
	for _, e := range members {
		f, known := turnField{}, false
		name := raw[e.nameStart:e.nameEnd]
		key := ""
		if bytes.IndexByte(name, '\\') >= 0 {
			key = transcriptDecodeName(raw[e.nameStart-1 : e.nameEnd+1])
			f, known = turnFieldTable.byName[key]
		} else {
			f, known = turnFieldTable.byName[string(name)]
		}
		if !known {
			if key == "" {
				key = string(name)
			}
			for _, field := range turnFieldTable.names {
				if strings.EqualFold(key, field) {
					return json.Unmarshal(raw, d)
				}
			}
			fields = append(fields, -1)
			continue
		}
		if f.kind == turnFieldWhole {
			return json.Unmarshal(raw, d)
		}
		fields = append(fields, int16(f.index))
	}
	v := reflect.ValueOf(d).Elem()
	for i, e := range members {
		if fields[i] < 0 {
			continue
		}
		index := int(fields[i])
		kind := turnFieldTable.kinds[index]
		value := raw[e.valueStart:e.valueEnd]
		fv := v.Field(index)
		if kind != turnFieldGeneric && kind != turnFieldUnmarshaler && string(value) == "null" {
			continue // encoding/json leaves a scalar field unchanged for null.
		}
		switch kind {
		case turnFieldString:
			if value[0] == '"' {
				if text, ok := transcriptUnquote(value); ok {
					fv.SetString(text)
					continue
				}
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
		case turnFieldUnmarshaler:
			// encoding/json hands a validated value, null included, to the
			// field's UnmarshalJSON unchanged.
			if err := fv.Addr().Interface().(json.Unmarshaler).UnmarshalJSON(value); err != nil {
				return err
			}
			continue
		}
		if err := json.Unmarshal(value, fv.Addr().Interface()); err != nil {
			return err
		}
	}
	return nil
}

// transcriptUnquote decodes a validated JSON string literal as encoding/json
// does. It reports false, leaving the decode to encoding/json, for a surrogate
// escape that is not a valid pair, which encoding/json replaces.
func transcriptUnquote(quoted []byte) (string, bool) {
	body := quoted[1 : len(quoted)-1]
	escape := bytes.IndexByte(body, '\\')
	if escape < 0 {
		return string(body), true
	}
	out := make([]byte, 0, len(body))
	out = append(out, body[:escape]...)
	for i := escape; i < len(body); {
		c := body[i]
		if c != '\\' {
			j := i + 1
			for j < len(body) && body[j] != '\\' {
				j++
			}
			out = append(out, body[i:j]...)
			i = j
			continue
		}
		switch body[i+1] {
		case '"', '\\', '/':
			out = append(out, body[i+1])
		case 'b':
			out = append(out, '\b')
		case 'f':
			out = append(out, '\f')
		case 'n':
			out = append(out, '\n')
		case 'r':
			out = append(out, '\r')
		case 't':
			out = append(out, '\t')
		case 'u':
			r, _ := captureHex(body[i+2 : i+6])
			if r >= 0xD800 && r <= 0xDFFF {
				if r > 0xDBFF || i+12 > len(body) || body[i+6] != '\\' || body[i+7] != 'u' {
					return "", false
				}
				low, _ := captureHex(body[i+8 : i+12])
				if low < 0xDC00 || low > 0xDFFF {
					return "", false
				}
				out = utf8.AppendRune(out, (r-0xD800)<<10|(low-0xDC00)+0x10000)
				i += 12
				continue
			}
			out = utf8.AppendRune(out, r)
			i += 6
			continue
		}
		i += 2
	}
	return string(out), true
}
