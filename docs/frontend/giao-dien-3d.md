# Giao diện Ma Sói — bản đặc tả cho frontend (hướng pixel 3D)

Tài liệu này nói **giao diện trông thế nào và chạy theo dữ liệu server ra sao**. Luật và dữ liệu là của backend: `docs/rules.md`, `docs/api/` (README, `screens/*.md`, `secrecy.md`, `errors.md`, `schema/*.json`). Khi tài liệu này và `docs/api/` vênh nhau, **`docs/api/` đúng**.

Bản chạy thử tham khảo: `web/prototype/ban-3d.html` (mở thẳng bằng trình duyệt). Mã nguồn dễ đọc: `web/prototype/ban-3d.template.html`, `web/prototype/villager.js`. Bản chạy thử dùng dữ liệu giả; app thật lấy mọi thứ từ `view` và `events`.

---

## 1. Khung chung

- Điện thoại **cầm ngang**, máy chuẩn iPhone 12: **844 × 390**. Máy dọc: lớp phủ "Xoay ngang điện thoại".
- Lề an toàn trái/phải theo `env(safe-area-inset-*)`, tối thiểu 16px.
- Hai lớp chồng nhau:
  1. **Cảnh 3D** (three.js) vẽ ở độ phân giải thấp: chiều cao khung vẽ ≈ **270 điểm ảnh**, rồi phóng to bằng `image-rendering: pixelated`. Mọi texture dùng `NearestFilter`, không mipmap.
  2. **Lớp giao diện DOM** phía trên: thanh trên, nhãn tên, bong bóng thoại, bảng trò chuyện, hai tay, bảng cài đặt, màn kết thúc. Chữ là chữ thật (không vẽ vào canvas) để sắc nét và đọc được bằng trình đọc màn hình.
- Chữ: `VT323` (thân), `Grenze Gotisch` 600 (tiêu đề, tên vai). Màu và nút theo `docs/frontend/ban-giao-frontend.md` mục 2–4 (token màu, nút vàng/đêm/máu).

## 2. Cảnh 3D

### Bố cục quảng trường
- Đống lửa ở tâm (0,0,0). Ghế (gốc cây) xếp vòng tròn bán kính **2.5**, nhà xếp vòng **4.1**, cùng góc với ghế.
- Ghế `i` (theo `view.players` thứ tự ghế) ở góc `A0 + i·2π/N`, với **ghế của mình luôn là `A0 = π/2`** (xoay cả vòng sao cho mình ở gần camera).
- Giàn treo cổ cố định phía bên kia quảng trường (`z = -7.4`).
- Nền xa: hình trụ bán kính 16 dán tranh `nen/ngay.webp` và `nen/dem.webp` (lặp gương 4 lần), hai lớp mờ chồng chéo theo độ tối.
- Mặt đất: cỏ (texture nhiễu + dither, ngọn cỏ, hoa), sân đất tròn có mép lởm chởm, vòng đất mòn chỗ ghế, vết cháy quanh lửa; ~300 bụi cỏ billboard, sỏi.

### Camera
- Ngồi tại ghế của mình, cao 1.18, nhìn về tâm, cúi nhẹ.
- **Vuốt ngang** để quay đầu, giới hạn ±112°; chuột kéo/phím mũi tên trên máy tính. Quay mượt (lerp).
- Bản đồ nhỏ góc phải: vòng ghế, quạt hướng nhìn; chạm một ghế để quay sang và chọn người đó.

### Nhân vật
- Dân làng là **billboard 32×64** vẽ bằng code từ bộ phận (da, tóc, kiểu tóc, mũ, áo, tạp dề/gilê/khăn choàng, quần/váy, giày, râu) — xem `villager.js`. Tạo hình **ngẫu nhiên theo hạt giống cố định từ id người chơi** để mọi máy thấy cùng một người giống nhau. Tạo hình **không bao giờ phụ thuộc vai**.
- Sprite không chịu ánh sáng: mỗi khung hình nhân màu theo ngày/đêm và ánh lửa bập bùng.
- Ảnh đại diện trong trò chuyện = cắt vùng đầu (22×22) từ chính canvas nhân vật.
- Người chết: ban ngày đầu lâu trên gốc cây, ban đêm bia mộ thay nhà.
- Sói (đồng bọn, chỉ phe sói thấy): `web/prototype/wolf.png` (pixel hoá từ thẻ Sói), đứng trước nhà.

