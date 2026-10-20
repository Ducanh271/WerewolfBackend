package i18n

import "strings"

// vi maps announcement keys to the Vietnamese text read aloud. %s is replaced
// by the names, if the message takes any.
var vi = map[string]string{
	RoleViewStarted:      "Mời mọi người xem vai của mình. Nhớ che màn hình nhé.",
	ChiefElectionStarted: "Mời cả làng bầu Trưởng làng.",
	NightStarted:         "Trời đã tối, tất cả nhắm mắt lại.",
	NightStepCalled:      "%s hãy mở mắt, thực hiện xong thì nhắm mắt lại.",
	DawnStarted:          "Trời đã sáng, mọi người mở mắt.",
	DawnSomeoneDied:      "Đêm qua, %s đã chết.",
	DawnProtected:        "Đêm qua không ai chết, có người đã được bảo vệ.",
	DawnNobodyDied:       "Không có ai chết.",
	DayStarted:           "Cả làng bắt đầu thảo luận và đề cử.",
	TrialStarted:         "%s lên giàn bào chữa.",
	VerdictStarted:       "Hết giờ bào chữa. Cả làng bỏ phiếu: treo hay tha.",
	ChiefElected:         "%s được bầu làm Trưởng làng.",
	ChiefSucceeded:       "Trưởng làng mới là %s.",
	DayHanged:            "%s đã bị treo.",
	DaySpared:            "Làng tha cho %s.",
	DayNobodyHanged:      "Không ai bị treo.",
	ChiefLost:            "Trưởng làng không kịp chọn người kế nhiệm, danh hiệu đã mất.",
	HunterRevealed:       "%s là Thợ săn.",
	HunterShot:           "%s bắn %s.",
	HunterTimedOut:       "Hết giờ, %s không bắn ai.",
	NightBlocked:         "Bạn không thể sử dụng chức năng đêm nay.",
	NightWolfDoomed:      "Bạn đã trúng độc và chết từ lúc này, bạn không được vote.",
	NightPackDoomed:      "%s không thể vote đêm nay.",
	NightGrief:           "Người yêu của bạn đã chết, bạn đau buồn mà chết từ lúc này và không dùng được chức năng.",
	NoticeNoLove:         "Bạn không ở trong cặp đôi.",
	NoticeInLove:         "Bạn yêu %s, vai %s.",
	NightPhoneCheck:      "Mọi người mở mắt xem điện thoại, rồi nhắm mắt lại.",
	NightLoversMeet:      "Cặp đôi hãy mở mắt nhìn nhau.",
	NoticeNotCursed:      "Bạn không bị nguyền.",
	NoticeCursed:         "Bạn đã bị nguyền.",
	WinVillage:           "Phe Dân chiến thắng, tất cả bầy sói đã bị diệt.",
	WinWolves:            "Phe Sói chiến thắng.",
	WinLovers:            "Cặp đôi khác phe chiến thắng.",
	WinDraw:              "Tất cả đã chết, ván đấu hòa.",
	GameAborted:          "Bàn đã bỏ phiếu huỷ ván. Ván kết thúc, không có phe thắng.",

	SummaryBite:            "Sói cắn %s",
	SummaryHeal:            "Phù thủy cứu %s",
	SummaryWitchPoison:     "Phù thủy đầu độc %s",
	SummaryGuardProtect:    "Bảo vệ che chở %s",
	SummarySeerWolf:        "Tiên tri soi %s: Sói",
	SummarySeerVillager:    "Tiên tri soi %s: không phải Sói",
	SummaryWolfSeerIsSeer:  "Sói tiên tri soi %s: Tiên tri",
	SummaryWolfSeerNotSeer: "Sói tiên tri soi %s: không phải Tiên tri",
	SummaryDetectiveWolf:   "Thám tử dò %s: có Sói",
	SummaryDetectiveClear:  "Thám tử dò %s: không có Sói",
	SummaryVisitorBlock:    "Nguyệt Nữ chặn %s",
	SummaryCurse:           "Sói nguyền nguyền %s",
	SummaryKnightPoison:    "Hiệp sĩ kiếm gỉ đầu độc %s",
	SummaryCupidPair:       "Cupid ghép đôi %s và %s",
	SummaryNightLabel:      "Đêm %s",
	SummaryDayLabel:        "Ngày %s",
	SummaryDawnDeaths:      "%s chết",
	SummaryDawnNobodyDied:  "Không ai chết",
	SummaryAborted:         "Bàn bỏ phiếu huỷ ván. Không có phe thắng.",

	// Error texts (errors.go): "error.<code>" and "error.<code>.<situation>".
	ErrorGeneric:                     "Có lỗi xảy ra, bạn thử lại hoặc mở lại ứng dụng nhé.",
	"error.room_not_found":           "Không tìm thấy phòng này, bạn kiểm tra lại mã phòng nhé.",
	"error.room_closed":              "Phòng này đã đóng, bạn tạo phòng mới nhé.",
	"error.invalid_nickname":         "Tên chỉ gồm chữ, số và dấu cách, dài từ 1 đến 20 ký tự.",
	"error.nickname_taken":           "Trong phòng đã có người tên này, bạn chọn tên khác nhé.",
	"error.room_full":                "Phòng đã đủ người, không vào thêm được.",
	"error.wrong_phase.join":         "Ván đã bắt đầu nên không vào mới được; nếu bạn đã ở trong ván, hãy vào lại bằng tên và PIN.",
	"error.bad_credentials":          "Tên hoặc PIN không đúng, bạn kiểm tra lại nhé.",
	"error.bad_credentials.token":    "Máy này không còn vào được phòng, bạn vào lại bằng tên và PIN nhé.",
	"error.pin_locked":               "Bạn nhập sai PIN quá nhiều lần, thử lại sau %s giây nhé.",
	"error.kicked_by_host":           "Chủ phòng đã mời bạn ra khỏi phòng.",
	"error.replaced_by_other_device": "Bạn vừa vào phòng từ một máy khác, máy này đã thoát.",
	"error.too_many_rooms":           "Máy chủ đang bận, bạn thử tạo phòng lại sau ít phút nhé.",
	"error.too_many_connections":     "Máy chủ đang bận, bạn thử lại sau ít phút nhé.",
	"error.rate_limited":             "Bạn bấm nhanh quá, chờ một chút nhé.",
	"error.not_host":                 "Chỉ chủ phòng mới làm được việc này.",
	"error.unknown_player":           "Người này không còn ở trong phòng.",
	"error.seats_mismatch":           "Danh sách chỗ ngồi chưa khớp với người trong phòng, bạn xếp lại nhé.",
	"error.seats_not_set":            "Còn người chưa được xếp chỗ, bạn xếp chỗ trước khi chia vai nhé.",
	"error.roles_not_configured":     "Bạn chưa chọn vai cho ván này.",
	"error.timer_out_of_range":       "Thời gian này nằm ngoài giới hạn cho phép, bạn chọn lại nhé.",
	"error.too_few_players":          "Cần ít nhất 5 người mới chơi được.",
	"error.role_count_mismatch":      "Số vai phải bằng đúng số người trong phòng.",
	"error.no_wolves":                "Ván cần ít nhất một Sói.",
	"error.too_many_wolves":          "Số Sói vượt giới hạn cho số người này, tính cả người sẽ bị nguyền.",
	"error.role_needs_more_players":  "Có vai cần thêm người mới dùng được: Thám tử cần trên 8 người, Cupid cần từ 10 người.",
	"error.role_required":            "Ván cần ít nhất một Dân làng.",
	"error.already_dealt":            "Vai đã được chia rồi.",
	"error.wrong_phase":              "Bước này đã qua rồi, bạn không làm được nữa.",
	"error.not_your_turn":            "Chưa tới lượt bạn.",
	"error.already_acted":            "Bạn đã chọn rồi, không đổi được nữa.",
	"error.ability_used":             "Bạn đã dùng chức năng này rồi, mỗi ván chỉ được một lần.",
	"error.potion_used":              "Bình thuốc này đã dùng hết rồi.",
	"error.nothing_to_heal":          "Đêm nay không có ai bị cắn để cứu.",
	"error.same_guard_target":        "Bạn không được bảo vệ cùng một người hai đêm liền.",
	"error.same_visitor_target":      "Bạn không được ghé cùng một người hai đêm liền.",
	"error.invalid_target":           "Không chọn được người này, bạn chọn người khác nhé.",
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

// nightCallNames are the names read aloud when a role's turn comes at night
// (rules §3, NightStepCalled): the same as the role card's name, except the
// wolf pack step, always called "Bầy sói" even when the only wolf role dealt
// is a lone Werewolf (whose card says "Sói").
var nightCallNames = map[string]string{
	"Cupid":      "Cupid",
	"Visitor":    "Nguyệt Nữ",
	"Guard":      "Bảo vệ",
	"CursedWolf": "Sói nguyền",
	"WolfSeer":   "Sói tiên tri",
	"Werewolf":   "Bầy sói",
	"Witch":      "Phù thủy",
	"Seer":       "Tiên tri",
	"Detective":  "Thám tử",
}

// NightCallName is the name read aloud for role's turn at night. ok is false
// for a night step that belongs to no role (the phone check, the lovers
// looking at each other): those read their own voice line instead, so
// NightStepCalled gets no "say" for them (encode.go).
func NightCallName(role string) (name string, ok bool) {
	name, ok = nightCallNames[role]
	return name, ok
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

// roleCards is the card text of each role, keyed by role ID. The order of the
// rules §9b table is the role's own (engine.CardOrder).
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
