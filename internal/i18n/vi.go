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
	NightLoversMeet: "Cặp đôi hãy mở mắt nhìn nhau.",
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

// RoleCard is the text of one role's card as printed (rules §9b): Team and
// Goal are the role's own side. A phone shows the player's current side
// instead (see Side). Every card has all four fields set.
type RoleCard struct {
	Name        string
	Team        string
	Goal        string
	Description string
}

// Team labels and goals of the cards (rules §9b "Mục tiêu theo phe", §8).
const (
	teamVillage = "Dân"
	teamWolves  = "Sói"
	teamThird   = "Phe thứ ba"
	goalVillage = "Dân thắng khi không còn con Sói nào."
	goalWolves  = "Sói thắng khi số Sói ≥ số Dân."
	goalThird   = "Cặp đôi khác phe thắng khi chỉ còn hai người họ sống."
)

// The sides a card can show.
const (
	SideVillage = "village"
	SideWolves  = "wolves"
	SideThird   = "third" // a cross-team couple
)

// Side returns the team label and the goal of a side; an unknown side returns
// the side itself twice.
func Side(side string) (team, goal string) {
	switch side {
	case SideVillage:
		return teamVillage, goalVillage
	case SideWolves:
		return teamWolves, goalWolves
	case SideThird:
		return teamThird, goalThird
	}
	return side, side
}

// roleCardOrder is the order of the rules §9b table; the public role list of a
// game follows it. Keys are the role IDs.
var roleCardOrder = []string{
	"Werewolf", "CursedWolf", "WolfSeer", "Villager", "Seer", "Guard", "Witch",
	"Hunter", "Visitor", "Detective", "RustyKnight", "Cupid",
}

var roleCards = map[string]RoleCard{
	"Werewolf": {Name: "Sói", Team: teamWolves, Goal: goalWolves,
		Description: "Mỗi đêm cùng bầy chọn một người để cắn, hoặc chọn không cắn ai. Ban ngày giả làm dân lành."},
	"CursedWolf": {Name: "Sói nguyền", Team: teamWolves, Goal: goalWolves,
		Description: "Trong 3 đêm đầu, được nguyền một người biến thành Sói (một lần duy nhất). Vẫn cắn cùng bầy."},
	"WolfSeer": {Name: "Sói tiên tri", Team: teamWolves, Goal: goalWolves,
		Description: "Mỗi đêm soi một người xem có phải Tiên tri không; cả bầy biết kết quả. Vẫn cắn cùng bầy."},
	"Villager": {Name: "Dân làng", Team: teamVillage, Goal: goalVillage,
		Description: "Không có năng lực ban đêm. Vũ khí của bạn là lá phiếu và khả năng suy luận ban ngày."},
	"Seer": {Name: "Tiên tri", Team: teamVillage, Goal: goalVillage,
		Description: "Mỗi đêm soi một người, biết người đó là Sói hay không."},
	"Guard": {Name: "Bảo vệ", Team: teamVillage, Goal: goalVillage,
		Description: "Mỗi đêm che một người khỏi bị Sói cắn. Được che chính mình, không che một người hai đêm liền."},
	"Witch": {Name: "Phù thủy", Team: teamVillage, Goal: goalVillage,
		Description: "Có 1 bình cứu và 1 bình độc cho cả ván. Khi còn bình cứu, bạn biết ai bị cắn và có thể cứu. Bình độc giết một người bất kỳ."},
	"Hunter": {Name: "Thợ săn", Team: teamVillage, Goal: goalVillage,
		Description: "Khi chết vì bất kỳ lý do gì, bạn lộ vai và bắt buộc bắn một người còn sống."},
	"Visitor": {Name: "Nguyệt Nữ", Team: teamVillage, Goal: goalVillage,
		Description: "Mỗi đêm bắt buộc ghé một người; người đó mất chức năng đêm đó. Không ghé một người hai đêm liền."},
	"Detective": {Name: "Thám tử", Team: teamVillage, Goal: goalVillage,
		Description: "Một lần duy nhất: chọn một người, biết trong vùng gồm người đó và hai người sống ngồi cạnh có Sói hay không."},
	"RustyKnight": {Name: "Hiệp sĩ kiếm gỉ", Team: teamVillage, Goal: goalVillage,
		Description: "Nếu bị Sói cắn chết, con Sói còn sống gần nhất bên trái bạn trúng độc và chết vào đêm sau."},
	"Cupid": {Name: "Cupid", Team: teamVillage, Goal: goalVillage,
		Description: "Đêm đầu ghép đôi hai người (được chọn chính mình). Một người chết thì người kia chết theo."},
}

// Card returns the card of a role by its ID.
func Card(role string) (RoleCard, bool) {
	c, ok := roleCards[role]
	return c, ok
}

// CardOrder lists the role IDs in the order of the rules §9b table.
func CardOrder() []string { return append([]string(nil), roleCardOrder...) }

// Kinds of the extra cards (rules §9b "Thẻ phụ").
const (
	StatusCursed      = "cursed"
	StatusLovers      = "lovers"
	StatusLoversCross = "lovers_cross"
	StatusChief       = "chief"
)

// statusCards: the name of each extra card and its text; %s is the partner.
var statusCards = map[string][2]string{
	StatusCursed:      {"Bị nguyền", "Từ giờ bạn thuộc phe Sói và dậy cắn cùng bầy, vẫn giữ năng lực cũ."},
	StatusLovers:      {"Cặp đôi cùng phe", "Bạn và %s yêu nhau. Một người chết thì người kia chết theo. Vẫn chơi cho phe của mình."},
	StatusLoversCross: {"Cặp đôi khác phe", "Bạn và %s thuộc phe thứ ba: chỉ thắng khi hai bạn là hai người cuối còn sống. Dân hay Sói thắng thì hai bạn đều thua."},
	StatusChief:       {"Trưởng làng", "Phiếu của bạn tính x2 khi đề cử và vote giết/tha. Khi chết, bạn chọn người kế nhiệm."},
}

// StatusCard returns the name and text of an extra card; partner fills the
// lovers' cards. An unknown kind returns the kind itself.
func StatusCard(kind, partner string) (name, text string) {
	c, ok := statusCards[kind]
	if !ok {
		return kind, kind
	}
	return c[0], strings.Replace(c[1], "%s", partner, 1)
}
