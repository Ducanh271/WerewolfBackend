package roles_test

import (
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/i18n"
	"masoi/internal/roles"

	. "masoi/internal/engine/enginetest"
)

func TestChiefElection_StartsAfterRoleViewAndLastsOneMinute(t *testing.T) {
	g := newTable(t)
	g.ExpectPhase(engine.PhaseRoleView)
	g.Election()
	g.ExpectPhase(engine.PhaseChiefElection)
	if got := g.State().Settings.ChiefElection; got != time.Minute {
		t.Fatalf("election lasts %v, want 1m", got)
	}
}

func TestChiefElection_MostVotesWins(t *testing.T) {
	g := newTable(t)
	g.Election().
		Vote("An", "Phuc").Vote("Binh", "Phuc").Vote("Chi", "Phuc").
		Vote("Dung", "Em").Vote("Em", "Em").
		Close(1)
	g.ExpectChief("Phuc").ExpectPhase(engine.PhaseNight)
}

func TestChiefElection_VoterMayChangeVoteUntilTimeIsUp(t *testing.T) {
	g := newTable(t)
	g.Election().
		Vote("An", "Phuc").Vote("Binh", "Phuc").
		Vote("An", "Giang").Vote("Binh", "Giang").
		Close(1)
	g.ExpectChief("Giang")
}

func TestChiefElection_VotingForYourselfIsAllowed(t *testing.T) {
	g := newTable(t)
	g.Election().Vote("Hoa", "Hoa").Close(1)
	g.ExpectChief("Hoa")
}

func TestChiefElection_ChiefKeepsTheirRole(t *testing.T) {
	g := newTable(t)
	g.Election().Vote("An", "Em").Vote("Em", "Em").Close(1)
	g.ExpectChief("Em").ExpectRole("Em", roles.Seer)
}

func TestChiefElection_ChiefOptionLetsAScenarioPickTheChief(t *testing.T) {
	g := NewGame(t, append(tenPlayers(), Chief("Binh"))...)
	g.Night()
	g.ExpectChief("Binh")
}

func TestChiefElection_TieIsSettledBySeedAmongTheTiedOnly(t *testing.T) {
	winners := map[engine.PlayerID]bool{}
	for seed := int64(0); seed < 30; seed++ {
		g := newTable(t)
		g.Election().Vote("An", "Phuc").Vote("Binh", "Giang").Vote("Chi", "Hoa").Vote("Dung", "Hoa").
			Vote("Em", "Phuc").Vote("Lan", "Giang").Close(seed)
		winners[g.State().Chief] = true
	}
	if len(winners) != 3 || !winners["Phuc"] || !winners["Giang"] || !winners["Hoa"] {
		t.Fatalf("a three-way tie should give each tied player a chance and nobody else, got %v", winners)
	}
}

func TestChiefElection_NobodyVotes_RandomAmongAllPlayersBySeed(t *testing.T) {
	winners := map[engine.PlayerID]bool{}
	for seed := int64(0); seed < 60; seed++ {
		g := newTable(t)
		g.Election().Close(seed)
		again := newTable(t)
		again.Election().Close(seed)
		if g.State().Chief == "" || g.State().Chief != again.State().Chief {
			t.Fatalf("seed %d: same seed must give the same Chief, got %q and %q", seed, g.State().Chief, again.State().Chief)
		}
		winners[g.State().Chief] = true
	}
	if len(winners) < 5 {
		t.Fatalf("with nobody voting any player may win, but 60 seeds only gave %v", winners)
	}
}

func TestChiefElection_InvalidVotesAreRejected(t *testing.T) {
	g := newTable(t)
	e := g.Election()
	e.VoteExpectErr("An", "Nobody", engine.ErrInvalidTarget)
	e.VoteExpectErr("Stranger", "An", engine.ErrUnknownPlayer)
	e.Close(1)
	g.DoExpectErr(engine.VoteChief{Voter: "An", Candidate: "Binh"}, engine.ErrWrongPhase)
}

func TestChiefElection_ResultIsAnnouncedPublicly(t *testing.T) {
	g := newTable(t)
	g.Election().Vote("An", "Em").Close(1)
	for _, e := range g.Events() {
		if c, ok := e.(engine.ChiefElected); ok {
			if c.Chief != "Em" || c.AnnouncementKey() != i18n.ChiefElected {
				t.Fatalf("unexpected announcement %+v", c)
			}
			return
		}
	}
	t.Fatal("no ChiefElected event")
}
