package engine

// Fake roles for engine tests. The engine must not import internal/roles
// (dependencies point inward), so tests register their own roles with the same
// IDs and the same rules §2 constraints as the real ones.

const (
	roleWerewolf   RoleID = "Werewolf"
	roleCursedWolf RoleID = "CursedWolf"
	roleVillager   RoleID = "Villager"
	roleSeer       RoleID = "Seer"
	roleDetective  RoleID = "Detective"
	roleCupid      RoleID = "Cupid"
)

type fakeRole struct {
	id   RoleID
	team Team
	call *NightCall
}

func (r fakeRole) ID() RoleID                                   { return r.id }
func (r fakeRole) BaseTeam() Team                               { return r.team }
func (r fakeRole) NightCall() *NightCall                        { return r.call }
func (r fakeRole) Decide(State, Command) ([]Event, bool, error) { return nil, false, nil }
func (r fakeRole) Evolve(*State, Event) bool                    { return false }

// fakeMultiRole may be dealt to several players, like Villager and the plain
// Werewolf.
type fakeMultiRole struct {
	fakeRole
}

func (fakeMultiRole) AllowsMultipleCopies() bool { return true }

// fakeMinPlayersRole is a role with a minimum table size, like Detective and
// Cupid.
type fakeMinPlayersRole struct {
	fakeRole
	min int
}

func (r fakeMinPlayersRole) MinPlayers() int { return r.min }

// fakeCurseRole turns a villager into a wolf, like the cursed wolf.
type fakeCurseRole struct {
	fakeRole
}

func (fakeCurseRole) CursesVillager() bool { return true }

func init() {
	Register(fakeMultiRole{fakeRole{id: roleWerewolf, team: TeamWolves, call: &NightCall{Order: 4}}})
	Register(fakeCurseRole{fakeRole{id: roleCursedWolf, team: TeamWolves, call: &NightCall{Order: 3, Nights: []int{1, 2, 3}}}})
	Register(fakeMultiRole{fakeRole{id: roleVillager, team: TeamVillage}})
	Register(fakeRole{id: roleSeer, team: TeamVillage, call: &NightCall{Order: 7}})
	Register(fakeMinPlayersRole{fakeRole{id: roleDetective, team: TeamVillage, call: &NightCall{Order: 8}}, 9})
	Register(fakeMinPlayersRole{fakeRole{id: roleCupid, team: TeamVillage, call: &NightCall{Order: 0, Nights: []int{1}}}, 10})
}
