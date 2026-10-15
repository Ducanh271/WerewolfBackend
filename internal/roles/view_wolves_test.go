package roles_test

import (
	"testing"

	"masoi/internal/roles"
)

func TestView_Wolves_SeePackLiveVotesAndWolfSeerResults(t *testing.T) {
	result := roles.WolfSeerChecked{WolfSeer: "Binh", Target: "Em", IsSeer: true}
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"the pack, without roles", func(t *testing.T) {
			g := newTable(t)
			for _, wolf := range []string{"An", "Binh", "Chi"} {
				g.As(wolf).Pack("An", "Binh", "Chi").NoRoleExceptOwn()
			}
		}},
		{"live bite votes", func(t *testing.T) {
			g := newTable(t)
			g.Night().Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Giang")
			g.As("Chi").WolfVote("An", "Phuc").WolfVote("Binh", "Giang").WolfVotesInFeed(2)
		}},
		{"a cursed player joins the pack at the phone check", func(t *testing.T) {
			g := cursedTable(t)
			n := g.Night().CursedWolf("An").Curses("Phuc")
			g.As("Phuc").NoPack()
			g.As("An").Pack("An", "Binh")
			n.PhoneCheck()
			g.As("Phuc").Pack("An", "Binh", "Phuc")
			g.As("An").Pack("An", "Binh", "Phuc")
		}},
		{"the Wolf seer's result goes to the whole pack", func(t *testing.T) {
			g := cursedTable(t)
			g.Night().CursedWolf("An").Curses("Phuc").PhoneCheck()
			g.Inject(result)
			for _, wolf := range []string{"An", "Binh", "Phuc"} {
				g.As(wolf).Event(result)
			}
			g.As("Em").NoEvent(result)
			g.As("Hoa").NoEvent(result)
		}},
		{"villagers see none of it", func(t *testing.T) {
			g := newTable(t)
			g.Night().Wolf("An").Bites("Phuc")
			g.As("Hoa").NoPack().WolfVotesInFeed(0)
			g.As("Dung").NoPack().WolfVotesInFeed(0)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

// Rules §5: the curse takes effect at once; the cursed player then sees the
// pack and its votes, but not the votes cast before the curse.
func TestView_CursedPlayer_SeesPackVotesOnlyFromTheCurseOn(t *testing.T) {
	g := cursedTable(t)
	g.Night().Guard("Dung").Protects("Giang").Wolf("An").Bites("Giang").Wolf("Binh").Bites("Giang")
	g.As("Phuc").NoPack().WolfVotesInFeed(0)
	g.Dawn().ExpectNobodyDied()
	g.Day().NobodyHanged()

	night2 := g.Night().CursedWolf("An").Curses("Phuc").PhoneCheck()
	g.As("Phuc").Cursed(true).Pack("An", "Binh", "Phuc").WolfVotesInFeed(0)

	night2.Wolf("An").Bites("Hoa").Wolf("Phuc").Bites("Hoa")
	g.As("Phuc").WolfVote("An", "Hoa").WolfVotesInFeed(2) // night 2 only
	g.As("Binh").WolfVotesInFeed(4)                       // both nights
}
