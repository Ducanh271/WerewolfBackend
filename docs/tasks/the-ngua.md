# Phiếu tính năng: thẻ vai ngửa ở màn Xem vai, úp trong ván

Chạy bằng `/tinh-nang docs/tasks/the-ngua.md`. Repo này là bản đúng; mọi đoạn tài liệu tự soạn lại.

## Quyết định (áp dụng cả hai chế độ chơi)
- **Màn `RoleView`**: thẻ hiện ngửa. Chạm 1 lần = úp, chạm lần nữa = lật lại. Chạm đúp khi đang ngửa = mở trang cách chơi của vai.
- **Trong ván**: thẻ nhỏ ở góc màn **luôn úp**, mặt sau giống nhau cho mọi vai. Chỉ khi chạm mới lật và phóng to ra khung xem thẻ.
- Ban đêm mỗi máy thấy hai bàn tay của vai mình; ở `inPerson` mọi máy thấy, ở `online` chỉ rõ khi tới lượt (`docs/frontend/giao-dien-3d.md` mục 16). Máy tới lượt chạm một người để chọn, chạm tay để xác nhận.

**Không đổi:** màn che người chết (nhấn giữ mới xem), câu trả lời ở bước Kiểm tra điện thoại (nhấn giữ), lượt gọi chạy đủ giờ, lỗi giống nhau.

Làm sau B2, vì B2 tách `docs/api/secrecy.md` theo chế độ.

## Việc cần làm

1. **Tài liệu**, tự soạn lại đúng các chỗ sau:
   - `docs/rules.md` §9b: câu "Thẻ có kích thước cố định, che mặc định, nhấn giữ mới hiện."
   - `docs/api/secrecy.md`: mục **Role cards**, mục **Guide**, câu "like a role card" ở mục Phone check.
   - `docs/api/screens/role-view.md`: đoạn "Show the card at a fixed size, covered…".
   - `docs/api/screens/guide.md`: đoạn "The guide is public, but…" và ví dụ `firstTime`.
   - `docs/api/screens/phone-check.md`: "behind press-and-hold like the role card".
   - `docs/protocol.md`: dòng `me.roleCard`, cụm "(fixed size, hidden until pressed)".
   - `docs/guide-text.md` mục 3, gạch đầu dòng thứ hai.

2. **Code**: chữ hướng dẫn do server sinh (`GET /guide`, mảng `firstTime`) ở `internal/i18n/guide.go`, câu "Thẻ vai che mặc định; nhấn giữ để xem, thả tay là ẩn lại ngay…". Soạn câu mới theo quyết định trên, đưa tôi duyệt ở điểm dừng 1. Chạy `make api-docs`; test so chữ guide (nếu có) cập nhật theo câu mới.

3. **Kiểm tra cuối**: `grep -rn "nhấn giữ\|press-and-hold\|pressed" docs/ internal/i18n/`: mọi chỗ còn lại phải chỉ nói về người chết hoặc bước Kiểm tra điện thoại.
