package ws

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// docs/api/schema and docs/api/examples are generated: `make api-docs` runs
// TestAPIDocs with -update to rewrite them; without it the test compares and
// fails on any difference, so the docs cannot drift from the server.
var updateAPIDocs = flag.Bool("update", false, "rewrite docs/api/schema and docs/api/examples")

const apiDocsDir = "../../../docs/api"

// generatedDirs are the folders of docs/api the test owns entirely.
var generatedDirs = []string{"schema", "examples"}

func TestAPIDocs(t *testing.T) {
	docs, err := readGoDocs()
	if err != nil {
		t.Fatal(err)
	}
	cmdDocs, err := commandDocs(docs)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for name, build := range map[string]func() (map[string]any, []error){
		"server": func() (map[string]any, []error) { return serverSchema(docs) },
		"client": func() (map[string]any, []error) { return clientSchema(docs, cmdDocs) },
		"http":   func() (map[string]any, []error) { return httpSchema(docs) },
	} {
		doc, errs := build()
		for _, err := range errs {
			t.Errorf("%s schema: %v", name, err)
		}
		files["schema/"+name+".schema.json"] = prettyJSON(doc)
	}
	v := newValidator(t, files)

	ex := playExampleGames(t)
	ex.validate(t, v)
	for name, b := range ex.files {
		files["examples/"+name] = b
	}
	syncAPIDocs(t, files)
}

// The events in the source are the events with a schema.
func TestAPIDocs_EveryEventInTheSource(t *testing.T) {
	docs, err := readGoDocs()
	if err != nil {
		t.Fatal(err)
	}
	var listed []string
	for _, ev := range eventTypes {
		tp := reflect.TypeOf(ev)
		listed = append(listed, tp.PkgPath()+"."+tp.Name())
	}
	sort.Strings(listed)
	if !reflect.DeepEqual(listed, docs.events) {
		t.Fatalf("eventTypes (apischema_msgs_test.go) differs from the events in the source:\n listed %v\n source %v", listed, docs.events)
	}
}

func TestAPIDocs_CoreArgsNameEveryCoreCommand(t *testing.T) {
	var got, want []string
	for n := range coreArgs {
		got = append(got, n)
	}
	for n := range coreCommands {
		want = append(want, n)
	}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("coreArgs %v, core commands %v", got, want)
	}
}

// syncAPIDocs writes files under docs/api (-update) or checks the folders
// hold exactly them.
func syncAPIDocs(t *testing.T, files map[string][]byte) {
	t.Helper()
	if *updateAPIDocs {
		for _, d := range generatedDirs {
			if err := os.RemoveAll(filepath.Join(apiDocsDir, d)); err != nil {
				t.Fatal(err)
			}
		}
		for name, b := range files {
			path := filepath.Join(apiDocsDir, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	stale := func(format string, args ...any) {
		t.Helper()
		t.Errorf(format+" — run `make api-docs`", args...)
	}
	onDisk := map[string]bool{}
	for _, d := range generatedDirs {
		_ = filepath.WalkDir(filepath.Join(apiDocsDir, d), func(path string, e fs.DirEntry, err error) error {
			if err == nil && !e.IsDir() {
				rel, _ := filepath.Rel(apiDocsDir, path)
				onDisk[filepath.ToSlash(rel)] = true
			}
			return nil
		})
	}
	for name := range onDisk {
		if _, ok := files[name]; !ok {
			stale("docs/api/%s is no longer generated", name)
		}
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		got, err := os.ReadFile(filepath.Join(apiDocsDir, name))
		switch {
		case err != nil:
			stale("docs/api/%s is missing", name)
		case !bytes.Equal(got, files[name]):
			stale("docs/api/%s is stale", name)
		}
	}
}

// prettyJSON is v as indented JSON, object keys sorted with "$schema",
// "title", "description" and "type" first, ending with a newline.
func prettyJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var tree any
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := dec.Decode(&tree); err != nil {
		panic(err)
	}
	var out bytes.Buffer
	writePretty(&out, tree, "")
	out.WriteByte('\n')
	return out.Bytes()
}

var firstKeys = []string{"$schema", "title", "description", "type"}

func keyRank(k string) int {
	for i, f := range firstKeys {
		if k == f {
			return i
		}
	}
	return len(firstKeys)
}

func writePretty(out *bytes.Buffer, v any, indent string) {
	inner := indent + "  "
	switch x := v.(type) {
	case map[string]any:
		if len(x) == 0 {
			out.WriteString("{}")
			return
		}
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if ri, rj := keyRank(keys[i]), keyRank(keys[j]); ri != rj {
				return ri < rj
			}
			return keys[i] < keys[j]
		})
		out.WriteString("{\n")
		for i, k := range keys {
			kb, _ := json.Marshal(k)
			out.WriteString(inner)
			out.Write(kb)
			out.WriteString(": ")
			writePretty(out, x[k], inner)
			if i < len(keys)-1 {
				out.WriteByte(',')
			}
			out.WriteByte('\n')
		}
		out.WriteString(indent + "}")
	case []any:
		if len(x) == 0 {
			out.WriteString("[]")
			return
		}
		out.WriteString("[\n")
		for i, item := range x {
			out.WriteString(inner)
			writePretty(out, item, inner)
			if i < len(x)-1 {
				out.WriteByte(',')
			}
			out.WriteByte('\n')
		}
		out.WriteString(indent + "]")
	default:
		enc := json.NewEncoder(out)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(x)
		out.Truncate(out.Len() - 1) // Encode's newline
	}
}

// validator checks JSON against the generated schemas, with the real JSON
// Schema 2020-12 rules.
type validator struct {
	t       *testing.T
	c       *jsonschema.Compiler
	schemas map[string]*jsonschema.Schema
}

// newValidator compiles the schemas, after checking each one against the
// draft 2020-12 meta-schema.
func newValidator(t *testing.T, files map[string][]byte) *validator {
	t.Helper()
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	meta, err := c.Compile(schemaDraft)
	if err != nil {
		t.Fatal(err)
	}
	v := &validator{t: t, c: c, schemas: map[string]*jsonschema.Schema{}}
	for name, b := range files {
		if !strings.HasPrefix(name, "schema/") {
			continue
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		if err := meta.Validate(doc); err != nil {
			t.Fatalf("%s is not a valid draft 2020-12 schema: %v", name, err)
		}
		if err := c.AddResource(name, doc); err != nil {
			t.Fatal(err)
		}
	}
	return v
}

// check validates b against loc ("schema/server.schema.json", or with a
// "#/$defs/..." fragment).
func (v *validator) check(what, loc string, b []byte) {
	v.t.Helper()
	sch, ok := v.schemas[loc]
	if !ok {
		var err error
		if sch, err = v.c.Compile(loc); err != nil {
			v.t.Fatalf("compile %s: %v", loc, err)
		}
		v.schemas[loc] = sch
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		v.t.Errorf("%s: %v", what, err)
		return
	}
	if err := sch.Validate(inst); err != nil {
		v.t.Errorf("%s does not match %s: %v\n%s", what, loc, err, b)
	}
}
