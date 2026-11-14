# Giao diện Ma Sói — bản đặc tả cho frontend (hướng pixel 3D)

Tài liệu này nói **giao diện trông thế nào và chạy theo dữ liệu server ra sao**. Luật và dữ liệu là của backend: `docs/rules.md`, `docs/api/` (README, `screens/*.md`, `secrecy.md`, `errors.md`, `schema/*.json`). Khi tài liệu này và `docs/api/` vênh nhau, **`docs/api/` đúng**.

**Chờ backend:** thẻ úp ở góc và link `/j/{code}` là quyết định đã chốt nhưng chưa có trong `docs/api/`; xem các việc B1–B4 trong `ke-hoach.md`. Chưa xong việc nào thì chưa code phần phụ thuộc vào nó.

Bản chạy thử tham khảo: `web/prototype/ban-3d.html` (không nằm trong git; tạo bằng `python3 web/prototype/build.py`, rồi mở thẳng bằng trình duyệt). Mã nguồn dễ đọc: `web/prototype/ban-3d.template.html`, `web/prototype/villager.js`. Bản chạy thử dùng dữ liệu giả; app thật lấy mọi thứ từ `view` và `events`.

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
| Trước khi vào phòng | 2D: trang đầu, tạo phòng (có ô chọn chế độ), vào phòng, PIN | `POST /rooms {mode}`, `join`, `rejoin` |
| `Lobby` | **Sảnh 3D** trên quảng trường, hiện chế độ phòng (mục 10). Các bảng cài đặt (xếp chỗ, chỉ `inPerson`; chọn vai; giờ & âm thanh) là 2D chồng lên | xem `screens/lobby.md` |
| `RoleView` | Quảng trường ban ngày, thẻ vai ngửa ở giữa | chạm = úp/ngửa, chạm đúp = cách chơi |
| `ChiefElection` | Ban ngày, chạm người để bầu | `voteChief` |
| `Night` + `nightCall.role` | Ban đêm, nhà | xem mục 4 |
| `Night` + `nightCall.role = PhoneCheck` (`me.phoneCheck`), chỉ `inPerson` | Ban đêm, hộp thông báo **nhấn giữ mới hiện** | `phone-check.md` |
| `Night` + `nightCall.role = LoversMeet` | Ban đêm như mọi bước; máy cặp đôi có `me.turn.step = LoversMeet`, không gửi gì. `online`: máy cặp đôi hiện hình xác nhận (mục 12) | `night.md` |
| `Dawn` | Chuyển đêm → ngày; đọc `DawnAnnounced`; ghế người chết chuyển đầu lâu | — |
| `Day` | Ban ngày: đề cử, trò chuyện | `nominate`, `skipDay`, `abortVote` |
| `NominationLock` | Ban ngày, đếm ngược chốt đề cử | — |
| `Trial` + `votes.defendant` | **Cảnh lên giàn**: camera ra trước giàn, đám đông bóng lưng, đếm ngược theo `deadline` | người bị đề cử thanh minh (nói thật quanh bàn / trò chuyện) |
| `Verdict` | Cùng cảnh giàn; **tay trái = Tha, tay phải = Treo**; bảng phiếu | `verdict {kill}` (trừ người bị xử) |
| Sự kiện `Hanged` (lúc `Verdict` kết thúc) | **Cảnh treo cổ** chạy theo sự kiện, không theo pha (mục 5) | — |
| Sự kiện `Spared` | Thẻ chữ "Dân làng tha cho …". Sau đó là `Day` nếu còn giờ, `Dusk` nếu hết giờ | — |
| `Dusk` | Chuyển ngày → đêm (mục 13) | `abortVote` |
| `HunterShot` | **Cảnh Thợ săn** (mục 17): lật thẻ Thợ săn → máy Thợ săn (`me.canShoot`) vào ống ngắm; máy khác chờ | `hunterShoot` |
| `ChiefSuccession` | Máy Trưởng làng chết (`me.id == view.succession`): tay chìa khoá/cuộn giấy + chọn | `nameSuccessor` |
| `Ended` | **Màn kết thúc lật thẻ** (mục 6) | `newGame` (chủ phòng) |
| Người chết (mọi pha trừ `Ended`) | Lớp che phẳng, nhấn giữ mới xem (mục 8) | 2 bộ chọn công khai ở trên là ngoại lệ |

Bản chạy thử có nhiều chỗ lệch backend (tự lên giàn khi đề cử, thiếu nút Bỏ qua/Huỷ ván, phiếu giả, thiếu dải lời dẫn…): danh sách và cách sửa ở `doi-chieu-backend.md` — **luôn làm theo backend**.

Nguyên tắc: **màn hình dựng lại hoàn toàn từ `view` mới nhất** (mỗi `view` thay cái trước). Hoạt ảnh chuyển cảnh chạy khi **pha đổi** hoặc khi có sự kiện tương ứng trong `events` (`TrialStarted`, `Hanged`, `Spared`, `GameEnded`…). Hoạt ảnh không được làm chậm trạng thái thật: nếu `view` mới đến giữa hoạt ảnh, cắt hoạt ảnh và vẽ theo `view`.

