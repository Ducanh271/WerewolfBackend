package roles_test

import (
	"testing"

	"masoi/internal/engine"

	. "masoi/internal/engine/enginetest"
)

// pairedDay1 plays night 1 with Cupid pairing a and b, and nobody dying; the
// day follows.
func pairedDay1(t testing.TB, override map[string]engine.RoleID, a, b string) *Game {
	g := cupidTableWith(t, override)
	g.Night().Cupid("Lan").Pairs(a, b)
	g.Dawn().ExpectNobodyDied()
	return g
}

// announcedDead is the list of the last morning announcement.
func announcedDead(t testing.TB, g *Game) []engine.PlayerID {
	evs := g.Events()
	for i := len(evs) - 1; i >= 0; i-- {
		if a, ok := evs[i].(engine.DawnAnnounced); ok {
			if a.Kind != engine.DawnDeaths {
				t.Fatalf("the morning announced %v, want deaths", a.Kind)
			}
			return a.Dead
		}
	}
	t.Fatal("no morning announcement")
	return nil
}

func wantDead(t testing.TB, got []engine.PlayerID, want ...string) {
	if len(got) != len(want) {
		t.Fatalf("announced dead %v, want %v", got, want)
	}
	for i := range want {
		if string(got[i]) != want[i] {
			t.Fatalf("announced dead %v, want %v", got, want)
		}
	}
}

// Rules §7, §9 step 7: a lover who dies in the night takes the partner the same
// night, and the morning lists both in seating order, with no cause.
func TestLove_LoverBittenInTheNight_PartnerDiesTheSameNight_AnnouncedInSeatingOrder(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Em", "Chi")
	n.Wolf("An").Bites("Em").Wolf("Binh").Bites("Em")
	g.Dawn().ExpectDead("Em", "Chi")
	wantDead(t, announcedDead(t, g), "Chi", "Em") // the pipeline killed Em first
}

func TestLove_LoverPoisonedByTheWitch_PartnerDiesTheSameNight(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("Em", "Chi").Witch("Giang").Poisons("Chi")
	g.Dawn().ExpectDead("Em", "Chi")
}

func TestLove_PoisonedWolfDiesAtTheNextPackStep_ItsPartnerDiesWithIt(t *testing.T) {
	g := cupidTableWith(t, map[string]engine.RoleID{"Khoa": RustyKnight})
	n := g.Night().Cupid("Lan").Pairs("An", "Em")
	n.Wolf("An").Bites("Khoa").Wolf("Binh").Bites("Khoa") // the Knight poisons An, next on his left
	g.Dawn().ExpectDead("Khoa")
	g.Day().NobodyHanged()
	g.Night().Wolf("Binh").Bites("Dung")
	g.Dawn().ExpectDead("An", "Em", "Dung") // An dies at step 6, his partner Em at step 7
}

// Rules §7, §8: one death by day. The survivor does not die at once: they stay
// alive through the day, then die in the next night, announced with the others.
func TestLove_LoverHangedByDay_PartnerLivesOnAndDiesInTheNextNightWithoutACall(t *testing.T) {
	g := pairedDay1(t, nil, "Chi", "Khoa")
	g.Day().Hanged("Chi")
	g.ExpectAlive("Khoa")
	g.Night().TurnOf(Seer) // Khoa has no night call and is alive for everybody
	g.As("Em").PublicTable()
	g.As("Hoa").Dead("Chi")
	g.Dawn().ExpectDead("Khoa")
	wantDead(t, announcedDead(t, g), "Khoa")
}

func TestLove_LoverShotByTheHunter_PartnerDiesInTheNextNightToo(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Chi", "Khoa")
	n.Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	g.Dawn().ExpectDead("Phuc")
	g.Hunter("Phuc").Shoots("Chi")
	g.ExpectAlive("Khoa") // the shot at dawn is a death by day
	g.Day().NobodyHanged()
	g.Night()
	g.Dawn().ExpectDead("Khoa")
}

// Rules §7: until the night the survivor is a living player: they vote, may be
// nominated, may be hanged (and then die at once).
func TestLove_SurvivorOfTheDay_VotesAndCanBeHangedLikeAnyone(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Chi", "Khoa")
	n.Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	g.Dawn()
	g.Hunter("Phuc").Shoots("Chi")
	day := g.Day()
	g.Do(engine.Nominate{Voter: g.Player("Khoa"), Target: g.Player("Binh")}) // the survivor may nominate
	day.Hanged("Khoa")
	g.ExpectDead("Phuc", "Chi", "Khoa")
	g.Night()
	g.Dawn().ExpectNobodyDied() // nobody is left to grieve
}

func TestLove_GriefIsNotAWinCheckDeath_SurvivorCountsAliveAtDusk(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("Chi", "Khoa")
	killOff(g, "An", "Binh", "Chi", "Khoa", "Em", "Dung")
	g.Dawn()
	g.Day().Hanged("Chi") // 2 wolves, 3 villagers with Khoa; 2 against 2 if Khoa were counted dead
	g.ExpectPhase(engine.PhaseDusk)
	g.ExpectAlive("An", "Binh", "Khoa", "Em", "Dung")
}

