package engine

// Audience says who may know about an event. An event nobody is named for is
// seen only by full views (dead players, and everybody after the game).
type Audience struct {
	Everyone bool       // every player
	Wolves   bool       // the living wolf team at the moment of the event
	Players  []PlayerID // these players
}

// Audienced is implemented by role events that some living viewer may see,
// e.g. the Seer's result. s is the state just before the event. Role events
// without it stay hidden: forgetting it can hide information, never leak it.
type Audienced interface {
	Audience(s State) Audience
}

// Visible returns the events of log the viewer may know about, in order. s is
// the current state; log is every event of the game, from the first join.
// Full viewers get the whole log. Everybody else gets each event whose
// audience, judged at the time of the event, included them: a player cursed on
// night 2 sees the pack's votes from then on, not those of night 1.
func Visible(s State, log []Event, v Viewer) []Event {
	if seesAll(s, v) {
		return append([]Event(nil), log...)
	}
	var out []Event
	at := NewState()
	for _, e := range log {
		if audienceOf(at, e).includes(at, v) {
			out = append(out, e)
		}
		at = Evolve(at, e)
	}
	return out
}

func (a Audience) includes(s State, v Viewer) bool {
	if a.Everyone {
		return true
	}
	if a.Wolves {
		if p := s.Players[v.Player]; p != nil && p.Alive && VisibleTeam(p) == TeamWolves {
			return true
		}
	}
	for _, id := range a.Players {
		if id == v.Player {
			return true
		}
	}
	return false
}

// audienceOf lists the audience of the engine's own events. PlayerDied is
// hidden on purpose: the deaths of a night come in pipeline order, which would
// tell who was bitten and who was poisoned. The living learn of deaths from
// the announcements and the alive flags of the view. RolesDealt is hidden too:
// each player reads their own role from the view.
func audienceOf(s State, e Event) Audience {
	switch ev := e.(type) {
	case PlayerJoined, PlayerLeft, PlayerKicked, SeatsSet, TimersConfigured, RolesConfigured,
		ChiefElectionStarted, ChiefVoteCast, ChiefElected,
		ChiefSuccessionStarted, ChiefSucceeded, ChiefLost,
		NightStarted, NightEnded, DawnAnnounced,
		DayStarted, DayEnded, NominationCast, SkipSet,
		NominationsLocking, NominationsReopened, DayClockExpired,
		TrialStarted, VerdictStarted, VerdictCast, Spared, Hanged,
		HunterRevealed, HunterShotStarted, HunterShot, HunterTimedOut,
		GameEnded, NightStepCalled, PhoneCheckAnnounced, LoversMeetAnnounced,
		AbortVoteCast, NewGameStarted:
		return Audience{Everyone: true}
	case PhoneChecked:
		return Audience{Players: []PlayerID{ev.Player}}
	case TeamChanged:
		return Audience{} // the phone check tells the player and the pack
	case Audienced:
		return ev.Audience(s)
	}
	return Audience{}
}
