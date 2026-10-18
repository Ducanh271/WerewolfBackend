# Lời dẫn của máy làm loa

Bản kiểm kê **mọi câu** máy làm loa (máy đang làm loa, hoặc mọi máy ở chế
độ "Mọi máy cùng phát", §1) đọc ra tiếng trong một ván, để anh duyệt chữ trước
khi cài vào code (audit [L3], `docs/audit.md`). Nguồn luật: `docs/rules.md`
§1–§9.

Không viết code trong bước này. File này chỉ chốt **chữ**; việc nối chữ vào
sự kiện nào, có cần thêm sự kiện công khai mới không, để làm ở bước lập trình
T25 sau khi chữ được duyệt.

## Cách đọc bảng

- **Trạng thái "Có sẵn"**: khoá đã có trong `internal/i18n/vi.go`, chữ lấy
  đúng như đang chạy, không đổi gì — nêu ở đây để thấy đủ cả bức tranh.
- **Trạng thái "MỚI"**: chưa có khoá, chưa có chữ trong code. 7 khoá đầu là
  đúng danh sách `docs/audit.md` [L3] đã liệt ra phải thêm: `NightStarted`,
  `NightStepCalled`, `DayStarted`, `TrialStarted`, `VerdictStarted`,
  `ChiefElectionStarted`, `HunterTimedOut`. Thêm 2 khoá anh vừa quyết định
  thêm ngoài danh sách audit: `DawnStarted`, `RoleViewStarted`.
- **Trạng thái "Có sẵn (đổi chữ)"**: khoá đã có, nhưng anh vừa sửa lại chữ —
  cần đổi trong `vi.go` ở bước lập trình, không phải thêm khoá mới.
- **Trạng thái "Im lặng (theo luật)"**: luật không yêu cầu loa đọc gì ở bước
  này — màn hình tự hiện đếm giờ hoặc nút bấm, không có lời dẫn. Không tự
  thêm chữ ở đây vì luật không nói, theo quy tắc "không tự đặt luật" của
  CLAUDE.md.
- Chữ không bao giờ nêu ai đang sống/chết ngoài những gì chính luật cho công
  khai, không nêu vai trừ Thợ săn (§1), và **lượt gọi giả giống y lượt gọi
  thật** (§3) — cột "An toàn bí mật" nhắc lại điều này ở từng dòng cần.

---

## 1. Xem vai (§2 bước 5)

| Khi nào | Chữ loa đọc | Khoá | Trạng thái |
| --- | --- | --- | --- |
| Ngay sau khi chia vai, mở pha Xem vai có đếm ngược | "Mời mọi người xem vai của mình. Nhớ che màn hình nhé." | `RoleViewStarted` | **MỚI** — luật §2 không bắt loa đọc gì ở bước này; câu này là dẫn chương trình do anh chọn thêm, không phải chữa luật. |

## 2. Bầu Trưởng làng (§7b)

| Khi nào | Chữ loa đọc | Khoá | Trạng thái |
| --- | --- | --- | --- |
| Hết giờ Xem vai, mở pha bầu Trưởng làng | "Mời cả làng bầu Trưởng làng." | `ChiefElectionStarted` | **MỚI** |
| Hết giờ bầu / mọi người đã bầu xong | "{name} được bầu làm Trưởng làng." | `ChiefElected` | Có sẵn |

An toàn bí mật: không có gì để lộ, bầu cử đã công khai theo luật (§7b).

## 3. Ban đêm (§3)

| Khi nào | Chữ loa đọc | Khoá | Trạng thái |
| --- | --- | --- | --- |
| Ngay khi đêm bắt đầu (đêm 1 sau bầu Trưởng làng, đêm sau sau hoàng hôn) | "Trời đã tối, tất cả nhắm mắt lại." | `NightStarted` | **MỚI** — câu lấy đúng câu đầu §3 ("Tất cả nhắm mắt"). |
| Đến lượt gọi của **từng vai có trong ván** (Cupid, Nguyệt Nữ, Bảo vệ, Sói nguyền, Sói tiên tri, Bầy sói, Phù thủy, Tiên tri, Thám tử), **thật hay giả giống nhau** | "{Vai} hãy mở mắt, thực hiện xong thì nhắm mắt lại." | `NightStepCalled` | **MỚI** — một khoá dùng chung cho mọi vai, chỉ đổi tên vai (xem bảng tên dưới). |
| Bước Kiểm tra điện thoại (0b, 3b — §3) | "Mọi người mở mắt xem điện thoại, rồi nhắm mắt lại." | `NightPhoneCheck` | Có sẵn |
| Bước Cặp đôi mở mắt nhìn nhau (0c — §3, đêm 1 có Cupid) | "Cặp đôi hãy mở mắt nhìn nhau." | `NightLoversMeet` | Có sẵn |

