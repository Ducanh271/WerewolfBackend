# Cập nhật 01 — Cảnh Thợ săn chết + bảng "Đang gọi" ban đêm

Loại: **chỉ giao diện** (hình/âm). Không cần trường, lệnh hay pha mới từ server.
Đặc tả đã sửa: `docs/frontend/giao-dien-3d.md` **mục 4** (bảng "Đang gọi") và **mục 17** (mới).
Bản chạy thử tham khảo: `web/prototype/ban-3d.template.html` — các hàm `revealHunter`, `hunterShot`, `fire`, `autoAim`, `stepAim`, `buildTgt`, `showCall`, `callStepSecond` và CSS `#scope`, `#tgt`, `.aimS`, `#aimName`, `#callBar`.
(Bản chạy thử có nút "Bảng thử" để giả lập; đó là đồ thử, **không** chép vào app.)

## Việc cần làm
1. **Bảng "Đang gọi" ban đêm** (mục 4): dựng từ `view.nightCall` + `deadline` + `say` của `NightStepCalled`. Giống hệt nhau trên mọi máy còn sống.
2. **Cảnh Thợ săn** (mục 17), pha `HunterShot`:
   - lật thẻ khi có `HunterRevealed`;
   - máy có `view.me.canShoot` thì vào ống ngắm, gồm: ống to, tên dưới tâm, tâm tự bám, dải tên, nút ‹ ›;
   - máy khác chờ;
   - phát súng khi có `HunterShot`, cuộn giấy dùng câu `say`;
   - hết giờ khi có `HunterTimedOut`, ghi "không ai bị bắn".

## Ràng buộc
- Danh sách người bắn được lấy từ `view.players[].alive`, **không** lọc theo view đầy đủ của người chết (`docs/api/screens/hunter.md`).
- Chế độ `inPerson`: các máy khác chờ im lặng, không có chấm laser. Chế độ `online`: được có chấm laser trang trí.
- Client **không bao giờ tự chọn người chết**. Bản chạy thử đang "hết giờ thì bắn ngẫu nhiên"; đây **chưa phải luật**, nên bỏ qua phần đó cho tới khi có `/tinh-nang` đổi luật + backend.
- Hoạt ảnh không được giữ trạng thái: có `view` mới thì cắt hoạt ảnh và vẽ theo `view`.

- [ ] đã áp
