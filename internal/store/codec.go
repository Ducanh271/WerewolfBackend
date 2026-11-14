package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"masoi/internal/engine"
)

// EncodeEvent turns an event into its stored form: the Go type name and the
// JSON of its fields (docs/database.md "Event encoding"). Only registered
// events can be encoded, so a log can always be read back.
func EncodeEvent(e engine.Event) (name string, data []byte, err error) {
	t := reflect.TypeOf(e)
	if t == nil {
		return "", nil, fmt.Errorf("store: encode nil event")
	}
	if reg, ok := engine.EventType(t.Name()); !ok || reg != t {
		return "", nil, fmt.Errorf("store: event %T is not registered (engine.RegisterEvents)", e)
	}
	data, err = json.Marshal(e)
	if err != nil {
		return "", nil, fmt.Errorf("store: encode %s: %w", t.Name(), err)
	}
	return t.Name(), data, nil
}

// DecodeEvent is the reverse of EncodeEvent. An unknown name, or a field the
// event does not have, is an error: a log is never read half right.
func DecodeEvent(name string, data []byte) (engine.Event, error) {
	t, ok := engine.EventType(name)
	if !ok {
		return nil, fmt.Errorf("store: unknown event type %q", name)
	}
	v := reflect.New(t)
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v.Interface()); err != nil {
		return nil, fmt.Errorf("store: decode %s: %w", name, err)
	}
	return v.Elem().Interface().(engine.Event), nil
}
