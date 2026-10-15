package engine

// afterDeaths is what happens once the deaths of a step have been announced.
// Every source of death (the night, the hanging, lovers in T19) calls it, so
// the order of consequences lives in one place (rules §9 steps 8 and 9, §7b):
//
//  1. every dead Hunter is revealed and shoots, one after the other; a shot
//     victim who is a Hunter shoots too (settle starts each shot, and
//     shotResolved, once the Hunter chose or ran out of time, comes back here);
//  2. by day, the surviving partner of a dead lover is marked to die in the
//     next night (LoverGrieving, rules §7); at night they died already (§9
//     step 7). They stay alive, vote and count for the win check until then;
//  3. the win check (rules §8), also when nobody died; a finished game skips
//     step 4;
//  4. a dead Chief names a successor.
//
// s is the state the step started from and emitted the events the step already
// produced (the deaths among them); deaths never change who is Chief.
func afterDeaths(s State, emitted []Event, dead []PlayerID) []Event {
	var shooters []PlayerID
	for _, id := range dead {
		if isDeathShooter(s.Players[id]) {
			shooters = append(shooters, id)
		}
	}
	return settle(s, emitted, shooters, dead)
}

// settle continues the aftermath: the first of the dead Hunters is revealed and
// gets his shot, or, with none left, the win check and the Chief follow. dead
// is everybody who died in the chain so far.
func settle(s State, emitted []Event, shooters, dead []PlayerID) []Event {
	after := Fold(s, emitted...)
	var evs []Event // reveals of Hunters with nobody left to shoot
	for len(shooters) > 0 {
		hunter := shooters[0]
		shooters = shooters[1:]
		reveal := HunterRevealed{Hunter: hunter}
		evs = append(evs, reveal)
		if !anyoneAlive(after) {
			continue
		}
		return append(evs,
			HunterChainNoted{Queue: shooters, Dead: dead},
			HunterShotStarted{Hunter: hunter, Duration: s.Settings.HunterShot})
	}
	evs = append(evs, griefMarks(after, dead)...)
	// The win check runs by day only, after the announcement and the shots, on
	// the situation of this moment (rules §8). A night without deaths still
	// gets it: a curse may have turned the numbers.
	if out, over := CheckWin(after); over {
		return append(evs, GameEnded{Outcome: out})
	}
	for _, id := range dead {
		if id == s.Chief {
			evs = append(evs, ChiefSuccessionStarted{Former: id, Duration: s.Settings.ChiefSuccession})
		}
	}
	return evs
}

func anyoneAlive(s State) bool {
	for _, p := range s.Players {
		if p.Alive {
			return true
		}
	}
	return false
}
