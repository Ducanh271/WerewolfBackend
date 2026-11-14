# Cập nhật 03 — Trúng tên, bị nguyền, trái tim cặp đôi

Loại: **chỉ giao diện**. Dữ liệu đã có sẵn trong hợp đồng:
- `inPerson`: `view.me.phoneCheck` (`cursedText`, `loveText`).
- `online` (`docs/api/screens/phone-check.md` mục Online): sự kiện `PhoneChecked` (`cursed`, `love.inLove`, `love.partner`, `love.partnerRole`, `love.colour`), `view.lover`, `me.cursed`, `me.statusCards`.
- Cả hai: `me.turn.step` `LoversMeet`, `nightCall.role`, `deadline`, `view.mode`.

Đặc tả: `docs/frontend/giao-dien-3d.md` **mục 18** (mới) và bảng "Tin riêng khi chơi `online`" ở **mục 12**.
Bản chạy thử: `web/prototype/ban-3d.template.html` — `phoneCheck`, `fxLove`, `fxCurse`, `fxCalm`, `loversMeet`, `heartURL`, `ARROW_URL`, `EYES_URL`, `LINK`, `fx.pc` trong `Snd`; CSS `#pc`, `.pcCard`, `#pcLink`, `.pcSm`, `.lb.love`, `#callBar.side`. Bỏ qua phần giả lập (`demoPartner`, `linkOf`, nút dev, `players.forEach(p => p.joined = true)`).

## Việc cần làm
1. `inPerson`, pha `Night` với `nightCall.role === 'PhoneCheck'`: hộp nhấn giữ (giữ nguyên quy tắc bí mật).
2. `online`, khi nhận sự kiện `PhoneChecked` của mình, chọn cảnh theo bảng ở mục 18:
   - trúng tên;
   - bị nguyền;
   - sợi chỉ đổi màu (người yêu của mình vừa bị nguyền).
3. `nightCall.role === 'LoversMeet'` trên máy có `me.turn.step === 'LoversMeet'`: cảnh trái tim. Các máy khác giữ màn đêm thường.
4. Nhãn người yêu và thẻ phụ cặp đôi / bị nguyền trong khung xem thẻ.

## Ràng buộc
- Phòng online **không có** bước `PhoneCheck` và không có `me.phoneCheck`. Bản chạy thử chiếu cảnh online trong bước Kiểm tra điện thoại; app thật thì chạy cảnh lúc nhận `PhoneChecked`, trong bước Cupid hoặc Sói nguyền.
- Chữ hiển thị lấy nguyên văn:
  - `inPerson`: `loveText`, `cursedText`;
  - `online`: `text` của thẻ phụ trong `me.statusCards`.
- Màu sợi chỉ lấy từ `love.colour`: `white` / `black` / `striped`.
- Sự kiện quyết định có chạy cảnh không; `view` quyết định vẽ gì. Máy vào lại chỉ nhận `view`, nên không chạy lại cảnh. Có `view` mới thì cắt hoạt ảnh và vẽ theo dữ liệu.
- Chế độ lấy từ `view.mode`.

- [ ] đã áp
