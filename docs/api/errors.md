# Errors

A refused message gets an `error` back on that socket only, and changes
nothing (`README.md` "Messages"); an HTTP error has the same body. This page
lists every `code` the server sends, when, and the Vietnamese sentence to
show the player.

## How to show an error

- Show the sentence for the `code`. Never show `message`: it is an English
  detail for logs.
- **Situation.** A few codes mean two things at the table. The app knows
  which one from the message it sent, and picks the row with that
  situation; with no matching situation, it uses the row without one:
  - `wrong_phase` answering a `join` → situation `join`;
  - `bad_credentials` answering a `rejoin` with the saved token → situation
    `token` (by nickname + PIN → the row without a situation).
- **Placeholders.** `%s` is filled by the app: only `pin_locked` has one
  (seconds left, from `retryAfter`).
- **Unknown codes.** The server may add codes within a protocol version.
  A code the app does not know shows the shared sentence of the last group.
- The texts come from the server's `internal/i18n` (a later language adds
  one file there). The tables below are generated from it and from the
  server's source by `make api-docs`, and a test fails when a code is
  missing, so this page always lists every code.
- Some codes close the socket after them: `kicked_by_host` and
  `replaced_by_other_device` (stop reconnecting, show the join screen),
  `message_too_large` over 64 KB, `rate_limited` held for 3 seconds. See
  `README.md` "Joining and rejoining" and "Limits".
- No error ever tells a player something secret: a player who does not hold
  the night step being called always gets the same `not_your_turn`, whatever
  their role or state, and the texts below name no role and nobody's state.
  See [secrecy.md](secrecy.md).

Real error frames: `examples/server/error.<code>.json`.

## Codes

<!-- generated: errors -->
### Joining and rejoining

| code | situation | when | text |
|---|---|---|---|
| `room_not_found` |  | `join`, `rejoin` or `GET /join/{code}` with a code that does not exist or has expired (HTTP 404) | Không tìm thấy phòng này, bạn kiểm tra lại mã phòng nhé. |
| `room_closed` |  | the room is shutting down while the phone joins or acts; `POST /rooms` while the server shuts down (HTTP 503) | Phòng này đã đóng, bạn tạo phòng mới nhé. |
| `invalid_nickname` |  | `join` with an empty name, over 20 characters, or with symbols or emoji (rules §2 "Nickname, rời phòng và đá người") | Tên chỉ gồm chữ, số và dấu cách, dài từ 1 đến 20 ký tự. |
| `nickname_taken` |  | `join` with a name already in the room (case and accents normalised) | Trong phòng đã có người tên này, bạn chọn tên khác nhé. |
| `room_full` |  | `join` when the room already has 30 players | Phòng đã đủ người, không vào thêm được. |
| `wrong_phase` | `join` | `join` once the roles are dealt: a new player cannot enter a running or ended game | Ván đã bắt đầu nên không vào mới được; nếu bạn đã ở trong ván, hãy vào lại bằng tên và PIN. |
| `bad_credentials` |  | `rejoin` with nickname + PIN: wrong PIN or unknown nickname | Tên hoặc PIN không đúng, bạn kiểm tra lại nhé. |
| `bad_credentials` | `token` | `rejoin` with the saved token, which no longer works: another phone took the player over with the PIN, or the host kicked them. Show the join screen | Máy này không còn vào được phòng, bạn vào lại bằng tên và PIN nhé. |
| `pin_locked` |  | 5 wrong PINs for that nickname (rules §2 "Vào lại phòng"). `%s` is the seconds left: `retryAfter` / 1000, rounded up | Bạn nhập sai PIN quá nhiều lần, thử lại sau %s giây nhé. |
| `kicked_by_host` |  | the host kicked this player from the lobby; the socket closes. Do not reconnect | Chủ phòng đã mời bạn ra khỏi phòng. |
| `replaced_by_other_device` |  | the player joined from another phone or tab (by token or by PIN); the socket closes. Do not reconnect | Bạn vừa vào phòng từ một máy khác, máy này đã thoát. |
| `too_many_rooms` |  | `POST /rooms` when the server holds its maximum number of rooms (HTTP 503) | Máy chủ đang bận, bạn thử tạo phòng lại sau ít phút nhé. |
| `too_many_connections` |  | `GET /ws` refused: the server or this network has too many open sockets (HTTP 503) | Máy chủ đang bận, bạn thử lại sau ít phút nhé. |
| `rate_limited` |  | over 10 messages a second; the message was ignored. Staying over the limit for 3 seconds closes the socket | Bạn bấm nhanh quá, chờ một chút nhé. |

### Host setting up the game (lobby)

