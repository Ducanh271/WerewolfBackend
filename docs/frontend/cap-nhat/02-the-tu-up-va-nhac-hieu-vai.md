# Cập nhật 02 — Thẻ tự úp bay về góc + nhạc hiệu từng vai

Loại: **chỉ giao diện**. Không cần dữ liệu server mới.
Đặc tả đã sửa: `docs/frontend/giao-dien-3d.md` **mục 7** (bảng nhạc hiệu), **mục 10** (câu cuối, xem vai), **mục 14** (thẻ góc luôn úp).
Bản chạy thử tham khảo: `web/prototype/ban-3d.template.html` — `roleView` (đoạn sau "Tôi đã xem"), `refreshMyCard`, `openCardView`, `closeCardView`, `CALL` và `fx.call` trong `Snd`, chỗ gọi `Snd.fx.call` trong `showCall`.

## Việc cần làm
1. **Xem vai** (`RoleView`): hết `deadline` hoặc bấm "Tôi đã xem" → chữ mờ → thẻ lật úp → bay về góc trái trên.
2. **Thẻ góc**: luôn hiện mặt sau. Khi chạm: phóng to từ góc, còn úp, rồi lật ngửa. Khi đóng: lật úp rồi tắt.
3. **Nhạc hiệu**: mỗi khi `nightCall.role` đổi (hoặc có `NightStepCalled`), phát đoạn tương ứng. Đặt chung với bảng "Đang gọi" của cập nhật 01.

## Ràng buộc
- Nhạc hiệu chỉ phụ thuộc `nightCall.role` (công khai), **không** phụ thuộc vai của người cầm máy.
  - `inPerson`: chỉ máy loa phát.
  - `online`: mọi máy phát.
- Thẻ góc không bao giờ để lộ mặt vai khi chưa chạm.
- Làm **sau** cập nhật 01, vì dùng chung bảng "Đang gọi".

- [ ] đã áp
