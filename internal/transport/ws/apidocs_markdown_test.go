package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// The hand-written Markdown of docs/api holds generated regions, which
// TestAPIDocs rewrites (-update) or compares like the generated folders:
//
//	<!-- example: server/view.night.witch.json view.phase view.me.turn -->
//	...
//	<!-- /example -->
//
// is that example file as a JSON block, cut down to the listed paths when
// there are any (dot separated, from the top of the file; a last segment
// "people[:2]" keeps the first two items of that list), and
//
//	<!-- generated: errors -->
//	...
//	<!-- /generated -->
//
// is the error tables (apidocs_errors_test.go).
var mdRegion = regexp.MustCompile(`(?s)(<!-- (example|generated): ([^\n]*?) -->\n)(.*?)(<!-- /(example|generated) -->)`)

// syncMarkdown fills the regions of every .md file under docs/api from files
// (the generated folders, as built by this run).
func syncMarkdown(t *testing.T, files map[string][]byte) {
	t.Helper()
	var docs []string
	_ = filepath.WalkDir(apiDocsDir, func(path string, e fs.DirEntry, err error) error {
		if err == nil && !e.IsDir() && strings.HasSuffix(path, ".md") {
			docs = append(docs, path)
		}
		return nil
	})
	sort.Strings(docs)
	for _, path := range docs {
		old, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(apiDocsDir, path)
		got := fillRegions(t, rel, old, files)
		switch {
		case bytes.Equal(got, old):
		case *updateAPIDocs:
			if err := os.WriteFile(path, got, 0o644); err != nil {
				t.Fatal(err)
			}
		default:
			t.Errorf("docs/api/%s is stale — run `make api-docs`", rel)
		}
	}
}

func fillRegions(t *testing.T, doc string, src []byte, files map[string][]byte) []byte {
	t.Helper()
	return mdRegion.ReplaceAllFunc(src, func(m []byte) []byte {
		g := mdRegion.FindSubmatch(m)
		kind, arg, closer := string(g[2]), strings.Fields(string(g[3])), string(g[6])
		if closer != kind || len(arg) == 0 {
			t.Errorf("docs/api/%s: bad region %q", doc, g[1])
			return m
		}
		var body string
		var err error
		switch kind {
		case "example":
			body, err = exampleBlock(files, arg[0], arg[1:])
		case "generated":
			body, err = generatedBlock(arg[0])
		}
		if err != nil {
			t.Errorf("docs/api/%s: %v", doc, err)
			return m
		}
		return []byte(string(g[1]) + body + string(g[5]))
	})
}

func generatedBlock(name string) (string, error) {
	switch name {
	case "errors":
		return errorTables(), nil
	}
	return "", fmt.Errorf("no generated block %q", name)
}

// exampleBlock is the example as a ```json block, cut down to paths.
func exampleBlock(files map[string][]byte, name string, paths []string) (string, error) {
	b, ok := files["examples/"+name]
	if !ok {
		return "", fmt.Errorf("no example %s", name)
	}
	if len(paths) > 0 {
		var full map[string]any
		if err := json.Unmarshal(b, &full); err != nil {
			return "", fmt.Errorf("%s: %v", name, err)
		}
		cut := map[string]any{}
		for _, p := range paths {
			if err := copyPath(cut, full, strings.Split(p, ".")); err != nil {
				return "", fmt.Errorf("%s: %s: %v", name, p, err)
			}
		}
		b = prettyJSON(cut)
	}
	return "```json\n" + string(b) + "```\n", nil
}

var firstItems = regexp.MustCompile(`^(.+)\[:(\d+)\]$`)

// copyPath copies from[path] into to, creating the objects on the way.
func copyPath(to, from map[string]any, path []string) error {
	if len(path) == 1 {
		if m := firstItems.FindStringSubmatch(path[0]); m != nil {
			list, ok := from[m[1]].([]any)
			if !ok {
				return fmt.Errorf("%q is not a list", m[1])
			}
			n, _ := strconv.Atoi(m[2])
			to[m[1]] = list[:min(n, len(list))]
			return nil
		}
	}
	v, ok := from[path[0]]
	if !ok {
		return fmt.Errorf("no field %q", path[0])
	}
	if len(path) == 1 {
		to[path[0]] = v
		return nil
	}
	sub, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("%q is not an object", path[0])
	}
	next, ok := to[path[0]].(map[string]any)
	if !ok {
		next = map[string]any{}
		to[path[0]] = next
	}
	return copyPath(next, sub, path[1:])
}

var (
	mdLink    = regexp.MustCompile(`\]\(([^)\s#]*\.md)?(#[^)\s]*)?\)`)
	mdHeading = regexp.MustCompile(`(?m)^#{1,6} (.+)$`)
)

// Every relative link in docs/api leads to a file that exists, and an
// anchor to one of its headings (GitHub's slugs).
func TestAPIDocs_MarkdownLinks(t *testing.T) {
	_ = filepath.WalkDir(apiDocsDir, func(path string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range mdLink.FindAllStringSubmatch(string(src), -1) {
			file, anchor := m[1], strings.TrimPrefix(m[2], "#")
			if file == "" && anchor == "" {
				continue
			}
			target := path
			if file != "" {
				target = filepath.Join(filepath.Dir(path), file)
			}
			b, err := os.ReadFile(target)
			if err != nil {
				t.Errorf("%s: link to missing %s", path, file)
				continue
			}
			if anchor != "" && !hasHeading(b, anchor) {
				t.Errorf("%s: no heading #%s in %s", path, anchor, target)
			}
		}
		return nil
	})
}

func hasHeading(md []byte, anchor string) bool {
	for _, h := range mdHeading.FindAllStringSubmatch(string(md), -1) {
		if slug(h[1]) == anchor {
			return true
		}
	}
	return false
}

// slug is GitHub's heading anchor: lower case, punctuation dropped, spaces
// to dashes.
func slug(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case r == ' ':
			b.WriteRune('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r):
			b.WriteRune(r)
		}
	}
	return b.String()
}
