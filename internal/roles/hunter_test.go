package roles_test

import (
	"reflect"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/i18n"

	. "masoi/internal/engine/enginetest"
)

// hunterTable is the usual table with the Hunter in Phuc's seat and the Witch
// in Giang's.
func hunterTable(t *testing.T, opts ...Option) *Game {
	return NewGame(t, append([]Option{
		Seats(all...),
		Cast("An", Werewolf, "Binh", Werewolf, "Chi", Werewolf, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Villager, "Khoa", Villager, "Lan", Villager),
	}, opts...)...)
}

// bitHunter plays a night where the pack bites the Hunter.
func bitHunter(g *Game) *Night {
	return packBites(g.Night(), "Phuc")
}

// trace lists, in order, the dawn/hunter events of the game by name.
func trace(g *Game) []string {
	var out []string
	for _, e := range g.Events() {
		switch e.(type) {
		case engine.DawnAnnounced, engine.Hanged, engine.HunterRevealed, engine.HunterShotStarted,
			engine.HunterShot, engine.HunterTimedOut, engine.ChiefSucceeded, engine.ChiefSuccessionStarted:
			out = append(out, reflect.TypeOf(e).Name())
		}
	}
	return out
}

func revealedTo(g *Game, viewer string) bool {
	for _, p := range g.As(viewer).View().Players {
		if p.ID == g.Player("Phuc") {
			return p.Role == Hunter && p.Team == engine.TeamNone
		}
	}
	return false
}

func TestHunter_DiesBitten_RoleIsRevealedToEverybody(t *testing.T) {
	g := hunterTable(t)
	g.As("Hoa").NoRoleExceptOwn()
	bitHunter(g)
	g.Dawn().ExpectDead("Phuc")
	if !revealedTo(g, "Hoa") || !revealedTo(g, "An") {
		t.Fatal("the dead Hunter's role must be public, and only the role")
	}
	g.As("Hoa").RoleOf("Khoa", "") // nobody else is revealed
}

func TestHunter_Hanged_RoleIsRevealedToEverybody(t *testing.T) {
	g := hunterTable(t)
	g.Night()
	g.Dawn().ExpectNobodyDied()
	g.Day().Hanged("Phuc")
	g.ExpectPhase(engine.PhaseHunterShot)
	if !revealedTo(g, "Hoa") {
		t.Fatal("the hanged Hunter must be revealed")
	}
}

func TestHunter_Poisoned_RoleIsRevealedToEverybody(t *testing.T) {
	g := hunterTable(t)
	g.Night().Witch("Giang").Poisons("Phuc")
	g.Dawn().ExpectDead("Phuc")
	if !revealedTo(g, "Hoa") {
		t.Fatal("the poisoned Hunter must be revealed")
	}
}

func TestHunter_ShotByAnotherHunter_RoleIsRevealedAndHeShootsToo(t *testing.T) {
	g := hunterTable(t)
	g.HandBuild(func(s *engine.State) { s.Players[g.Player("Hoa")].Role = Hunter }) // only one Hunter per real game
	bitHunter(g)
	g.Dawn().ExpectDead("Phuc")
	g.Hunter("Phuc").Shoots("Hoa")
	g.ExpectPhase(engine.PhaseHunterShot)
	g.As("Hoa").CanShoot(true).Shooter("Hoa")
	g.Hunter("Hoa").Shoots("Khoa")
	g.ExpectDead("Phuc", "Hoa", "Khoa").ExpectPhase(engine.PhaseDawn)
	want := []string{"DawnAnnounced", "HunterRevealed", "HunterShotStarted", "HunterShot",
		"HunterRevealed", "HunterShotStarted", "HunterShot"}
	if got := trace(g); !reflect.DeepEqual(got, want) {
		t.Fatalf("events %v, want %v", got, want)
	}
}

func TestHunter_Flow_AnnouncementThenReveal_ThenFifteenSecondsOnlyHeCanShoot(t *testing.T) {
	g := hunterTable(t)
	bitHunter(g)
	d := g.Dawn().ExpectDead("Phuc").ExpectAnnouncement(i18n.DawnSomeoneDied).ExpectAnnouncement(i18n.HunterRevealed)
	_ = d
	want := []string{"DawnAnnounced", "HunterRevealed", "HunterShotStarted"}
	if got := trace(g); !reflect.DeepEqual(got, want) {
		t.Fatalf("events %v, want %v", got, want)
	}
	g.ExpectPhase(engine.PhaseHunterShot)
	if got := g.State().Settings.HunterShot; got != 15*time.Second {
		t.Fatalf("the shot lasts %v, want 15s", got)
	}
	g.As("Phuc").CanShoot(true).Shooter("Phuc")
	for _, other := range []string{"An", "Em", "Giang", "Hoa"} {
		g.As(other).CanShoot(false).Shooter("Phuc")
	}
	if got := i18n.Vi(i18n.HunterRevealed, "Phuc"); got != "Phuc là Thợ săn." {
		t.Fatalf("reveal line %q", got)
	}
}

func TestHunter_HangedFlow_HangingIsAnnouncedBeforeTheReveal(t *testing.T) {
	g := hunterTable(t)
	g.Night()
	g.Dawn()
	g.Day().Hanged("Phuc")
	want := []string{"DawnAnnounced", "Hanged", "HunterRevealed", "HunterShotStarted"}
	if got := trace(g); !reflect.DeepEqual(got, want) {
		t.Fatalf("events %v, want %v", got, want)
	}
	g.Hunter("Phuc").Shoots("An")
	g.ExpectPhase(engine.PhaseDusk).ExpectDead("Phuc", "An")
}

