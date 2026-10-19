package i18n

import (
	"strconv"
	"strings"
	"time"
)

// GuideTimers carries the real settings and constants that the in-app guide
// (T25) quotes, so a number in the guide can never drift from the number the
// game actually uses. Named placeholders ({seconds}, {pinTries},
// {lockMinutes}, {minPlayers}) are filled from these fields instead of being
// typed into the guide text, and instead of a positional %d (which would
// break if a future language reordered the sentence).
type GuideTimers struct {
	RoleView            time.Duration // rules §2 "xem vai"
	HunterShot          time.Duration // rules §6 "Thợ săn"
	LoversMeet          time.Duration // rules §3, the lovers looking at each other
	ChiefSuccession     time.Duration // rules §7b, naming a successor
	PINMaxFailures      int           // room.MaxPINFailures
	PINLockTime         time.Duration // room.PINLockTime
	CupidMinPlayers     int           // Cupid's own MinPlayersRule
	DetectiveMinPlayers int           // Detective's own MinPlayersRule
}

// fillVars replaces every "{name}" in tmpl with its value from vars. A name
// with no entry is left as-is, so a missing var is easy to spot in the
// rendered text instead of silently vanishing.
func fillVars(tmpl string, vars map[string]string) string {
	if len(vars) == 0 {
		return tmpl
	}
	out := tmpl
	for name, value := range vars {
		out = strings.ReplaceAll(out, "{"+name+"}", value)
	}
	return out
}

func seconds(d time.Duration) string { return strconv.Itoa(int(d.Seconds())) }
func minutes(d time.Duration) string { return strconv.Itoa(int(d.Minutes())) }

// QuickRules is the "Luật nhanh" of guide-text.md §1: 12 lines, no settings
// number in any of them, read before the first game.
func QuickRules() []string {
	return []string{
		"Ma Sói chơi bằng điện thoại của chính mỗi người, không cần ai làm quản trò riêng.",
		"Có hai phe chính: Dân làng và Sói; có thể có thêm phe thứ ba là một cặp đôi khác phe.",
		"Mỗi đêm, các vai có chức năng được gọi lần lượt: mở mắt làm việc trên điện thoại rồi nhắm mắt lại.",
		"Bầy sói luôn được gọi mỗi đêm để chọn cắn một người.",
		"Buổi sáng, loa đọc ai đã chết đêm qua — không nêu vai, không nêu lý do chết.",
		"Ban ngày, cả làng thảo luận và đề cử một người lên giàn.",
		"Người bị đề cử nhiều nhất lên giàn bào chữa, rồi cả làng bỏ phiếu treo hay tha; hoà phiếu đề cử thì người được đề cử sớm nhất lên giàn.",
		"Dân thắng khi diệt hết Sói. Sói thắng khi số Sói còn sống bằng hoặc nhiều hơn số Dân còn sống.",
		"Cặp đôi khác phe chỉ thắng khi chính hai người đó là hai người cuối cùng còn sống.",
		"Trưởng làng là danh hiệu công khai (không phải vai riêng): phiếu tính gấp đôi, khi chết được chọn người kế nhiệm.",
		"Giữa ban ngày, người còn sống có thể bỏ phiếu huỷ ván nếu cả bàn muốn dừng sớm.",
		"Người đã chết không được nói hay bỏ phiếu nữa, nhưng được xem toàn bộ vai và diễn biến từ lúc đó.",
	}
}

// GuideEntry is the per-role or per-status how-to of guide-text.md §2: when
// it is called, how to use it on the phone, what happens on timeout, and one
// common mistake. Every entry has the same shape; Note is the only optional
// field (an extra fact that does not fit the other four).
type GuideEntry struct {
	ID         string // the role ID (engine.RoleID) or a status key (StatusChief, StatusLovers, StatusCursed)
	Title      string
	WhenCalled string
	HowTo      string
	OnTimeout  string
	Gotcha     string
	Note       string
}

