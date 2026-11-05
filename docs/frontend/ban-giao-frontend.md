# Bàn giao giao diện cho frontend — Ma Sói

Tài liệu này là nguồn chính để code giao diện. Hình mẫu từng màn nằm trên canvas thiết kế "Ma Sói — Wireframe" (tên bảng ghi trong ngoặc, ví dụ `C4`). Luật và dữ liệu: `docs/rules.md`, `docs/api/`. Bí mật: `docs/api/secrecy.md`.

## 1. Khung màn hình

- Máy chuẩn iPhone 12 cầm ngang: **844 × 390**. Chỉ hỗ trợ ngang; máy dọc thì hiện lớp phủ "Xoay ngang điện thoại" (A1).
- Lề an toàn: **47px trái/phải** (tai thỏ), **10px trên/dưới**. Nội dung chính nằm trong 750 × 370.
- Khung chung của mọi màn trong ván, từ trên xuống:
  - Thanh trên cao 40px: nút menu 40×32, tiêu đề (Grenze 28px), dòng phụ, đồng hồ (VT323 34px, vàng).
  - Vùng giữa: cột trái 150px, bàn ghế ở giữa, cột phải 150px, cách nhau 8px.
  - Dải lời dẫn cao 30px ở đáy, đọc đúng chữ `say` của server.

## 2. Màu (token)

| Token | Đêm | Ngày | Dùng cho |
|---|---|---|---|
| `bg` | `#0d1322` | `#bfe3e8` | màu nền dự phòng dưới ảnh nền |
| `overlay` | `rgba(13,19,34,.8)` | `rgba(214,236,228,.5)` | lớp phủ trên ảnh nền cho chữ dễ đọc |
| `band` | `rgba(13,19,34,.55)` | `rgba(243,231,201,.92)` | nền thanh trên và dải lời dẫn |
| `panel` | `#10182a` | `#f3e7c9` (giấy da) | nền khung, ô nhập |
| `fg` | `#e8e0cc` (màu xương) | `#2a2420` (mực) | chữ chính |
| `muted` | `#a9b3b8` | `#4a4236` | chữ phụ, nhãn |
| `line` | `#3a4a5e` | `#6d9566` | đường kẻ, viền nét đứt |
| `btn-bg` / `btn-bd` | `#182340` / `#4a584d` | `#f3e7c9` / `#4a584d` | nút thường |
| `shadow` | `#05070d` | `#2f4029` | bóng đổ 3px của nút, thẻ |

Màu nhấn dùng chung:
- **Vàng `#c59c47`** (viền sáng `#f0d58e`, chữ trên nền vàng `#1a1408`): nút chính, ghế đang chọn, đồng hồ.
- **Máu `#be3424`** / nền nút nguy hiểm `#7a1c1c`: Huỷ ván, đầu độc, đánh dấu người lên giàn (ngày).
- Màu phép, chỉ dùng điểm xuyết: xanh `#6be36b`, tím `#ad4acb`, lửa `#d85834`.
- Ngày dùng thêm: cỏ `#a9d18e` (mặt bàn), lá sẫm `#2f4029`.

## 3. Chữ

- Tải từ Google Fonts: `Grenze Gotisch` (400, 600) và `VT323`. Cả hai đủ dấu tiếng Việt. Không dùng Press Start 2P, Pixelify Sans, Jersey, Jacquard (thiếu dấu tiếng Việt).
- **Grenze Gotisch 600**: tiêu đề màn, tên vai, tên người được chọn (26–30px), chữ to giữa bàn.
- **VT323**: mọi chữ còn lại. Thân 19–20px, nhãn phụ 16px, chữ trong ghế 18px (tên) và 15px (nhãn), đồng hồ 34px.
- VT323 nhìn nhỏ hơn font thường cùng cỡ: đừng để chữ dưới 15px.

## 4. Thành phần

### Nút
Góc vuông, viền 2px, bóng trong `inset -3px -3px 0` và bóng ngoài `3px 3px 0 shadow`. Cao tối thiểu 44px (nút trên thanh trên 32px).

| Loại | Nền | Viền | Chữ | Bóng trong |
|---|---|---|---|---|
| Thường (đêm) | `#182340` | `#4a584d` | `#e8e0cc` | `#0d1322` |
| Thường (ngày) | `#f3e7c9` | `#4a584d` | `#2a2420` | `#d9c99e` |
| Chính | `#c59c47` | `#f0d58e` | `#1a1408` | `#8a6a2a` |
| Nguy hiểm | `#7a1c1c` | `#be3424` | `#f3e3d0` | `#4a0f10` |