| code | situation | when | text |
|---|---|---|---|
| `not_host` |  | a host-only message from a player who is not, or no longer, the host (e.g. right after `transferHost`) | Chỉ chủ phòng mới làm được việc này. |
| `unknown_player` |  | `kick` or `transferHost` naming someone who just left the room | Người này không còn ở trong phòng. |
| `seats_mismatch` |  | `setSeats` whose list no longer matches the room: someone joined or left meanwhile | Danh sách chỗ ngồi chưa khớp với người trong phòng, bạn xếp lại nhé. |
| `seats_not_set` |  | `dealRoles` while someone has no seat (joined after `setSeats`) | Còn người chưa được xếp chỗ, bạn xếp chỗ trước khi chia vai nhé. |
| `roles_not_configured` |  | `dealRoles` before any `configureRoles` | Bạn chưa chọn vai cho ván này. |
| `timer_out_of_range` |  | `configureTimers` with a time outside its range (rules §2 "Thời gian mặc định") | Thời gian này nằm ngoài giới hạn cho phép, bạn chọn lại nhé. |
| `too_few_players` |  | `configureRoles` or `dealRoles` with fewer than 5 players | Cần ít nhất 5 người mới chơi được. |
| `role_count_mismatch` |  | `configureRoles` or `dealRoles` when the number of roles is not the number of players (also after a `leave` or `kick`) | Số vai phải bằng đúng số người trong phòng. |
| `no_wolves` |  | the role list has no wolf role | Ván cần ít nhất một Sói. |
| `too_many_wolves` |  | wolves over 30% of the players, the one to be cursed included (rules §2 "Giới hạn số sói") | Số Sói vượt giới hạn cho số người này, tính cả người sẽ bị nguyền. |
| `role_needs_more_players` |  | Detective with 8 players or fewer, Cupid with fewer than 10 (rules §2 "Vai theo số người") | Có vai cần thêm người mới dùng được: Thám tử cần trên 8 người, Cupid cần từ 10 người. |
| `role_required` |  | the role list has no Villager (rules §2 "Số lượng mỗi vai") | Ván cần ít nhất một Dân làng. |
| `already_dealt` |  | `dealRoles` a second time (double tap) | Vai đã được chia rồi. |

### During the game

| code | situation | when | text |
|---|---|---|---|
| `wrong_phase` |  | a command the current phase does not take, usually a tap that lands just after the phase or step ended | Bước này đã qua rồi, bạn không làm được nữa. |
| `not_your_turn` |  | a night command outside the sender's own step, usually a tap just after it ended. Every player who does not hold the step gets exactly this, whatever their role or state ([secrecy](secrecy.md)) | Chưa tới lượt bạn. |
| `blocked` |  | a night command while `me.turn.blocked` (rules §6 "Nguyệt Nữ"); the text is `me.turn.notice`'s and never says who | Bạn không thể sử dụng chức năng đêm nay. |
| `already_acted` |  | a second choice where the first is final: Wolf seer, Cupid, Cursed wolf (double tap) | Bạn đã chọn rồi, không đổi được nữa. |
| `ability_used` |  | `detectiveCheck` a second time in the game (rules §6 "Thám tử") | Bạn đã dùng chức năng này rồi, mỗi ván chỉ được một lần. |
| `potion_used` |  | `witchHeal` or `witchPoison` with that potion already used (rules §6 "Phù thủy") | Bình thuốc này đã dùng hết rồi. |
| `nothing_to_heal` |  | `witchHeal` when nobody is bitten tonight | Đêm nay không có ai bị cắn để cứu. |
| `same_guard_target` |  | `guardProtect` on last night's player (rules §6 "Bảo vệ") | Bạn không được bảo vệ cùng một người hai đêm liền. |
| `same_visitor_target` |  | `visitorVisit` on last night's player (rules §6 "Nguyệt Nữ") | Bạn không được ghé cùng một người hai đêm liền. |
| `invalid_target` |  | a target the role may not pick: a dead player, a wolf for the curse, oneself where not allowed, the same player twice for Cupid | Không chọn được người này, bạn chọn người khác nhé. |

### Only a buggy client or a server bug

All of these, and any code the app does not know, show:

> Có lỗi xảy ra, bạn thử lại hoặc mở lại ứng dụng nhé.

| code | when |
|---|---|
| `bad_message` | malformed JSON, an unknown or wrongly typed field, wrong capitals, a `seed` |
| `unknown_type` | a message `type` the server does not know |
| `unknown_command` | a command `name` the server does not know |
| `message_too_large` | a message over 4 KB |
| `not_joined` | any message before `join` or `rejoin` succeeded |
| `already_joined` | `join` or `rejoin` a second time on the same socket |
| `disconnected` | a message that still reaches the room from a socket another phone replaced (it was told `replaced_by_other_device` first) |
| `secret_in_url` | `GET /ws` with a query string (HTTP 400) |
| `bad_origin` | the page's origin is not configured on the server (HTTP 403) |
| `forbidden` | `kick` or `transferHost` naming oneself |
| `invalid_player_id` | never from the wire: the server makes player ids |
| `duplicate_seat` | `setSeats` listing a player twice |
| `mode_unsupported` | `POST /rooms` with `mode: "online"` (HTTP 400) |
| `invalid_sound_mode` | never from the wire: an unknown sound mode is `bad_message` first |
| `unknown_role` | `configureRoles` with a role that does not exist |
| `invalid_role_count` | `configureRoles` with a count below 1 |
| `role_must_be_unique` | `configureRoles` with more than one of a role allowed once (rules §2 "Số lượng mỗi vai"); cap it in the picker |
| `too_many_players` | never in practice: a room is full at 30 |
| `self_nomination` | `nominate` oneself |
| `already_spared` | `nominate` someone in `votes.spared` |
| `on_trial` | `verdict` from `votes.defendant` |
| `not_alive` | a vote or night command from a dead player |
| `internal` | a bug in the server; the message changed nothing |

<!-- /generated -->
