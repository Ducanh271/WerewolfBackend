package ws

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"masoi/internal/engine"
)

var (
	durationType = reflect.TypeOf(time.Duration(0))
	timeType     = reflect.TypeOf(time.Time{})
	phaseType    = reflect.TypeOf(engine.Phase(0))
	teamType     = reflect.TypeOf(engine.Team(0))
	dawnKindType = reflect.TypeOf(engine.DawnKind(0))

	// fixedShape lists the types whose fields are all sent, empty ones too:
	// the phone check notice must have the same keys for everybody, so the
	// screen cannot tell who was cursed or who is in love (rules §3).
	fixedShape = map[reflect.Type]bool{
		reflect.TypeOf(engine.PhoneChecked{}):   true,
		reflect.TypeOf(engine.PhoneCheckView{}): true,
		reflect.TypeOf(engine.LoveNotice{}):     true,
	}
)

// toTree turns an engine value (a View or an event, already projected for one
// viewer) into plain maps, slices and scalars ready for encoding/json. It only
// copies what it is given: what a viewer may see is decided before, by
// engine.ViewFor and engine.Visible.
//
// Field names become camelCase. Durations are milliseconds, times Unix
// milliseconds, phases and teams their names. Absent parts are omitted: nil
// pointers, empty maps, lists and strings, empty objects, a zero time, no
// team. Booleans and numbers always stay, and so do the entries inside a map
// or list (a nomination of "" means no one). The types in fixedShape keep
// every field.
func toTree(v any) any {
	out, _ := encodeValue(reflect.ValueOf(v), false)
	return out
}

func encodeValue(v reflect.Value, keepEmpty bool) (any, bool) {
	if !v.IsValid() {
		return nil, false
	}
	switch v.Type() {
	case durationType:
		return v.Int() / int64(time.Millisecond), true
	case timeType:
		t := v.Interface().(time.Time)
		return t.UnixMilli(), !t.IsZero()
	case phaseType:
		return engine.Phase(v.Int()).String(), true
	case dawnKindType:
		return dawnKindName(engine.DawnKind(v.Int())), true
	case teamType:
		t := engine.Team(v.Int())
		return t.String(), t != engine.TeamNone
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return nil, false
		}
		return encodeValue(v.Elem(), keepEmpty)
	case reflect.String:
		return v.String(), keepEmpty || v.String() != ""
	case reflect.Bool:
		return v.Bool(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.Slice, reflect.Array:
		list := make([]any, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			if item, ok := encodeValue(v.Index(i), true); ok {
				list = append(list, item)
			}
		}
		return list, keepEmpty || len(list) > 0
	case reflect.Map:
		m := map[string]any{}
		for _, k := range v.MapKeys() {
			if item, ok := encodeValue(v.MapIndex(k), true); ok {
				m[mapKey(k)] = item
			}
		}
		return m, keepEmpty || len(m) > 0
	case reflect.Struct:
		m := map[string]any{}
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if !f.IsExported() || f.Anonymous {
				continue // markers (EventMarker...) carry nothing
			}
			if item, ok := encodeValue(v.Field(i), fixedShape[v.Type()]); ok {
				m[camel(f.Name)] = item
			}
		}
		return m, keepEmpty || len(m) > 0
	}
	return nil, false
}

func mapKey(k reflect.Value) string {
	if k.Kind() == reflect.String {
		return k.String()
	}
	return fmt.Sprint(k.Interface())
}

// camel lowers the leading capitals of a Go name: Role -> role, NightCall ->
// nightCall, ID -> id, PlayerID -> playerID.
func camel(name string) string {
	r := []rune(name)
	n := 0
	for n < len(r) && unicode.IsUpper(r[n]) {
		n++
	}
	switch {
	case n == 0:
		return name
	case n == len(r) || n == 1:
		return strings.ToLower(string(r[:n])) + string(r[n:])
	default:
		return strings.ToLower(string(r[:n-1])) + string(r[n-1:])
	}
}

// dawnKindName is the "kind" of DawnAnnounced in docs/protocol.md.
func dawnKindName(k engine.DawnKind) string {
	switch k {
	case engine.DawnDeaths:
		return "deaths"
	case engine.DawnProtected:
		return "protected"
	default:
		return "nobodyDied"
	}
}