// RoleGuides is guide-text.md §2 for every role, in the order of the rules
// §9b table (engine.CardOrder). Adding a role without adding its entry here
// fails TestGuide_EveryRegisteredRoleHasAnEntry (internal/roles/guide_test.go).
func RoleGuides(t GuideTimers) []GuideEntry {
	return []GuideEntry{
		{
			ID:         "Werewolf",
			Title:      "Sói",
			WhenCalled: "Được gọi mỗi đêm cùng cả bầy, kể cả đêm đầu tiên.",
			HowTo:      `Chọn một người sống để cắn, hoặc chọn "Không cắn ai"; thấy phiếu của cả bầy theo thời gian thực, đổi phiếu tuỳ ý cho tới hết giờ.`,
			OnTimeout:  `Tính phiếu lúc đó — người được quá nửa phiếu bị cắn; nếu "Không cắn ai" được quá nửa, hoặc không ai quá nửa, đêm đó không ai bị cắn.`,
			Gotcha:     "Đổi phiếu phút cuối rồi quên đổi lại — phiếu tính đúng lúc hết giờ, không phải phiếu lúc đầu.",
		},
		{
			ID:         "CursedWolf",
			Title:      "Sói nguyền",
			WhenCalled: "Được gọi riêng, ngay trước bầy sói, trong 3 đêm đầu (kể cả sau khi đã nguyền xong, để không ai đoán được mình đã dùng hay chưa).",
			HowTo:      "Chọn một người sống, không phải sói, để nguyền thành Sói — chỉ dùng được một lần cả ván.",
			OnTimeout:  "Lượt nguyền chưa mất, dùng được ở đêm sau (miễn còn trong 3 đêm đầu).",
			Gotcha:     "Để dành nguyền quá lâu — hết 3 đêm đầu mà chưa nguyền thì mất chức năng, thành sói thường.",
		},
		{
			ID:         "WolfSeer",
			Title:      "Sói tiên tri",
			WhenCalled: "Được gọi mỗi đêm, ngay trước bầy sói (sau Sói nguyền nếu ván có cả hai).",
			HowTo:      `Chọn một người để soi; kết quả chỉ là "là Tiên tri" hay "không phải Tiên tri", báo cho cả bầy ngay sau đó.`,
			OnTimeout:  "Đêm đó không soi ai, cả bầy không có thêm thông tin.",
			Gotcha:     "Nhầm với Tiên tri (vai phe Dân) vì tên gần giống — Sói tiên tri thuộc phe Sói, vẫn dậy cắn cùng bầy.",
		},
		{
			ID:         "Villager",
			Title:      "Dân làng",
			WhenCalled: "Không có lượt đêm.",
			HowTo:      "Không có hành động ban đêm; ban ngày dùng phiếu như mọi người còn sống (đề cử, bỏ phiếu treo/tha).",
			OnTimeout:  "Không áp dụng.",
			Gotcha:     `Cứ chờ "lượt của mình" ban đêm — Dân làng không có lượt, nhắm mắt ngủ cho tới sáng.`,
		},
		{
			ID:         "Seer",
			Title:      "Tiên tri",
			WhenCalled: "Được gọi mỗi đêm.",
			HowTo:      "Chọn một người để soi; biết người đó là Sói hay không phải Sói.",
			OnTimeout:  "Đêm đó không soi ai, không có kết quả.",
			Gotcha:     `Nghĩ soi ra được vai cụ thể — Tiên tri chỉ biết "là Sói" hay "không phải Sói", không biết vai gì.`,
		},
		{
			ID:         "Guard",
			Title:      "Bảo vệ",
			WhenCalled: "Được gọi mỗi đêm.",
			HowTo:      "Chọn một người để che khỏi bị sói cắn đêm đó (được chọn chính mình).",
			OnTimeout:  "Đêm đó không che ai.",
			Gotcha:     "Chọn lại đúng người đã che đêm trước — app không cho chọn, vì không được che cùng một người hai đêm liên tiếp.",
		},
		{
			ID:         "Witch",
			Title:      "Phù thủy",
			WhenCalled: "Được gọi mỗi đêm.",
			HowTo:      "Khi còn bình cứu, thấy ai bị sói cắn đêm đó và chọn cứu hay không (được tự cứu mình); luôn được chọn dùng bình độc giết một người sống bất kỳ, kể cả đêm đã cứu người.",
			OnTimeout:  "Không dùng bình nào đêm đó; bình chưa dùng vẫn còn cho đêm sau.",
			Gotcha:     "Tưởng chỉ được chọn cứu hoặc độc trong một đêm — được dùng cả hai bình cùng một đêm nếu muốn.",
		},
		{
			ID:         "Hunter",
			Title:      "Thợ săn",
			WhenCalled: "Không có lượt đêm; chỉ hoạt động ngay khi chết, vì bất kỳ lý do gì.",
			HowTo:      fillVars("Lúc chết, vai được công khai ngay; trong {seconds} giây, chọn một người sống bất kỳ để bắn.", map[string]string{"seconds": seconds(t.HunterShot)}),
			OnTimeout:  "Không bắn ai.",
			Gotcha:     "Nghĩ mình còn cơ hội bắn lần sau — Thợ săn chỉ bắn đúng một lần, ngay lúc chết.",
		},
		{
			ID:         "Visitor",
			Title:      "Nguyệt Nữ",
			WhenCalled: "Được gọi đầu tiên mỗi đêm (đêm 1 thì sau Cupid và cặp đôi).",
			HowTo:      `Bắt buộc chọn một người sống khác mình để "ghé"; người đó mất chức năng ban đêm đêm đó.`,
			OnTimeout:  "Server tự chọn ngẫu nhiên một người hợp lệ (không phải chính mình, không trùng người đêm trước).",
			Gotcha:     "Chọn lại đúng người đã ghé đêm trước — app không cho chọn.",
		},
		{
			ID:         "Detective",
			Title:      fillVars("Thám tử (ván từ {minPlayers} người)", map[string]string{"minPlayers": strconv.Itoa(t.DetectiveMinPlayers)}),
			WhenCalled: "Được gọi mỗi đêm, kể cả sau khi đã dùng (lượt gọi giả, vẫn chạy đủ giờ, để không lộ đã dùng).",
			HowTo:      "Chọn một người sống (được chọn chính mình); biết vùng gồm người đó và hai người sống ngồi gần nhất có Sói hay không — không biết số lượng hay tên.",
			OnTimeout:  "Khi chưa dùng: đêm đó bỏ qua, lượt vẫn còn cho đêm sau.",
			Gotcha:     `Nghĩ biết được số lượng hay tên sói trong vùng — app chỉ báo "có Sói" hay "không có Sói", không hơn.`,
		},
		{
			ID:         "RustyKnight",
			Title:      "Hiệp sĩ kiếm gỉ",
			WhenCalled: "Không có lượt đêm, năng lực bị động, tự kích hoạt.",
			HowTo:      "Không cần chạm gì; nếu bị sói cắn chết mà không được Bảo vệ che, tự động làm một con sói còn sống gần nhất bên trái trúng độc.",
			OnTimeout:  "Không áp dụng.",
			Gotcha:     "Tưởng có nút bấm ban đêm — Hiệp sĩ không làm gì cả, năng lực chỉ kích hoạt khi bị cắn chết.",
			Note:       "Con sói trúng độc không chết ngay đêm đó — nó sống thêm đúng 24 giờ và chết vào đêm sau; cái chết đó được loa đọc sáng hôm sau cùng các cái chết khác.",
		},
		{
			ID:         "Cupid",
			Title:      fillVars("Cupid (ván từ {minPlayers} người)", map[string]string{"minPlayers": strconv.Itoa(t.CupidMinPlayers)}),
			WhenCalled: "Được gọi đầu tiên, chỉ đêm đầu tiên.",
			HowTo:      "Chọn hai người sống khác nhau (được chọn chính mình) để thành một cặp đôi.",
			OnTimeout:  "Server tự ghép ngẫu nhiên hai người sống.",
			Gotcha:     "Nghĩ còn đêm sau để ghép đôi — Cupid chỉ được gọi đúng một lần, đêm đầu tiên.",
		},
	}
}