## 4. Ban đêm

- `nightCall.role` cho biết vai đang được gọi (giống nhau trên mọi máy). Thanh trên: "Đêm N · {Vai} đang được gọi" + đồng hồ theo `deadline`.
- **Bảng "Đang gọi"** (giữa, dưới thanh trên; **giống hệt nhau trên mọi máy còn sống**, cả máy đang tới lượt, cả hai chế độ): ảnh thẻ của vai đang gọi (`nightCall.role`; `Werewolf` → thẻ Sói, chữ "Bầy sói"; `PhoneCheck`/`LoversMeet` → mặt sau thẻ + tên bước), dòng nhỏ "Đang gọi", tên vai to, đồng hồ `00:ss` theo `deadline`, câu `say` của sự kiện `NightStepCalled` (một dòng, cắt "…"), thanh thời gian vàng rút dần theo `nightCall.duration`. Khi bước đổi: bảng mở lại kiểu cuộn dọc (0,5 s, `steps(5)`); bảng không có tiếng riêng, chỉ có nhạc hiệu của bước (mục 7). Ẩn khi cuộn giấy thông báo đang mở và trong cảnh cắt. Không bao giờ ghi tên người chơi.
- **Hình ảnh theo chế độ phòng** (`view.mode`, bảng đầy đủ ở mục 16):
  - `inPerson`: mọi máy còn sống cùng một cảnh làng tối mờ, không màn che đen, không chớp sáng toàn màn, không đổi độ sáng theo lượt. Mọi máy thấy tay của vai mình. Máy tới lượt (`me.turn.step`) chỉ thêm nhãn trên tay và tay bấm được. Ban đêm không máy nào phát tiếng, trừ máy loa đọc `say` và phát nhạc hiệu (mục 7).
  - `online`: như bản chạy thử. Không tới lượt (`me.turn` vắng) thì màn **tối hẳn** (che 93%), tay hạ xuống và tối đi. Tới lượt thì căn phòng sáng lên, tay vai mình trồi lên, hiệu ứng đầy đủ kèm chớp màu và âm thanh.
- **Cách bấm, cả hai chế độ: chạm một ngôi nhà để chọn → chạm "tay làm" để gửi lệnh** (Q27). Bảng vai → tay → lệnh ở `docs/frontend/ban-giao-frontend.md` mục 5.
- `me.turn.blocked/doomed/grief`: hiện `me.turn.notice`, tay không bấm được.
- **Bầy sói** (`view.pack`): đồng bọn hiện là sprite sói trước nhà, nhãn viền đỏ; nhà sói không có "zzz", không chọn được. Phiếu từng con sói (`pack.votes`) hiện nhãn "→ tên" trên nhà bị chọn. Nút "Không cắn ai" → `wolfVote {target: ""}`.
- Nhà người còn sống có "zzz" bay lên — **trên mọi máy giống nhau** (kể cả nhà sói khi người xem không phải sói).
- Phù thủy: tay trái = cứu (chạm thẳng, `witchHeal`), tay phải = chọn nhà rồi chạm để đầu độc (`witchPoison`).

## 5. Ban ngày

- Đề cử: chạm người → chạm tay trái (chỉ trỏ) → `nominate {target}`. Nhãn trên đầu: số phiếu đề cử (`view.votes`).
- **Trò chuyện** (luật ở mục 15): nút "Trò chuyện" có chấm đỏ đếm tin chưa đọc; bảng "Bàn tán" bên trái: ảnh đại diện + tên (màu riêng mỗi người) + nội dung; tin mình bên phải viền vàng; bong bóng thoại trên đầu người nói 4 giây. **Lệnh gửi/nhận tin, giới hạn và ai được nhắn: `docs/api/screens/day.md` mục Chat**, tóm tắt ở mục 15.
- **Cảnh lên giàn** (`Trial`): vòng đen khép → mở ra trước giàn; người bị xử (ảnh nửa thân) đứng sau lan can; đám đông bóng lưng tối nhún nhẹ; thẻ chữ "{tên} bị đưa lên giàn" 1.8 giây; khung đếm ngược bên trái (không che mặt người bị xử) + thanh thời gian; tim đập (âm thanh).
- **Bỏ phiếu** (`Verdict`): cùng cảnh; hai tay ngày hiện nhãn "Chạm: tha" / "Chạm: treo"; bảng "Treo x · Tha y · chưa z" từ `view.votes`; người bị xử không có tay bấm.
- **Treo cổ** (sự kiện `Hanged`): 0–0.7s bàn tay pixel trồi lên từ đám đông; 0.5–2.2s nâng người bị xử lên và đẩy ra dưới xà; 1.6–2.8s thòng lọng hạ tới cổ; **3.5s cắt đen ngay** (không chiếu cảnh chết), chữ "{tên} đã bị treo cổ"; mở ra quảng trường, gốc cây có đầu lâu.
  - Chạy theo sự kiện `Hanged`, không chờ `Dusk`. Sau khi treo, pha kế có thể là `HunterShot`, `ChiefSuccession` hoặc `Ended`, và đồng hồ của pha đó đã chạy. Khi đó **rút gọn còn cắt đen + chữ ~1 giây** rồi vẽ theo `view`. Chỉ khi pha kế là `Dusk` mới chạy đủ, và phải xong trước `deadline`.

