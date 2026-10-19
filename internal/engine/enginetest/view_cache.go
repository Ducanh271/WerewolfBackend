package enginetest

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"masoi/internal/engine"
)

// Outsider is a viewer who never joined the room; the differential check
// looks through their phone too.
const Outsider engine.PlayerID = "outsider"

// ViewCacheDiff compares the fast view path with the reference one for one
// viewer (audit R5): c.ViewFor must equal engine.ViewFor and c.Visible must
// equal engine.Visible, deep-equal. It returns "" when they agree, else what
// differs.
func ViewCacheDiff(c *engine.ViewCache, s engine.State, log []engine.Event, v engine.Viewer) string {
	if got, want := c.ViewFor(s, log, v), engine.ViewFor(s, log, v); !reflect.DeepEqual(got, want) {
		return fmt.Sprintf("view of %q after %d events:\n fast %+v\n  ref %+v", v.Player, len(log), got, want)
	}
	if got, want := c.Visible(s, log, v), engine.Visible(s, log, v); !reflect.DeepEqual(got, want) {
		return fmt.Sprintf("events of %q after %d events: fast has %d, ref %d\n fast %+v\n  ref %+v", v.Player, len(log), len(got), len(want), got, want)
	}
	return ""
}

// Viewers is everybody who ever joined the log, in a stable order, plus the
// Outsider.
func Viewers(log []engine.Event) []engine.Viewer {
	ids := map[engine.PlayerID]bool{Outsider: true}
	for _, e := range log {
		if j, ok := e.(engine.PlayerJoined); ok {
			ids[j.Player] = true
		}
	}
	out := make([]engine.Viewer, 0, len(ids))
	for id := range ids {
		out = append(out, engine.PlayerViewer(id))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Player < out[j].Player })
	return out
}

// checkViewCache runs ViewCacheDiff for every viewer; record calls it after
// every event, so each scenario is also a differential test of the cache.
// The full replay costs one replay per viewer per event, so under -short
// (make test-fast) only one viewer is checked per event, each in turn.
func (g *Game) checkViewCache() {
	g.t.Helper()
	viewers := Viewers(g.events)
	if testing.Short() && !g.allViewers {
		viewers = viewers[len(g.events)%len(viewers):][:1]
	}
	for _, v := range viewers {
		if diff := ViewCacheDiff(&g.views, g.state, g.events, v); diff != "" {
			g.t.Fatalf("view cache differs from the full replay: %s", diff)
		}
	}
}

// CheckEveryViewer makes every event of the scenario check every viewer
// against the full replay, under -short too.
func CheckEveryViewer() Option {
	return func(s *setup) { s.allViewers = true }
}
