package engine

import (
	"testing"

	"masoi/internal/i18n"
)

// This file proves each new voice line of docs/voice-lines.md (audit [L3]) is
// emitted at the moment the table says, tied to its i18n key. The text itself
// is checked in internal/i18n (vi_test.go); here only the event and the
// moment matter.

// TestDealRoles_EmitsRoleViewStarted: rules §2 bước 5, right after the deal.
func TestDealRoles_EmitsRoleViewStarted(t *testing.T) {
	s := lobbyOf(t, 7)
	s = apply(t, s, SetSeats{Order: seatOrder(7)})
	s = apply(t, s, ConfigureRoles{Roles: map[RoleID]int{roleWerewolf: 2, roleSeer: 1, roleVillager: 4}})

	evs, err := Decide(s, DealRoles{Seed: 1})
	if err != nil {
		t.Fatalf("DealRoles rejected: %v", err)
	}
	if len(evs) != 2 {
		t.Fatalf("events %v, want RolesDealt then RoleViewStarted", evs)
	}
	rv, ok := evs[1].(RoleViewStarted)
	if !ok || rv.Duration != s.Settings.RoleView {
		t.Fatalf("event %#v, want RoleViewStarted{%v}", evs[1], s.Settings.RoleView)
	}
	if rv.AnnouncementKey() != i18n.RoleViewStarted {
		t.Fatalf("key = %q, want %q", rv.AnnouncementKey(), i18n.RoleViewStarted)
	}
}

// TestRoleView_TimeoutEmitsChiefElectionStarted: rules §7b, right after the
// role-view countdown.
func TestRoleView_TimeoutEmitsChiefElectionStarted(t *testing.T) {
	s := smallTable(t)
	evs, err := Decide(s, PhaseTimeout{Phase: s.Phase, Seq: s.Seq})
	if err != nil {
		t.Fatalf("PhaseTimeout rejected: %v", err)
	}
	ces, ok := evs[0].(ChiefElectionStarted)
	if !ok || ces.Duration != s.Settings.ChiefElection {
		t.Fatalf("event %#v, want ChiefElectionStarted{%v}", evs[0], s.Settings.ChiefElection)
	}
	if ces.AnnouncementKey() != i18n.ChiefElectionStarted {
		t.Fatalf("key = %q, want %q", ces.AnnouncementKey(), i18n.ChiefElectionStarted)
	}
}

// TestChiefElection_TimeoutEmitsNightStarted: rules §3, night 1 opens right
// after the election.
func TestChiefElection_TimeoutEmitsNightStarted(t *testing.T) {
	s, _ := timeout(t, smallTable(t)) // -> ChiefElection
	evs, err := Decide(s, PhaseTimeout{Phase: s.Phase, Seq: s.Seq})
	if err != nil {
		t.Fatalf("PhaseTimeout rejected: %v", err)
	}
	// evs[0] is the election result (ChiefElected); the night it opens follows
	// right after.
	ns, ok := evs[1].(NightStarted)
	if !ok || ns.Night != 1 {
		t.Fatalf("event %#v, want NightStarted{Night: 1}", evs[1])
	}
	if ns.AnnouncementKey() != i18n.NightStarted {
		t.Fatalf("key = %q, want %q", ns.AnnouncementKey(), i18n.NightStarted)
	}
}

// bitingWolves sets every living wolf's vote to target, as if the pack had
// just agreed on a bite (rules §9 step 3), so the night resolves with a
// death.
func bitingWolves(s State, target PlayerID) State {
	s.Pending.WolfVotes = map[PlayerID]PlayerID{}
	for id, p := range s.Players {
		if p.Alive && p.Team == TeamWolves {
			s.Pending.WolfVotes[id] = target
		}
	}
	return s
}

// TestNight_DawnStartedComesRightAfterNightEndedAndBeforeTheDeathIsAnnounced:
// rules §4, "Trời đã sáng" opens the morning before the narrator says who
// died.
func TestNight_DawnStartedComesRightAfterNightEndedAndBeforeTheDeathIsAnnounced(t *testing.T) {
	s, _ := timeout(t, smallTable(t)) // -> ChiefElection
	s, _ = timeout(t, s)              // -> Night 1

	var target PlayerID
	for id, p := range s.Players {
		if p.Team != TeamWolves {
			target = id
			break
		}
	}
	s = bitingWolves(s, target)

	var all []Event
	for s.Phase == PhaseNight {
		var evs []Event
		s, evs = timeout(t, s)
		all = append(all, evs...)
	}

	endIdx := -1
	for i, e := range all {
		if _, ok := e.(NightEnded); ok {
			endIdx = i
			break
		}
	}
	if endIdx == -1 || endIdx+1 >= len(all) {
		t.Fatalf("no NightEnded followed by another event: %v", all)
	}
	started, ok := all[endIdx+1].(DawnStarted)
	if !ok {
		t.Fatalf("event after NightEnded = %#v, want DawnStarted", all[endIdx+1])
	}
	if started.AnnouncementKey() != i18n.DawnStarted {
		t.Fatalf("key = %q, want %q", started.AnnouncementKey(), i18n.DawnStarted)
	}
	for _, e := range all[:endIdx+2] {
		if _, ok := e.(DawnAnnounced); ok {
			t.Fatalf("DawnAnnounced came at or before DawnStarted: %v", all)
		}
		if _, ok := e.(PlayerDied); ok {
			t.Fatalf("PlayerDied came at or before DawnStarted: %v", all)
		}
	}
	var sawDeath bool
	for _, e := range all[endIdx+2:] {
		if d, ok := e.(PlayerDied); ok && d.Player == target {
			sawDeath = true
		}
	}
	if !sawDeath {
		t.Fatalf("the bitten player never died: %v", all)
	}
}