## 6. Màn kết thúc (`Ended`)

- `view.winner`: `Wolves` / `Village` / `Lovers`; `view.draw` khi hoà/huỷ; `view.winners` danh sách người thắng (kể cả đã chết); `view.players[]` có `role`, `team`, `card`.
- Tiêu đề phóng to rồi dội vào giữa: "Bầy Sói chiến thắng" (đỏ, vết cào, tàn lửa, cảnh đêm) · "Dân làng chiến thắng" (vàng, tia nắng xoay, cảnh ngày) · "Cặp đôi chiến thắng" (hồng, trái tim bay, chạng vạng) · hoà/huỷ: xám, không hiệu ứng. Dòng phụ = câu cuối của `view.summary`.
- Lưới thẻ 2 hàng theo thứ tự ghế, **úp**, rồi **lật lần lượt cách 0.3 giây** (CSS 3D `rotateY`), tên vai trên dải tối chân thẻ, đầu lâu/trái tim ở mép trên.
- Lật xong: người trong `view.winners` nhô lên, viền sáng màu phe; người thua tối đi. Nút "Ván mới" (chỉ chủ phòng, `newGame`) và "Diễn biến" (hiện `view.summary`).

## 7. Âm thanh

- Tổng hợp bằng WebAudio, không file. Nhạc nền ngày (đàn dây gảy, Rê dorian), đêm (ù trầm, gió, chuông, lửa lách tách), hiệu ứng (bấm, chọn, xác nhận, tin nhắn, vòng đen, tim đập, phiếu, huỵch, lật thẻ, 3 đoạn nhạc thắng). Mã: `Snd` trong `web/prototype/ban-3d.template.html`.
- Bảng **Cài đặt** có ở cả hai chế độ, riêng từng máy: bật/tắt + âm lượng riêng Nhạc nền / Hiệu ứng, lưu `localStorage` (bọc try/catch). Âm thanh chỉ bắt đầu sau lần chạm đầu tiên. **Không có nút "Im lặng ban đêm"**: chế độ phòng quyết định.
- **`inPerson`**: âm thanh không được phụ thuộc vai hay hành động của mình.
  - Ban ngày: máy loa (`speaker: true`) phát nhạc nền và tiếng hiệu ứng. Máy khác **mặc định im**, tự bật được trong Cài đặt; bật lên thì phát giống nhau, không phụ thuộc vai.
  - **Ban đêm im hết**, kể cả máy đã tự bật. Chỉ máy loa đọc `say` và phát nhạc hiệu (dưới).
  - Không rung, không thông báo hệ thống.
- **`online`**: mọi máy đủ nhạc nền và hiệu ứng cả ngày lẫn đêm, kể cả tiếng hiệu ứng vai. Mỗi máy tự đọc `say` (`sound: allPhones`, mọi máy có `speaker: true`).
- Máy có `speaker: true` đọc `say` bằng giọng đọc; nhạc nền giảm âm (ducking) khi đang đọc.
- **Nhạc hiệu từng vai** (khi bước đêm bắt đầu: pha `Night` có `nightCall.role` mới / sự kiện `NightStepCalled`): mỗi bước có một đoạn 1,5–2,5 s riêng, phát **trước** câu `say`. Chỉ phụ thuộc `nightCall.role` (công khai), không phụ thuộc vai của người cầm máy nên không lộ gì. `online`: mọi máy phát. `inPerson`: **chỉ máy loa phát** (12 máy kêu lệch nhau nghe rối); đây là ngoại lệ thứ hai của "ban đêm im hết", cùng với câu `say`. Mã: `CALL` trong `Snd`.

  | `nightCall.role` | Nhạc hiệu |
  |---|---|
  | `Cupid` | đàn hạc vuốt lên + hai nốt "tim" |
  | `Visitor` (Nguyệt Nữ) | chuông celesta rơi xuống thang ngũ cung, nền trầm |
  | `Guard` (Bảo vệ) | tù và ba nốt đi lên + tiếng khiên kim loại |
  | `CursedWolf` (Sói nguyền) | hợp âm răng cưa trầm tối + gió |
  | `WolfSeer` (Sói tiên tri) | hai nốt thuỷ tinh nghịch tai + nhịp trống trầm |
  | `Werewolf` (Bầy sói) | tiếng sói tru (hai con, rung) + ba nhịp trống |
  | `Witch` (Phù thuỷ) | nồi thuốc sủi bọt + giai điệu năm nốt quái |
  | `Seer` (Tiên tri) | arpeggio pha lê đi lên + lấp lánh |
  | `Detective` (Thám tử) | đồng hồ tích tắc + ba nốt bí ẩn đi xuống |
  | `PhoneCheck` (chỉ `inPerson`) | chuông đôi nhẹ |
  | `LoversMeet` | tiếng tim đập + hợp âm ấm |

