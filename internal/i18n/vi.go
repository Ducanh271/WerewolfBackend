package i18n

import "strings"

// vi maps announcement keys to the Vietnamese text read aloud. %s is replaced
// by the names, if the message takes any.
var vi = map[string]string{
	DawnSomeoneDied: "Đêm qua, %s đã chết.",
	DawnProtected:   "Không có ai chết, được bảo vệ.",
	DawnNobodyDied:  "Không có ai chết.",
	ChiefElected:    "%s đã được bầu làm Trưởng làng.",
	ChiefSucceeded:  "Trưởng làng mới là %s.",
	DayHanged:       "%s đã bị treo.",
	DaySpared:       "Làng tha cho %s.",
	DayNobodyHanged: "Không ai bị treo.",
	ChiefLost:       "Trưởng làng không kịp chọn người kế nhiệm, danh hiệu đã mất.",
	HunterRevealed:  "%s là Thợ săn.",
	HunterShot:      "%s bắn %s.",
	NightBlocked:    "Bạn không thể sử dụng chức năng đêm nay.",
	NightWolfDoomed: "Bạn đã trúng độc và chết từ lúc này, bạn không được vote.",
	NightPackDoomed: "%s không thể vote đêm nay.",
	NightGrief:      "Người yêu của bạn đã chết, bạn đau buồn mà chết từ lúc này và không dùng được chức năng.",
	NoticeNoLove:    "Bạn không ở trong cặp đôi.",
	NoticeInLove:    "Bạn yêu %s, vai %s.",
	NightPhoneCheck: "Mọi người mở mắt xem điện thoại, rồi nhắm mắt lại.",
	NoticeNotCursed: "Bạn không bị nguyền.",
	NoticeCursed:    "Bạn đã bị nguyền.",
	WinVillage:      "Phe Dân chiến thắng, tất cả bầy sói đã bị diệt.",
	WinWolves:       "Phe Sói chiến thắng.",
	WinLovers:       "Cặp đôi khác phe chiến thắng.",
	WinDraw:         "Tất cả đã chết, ván đấu hòa.",
}

// Vi returns the Vietnamese text for key with names joined in the order given.
// An unknown key returns the key itself so a missing string is easy to spot.
func Vi(key string, names ...string) string {
	text, ok := vi[key]
	if !ok {
		return key
	}
	if strings.Count(text, "%s") > 1 { // one name per slot
		for _, name := range names {
			text = strings.Replace(text, "%s", name, 1)
		}
		return text
	}
	return strings.Replace(text, "%s", strings.Join(names, ", "), 1)
}
