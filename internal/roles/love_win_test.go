package roles_test

import (
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"

	. "masoi/internal/engine/enginetest"
)

func expectWinner(t testing.TB, g *Game, team engine.Team) {
	t.Helper()
	g.ExpectPhase(engine.PhaseEnded)
	if w := g.State().Winner; w == nil || *w != team {
		t.Fatalf("winner %v, want %v", w, team)
	}
}

// Rules §8: a couple of one wolf and one villager wins when it is the last two
// alive; the third team is worked out for the couple, Player.Team stays.
func TestWin_CrossTeamCouple_WinsWhenTheyAreTheLastTwoAlive(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("An", "Em")
	killOff(g, "An", "Em", "Chi")
	g.Dawn()
	g.Day().Hanged("Chi")
	expectWinner(t, g, engine.TeamLovers)
	if g.State().Players[g.Player("An")].Team != engine.TeamWolves {
		t.Fatal("the wolf lover must stay on the wolf team")
	}
}

// The couple counts on neither side, and a wolf lover still blocks the
// village's win: the game goes on while villagers and the couple remain.
func TestWin_CrossTeamCouple_LeftOutOfTheCounts_WolfLoverStillBlocksTheVillage(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("An", "Em")
	killOff(g, "An", "Em", "Chi", "Dung")
	g.Dawn()
	g.Day().Hanged("Dung")
	g.ExpectPhase(engine.PhaseDusk) // no village win (a wolf lives), no wolf win (no wolf outside the couple)
	g.Night()
	g.Dawn()
	g.Day().Hanged("Chi")
	expectWinner(t, g, engine.TeamLovers)
}

// Rules §7: Cupid who pairs himself with a wolf belongs to the third team.
func TestWin_CupidPairedWithAWolf_IsInTheThirdTeam(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("Lan", "Binh")
	killOff(g, "Lan", "Binh", "Chi")
	g.Dawn()
	g.Day().Hanged("Chi")
	expectWinner(t, g, engine.TeamLovers)
	if g.State().Players[g.Player("Lan")].Team != engine.TeamVillage {
		t.Fatal("Cupid's own team is computed with the couple, Player.Team stays")
	}
	announced := false
	for _, e := range g.Events() {
		if a, ok := e.(engine.GameEnded); ok {
			announced = a.AnnouncementKey() == i18n.WinLovers
		}
	}
	if !announced {
		t.Fatal("the couple's win was not announced")
	}
}

// A wolf lover dies with the partner: when the village kills the couple, the
// rest of the wolves decide the game as usual.
func TestWin_TwoWolvesInLove_AreTheWolfTeamAndDieTogether(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("An", "Binh")
	killOff(g, "An", "Binh", "Chi", "Dung", "Em", "Hoa")
	g.Dawn()
	g.Day().Hanged("An") // Binh follows in the night
	g.Night()
	g.Dawn().ExpectDead("Binh")
	expectWinner(t, g, engine.TeamVillage)
}