// TestDawn_TimeoutEmitsDayStarted: rules §4, discussion opens right after the
// morning announcement.
func TestDawn_TimeoutEmitsDayStarted(t *testing.T) {
	s, _ := timeout(t, smallTable(t)) // -> ChiefElection
	s, _ = timeout(t, s)              // -> Night 1
	for s.Phase == PhaseNight {
		s, _ = timeout(t, s)
	}
	if s.Phase != PhaseDawn {
		t.Fatalf("phase = %s, want Dawn", s.Phase)
	}
	evs, err := Decide(s, PhaseTimeout{Phase: s.Phase, Seq: s.Seq})
	if err != nil {
		t.Fatalf("PhaseTimeout rejected: %v", err)
	}
	ds, ok := evs[0].(DayStarted)
	if !ok || ds.Duration != s.Settings.Day {
		t.Fatalf("event %#v, want DayStarted{%v}", evs[0], s.Settings.Day)
	}
	if ds.AnnouncementKey() != i18n.DayStarted {
		t.Fatalf("key = %q, want %q", ds.AnnouncementKey(), i18n.DayStarted)
	}
}

// TestNominationLock_TimeoutEmitsTrialStarted: rules §4 bước 2, the top
// nominee goes up once the countdown runs out.
func TestNominationLock_TimeoutEmitsTrialStarted(t *testing.T) {
	s := handBuilt(nil, alive("a", V), alive("b", V), alive("c", V))
	s.Phase = PhaseNominationLock
	s.Day.Nominations = map[PlayerID]Nomination{"a": {Target: "b", Order: 1}}
	s.Settings = DefaultSettings()

	evs, err := Decide(s, PhaseTimeout{Phase: s.Phase, Seq: s.Seq})
	if err != nil {
		t.Fatalf("PhaseTimeout rejected: %v", err)
	}
	ts, ok := evs[0].(TrialStarted)
	if !ok || ts.Defendant != "b" || ts.Duration != s.Settings.Defense {
		t.Fatalf("event %#v, want TrialStarted{Defendant: b, Duration: %v}", evs[0], s.Settings.Defense)
	}
	if ts.AnnouncementKey() != i18n.TrialStarted {
		t.Fatalf("key = %q, want %q", ts.AnnouncementKey(), i18n.TrialStarted)
	}
}

// TestTrial_TimeoutEmitsVerdictStarted: rules §4 bước 3→4, the kill-or-spare
// vote opens once the defense is over.
func TestTrial_TimeoutEmitsVerdictStarted(t *testing.T) {
	s := handBuilt(nil, alive("a", V), alive("b", V), alive("c", V))
	s.Phase = PhaseTrial
	s.Day.Defendant = "b"
	s.Settings = DefaultSettings()

	evs, err := Decide(s, PhaseTimeout{Phase: s.Phase, Seq: s.Seq})
	if err != nil {
		t.Fatalf("PhaseTimeout rejected: %v", err)
	}
	vs, ok := evs[0].(VerdictStarted)
	if !ok || vs.Duration != s.Settings.Verdict {
		t.Fatalf("event %#v, want VerdictStarted{%v}", evs[0], s.Settings.Verdict)
	}
	if vs.AnnouncementKey() != i18n.VerdictStarted {
		t.Fatalf("key = %q, want %q", vs.AnnouncementKey(), i18n.VerdictStarted)
	}
}

// TestHunterShot_TimeoutEmitsHunterTimedOut: rules §6, 15 seconds pass with no
// choice.
func TestHunterShot_TimeoutEmitsHunterTimedOut(t *testing.T) {
	s := handBuilt(nil, alive("h", V), alive("w", W), alive("a", V), alive("b", V))
	s.Phase = PhaseHunterShot
	s.Shot = Shot{Hunter: "h", Resume: PhaseDay}
	s.Settings = DefaultSettings()

	evs, err := Decide(s, PhaseTimeout{Phase: s.Phase, Seq: s.Seq})
	if err != nil {
		t.Fatalf("PhaseTimeout rejected: %v", err)
	}
	hto, ok := evs[0].(HunterTimedOut)
	if !ok || hto.Hunter != "h" {
		t.Fatalf("event %#v, want HunterTimedOut{Hunter: h}", evs[0])
	}
	if hto.AnnouncementKey() != i18n.HunterTimedOut {
		t.Fatalf("key = %q, want %q", hto.AnnouncementKey(), i18n.HunterTimedOut)
	}
}
