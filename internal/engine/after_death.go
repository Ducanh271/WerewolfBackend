package engine

// afterDeaths is what happens once the deaths of a step have been announced.
// Every source of death (the night here, the hanging in T7, lovers in T19)
// calls it, so the order of consequences lives in one place:
//
//  1. the Hunter shoots (T13 adds the step here, rules §7b: the Hunter shoots
//     before the Chief passes on the title);
//  2. the win check (rules §8, §9 step 9); a finished game skips step 3;
//  3. a dead Chief names a successor.
//
// s is the state the step started from and emitted the events the step already
// produced (the deaths among them); deaths never change who is Chief.
func afterDeaths(s State, emitted []Event, dead []PlayerID) []Event {
	if len(dead) > 0 {
		if out, over := CheckWin(Fold(s, emitted...)); over {
			return []Event{GameEnded{Outcome: out}}
		}
	}
	var evs []Event
	for _, id := range dead {
		if id == s.Chief {
			evs = append(evs, ChiefSuccessionStarted{Former: id, Duration: s.Settings.ChiefSuccession})
		}
	}
	return evs
}
