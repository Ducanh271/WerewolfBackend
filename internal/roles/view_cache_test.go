package roles_test

import (
	"testing"

	"masoi/internal/engine"
	"masoi/internal/roles"

	. "masoi/internal/engine/enginetest"
)

// These scenarios are the differential test of the fast view path (audit
// R5): after every event, every phone's view and event list from
// engine.ViewCache must equal the full replay of engine.ViewFor and
// engine.Visible (CheckEveryViewer, see enginetest.ViewCacheDiff). Each one
// covers a moment where what a phone may see changes after the fact.

func TestViewCache_BittenHunterChiefShootsThenNamesASuccessor_EveryViewMatchesTheFullReplay(t *testing.T) {
	g := hunterTable(t, Chief("Phuc"), CheckEveryViewer())
	bitHunter(g)
	g.Dawn().ExpectDead("Phuc")
	g.As("Phuc").Full()
	g.Hunter("Phuc").Shoots("Hoa")
	g.ExpectDead("Phuc", "Hoa").ExpectPhase(engine.PhaseChiefSuccession)
	g.As("Hoa").Full()
	g.NameSuccessor("Phuc", "Em")
	g.ExpectChief("Em").ExpectPhase(engine.PhaseDawn)
	g.Day().Hanged("An")
}

func TestViewCache_HangedHunterChiefShootsThenNamesASuccessor_EveryViewMatchesTheFullReplay(t *testing.T) {
	g := hunterTable(t, Chief("Phuc"), CheckEveryViewer())
	g.Night()
	g.Dawn().ExpectNobodyDied()
	g.Day().Hanged("Phuc")
	g.Hunter("Phuc").Shoots("An")
	g.ExpectDead("Phuc", "An").ExpectPhase(engine.PhaseChiefSuccession)
	g.NameSuccessor("Phuc", "Lan")
	g.ExpectChief("Lan")
	g.Night()
}

// A player cursed on night 2 (the cursedTable cast) sees the pack's votes from then on, not those of
// night 1 (view_audience.go, Visible): judged when each vote was cast, not now.
func TestViewCache_CursedOnNightTwo_SeesOnlyThePackVotesFromThen(t *testing.T) {
	g := NewGame(t, Seats(all...), CheckEveryViewer(),
		Cast("An", CursedWolf, "Binh", Werewolf, "Chi", Villager, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Villager, "Khoa", Villager, "Lan", Villager))
	g.Night().Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Hoa")
	g.Dawn().ExpectDead("Hoa")
	g.Day().NobodyHanged()
	n := g.Night().CursedWolf("An").Curses("Phuc")
	n.PhoneCheck()
	n.Wolf("An").Bites("Em").Wolf("Phuc").Bites("Em")
	g.Dawn().ExpectDead("Em")

	votes := 0
	for _, e := range g.As("Phuc").Feed() {
		if _, ok := e.(roles.WolfVoted); ok {
			votes++
		}
	}
	if votes != 2 {
		t.Fatalf("the cursed player sees %d pack votes, want the 2 of night 2", votes)
	}
}

func TestViewCache_GameEnds_EverybodySeesTheWholeLogAndTheSummary(t *testing.T) {
	g := NewGame(t, append(tenPlayers(), Chief("Lan"), CheckEveryViewer())...)
	bitesOnTheseNights(g, "Phuc", "Giang", "Hoa", "Khoa")
	g.Dawn().ExpectDead("Khoa")
	g.ExpectPhase(engine.PhaseEnded)
	for _, name := range all {
		if s := g.As(name).Full(); len(s.View().Summary) == 0 {
			t.Fatalf("%s sees no summary", name)
		}
	}
}
