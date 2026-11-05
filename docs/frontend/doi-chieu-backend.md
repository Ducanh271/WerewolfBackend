# Đối chiếu giao diện ↔ backend, và quy trình thêm tính năng

Mục đích: giao diện (bản chạy thử pixel 3D, `docs/frontend/giao-dien-3d.md`) không làm điều gì server không cho, và mọi tính năng mới đi đúng đường để hai bên không lệch nhau.

## 1. Nguyên tắc: hợp đồng trước, giao diện sau

Mọi thứ trên màn hình thuộc một trong hai loại:

| Loại | Ví dụ | Cần backend? |
|---|---|---|
| **Chỉ hình ảnh / âm thanh** | camera, ánh lửa, cách vẽ nhà, hoạt ảnh treo cổ, nhạc nền, bảng cài đặt âm thanh | Không. Chỉ sửa `giao-dien-3d.md` rồi code. |
| **Có dữ liệu hoặc lệnh** | ai thắng, ai bị đề cử, trò chuyện, chọn tạo hình nhân vật, thả biểu cảm | Có. Phải có trong `docs/api/` trước khi frontend làm. |

Câu hỏi kiểm tra cho mỗi chi tiết trên màn hình: **"Số liệu / chữ / lệnh này lấy từ trường nào trong `docs/api/screens/*.md`?"** Không chỉ ra được trường nào thì hoặc là chỉ hình ảnh, hoặc là đang tự bịa — bịa thì phải dừng lại.

## 2. Những chỗ bản chạy thử đang lệch backend (phải sửa khi code thật)

| # | Bản chạy thử | Backend nói | Sửa thế nào |
|---|---|---|---|
| 1 | Đề cử một người là **ngay lập tức lên giàn** | Đề cử gom trong pha `Day`; đủ điều kiện mới sang `NominationLock` (đếm ngược, đổi phiếu thì quay lại), rồi `Trial` (`day.md`) | Frontend không tự chuyển cảnh. Chỉ chuyển cảnh khi `view.phase` thành `Trial` / có sự kiện `TrialStarted`. Đề cử đổi được tới khi khoá; có lựa chọn "không ai" (`target: ""`). |
| 2 | Không có nút **Bỏ qua ngày** và **Huỷ ván** | `skipDay {on}`, `abortVote {on}` ở các pha ngày (`day.md`) | Thêm hai nút vào lớp DOM ban ngày (cột phải như C7), hiện số phiếu từ `view.votes.skips`, `abortVotes`. |
| 3 | Phiếu treo/tha là **số giả**, đếm 1 người 1 phiếu | `view.votes.verdicts` live; **Trưởng làng tính 2**; hoà thì tha; người bị xử không bỏ phiếu | Đếm từ `view.votes`, nhân đôi theo `view.chief`; máy người bị xử không hiện tay bấm. |
| 4 | Không có **dải lời dẫn** | Mọi sự kiện có `say` là thông báo: máy loa đọc, **mọi máy hiện chữ** (`README.md`) | Thêm dải chữ lời dẫn ở đáy (như `.sy` trong wireframe), cả trong cảnh 3D. |
| 5 | Ban đêm máy **không tới lượt bị che đen 93%**, máy tới lượt **sáng lên** | `secrecy.md`: độ sáng, hoạt ảnh, thời gian **giống nhau trên mọi máy** — người ngồi cạnh thấy máy sáng lên là biết ai đang dậy | **Cần cậu quyết** (mục 4). |
| 6 | Hiệu ứng âm thanh khi chọn/xác nhận ban đêm | `secrecy.md`: không âm thanh phụ thuộc hành động | Chế độ chơi trực tiếp: tắt tiếng do mình bấm ban đêm (đã ghi trong đặc tả mục 7). |
| 7 | Trò chuyện là tin giả, ai cũng nhắn được | Backend đã có phần nhắn tin nhưng **tài liệu chưa có trong `docs/api/`** | Viết tài liệu trò chuyện vào `docs/api/screens/` + `make api-docs` trước khi frontend làm (mục 3). |
| 8 | Tạo hình dân làng ngẫu nhiên theo số thứ tự | Server không có trường tạo hình | Tạm thời: sinh từ **id người chơi** (`p1`…) để mọi máy giống nhau. Muốn người chơi tự chọn → tính năng backend mới. |
| 9 | Treo cổ: hoạt ảnh ~6 giây | Pha `Dusk` có thời lượng theo `timers` (mặc định 10 giây) | Hoạt ảnh phải xong trước `deadline`; `view` mới đến thì cắt hoạt ảnh. |
| 10 | Chưa có: chờ xem vai, bầu Trưởng làng, Kiểm tra điện thoại, Cặp đôi nhìn nhau, Thợ săn bắn, Trưởng làng chết, màn che người chết, sảnh | Có đủ trong `role-view.md`, `chief.md`, `phone-check.md`, `night.md` (LoversMeet), `hunter.md`, `secrecy.md` (người chết), `lobby.md` | Làm theo `ke-hoach-frontend.md` bước 3, 5, 9. |
| 11 | Màn kết thúc chỉ có 3 kết cục | Còn **hoà** (`view.draw`) và **huỷ ván** (`GameEnded.outcome.aborted`) | Thêm hai biến thể: xám, không hiệu ứng, vẫn lật thẻ. |
| 13 | Bầu Trưởng làng xong thẻ bay tới người thắng; Trưởng làng chết thì thẻ bay sang người kế nhiệm | Luật §7b: người chết **chọn** kế nhiệm, hết giờ thì mất danh hiệu — không bầu lại | Giao diện theo luật (đặc tả mục 11). Muốn bầu lại thì sửa luật + backend trước. |
| 14 | Bản chạy thử tự bịa người chết lúc trời sáng và kết quả soi | Lấy từ `DawnAnnounced` và sự kiện kết quả (`SeerChecked`…) | Không suy ra kết quả trên máy. |
| 12 | Thẻ vai hiện ngửa (Q25) | Repo backend có thể chưa sửa câu "nhấn giữ để xem thẻ" | Chạy `docs/tasks/the-ngua.md` (lệnh `/tinh-nang`). |

