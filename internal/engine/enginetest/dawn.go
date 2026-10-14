package enginetest

import "masoi/internal/engine"

// Dawn ends the night and returns what happened at sunrise. Roles that did not
// act time out and do nothing. It fails if no night is running.
func (g *Game) Dawn() *Dawn {
	g.t.Helper()
	if g.stage != stageNight {
		g.t.Fatalf("Dawn() called but no night is running: call Night() first")
	}
	before := g.deadSet()
	mark := len(g.events)
	cmds, err := g.driver.EndNight(g.state)
	g.doAll(cmds, err)
	g.stage = stageDawn

	d := &Dawn{g: g, announced: announcementKeys(g.events[mark:])}
	for _, name := range g.deadSet() {
		if !contains(before, name) {
			d.died = append(d.died, name)
		}
	}
	return d
}

// Dawn is the morning after one night.
type Dawn struct {
	g         *Game
	died      []string
	announced []string
}

// ExpectDead checks exactly who died during the night, in any order.
func (d *Dawn) ExpectDead(names ...string) *Dawn {
	d.g.t.Helper()
	if !sameNames(d.died, names) {
		d.g.t.Fatalf("expected these players to die in the night: %v, but died: %v", names, d.died)
	}
	return d
}

// ExpectNobodyDied checks that the night killed nobody.
func (d *Dawn) ExpectNobodyDied() *Dawn {
	d.g.t.Helper()
	if len(d.died) > 0 {
		d.g.t.Fatalf("expected nobody to die in the night, but died: %v", d.died)
	}
	return d
}

// ExpectAnnouncement checks that the narrator announced the message with the
// given i18n key at dawn.
func (d *Dawn) ExpectAnnouncement(key string) *Dawn {
	d.g.t.Helper()
	if !contains(d.announced, key) {
		d.g.t.Fatalf("expected announcement %q, but the dawn announced: %v", key, d.announced)
	}
	return d
}

// Player returns the ID of a seated player.
func (d *Dawn) Player(name string) engine.PlayerID { return d.g.Player(name) }

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