### Tên vai dùng khi gọi (cho khoá `NightStepCalled` mới)

Dùng đúng tên đã có trong `i18n.roleCards[...].Name`, trừ **Bầy sói**: thẻ
vai của "Werewolf" ghi "Sói" (số ít, dùng khi giới thiệu vai), nhưng lúc gọi
đêm phải gọi đúng tên bầy theo đúng chữ bảng §3: **"Bầy sói"**.

| RoleID | Tên gọi khi vào lượt đêm |
| --- | --- |
| `Cupid` | Cupid |
| `Visitor` | Nguyệt Nữ |
| `Guard` | Bảo vệ |
| `CursedWolf` | Sói nguyền |
| `WolfSeer` | Sói tiên tri |
| `Werewolf` | **Bầy sói** (không phải "Sói") |
| `Witch` | Phù thủy |
| `Seer` | Tiên tri |
| `Detective` | Thám tử |

`Hunter`, `RustyKnight`, `Villager` không có lượt gọi đêm (§3, §6) — không
xuất hiện trong bảng này.

An toàn bí mật: câu "{Vai} hãy mở mắt..." **giống nhau tuyệt đối** cho vai
còn sống và vai đã chết/vắng mặt (gọi giả, §3), và không đổi theo người giữ
vai có hành động hay không. Hai bước Kiểm tra điện thoại và Cặp đôi mở mắt
nhìn nhau không cần thêm câu "hãy mở mắt" riêng — chữ của chính hai bước đó
(đã có sẵn) đủ làm hiệu lệnh, tránh đọc hai câu chồng nhau.

## 4. Buổi sáng (§4 "Thông báo buổi sáng")

| Khi nào | Chữ loa đọc | Khoá | Trạng thái |
| --- | --- | --- | --- |
| Đêm vừa kết thúc, trước khi đọc ai chết | "Trời đã sáng, mọi người mở mắt." | `DawnStarted` | **MỚI** |

| Tình huống | Chữ loa đọc | Khoá | Trạng thái |
| --- | --- | --- | --- |
| Có người chết | "Đêm qua, {tên} đã chết." (nhiều người: đọc theo thứ tự ghế) | `DawnSomeoneDied` | Có sẵn |
| Không ai chết vì Bảo vệ che đúng người bị cắn | "Đêm qua không ai chết, có người đã được bảo vệ." | `DawnProtected` | Có sẵn (đổi chữ) |
| Không ai chết vì lý do khác | "Không có ai chết." | `DawnNobodyDied` | Có sẵn |

An toàn bí mật: không nêu nguyên nhân chết hay vai (§4); ba câu trên là toàn
bộ những gì luật cho phép loa nói buổi sáng.

## 5. Ban ngày (§4 "Thảo luận và xử trảm")

| Khi nào | Chữ loa đọc | Khoá | Trạng thái |
| --- | --- | --- | --- |
| Ngay sau thông báo buổi sáng, mở giờ thảo luận + đề cử | "Cả làng bắt đầu thảo luận và đề cử." | `DayStarted` | **MỚI** |
| Đếm ngược 10 giây chốt đề cử (§4 bước 2) | *(không có)* | — | Im lặng (theo luật) — luật chỉ nói "bắt đầu đếm ngược chốt 10 giây", không có câu loa đọc; màn hình tự hiện số đếm lùi. |
| Chốt xong, người lên giàn, mở giờ bào chữa | "{name} lên giàn bào chữa." | `TrialStarted` | **MỚI** |
| Hết giờ bào chữa, mở vote giết/tha | "Hết giờ bào chữa. Cả làng bỏ phiếu: treo hay tha." | `VerdictStarted` | **MỚI** |
| Treo | "{name} đã bị treo." | `DayHanged` | Có sẵn |
| Tha | "Làng tha cho {name}." | `DaySpared` | Có sẵn |
| Hết giờ ban ngày, không ai được đề cử | "Không ai bị treo." | `DayNobodyHanged` | Có sẵn |

An toàn bí mật: tên người lên giàn đã công khai từ đề cử (§4), không phải
thông tin bí mật.

## 6. Thợ săn (§6 "Thợ săn")

| Khi nào | Chữ loa đọc | Khoá | Trạng thái |
| --- | --- | --- | --- |
| Thợ săn chết (bất kỳ lý do), ngay sau thông báo người chết | "{name} là Thợ săn." | `HunterRevealed` | Có sẵn |
| Trong lúc chờ Thợ săn chọn (15 giây) | *(không có)* | — | Im lặng (theo luật) — rules §6: "cả bàn im lặng chờ 15 giây". |
| Thợ săn bắn xong | "{hunter} bắn {target}." | `HunterShot` | Có sẵn |
| Hết 15 giây, Thợ săn chưa chọn | "Hết giờ, {hunter} không bắn ai." | `HunterTimedOut` | **MỚI** |