## 8. Bí mật (đọc `docs/api/secrecy.md` trước khi làm)

- **Người chết**: màn hình là **lớp che phẳng tối** không đổi khi có tin nhắn; **nhấn giữ** mới thấy cảnh. Ngoại lệ: bộ chọn của Thợ săn chết và Trưởng làng chết hiện công khai, chỉ có tên người còn sống theo `view.players[].alive`.
- Bước Kiểm tra điện thoại (chỉ `inPerson`): hộp thông báo cùng cỡ, cùng màu cho mọi câu trả lời, nhấn giữ mới hiện.
- Đồng hồ luôn theo `deadline` server, không kết thúc sớm khi đã chọn.
- Không có gì trên màn hình máy khác cho biết ai vừa chọn (chỉ máy người chọn, và bầy sói thấy phiếu của nhau).
- Cái gì áp dụng cho chế độ nào: bảng ở mục 16.

## 9. Hiệu năng
- Cảnh vẽ ở ~573×270 rồi phóng to: iPhone 12 chạy mượt. Giữ số material ít (dùng chung material khi được), tránh tạo texture mới mỗi khung hình.
- Tôn trọng `prefers-reduced-motion`: tắt lắc tay, nhún, tàn lửa nhiều, lật thẻ tức thì.
- Precache toàn bộ `web/assets/` (~3.6 MB) và font trong service worker của PWA.

## 10. Sảnh chờ → chia vai → xem vai

- **Trước sảnh** (2D, DOM): trang đầu, tạo phòng (có ô chọn **Chơi quanh bàn / Chơi online**), vào phòng, PIN. Dựng từ `screens/lobby.md` và token ở `ban-giao-frontend.md` mục 2–4.
- **Sảnh (`Lobby`)**: 3D, quảng trường lúc chiều, lửa chưa đốt. Mỗi người vào phòng (`people`, sự kiện `PlayerJoined`) hiện ra trên gốc cây của mình kèm làn bụi và tiếng "bụp"; ghế trống không có ai. Khung trái: **mã phòng** lớn + **mã QR** (mở thẳng app, ví dụ `https://<tên-miền>/j/K7QP`; dựng từ `location.origin`, không ghi cứng tên miền) + số người + **chế độ phòng**. Khung phải: các vai trong ván (thẻ nhỏ × số lượng, từ cấu hình vai), nút Xếp chỗ (chỉ `inPerson`) / Chọn vai / Giờ & âm thanh (`online` không hiện giờ `phoneCheck`) / đổi chế độ (`configureMode`) (chủ phòng, các bảng này là 2D), nút **Chia vai** sáng nhấp nháy khi đủ điều kiện (`dealRoles`). Người không phải chủ phòng thấy "Chờ chủ phòng chia vai".
- **Chia vai**: khi pha đổi `Lobby → RoleView`, lửa bùng lên, lá bài úp bay từ đống lửa tới từng người (lệch nhau 90 ms, tiếng sột soạt).
- **Xem vai (`RoleView`)**, cả hai chế độ: thẻ của mình lật ngửa giữa màn, bên phải: tên vai, phe, mô tả, mục tiêu (từ `me.roleCard`). Chạm thẻ = úp/ngửa, chạm đúp = cách chơi. Hết `deadline` (hoặc bấm "Tôi đã xem"): chữ mờ đi, **thẻ tự lật úp** (0,6 s, tiếng lật), rồi thu nhỏ bay vào **góc trái trên** thành thẻ nhỏ **mặt sau** (0,7 s, tiếng vút), thẻ nhỏ nảy nhẹ khi tới.

## 11. Trưởng làng

- **Bầu (`ChiefElection`)**: ban ngày, chạm người rồi chạm tay chỉ trỏ để bầu (`voteChief`, đổi được). Nhãn trên đầu hiện số phiếu từ `view.votes.chief`. 2 giây cuối có tiếng trống dồn.
- **Kết quả (`ChiefElected`)**: thẻ Trưởng làng hiện to giữa màn (bật nảy, viền vàng, kèn), dòng chữ "{tên} là Trưởng làng"; camera quay sang người đó; thẻ thu nhỏ bay tới đầu họ → tia sáng vàng, tiếng chuông. Từ đó:
  - **Thẻ nhỏ lơ lửng trên đầu** Trưởng làng (ban đêm trên mái nhà), lắc nhẹ.
  - Nhân vật mặc **trang phục Trưởng làng** giống tranh thẻ: áo choàng xanh két viền vàng, dây chuyền chìa khoá vàng, mặt nạ đen (`paintVillager(..., { chief: true })`).
  - Nếu mình là Trưởng làng: thẻ bay vào thẻ nhỏ ở góc; mục "Thẻ phụ" trong xem thẻ có dòng Trưởng làng.
