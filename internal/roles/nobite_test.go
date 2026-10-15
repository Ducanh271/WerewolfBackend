package roles_test

import (
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"

	. "masoi/internal/engine/enginetest"
)

// The usual witchTable: wolves An, Binh, Chi; Guard Dung; Seer Em; Witch Lan.

func TestNight_PackVotesNoBite_NobodyDies_AnnouncesNobodyDied(t *testing.T) {
	g := witchTable(t)
	g.Night().Wolf("An").BitesNobody().Wolf("Binh").BitesNobody().Wolf("Chi").Bites("Phuc")
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnNobodyDied)
}

func TestNight_NoBiteNeedsMoreThanHalf(t *testing.T) {
	tests := []struct {
		name string
		vote func(n *Night) *Night
		dead []string
	}{
		{"two of three: no bite", func(n *Night) *Night {
			return n.Wolf("An").BitesNobody().Wolf("Binh").BitesNobody().Wolf("Chi").Bites("Phuc")
		}, nil},
		{"one of three: the bite of the other two wins", func(n *Night) *Night {
			return n.Wolf("An").BitesNobody().Wolf("Binh").Bites("Phuc").Wolf("Chi").Bites("Phuc")
		}, []string{"Phuc"}},
		{"a tie of no bite and a target: nobody", func(n *Night) *Night {
			return n.Wolf("An").BitesNobody().Wolf("Binh").Bites("Phuc")
		}, nil},
		{"a wolf changes the vote to no bite", func(n *Night) *Night {
			return n.Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc").Wolf("Binh").BitesNobody().Wolf("Chi").BitesNobody()
		}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := witchTable(t)
			tc.vote(g.Night())
			d := g.Dawn()
			if tc.dead == nil {
				d.ExpectNobodyDied().ExpectAnnouncement(i18n.DawnNobodyDied)
			} else {
				d.ExpectDead(tc.dead...)
			}
		})
	}
}

func TestNight_NoBiteCountsOnlyWolvesWhoCanVote(t *testing.T) {
	g := witchTable(t)
	n := g.Night()
	g.HandBuild(func(s *engine.State) { s.Players[g.Player("Chi")].Alive = false }) // two voters left
	n.Wolf("An").BitesNobody().Wolf("Binh").BitesNobody()
	g.Dawn().ExpectNobodyDied()
}

func TestNight_NoBiteAndGuard_NobodyIsProtectedSoTheAnnouncementIsTheUsualOne(t *testing.T) {
	g := witchTable(t)
	g.Night().Guard("Dung").Protects("Phuc").Wolf("An").BitesNobody().Wolf("Binh").BitesNobody()
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnNobodyDied)
}

func TestNight_NoBite_WitchSeesNobodyBitten_AndKeepsHerPotion(t *testing.T) {
	g := witchTable(t)
	g.Night().Wolf("An").BitesNobody().Wolf("Binh").BitesNobody().TurnOf(Witch)
	g.As("Lan").BiteShown(true).Bitten("")
	g.Dawn().ExpectNobodyDied()
	g.Day().NobodyHanged()
	bite(g.Night(), "Phuc", "An", "Binh").Witch("Lan").Heals() // the heal potion was never used
	g.Dawn().ExpectNobodyDied()
}

func TestNight_NoBite_PackSeesTheVote_NobodyElseSeesAnything(t *testing.T) {
	g := witchTable(t)
	g.Night().Wolf("An").BitesNobody()
	g.As("Binh").WolfVote("An", "").WolfVotesInFeed(1)
	g.As("Em").NoPack().WolfVotesInFeed(0).NoSecretsOfOthers()
	g.Dawn()
	for _, who := range []string{"Em", "Dung", "Phuc"} {
		g.As(who).NoSecretsOfOthers().WolfVotesInFeed(0)
	}
}

func TestNight_NoBite_WolvesCanStillBiteAfterwards(t *testing.T) {
	g := witchTable(t)
	g.Night().Wolf("An").BitesNobody().Wolf("Binh").BitesNobody()
	g.Dawn().ExpectNobodyDied()
	g.Day().NobodyHanged()
	bite(g.Night(), "Phuc", "An", "Binh")
	g.Dawn().ExpectDead("Phuc")
}

func TestNight_NoBite_StillRefusesDeadTargets(t *testing.T) {
	g := witchTable(t)
	n := g.Night()
	g.HandBuild(func(s *engine.State) { s.Players[g.Player("Phuc")].Alive = false })
	n.Wolf("An").CannotBite("Phuc", engine.ErrInvalidTarget)
}