### Ngày ↔ đêm
- Một biến `night` 0→1 trong ~2.2 giây điều khiển: ánh sáng bán cầu, nắng, sương mù, nền, lửa (bùng lên khi `night > 0.35`), tàn lửa/khói.
- Chập tối: dân làng mờ dần, nhà mọc lên từ đất (scale Y). Rạng sáng: ngược lại.
- Lửa: 4 sprite ngọn lửa cộng màu, 4 khung hình luân phiên ~11 khung/giây, một PointLight dao động độ sáng; lớp vignette DOM ấm theo nhịp lửa.

## 3. Pha của server → cảnh

| `view.phase` (+ trường) | Cảnh | Ai làm gì |
|---|---|---|
| `Lobby` | Giao diện 2D (A/B trong canvas thiết kế), chưa vào quảng trường | xem `screens/lobby.md` |
| `RoleView` | Quảng trường ban ngày, thẻ vai ngửa ở giữa (C1) | chạm úp, chạm đúp xem cách chơi |
| `ChiefElection` | Ban ngày, chạm người để bầu | `voteChief` |
| `Night` + `nightCall.role` | Ban đêm, nhà | xem mục 4 |
| `Night` + `me.phoneCheck` | Ban đêm, hộp thông báo **nhấn giữ mới hiện** | `phone-check.md` |
| `Dawn` | Chuyển đêm → ngày; đọc `DawnAnnounced`; ghế người chết chuyển đầu lâu | — |
| `Day` | Ban ngày: đề cử, trò chuyện | `nominate`, `skipDay`, `abortVote` |
| `NominationLock` | Ban ngày, đếm ngược chốt đề cử | — |
| `Trial` + `votes.defendant` | **Cảnh lên giàn**: camera ra trước giàn, đám đông bóng lưng, đếm ngược theo `deadline` | người bị đề cử thanh minh (nói thật quanh bàn / trò chuyện) |
| `Verdict` | Cùng cảnh giàn; **tay trái = Tha, tay phải = Treo**; bảng phiếu | `verdict {kill}` (trừ người bị xử) |
| `Dusk` + sự kiện `Hanged` | **Cảnh treo cổ** rồi chuyển ngày → đêm | — |
| `Dusk` + `Spared`/không ai | Thẻ chữ "Dân làng tha cho …" rồi chuyển sang đêm | — |
| `HunterShot` | Máy Thợ săn (`me.canShoot`): tay cầm súng + chọn người; máy khác: chờ | `hunterShoot` |
| `ChiefSuccession` | Máy Trưởng làng chết (`me.id == view.succession`): tay chìa khoá/cuộn giấy + chọn | `nameSuccessor` |
| `Ended` | **Màn kết thúc lật thẻ** (mục 6) | `newGame` (chủ phòng) |

Bản chạy thử có 12 chỗ lệch backend (tự lên giàn khi đề cử, thiếu nút Bỏ qua/Huỷ ván, phiếu giả, thiếu dải lời dẫn…): danh sách và cách sửa ở `doi-chieu-backend.md` — **luôn làm theo backend**.

Nguyên tắc: **màn hình dựng lại hoàn toàn từ `view` mới nhất** (mỗi `view` thay cái trước). Hoạt ảnh chuyển cảnh chạy khi **pha đổi** hoặc khi có sự kiện tương ứng trong `events` (`TrialStarted`, `Hanged`, `Spared`, `GameEnded`…). Hoạt ảnh không được làm chậm trạng thái thật: nếu `view` mới đến giữa hoạt ảnh, cắt hoạt ảnh và vẽ theo `view`.

## 4. Ban đêm

- `nightCall.role` cho biết vai đang được gọi (giống nhau trên mọi máy). Thanh trên: "Đêm N · {Vai} đang được gọi" + đồng hồ theo `deadline`.
- **Không tới lượt mình** (`me.turn` vắng): màn hình **gần đen** (che 93%), tay mình hạ xuống và tối đi.
- **Tới lượt** (`me.turn.step`): căn phòng sáng lên, tay vai mình trồi lên. **Chạm một ngôi nhà để chọn → chạm "tay làm" để gửi lệnh** (Q27). Bảng vai → tay → lệnh ở `docs/frontend/ban-giao-frontend.md` mục 5.
- `me.turn.blocked/doomed/grief`: hiện `me.turn.notice`, tay không bấm được.
- **Bầy sói** (`view.pack`): đồng bọn hiện là sprite sói trước nhà, nhãn viền đỏ; nhà sói không có "zzz", không chọn được. Phiếu từng con sói (`pack.votes`) hiện nhãn "→ tên" trên nhà bị chọn. Nút "Không cắn ai" → `wolfVote {target: ""}`.
- Nhà người còn sống có "zzz" bay lên — **trên mọi máy giống nhau** (kể cả nhà sói khi người xem không phải sói).
- Phù thủy: tay trái = cứu (chạm thẳng, `witchHeal`), tay phải = chọn nhà rồi chạm để đầu độc (`witchPoison`).

