# Hướng dẫn dự án Ma Sói (cho người không đọc code)

Backend của app quản trò Ma Sói chơi trực tiếp. Mỗi người một điện thoại; máy chủ phòng là loa đọc lời dẫn. **Server là nguồn sự thật duy nhất**: điện thoại chỉ gửi ý định ("tôi soi An") và nhận màn hình của riêng mình.

## 1. Các tầng và quy tắc ai gọi ai

```
điện thoại ──WebSocket──► transport/ws ──► room ──► engine ◄── roles
                              │              │         │
                              │              └► store  └► i18n (câu tiếng Việt)
                         cmd/server, cmd/demo (khởi động, nối dây)
```

| Tầng | Thư mục | Việc của nó |
|---|---|---|
| engine | `internal/engine` | **Luật chơi thuần túy**: trạng thái, lệnh, sự kiện, đêm/ngày, kiểm tra thắng, "view" cho từng người. Không mạng, không giờ, không ngẫu nhiên. |
| roles | `internal/roles` | Mỗi vai một file (Sói, Dân, Bảo vệ, Tiên tri...). Tự đăng ký vào engine. |
| room | `internal/room` | Một phòng = một goroutine: nhận lệnh lần lượt, hẹn giờ, phát view, vào lại phòng (token/PIN), chọn máy làm loa (`SpeakerPolicy`: `InPersonSpeaker` một máy, nhớ máy đang đọc; `AllPhones` mọi máy, chọn bằng lệnh `configureSound`). `Manager` tạo phòng theo mã và xoá phòng bỏ không 2 giờ. |
| transport | `internal/transport/ws` | HTTP + WebSocket: đọc JSON, kiểm tra, đổi thành lệnh; đổi view thành JSON. Không có luật chơi. `wsclient/` là "điện thoại giả" cho test và demo. |
| store | `internal/store` | Kho sự kiện trong RAM (sau này đổi sang database mà không sửa chỗ khác). |
| i18n | `internal/i18n` | Câu đọc tiếng Việt (`vi.go`) và mã câu (`keys.go`). |
| cmd | `cmd/server`, `cmd/demo` | `main`: ghép các tầng lại. |

**Quy tắc cứng** (xem `CLAUDE.md`):
- Chỉ gọi vào trong: `transport → room → engine`, `roles → engine`. Engine không bao giờ import room/transport/store.
- Engine không I/O, không `time.Now()`, không `rand` toàn cục: giờ và seed do `room` đưa vào.
- Mọi thay đổi trạng thái đi qua sự kiện: `Decide(state, lệnh) → sự kiện`, rồi `Evolve(state, sự kiện) → state mới`.
- Server **không bao giờ** gửi nguyên state. Mọi tin gửi đi đều qua `engine.ViewFor` (màn hình hiện tại) và `engine.Visible` (sự kiện người đó được biết). Lộ thông tin là loại lỗi nguy hiểm nhất.
- Mã bí mật (seed chia bài, token, PIN) luôn do server sinh bằng `crypto/rand`, không bao giờ nhận từ điện thoại.

## 2. Công nghệ và lý do

- **Go (chỉ thư viện chuẩn là chính)**: đơn giản, chạy nhiều kết nối đồng thời tốt, test có `-race` bắt lỗi chạy song song. Chỉ thêm `coder/websocket` (kết nối) và `testify` (test).
- **WebSocket**: server phải đẩy tin xuống điện thoại ngay (đến lượt bạn, trời sáng, có người bị treo) mà không để điện thoại hỏi liên tục.
- **JSON**: dễ đọc, dễ gỡ lỗi, app/web nào cũng nói được. Mẫu tin nhắn nằm ở `docs/protocol.md`.
- **Lưu trong RAM**: phiên bản 1 chưa cần database; tắt server là mất phòng. Có interface `store.EventStore` để thay sau.
- **Event sourcing (nhật ký sự kiện)**: ván chơi là danh sách sự kiện; trạng thái là kết quả "gấp" các sự kiện. Nhờ vậy vào lại phòng, xem lại lịch sử khi chết, và viết test kịch bản đều dễ.
- **Một goroutine mỗi phòng**: mọi lệnh và mọi hết giờ của một phòng chạy lần lượt, không bao giờ đua nhau.

## 3. Hành trình của lệnh "Tiên tri soi An"

Điện thoại Tiên tri gửi: `{"type":"command","name":"seerCheck","args":{"target":"p1"}}` (p1 là An). Tin **không** ghi ai gửi; server biết từ kết nối.

