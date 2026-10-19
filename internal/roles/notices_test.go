package roles_test

import (
	"testing"

	"masoi/internal/engine"

	. "masoi/internal/engine/enginetest"
)

// The phone shows these notices as text sent by the server (plan T30e); the
// wording is the approved one in i18n/vi.go. Rules §3, §6, §7.

func TestNotice_BlockedPlayer_IsToldInWords_NeverByWhom(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Em")
	n.Wolf("An").Bites("Phuc").TurnOf(Seer)

	g.As("Em").Turn(Seer).Blocked(true).TurnNotice("Bạn không thể sử dụng chức năng đêm nay.")
}

func TestNotice_CalledPlayerNotBlocked_GetsNoText(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Dung")
	n.Wolf("An").Bites("Phuc").TurnOf(Seer)

	g.As("Em").Turn(Seer).Blocked(false).TurnNotice("")
}

func TestNotice_PoisonedWolf_IsToldInWords_AndThePackToo(t *testing.T) {
	g := knightTable(t)
	bite(g.Night(), "Binh", "An", "Chi")
	g.Dawn()
	g.Day().NobodyHanged()
	g.As("Chi").TurnNotice("").PackDoomedText()

	g.Night().TurnOf(Werewolf)
	g.As("Chi").Doomed(true).TurnNotice("Bạn đã trúng độc và chết từ lúc này, bạn không được vote.").
		PackDoomedText("Chi không thể vote đêm nay.")
	g.As("An").Doomed(false).TurnNotice("").PackDoomedText("Chi không thể vote đêm nay.")
}

func TestNotice_GrievingLover_IsToldInWords(t *testing.T) {
	g := pairedDay1(t, nil, "An", "Chi")
	g.Day().Hanged("Chi")
	g.Night().TurnOf(CursedWolf)

	g.As("An").Grief(true).TurnNotice("Người yêu của bạn đã chết, bạn đau buồn mà chết từ lúc này và không dùng được chức năng.")
}

// The curse line is sent only when the role list has the Cursed wolf, the
// love line only when it has Cupid: the same for every phone, only the words
// differ.
func TestNotice_PhoneCheck_LinesFollowThePublicRoleList(t *testing.T) {
	const (
		notCursed = "Bạn không bị nguyền."
		cursed    = "Bạn đã bị nguyền."
		noLove    = "Bạn không ở trong cặp đôi."
	)

	t.Run("Cursed wolf only", func(t *testing.T) {
		g := cursedTable(t)
		g.Night().CursedWolf("An").Curses("Phuc").PhoneCheck()
		g.As("Phuc").PhoneNoticeText(cursed, "")
		g.As("Hoa").PhoneNoticeText(notCursed, "")
	})

	t.Run("Cupid only", func(t *testing.T) {
		g := cupidTableWith(t, map[string]engine.RoleID{"An": Werewolf})
		g.Night().Cupid("Lan").Pairs("Em", "Chi").PhoneCheck()
		g.As("Em").PhoneNoticeText("", "Bạn yêu Chi, vai Dân làng.")
		g.As("Chi").PhoneNoticeText("", "Bạn yêu Em, vai Tiên tri.")
		g.As("Hoa").PhoneNoticeText("", noLove)
	})

	t.Run("Cupid and Cursed wolf", func(t *testing.T) {
		g := cupidTable(t)
		n := g.Night().Cupid("Lan").Pairs("Em", "Chi").PhoneCheck()
		g.As("Em").PhoneNoticeText(notCursed, "Bạn yêu Chi, vai Dân làng.")
		g.As("Hoa").PhoneNoticeText(notCursed, noLove)

		n.CursedWolf("An").Curses("Em").PhoneCheck()
		g.As("Em").PhoneNoticeText(cursed, "Bạn yêu Chi, vai Dân làng.")
		g.As("Hoa").PhoneNoticeText(notCursed, noLove)
	})
}
