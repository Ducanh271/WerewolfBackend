package engine

import (
	"reflect"
	"testing"
)

func TestDawn_SeveralDeaths_NamesInSeatingOrderNotResolutionOrder(t *testing.T) {
	seats := []PlayerID{"a", "b", "c", "d", "e"}
	tests := []struct {
		name       string
		resolution []PlayerID // order in which the pipeline killed them
		want       []PlayerID // order the narrator reads
	}{
		{"bitten seated after poisoned", []PlayerID{"d", "b"}, []PlayerID{"b", "d"}},
		{"bitten seated before poisoned", []PlayerID{"b", "d"}, []PlayerID{"b", "d"}},
		{"three deaths", []PlayerID{"e", "a", "c"}, []PlayerID{"a", "c", "e"}},
		{"one death", []PlayerID{"c"}, []PlayerID{"c"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := State{Seats: seats}
			ev := announceDawn(&Resolution{State: s, Deaths: tc.resolution})
			got := ev.(DawnAnnounced)
			if got.Kind != DawnDeaths || !reflect.DeepEqual(got.Dead, tc.want) {
				t.Fatalf("announced %v (kind %v), want %v", got.Dead, got.Kind, tc.want)
			}
		})
	}
}
