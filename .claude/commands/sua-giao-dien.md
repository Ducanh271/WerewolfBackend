---
description: Sửa phần chỉ thuộc hình ảnh/âm thanh (không đụng dữ liệu server)
argument-hint: <mô tả thay đổi>
---
Thay đổi giao diện: $ARGUMENTS

- Chỉ được sửa `web/` và `docs/frontend/giao-dien-3d.md`. **Không** sửa Go, `docs/api/`, `docs/rules.md`.
- Nếu thay đổi cần dữ liệu hoặc lệnh mới từ server: DỪNG, nói rõ cần gì, đề nghị chạy `/tinh-nang` thay vì làm tiếp.
- Kiểm tra `docs/api/secrecy.md`: thay đổi có làm hình, tiếng hoặc thời gian khác nhau giữa các máy theo vai/hành động không?
- Cập nhật đặc tả trước, rồi code, rồi `cd web && npm run check`. Báo cáo theo mẫu trong `CLAUDE.md`.
