---
description: Thêm/sửa một tính năng cần dữ liệu server — theo thứ tự luật → backend → hợp đồng → giao diện → frontend
argument-hint: <mô tả tính năng hoặc đường dẫn phiếu tính năng>
---
Tính năng: $ARGUMENTS

Làm theo quy trình "hợp đồng trước" trong `docs/frontend/doi-chieu-backend.md` mục 3. Có **3 điểm dừng chờ tôi duyệt**.

**A. Phân tích (dừng 1)**
- Điền phiếu tính năng (mẫu trong `doi-chieu-backend.md`): ai thấy, pha nào, dữ liệu server gửi, lệnh app gửi, lỗi, ảnh hưởng bí mật (`docs/api/secrecy.md`).
- Nêu rõ: có đổi luật không (mục nào của `docs/rules.md`)? Đụng phần nào của Go? Màn hình nào trong `docs/frontend/giao-dien-3d.md`?
- Nếu tính năng chỉ là hình ảnh/âm thanh (không cần dữ liệu mới): nói vậy và chỉ đề xuất phần C–E.

**B. Luật + backend (dừng 2 — sau khi xong)**
- Sửa `docs/rules.md` nếu cần. Code Go + test. `make api-docs`, viết/sửa `docs/api/screens/*.md`, `secrecy.md` nếu liên quan. `make test` phải qua.
- Báo cáo các thay đổi hợp đồng (trường/lệnh/sự kiện/lỗi mới) trước khi sang frontend.

**C–E. Giao diện + frontend (dừng 3 — sau khi xong)**
- C. Cập nhật `docs/frontend/giao-dien-3d.md` (trông thế nào, lấy từ trường nào).
- D. `cd web && npm run gen:types`, code `web/`.
- E. `cd web && npm run check` phải qua. Báo cáo theo mẫu trong `CLAUDE.md`.