- **Khi Trưởng làng chết** (`ChiefSuccession`): "{tên} chọn người kế nhiệm…", rồi thẻ bay theo đường vòng từ người cũ sang người mới (`ChiefSucceeded`), trang phục chuyển theo. Hết giờ không chọn (`ChiefLost`): thẻ mờ tan, không ai giữ. *Theo luật §7b là người chết chọn kế nhiệm, không bầu lại; muốn bầu lại phải sửa luật + backend trước.*

## 12. Hiệu ứng khi dùng chức năng ban đêm

Chỉ hiện trên **máy người dùng chức năng** (và cả bầy với phiếu sói). Camera tự quay về nhà mục tiêu.

- `online`: hiệu ứng đầy đủ như bảng, màn hình chớp màu của hiệu ứng, có âm thanh.
- `inPerson`: vẽ dịu trên nhà mục tiêu, **không chớp và không làm cả màn sáng lên**, không có cột âm thanh.

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

- Kết quả hiện chữ lớn giữa màn (xanh = an toàn, đỏ = Sói) và ghi vào **Ghi chép của bạn**. Ở `inPerson` chữ không được làm cả màn sáng lên.
- Kết quả **chỉ lấy từ sự kiện server**, không tự suy ra.

**Tin riêng khi chơi `online`** (`phone-check.md` mục Online). Không có bước Kiểm tra điện thoại, nên:

| Lúc nào | Máy nào | Hiện gì | Lấy từ |
|---|---|---|---|
| Cupid vừa bắn | máy Cupid | mũi tên như bảng trên | `LoversPaired` |
| Cupid vừa bắn, hoặc hết giờ server tự ghép | máy hai người bị bắn | cảnh trúng tên (mục 18), kèm chữ của thẻ phụ | sự kiện `PhoneChecked`, rồi `view.lover`, `me.statusCards` |
| Bước `LoversMeet` | máy hai người yêu | cảnh "trái tim" để xác nhận là một cặp (mục 18) | `me.turn.step`, `view.lover` |
| Sói nguyền vừa nguyền | máy người bị nguyền | cảnh bị nguyền (mục 18), kèm chữ của thẻ phụ | sự kiện `PhoneChecked`, rồi `me.cursed`, `me.statusCards` |
| Sói nguyền vừa nguyền một người đang yêu | máy hai người yêu | người bị nguyền: cảnh bị nguyền; người kia: cảnh sợi chỉ đổi màu (mục 18) | sự kiện `PhoneChecked`, rồi `view.lover.colour`, `me.statusCards` |

Hình bầy sói thấy đồng bọn mới lấy từ `view.pack`, không có sự kiện riêng. Hai người bị bắn không được biết ai là Cupid.

## 13. Thông báo trời sáng / hoàng hôn

- **Trời sáng (`Dawn` + `DawnAnnounced`)**: lớp nắng vàng dâng từ chân màn rồi tan, tiếng chim hót; **cuộn giấy da** mở từ trên xuống (tiêu đề "Trời đã sáng", câu `say` của server). Có người chết: chuông nhà thờ 2 tiếng, chữ tiêu đề đỏ sẫm, camera quay về chỗ người đó, gốc cây của họ hiện đầu lâu kèm làn bụi. Cuộn giấy tự cuộn lại sau ~4 giây.
- **Hoàng hôn (`Dusk` → `Night`)**: cuộn giấy "Màn đêm buông xuống — Cả làng đi ngủ", tiếng tù và trầm.

## 14. Thẻ của bạn (góc trái trên)

- Thẻ nhỏ 46×62 luôn ở góc trái trên (trừ sảnh, cảnh lên giàn, màn kết thúc), cả hai chế độ, **luôn úp (mặt sau giống nhau cho mọi vai)** để người ngồi cạnh không nhìn trộm được. Chạm → thẻ úp phóng to từ góc ra giữa (0,45 s) rồi **mới lật ngửa**; đóng (nút Đóng hoặc chạm nền) → lật úp rồi tắt. Khung xem thẻ: thẻ to (chạm = úp/ngửa, chạm đúp = cách chơi), tên vai, phe, mô tả, mục tiêu, **thẻ phụ** (`me.statusCards`: Trưởng làng, cặp đôi, bị nguyền), **Ghi chép của bạn** (kết quả soi, việc đã làm mỗi đêm — dựng từ sự kiện riêng của mình).
- Ở `inPerson`, ban đêm khung xem thẻ giữ cùng tông tối của cảnh đêm, không làm màn sáng hơn.

## 15. Trò chuyện — luật đã chốt
Theo `docs/rules.md` §4 "Nhắn tin" và `docs/api/screens/day.md` mục Chat, **như nhau ở cả hai chế độ**:

| Pha | Ai được nhắn |
|---|---|
| `ChiefElection`, `Day`, `NominationLock` | mọi người còn sống |
| `Trial` | chỉ người trên giàn (`view.votes.defendant`) |
| còn lại, gồm sảnh, `RoleView`, `Verdict`, ban đêm, `Ended` | không ai |

