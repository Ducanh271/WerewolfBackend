package roles_test

import (
	"testing"

	. "masoi/internal/engine/enginetest"
)

// tenPlayers is the usual table of these scenarios: three wolves (An, Binh,
// Chi), the Guard (Dung), the Seer (Em) and five villagers.
func tenPlayers() []Option {
	return []Option{
		Seats("An", "Binh", "Chi", "Dung", "Em", "Phuc", "Giang", "Hoa", "Khoa", "Lan"),
		Cast("An", Werewolf, "Binh", Werewolf, "Chi", Werewolf, "Dung", Guard, "Em", Seer,
			"Phuc", Villager, "Giang", Villager, "Hoa", Villager, "Khoa", Villager, "Lan", Villager),
	}
}

func newTable(t testing.TB) *Game { return NewGame(t, tenPlayers()...) }
