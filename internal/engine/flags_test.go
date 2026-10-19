package engine

import "testing"

// seatedTable is three seated players A, B, C on the given night, B dead.
func seatedTable(night int) State {
	s := NewState()
	s.Night = night
	s.Seats = []PlayerID{"A", "B", "C"}
	roles := map[PlayerID]RoleID{"A": roleWerewolf, "B": roleDetective, "C": roleVillager}
	for _, id := range s.Seats {
		s.Players[id] = &Player{ID: id, Role: roles[id], Alive: id != "B"}
	}
	return s
}

func TestSetFlag_CreatesTheMapAndOverwrites(t *testing.T) {
	p := &Player{ID: "A"}
	SetFlag(p, "x", 1)
	SetFlag(p, "x", 3)
	SetFlag(p, "y", 2)
	if p.Flags["x"] != 3 || p.Flags["y"] != 2 || len(p.Flags) != 2 {
		t.Fatalf("flags = %v, want x=3 y=2", p.Flags)
	}
}

func TestSeatNumber(t *testing.T) {
	s := seatedTable(1)
	tests := []struct {
		id   PlayerID
		want int
	}{{"A", 1}, {"C", 3}, {"Z", 0}, {"", 0}}
	for _, tc := range tests {
		if got := SeatNumber(s, tc.id); got != tc.want {
			t.Errorf("SeatNumber(%q) = %d, want %d", tc.id, got, tc.want)
		}
	}
}

func TestHolderOf(t *testing.T) {
	s := seatedTable(1)
	tests := []struct {
		name string
		role RoleID
		want PlayerID
	}{
		{"living holder", roleWerewolf, "A"},
		{"dead holder still holds the role", roleDetective, "B"},
		{"nobody holds it", roleWolfSeer, ""},
	}
	for _, tc := range tests {
		got := HolderOf(s, tc.role)
		if (got == nil && tc.want != "") || (got != nil && got.ID != tc.want) {
			t.Errorf("%s: HolderOf(%s) = %v, want %q", tc.name, tc.role, got, tc.want)
		}
	}
}

func TestLastTarget_SamePlayerOnlyRefusedTheNightAfter(t *testing.T) {
	last := LastTarget{Night: "t.night", Target: "t.target"}
	tests := []struct {
		name     string
		recorded int // night of the recorded choice of C, 0 for none
		tonight  int
		target   PlayerID
		want     bool
	}{
		{"no choice yet", 0, 1, "C", false},
		{"same player the night after", 1, 2, "C", true},
		{"another player the night after", 1, 2, "A", false},
		{"same player two nights later", 1, 3, "C", false},
	}
	for _, tc := range tests {
		s := seatedTable(tc.tonight)
		me := s.Players["A"]
		if tc.recorded > 0 {
			last.Record(s, me, tc.recorded, "C")
		}
		if got := last.Is(s, me, tc.target); got != tc.want {
			t.Errorf("%s: Is = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestLivingCount(t *testing.T) {
	s := seatedTable(1)
	if got := livingCount(s); got != 2 {
		t.Fatalf("livingCount = %d, want 2", got)
	}
	for _, p := range s.Players {
		p.Alive = false
	}
	if got := livingCount(s); got != 0 {
		t.Fatalf("livingCount with everybody dead = %d, want 0", got)
	}
}
