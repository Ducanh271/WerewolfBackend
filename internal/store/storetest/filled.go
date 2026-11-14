// Package storetest helps the store tests: every registered event with every
// field filled, so each field shows up in an encoding with its name and kind.
package storetest

import (
	"reflect"
	"time"

	"masoi/internal/engine"
)

// FilledEvents returns one value of every registered event (sorted by name)
// with every exported field set to a fixed non-zero value.
func FilledEvents() []engine.Event {
	var out []engine.Event
	for _, e := range engine.EventTypes() {
		v := reflect.New(reflect.TypeOf(e)).Elem()
		fill(v, 1)
		out = append(out, v.Interface().(engine.Event))
	}
	return out
}

var (
	durationType = reflect.TypeOf(time.Duration(0))
	timeType     = reflect.TypeOf(time.Time{})
)

func fill(v reflect.Value, n int) {
	switch v.Kind() {
	case reflect.Struct:
		if v.Type() == timeType {
			v.Set(reflect.ValueOf(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)))
			return
		}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				fill(v.Field(i), n+i)
			}
		}
	case reflect.String:
		v.SetString("p" + string(rune('0'+n%10)))
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Type() == durationType {
			v.SetInt(int64(time.Duration(n) * time.Second))
			return
		}
		v.SetInt(int64(n))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(n))
	case reflect.Slice:
		s := reflect.MakeSlice(v.Type(), 1, 1)
		fill(s.Index(0), n+1)
		v.Set(s)
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k, e := reflect.New(v.Type().Key()).Elem(), reflect.New(v.Type().Elem()).Elem()
		fill(k, n+1)
		fill(e, n+2)
		m.SetMapIndex(k, e)
		v.Set(m)
	case reflect.Pointer:
		p := reflect.New(v.Type().Elem())
		fill(p.Elem(), n)
		v.Set(p)
	}
}
