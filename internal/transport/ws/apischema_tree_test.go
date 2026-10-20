package ws

import (
	"fmt"
	"reflect"

	"masoi/internal/engine"
)

// The API schema (docs/api/schema, JSON Schema draft 2020-12) is generated
// from the Go types. treeSchema describes what toTree (jsonenc.go) makes of a
// type, rule for rule, so the schema says exactly what the encoder may send;
// every frame of the example games is then validated against it
// (apidocs_test.go), which catches a rule here that drifted from the encoder.

var (
	playerIDType   = reflect.TypeOf(engine.PlayerID(""))
	roleIDType     = reflect.TypeOf(engine.RoleID(""))
	linkColourType = reflect.TypeOf(engine.LinkColour(""))
)

// alwaysFilled lists string fields the server always fills, though toTree
// would leave an empty one out: the schema marks them required. The example
// games check it.
var alwaysFilled = map[string]bool{
	"Person.ID": true, "Person.Nickname": true,
	"PlayerView.ID": true, "PlayerView.Nickname": true,
	"MyView.ID":          true,
	"NightCallView.Role": true,
	"TurnView.Step":      true,
}

// schemaGen collects the named types it meets as $defs.
type schemaGen struct {
	defs  map[string]any
	owner map[string]reflect.Type
	docs  goDocs
	errs  []error
}

func newSchemaGen(docs goDocs) *schemaGen {
	return &schemaGen{defs: map[string]any{}, owner: map[string]reflect.Type{}, docs: docs}
}

func (g *schemaGen) fail(format string, args ...any) {
	g.errs = append(g.errs, fmt.Errorf(format, args...))
}

func ref(name string) map[string]any { return map[string]any{"$ref": "#/$defs/" + name} }

// define adds a $def once; two Go types with one name are an error.
func (g *schemaGen) define(name string, t reflect.Type, build func() map[string]any) map[string]any {
	if prev, ok := g.owner[name]; ok {
		if prev != t {
			g.fail("two types named %s: %v and %v", name, prev, t)
		}
		return ref(name)
	}
	g.owner[name] = t
	g.defs[name] = map[string]any{} // placeholder against recursion
	g.defs[name] = build()
	return ref(name)
}

// enumDef defines a string enum.
func (g *schemaGen) enumDef(name string, t reflect.Type, values []string, desc string) map[string]any {
	return g.define(name, t, func() map[string]any {
		return map[string]any{"type": "string", "enum": values, "description": desc}
	})
}

// treeSchema is the schema of a value of type t encoded by toTree. keepEmpty
// is encodeValue's flag: true inside lists and maps and for the fields of a
// fixedShape type.
func (g *schemaGen) treeSchema(t reflect.Type, keepEmpty bool) map[string]any {
	switch t {
	case durationType:
		return map[string]any{"type": "integer", "minimum": 0, "description": "Milliseconds."}
	case timeType:
		return map[string]any{"type": "integer", "description": "Unix time in milliseconds."}
	case phaseType:
		return g.enumDef("Phase", t, phaseNames(), "Phase of the game.")
	case dawnKindType:
		return g.enumDef("DawnKind", t, []string{dawnKindName(engine.DawnDeaths), dawnKindName(engine.DawnProtected), dawnKindName(engine.DawnNobodyDied)},
			"Which morning announcement was read (rules §4 \"Thông báo buổi sáng\").")
	case soundType:
		return g.enumDef("SoundMode", t, soundModeNames(), "Who reads the announcements aloud (rules §1 \"Âm thanh\").")
	case teamType:
		return g.enumDef("Team", t, teamNames(), "A team (side).")
	case playerIDType:
		return g.define("PlayerID", t, func() map[string]any {
			return map[string]any{"type": "string", "description": "Public id of a player in the room, e.g. \"p3\". Not a secret, not a credential."}
		})
	case roleIDType:
		return g.define("RoleID", t, func() map[string]any {
			return map[string]any{"type": "string", "description": "A role, e.g. \"Werewolf\"; as a night step also \"PhoneCheck\" or \"LoversMeet\"."}
		})
	case linkColourType:
		values := []string{string(engine.LinkWhite), string(engine.LinkBlack), string(engine.LinkStriped)}
		if keepEmpty {
			values = append(values, "")
		}
		return map[string]any{"type": "string", "enum": values}
	}
	switch t.Kind() {
	case reflect.Pointer:
		return g.treeSchema(t.Elem(), keepEmpty)
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": "array", "items": g.treeSchema(t.Elem(), true)}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": g.treeSchema(t.Elem(), true)}
	case reflect.Struct:
		return g.define(t.Name(), t, func() map[string]any { return g.structSchema(t) })
	}
	g.fail("no schema rule for %v (kind %v)", t, t.Kind())
	return map[string]any{}
}

func (g *schemaGen) structSchema(t reflect.Type) map[string]any {
	props := map[string]any{}
	required := []string{}
	fixed := fixedShape[t]
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() || f.Anonymous {
			continue
		}
		key := camel(f.Name)
		s := g.treeSchema(f.Type, fixed)
		if doc := g.docs.field(t, f.Name); doc != "" {
			s = withDescription(s, doc)
		}
		props[key] = s
		if alwaysSent(f.Type, fixed) || alwaysFilled[t.Name()+"."+f.Name] {
			required = append(required, key)
		}
	}
	out := map[string]any{"type": "object", "properties": props, "required": required, "additionalProperties": false}
	if doc := g.docs.typ(t); doc != "" {
		out["description"] = doc
	}
	return out
}

// alwaysSent says whether toTree always writes a value of type t (the
// "keep" result of encodeValue for every value), given keepEmpty.
func alwaysSent(t reflect.Type, keepEmpty bool) bool {
	switch t {
	case durationType, phaseType, dawnKindType, soundType:
		return true
	case timeType, teamType:
		return false
	}
	switch t.Kind() {
	case reflect.Pointer, reflect.Interface:
		return false
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return keepEmpty
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return true
	case reflect.Struct:
		if keepEmpty {
			return true
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.IsExported() && !f.Anonymous && alwaysSent(f.Type, fixedShape[t]) {
				return true
			}
		}
	}
	return false
}

// withDescription adds a description next to s ($ref may have siblings in
// draft 2020-12). s is copied: $refs are shared.
func withDescription(s map[string]any, desc string) map[string]any {
	out := map[string]any{}
	for k, v := range s {
		out[k] = v
	}
	out["description"] = desc
	return out
}

func phaseNames() []string {
	var out []string
	for p := engine.Phase(0); p.String() != "Unknown"; p++ {
		out = append(out, p.String())
	}
	return out
}

func soundModeNames() []string {
	return []string{soundModeName(engine.OneSpeaker), soundModeName(engine.AllPhonesSpeak)}
}

func teamNames() []string {
	var out []string
	for t := engine.TeamNone + 1; t.String() != engine.TeamNone.String(); t++ {
		out = append(out, t.String())
	}
	return out
}
