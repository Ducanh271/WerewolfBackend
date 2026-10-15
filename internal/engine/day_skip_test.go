package engine

import "testing"

func TestDay_SkipThreshold_IsMoreThanSixtySixPercentOfTheLiving(t *testing.T) {
	tests := []struct {
		living, skips int
		ends          bool
	}{
		{6, 3, false},    // 50%
		{6, 4, true},     // 66.7%
		{5, 3, false},    // 60%
		{5, 4, true},     // 80%
		{3, 2, true},     // 66.7%
		{9, 5, false},    // 55.6%
		{9, 6, true},     // 66.7%
		{15, 10, true},   // 66.7%
		{100, 66, false}, // exactly 66% is not more than 66%
		{100, 67, true},
	}
	for _, tc := range tests {
		s := NewState()
		s.Phase = PhaseDay
		var voters []PlayerID
		for i := 0; i < tc.living; i++ {
			id := PlayerID(string(rune('a' + i)))
			s.Players[id] = &Player{ID: id, Alive: true}
			voters = append(voters, id)
		}
		for _, id := range voters[:tc.skips-1] {
			s = Fold(s, SkipSet{Voter: id, On: true})
		}
		evs, err := Decide(s, SkipDay{Voter: voters[tc.skips-1], On: true})
		if err != nil {
			t.Fatalf("%d of %d: %v", tc.skips, tc.living, err)
		}
		_, ended := evs[len(evs)-1].(DayEnded)
		if ended != tc.ends {
			t.Errorf("%d skips of %d living: day ends = %v, want %v", tc.skips, tc.living, ended, tc.ends)
		}
	}
}
