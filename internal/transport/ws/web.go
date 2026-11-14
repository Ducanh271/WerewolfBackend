package ws

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

// webApp serves the built web app (web/dist) from Config.WebDir: the page a
// phone opens, its files, and the QR link /j/{code}, which opens the same page
// (the page reads the code from its own address). It is not embedded in the
// binary: web/dist is not in git.
//
// The folder is opened as an os.Root, so neither ".." nor a symlink reaches a
// file outside it, and hidden files (a ".env" next to a wrongly set WEB_DIR)
// are never served.
type webApp struct {
	files fs.FS
}

func newWebApp(dir string) (*webApp, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("WEB_DIR %q: %w", dir, err)
	}
	files := root.FS()
	if st, err := fs.Stat(files, "index.html"); err != nil || st.IsDir() {
		return nil, fmt.Errorf("WEB_DIR %q has no index.html (build it with `cd web && npm run build`)", dir)
	}
	return &webApp{files: files}, nil
}

// pageHeaders go on every file: no sniffing, no framing by another site (the
// role card is on this page), and the room code in /j/{code} does not leave
// in a Referer.
func pageHeaders(h http.Header) {
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Content-Security-Policy", "frame-ancestors 'none'")
	h.Set("Referrer-Policy", "no-referrer")
}

// page serves index.html, never cached: a new build reaches phones at once.
func (a *webApp) page(w http.ResponseWriter, r *http.Request) {
	a.serve(w, r, "index.html")
}

// file serves one file of the build.
func (a *webApp) file(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if p == "" {
		p = "index.html"
	}
	a.serve(w, r, p)
}

func (a *webApp) serve(w http.ResponseWriter, r *http.Request, name string) {
	for _, part := range strings.Split(name, "/") {
		if strings.HasPrefix(part, ".") {
			http.NotFound(w, r)
			return
		}
	}
	f, err := a.files.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	seeker, ok := f.(io.ReadSeeker)
	if err != nil || st.IsDir() || !ok {
		http.NotFound(w, r) // a folder is never listed
		return
	}
	pageHeaders(w.Header())
	// Vite names the files under assets/ by their content, so they may be
	// cached for good; everything else is checked again on each load.
	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	http.ServeContent(w, r, st.Name(), st.ModTime(), seeker)
}
