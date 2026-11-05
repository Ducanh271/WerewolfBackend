# web/ — app Ma Sói (PWA)

Đọc thêm `../CLAUDE.md` (quy tắc chung, thứ tự nguồn sự thật).

- Stack đã chốt: Vite, TypeScript, three.js. Không thêm thư viện lớn khi chưa hỏi.
- Kiểu dữ liệu: `src/api/types.ts` sinh từ `../docs/api/schema/*.json` bằng `npm run gen:types`. Không sửa tay.
- Mỗi `view` **thay hẳn** cái trước; dựng màn hình từ `view` mới nhất. `events` có `reset: true` thì thay cả danh sách.
- `switch` trên `view.phase` và `event.type` phải đầy đủ, nhánh mặc định là `never` để TypeScript báo khi backend thêm pha/sự kiện.
- Đồng hồ luôn theo `deadline`. Hoạt ảnh không được giữ màn hình cũ khi `view` mới đã tới.
- Ảnh lấy qua `assets/manifest.json`. Ảnh mới: `python3 tools/chuan_hoa_anh.py`.
- Màn chuẩn 844 × 390 ngang, nút ≥ 44px, tôn trọng `prefers-reduced-motion`.
- Bản chạy thử `prototype/ban-3d.html` chỉ để chép **hình ảnh**; dữ liệu và luồng lấy từ `../docs/api/`. Các chỗ nó lệch backend: `../docs/frontend/doi-chieu-backend.md`.

Lệnh: `npm run dev` · `npm run check` (gen:types + tsc + test + build — phải qua trước khi báo xong) · `npm test` (phát lại `../docs/api/examples/`).