// StatusGuides is guide-text.md §2 for the three statuses that are not a
// role: Chief, Lovers and Cursed. They reuse the status keys already defined
// for StatusCard (vi.go) so the two never drift apart.
func StatusGuides(t GuideTimers) []GuideEntry {
	return []GuideEntry{
		{
			ID:         StatusChief,
			Title:      "Trưởng làng (danh hiệu công khai, không phải vai riêng)",
			WhenCalled: "Bầu ngay sau khi xem vai, trước đêm 1; giữ cho tới khi chết.",
			HowTo: fillVars(
				"Lúc bầu, chọn một người (được tự bầu mình); khi đang giữ danh hiệu, phiếu đề cử và phiếu giết/tha của bạn tự tính gấp đôi, không cần làm gì thêm; khi Trưởng làng chết, trong {seconds} giây chọn một người sống kế nhiệm.",
				map[string]string{"seconds": seconds(t.ChiefSuccession)}),
			OnTimeout: "Lúc bầu, không ai bầu cho ai thì server chọn ngẫu nhiên một người. Lúc chọn kế nhiệm, hết giờ thì danh hiệu mất, không ai thay.",
			Gotcha:    "Nghĩ vai gốc của mình đổi khi làm Trưởng làng — vai và phe gốc không đổi, chỉ có phiếu tính gấp đôi.",
			Note:      "Đang bị đề cử lên giàn thì mất quyền bỏ phiếu như mọi người khác, kể cả khi đang là Trưởng làng.",
		},
		{
			ID:    StatusLovers,
			Title: "Cặp đôi (trạng thái do Cupid ghép, không phải vai riêng)",
			WhenCalled: fillVars(
				"Không có lượt riêng; biết nhau ở bước Kiểm tra điện thoại ngay sau lượt Cupid, đêm 1; ngay sau đó mở mắt nhìn nhau {seconds} giây (chỉ hai người, người khác vẫn nhắm mắt).",
				map[string]string{"seconds": seconds(t.LoversMeet)}),
			HowTo:     "Không chọn gì; điện thoại hiện người yêu là ai, vai gì, và màu liên kết theo phe hiện tại của cặp.",
			OnTimeout: "Không áp dụng.",
			Gotcha: "Một người chết ban đêm thì người kia chết theo cùng đêm đó; một người chết ban ngày (bị treo, bị Thợ săn bắn) thì người kia chưa chết ngay, mà chết vào đêm kế tiếp. " +
				"Nếu cặp gồm một người phe Sói và một người phe Dân, hai người thuộc phe thứ ba: chỉ thắng khi chính hai người đó là hai người cuối cùng còn sống, dù phe Dân hay phe Sói thắng thì cả hai đều thua.",
		},
		{
			ID:         StatusCursed,
			Title:      "Bị nguyền (trạng thái do Sói nguyền gây ra, không phải vai riêng)",
			WhenCalled: "Không có lượt riêng; báo ở bước Kiểm tra điện thoại ngay sau lượt Sói nguyền, các đêm 1–3.",
			HowTo:      `Không chọn gì; điện thoại chỉ báo "Bạn đã bị nguyền" hoặc "Bạn không bị nguyền".`,
			OnTimeout:  "Không áp dụng.",
			Gotcha:     "Từ lúc bị nguyền, bạn thuộc phe Sói, dậy cắn cùng bầy và được tính là Sói khi xét thắng thua — nhưng vẫn giữ nguyên năng lực vai cũ (Tiên tri, Bảo vệ, Thợ săn... vẫn dùng được như trước).",
		},
	}
}

