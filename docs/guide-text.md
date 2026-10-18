# Chữ cho hướng dẫn trong app

T25 bước 2: chữ tiếng Việt cho "hướng dẫn luật trong app" và "cách chơi theo
vai" (merge từ T28b). Công khai, giống nhau cho mọi người — không có vấn đề
bí mật, nhưng tránh gợi ý chiến thuật làm lộ cơ chế ẩn (ví dụ không nói
"Bảo vệ hay che chính mình đêm đầu" — đó là mẹo chơi, không phải luật).

Không viết code ở bước này. Chưa đụng `internal/i18n/vi.go`.

Nguồn: `docs/rules.md` (toàn bộ). Phần "Mô tả ngắn" của từng vai
(`internal/i18n/vi.go`, `roleCards`, đúng bảng §9b) **không viết lại** — mục 2
dưới đây là hướng dẫn thao tác trên điện thoại, bổ sung thêm cho thẻ vai sẵn
có, không thay thế nó.

**T25 bước 3b (đã triển khai, `internal/i18n/guide.go`):** vài câu ở mục 2 và
3 có một số thật (thời gian, số lần, số người) nay viết bằng placeholder có
tên (`{seconds}`, `{pinTries}`, `{lockMinutes}`, `{minPlayers}`) thay cho số
gõ cứng, để server điền từ `Settings`/hằng số thật — không gõ số trực tiếp
trong code, và không dùng `%d` vị trí (để sau thêm ngôn ngữ khác không bị vỡ
nếu câu đổi thứ tự). Mục 1 (Luật nhanh) và mục 4 (Mẹo chơi) không đổi — không
có số nào trong đó. Câu khoá đề cử 10 giây: giữ nguyên như đã duyệt, không
thêm số.

---

## 1. "Luật nhanh" (≤ 1 phút đọc, 12 dòng)

Dành cho người chưa biết luật, đọc trên điện thoại trước khi vào ván.

1. Ma Sói chơi bằng điện thoại của chính mỗi người, không cần ai làm quản trò riêng.
2. Có hai phe chính: Dân làng và Sói; có thể có thêm phe thứ ba là một cặp đôi khác phe.
3. Mỗi đêm, các vai có chức năng được gọi lần lượt: mở mắt làm việc trên điện thoại rồi nhắm mắt lại.
4. Bầy sói luôn được gọi mỗi đêm để chọn cắn một người.
5. Buổi sáng, loa đọc ai đã chết đêm qua — không nêu vai, không nêu lý do chết.
6. Ban ngày, cả làng thảo luận và đề cử một người lên giàn.
7. Người bị đề cử nhiều nhất lên giàn bào chữa, rồi cả làng bỏ phiếu treo hay tha; hoà phiếu đề cử thì người được đề cử sớm nhất lên giàn.
8. Dân thắng khi diệt hết Sói. Sói thắng khi số Sói còn sống bằng hoặc nhiều hơn số Dân còn sống.
9. Cặp đôi khác phe chỉ thắng khi chính hai người đó là hai người cuối cùng còn sống.
10. Trưởng làng là danh hiệu công khai (không phải vai riêng): phiếu tính gấp đôi, khi chết được chọn người kế nhiệm.
11. Giữa ban ngày, người còn sống có thể bỏ phiếu huỷ ván nếu cả bàn muốn dừng sớm.
12. Người đã chết không được nói hay bỏ phiếu nữa, nhưng được xem toàn bộ vai và diễn biến từ lúc đó.

---

## 2. Cách chơi trên điện thoại, theo vai

Mỗi vai: khi nào được gọi, cách dùng trên điện thoại, nếu hết giờ chưa làm
thì sao, và một điều dễ nhầm. Vai không có lượt đêm chỉ ghi "không có lượt
đêm".

### Sói (rules.md §5)
- Khi nào được gọi: được gọi mỗi đêm cùng cả bầy, kể cả đêm đầu tiên.
- Cách dùng: chọn một người sống để cắn, hoặc chọn "Không cắn ai"; thấy phiếu của cả bầy theo thời gian thực, đổi phiếu tuỳ ý cho tới hết giờ.
- Nếu hết giờ: tính phiếu lúc đó — người được quá nửa phiếu bị cắn; nếu "Không cắn ai" được quá nửa, hoặc không ai quá nửa, đêm đó không ai bị cắn.
- Dễ nhầm: đổi phiếu phút cuối rồi quên đổi lại — phiếu tính đúng lúc hết giờ, không phải phiếu lúc đầu.