Khi nhấn: dịch nút 2px xuống phải và bỏ bóng ngoài.

### Khung (cắt 9 phần từ khung thẻ)
```css
.khung {
  border: 9px solid transparent;
  border-image: url(web/assets/ui/khung-lang-9.webp) 36 / 9px round;
  image-rendering: pixelated;
  background: var(--panel);
  background-clip: padding-box;
}
```
- Phe Sói dùng `khung-soi-9.webp` (nền `#1a1216`). `khung-xam-9.webp` cho trạng thái tắt.
- Hoa văn góc (tuỳ chọn, cho khung quan trọng): `khung-lang-goc.webp` / `khung-soi-goc.webp`, 38×38, đặt tại `-9px` mỗi góc, lật bằng `scaleX(-1)` / `scaleY(-1)` / `scale(-1,-1)`.

### Bàn ghế (`Ban`)
- **5–16 người**: vòng bầu dục. Ghế 84×38, xếp đều theo chiều kim đồng hồ, ghế 1 ở đỉnh. Tâm ghế nằm trên elip có bán kính `W/2 − 44` và `H/2 − 21`.
- **17–30 người**: bàn chữ nhật chiếm cả bề ngang (B-ban-30-nguoi). Ghế 80×36, hoặc 66×36 khi trên 24 người, chạy quanh 4 cạnh. Vùng hành động nằm giữa bàn.
- Ghế: dòng 1 là "số · tên" (18px), dòng 2 là nhãn (15px: "bạn", "Trưởng làng", "→ Phúc", "2 phiếu"…). Tên dài thì cắt "…".
- Trạng thái ghế:

| Trạng thái | Đêm | Ngày |
|---|---|---|
| thường | nền `#182340`, viền `#4a584d` | nền `#f3e7c9`, viền `#4a584d` |
| đang chọn | nền vàng, chữ `#1a1408` | như đêm |
| đánh dấu | viền 3px vàng | viền 3px `#be3424` |
| đã chết | viền nét đứt, chữ gạch ngang `#6b7480`, không bóng | chữ `#7d837c` |
| không chọn được | viền `#26304a`, chữ mờ | viền `#d9d0b6` |

- Mặt bàn: đêm `rgba(13,19,34,.55)` viền đứt `#3a4a5e`; ngày cỏ `#a9d18e` viền `#6d9566`. Giữa bàn có 3 dòng: nhãn nhỏ, chữ to (Grenze), dòng phụ.

### Thẻ vai (hiện ngửa — rules §9b)
- Tranh thẻ tỉ lệ 3:4. Trên màn lượt đêm thẻ cỡ **156×209**, đặt tại (47, 56). Màn Xem vai (C1) 200×268. Bóng `4px 4px 0 #05070d`.
- Chạm 1 lần: úp thẻ (đổi sang `the/mat-sau.webp`). Chạm lần nữa: lật lại. Chạm đúp khi đang ngửa: mở trang hướng dẫn của vai (`GET /guide` → `roles[]`).
- Chữ trên thẻ do code in (tên, phe, mô tả, mục tiêu từ `me.roleCard`), không có chữ trong tranh.

### Hai tay (góc nhìn thứ nhất)
- Ảnh cao 360px. Ban đêm hiển thị **cao 205px**, ban ngày **cao 150px**. Rộng theo tỉ lệ ảnh.
- Tay trái dính góc dưới trái (`left:0; bottom:0`), tay phải dính góc dưới phải. Được phép đè lên thẻ và mép bàn.
- Nhãn trên tay: cách đáy 8px, cách mép 8px, VT323 18px, nền vàng. Sau khi xong: nền `#10182a`, chữ `#a9b3b8`.
- Hai trạng thái cùng bên có cùng kích thước ảnh: đổi `src` là đủ, tay không nhảy.
- Ban ngày, dải lời dẫn thụt vào `padding-left:130px; padding-right:110px` để không bị tay che.

## 5. Tương tác lượt đêm

Quy tắc chung: **chạm một người trên bàn để chọn → chạm "tay làm" để xác nhận**. Không có nút Xác nhận riêng (thay cho Q16). Mọi lựa chọn là cuối cùng, trừ phiếu cắn của bầy sói.