- Người chết đọc được, không nhắn được. Mọi máy đọc được lúc nào cũng được; hết ván vẫn đọc được.
- Không có kênh riêng (kể cả bầy sói và cặp đôi). Chỉ hiện tin sau `NewGameStarted` gần nhất.
- 200 ký tự một tin; 1 tin/2 giây, dồn tối đa 5; 2000 tin một ván. Hiện `data.text` dạng chữ thuần, không HTML, không tạo link.

## 16. Hai chế độ chơi

Chủ phòng chọn khi tạo phòng (`POST /rooms {mode}`), đổi được trong sảnh trước khi chia vai. **Mọi máy theo chế độ của phòng, không có công tắc riêng từng máy.** Code rẽ nhánh theo `view.mode`, không theo cài đặt của máy. Đổi chế độ bằng lệnh `configureMode {mode}`.

| | Chơi quanh bàn (`inPerson`) | Chơi online (`online`) |
|---|---|---|
| Âm thanh mặc định của phòng | `oneSpeaker` | `allPhones` |
| Xếp chỗ | chủ phòng xếp (`setSeats`) | không có nút; server bốc lúc chia vai (`view.players` trống trong sảnh) |
| Kiểm tra điện thoại | có, theo lịch | không có bước này; người bị bắn, bị nguyền được báo ngay trên máy mình (mục 12, 18) |
| Trao đổi ban ngày | nói trực tiếp và nhắn tin | chỉ nhắn tin; gọi thoại làm sau |
| Cảnh đêm, không tới lượt | cùng một cảnh làng tối mờ trên mọi máy còn sống | màn tối hẳn (che 93%), tay hạ xuống |
| Cảnh đêm, tới lượt | cùng cảnh đó, chỉ thêm nhãn trên tay và tay bấm được; độ sáng không đổi | căn phòng sáng lên, tay trồi lên |
| Tay của vai mình ban đêm | mọi máy đều thấy | chỉ rõ khi tới lượt |
| Hiệu ứng vai (mục 12) | vẽ dịu, không chớp, không làm cả màn sáng | đầy đủ, có chớp màu |
| Tiếng ban đêm | im hết, trừ máy loa đọc `say` và phát nhạc hiệu | nhạc nền và hiệu ứng đầy đủ |
| Tiếng ban ngày | máy loa phát; máy khác mặc định im, tự bật được | mọi máy phát |
| Đọc lời dẫn `say` | một máy loa | mỗi máy tự đọc |
| Cài đặt bật/tắt + âm lượng từng máy | có | có |

**Giống nhau ở cả hai chế độ:**
- Đồng hồ theo `deadline`, lượt gọi chạy đủ giờ, gọi giả như gọi thật.
- Server không gửi gì ngoài phần máy đó được biết.
- Lớp che người chết và 2 bộ chọn công khai (mục 8).
- Thẻ: `RoleView` ngửa, thẻ nhỏ trong ván luôn úp (mục 10, 14).
- Luật nhắn tin (mục 15).

**Đổi chế độ trong sảnh:** âm thanh của phòng về mặc định của chế độ mới; chủ phòng vẫn chỉnh lại được (`configureSound`). Chỗ đã xếp bị bỏ.

**Ván mới** (`newGame`): giữ chế độ, như giờ và âm thanh; chủ phòng đổi được ở sảnh. `online` thì vòng tròn được bốc lại lúc chia vai.

## 17. Thợ săn chết (`HunterShot`)

Bản chạy thử: nút dev "Thợ săn chết: bạn" / "người khác". Thứ tự:
1. **Báo tử trước**: nếu Thợ săn chết trong đêm, cảnh trời sáng (mục 13) chạy xong mới tới đây; nếu chết vì treo cổ, cảnh treo (mục 5) chạy xong mới tới.
2. **Lật thẻ** (sự kiện `HunterRevealed`): thẻ úp phóng to giữa màn, lật ngửa ra thẻ Thợ săn, chữ "X là Thợ săn!", tiếng ngân + tiếng lên cò. Chỉ lộ **vai**, không lộ phe (`hunter.md`).
3. **Máy Thợ săn** (`view.me.canShoot`):
   - **Ống ngắm to**: vòng tròn đường kính ~92% chiều cao màn, vạch chữ thập; ngoài vòng chỉ **tối mờ 55%** (vẫn thấy người xung quanh). Camera zoom nhẹ (FOV 64 → 48). Hai tay cầm súng (`vai/tho-san/hai-tay.webp`) ở giữa đáy, cao 40% màn. Dải trên "Bạn là Thợ săn — bắn một người" + đồng hồ theo `deadline`, tiếng tim đập. Ẩn dòng gợi ý dưới đáy và chữ ở hai tay.
   - **Tên trong ống ngắm**: nhãn đỏ ngay dưới tâm ghi tên người đang ngắm ("Chưa ngắm ai" nếu chưa có). Nhãn tên trên đầu mọi người nổi **trên** lớp ống ngắm.
   - **Tâm tự bám**: người còn sống gần tâm nhất (theo chiều ngang màn, trong ±0,3) tự thành người được chọn, có tiếng "tách" nhỏ khi đổi người; thả tay sau khi vuốt thì camera hút về đúng người đó. Chỉ bám khi camera đã đứng yên.
   - **Dải tên**: hàng nút (ảnh đại diện + tên) dưới dải trên, chỉ người **còn sống theo `view.players[].alive`**, trừ chính Thợ săn; người đang ngắm tô đỏ. Chạm tên → camera quay tới người đó. Dài quá thì cuộn ngang.
   - **Nút ‹ ›** hai bên giữa màn: nhảy sang người kế bên trái/phải (theo góc nhìn, không theo thứ tự ghế).
   - Chạm súng hoặc nút "Bắn {tên}" → `hunterShoot {target}`. Không có lựa chọn "không bắn".