### Sói nguyền (rules.md §5)
- Khi nào được gọi: được gọi riêng, ngay trước bầy sói, trong 3 đêm đầu (kể cả sau khi đã nguyền xong, để không ai đoán được mình đã dùng hay chưa).
- Cách dùng: chọn một người sống, không phải sói, để nguyền thành Sói — chỉ dùng được một lần cả ván.
- Nếu hết giờ: lượt nguyền chưa mất, dùng được ở đêm sau (miễn còn trong 3 đêm đầu).
- Dễ nhầm: để dành nguyền quá lâu — hết 3 đêm đầu mà chưa nguyền thì mất chức năng, thành sói thường.

### Sói tiên tri (rules.md §5)
- Khi nào được gọi: được gọi mỗi đêm, ngay trước bầy sói (sau Sói nguyền nếu ván có cả hai).
- Cách dùng: chọn một người để soi; kết quả chỉ là "là Tiên tri" hay "không phải Tiên tri", báo cho cả bầy ngay sau đó.
- Nếu hết giờ: đêm đó không soi ai, cả bầy không có thêm thông tin.
- Dễ nhầm: nhầm với Tiên tri (vai phe Dân) vì tên gần giống — Sói tiên tri thuộc phe Sói, vẫn dậy cắn cùng bầy.

### Dân làng (rules.md §6)
- Khi nào được gọi: không có lượt đêm.
- Cách dùng: không có hành động ban đêm; ban ngày dùng phiếu như mọi người còn sống (đề cử, bỏ phiếu treo/tha).
- Nếu hết giờ: không áp dụng.
- Dễ nhầm: cứ chờ "lượt của mình" ban đêm — Dân làng không có lượt, nhắm mắt ngủ cho tới sáng.

### Tiên tri (rules.md §6)
- Khi nào được gọi: được gọi mỗi đêm.
- Cách dùng: chọn một người để soi; biết người đó là Sói hay không phải Sói.
- Nếu hết giờ: đêm đó không soi ai, không có kết quả.
- Dễ nhầm: nghĩ soi ra được vai cụ thể — Tiên tri chỉ biết "là Sói" hay "không phải Sói", không biết vai gì.

### Bảo vệ (rules.md §6)
- Khi nào được gọi: được gọi mỗi đêm.
- Cách dùng: chọn một người để che khỏi bị sói cắn đêm đó (được chọn chính mình).
- Nếu hết giờ: đêm đó không che ai.
- Dễ nhầm: chọn lại đúng người đã che đêm trước — app không cho chọn, vì không được che cùng một người hai đêm liên tiếp.

### Phù thủy (rules.md §6)
- Khi nào được gọi: được gọi mỗi đêm.
- Cách dùng: khi còn bình cứu, thấy ai bị sói cắn đêm đó và chọn cứu hay không (được tự cứu mình); luôn được chọn dùng bình độc giết một người sống bất kỳ, kể cả đêm đã cứu người.
- Nếu hết giờ: không dùng bình nào đêm đó; bình chưa dùng vẫn còn cho đêm sau.
- Dễ nhầm: tưởng chỉ được chọn cứu hoặc độc trong một đêm — được dùng cả hai bình cùng một đêm nếu muốn.

### Thợ săn (rules.md §6)
- Khi nào được gọi: không có lượt đêm; chỉ hoạt động ngay khi chết, vì bất kỳ lý do gì.
- Cách dùng: lúc chết, vai được công khai ngay; trong {seconds} giây, chọn một người sống bất kỳ để bắn.
- Nếu hết giờ: không bắn ai.
- Dễ nhầm: nghĩ mình còn cơ hội bắn lần sau — Thợ săn chỉ bắn đúng một lần, ngay lúc chết.

### Nguyệt Nữ (rules.md §6)
- Khi nào được gọi: được gọi đầu tiên mỗi đêm (đêm 1 thì sau Cupid và cặp đôi).
- Cách dùng: bắt buộc chọn một người sống khác mình để "ghé"; người đó mất chức năng ban đêm đêm đó.
- Nếu hết giờ: server tự chọn ngẫu nhiên một người hợp lệ (không phải chính mình, không trùng người đêm trước).
- Dễ nhầm: chọn lại đúng người đã ghé đêm trước — app không cho chọn.

### Thám tử (rules.md §6, ván từ {minPlayers} người)
- Khi nào được gọi: được gọi mỗi đêm, kể cả sau khi đã dùng (lượt gọi giả, vẫn chạy đủ giờ, để không lộ đã dùng).
- Cách dùng: chọn một người sống (được chọn chính mình); biết vùng gồm người đó và hai người sống ngồi gần nhất có Sói hay không — không biết số lượng hay tên.
- Nếu hết giờ (khi chưa dùng): đêm đó bỏ qua, lượt vẫn còn cho đêm sau.
- Dễ nhầm: nghĩ biết được số lượng hay tên sói trong vùng — app chỉ báo "có Sói" hay "không có Sói", không hơn.