| Vai (`nightCall.role`) | Tay làm | Gửi | Ảnh sau khi xong |
|---|---|---|---|
| `Cupid` | phải (mũi tên), cần chọn đủ 2 người | `cupidPair {a,b}` | cả hai tay `-da-dung` |
| `Visitor` (Nguyệt Nữ) | phải (mặt dây chuyền) | `visitorVisit {target}` | — |
| `Guard` (Bảo vệ) | phải (bàn tay sáng) | `guardProtect {target}` | — |
| `CursedWolf` (Sói nguyền) | trái (lửa nguyền) | `cursedWolfCurse {target}` | tay trái `-da-dung`, giữ suốt ván |
| `WolfSeer` (Sói tiên tri) | phải (cầu đỏ) | `wolfSeerCheck {target}` | — |
| `Werewolf` (bầy sói) | phải (móng vuốt); nút "Không cắn ai" trong khung | `wolfVote {target}` (`""` = không cắn) | đổi phiếu được tới hết giờ |
| `Witch` (Phù thủy) | trái = cứu người bị cắn (chạm thẳng); phải = chọn người rồi chạm để đầu độc | `witchHeal`, `witchPoison {target}` | mỗi bên `-da-dung` khi hết bình, giữ suốt ván |
| `Seer` (Tiên tri) | phải (cầu xanh) | `seerCheck {target}` | — |
| `Detective` (Thám tử) | trái (kính lúp) | `detectiveCheck {target}` | cả hai tay `-da-dung`, giữ suốt ván |

- Khi không tới lượt (C3): vẫn hiện thẻ và tay của vai mình nhưng tay không bấm được, không có nhãn.
- Bị chặn hay trúng độc: thông báo `me.turn.notice` thay vào khung thông tin bên phải, tay không bấm được.
- Ngày: tay chung trong `web/assets/chung/`. Tay trái chỉ trỏ (`-da-vote` sau khi đề cử), tay phải cầm chĩa ba. Thợ săn (C12b) và Trưởng làng (C13) hiện tay riêng, chỉ để trang trí; việc chọn vẫn bằng lưới tên và nút chính.

## 6. Bố cục lượt đêm (thành phần `LuotDem`)

| Phần | Vị trí / cỡ |
|---|---|
| Thanh trên | x 47→797, y 10, cao 40 |
| Thẻ | (47, 56), 156×209 |
| Bàn | (205, 54), 434×300 |
| Khung thông tin | phải 47, y 58, 150×170 — tiêu đề, nhãn, giá trị lớn (30px), gợi ý, nút phụ (tuỳ chọn) |
| Tay | góc dưới, cao 205 |

## 7. Nền và giao diện theo màn

- Nền: `nen/dem.webp` và `nen/ngay.webp`, cùng bố cục, `background-size: cover`. Lúc trời sáng và hoàng hôn, chuyển dần (crossfade) giữa hai ảnh khoảng 1,5 giây.
- **Đêm**: A1–A9, B1–B5, C1, C3, C4*, C5, C14, D1–D2, E1–E8.
- **Ngày**: C2, C6–C13.
- D1 (màn che người chết) là nền phẳng `#0d1322`, không ảnh, không đổi khi có tin nhắn (secrecy.md).

## 8. Tài nguyên ảnh

- Thư mục `web/assets/` sinh bằng `web/tools/chuan_hoa_anh.py`, quy ước ở `docs/frontend/quy-uoc-tai-nguyen-anh.md`.
- Code đọc `web/assets/manifest.json`:
```js
import m from './assets/manifest.json'
const v = m.vai[view.me.role]            // 'Witch' → { thuMuc, the, tay }
const tayTrai = v.tay[daDung ? 'tay-trai-da-dung' : 'tay-trai'] ?? v.tay['tay-trai']
const theUp = m.matSau
const khung = m.ui['khung-lang-9']
```
- Trưởng làng là danh hiệu: `m.danhHieu.chief`.
- Tải trước (precache trong service worker) toàn bộ `web/assets/`, khoảng 3,6 MB, để vào ván không phải chờ.

## 9. Việc còn mở
- Thám tử, tay phải: hai trạng thái lệch vị trí cuốn sổ (cần ảnh mới).
- Dân làng ban ngày: màu da tay trái lệch giữa hai trạng thái.
- Hiệp sĩ kiếm gỉ: đã có ảnh tay nhưng chưa có màn dùng (không có lượt đêm). Có thể hiện ở C3 khi người chơi là Hiệp sĩ.
