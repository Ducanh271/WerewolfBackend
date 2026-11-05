# Task cho agent backend: thẻ vai hiện ngửa (Q25) + hai tay (Q26)

Dán nguyên file này cho agent làm việc trên repo `WerewolfBackend`.

## Bối cảnh
Thiết kế giao diện đã chốt (claude/thiet-ke.md, Q25, Q26):
- Thẻ vai **hiện ngửa** trên máy của chính người chơi suốt ván. Chạm 1 lần = úp thẻ, chạm lần nữa = lật lại. Chạm đúp khi thẻ ngửa = mở trang hướng dẫn của vai đó. App **không còn bắt nhấn giữ** để xem thẻ.
- Ban đêm mỗi máy hiện thẻ của mình và hai bàn tay của vai (góc nhìn thứ nhất). Chỉ máy đang tới lượt có tay bấm được: chạm một người trên bàn để chọn, chạm tay để xác nhận.

**Không đổi:** màn che người chết (nhấn giữ mới xem), câu trả lời ở bước Kiểm tra điện thoại (nhấn giữ), không âm thanh/rung/nháy, lượt gọi chạy đủ giờ, lỗi giống nhau.

## Việc cần làm

1. **Tài liệu** — sửa đúng các chỗ sau (bản đã sửa sẵn nằm trong Project, cùng đường dẫn, có thể chép đè):
   - `docs/rules.md` §9b, câu "Thẻ có kích thước cố định, che mặc định, nhấn giữ mới hiện."
   - `docs/api/secrecy.md`: mục **Role cards**, mục **Night** (2 gạch đầu dòng đầu), mục **Guide**, câu "like a role card" ở mục Phone check.
   - `docs/api/screens/role-view.md`: đoạn "Show the card at a fixed size, covered…".
   - `docs/api/screens/guide.md`: đoạn "The guide is public, but…" và ví dụ `firstTime`.
   - `docs/api/screens/phone-check.md`: "behind press-and-hold like the role card".
   - `docs/protocol.md`: dòng `me.roleCard`, "(fixed size, hidden until pressed)" → "(fixed size, face up; tap to turn face down, double tap for the role's guide page)".
   - `docs/guide-text.md` mục 3, gạch đầu dòng thứ hai.

2. **Code** — chữ hướng dẫn do server sinh (`GET /guide`, mảng `firstTime`):
   - Tìm chuỗi `Thẻ vai che mặc định; nhấn giữ để xem` trong code Go và thay bằng:
     `Thẻ vai hiện ngửa trên máy bạn: chạm một lần để úp, chạm lần nữa để lật lại, chạm đúp khi thẻ đang ngửa để xem cách chơi. Chơi quanh bàn thì giữ màn hình của mình kín đáo, úp thẻ khi không cần xem.`
   - Chạy lại bộ sinh ví dụ (các khối `<!-- example: http/guide.reply.json … -->`) để ví dụ trong `guide.md` khớp code.
   - Chạy toàn bộ test; test so chữ guide (nếu có) cập nhật theo chuỗi mới.

3. **Kiểm tra cuối**: `grep -rn "nhấn giữ\|press-and-hold\|pressed" docs/` — mọi chỗ còn lại phải chỉ nói về người chết hoặc bước Kiểm tra điện thoại.
