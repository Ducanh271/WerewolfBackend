package roles_test

import (
	"reflect"
	"strings"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"

	. "masoi/internal/engine/enginetest"
)

// Rules §9b: every role has a card with a name, a team, the team's goal and a
// short description, and the card's team is the role's own team.
func TestRoleCards_EveryRoleHasAFullCard_OnItsOwnTeam(t *testing.T) {
	for id, role := range engine.Registry {
		c, ok := i18n.Card(string(id))
		if !ok {
			t.Errorf("%s has no card", id)
			continue
		}
		if c.Name == "" || c.Team == "" || c.Goal == "" || c.Description == "" {
			t.Errorf("%s has an incomplete card %+v", id, c)
		}
		wantTeam, wantGoal := "Dân", "Dân thắng khi không còn con Sói nào."
		if role.BaseTeam() == engine.TeamWolves {
			wantTeam, wantGoal = "Sói", "Sói thắng khi số Sói ≥ số Dân."
		}
		if c.Team != wantTeam || c.Goal != wantGoal {
			t.Errorf("%s card is on %q with goal %q, want %q / %q", id, c.Team, c.Goal, wantTeam, wantGoal)
		}
	}
}

// Secrecy: the card on every phone has the same shape, whatever the role: the
// same fields, all of them set.
func TestRoleCards_SameShapeForEveryRole_EachPlayerSeesOnlyTheirOwn(t *testing.T) {
	g := cupidTable(t)
	g.Night()
	for _, name := range all {
		me := g.State().Players[g.Player(name)]
		seen := g.As(name).OwnCard(me.Role).NoRoleExceptOwn().CardOf(name, me.Role)
		card := reflect.ValueOf(*seen.View().Me.RoleCard)
		for i := 0; i < card.NumField(); i++ {
			if card.Field(i).String() == "" {
				t.Fatalf("%s's card has an empty %s", name, card.Type().Field(i).Name)
			}
		}
	}
}

// Rules §9b: the role list of the game, with each role's name and short
// description, is public: the same on every phone, in the order of the table.
func TestRoleCards_PublicRoleListCarriesNameAndDescription(t *testing.T) {
	g := cupidTable(t)
	g.Night()
	var want []engine.RoleListCard
	for _, id := range []engine.RoleID{Werewolf, CursedWolf, Villager, Seer, Guard, Witch, Hunter, Cupid} {
		c := CardText(id)
		want = append(want, engine.RoleListCard{Role: id, Name: c.Name, Description: c.Description})
	}
	for _, name := range all {
		if got := g.As(name).View().RoleCards; !reflect.DeepEqual(got, want) {
			t.Fatalf("%s sees the role list %+v, want %+v", name, got, want)
		}
	}
}

// Dead players, and everybody once the game is over, see every player's card.
func TestRoleCards_DeadPlayersAndTheEndScreenSeeEveryCard(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Em", "Dung")
	n.Wolf("An").Bites("Khoa").Wolf("Binh").Bites("Khoa")
	g.Dawn().ExpectDead("Khoa")
	for _, name := range all {
		g.As("Khoa").CardOf(name, g.State().Players[g.Player(name)].Role)
	}
	g.As("Em").CardOf("An", "").CardOf("Khoa", "")

	killOff(g, "An", "Em")
	g.Day().Hanged("An")
	g.ExpectPhase(engine.PhaseEnded)
	for _, viewer := range []string{"Em", "Binh"} {
		for _, name := range all {
			g.As(viewer).CardOf(name, g.State().Players[g.Player(name)].Role)
		}
	}
}

// Extra cards: the curse shows from the phone check on, never before.
func TestStatusCards_CursedFromThePhoneCheckOn(t *testing.T) {
	g := cupidTableWith(t, map[string]engine.RoleID{"Lan": Villager})
	n := g.Night().CursedWolf("An").Curses("Em")
	g.As("Em").StatusCards()
	n.PhoneCheck()
	g.As("Em").StatusCards(i18n.StatusCursed).OwnCard(Seer) // the role card stays the Seer's
	g.As("Chi").StatusCards()
	g.As("Hoa").StatusCards(i18n.StatusChief)
}

// The Chief's card follows the title: after the election, then to the
// successor; the former Chief no longer holds it.
func TestStatusCards_ChiefFollowsTheTitle(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Em", "Dung")
	g.As("Hoa").StatusCards(i18n.StatusChief)
	g.As("Em").StatusCards()
	n.Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Hoa")
	g.Dawn().ExpectDead("Hoa")
	g.NameSuccessor("Hoa", "Em")
	g.As("Em").StatusCards(i18n.StatusLovers, i18n.StatusChief)
	g.As("Hoa").StatusCards()
}

func loverCardText(t testing.TB, g *Game, who string) string {
	t.Helper()
	cards := g.As(who).View().Me.StatusCards
	if len(cards) != 1 {
		t.Fatalf("%s holds %+v, want one lover card", who, cards)
	}
	return cards[0].Text
}

// The lovers' card: same team or cross-team, naming the partner, from the
// phone check on; a curse turns it into the cross-team card at the next check.
// Cupid holds none.
func TestStatusCards_Lovers(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Em", "Chi")
	g.As("Em").StatusCards()
	n.PhoneCheck()
	g.As("Em").StatusCards(i18n.StatusLovers)
	g.As("Chi").StatusCards(i18n.StatusLovers)
	g.As("Lan").StatusCards()
	if text := loverCardText(t, g, "Em"); !strings.HasPrefix(text, "Bạn và Chi yêu nhau.") {
		t.Fatalf("Em's lover card reads %q", text)
	}
	n.CursedWolf("An").Curses("Em")
	g.As("Chi").StatusCards(i18n.StatusLovers) // not before the check
	n.PhoneCheck()
	g.As("Em").StatusCards(i18n.StatusCursed, i18n.StatusLoversCross)
	g.As("Chi").StatusCards(i18n.StatusLoversCross)
	if text := loverCardText(t, g, "Chi"); !strings.HasPrefix(text, "Bạn và Em thuộc phe thứ ba") {
		t.Fatalf("Chi's lover card reads %q", text)
	}
}

// Once on the third team, always on it: the card stays after the couple broke.
func TestStatusCards_CrossTeamCardStaysAfterThePartnerDied(t *testing.T) {
	g := pairedDay1(t, nil, "Binh", "Chi")
	g.As("Chi").StatusCards(i18n.StatusLoversCross)
	g.Day().Hanged("Binh")
	g.ExpectAlive("Chi")
	g.As("Chi").StatusCards(i18n.StatusLoversCross)
}
