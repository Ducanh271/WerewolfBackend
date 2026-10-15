package engine

import "math/rand"

// tallyChief picks the winner of the election (rules §7b): the most votes
// wins; a tie is settled at random among the tied players, and if nobody voted
// the pick is random among all players. The only randomness is the seed.
func tallyChief(s State, seed int64) PlayerID {
	counts := map[PlayerID]int{}
	best := 0
	for _, cand := range s.ChiefVotes {
		counts[cand]++
		if counts[cand] > best {
			best = counts[cand]
		}
	}
	// Candidates in seating order so the same seed always gives the same pick.
	var pool []PlayerID
	for _, id := range s.Seats {
		if best == 0 || counts[id] == best {
			pool = append(pool, id)
		}
	}
	if len(pool) == 1 {
		return pool[0]
	}
	return pool[rand.New(rand.NewSource(seed)).Intn(len(pool))]
}