## 5. Ban ngày

- Đề cử: chạm người → chạm tay trái (chỉ trỏ) → `nominate {target}`. Nhãn trên đầu: số phiếu đề cử (`view.votes`).
- **Trò chuyện** (luật ở mục 15): nút "Trò chuyện" có chấm đỏ đếm tin chưa đọc; bảng "Bàn tán" bên trái: ảnh đại diện + tên (màu riêng mỗi người) + nội dung; tin mình bên phải viền vàng; bong bóng thoại trên đầu người nói 4 giây. Ban đêm khoá. **Lệnh gửi/nhận tin, giới hạn và ai được nhắn: theo tài liệu trò chuyện của backend** (người chết không nhắn được nếu luật nói vậy).
- **Cảnh lên giàn** (`Trial`): vòng đen khép → mở ra trước giàn; người bị xử (ảnh nửa thân) đứng sau lan can; đám đông bóng lưng tối nhún nhẹ; thẻ chữ "{tên} bị đưa lên giàn" 1.8 giây; khung đếm ngược bên trái (không che mặt người bị xử) + thanh thời gian; tim đập (âm thanh).
- **Bỏ phiếu** (`Verdict`): cùng cảnh; hai tay ngày hiện nhãn "Chạm: tha" / "Chạm: treo"; bảng "Treo x · Tha y · chưa z" từ `view.votes`; người bị xử không có tay bấm.
- **Treo cổ** (sự kiện `Hanged`): 0–0.7s bàn tay pixel trồi lên từ đám đông; 0.5–2.2s nâng người bị xử lên và đẩy ra dưới xà; 1.6–2.8s thòng lọng hạ tới cổ; **3.5s cắt đen ngay** (không chiếu cảnh chết), chữ "{tên} đã bị treo cổ"; mở ra quảng trường, gốc cây có đầu lâu.

## 6. Màn kết thúc (`Ended`)

- `view.winner`: `Wolves` / `Village` / `Lovers`; `view.draw` khi hoà/huỷ; `view.winners` danh sách người thắng (kể cả đã chết); `view.players[]` có `role`, `team`, `card`.
- Tiêu đề phóng to rồi dội vào giữa: "Bầy Sói chiến thắng" (đỏ, vết cào, tàn lửa, cảnh đêm) · "Dân làng chiến thắng" (vàng, tia nắng xoay, cảnh ngày) · "Cặp đôi chiến thắng" (hồng, trái tim bay, chạng vạng) · hoà/huỷ: xám, không hiệu ứng. Dòng phụ = câu cuối của `view.summary`.
- Lưới thẻ 2 hàng theo thứ tự ghế, **úp**, rồi **lật lần lượt cách 0.3 giây** (CSS 3D `rotateY`), tên vai trên dải tối chân thẻ, đầu lâu/trái tim ở mép trên.
- Lật xong: người trong `view.winners` nhô lên, viền sáng màu phe; người thua tối đi. Nút "Ván mới" (chỉ chủ phòng, `newGame`) và "Diễn biến" (hiện `view.summary`).

## 7. Âm thanh

- Tổng hợp bằng WebAudio, không file. Nhạc nền ngày (đàn dây gảy, Rê dorian), đêm (ù trầm, gió, chuông, lửa lách tách), hiệu ứng (bấm, chọn, xác nhận, tin nhắn, vòng đen, tim đập, phiếu, huỵch, lật thẻ, 3 đoạn nhạc thắng). Mã: `Snd` trong `web/prototype/ban-3d.template.html`.
- Bảng **Cài đặt**: bật/tắt + âm lượng riêng Nhạc nền / Hiệu ứng, lưu `localStorage` (bọc try/catch). Âm thanh chỉ bắt đầu sau lần chạm đầu tiên.
- **Bắt buộc theo `secrecy.md` khi chơi trực tiếp (`mode: inPerson`)**: âm thanh không được phụ thuộc vai hay hành động của mình.
  - Nhạc nền và tiếng lửa giống nhau trên mọi máy: được.
  - **Ban đêm tắt mọi hiệu ứng do mình bấm** (chọn, xác nhận, tay) — tiếng "tách" lúc nửa đêm lộ người đang hành động.
  - Không rung, không thông báo hệ thống.
  - Máy loa (`speaker: true`) đọc `say` bằng giọng đọc; nhạc nền nên giảm âm (ducking) khi đang đọc.