func TestHunter_NoChoiceInTime_NobodyIsShot(t *testing.T) {
	g := hunterTable(t)
	bitHunter(g)
	g.Dawn().ExpectDead("Phuc")
	g.TimeOut()
	g.ExpectPhase(engine.PhaseDawn).ExpectAlive("An", "Em", "Hoa")
	g.Day().NobodyHanged()
}

func TestHunter_ShotIsMandatory_OnlyALivingPlayerIsAValidChoice(t *testing.T) {
	g := hunterTable(t)
	bitHunter(g)
	g.Dawn()
	g.Hunter("Phuc").CannotShoot("", engine.ErrInvalidTarget)
	g.Hunter("Phuc").CannotShoot(g.Player("Phuc"), engine.ErrInvalidTarget)
	g.DoExpectErr(engine.HunterShoot{By: g.Player("Hoa"), Target: g.Player("An")}, engine.ErrNotYourTurn)
	g.ExpectPhase(engine.PhaseHunterShot)
}

func TestHunter_Shoots_AnnouncedAndTheTargetDiesAtOnce(t *testing.T) {
	g := hunterTable(t)
	bitHunter(g)
	g.Dawn()
	g.Hunter("Phuc").Shoots("An")
	g.ExpectDead("Phuc", "An").ExpectPhase(engine.PhaseDawn)
	if got := i18n.Vi(i18n.HunterShot, "Phuc", "An"); got != "Phuc bắn An." {
		t.Fatalf("shot line %q", got)
	}
	g.As("Hoa").Event(engine.HunterShot{Hunter: g.Player("Phuc"), Target: g.Player("An")})
	g.Day().NobodyHanged()
}

func TestHunter_IsAlsoChief_ShootsFirstThenPassesTheTitle(t *testing.T) {
	g := hunterTable(t)
	g.Election().Vote("An", "Phuc").Close(1)
	bitHunter(g)
	g.Dawn().ExpectDead("Phuc")
	g.ExpectPhase(engine.PhaseHunterShot)
	g.Hunter("Phuc").Shoots("An")
	g.ExpectPhase(engine.PhaseChiefSuccession)
	g.NameSuccessor("Phuc", "Em")
	g.ExpectChief("Em").ExpectPhase(engine.PhaseDawn)
	want := []string{"DawnAnnounced", "HunterRevealed", "HunterShotStarted", "HunterShot", "ChiefSuccessionStarted", "ChiefSucceeded"}
	if got := trace(g); !reflect.DeepEqual(got, want) {
		t.Fatalf("events %v, want %v", got, want)
	}
}

func TestHunter_ShotKillsTheChief_TheChiefStillPassesTheTitle(t *testing.T) {
	g := hunterTable(t)
	g.Election().Vote("An", "Hoa").Close(1)
	bitHunter(g)
	g.Dawn()
	g.Hunter("Phuc").Shoots("Hoa")
	g.ExpectPhase(engine.PhaseChiefSuccession)
	g.NameSuccessor("Hoa", "Em").ExpectChief("Em")
}

func TestHunter_Cursed_KeepsHisAbilityAndShootsAsAWolf(t *testing.T) {
	g := cursedTable(t)
	g.Night().CursedWolf("An").Curses("Phuc").
		Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc").Wolf("Phuc").Bites("Phuc")
	g.Dawn().ExpectDead("Phuc")
	g.As("Phuc").CanShoot(true)
	g.Hunter("Phuc").Shoots("Em")
	g.ExpectDead("Phuc", "Em")
}

func TestHunter_VisitedByTheVisitor_StillShoots(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Phuc")
	packBites(n, "Phuc")
	g.Dawn().ExpectDead("Phuc")
	g.Hunter("Phuc").Shoots("An")
	g.ExpectDead("Phuc", "An")
}

func TestHunter_ShotKillsTheLastWolf_VillageWinsRightAfterTheShot(t *testing.T) {
	g := NewGame(t,
		Seats("An", "Phuc", "Giang", "Hoa", "Khoa"),
		Cast("An", Werewolf, "Phuc", Hunter, "Giang", Villager, "Hoa", Villager, "Khoa", Villager))
	g.Night().Wolf("An").Bites("Phuc")
	g.Dawn().ExpectDead("Phuc").ExpectAnnouncement(i18n.HunterRevealed)
	g.ExpectPhase(engine.PhaseHunterShot) // the win check waits for the shot
	g.Hunter("Phuc").Shoots("An")
	g.ExpectPhase(engine.PhaseEnded)
	if w := g.State().Winner; w == nil || *w != engine.TeamVillage {
		t.Fatalf("winner %v, want the village", w)
	}
}

func TestHunter_ShotFlipsTheOutcome_WinCheckRunsAfterTheShot(t *testing.T) {
	// 2 wolves against 3 villagers and the Hunter: the bite alone ends nothing,
	// but after the shot it is 2 against 2.
	g := NewGame(t,
		Seats("An", "Binh", "Phuc", "Giang", "Hoa", "Khoa", "Lan"),
		Cast("An", Werewolf, "Binh", Werewolf, "Phuc", Hunter, "Giang", Villager, "Hoa", Villager,
			"Khoa", Villager, "Lan", Villager))
	g.HandBuild(func(s *engine.State) { s.Players[g.Player("Khoa")].Alive = false })
	g.Night().Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	g.Dawn().ExpectDead("Phuc")
	g.ExpectPhase(engine.PhaseHunterShot)
	g.Hunter("Phuc").Shoots("Giang")
	g.ExpectPhase(engine.PhaseEnded)
	if w := g.State().Winner; w == nil || *w != engine.TeamWolves {
		t.Fatalf("winner %v, want the wolves", w)
	}
}
