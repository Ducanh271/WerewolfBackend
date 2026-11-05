# Kế hoạch code frontend theo từng bước (kèm câu lệnh cho Claude Code)

Mỗi bước nhỏ, chạy được, có thể kiểm tra ngay. **Làm xong một bước, cậu chạy thử, rồi mới dán câu lệnh bước sau.** Mỗi câu lệnh đã ghi rõ tài liệu cần đọc, nên Claude Code không phải đoán.

Cấu trúc: frontend nằm ở `web/` **trong chính repo backend**, nên `docs/api/` chỉ có một bản duy nhất. Mỗi bước chạy bằng lệnh `/buoc <số>` (file `.claude/commands/buoc.md`).

---

## Bước 1 — Khung dự án + kiểu dữ liệu từ schema
> Đọc `CLAUDE.md`, `docs/api/README.md` và `docs/frontend/giao-dien-3d.md` mục 1. Tạo dự án Vite + TypeScript trong `web/`, cài `three` (ghim phiên bản). Sinh kiểu TypeScript từ `docs/api/schema/server.schema.json`, `client.schema.json`, `http.schema.json` bằng `json-schema-to-typescript` vào `src/api/types.ts`, kèm các script `gen:types` (đọc `../docs/api/schema/`), `check` (= gen:types + `tsc --noEmit` + test + build) trong `web/package.json`. Chưa làm giao diện. Xong thì chạy `cd web && npm run check` và báo lại cấu trúc thư mục.

## Bước 2 — Kết nối server và kho trạng thái
> Đọc `docs/api/README.md` (Connecting, Joining and rejoining, Messages, Limits). Viết `src/net/socket.ts`: mở `/ws`, gửi `join`/`rejoin` trong 10 giây, lưu token theo mã phòng trong localStorage (try/catch), tự kết nối lại với độ trễ tăng dần, dừng đúng các trường hợp README nêu. Viết `src/state/store.ts`: giữ `view` mới nhất (thay hẳn, không trộn), danh sách `events` (xử lý `reset` và `from` đúng README), `deadline`, `speaker`, `people`; phát sự kiện khi đổi. Viết test bằng các file trong `docs/api/examples/` (phát lại `examples/server/*.json` vào store và kiểm tra kết quả). Chưa có giao diện.

## Bước 3 — Các màn 2D ngoài ván
> Đọc `docs/api/screens/lobby.md`, `docs/api/errors.md`, `docs/frontend/ban-giao-frontend.md` mục 1–4 và 7. Làm các màn A1–A9, B1–B5 (trang đầu, tạo phòng, vào phòng, PIN, sảnh, xếp chỗ, chọn vai, giờ & âm thanh) bằng DOM theo token màu/chữ/nút/khung 9 lát trong bàn giao. Lỗi hiện đúng câu trong `errors.md`. Dùng ảnh trong `web/assets/` qua `web/assets/manifest.json`.

## Bước 3b — Sảnh chờ 3D, chia vai, xem vai, thẻ của bạn
> Đọc `docs/frontend/giao-dien-3d.md` mục 10 và 14, `docs/api/screens/lobby.md`, `role-view.md`. Làm sảnh trên cảnh quảng trường (người vào hiện dần, mã phòng + QR, khung vai, nút Chia vai cho chủ phòng), hoạt ảnh chia bài khi pha đổi sang `RoleView`, màn xem vai, thẻ nhỏ góc trái và khung xem thẻ (thẻ phụ từ `me.statusCards`, ghi chép từ sự kiện riêng). Làm sau Bước 4–5.

## Bước 4 — Cảnh 3D quảng trường (chưa nối dữ liệu)
> Đọc `docs/frontend/giao-dien-3d.md` mục 2 và mã `web/prototype/ban-3d.template.html` + `web/prototype/villager.js`. Chuyển cảnh 3D của bản chạy thử thành module TypeScript trong `src/scene/`: renderer độ phân giải thấp, mặt đất, nền xa, đống lửa, ghế, nhà, bia mộ, giàn treo cổ, dân làng vẽ từ bộ phận (hạt giống = id người chơi), sprite sói, camera ngồi tại ghế + vuốt quay đầu + bản đồ nhỏ, chuyển ngày/đêm bằng một biến `night`. Giữ nguyên hình ảnh như bản chạy thử. Thêm trang `/dev/scene` có nút giả lập (ngày/đêm, số người, giết 1) để kiểm tra.

## Bước 5 — Nối cảnh 3D với `view`
> Đọc `docs/frontend/giao-dien-3d.md` mục 3–5 và `docs/api/screens/role-view.md`, `chief.md`, `night.md`, `dawn.md`, `day.md`. Viết `src/scene/director.ts`: nhận `view`/`events` từ store, quyết định cảnh (bảng mục 3), đặt người ngồi theo `view.players` (mình ở `A0`), sống/chết, ngày/đêm, lượt đêm (`me.turn`, `nightCall`, `pack`). Chạm nhà/người + chạm tay gửi đúng lệnh trong `night.md` / `day.md`. Hiện lỗi server như mọi lỗi khác. Không thêm trường nào không có trong schema.