## 8. Bí mật (đọc `docs/api/secrecy.md` trước khi làm)

- **Người chết**: màn hình là **lớp che phẳng tối** không đổi khi có tin nhắn; **nhấn giữ** mới thấy cảnh. Ngoại lệ: bộ chọn của Thợ săn chết và Trưởng làng chết hiện công khai, chỉ có tên người còn sống theo `view.players[].alive`.
- Bước Kiểm tra điện thoại: hộp thông báo cùng cỡ, cùng màu cho mọi câu trả lời, nhấn giữ mới hiện.
- Đồng hồ luôn theo `deadline` server, không kết thúc sớm khi đã chọn.
- Không có gì trên màn hình máy khác cho biết ai vừa chọn (chỉ máy người chọn, và bầy sói thấy phiếu của nhau).

## 9. Hiệu năng
- Cảnh vẽ ở ~573×270 rồi phóng to: iPhone 12 chạy mượt. Giữ số material ít (dùng chung material khi được), tránh tạo texture mới mỗi khung hình.
- Tôn trọng `prefers-reduced-motion`: tắt lắc tay, nhún, tàn lửa nhiều, lật thẻ tức thì.
- Precache toàn bộ `web/assets/` (~3.6 MB) và font trong service worker của PWA.

## 10. Sảnh chờ → chia vai → xem vai

- **Sảnh (`Lobby`)**: quảng trường lúc chiều, lửa chưa đốt. Mỗi người vào phòng (`people`, sự kiện `PlayerJoined`) hiện ra trên gốc cây của mình kèm làn bụi và tiếng "bụp"; ghế trống không có ai. Khung trái: **mã phòng** lớn + **mã QR** (link `GET /join/{code}`) + số người. Khung phải: các vai trong ván (thẻ nhỏ × số lượng, từ cấu hình vai), nút Xếp chỗ / Chọn vai / Giờ & âm thanh (chủ phòng), nút **Chia vai** sáng nhấp nháy khi đủ điều kiện (`dealRoles`). Người không phải chủ phòng thấy "Chờ chủ phòng chia vai".
- **Chia vai**: khi pha đổi `Lobby → RoleView`, lửa bùng lên, lá bài úp bay từ đống lửa tới từng người (lệch nhau 90 ms, tiếng sột soạt).
- **Xem vai (`RoleView`)**: thẻ của mình lật ngửa giữa màn, bên phải: tên vai, phe, mô tả, mục tiêu (từ `me.roleCard`). Chạm thẻ = úp/ngửa, chạm đúp = cách chơi. Hết `deadline` (hoặc bấm "Tôi đã xem"): thẻ thu nhỏ bay vào **góc trái trên** thành thẻ nhỏ.

## 11. Trưởng làng

- **Bầu (`ChiefElection`)**: ban ngày, chạm người rồi chạm tay chỉ trỏ để bầu (`voteChief`, đổi được). Nhãn trên đầu hiện số phiếu từ `view.votes.chief`. 2 giây cuối có tiếng trống dồn.
- **Kết quả (`ChiefElected`)**: thẻ Trưởng làng hiện to giữa màn (bật nảy, viền vàng, kèn), dòng chữ "{tên} là Trưởng làng"; camera quay sang người đó; thẻ thu nhỏ bay tới đầu họ → tia sáng vàng, tiếng chuông. Từ đó:
  - **Thẻ nhỏ lơ lửng trên đầu** Trưởng làng (ban đêm trên mái nhà), lắc nhẹ.
  - Nhân vật mặc **trang phục Trưởng làng** giống tranh thẻ: áo choàng xanh két viền vàng, dây chuyền chìa khoá vàng, mặt nạ đen (`paintVillager(..., { chief: true })`).
  - Nếu mình là Trưởng làng: thẻ bay vào thẻ nhỏ ở góc; mục "Thẻ phụ" trong xem thẻ có dòng Trưởng làng.
