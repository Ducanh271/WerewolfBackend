package ws

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// sourceDirs are the packages whose types reach the wire, relative to this
// package. Their doc comments become the schema's descriptions.
var sourceDirs = map[string]string{
	"masoi/internal/engine": "../../engine",
	"masoi/internal/roles":  "../../roles",
	"masoi/internal/room":   "../../room",
}

// goDocs holds the doc comments of struct types and their fields, keyed by
// "package path.Type" and "package path.Type.Field", and the events found in
// the source (struct types embedding EventMarker).
type goDocs struct {
	types  map[string]string
	events []string // "package path.Type", sorted
}

func (d goDocs) typ(t reflect.Type) string { return d.types[t.PkgPath()+"."+t.Name()] }

func (d goDocs) field(t reflect.Type, name string) string {
	return d.types[t.PkgPath()+"."+t.Name()+"."+name]
}

// readGoDocs parses the non-test Go files of sourceDirs.
func readGoDocs() (goDocs, error) {
	d := goDocs{types: map[string]string{}}
	for pkg, dir := range sourceDirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			return d, err
		}
		for _, path := range files {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return d, err
			}
			f, err := parser.ParseFile(token.NewFileSet(), path, src, parser.ParseComments)
			if err != nil {
				return d, err
			}
			d.addFile(pkg, f)
		}
	}
	sort.Strings(d.events)
	return d, nil
}

func (d *goDocs) addFile(pkg string, f *ast.File) {
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts := spec.(*ast.TypeSpec)
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			key := pkg + "." + ts.Name.Name
			doc := ts.Doc
			if doc == nil && len(gen.Specs) == 1 {
				doc = gen.Doc
			}
			d.types[key] = commentText(doc)
			for _, field := range st.Fields.List {
				if len(field.Names) == 0 {
					if embedsEventMarker(field.Type) {
						d.events = append(d.events, key)
					}
					continue
				}
				text := commentText(field.Doc)
				if text == "" {
					text = commentText(field.Comment)
				}
				for _, n := range field.Names {
					d.types[key+"."+n.Name] = text
				}
			}
		}
	}
}

func embedsEventMarker(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name == "EventMarker"
	case *ast.SelectorExpr:
		return x.Sel.Name == "EventMarker"
	}
	return false
}

// commentText is a comment as one line.
func commentText(g *ast.CommentGroup) string {
	if g == nil {
		return ""
	}
	return strings.Join(strings.Fields(g.Text()), " ")
}