## Bước 6 — Hai tay góc nhìn thứ nhất
> Đọc `docs/frontend/ban-giao-frontend.md` mục 4–5 và `docs/frontend/giao-dien-3d.md` mục 4. Làm lớp tay DOM: ảnh theo vai từ `manifest.json` (trạng thái `-da-dung`), nhãn trên tay, lắc nhẹ khi quay đầu, hạ xuống khi ngủ/đang trò chuyện. Mỗi vai đúng "tay làm" và lệnh như bảng.

## Bước 6b — Trưởng làng, hiệu ứng chức năng đêm, trời sáng
> Đọc `docs/frontend/giao-dien-3d.md` mục 11–13 và `chief.md`, `night.md` (sự kiện kết quả), `dawn.md`. Làm bầu Trưởng làng + thẻ bay + trang phục + kế nhiệm; hiệu ứng từng vai (chỉ trên máy người dùng, kết quả lấy từ sự kiện); cuộn giấy trời sáng/hoàng hôn kèm âm thanh. Tôn trọng 'Im lặng ban đêm'.

## Bước 7 — Trò chuyện ban ngày
> Luật: chỉ sảnh và ban ngày, không nhắn ban đêm, không kênh riêng (`giao-dien-3d.md` mục 15). Đọc tài liệu trò chuyện của backend (lệnh gửi, sự kiện nhận, giới hạn, ai được nhắn) và `docs/frontend/giao-dien-3d.md` mục 5. Làm bảng "Bàn tán": ảnh đại diện cắt từ nhân vật, tên màu riêng, tin của mình bên phải, chấm đỏ tin chưa đọc, bong bóng thoại trên đầu 4 giây. Khoá ban đêm. Theo đúng giới hạn server.

## Bước 8 — Cảnh lên giàn, bỏ phiếu, treo cổ
> Đọc `docs/api/screens/day.md` (Trial, Verdict, Dusk, sự kiện TrialStarted/VerdictCast/Hanged/Spared) và `docs/frontend/giao-dien-3d.md` mục 5. Chuyển cảnh lên giàn, bỏ phiếu bằng hai tay (`verdict {kill}`), treo cổ với mốc thời gian như mô tả (cắt đen, không chiếu cảnh chết). Đồng hồ theo `deadline`. Nếu `view` mới đến giữa hoạt ảnh thì cắt hoạt ảnh.

## Bước 9 — Thợ săn, Trưởng làng chết, người chết, Kiểm tra điện thoại
> Đọc `docs/api/screens/hunter.md`, `chief.md` (Succession), `phone-check.md`, và **toàn bộ `docs/api/secrecy.md`**. Làm bộ chọn công khai của Thợ săn/Trưởng làng chết (chỉ tên người sống theo `view.players[].alive`, có tay riêng), lớp che người chết (nhấn giữ mới xem, không đổi khi có tin), hộp Kiểm tra điện thoại nhấn giữ.

## Bước 10 — Màn kết thúc
> Đọc `docs/api/screens/end.md` và `docs/frontend/giao-dien-3d.md` mục 6. Làm màn kết thúc lật thẻ theo `view.winner`, `view.winners`, `view.draw`, `view.players[].card`, `view.summary`; nút "Ván mới" chỉ chủ phòng (`newGame`).

## Bước 11 — Âm thanh và cài đặt
> Đọc `docs/frontend/giao-dien-3d.md` mục 7 và lớp `Snd` trong `web/prototype/ban-3d.template.html`. Chuyển thành `src/audio/`. Bảng Cài đặt bật/tắt + âm lượng. **Khi chơi trực tiếp, ban đêm tắt mọi hiệu ứng do mình bấm.** Máy loa đọc `say` bằng giọng đọc và giảm nhạc khi đọc.

## Bước 12 — PWA, hiệu năng, kiểm tra cuối
> Thêm service worker precache `web/assets/` và font, manifest PWA khoá màn ngang. Kiểm tra trên khung 844×390: không chữ nào tràn, mọi nút ≥ 44px, `prefers-reduced-motion`. Chạy một ván giả từ đầu tới cuối bằng `docs/api/examples/` và liệt kê chỗ nào giao diện còn lệch đặc tả.

---

### Mẹo khi làm với Claude Code
- Mỗi bước mở **phiên mới** hoặc `/clear`, dán câu lệnh của bước đó.
- Khi nó hỏi lại hoặc đề xuất đổi thiết kế: so với `docs/frontend/giao-dien-3d.md`; muốn đổi thì sửa tài liệu trước rồi mới cho code.
- Bắt nó chạy `cd web && npm run build` và test sau mỗi bước, đọc lỗi rồi sửa, trước khi báo xong.
- Backend đổi (thêm trường, lệnh, như phần trò chuyện): chạy lại `cd web && npm run gen:types`, rồi bảo nó đọc phần tài liệu đổi.
