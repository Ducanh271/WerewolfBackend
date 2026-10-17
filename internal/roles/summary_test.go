package roles_test

import (
	"reflect"
	"testing"

	"masoi/internal/engine"

	. "masoi/internal/engine/enginetest"
)

// TestSummary_NarratesNightsAndDaysInVietnamese is the acceptance test of
// audit K4: the end-of-game summary reads like the table's own story, one
// line per night and day, in Vietnamese, from facts that stayed secret from
// living players during play (who the pack bit, who the Witch healed...).
func TestSummary_NarratesNightsAndDaysInVietnamese(t *testing.T) {
	g := NewGame(t, Seats(all...),
		Cast("An", Werewolf, "Binh", Werewolf, "Chi", Werewolf, "Dung", Guard, "Em", Seer,
			"Phuc", Villager, "Giang", Villager, "Hoa", Villager, "Khoa", Villager, "Lan", Witch),
		Chief("Hoa"))

	// Night 1: the Guard shields the wrong player, the pack bites Phuc, the
	// Witch heals him (saved), the Seer checks Dung (not a wolf).
	n := g.Night().Guard("Dung").Protects("Hoa")
	packBites(n, "Phuc").Witch("Lan").Heals()
	n.Seer("Em").Checks("Dung")
	g.Dawn().ExpectNobodyDied()
	g.Day().NobodyHanged()

	// Night 2: the pack bites Giang; nobody saves him.
	packBites(g.Night(), "Giang")
	g.Dawn().ExpectDead("Giang")

	// Day 2: the table votes to call the game off (audit K1): 6 of the 9
	// living players, exactly 2/3.
	g.Day()
	for _, voter := range []string{"An", "Binh", "Chi", "Dung", "Em", "Phuc"} {
		g.Do(engine.CastAbortVote{Voter: g.Player(voter), Yes: true})
	}
	g.ExpectPhase(engine.PhaseEnded)

	want := []string{
		"Hoa được bầu làm Trưởng làng.",
		"Đêm 1: Bảo vệ che chở Hoa. Sói cắn Phuc, Phù thủy cứu Phuc. Tiên tri soi Dung: không phải Sói. → Không ai chết.",
		"Ngày 1: Không ai bị treo.",
		"Đêm 2: Sói cắn Giang. → Giang chết.",
		"Bàn bỏ phiếu huỷ ván. Không có phe thắng.",
	}
	for _, who := range []string{"An", "Giang", "Khoa"} { // a wolf, the dead, a bystander: everybody sees it
		if got := g.As(who).View().Summary; !reflect.DeepEqual(got, want) {
			t.Fatalf("%s sees summary:\n%v\nwant:\n%v", who, got, want)
		}
	}
}
