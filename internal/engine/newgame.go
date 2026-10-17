package engine

import "fmt"

// NewGame starts another game in the same room once the previous one ended,
// by a win or by an abort vote (audit K3): the table does not have to
// recreate the room and have everyone rejoin for a second game. The room
// layer checks that the sender is the host; the engine only checks the phase.
type NewGame struct {
	CommandMarker
}

// NewGameStarted resets the match while keeping the room: the players, their
// seats and the configured role list survive; everything about the game that
// just ended (who is alive, who holds which role, the night/day state, the
// winner) is cleared so the host can deal again with a fresh seed.
type NewGameStarted struct {
	EventMarker
}

func decideNewGame(s State, _ NewGame) ([]Event, error) {
	if s.Phase != PhaseEnded {
		return nil, fmt.Errorf("%w: a new game can only start once the previous one ended", ErrWrongPhase)
	}
	return []Event{NewGameStarted{}}, nil
}

func evolveNewGame(s State, e Event) (State, bool) {
	if _, ok := e.(NewGameStarted); !ok {
		return s, false
	}
	out := s.clone()
	out.Phase = PhaseLobby
	out.Dealt = false
	out.Night, out.NightStep = 0, 0
	out.Pending = NightActions{}
	out.Day = DayState{}
	out.Lovers = nil
	out.Winner = nil
	out.Winners = nil
	out.Chief = ""
	out.ChiefVotes = nil
	out.Succession = Succession{}
	out.Shot = Shot{}
	for id, p := range out.Players {
		out.Players[id] = &Player{ID: id, Nickname: p.Nickname, Team: TeamNone, Alive: true, Flags: map[string]int{}}
	}
	out.Seq++
	return out, true
}
