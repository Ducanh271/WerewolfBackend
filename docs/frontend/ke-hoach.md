# Kế hoạch code frontend theo từng bước (kèm câu lệnh cho Claude Code)

Mỗi bước nhỏ, chạy được, có thể kiểm tra ngay. **Làm xong một bước, cậu chạy thử, rồi mới dán câu lệnh bước sau.** Mỗi câu lệnh đã ghi rõ tài liệu cần đọc, nên Claude Code không phải đoán.

Cấu trúc: frontend nằm ở `web/` **trong chính repo backend**, nên `docs/api/` chỉ có một bản duy nhất. Mỗi bước chạy bằng lệnh `/buoc <số>` (file `.claude/commands/buoc.md`).

---

## Việc backend trước khi code frontend

Chạy từng việc bằng `/tinh-nang`, theo thứ tự. B1, B2, B4 xong trước Bước 2; B3 xong trước Bước 3b.

- **B1 — Ví dụ còn thiếu.** Ván mẫu sinh thêm `view.day`, `view.dusk`, `error.chat_full`; `make api-docs`.
- **B2 — Hai chế độ chơi.** Luật `docs/rules.md` §1 + mục mới về chế độ, sửa §10. Server nhận `mode: online` (bỏ `mode_unsupported`); `view.mode` ở sảnh và trong ván; lệnh đổi chế độ trong sảnh cho chủ phòng; `online` mặc định `allPhones`. `secrecy.md` tách theo chế độ: luật hình/tiếng/độ sáng chỉ bắt buộc với `inPerson`; đồng hồ theo `deadline`, không lộ dữ liệu, lớp che người chết áp dụng cả hai; bước Kiểm tra điện thoại chỉ có ở `inPerson`. Đổi chế độ thì âm thanh về mặc định của chế độ mới, chủ phòng vẫn chỉnh lại được; `newGame` giữ chế độ.
- **B3 — Thẻ vai.** `/tinh-nang docs/tasks/the-ngua.md`.
- **B4 — Phục vụ app và mã QR.** Server Go phục vụ `web/dist`, link `/j/{code}` mở thẳng app; sửa `docs/api/README.md` và `docs/huong-dan-du-an.md`. Thư mục app đọc từ biến môi trường `WEB_DIR` (không nhúng vào binary, vì `web/dist` không nằm trong git). Server chấp nhận `Origin` trùng chính nó, để đổi tên miền không phải cấu hình lại `ALLOWED_ORIGINS`. Lúc phát triển dùng proxy của Vite.
  - Chưa deploy: tạm chạy `ngrok http 8080`, HTTPS do ngrok lo, tên miền có thể đổi. Qua ngrok mọi máy chung một IP (`clientIP`), nên chung giới hạn 100 kết nối và 30 kết nối mới/phút. Khi deploy thật mới tính reverse proxy và tin `X-Forwarded-For`.

---

## Bước 1 — Khung dự án + kiểu dữ liệu từ schema
> Đọc `CLAUDE.md`, `docs/api/README.md` và `docs/frontend/giao-dien-3d.md` mục 1. Tạo dự án Vite + TypeScript trong `web/`, cài `three` (ghim phiên bản). Sinh kiểu TypeScript từ `docs/api/schema/server.schema.json`, `client.schema.json`, `http.schema.json` bằng `json-schema-to-typescript` vào `src/api/types.ts`, kèm các script `gen:types` (đọc `../docs/api/schema/`, và sinh bảng câu lỗi từ `../docs/api/errors.md` vì chưa có file máy đọc được), `check` (= gen:types + `tsc --noEmit` + test + build) trong `web/package.json`. Tạo `src/i18n/vi.ts` cho chữ của app. Thêm job `cd web && npm ci && npm run check` vào `.github/workflows/`. Chưa làm giao diện. Xong thì chạy `cd web && npm run check` và báo lại cấu trúc thư mục.

## Bước 2 — Kết nối server và kho trạng thái
> Đọc `docs/api/README.md` (Connecting, Joining and rejoining, Messages, Limits). Viết `src/net/socket.ts`: mở `/ws`, gửi `join`/`rejoin` trong 10 giây, lưu token theo mã phòng trong localStorage (try/catch), tự kết nối lại với độ trễ tăng dần, dừng đúng các trường hợp README nêu. Viết `src/state/store.ts`: giữ `view` mới nhất (thay hẳn, không trộn), danh sách `events` (xử lý `reset` và `from` đúng README), `deadline`, `speaker`, `people`; phát sự kiện khi đổi. Giữ cả `view.mode`. Cấu hình proxy của Vite tới server Go (`/rooms`, `/join`, `/guide`, `/ws`). Viết test bằng các file trong `docs/api/examples/` (phát lại `examples/server/*.json` vào store và kiểm tra kết quả). Chưa có giao diện.

