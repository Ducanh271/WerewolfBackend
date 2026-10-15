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
	return strings.Replace(text, "%s", strings.Join(names, ", "), 1)
}
