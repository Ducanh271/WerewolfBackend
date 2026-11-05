---
description: Soát lệch giữa code web/, hợp đồng docs/api/ và đặc tả giao diện — chỉ đọc, không sửa
argument-hint: [phạm vi, vd "ban đêm" hoặc để trống = tất cả]
---
Soát lệch (phạm vi: $ARGUMENTS). **Chỉ đọc và báo cáo, không sửa file nào.**

1. Chạy `cd web && npm run check` và `make test`; ghi kết quả.
2. Với mỗi màn/pha trong phạm vi, so ba nguồn: `docs/api/screens/*.md` (+ schema), `docs/frontend/giao-dien-3d.md`, code trong `web/src/`.
3. Tìm:
   - Code dùng trường/lệnh/pha không có trong schema, hoặc bỏ sót trường/lệnh/lỗi mà tài liệu yêu cầu.
   - Pha hoặc sự kiện chưa có nhánh xử lý.
   - Chỗ code tự suy ra điều server phải quyết (kết quả soi, người chết, thắng thua, đồng hồ).
   - Vi phạm `docs/api/secrecy.md` (hình, tiếng, thời gian khác nhau giữa các máy).
   - Mục trong `docs/frontend/doi-chieu-backend.md` mục 2 chưa xử lý.
4. Trả về bảng: `mức độ (nặng/vừa/nhẹ) | chỗ (file:dòng) | lệch gì | nguồn đúng (file:dòng) | đề xuất sửa`. Sắp nặng trước.