**A. Đi vào (transport)**
1. `transport/ws/server.go` → `serveWS`: kiểm tra không có query, kiểm tra Origin, mở WebSocket. (Lúc vào phòng: `handle.go` → `join`/`rejoin`.)
2. `session.go` → `run`: đọc từng tin, gọi `handle`.
3. `handle.go` → `handle`: chặn nếu quá tốc độ (`ratelimit.go` → `allow`) hoặc quá 4 KB.
4. `decode.go` → `messageType`, rồi `handle.go` → `dispatch`, `decodeStrict`: tin phải đúng mẫu, **thừa trường (như `seed`) là bị từ chối**.
5. `decode.go` → `buildCommand` → `commands.go` → `commandTable["seerCheck"]`: tạo `roles.SeerCheck{By: <người gửi>, Target: "p1"}`.

**B. Vào phòng (room)**
6. `room/submit.go` → `Client.Submit`: chuyển việc cho goroutine của phòng qua `room.go` → `do`.
7. Trong đó: kiểm tra kết nối còn là của người này (`live`), `actor.go` → `actorOf` (không giả danh người khác), lệnh dành cho chủ phòng thì kiểm tra `hostOnly`.
8. `room.go` → `execute` → `apply`.

**C. Luật (engine + roles)**
9. `engine/decide.go` → `Decide`: không phải lệnh chung nên hỏi từng vai trong `Registry`.
10. `roles/seer.go` → `Decide`: `engine/night_actor.go` → `NightActor` (đang là bước gọi Tiên tri? còn sống?), chưa soi đêm nay, `LivingTarget` (An còn sống). Sai thì trả lỗi như `ErrNotYourTurn` (danh sách ở `engine/errors.go`). `NightActor` kiểm tra **trước tiên** người gửi có giữ vai không: ai không giữ vai luôn nhận đúng một lỗi `not your turn`, dù bị chặn, đã chết hay sai vai, để lỗi không tiết lộ gì. Đúng thì sinh sự kiện `SeerChecked{Seer, Target, IsWolf}`.

**D. Ghi nhận**
11. `room.go` → `apply`: `store.Append` (ghi nhật ký trước), rồi `engine.Fold` → `Evolve` → `roles/seer.go` → `Evolve` đặt `Pending.SeerTarget`.
12. `room/timers.go` → `rearm`: chỉnh lại hẹn giờ. (Bước đêm **không kết thúc sớm**; luôn chạy đủ giờ để không lộ ai còn sống.) Khi bàn có Sói nguyền, sau lượt Sói nguyền (đêm 1–3) có thêm bước **Kiểm tra điện thoại** (5 giây, `engine/phone_check.go`); khi bàn có Cupid, đêm 1 có thêm một lần kiểm tra ngay sau lượt Cupid, rồi bước **Cặp đôi mở mắt nhìn nhau** (10 giây, `engine/lovers_meet.go`, chỉ hai người yêu thấy đến lượt mình). Lịch các bước này cố định theo danh sách vai và số đêm. Ở mỗi lần kiểm tra, mọi người sống nhận **một** thông báo cùng một dạng, gộp “bị nguyền / không bị nguyền” và “yêu ai / không ở trong cặp đôi”. Người bị nguyền, bầy sói và cặp đôi chỉ thấy thay đổi ở bước này. Thẻ vai (rules §9b) đi kèm view: `engine/view_cards.go` lấy chữ từ `i18n/vi.go` (`me.roleCard`, `me.statusCards`, danh sách vai công khai `roleCards`), frontend không giữ chữ nào của thẻ. Người yêu chết ban ngày thì người kia “đau buồn mà chết” ở đêm kế tiếp, dùng chung cơ chế chết muộn với độc của Hiệp sĩ (`engine/delayed_death.go`).

**E. Gửi kết quả về (chỉ Tiên tri thấy)**
13. `room/broadcast.go` → `changed` → `send`, **cho từng điện thoại có màn hình thay đổi** (máy nào màn hình y như cũ và không có sự kiện mới thì không nhận gì, nên khi Tiên tri soi, các máy khác không nhận tin nào, không ai đoán được Tiên tri còn sống):
    - `engine/view_project.go` → `ViewFor` (màn hình của riêng người đó);
    - `engine/view_audience.go` → `Visible`: dùng `SeerChecked.Audience` (`roles/seer.go`) = chỉ Tiên tri. Người khác không nhận sự kiện này.
14. `transport/ws/session.go` → `Deliver` → `encode.go` → `frames` → `jsonenc.go` → `toTree`: đổi sang JSON (`view` rồi `events`).
15. `session.go` → `writeLoop` gửi xuống điện thoại Tiên tri: sự kiện `SeerChecked` với `isWolf`.

## 4. Triệu chứng → tầng nghi ngờ → file xem đầu tiên

