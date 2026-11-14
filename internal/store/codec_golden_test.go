package store_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	_ "masoi/internal/roles" // registers every role's events
	"masoi/internal/store"
	"masoi/internal/store/storetest"
)

var update = flag.Bool("update", false, "rewrite testdata/events.golden.json")

// The stored format is a contract with every log already written
// (docs/database.md "Event encoding"). Each registered event, with every field
// filled, is encoded and compared with the golden file: renaming a field or an
// event, or reordering an enum, fails here. Such a change needs a migration
// that rewrites old rows; only then run
//
//	go test ./internal/store -run TestCodec_Golden -update
//
// A new event or field is added to the golden file the same way.
func TestCodec_GoldenEncodingOfEveryEvent(t *testing.T) {
	got := map[string]json.RawMessage{}
	for _, filled := range storetest.FilledEvents() {
		name, data, err := store.EncodeEvent(filled)
		if err != nil {
			t.Fatalf("encode %T: %v", filled, err)
		}
		back, err := store.DecodeEvent(name, data)
		if err != nil {
			t.Fatalf("decode %s %s: %v", name, data, err)
		}
		if !reflect.DeepEqual(filled, back) {
			t.Errorf("%s does not round-trip:\n before %#v\n after  %#v", name, filled, back)
		}
		got[name] = data
	}
	out, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	out = append(out, '\n')
	path := filepath.Join("testdata", "events.golden.json")
	if *update {
		if err := os.WriteFile(path, out, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (create it with -update)", err)
	}
	if !bytes.Equal(want, out) {
		t.Errorf("the stored event format changed (%s); old logs would no longer read back.\n"+
			"Write a migration first, then -update. Got:\n%s", path, out)
	}
}