- **Khi Trưởng làng chết** (`ChiefSuccession`): "{tên} chọn người kế nhiệm…", rồi thẻ bay theo đường vòng từ người cũ sang người mới (`ChiefSucceeded`), trang phục chuyển theo. Hết giờ không chọn (`ChiefLost`): thẻ mờ tan, không ai giữ. *Theo luật §7b là người chết chọn kế nhiệm, không bầu lại; muốn bầu lại phải sửa luật + backend trước.*

## 12. Hiệu ứng khi dùng chức năng ban đêm

Chỉ hiện trên **máy người dùng chức năng** (và cả bầy với phiếu sói). Camera tự quay về nhà mục tiêu; màn hình chớp nhẹ màu của hiệu ứng.

| Vai | Hình ảnh trên nhà mục tiêu | Âm thanh | Kết quả (từ sự kiện) |
|---|---|---|---|
| Bảo vệ | vòm khiên vàng trùm nhà, tia sáng vàng | rải nốt cao lấp lánh | — |
| Sói (phiếu) | 3 vết cào đỏ rạch ngang, khói đỏ | gầm gừ + tiếng cào | — |
| Sói nguyền | lửa xanh lục xoáy quanh nhà bốc lên | âm ma quái trượt xuống | — |
| Sói tiên tri | cột sáng đỏ từ trời rọi xuống | ù trầm dâng lên | "{tên} là / không phải Tiên tri" (`WolfSeerChecked`) |
| Tiên tri | cột sáng xanh + con mắt hiện trên mái | hợp âm huyền bí | "{tên} là Sói / không phải Sói" (`SeerChecked`) |
| Phù thủy · cứu | tia lấp lánh xanh lá bốc lên | rải hợp âm đi lên | (`WitchHealed`) |
| Phù thủy · độc | khói tím cuộn + đầu lâu hiện trên mái | tiếng bọt sôi + trầm | (`WitchPoisoned`) |
| Nguyệt Nữ | trăng khuyết bạc hạ xuống mái, bụi tím | chuông bạc | (`PlayerVisited`) |
| Thám tử | kính lúp lướt qua 3 nhà (mục tiêu + 2 người sống cạnh) | tích tắc + chuông | "Vùng …: có / không có Sói" (`DetectiveChecked`) |
| Cupid | mũi tên hồng bay vòng cung giữa 2 nhà, kéo vệt trái tim | dây cung + đàn hạc | (`LoversPaired`) |

- Kết quả hiện chữ lớn giữa màn (xanh = an toàn, đỏ = Sói) và ghi vào **Ghi chép của bạn**.
- Kết quả **chỉ lấy từ sự kiện server**, không tự suy ra.
- Chơi trực tiếp: âm thanh của các hiệu ứng này tắt khi bật "Im lặng ban đêm" (mặc định **bật** khi `mode: inPerson`).

## 13. Thông báo trời sáng / hoàng hôn

- **Trời sáng (`Dawn` + `DawnAnnounced`)**: lớp nắng vàng dâng từ chân màn rồi tan, tiếng chim hót; **cuộn giấy da** mở từ trên xuống (tiêu đề "Trời đã sáng", câu `say` của server). Có người chết: chuông nhà thờ 2 tiếng, chữ tiêu đề đỏ sẫm, camera quay về chỗ người đó, gốc cây của họ hiện đầu lâu kèm làn bụi. Cuộn giấy tự cuộn lại sau ~4 giây.
- **Hoàng hôn (`Dusk` → `Night`)**: cuộn giấy "Màn đêm buông xuống — Cả làng đi ngủ", tiếng tù và trầm.

## 14. Thẻ của bạn (góc trái trên)

- Thẻ nhỏ 46×62 luôn ở góc trái trên (trừ sảnh, cảnh lên giàn, màn kết thúc). Chạm → khung xem thẻ: thẻ to (chạm = úp/ngửa, chạm đúp = cách chơi), tên vai, phe, mô tả, mục tiêu, **thẻ phụ** (`me.statusCards`: Trưởng làng, cặp đôi, bị nguyền), **Ghi chép của bạn** (kết quả soi, việc đã làm mỗi đêm — dựng từ sự kiện riêng của mình).

## 15. Trò chuyện — luật đã chốt
- Chỉ ở **sảnh** và **ban ngày** (`Day`, `NominationLock`, `Trial`, `Verdict`). **Ban đêm không ai nhắn được. Không có kênh riêng** (kể cả bầy sói).