| Triệu chứng | Tầng | Xem trước |
|---|---|---|
| Luật sai (ai thắng, ai chết, thứ tự đêm) | engine / roles | `engine/night_resolve.go`, `engine/win.go` (thắng thua chỉ xét ban ngày, trong `engine/after_death.go`), file vai trong `roles/`; luật gốc `docs/rules.md` |
| Lệnh bị báo `wrong_phase` / `not_your_turn` | engine | `engine/decide.go`, `engine/night_actor.go`, `engine/day_nominate.go` |
| Hết giờ mà không sang pha | room | `room/timers.go` (`rearm`, `fire`), `room/dayclock.go`; rồi `engine/day.go`, `engine/night.go` (timeout cũ bị bỏ qua nếu `Phase`/`Seq` lệch) |
| Chủ phòng/máy làm loa sai người | room | `room/speaker.go` (`InPersonSpeaker`, `AllPhones`), `room/join.go` (`passHost`, `armHostGrace` — chủ phòng rớt mạng chưa quá 60 giây (`hostDisconnectGrace`) thì chưa mất quyền) |
| Vào lại phòng lỗi (`bad_credentials`, `pin_locked`) | room | `room/identity.go`, `room/join.go` (`RejoinToken`, `RejoinPIN`); `transport/ws/handle.go` → `rejoin` |
| Rớt mạng rồi mất chỗ / bị đá | room + transport | `room/join.go` (`Disconnect`, `reattach` gọi `Sink.Evict` khi vào lại cướp chỗ máy cũ), `transport/ws/session.go` (`finish`, `Evict`) |
| Tin nhắn bị từ chối (`bad_message`, `unknown_type`...) | transport | `transport/ws/decode.go`, `commands.go`, mã lỗi ở `errors.go` |
| Lệnh nhận lỗi `internal`, log có "panic in room" | room | `room/room.go` (`safe`): lỗi lập trình trong engine/vai bị bắt lại, phòng chạy tiếp; xem `stack` trong log |
| Bị ngắt kết nối bất ngờ | transport | `ratelimit.go` (flood), `session.go` (hàng đợi đầy = "too slow"; ping/pong không trả lời; quá N lần join/rejoin sai), `server.go` (Origin, `connguard.go` giới hạn kết nối theo IP/tổng) |
| Lộ thông tin (thấy vai người khác, thấy phiếu sói) | engine view | `engine/view_project.go`, `engine/view_audience.go` (`Audience` của sự kiện), `roles/view_*_test.go`; sau đó `transport/ws/jsonenc.go` |
| Câu đọc tiếng Việt sai/thiếu | i18n + transport | `i18n/vi.go`; ghép tên người ở `transport/ws/encode.go` → `announcedPlayers` |
| Định dạng JSON gửi ra sai | transport | `encode.go`, `jsonenc.go`, đối chiếu `docs/protocol.md` |
| Phòng biến mất | room | `room/manager.go` (`sweep`, `IdleTimeout` 2 giờ) |

## 5. Lệnh hay dùng

```bash
make test                      # toàn bộ test (có -race)
make lint                      # go vet + golangci-lint
make fmt                       # định dạng code
make run                       # chạy server (cổng :8080, đổi bằng ADDR=:9000)
make demo                      # server + 5 bot chơi 1 ván, in lời dẫn tiếng Việt
make demo PLAYERS=12 ROLES=CursedWolf=1,WolfSeer=1,Seer=1 SPEED=20   # SPEED: đồng hồ phòng chạy nhanh gấp N lần (chỉ demo; server thật luôn dùng đồng hồ thật). Số người và vai tuỳ chọn (Dân làng tự điền); make demo-12 là bàn 12 người đủ vai, mặc định SPEED=20 (một ván dưới 1 phút)
SIM_GAMES=50 go test ./internal/roles -run TestSim -v          # 300 ván bot ngẫu nhiên (SIM_SEED=<seed> để chạy lại đúng 1 ván lỗi)
go test ./internal/roles -run TestNight_SeerChecks -v      # một test (khớp theo tên)
go test ./internal/transport/ws -run E2E -v                # test cả đường truyền
go test ./... -race -count=20 -run TestTên                 # chạy lặp để bắt test chập chờn
```

- **Xem log**: server in log ra màn hình (`slog`), chỉ ghi phòng, người chơi, loại sự kiện; **không bao giờ** ghi token/PIN. Chỉnh bằng biến môi trường: `ALLOWED_ORIGINS=https://app.example` (danh sách origin được phép), `ALLOW_ANY_ORIGIN=1` (chỉ khi thử nghiệm).
- **Tin `events` có `reset: true` thay cả danh sách sự kiện, không được nối thêm** (ví dụ người chơi vừa chết nên thấy hết lịch sử; các sự kiện từng bị giấu được chèn lại đúng chỗ nên chỉ số dịch đi). Client nối thêm sẽ hiện trùng dòng, đó là nguyên nhân dòng lặp trong demo trước đây. Điện thoại thật sau này phải xử lý đúng: thay danh sách, rồi chỉ hiện phần mới (xem `afterSubsequence` trong `cmd/demo/printer.go`). Chi tiết ở `docs/protocol.md`.
- **Xem một ván bằng mắt**: `make demo`. Để xem tin JSON thật, mở `internal/transport/ws/wsclient` hoặc chạy test E2E với `-v`.