## Bước 3 — Các màn 2D ngoài ván
> Đọc `docs/api/screens/lobby.md`, `docs/api/errors.md`, `docs/frontend/ban-giao-frontend.md` mục 1–4 và 7. Không có bản vẽ wireframe: dựng từ `lobby.md` và token. Làm các màn 2D: trang đầu, tạo phòng (có ô chọn chế độ), vào phòng, PIN, và các bảng cài đặt (xếp chỗ, chọn vai, giờ & âm thanh, đổi chế độ) bằng DOM theo token màu/chữ/nút/khung 9 lát trong bàn giao. Sảnh chờ là 3D, làm ở Bước 3b; ở bước này chỉ cần danh sách tạm để thử nối. Lỗi hiện đúng câu trong `errors.md`. Dùng ảnh trong `web/assets/` qua `web/assets/manifest.json`.

## Bước 3b — Sảnh chờ 3D, chia vai, xem vai, thẻ của bạn
> Đọc `docs/frontend/giao-dien-3d.md` mục 10 và 14, `docs/api/screens/lobby.md`, `role-view.md`. Làm sảnh trên cảnh quảng trường (người vào hiện dần, mã phòng + QR, chế độ phòng, khung vai, nút Chia vai cho chủ phòng), hoạt ảnh chia bài khi pha đổi sang `RoleView`, màn xem vai (thẻ ngửa), thẻ nhỏ góc trái **luôn úp** và khung xem thẻ (thẻ phụ từ `me.statusCards`, ghi chép từ sự kiện riêng). Làm sau Bước 4–5 và sau B3.

## Bước 4 — Cảnh 3D quảng trường (chưa nối dữ liệu)
> Đọc `docs/frontend/giao-dien-3d.md` mục 2 và mã `web/prototype/ban-3d.template.html` + `web/prototype/villager.js`. Chuyển cảnh 3D của bản chạy thử thành module TypeScript trong `src/scene/`: renderer độ phân giải thấp, mặt đất, nền xa, đống lửa, ghế, nhà, bia mộ, giàn treo cổ, dân làng vẽ từ bộ phận (hạt giống = id người chơi), sprite sói, camera ngồi tại ghế + vuốt quay đầu + bản đồ nhỏ, chuyển ngày/đêm bằng một biến `night`. Giữ nguyên hình ảnh như bản chạy thử. Thêm trang `/dev/scene` có nút giả lập (ngày/đêm, số người, giết 1) để kiểm tra.

## Bước 5 — Nối cảnh 3D với `view`
> Đọc `docs/frontend/giao-dien-3d.md` mục 3–5, 16 và `docs/api/screens/role-view.md`, `chief.md`, `night.md`, `dawn.md`, `day.md`. Viết `src/scene/director.ts`: nhận `view`/`events` từ store, quyết định cảnh (bảng mục 3), đặt người ngồi theo `view.players` (mình ở `A0`), sống/chết, ngày/đêm, lượt đêm (`me.turn`, `nightCall`, `pack`). Cảnh đêm rẽ theo `view.mode` (bảng mục 16). Chạm nhà/người + chạm tay gửi đúng lệnh trong `night.md` / `day.md`. Hiện lỗi server như mọi lỗi khác. Không thêm trường nào không có trong schema.

## Bước 6 — Hai tay góc nhìn thứ nhất
> Đọc `docs/frontend/ban-giao-frontend.md` mục 4–5 và `docs/frontend/giao-dien-3d.md` mục 4, 16. Làm lớp tay DOM: ảnh theo vai từ `manifest.json` (trạng thái `-da-dung`), nhãn trên tay, lắc nhẹ khi quay đầu. Không tới lượt: `inPerson` vẫn hiện tay, `online` hạ xuống và tối đi. Hạ xuống khi đang trò chuyện. Mỗi vai đúng "tay làm" và lệnh như bảng.

## Bước 6b — Trưởng làng, hiệu ứng chức năng đêm, trời sáng
> Đọc `docs/frontend/giao-dien-3d.md` mục 11–13 và `chief.md`, `night.md` (sự kiện kết quả), `dawn.md`. Làm bầu Trưởng làng + thẻ bay + trang phục + kế nhiệm; hiệu ứng từng vai (chỉ trên máy người dùng, kết quả lấy từ sự kiện; `inPerson` vẽ dịu, không chớp, không tiếng; `online` đầy đủ); cuộn giấy trời sáng/hoàng hôn kèm âm thanh.

