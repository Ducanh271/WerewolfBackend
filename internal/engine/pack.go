package engine

// WolfVoters lists the wolves who may vote on the bite tonight: every living
// player of the wolf team who is not blocked and not doomed by the Rusty
// knight's poison. The count of rules §5 ("quá nửa
// số phiếu") is always taken over this list, so a blocked wolf neither votes
// nor counts.
func WolfVoters(s State) []PlayerID {
	var voters []PlayerID
	for _, id := range s.Seats {
		if p := s.Players[id]; p != nil && p.Alive && p.Team == TeamWolves && !Blocked(s, id) && !Doomed(s, id) {
			voters = append(voters, id)
		}
	}
	return voters
}

// NoBite is the pack's explicit "Không cắn ai" option. It is counted like a
// target in WolfVotes (rules §5).
const NoBite PlayerID = ""

// BiteTarget returns the player the pack bites: the one with more than half of
// the voters' votes. Votes of players who cannot vote tonight do not count.
// ok is false when nobody has more than half, or when NoBite has, so nobody is
// bitten.
func BiteTarget(s State, voters []PlayerID) (target PlayerID, ok bool) {
	counts := map[PlayerID]int{}
	for _, wolf := range voters {
		if t, voted := s.Pending.WolfVotes[wolf]; voted {
			counts[t]++
		}
	}
	for t, n := range counts {
		if 2*n > len(voters) {
			return t, t != NoBite
		}
	}
	return "", false
}