### Viết test kịch bản để tái hiện lỗi

Test kịch bản nằm trong `internal/roles/*_test.go`, dùng bộ dựng `enginetest` để một người không phải lập trình viên cũng đọc được. Tên test là tên luật. Ví dụ (rút từ `roles/seer_test.go`):

```go
func TestNight_SeerChecksACursedPlayer_SeesAWolf(t *testing.T) {
    g := newTable(t)                       // bàn 10 người đã chia vai sẵn
    g.Night().Seer("Em").Checks("Phuc").   // đêm 1: Tiên tri Em soi Phuc
        ExpectSeerSaw("Em", "Phuc", true)  // mong đợi: thấy là Sói
}
```

Cách làm: (1) đặt tên test theo luật bị sai; (2) dựng bàn bằng `newTable` hoặc `enginetest.NewGame`; (3) viết các bước `Night()`, `Day()`...; (4) viết `Expect...` cho kết quả đúng theo `docs/rules.md`; (5) chạy, thấy đỏ rồi mới sửa code. Các bảng trường hợp (table-driven) cho nhiều đầu vào cùng lúc có ví dụ ở `TestNight_SeerChoiceIsRejectedWhenInvalid`. **Lưu ý khi bàn có Nguyệt Nữ (`Visitor`)**: Nguyệt Nữ bắt buộc ghé, hết giờ mà chưa chọn thì server random một người. Vì vậy test nào có Nguyệt Nữ trong bàn phải cho cô ấy ghé tường minh (`n.Visitor("Khoa").Visits("Hoa")`), nếu không cô ấy sẽ ghé ngẫu nhiên và làm người đó bị chặn, khiến kết quả test sai khó hiểu. Muốn test việc random thì dùng `n.Visitor("Khoa").RunsOutOfTime(seed)` với seed cố định. Với lỗi ở phòng/mạng, xem mẫu `room/*_test.go` (đồng hồ giả `FakeClock`) và `transport/ws/*_test.go`.

## 6. Chỗ chỉnh nhanh

| Muốn đổi | File |
|---|---|
| Số người tối thiểu/tối đa (5–30), tỉ lệ Sói tối đa (30%) | `internal/engine/limits.go` (`MinPlayers`, `MaxPlayers`, `MaxWolfPercent`) |
| Thời gian mặc định (xem vai 30s, mỗi bước đêm 30s, ngày 7 phút, bào chữa 1 phút...) | `internal/engine/state.go` → `DefaultSettings` |
| Khoảng cho phép khi chủ phòng chỉnh giờ | `internal/engine/timers_config.go` (`shortTimer`, `dayTimer`, `pauseTimer`) |
| Câu đọc tiếng Việt | `internal/i18n/vi.go` (mã câu ở `keys.go`) |
| Số ký tự mã phòng, bảng chữ, thời gian phòng bỏ không | `internal/room/manager.go` (`CodeLength`, `CodeAlphabet`, `IdleTimeout`) |
| Giới hạn tin nhắn (4 KB, 10 tin/giây, dồn 20, flood 3s, chờ tin đầu 10s, tối đa 8 lần join/rejoin sai) | `internal/transport/ws/config.go` (các hằng `Default...`) |
| Ping/pong (chu kỳ 12s, chờ pong 10s) | `internal/transport/ws/config.go` (`DefaultPingInterval`, `DefaultPongTimeout`), `session.go` (`sendPing`) |
| Giới hạn kết nối (tổng 2000, mỗi IP 100, kết nối mới mỗi IP 30/phút) | `internal/transport/ws/config.go`, `connguard.go` |
| Số lần sai PIN, thời gian khoá, cỡ bảng `fails` | `internal/room/identity.go` (`maxPINFailures`, `pinLockTime`, `maxPINFailEntries`) |
| Thời gian chủ phòng được rớt mạng trước khi mất quyền (60s) | `internal/room/join.go` (`hostDisconnectGrace`) |
| Thêm lệnh cho vai mới | `internal/transport/ws/commands.go` (`commandTable`) + `docs/protocol.md` |

Nhớ: luật chơi mới hay mơ hồ thì **hỏi trước**, không tự bịa (xem `CLAUDE.md`). Sau khi sửa, chạy `make fmt test lint`.