4. **Máy khác**: dải "X (Thợ săn) đang ngắm…" + đồng hồ. Chế độ **online**: chấm laser đỏ lượn qua dân làng (chỉ trang trí, **không** phải chỗ Thợ săn ngắm thật — server không gửi). Chế độ **trực tiếp** (`inPerson`): không laser, không tiếng, chờ im lặng (rules §6, `secrecy.md`).
5. **Phát súng** (sự kiện `HunterShot`): tiếng súng, chớp sáng toàn màn, rung màn, vệt đạn từ chỗ Thợ săn tới người bị bắn, người đó ngã khỏi gốc cây → đầu lâu + bụi; cuộn giấy "{Y} đã bị Thợ săn bắn chết" (câu `say` của server), chuông 1 tiếng.
6. **Hết giờ** (`HunterTimedOut`): theo luật hiện tại **không ai bị bắn** → cuộn giấy "Thợ săn không bắn ai". (Bản chạy thử đang làm "súng tự cướp cò, bắn ngẫu nhiên" theo ý chủ dự án — **chỉ làm khi luật và backend đã đổi** qua `/tinh-nang`; client không bao giờ tự chọn người chết.)
7. Thợ săn bị Thợ săn khác bắn: lặp lại từ bước 2 với người mới.

## 18. Kiểm tra điện thoại, trúng tên, bị nguyền, cặp đôi (`PhoneCheck`, `PhoneChecked`, `LoversMeet`)

Nguồn: `docs/api/screens/phone-check.md` (cả mục Online), `night.md` ("Lovers look at each other"), rules §1b, §3, §7. Bản chạy thử: nút dev "Chế độ: Online/Trực tiếp", "Người yêu: …", "KTĐT: trúng tên / bị nguyền / cả hai / không gì". Bản chạy thử chiếu các cảnh online trong một bước Kiểm tra điện thoại; app thật **không** làm vậy, vì phòng online không có bước đó (xem "Chế độ online").

**Chế độ trực tiếp (`inPerson`) — giữ nguyên quy tắc bí mật:** bước `nightCall.role = PhoneCheck`, dữ liệu `me.phoneCheck`. Một hộp cùng cỡ, cùng màu cho mọi câu trả lời, chữ "Nhấn giữ để xem"; chỉ khi nhấn giữ mới hiện `loveText` và `cursedText`. Không hoạt ảnh, không tiếng riêng. Đồng hồ theo `deadline`.

**Chế độ online:** không có bước `PhoneCheck` và không có `me.phoneCheck` (`phone-check.md` mục Online). Server báo ngay lúc việc xảy ra, chỉ cho máy có tin mới, bằng **một sự kiện `PhoneChecked`** (`cursed`, `love.inLove`, `love.partner`, `love.partnerRole`, `love.colour`) kèm `view` mới (`view.lover`, `me.cursed`, `me.statusCards`). Sự kiện quyết định **có chạy cảnh không**. `view` quyết định **vẽ gì**: một máy vào lại sẽ nhận `view` mà không có sự kiện, nên chỉ vẽ trạng thái, không chạy lại cảnh.

- Lớp phủ toàn màn, tối tím; trên cùng "Đêm N"; bảng "Đang gọi" vẫn chạy bên dưới theo bước hiện tại (bước Cupid hoặc Sói nguyền).
- Chữ: **nguyên văn `text` của thẻ phụ** trong `me.statusCards` (`lovers`, `lovers_cross`, `cursed`), không tự viết câu. Màu sợi chỉ lấy từ `love.colour`: `white` / `black` / `striped`.
- Có `view` mới hoặc pha đổi giữa cảnh thì cắt cảnh và vẽ theo `view`. Cảnh phải xong trước `deadline` của bước đang chạy (bước Cupid hoặc Sói nguyền).

Chọn cảnh theo sự kiện `PhoneChecked` của mình:

| Sự kiện | Lúc nào | Cảnh |
|---|---|---|
| `love.inLove`, máy **chưa có** `view.lover` trước đó | Cupid vừa bắn, hoặc hết giờ server tự ghép | 1. Trúng tên |
| `cursed` | Sói nguyền vừa nguyền mình | 2. Bị nguyền |
| `love.inLove`, `cursed` sai, máy **đã có** `view.lover` | người yêu của mình vừa bị nguyền | 3. Sợi chỉ đổi màu |