## Bước 7 — Trò chuyện ban ngày
> Luật ở `giao-dien-3d.md` mục 15, như nhau ở cả hai chế độ. Đọc `docs/api/screens/day.md` mục Chat (lệnh gửi, sự kiện nhận, giới hạn, ai được nhắn), `docs/api/secrecy.md` (chữ thuần) và `docs/frontend/giao-dien-3d.md` mục 5. Làm bảng "Bàn tán": ảnh đại diện cắt từ nhân vật, tên màu riêng, tin của mình bên phải, chấm đỏ tin chưa đọc, bong bóng thoại trên đầu 4 giây. Ô nhập hiện theo pha và người như bảng mục 15, mọi máy đọc được. Theo đúng giới hạn server.

## Bước 8 — Cảnh lên giàn, bỏ phiếu, treo cổ
> Đọc `docs/api/screens/day.md` (Trial, Verdict, Dusk, sự kiện TrialStarted/VerdictCast/Hanged/Spared) và `docs/frontend/giao-dien-3d.md` mục 5. Chuyển cảnh lên giàn, bỏ phiếu bằng hai tay (`verdict {kill}`), treo cổ với mốc thời gian như mô tả (cắt đen, không chiếu cảnh chết), chạy theo sự kiện `Hanged`. Pha kế là `HunterShot`, `ChiefSuccession` hoặc `Ended` thì rút gọn còn cắt đen + chữ ~1 giây. Đồng hồ theo `deadline`. Nếu `view` mới đến giữa hoạt ảnh thì cắt hoạt ảnh.

## Bước 9 — Thợ săn, Trưởng làng chết, người chết, Kiểm tra điện thoại
> Đọc `docs/api/screens/hunter.md`, `chief.md` (Succession), `phone-check.md`, và **toàn bộ `docs/api/secrecy.md`**. Làm bộ chọn công khai của Thợ săn/Trưởng làng chết (chỉ tên người sống theo `view.players[].alive`, có tay riêng), lớp che người chết (nhấn giữ mới xem, không đổi khi có tin), hộp Kiểm tra điện thoại nhấn giữ.

## Bước 10 — Màn kết thúc
> Đọc `docs/api/screens/end.md` và `docs/frontend/giao-dien-3d.md` mục 6. Làm màn kết thúc lật thẻ theo `view.winner`, `view.winners`, `view.draw`, `view.players[].card`, `view.summary`; nút "Ván mới" chỉ chủ phòng (`newGame`).

## Bước 11 — Âm thanh và cài đặt
> Đọc `docs/frontend/giao-dien-3d.md` mục 7, 16 và lớp `Snd` trong `web/prototype/ban-3d.template.html`. Chuyển thành `src/audio/`. Bảng Cài đặt bật/tắt + âm lượng riêng Nhạc nền / Hiệu ứng, từng máy; không có nút "Im lặng ban đêm". **`inPerson`: ban ngày chỉ máy loa phát, máy khác mặc định im và tự bật được; ban đêm im hết. `online`: mọi máy đủ tiếng cả ngày lẫn đêm.** Máy có `speaker: true` đọc `say` bằng giọng đọc và giảm nhạc khi đọc.

## Bước 12 — PWA, hiệu năng, kiểm tra cuối
> Thêm service worker precache `web/assets/` và font, manifest PWA khoá màn ngang. App do server Go phục vụ (B4), tạm qua HTTPS của ngrok; service worker cần HTTPS. Kiểm tra trên khung 844×390: không chữ nào tràn, mọi nút ≥ 44px, `prefers-reduced-motion`. Chạy một ván giả từ đầu tới cuối bằng `docs/api/examples/` và liệt kê chỗ nào giao diện còn lệch đặc tả.

---

### Mẹo khi làm với Claude Code
- Mỗi bước mở **phiên mới** hoặc `/clear`, dán câu lệnh của bước đó.
- Khi nó hỏi lại hoặc đề xuất đổi thiết kế: so với `docs/frontend/giao-dien-3d.md`; muốn đổi thì sửa tài liệu trước rồi mới cho code.
- Bắt nó chạy `cd web && npm run build` và test sau mỗi bước, đọc lỗi rồi sửa, trước khi báo xong.
- Backend đổi (thêm trường, lệnh, như phần trò chuyện): chạy lại `cd web && npm run gen:types`, rồi bảo nó đọc phần tài liệu đổi.
