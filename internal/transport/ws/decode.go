package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"masoi/internal/engine"
)

// Client message types (docs/protocol.md).
const (
	typeJoin    = "join"
	typeRejoin  = "rejoin"
	typeLeave   = "leave"
	typeCommand = "command"
)

type joinMsg struct {
	Type     string `json:"type"`
	Room     string `json:"room"`
	Nickname string `json:"nickname"`
}

type rejoinMsg struct {
	Type     string `json:"type"`
	Room     string `json:"room"`
	Token    string `json:"token"`
	Nickname string `json:"nickname"`
	PIN      string `json:"pin"`
}

type leaveMsg struct {
	Type string `json:"type"`
}

type commandMsg struct {
	Type string          `json:"type"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

// decodeStrict decodes one JSON object into v. Keys must match a field's json
// tag exactly (encoding/json alone also accepts other capitals), unknown keys
// are refused, and nothing may follow the object.
func decodeStrict(data []byte, v any) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil || keys == nil {
		return badMessage("not a JSON object")
	}
	allowed := jsonKeys(reflect.TypeOf(v).Elem())
	for k := range keys {
		if !allowed[k] {
			return badMessage(fmt.Sprintf("unknown field %q", k))
		}
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return badMessage("wrong field type: " + err.Error())
	}
	if _, err := dec.Token(); err != io.EOF {
		return badMessage("unexpected data after the message")
	}
	return nil
}

func jsonKeys(t reflect.Type) map[string]bool {
	keys := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		if name != "" && name != "-" {
			keys[name] = true
		}
	}
	return keys
}

// messageType reads the type of a client message without looking at the rest.
func messageType(data []byte) (string, error) {
	var head struct {
		Type any `json:"type"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return "", badMessage("not a JSON object")
	}
	typ, ok := head.Type.(string)
	if !ok || typ == "" {
		return "", badMessage("missing message type")
	}
	return typ, nil
}

func unknownType(typ string) error {
	return &protoError{codeUnknownType, fmt.Sprintf("unknown message type %q", typ)}
}

// buildCommand turns a `command` message into a function that builds the
// engine command for the connection's player. The player never comes from the
// message.
func buildCommand(m commandMsg) (func(by engine.PlayerID) engine.Command, error) {
	spec, ok := commandTable[m.Name]
	if !ok {
		return nil, &protoError{codeUnknownCommand, fmt.Sprintf("unknown command %q", m.Name)}
	}
	args := m.Args
	if len(args) == 0 || string(args) == "null" {
		args = json.RawMessage("{}")
	}
	return spec(args)
}
