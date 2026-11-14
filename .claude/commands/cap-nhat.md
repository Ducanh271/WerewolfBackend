---
description: Áp một gói cập nhật giao diện (docs/frontend/cap-nhat/NN-*.md) vào web/
argument-hint: <đường dẫn file cập nhật, vd docs/frontend/cap-nhat/01-tho-san-va-bang-dang-goi.md>
---
Áp cập nhật: $ARGUMENTS

1. Đọc file cập nhật, rồi **chỉ** các mục đặc tả nó nêu trong `docs/frontend/giao-dien-3d.md` và các tài liệu `docs/api/` nó trích. Không đọc lại toàn bộ tài liệu.
2. Đọc phần bản chạy thử được nêu tên (hàm/CSS) trong `web/prototype/ban-3d.template.html` để lấy hình dáng và nhịp hoạt ảnh. Không chép mã giả lập ("Bảng thử", dữ liệu giả).
3. Nếu cập nhật cần dữ liệu/lệnh server chưa có trong `docs/api/`: DỪNG, nói rõ cần gì, đề nghị `/tinh-nang`.
4. Đưa kế hoạch (file sẽ sửa, trường API dùng kèm file:dòng) rồi **dừng chờ duyệt**.
5. Làm xong: `cd web && npm run check`, báo cáo theo mẫu trong `CLAUDE.md`, và đánh dấu `- [x] đã áp` ở cuối file cập nhật.