// A role with a night call dies at its FIRST call of the night, which still
// runs, and cannot use its ability. A wolf is not counted in the bite vote.
func TestLove_GrievingCursedWolf_DiesAtItsOwnStep_CannotCurse_NotCountedInTheBite(t *testing.T) {
	g := pairedDay1(t, nil, "An", "Chi")
	g.Day().Hanged("Chi")
	n := g.Night()
	g.As("An").Doomed(false)
	n.TurnOf(CursedWolf)
	g.As("An").Turn(CursedWolf).Doomed(true).Grief(true)
	g.As("Binh").Doomed(false).Grief(false)
	g.As("Em").PublicTable()
	n.CursedWolf("An").CannotCurse("Em", engine.ErrNotYourTurn)
	n.TurnOf(Werewolf)
	g.As("Binh").Turn(Werewolf).PackDoomed("An")
	n.Wolf("An").CannotBite("Em", engine.ErrNotYourTurn)
	n.Wolf("Binh").Bites("Dung") // the only voter: more than half
	for _, who := range []string{"Em", "Hoa", "Dung"} {
		for _, p := range g.As(who).View().Players {
			if p.ID == g.Player("An") && !p.Alive {
				t.Fatalf("%s sees An dead during the night", who)
			}
		}
	}
	g.Dawn().ExpectDead("An", "Dung")
}

func TestLove_GrievingVisitor_DiesAtHerStep_BlocksNobody(t *testing.T) {
	g := pairedDay1(t, map[string]engine.RoleID{"Khoa": Visitor}, "Chi", "Khoa")
	g.Day().Hanged("Chi")
	n := g.Night()
	n.Visitor("Khoa").CannotVisit("Em", engine.ErrNotYourTurn)
	n.Visitor("Khoa").RunsOutOfTime(7) // no random visit either
	if len(g.State().Pending.Blocked) != 0 {
		t.Fatalf("a grieving Visitor blocked %v", g.State().Pending.Blocked)
	}
	g.Dawn().ExpectDead("Khoa")
}

func TestLove_GrievingHunter_DiesAtStepSevenAndStillShoots(t *testing.T) {
	g := pairedDay1(t, nil, "Chi", "Phuc")
	g.Day().Hanged("Chi")
	g.Night()
	g.Dawn().ExpectDead("Phuc")
	g.As("Phuc").CanShoot(true)
	g.Hunter("Phuc").Shoots("Binh")
	g.ExpectDead("Chi", "Phuc", "Binh")
}

func TestLove_GrievingChief_PassesTheTitleAfterTheMorning(t *testing.T) {
	g := pairedDay1(t, nil, "Chi", "Hoa")
	g.Day().Hanged("Chi")
	g.Night()
	g.Dawn().ExpectDead("Hoa")
	g.ExpectPhase(engine.PhaseChiefSuccession)
	g.NameSuccessor("Hoa", "Em").ExpectChief("Em")
}

func TestLove_GrievingChief_WinCheckComesBeforeTheTitle(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("Chi", "Hoa")
	killOff(g, "An", "Binh", "Chi", "Hoa", "Em", "Dung")
	g.HandBuild(func(s *engine.State) { s.Chief = g.Player("Hoa") })
	g.Dawn()
	g.Day().Hanged("Chi") // 2 wolves, 3 villagers
	g.Night()
	g.Dawn().ExpectDead("Hoa") // 2 wolves, 2 villagers: the wolves win at the morning
	g.ExpectPhase(engine.PhaseEnded)
}

// The chain of rules §8: the hanged Hunter shoots his own lover; the partner is
// dead already, nobody is left to grieve. A shot at somebody else marks that
// victim's partner, after the whole chain.
func TestLove_HangedHunterShootsHisLover_NobodyGrieves(t *testing.T) {
	g := pairedDay1(t, nil, "Phuc", "Chi")
	g.Day().Hanged("Phuc")
	g.Hunter("Phuc").Shoots("Chi")
	g.ExpectDead("Phuc", "Chi")
	g.Night()
	g.Dawn().ExpectNobodyDied()
}

func TestLove_HangedHunterShootsAnotherHunter_LoverIsMarkedAfterTheWholeChain(t *testing.T) {
	g := pairedDay1(t, nil, "Phuc", "Chi")
	g.HandBuild(func(s *engine.State) { s.Players[g.Player("Khoa")].Role = Hunter })
	g.Day().Hanged("Phuc")
	g.Hunter("Phuc").Shoots("Khoa")
	g.Hunter("Khoa").Shoots("Dung")
	g.ExpectDead("Phuc", "Khoa", "Dung")
	g.ExpectAlive("Chi")
	if g.State().Players[g.Player("Chi")].Flags[engine.FlagGrieving] == 0 {
		t.Fatal("Chi, whose lover was hanged, is not marked to die")
	}
	g.Night()
	g.Dawn().ExpectDead("Chi")
}
