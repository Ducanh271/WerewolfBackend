package engine

import (
	"fmt"
	"reflect"
	"sort"
)

// eventTypes maps an event's Go type name to its type. It is what lets a
// stored log be read back (internal/store): the engine's own events are added
// below, each role adds its own from its init with RegisterEvents.
var eventTypes = map[string]reflect.Type{}

// RegisterEvents adds event types to the registry, by type name. It panics on
// a pointer, an unnamed type or a duplicate name, which can only be a
// programming error at startup.
func RegisterEvents(evs ...Event) {
	for _, e := range evs {
		t := reflect.TypeOf(e)
		if t == nil || t.Kind() != reflect.Struct || t.Name() == "" {
			panic(fmt.Sprintf("engine: RegisterEvents(%T): events are named struct values", e))
		}
		if old, dup := eventTypes[t.Name()]; dup && old != t {
			panic(fmt.Sprintf("engine: two events named %q (%v, %v)", t.Name(), old, t))
		}
		eventTypes[t.Name()] = t
	}
}

// EventType returns the registered event type with that name.
func EventType(name string) (reflect.Type, bool) {
	t, ok := eventTypes[name]
	return t, ok
}

// EventTypes returns one zero value of every registered event, sorted by name.
func EventTypes() []Event {
	names := make([]string, 0, len(eventTypes))
	for n := range eventTypes {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]Event, 0, len(names))
	for _, n := range names {
		out = append(out, reflect.New(eventTypes[n]).Elem().Interface().(Event))
	}
	return out
}

func init() {
	RegisterEvents(
		AbortVoteCast{}, ChatSent{}, ChiefElected{}, ChiefElectionStarted{}, ChiefLost{},
		ChiefSucceeded{}, ChiefSuccessionStarted{}, ChiefVoteCast{}, DawnAnnounced{},
		DawnStarted{}, DayClockExpired{}, DayEnded{}, DayStarted{}, GameEnded{},
		Hanged{}, HunterChainNoted{}, HunterRevealed{}, HunterShot{},
		HunterShotStarted{}, HunterTimedOut{}, LoverGrieving{}, LoversMeetAnnounced{},
		ModeConfigured{},
		NewGameStarted{}, NightEnded{}, NightStarted{}, NightStepCalled{},
		NominationCast{}, NominationsLocking{}, NominationsReopened{},
		PhoneCheckAnnounced{}, PhoneChecked{}, PlayerDied{}, PlayerJoined{},
		PlayerKicked{}, PlayerLeft{}, RolesConfigured{}, RolesDealt{},
		RoleViewStarted{}, SeatsSet{}, SkipSet{}, SoundConfigured{}, Spared{},
		TeamChanged{}, TimersConfigured{}, TrialStarted{}, VerdictCast{},
		VerdictStarted{}, WolfPoisoned{},
	)
}