1. **Trúng tên**:
   - 0–0,43 s: mũi tên pixel (đuôi lông trắng, thân vàng, đầu tim hồng, dài ~300px) bay vòng cung từ trên trái vào giữa màn (tiếng vút).
   - 0,43 s: cắm phập — rung màn 4 nhịp, mũi tên rung, chớp hồng tỏa tròn, **tim lớn** bật ra (0 → 1,4 → 1), 18 tim nhỏ nổ tung ra xung quanh (tiếng phập + đàn hạc).
   - 1,2 s: mũi tên mờ đi; **thẻ của mình** (ngửa) trượt vào từ trái, **thẻ người yêu** (úp) từ phải; giữa hai thẻ là **sợi chỉ** dài 190px theo `love.colour`: trắng (`white`) / đen viền hồng (`black`) / trắng sọc đen (`striped`); tim lớn thu nhỏ đặt giữa sợi chỉ, đổi màu theo sợi chỉ, đập nhịp tim.
   - 1,9 s: thẻ người yêu lật ngửa ra vai `love.partnerRole`; tên người yêu dưới thẻ (tiếng tim đập).
   - Dòng chữ: `text` của thẻ phụ cặp đôi (hồng nhạt) + chú thích màu: "Sợi chỉ trắng · cặp đôi Phe Dân" / "Sợi chỉ đen · cặp đôi Phe Sói" / "Sợi chỉ trắng sọc đen · Phe thứ ba — thắng khi chỉ còn hai bạn sống" (rules §7).
   - Không lộ ai là Cupid (Cupid nhận `LoversPaired` riêng, hiệu ứng ở mục 12).
2. **Bị nguyền**:
   - Khói tím đỏ từ 8 phía bò vào (tiếng ù trầm).
   - Thẻ vai của mình hiện giữa, to hơn; trong ~0,9 s **vết nứt pixel** chạy từ mép thẻ vào trong (lõi đen, viền đỏ cam, tiếng rạn), thẻ ngả màu đỏ úa, viền phát sáng đỏ.
   - **Đôi mắt sói đỏ** cháy lên trên thẻ (nháy 4 nhịp), thẻ giật 3 lần (tiếng sói tru xa).
   - Dòng chữ: `text` của thẻ phụ `cursed` (đỏ) + **"Phe Sói"** to (lấy từ `me.roleCard.team` mới). Nếu mình đang yêu: thêm dòng nhạt là `text` của thẻ phụ cặp đôi mới, không có chú thích màu.
3. **Sợi chỉ đổi màu** (~2 s): thẻ của mình (ngửa) và thẻ người yêu (ngửa, vai `love.partnerRole`) hiện hai bên, sợi chỉ cũ chuyển dần sang màu mới `love.colour` (thường là `striped`), tim giữa sợi chỉ đổi màu theo và đập 2 nhịp (tiếng tim đập). Dòng chữ: `text` của thẻ phụ cặp đôi mới + chú thích màu như cảnh 1. **Không** nói người yêu bị nguyền và không hiện hình nguyền: thẻ phụ của server nói gì thì hiện đúng thế.

Máy bầy sói thấy đồng bọn mới từ `view.pack`, không có sự kiện và không có cảnh riêng (mục 12).

**Cặp đôi nhìn nhau (`LoversMeet`, 10 s) — "trái tim":** có ở cả hai chế độ.
- Mọi máy: bảng "Đang gọi: Cặp đôi" như mọi bước (mục 4).
- Máy hai người yêu (`me.turn.step` `LoversMeet`): đêm sáng lên một phần (che 30%); camera quay về **nhà người yêu** (`view.lover.partner`); một **dòng tim nhỏ** màu sợi chỉ bay liên tục từ chỗ mình tới nhà người ấy (vẽ đè lên lửa); **tim lớn** đập trên mái nhà, mỗi 1,2 s chớp hồng ở nhà người ấy (+ tiếng tim đập khi online). Bảng "Đang gọi" dời xuống giữa đáy màn để không che nhà, ẩn dòng gợi ý.
- `inPerson`: giống hệt về hình, **không tiếng** (những người khác đang nhắm mắt).
- **Sau đó, suốt ván** (chỉ máy hai người yêu): nhãn người yêu viền hồng, dòng nhỏ "♥ người bạn yêu"; khung xem thẻ có thẻ phụ cặp đôi ("bạn yêu X, vai Y" + màu sợi chỉ) và thẻ phụ bị nguyền (`me.statusCards`).

**Âm thanh (online):** vút, phập, đàn hạc, tim đập, ù trầm, rạn, sói tru, chuông trăng — mã `fx.pc` trong `Snd` (chuông trăng là của cảnh "không gì" trong bản chạy thử; app thật không dùng vì online không báo người không có tin). ⚠ Chơi online qua gọi thoại (Discord/Zalo) thì mic có thể thu tiếng này từ loa người khác; nên kèm gợi ý "đeo tai nghe" ở sảnh khi chọn chế độ online.