### Hiệp sĩ kiếm gỉ (rules.md §6)
- Khi nào được gọi: không có lượt đêm, năng lực bị động, tự kích hoạt.
- Cách dùng: không cần chạm gì; nếu bị sói cắn chết mà không được Bảo vệ che, tự động làm một con sói còn sống gần nhất bên trái trúng độc.
- Nếu hết giờ: không áp dụng.
- Dễ nhầm: tưởng có nút bấm ban đêm — Hiệp sĩ không làm gì cả, năng lực chỉ kích hoạt khi bị cắn chết.
- Lưu ý: con sói trúng độc không chết ngay đêm đó — nó sống thêm đúng 24 giờ và chết vào đêm sau; cái chết đó được loa đọc sáng hôm sau cùng các cái chết khác.

### Cupid (rules.md §7, ván từ {minPlayers} người)
- Khi nào được gọi: được gọi đầu tiên, chỉ đêm đầu tiên.
- Cách dùng: chọn hai người sống khác nhau (được chọn chính mình) để thành một cặp đôi.
- Nếu hết giờ: server tự ghép ngẫu nhiên hai người sống.
- Dễ nhầm: nghĩ còn đêm sau để ghép đôi — Cupid chỉ được gọi đúng một lần, đêm đầu tiên.

### Trưởng làng (rules.md §7b — danh hiệu công khai, không phải vai riêng)
- Khi nào được gọi: bầu ngay sau khi xem vai, trước đêm 1; giữ cho tới khi chết.
- Cách dùng: lúc bầu, chọn một người (được tự bầu mình); khi đang giữ danh hiệu, phiếu đề cử và phiếu giết/tha của bạn tự tính gấp đôi, không cần làm gì thêm; khi Trưởng làng chết, trong {seconds} giây chọn một người sống kế nhiệm.
- Nếu hết giờ: lúc bầu, không ai bầu cho ai thì server chọn ngẫu nhiên một người. Lúc chọn kế nhiệm, hết giờ thì danh hiệu mất, không ai thay.
- Dễ nhầm: nghĩ vai gốc của mình đổi khi làm Trưởng làng — vai và phe gốc không đổi, chỉ có phiếu tính gấp đôi.
- Lưu ý: đang bị đề cử lên giàn thì mất quyền bỏ phiếu như mọi người khác, kể cả khi đang là Trưởng làng.

### Cặp đôi (rules.md §7 — trạng thái do Cupid ghép, không phải vai riêng)
- Khi nào được gọi: không có lượt riêng; biết nhau ở bước Kiểm tra điện thoại ngay sau lượt Cupid, đêm 1; ngay sau đó mở mắt nhìn nhau {seconds} giây (chỉ hai người, người khác vẫn nhắm mắt).
- Cách dùng: không chọn gì; điện thoại hiện người yêu là ai, vai gì, và màu liên kết theo phe hiện tại của cặp.
- Nếu hết giờ: không áp dụng.
- Dễ nhầm: một người chết **ban đêm** thì người kia chết theo **cùng đêm đó**; một người chết **ban ngày** (bị treo, bị Thợ săn bắn) thì người kia chưa chết ngay, mà chết vào **đêm kế tiếp**. Nếu cặp gồm một người phe Sói và một người phe Dân, hai người thuộc **phe thứ ba**: chỉ thắng khi chính hai người đó là hai người cuối cùng còn sống, dù phe Dân hay phe Sói thắng thì cả hai đều thua.

### Bị nguyền (rules.md §5 — trạng thái do Sói nguyền gây ra, không phải vai riêng)
- Khi nào được gọi: không có lượt riêng; báo ở bước Kiểm tra điện thoại ngay sau lượt Sói nguyền, các đêm 1–3.
- Cách dùng: không chọn gì; điện thoại chỉ báo "Bạn đã bị nguyền" hoặc "Bạn không bị nguyền".
- Nếu hết giờ: không áp dụng.
- Dễ nhầm: từ lúc bị nguyền, bạn thuộc phe Sói, dậy cắn cùng bầy và được tính là Sói khi xét thắng thua — nhưng vẫn giữ nguyên năng lực vai cũ (Tiên tri, Bảo vệ, Thợ săn... vẫn dùng được như trước).

---

## 3. "Lần đầu chơi?" — việc cần nhớ