// FirstTimeChecklist is guide-text.md §3: what a first-time player must
// remember, not game rules.
func FirstTimeChecklist(t GuideTimers) []string {
	return []string{
		fillVars(
			"Che màn hình (tay, hoặc ngồi quay màn hình vào trong) mỗi khi điện thoại hiện vai của bạn (có {seconds} giây để xem) hoặc lượt đêm — người ngồi cạnh không được thấy.",
			map[string]string{"seconds": seconds(t.RoleView)}),
		"Thẻ vai che mặc định; nhấn giữ để xem, thả tay là ẩn lại ngay — đừng mở thẻ vai lâu để người khác liếc thấy.",
		"Nếu bạn đã chết, màn hình vẫn tối/im ban đêm cho tới khi bạn tự chạm vào xem — đừng để màn hình tự sáng, vì vậy sẽ lộ ra là bạn đã chết.",
		"Mỗi lượt gọi đêm chạy đủ thời gian như nhau dù vai đó còn sống hay đã chết — đừng đoán ai còn sống vì lượt nhanh hay chậm.",
		"Tắt âm bàn phím và rung khi chạm máy — tiếng gõ ban đêm lộ ra ai đang thao tác.",
		"Hết pin hoặc lỡ tắt app giữa ván: vào lại đúng phòng bằng cùng máy sẽ tự nhận lại vai ngay; đổi sang máy khác thì nhập đúng nickname và mã PIN 4 số được cấp lúc vào phòng.",
		fillVars(
			"Tự nhớ mã PIN của mình, không ai nhắc lại giúp — nhập sai PIN {pinTries} lần sẽ bị khoá vào lại bằng PIN trong {lockMinutes} phút.",
			map[string]string{"pinTries": strconv.Itoa(t.PINMaxFailures), "lockMinutes": minutes(t.PINLockTime)}),
		"Khi chết, bạn không được nói hay bỏ phiếu nữa, nhưng được xem toàn bộ vai thật và diễn biến của ván từ lúc đó.",
	}
}

// OptionalTips is guide-text.md §4: strategy tips, not rules. No number in
// any of them.
func OptionalTips() []string {
	return []string{
		"Tiên tri: cân nhắc trước khi nói huỵch toẹt kết quả soi ra ban ngày — nói sớm quá dễ lộ mình là Tiên tri.",
		"Thám tử: chỉ dùng được đúng một lần cả ván, có thể chờ thêm vài đêm để có manh mối trước khi chọn vùng dò.",
		"Phù thủy: không biết Bảo vệ đã che người bị cắn chưa, nên dùng bình cứu ngay đêm nào cũng có thể phí bình oan — cân nhắc trước khi cứu.",
		"Sói tiên tri: soi người khác nhau mỗi đêm thay vì soi lại người đã biết chắc, để tìm Tiên tri thật nhanh hơn.",
	}
}