## 7. Trưởng làng chết / kế nhiệm (§7b)

| Khi nào | Chữ loa đọc | Khoá | Trạng thái |
| --- | --- | --- | --- |
| Trưởng làng chết, trong lúc chờ 15 giây chọn người kế nhiệm | *(không có)* | — | Im lặng (theo luật) — §7b chỉ nói việc chọn, không nói loa đọc gì lúc chờ; người chết đã được công bố ở mục 4 hoặc mục 6 rồi. |
| Chọn xong | "Trưởng làng mới là {name}." | `ChiefSucceeded` | Có sẵn |
| Hết giờ, không chọn | "Trưởng làng không kịp chọn người kế nhiệm, danh hiệu đã mất." | `ChiefLost` | Có sẵn |

## 8. Huỷ ván (§8b)

| Khi nào | Chữ loa đọc | Trạng thái |
| --- | --- | --- |
| Trong lúc phiếu huỷ đang diễn ra (không có pha riêng, không dừng gì — §8b) | *(không có)* | Im lặng (theo luật) — phiếu hiện công khai trên màn hình như mọi phiếu ban ngày khác, không có "bắt đầu phiếu huỷ" để đọc. |
| Phiếu huỷ đạt 2/3 | dùng chung chữ "Bàn đã bỏ phiếu huỷ ván. Ván kết thúc, không có phe thắng." (xem mục 9 — `GameAborted`, trong `GameEnded`) | Có sẵn |

## 9. Kết thúc ván (§8)

| Kết quả | Chữ loa đọc | Khoá | Trạng thái |
| --- | --- | --- | --- |
| Phe Dân thắng | "Phe Dân chiến thắng, tất cả bầy sói đã bị diệt." | `WinVillage` | Có sẵn |
| Phe Sói thắng | "Phe Sói chiến thắng." | `WinWolves` | Có sẵn |
| Cặp đôi khác phe thắng | "Cặp đôi khác phe chiến thắng." | `WinLovers` | Có sẵn |
| Hòa (mọi người chết cùng lúc) | "Tất cả đã chết, ván đấu hòa." | `WinDraw` | Có sẵn |
| Huỷ ván | "Bàn đã bỏ phiếu huỷ ván. Ván kết thúc, không có phe thắng." | `GameAborted` | Có sẵn |

Bảng tổng kết chi tiết sau ván (T28a, audit [K4]: ai chết đêm nào, ai làm gì)
là **màn hình chữ**, không đọc bằng giọng — quá dài để đọc, và chỉ hiện sau
khi ván đã kết thúc, lúc ai cũng được xem toàn bộ (§4, §8). Câu thắng/thua ở
trên là phần duy nhất của "kết thúc ván" mà loa đọc ra tiếng.

---

## Tổng hợp khoá MỚI đã duyệt

7 khoá theo đúng danh sách audit [L3]:

| Khoá | Chữ đã duyệt |
| --- | --- |
| `NightStarted` | "Trời đã tối, tất cả nhắm mắt lại." |
| `NightStepCalled` | "{Vai} hãy mở mắt, thực hiện xong thì nhắm mắt lại." (tên vai theo bảng ở mục 3) |
| `ChiefElectionStarted` | "Mời cả làng bầu Trưởng làng." |
| `DayStarted` | "Cả làng bắt đầu thảo luận và đề cử." |
| `TrialStarted` | "{name} lên giàn bào chữa." |
| `VerdictStarted` | "Hết giờ bào chữa. Cả làng bỏ phiếu: treo hay tha." |
| `HunterTimedOut` | "Hết giờ, {hunter} không bắn ai." |

2 khoá thêm ngoài danh sách audit, anh vừa quyết định thêm:

| Khoá | Chữ đã duyệt |
| --- | --- |
| `DawnStarted` | "Trời đã sáng, mọi người mở mắt." |
| `RoleViewStarted` | "Mời mọi người xem vai của mình. Nhớ che màn hình nhé." |

1 khoá có sẵn, đổi chữ:

| Khoá | Chữ cũ | Chữ mới đã duyệt |
| --- | --- | --- |
| `DawnProtected` | "Không có ai chết, được bảo vệ." | "Đêm qua không ai chết, có người đã được bảo vệ." |

Tất cả đã duyệt. Bước tiếp theo (`guide-text.md`) làm riêng; chưa cài gì vào
`internal/i18n/vi.go` ở bước này.
