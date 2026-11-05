---
description: Làm một bước trong docs/frontend/ke-hoach.md (vd /buoc 5)
argument-hint: <số bước>
---
Làm **Bước $ARGUMENTS** trong `docs/frontend/ke-hoach.md`, đúng phạm vi bước đó, không hơn.

1. Đọc `CLAUDE.md`, `web/CLAUDE.md`, phần Bước $ARGUMENTS và mọi tài liệu bước đó nêu tên. Đọc `docs/frontend/doi-chieu-backend.md` để biết chỗ bản chạy thử lệch backend.
2. Viết **kế hoạch ngắn** rồi DỪNG chờ tôi duyệt:
   - File sẽ tạo/sửa.
   - Trường `view`, sự kiện, lệnh sẽ dùng — mỗi cái kèm nguồn (file:dòng trong `docs/api/`).
   - Chỗ nào tài liệu thiếu hoặc mâu thuẫn (nếu có) và câu hỏi cần tôi trả lời.
   - Cách kiểm tra.
3. Sau khi tôi duyệt: code, chạy `cd web && npm run check`, sửa tới khi qua. Không sửa file ngoài phạm vi; thấy việc khác cần làm thì ghi lại.
4. Tick bước này ở mục "Tiến độ" của `docs/frontend/ke-hoach.md`, chung commit với phần việc.
5. Báo cáo theo mẫu trong `CLAUDE.md`.
