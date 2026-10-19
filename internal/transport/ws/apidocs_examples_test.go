package ws

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"

	"masoi/internal/transport/ws/wsclient"
)

// The example games are played through the real server with a fake clock,
// fixed randomness and room code exampleRoom, so every run sends the same
// bytes. An example is a message as a phone received or sent it.

const exampleRoom = "K7QP"

// exampleHostKey replaces the real hostKey (drawn by the manager from
// crypto/rand) in the examples.
const exampleHostKey = "5f1c9e0b7a2d4e6f8a9b0c1d2e3f4a5b"

// exampleRandom makes tokens, PINs and seeds that are the same on every run.
type exampleRandom struct {
	mu                  sync.Mutex
	tokens, pins, seeds int
}

func (r *exampleRandom) Seed() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seeds++
	return int64(r.seeds)
}

func (r *exampleRandom) Token() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens++
	sum := sha256.Sum256([]byte(fmt.Sprint("token ", r.tokens)))
	return hex.EncodeToString(sum[:16])
}

func (r *exampleRandom) PIN() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pins++
	sum := sha256.Sum256([]byte(fmt.Sprint("pin ", r.pins)))
	return fmt.Sprintf("%04d", binary.BigEndian.Uint32(sum[:4])%10000)
}

// exampleSet collects the examples and everything to validate.
type exampleSet struct {
	t       *testing.T
	files   map[string][]byte
	clients []*wsclient.Client // every phone ever connected, in order
	sent    []sentMsg
	http    []sentMsg
	redact  map[string]string
}

type sentMsg struct {
	what, schema string
	body         []byte
}

func newExampleSet(t *testing.T) *exampleSet {
	return &exampleSet{t: t, files: map[string][]byte{}, redact: map[string]string{}}
}

// put stores an example once under name (a path under docs/api/examples).
func (x *exampleSet) put(name string, v any) {
	x.t.Helper()
	if _, dup := x.files[name]; dup {
		x.t.Fatalf("example %s recorded twice", name)
	}
	x.files[name] = x.clean(prettyJSON(v))
}

// clean replaces the secrets that are not drawn from exampleRandom.
func (x *exampleSet) clean(b []byte) []byte {
	keys := make([]string, 0, len(x.redact))
	for k := range x.redact {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b = bytes.ReplaceAll(b, []byte(k), []byte(x.redact[k]))
	}
	return b
}

// validate checks every message of the games against the schemas, and adds
// one example per event type from what the phones received.
func (x *exampleSet) validate(t *testing.T, v *validator) {
	t.Helper()
	for _, m := range x.sent {
		v.check(m.what, m.schema, m.body)
	}
	for _, m := range x.http {
		v.check(m.what, m.schema, m.body)
	}
	// Phones get the same messages and events many times over: each distinct
	// one is validated once, an events message as its envelope plus each of
	// its events.
	seen, checked := map[string]bool{}, map[string]bool{}
	check := func(what, loc string, msg any) {
		b, _ := json.Marshal(msg)
		key := loc + string(b)
		if !checked[key] {
			checked[key] = true
			v.check(what, loc, b)
		}
	}
	for i, c := range x.clients {
		for j, f := range c.Snapshot().Frames {
			what := fmt.Sprintf("phone %d frame %d (%s)", i, j, f.Type())
			if f.Type() != "events" {
				check(what, "schema/server.schema.json", f)
				continue
			}
			list, _ := f["events"].([]any)
			envelope := wsclient.Frame{}
			for k, val := range f {
				envelope[k] = val
			}
			envelope["events"] = []any{}
			check(what, "schema/server.schema.json", envelope)
			for _, ev := range list {
				check(what, "schema/server.schema.json#/$defs/Event", ev)
				typ, _ := ev.(map[string]any)["type"].(string)
				if !seen[typ] {
					seen[typ] = true
					x.put("events/"+typ+".json", ev)
				}
			}
		}
	}
	for _, ev := range eventTypes {
		name := fmt.Sprintf("%T", ev)
		name = name[strings.LastIndex(name, ".")+1:]
		if !seen[name] {
			t.Errorf("no example game sends the event %s", name)
		}
	}
	for name, b := range x.files {
		switch {
		case strings.HasPrefix(name, "events/"):
			v.check(name, "schema/server.schema.json#/$defs/Event", b)
		case strings.HasPrefix(name, "server/"):
			v.check(name, "schema/server.schema.json", b)
		case strings.HasPrefix(name, "client/"):
			v.check(name, "schema/client.schema.json", b)
		}
	}
	for name := range commandArgs() {
		if _, ok := x.files["client/command."+name+".json"]; !ok {
			t.Errorf("no example game sends the command %s", name)
		}
	}
}