## 3. Quy trình thêm một tính năng (có dữ liệu)

1. **Viết phiếu tính năng** (mẫu bên dưới) — cậu tự viết hoặc nhờ Claude trong Project.
2. **Luật**: nếu tính năng đổi cách chơi, sửa `docs/rules.md` trước (số mục rõ ràng).
3. **Backend** (Claude Code trên repo backend): code + test + chạy `make api-docs` để sinh lại `schema/`, `examples/`; viết/sửa `docs/api/screens/<màn>.md` (trường nào, lệnh nào, lỗi nào, ai thấy) và `secrecy.md` nếu liên quan bí mật.
4. **Đặc tả giao diện**: thêm vào `docs/frontend/giao-dien-3d.md` (trông thế nào, lấy từ trường nào).
5. **Frontend** (Claude Code trên repo frontend): `cd web && npm run gen:types` → code → test phát lại `examples/`.
6. Chép lại các file `docs/` đã đổi vào Project để mọi phiên Claude cùng thấy.

**Mẫu phiếu tính năng**

```
Tên: …
Mục đích (người chơi được gì): …
Ai thấy / ai dùng: mọi người | người sống | người chết | một vai | chủ phòng
Lúc nào (pha): …
Dữ liệu server phải gửi (trường trong view / sự kiện): …
Lệnh app gửi lên (tên + tham số) và điều kiện được gửi: …
Lỗi có thể có: …
Bí mật: có làm máy này khác máy kia (hình, tiếng, thời gian) không? …
Màn hình: mô tả + bản chạy thử nếu có
```

## 4. Hai việc cần quyết

1. **Độ sáng ban đêm khi chơi trực tiếp** (dòng 5). Đề xuất: ở chế độ chơi trực tiếp, **mọi máy còn sống ban đêm cùng một cảnh làng tối mờ** (không che đen), mọi máy đều thấy tay của vai mình; máy tới lượt chỉ khác ở **nhãn trên tay và tay bấm được**, không sáng lên. Màn che đen chỉ dùng khi chơi online (mỗi người một nơi). Khi đó cần sửa đoạn "Night" trong `secrecy.md` cho khớp.
2. **Luật trò chuyện** — *đã chốt 04/11*: chỉ ở sảnh và ban ngày, không nhắn ban đêm, không có kênh riêng (kể cả bầy sói). Còn cần ghi rõ trong `docs/api/screens/chat.md`: người chết có nhắn được không, giới hạn độ dài và tốc độ.

## 5. Để máy tự bắt lệch (cài sẵn trong frontend)

- **Kiểu sinh từ schema** (`cd web && npm run gen:types`): backend đổi tên trường → frontend **không build được**, lộ ngay chỗ lệch.
- **`switch` đầy đủ** trên `view.phase` và `event.type` với nhánh `never`: backend thêm pha/sự kiện mới → TypeScript báo thiếu xử lý.
- **Test phát lại `docs/api/examples/`**: mỗi file `view.*.json` dựng được một cảnh không lỗi; mỗi `command.*.json` khớp đúng lệnh app gửi.
- **Kiểm tra trước khi gộp code**: `cd web && npm run gen:types && npm run build && npm test` chạy trên CI mỗi lần đẩy code, kể cả khi chỉ backend đổi.
