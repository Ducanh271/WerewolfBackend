package roles_test

import (
	"strings"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/i18n"
)

// guideFixture is some GuideTimers, not the defaults, so a test asserting on
// these numbers proves the guide actually renders the values it was given
// instead of a number that happens to also be the default.
func guideFixture() i18n.GuideTimers {
	return i18n.GuideTimers{
		RoleView:            22 * time.Second,
		HunterShot:          37 * time.Second,
		LoversMeet:          14 * time.Second,
		ChiefSuccession:     19 * time.Second,
		PINMaxFailures:      4,
		PINLockTime:         3 * time.Minute,
		CupidMinPlayers:     11,
		DetectiveMinPlayers: 13,
	}
}

func entryByID(entries []i18n.GuideEntry, id string) (i18n.GuideEntry, bool) {
	for _, e := range entries {
		if e.ID == id {
			return e, true
		}
	}
	return i18n.GuideEntry{}, false
}

// T25 step 3b: every role in the Registry has a guide entry; adding a role
// without adding its guide text fails this test.
func TestGuide_EveryRegisteredRoleHasAnEntry(t *testing.T) {
	entries := i18n.RoleGuides(guideFixture())
	for id := range engine.Registry {
		e, ok := entryByID(entries, string(id))
		if !ok {
			t.Errorf("%s has no guide entry", id)
			continue
		}
		if e.Title == "" || e.WhenCalled == "" || e.HowTo == "" || e.OnTimeout == "" || e.Gotcha == "" {
			t.Errorf("%s has an incomplete guide entry %+v", id, e)
		}
	}
	if len(entries) != len(engine.Registry) {
		t.Errorf("RoleGuides has %d entries, Registry has %d roles", len(entries), len(engine.Registry))
	}
}

// The three statuses that are not a role (Chief, Lovers, Cursed) each have a
// full entry too (merged from T28b).
func TestGuide_StatusEntriesComplete(t *testing.T) {
	entries := i18n.StatusGuides(guideFixture())
	for _, id := range []string{i18n.StatusChief, i18n.StatusLovers, i18n.StatusCursed} {
		e, ok := entryByID(entries, id)
		if !ok {
			t.Fatalf("%s has no guide entry", id)
		}
		if e.Title == "" || e.WhenCalled == "" || e.HowTo == "" || e.OnTimeout == "" || e.Gotcha == "" {
			t.Errorf("%s has an incomplete guide entry %+v", id, e)
		}
	}
}

// Numbers quoted in the guide must come from the GuideTimers given, not a
// hardcoded default: render with values that differ from DefaultSettings and
// check each number lands in its own sentence, nowhere else.
func TestGuide_NumbersComeFromTheGivenTimers(t *testing.T) {
	tm := guideFixture()
	roles := i18n.RoleGuides(tm)
	statuses := i18n.StatusGuides(tm)
	checklist := i18n.FirstTimeChecklist(tm)

	hunter, _ := entryByID(roles, "Hunter")
	if !strings.Contains(hunter.HowTo, "37 giây") {
		t.Errorf("Hunter.HowTo = %q, want it to mention 37 giây", hunter.HowTo)
	}

	detective, _ := entryByID(roles, "Detective")
	if !strings.Contains(detective.Title, "13 người") {
		t.Errorf("Detective.Title = %q, want it to mention 13 người", detective.Title)
	}

	cupid, _ := entryByID(roles, "Cupid")
	if !strings.Contains(cupid.Title, "11 người") {
		t.Errorf("Cupid.Title = %q, want it to mention 11 người", cupid.Title)
	}

	chief, _ := entryByID(statuses, i18n.StatusChief)
	if !strings.Contains(chief.HowTo, "19 giây") {
		t.Errorf("Chief.HowTo = %q, want it to mention 19 giây", chief.HowTo)
	}

	lovers, _ := entryByID(statuses, i18n.StatusLovers)
	if !strings.Contains(lovers.WhenCalled, "14 giây") {
		t.Errorf("Lovers.WhenCalled = %q, want it to mention 14 giây", lovers.WhenCalled)
	}

	if !strings.Contains(checklist[0], "22 giây") {
		t.Errorf("checklist[0] = %q, want it to mention 22 giây", checklist[0])
	}
	pinLine := checklist[len(checklist)-2]
	if !strings.Contains(pinLine, "4 lần") || !strings.Contains(pinLine, "3 phút") {
		t.Errorf("pin checklist line = %q, want 4 lần and 3 phút", pinLine)
	}
}

// §1 and §4 carry no settings number, so they stay fixed regardless of
// GuideTimers; §1 is exactly the approved 12 lines.
func TestGuide_QuickRulesAndTipsAreFixed(t *testing.T) {
	if n := len(i18n.QuickRules()); n != 12 {
		t.Errorf("QuickRules has %d lines, want 12", n)
	}
	if n := len(i18n.OptionalTips()); n == 0 {
		t.Error("OptionalTips is empty")
	}
}
