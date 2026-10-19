package ws

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"masoi/internal/engine"
)

// clientSchema is docs/api/schema/client.schema.json: every WebSocket message
// a client may send.
func clientSchema(docs goDocs, commandDocs map[string]string) (map[string]any, []error) {
	g := newSchemaGen(docs)
	pid := g.treeSchema(playerIDType, false)
	g.defs["Join"] = object(map[string]any{
		"type": constant(typeJoin), "room": roomCodeSchema(), "nickname": str(),
		"hostKey": withDescription(str(), "Only the room's creator, on their own first join: the hostKey from POST /rooms."),
	}, "type", "room", "nickname")
	g.defs["RejoinWithToken"] = object(map[string]any{
		"type": constant(typeRejoin), "room": roomCodeSchema(), "token": withDescription(str(), "The token from welcome (same phone)."),
	}, "type", "room", "token")
	g.defs["RejoinWithPIN"] = object(map[string]any{
		"type": constant(typeRejoin), "room": roomCodeSchema(), "nickname": str(),
		"pin": map[string]any{"type": "string", "pattern": "^[0-9]{4}$"},
	}, "type", "room", "nickname", "pin")
	g.defs["Leave"] = object(map[string]any{"type": constant(typeLeave)}, "type")
	g.defs["Kick"] = object(map[string]any{"type": constant(typeKick), "target": pid}, "type", "target")
	g.defs["TransferHost"] = object(map[string]any{"type": constant(typeTransferHost), "target": pid}, "type", "target")

	args := commandArgs()
	names := make([]string, 0, len(args))
	for n := range args {
		names = append(names, n)
	}
	sort.Strings(names)
	var commands []string
	for _, n := range names {
		argsName := pascal(n) + "Args"
		g.defs[argsName] = g.argsSchema(n, args[n])
		props := map[string]any{"type": constant(typeCommand), "name": constant(n), "args": ref(argsName)}
		required := []string{"type", "name"}
		if args[n].NumField() > 0 {
			required = append(required, "args")
		}
		def := object(props, required...)
		if d := commandDocs[n]; d != "" {
			def["description"] = d
		}
		g.defs[pascal(n)+"Command"] = def
		commands = append(commands, pascal(n)+"Command")
	}
	g.defs["Command"] = oneOfRefs(commands)
	doc := schemaDoc("Ma Sói: client to server", "Every message a client may send on the game WebSocket (/ws). Unknown keys are refused.",
		oneOfRefs([]string{"Join", "RejoinWithToken", "RejoinWithPIN", "Leave", "Kick", "TransferHost", "Command"}), g)
	return doc, g.errs
}

// argsSchema is the schema of a command's args, from the json tags that
// decodeStrict accepts.
func (g *schemaGen) argsSchema(name string, t reflect.Type) map[string]any {
	switch name {
	case "configureSound":
		var modes []string
		for m := range soundModes {
			modes = append(modes, m)
		}
		sort.Strings(modes)
		return object(map[string]any{"mode": map[string]any{"type": "string", "enum": modes}}, "mode")
	case "configureTimers":
		props := map[string]any{}
		tm := reflect.TypeOf(timerMillis{})
		for i := 0; i < tm.NumField(); i++ {
			props[jsonName(tm.Field(i))] = map[string]any{"type": "integer", "minimum": 0, "maximum": maxTimerMillis}
		}
		timers := object(props)
		timers["description"] = "Milliseconds; only the timers to change. Allowed ranges: rules §2 \"Thời gian mặc định\"."
		return object(map[string]any{"timers": timers}, "timers")
	}
	return g.tagSchema(t)
}

// tagSchema is the schema of a struct encoded or decoded by encoding/json:
// keys from the json tags, required unless omitempty.
func (g *schemaGen) tagSchema(t reflect.Type) map[string]any {
	props := map[string]any{}
	var required []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		key := jsonName(f)
		if key == "" {
			continue
		}
		props[key] = g.tagValue(f.Type)
		if !strings.Contains(f.Tag.Get("json"), ",omitempty") {
			required = append(required, key)
		}
	}
	return object(props, required...)
}

func (g *schemaGen) tagValue(t reflect.Type) map[string]any {
	switch {
	case t == playerIDType || t == roleIDType:
		return g.treeSchema(t, false)
	case t.Kind() == reflect.Slice:
		return map[string]any{"type": "array", "items": g.tagValue(t.Elem())}
	case t.Kind() == reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": g.tagValue(t.Elem())}
	case t.Kind() == reflect.Struct:
		return g.define(pascal(strings.TrimSuffix(t.Name(), "JSON")), t, func() map[string]any { return g.tagSchema(t) })
	}
	return g.treeSchema(t, false)
}

func jsonName(f reflect.StructField) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "-" {
		return ""
	}
	return name
}

func pascal(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// commandDocs is the doc comment of the engine command each wire command
// becomes.
func commandDocs(docs goDocs) (map[string]string, error) {
	out := map[string]string{}
	table, err := newCommandTable(engine.Registry)
	if err != nil {
		return nil, err
	}
	for _, c := range goldenCommands {
		build, err := table.build(commandMsg{Type: typeCommand, Name: c.name, Args: json.RawMessage(c.args)})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", c.name, err)
		}
		out[c.name] = docs.typ(reflect.TypeOf(build("p1")))
	}
	return out, nil
}