- Che màn hình (tay, hoặc ngồi quay màn hình vào trong) mỗi khi điện thoại hiện vai của bạn (có {seconds} giây để xem) hoặc lượt đêm — người ngồi cạnh không được thấy.
- Thẻ vai che mặc định; nhấn giữ để xem, thả tay là ẩn lại ngay — đừng mở thẻ vai lâu để người khác liếc thấy.
- Nếu bạn đã chết, màn hình vẫn tối/im ban đêm cho tới khi bạn tự chạm vào xem — đừng để màn hình tự sáng, vì vậy sẽ lộ ra là bạn đã chết.
- Mỗi lượt gọi đêm chạy đủ thời gian như nhau dù vai đó còn sống hay đã chết — đừng đoán ai còn sống vì lượt nhanh hay chậm.
- Tắt âm bàn phím và rung khi chạm máy — tiếng gõ ban đêm lộ ra ai đang thao tác.
- Hết pin hoặc lỡ tắt app giữa ván: vào lại đúng phòng bằng **cùng máy** sẽ tự nhận lại vai ngay; đổi sang **máy khác** thì nhập đúng nickname và mã PIN 4 số được cấp lúc vào phòng.
- Tự nhớ mã PIN của mình, không ai nhắc lại giúp — nhập sai PIN {pinTries} lần sẽ bị khoá vào lại bằng PIN trong {lockMinutes} phút.
- Khi chết, bạn không được nói hay bỏ phiếu nữa, nhưng được xem toàn bộ vai thật và diễn biến của ván từ lúc đó.

---

## 4. Mẹo chơi (tuỳ chọn)

Đây là mẹo chiến thuật, không phải luật — app có thể không cần hiện mục này,
để người chơi tự khám phá cũng được. Tách riêng khỏi mục 2 để không lẫn với
"Dễ nhầm" (chỉ nói về hiểu sai luật hay bấm sai thao tác).

- **Tiên tri**: cân nhắc trước khi nói huỵch toẹt kết quả soi ra ban ngày —
  nói sớm quá dễ lộ mình là Tiên tri.
- **Thám tử**: chỉ dùng được đúng một lần cả ván, có thể chờ thêm vài đêm để
  có manh mối trước khi chọn vùng dò.
- **Phù thủy**: không biết Bảo vệ đã che người bị cắn chưa, nên dùng bình cứu
  ngay đêm nào cũng có thể phí bình oan — cân nhắc trước khi cứu.
- **Sói tiên tri**: soi người khác nhau mỗi đêm thay vì soi lại người đã biết
  chắc, để tìm Tiên tri thật nhanh hơn.

---

## Đã sửa theo xác nhận của anh

- **Sói nguyền được gọi giả đủ 3 đêm đầu dù đã nguyền xong** (anh xác nhận
  ĐÚNG: ngừng gọi sớm sẽ lộ, vì cả bàn nghe tên vai được gọi). Đã ghi thẳng
  vào `docs/rules.md` §3 và §5 (thêm một gạch đầu dòng ở mỗi nơi, giống câu
  đã có sẵn cho Thám tử), không còn là suy diễn của tôi nữa — guide-text.md
  mục 2 (Sói nguyền) đã khớp với rules.md.

## Đối chiếu với `rules.md` (mục 7 anh yêu cầu)

| Điểm cần đối chiếu | Luật nói gì | Đúng/sai trong guide-text.md |
| --- | --- | --- |
| Bầy sói có lựa chọn "Không cắn ai" không | **Có.** §5 Sói: "Danh sách vote của bầy có thêm lựa chọn rõ ràng 'Không cắn ai', tính như một mục tiêu." | Đúng — mục Sói, "Cách dùng" đã ghi đúng. |
| Phù thủy dùng được cả 2 bình trong 1 đêm không | **Được.** §6 Phù thủy: "Được dùng cả hai trong cùng một đêm." | Đúng — mục Phù thủy, "Cách dùng" đã ghi đúng; "Dễ nhầm" vừa sửa lại cũng khẳng định đúng điều này. |
| Nguyệt Nữ cấm ghé cùng người 2 đêm liền không | **Cấm.** §6 Nguyệt Nữ: "Không được chọn cùng một người hai đêm liên tiếp." | Đúng — mục Nguyệt Nữ, "Dễ nhầm" đã ghi đúng. |
| Cupid tự chọn mình được không | **Được.** §7: "Cupid chọn 2 người sống khác nhau (được chọn cả chính mình) thành cặp đôi." | Đúng — mục Cupid, "Cách dùng" đã ghi đúng. |
| Điều kiện thắng ở "Luật nhanh" dòng 8–9 có đúng §8 không | **Đúng.** §8: Dân thắng khi hết sói; Sói thắng khi số Sói còn sống ≥ số Dân còn sống; Cặp đôi khác phe chỉ thắng khi chính hai người đó là hai người cuối sống. | Đúng — hai dòng đó khớp nguyên văn luật, không cần sửa. |

Cả 5 điểm đều khớp sẵn với `rules.md`, không có chỗ nào luật nói mà
guide-text.md ghi sai, và không có chỗ nào luật không nói mà tôi phải tự
đoán thêm — nên không có câu hỏi mới ở mục này.

Dừng ở đây, chờ anh duyệt.
