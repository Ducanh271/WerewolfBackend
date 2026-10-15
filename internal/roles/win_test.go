package roles_test

import (
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"

	. "masoi/internal/engine/enginetest"
)

// bitesOnTheseNights plays one night per victim (the wolves An and Binh bite
// them) and a quiet day after each, except after the last night.
func bitesOnTheseNights(g *Game, victims ...string) {
	for i, victim := range victims {
		g.Night().Wolf("An").Bites(victim).Wolf("Binh").Bites(victim)
		if i < len(victims)-1 {
			g.Dawn().ExpectDead(victim)
			g.Day().NobodyHanged()
		}
	}
}

func TestWin_WolvesReachVillagersByBiting_GameEndsAtDawn(t *testing.T) {
	g := NewGame(t, append(tenPlayers(), Chief("Lan"))...)
	bitesOnTheseNights(g, "Phuc", "Giang", "Hoa", "Khoa")
	g.Dawn().ExpectDead("Khoa").ExpectAnnouncement(i18n.WinWolves)
	g.ExpectPhase(engine.PhaseEnded)
	if w := g.State().Winner; w == nil || *w != engine.TeamWolves {
		t.Fatalf("winner %v, want wolves", w)
	}
}

func TestWin_WolvesWinWithTheChiefDead_NoSuccession(t *testing.T) {
	g := NewGame(t, append(tenPlayers(), Chief("Khoa"))...)
	bitesOnTheseNights(g, "Phuc", "Giang", "Hoa", "Khoa")
	g.Dawn().ExpectDead("Khoa")
	g.ExpectPhase(engine.PhaseEnded).ExpectChief("Khoa") // the title stays: no succession was started
}

func TestWin_LastWolfHanged_VillageWinsAtDusk(t *testing.T) {
	g := NewGame(t, append(tenPlayers(), Chief("Lan"))...)
	for _, wolf := range []string{"An", "Binh", "Chi"} {
		g.Night()
		g.Dawn().ExpectNobodyDied()
		if wolf != "Chi" {
			g.Day().Hanged(wolf)
		}
	}
	g.Day().Hanged("Chi")
	g.ExpectPhase(engine.PhaseEnded)
	if w := g.State().Winner; w == nil || *w != engine.TeamVillage {
		t.Fatalf("winner %v, want village", w)
	}
}

func TestWin_HangingOneOfSeveralWolves_GameGoesOn(t *testing.T) {
	g := NewGame(t, append(tenPlayers(), Chief("Lan"))...)
	g.Night()
	g.Dawn().ExpectNobodyDied()
	g.Day().Hanged("An")
	g.ExpectPhase(engine.PhaseDusk)
}

func TestWin_AfterTheGameEnded_CommandsAreRejected(t *testing.T) {
	g := NewGame(t, append(tenPlayers(), Chief("Lan"))...)
	bitesOnTheseNights(g, "Phuc", "Giang", "Hoa", "Khoa")
	g.Dawn()
	g.ExpectPhase(engine.PhaseEnded)
	g.DoExpectErr(engine.Nominate{Voter: "An", Target: "Lan"}, engine.ErrWrongPhase)
	before := len(g.Events())
	g.TimeOut()
	if len(g.Events()) != before {
		t.Fatalf("a timeout after the end produced events: %v", g.Events()[before:])
	}
}

// Rules §8: the win check runs once the whole chain of a hanging is over, so a
// hanged Hunter's shot comes first.
func TestWin_HangedHunterShootsTheLastWolf_VillageWinsAfterTheShot_NotBefore(t *testing.T) {
	g := NewGame(t,
		Seats("An", "Phuc", "Giang", "Hoa", "Khoa"),
		Cast("An", Werewolf, "Phuc", Hunter, "Giang", Villager, "Hoa", Villager, "Khoa", Villager))
	g.Night()
	g.Dawn().ExpectNobodyDied()
	g.Day().Hanged("Phuc")
	g.ExpectPhase(engine.PhaseHunterShot) // nothing is decided yet
	g.Hunter("Phuc").Shoots("An")
	g.ExpectPhase(engine.PhaseEnded)
	if w := g.State().Winner; w == nil || *w != engine.TeamVillage {
		t.Fatalf("winner %v, want the village", w)
	}
}

// Hanging the Hunter alone makes it 2 wolves against 2 villagers; the verdict
// is read on the situation after his shot, not on the one before it.
func TestWin_HangedHunterLevelsTheNumbers_OutcomeFollowsTheShot(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		wolvesWin bool
	}{
		{"he shoots a wolf: the game goes on", "Binh", false},
		{"he shoots a villager: the wolves win", "Giang", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGame(t,
				Seats("An", "Binh", "Phuc", "Giang", "Hoa", "Khoa", "Lan"),
				Cast("An", Werewolf, "Binh", Werewolf, "Phuc", Hunter, "Giang", Villager, "Hoa", Villager,
					"Khoa", Villager, "Lan", Villager))
			g.Night()
			killOff(g, "An", "Binh", "Phuc", "Giang", "Hoa") // 2 wolves, the Hunter, 2 villagers
			g.Dawn().ExpectNobodyDied()
			g.Day().Hanged("Phuc") // 2 wolves against 2 villagers, before the shot
			g.ExpectPhase(engine.PhaseHunterShot)
			g.Hunter("Phuc").Shoots(tc.target)
			ended := g.State().Phase == engine.PhaseEnded
			if ended != tc.wolvesWin {
				t.Fatalf("ended=%v, want %v", ended, tc.wolvesWin)
			}
			if w := g.State().Winner; tc.wolvesWin && (w == nil || *w != engine.TeamWolves) {
				t.Fatalf("winner %v, want the wolves", w)
			}
		})
	}
}
