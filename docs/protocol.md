# Protocol

JSON over one WebSocket per screen. Implemented by `internal/transport/ws`.
Every message is an object with a `type` field. Field names are camelCase,
durations are milliseconds, times are Unix milliseconds.

## Rooms and the socket (HTTP)

| request | reply |
|---|---|
| `POST /rooms` (body optional: `{"mode": "inPerson"}`) | `201 {"room": "K7QP", "hostKey": "…"}`; `mode: "online"` → `400` with `mode_unsupported`; `503` `too_many_rooms` |
| `GET /join/{code}` | `200 {"room": "K7QP"}`, or `404` `room_not_found`. This is the link the QR code will carry (the UI is built later); case is ignored |
| `GET /guide` | `200` the in-app rules guide (below). Public, same for everyone, needs no room: usable from the lobby before anyone has joined |
| `GET /ws` | upgrades to the game socket |

### `GET /guide` (T25)

The in-app "rules guide" and "how to play your role" screens, in Vietnamese,
rendered server-side from the real settings and constants (a number here can
never drift from the number the game actually uses). Always `200`, same body
for every caller:

```json
{
  "quickRules": ["Ma Sói chơi bằng điện thoại của chính mỗi người, không cần ai làm quản trò riêng.", "..."],
  "roles": [
    {"role": "Werewolf", "title": "Sói", "whenCalled": "...", "howTo": "...", "onTimeout": "...", "gotcha": "..."},
    {"role": "RustyKnight", "title": "Hiệp sĩ kiếm gỉ", "whenCalled": "...", "howTo": "...", "onTimeout": "...", "gotcha": "...", "note": "..."}
  ],
  "statuses": [
    {"role": "chief", "title": "Trưởng làng (danh hiệu công khai, không phải vai riêng)", "whenCalled": "...", "howTo": "...", "onTimeout": "...", "gotcha": "...", "note": "..."},
    {"role": "lovers", "title": "Cặp đôi (...)", "...": "..."},
    {"role": "cursed", "title": "Bị nguyền (...)", "...": "..."}
  ],
  "firstTime": ["Che màn hình ...", "..."],
  "tips": ["Tiên tri: ...", "..."]
}
```

- `quickRules`: the 12-line "Luật nhanh" (`docs/guide-text.md` §1), fixed text, no settings number.
- `roles`: one entry per role in `engine.Registry` (`i18n.RoleGuides`), in the rules §9b card order. `note` is omitted when the role has none.
- `statuses`: the three states that are not a role — Chief, Lovers, Cursed (`i18n.StatusGuides`) — same shape as `roles`.
- `firstTime`: the "Lần đầu chơi?" checklist (`i18n.FirstTimeChecklist`).
- `tips`: optional strategy tips (`i18n.OptionalTips`), fixed text.
- Every number in `howTo`/`whenCalled`/`title` (the Hunter's shot time, the lovers looking at each other, a dead Chief's time to name a successor, the role-view time, the PIN retry/lock numbers, Cupid's and the Detective's minimum table size) is read from `engine.DefaultSettings()`, the roles' own `MinPlayers()`, and `room.MaxPINFailures`/`room.PINLockTime` — never typed into the text (`internal/i18n/guide.go`).

- The room code is 4 characters from `ABCDEFGHJKMNPQRSTUVWXYZ23456789` (no O, 0, I,
  1, L). It is not a secret. `hostKey` is: only the phone that created the room
  ever sees it, and only its first `join` needs to carry it (rules §2 "chủ
  phòng là người tạo phòng", audit S8). Presenting it again later, including
  after the title has moved on (`transferHost` or a dropped host, below),
  grants nothing.
- A room nobody has joined, left, dropped from or sent a command to for **2 hours**
  is deleted, and its sockets are closed.
- `/ws` takes **no query string** (a request with one gets `400` `secret_in_url`):
  the secret code and the PIN only ever travel inside messages, because URLs end
  up in logs. The room code goes in the first message, `join` or `rejoin`, which
  must arrive within 10 s. That 10 s is counted once, from when the socket
  opens, not reset by later messages; a connection that fails to join or
  rejoin 8 times in a row is also closed (bogus tokens, PINs guessed across
  nicknames).
- **Origin.** The `Origin` header must be one of the configured origins
  (`ALLOWED_ORIGINS`, comma separated, `scheme://host[:port]`). No `Origin` is
  refused too, unless `ALLOW_ANY_ORIGIN=1` (demo and tests), which allows
  everything. Refused → `403` `bad_origin`.
- **Connection limits.** The upgrade to `/ws` is refused with `503`
  `too_many_connections` when the server already holds its configured global
  cap of open sockets, when one IP already holds its (deliberately high —
  a table of players usually shares one IP) concurrent cap, or when that IP
  has opened new connections faster than its per-minute allowance (the real
  control on connection-flood abuse; defaults 2000 / 100 / 30 per minute).
- **Heartbeat.** The server pings every connection roughly every 12 s; a
  connection that does not pong within 10 s of a ping is closed with no
  further notice (the next connected phone takes over as speaker, rules §1).
  This is how a phone that lost signal, rather than one that closed cleanly,
  is noticed.
- **Limits per connection.** A message over 4 KB gets `error` `message_too_large`
  (over 64 KB the connection is closed). More than 10 messages per second
  (a burst of 20 is allowed) gets `error` `rate_limited` and the message is
  ignored; staying over the limit without a break for 3 s closes the connection.
- **Strict decoding.** An unknown `type` is `unknown_type`, an unknown command
  `unknown_command`. Unknown fields anywhere (a `seed`, a `by`, a `voter`),
  wrong field types, wrong capitals and trailing data are `bad_message`. A
  wrong field type names the field by its JSON key, e.g.
  `field "target" has wrong type`.
- A dropped connection never removes the player; reconnect with `rejoin`. A
  rejoin from another phone — by token (another tab, a reload racing the old
  one) or by nickname + PIN — tells the old socket why and closes it right
  away, without waiting for it to send anything: `error` `disconnected`,
  message `"disconnected: joined from another device"` for a token takeover,
  `"someone joined with your PIN"` for a PIN takeover.
- `welcome` always comes first, before the first `view`.

The secrecy rule: the server never sends `engine.State`. Every outgoing game
message is built from `engine.ViewFor` (the screen now) and `engine.Visible`
(the events this screen may know about). See `internal/engine/view.go` for who
sees what.

## Identity

- A `PlayerID` is **public**: it appears in every view (seats, votes). It is
  not a credential.
- On join the server returns a secret `token` (the phone stores it) and a
  4-digit `pin` (shown only on that player's screen). How they are used to
  come back is rules §2 "Vào lại phòng": `rejoin` with the token (same phone)
  or with nickname + PIN (other phone); a locked PIN rejoin is `pin_locked`.
- The connection decides who acts. Commands never carry `by`/`voter`; the
  server fills it from the connection.

## Client → server

### Joining
```json
{"type": "join",     "room": "K7QP", "nickname": "An", "hostKey": "…"}
{"type": "rejoin",   "room": "K7QP", "token": "…"}
{"type": "rejoin",   "room": "K7QP", "nickname": "An", "pin": "4821"}
{"type": "leave"}
{"type": "kick",         "target": "p5"}
{"type": "transferHost", "target": "p5"}
```
`hostKey` is only ever sent by the room's creator, on their own first `join`
(see above); every other phone omits it.

Reply: `{"type": "welcome", "player": "p3", "token": "…", "pin": "4821"}` (on
rejoin by token only `player`; rejoin by nickname + PIN returns a **new** `token`
and the old one stops working), then a full `view` and
a full `events` message. A refused rejoin is an `error` with code
`bad_credentials` or `pin_locked` (with `retryAfter` in ms). `join` with an
invalid nickname (empty, over 20 characters, or outside Vietnamese/Latin
letters, digits and spaces — audit S4) is `error` `invalid_nickname`.

`leave` (lobby only) removes the player for good; a closed connection keeps their
place. Creating a room in `Online` mode is refused with `mode_unsupported`.

`kick` (lobby only, host only; audit K2) removes another player for good,
the same as `leave` would for themselves: their token and PIN are revoked and
their connection, if any, is told why (`error` `disconnected`, message
"removed from the lobby by the host") and closed. Kicking yourself is
`forbidden`; kicking after the deal is `wrong_phase`; a non-host is
`not_host`. Either `leave` or `kick` only drops that one seat — the rest of
the seating and the chosen role list stay as they were; `dealRoles` still
checks the list against the new player count (rules §2) and refuses with the
usual error if it no longer fits.

`transferHost` gives another player in the room the host's title right away,
no vote and no need for them to agree, at any time including mid-game (so the
table can line up who will run the next game, T27b, without waiting for this
one to end). Only the current host may send it; transferring to yourself is
`forbidden`, to someone not in the room is `unknown_player`, and sending it
while not the host is `not_host`.

### Commands
```json
{"type": "command", "name": "nominate", "args": {"target": "p5"}}
```

| name | args | when | engine command |
|---|---|---|---|
| `setSeats` | `order: [id]` | lobby, host (the room checks) | `SetSeats` |
| `configureRoles` | `roles: {role: n}` | lobby, host | `ConfigureRoles` |
| `configureTimers` | `timers: {roleView, chiefElection, nightStep, phoneCheck, loversMeet, announce, day, nominationLock, defense, verdict, hunterShot, chiefSuccession}` in ms, only the ones to change | lobby, host | `ConfigureTimers`; out of range → `timer_out_of_range` |
| `configureSound` | `mode: "oneSpeaker"` (default) or `"allPhones"` | lobby, host | `ConfigureSound`; unknown mode → `bad_message` |
| `dealRoles` | – | lobby, host | `DealRoles` (seed from the room) |
| `voteChief` | `candidate` | Chief election | `VoteChief` |
| `nameSuccessor` | `successor` | `ChiefSuccession`, the dead Chief only (rules §7b) | `NameSuccessor` |
| `nominate` | `target` (`""` = no one) | day (rules §4) | `Nominate` |
| `skipDay` | `on: bool` | day (rules §4) | `SkipDay` |
| `verdict` | `kill: bool` | verdict (rules §4) | `CastVerdict` |
| `abortVote` | `on: bool` (true = vote to abort, false = withdraw) | `Day`, `NominationLock`, `Trial`, `Verdict`, `Dusk`; living players only (rules §8b) | `CastAbortVote`; when the vote passes (rules §8b) the game ends with `draw: true` and no `winner` |
| `newGame` | – | `Ended`, host only (rules §8c) | `NewGame`; back to the lobby with the same players, seats and role list; the host deals again with a fresh seed |
| `wolfVote` | `target` (required: `""` = "Không cắn ai"; a missing or `null` target → `bad_message`) | pack step (rules §5 "Sói") | `WolfVote`; `""` is a choice like any target (rules §5 "Sói"). The pack sees it in `view.pack.votes` as `""` for that wolf |
| `visitorVisit` | `target` (a living player, not herself, not last night's) | Visitor step (rules §6 "Nguyệt Nữ") | `VisitorVisit`; a missing choice is made by the server when the step ends |
| `guardProtect` | `target` (a living player, not last night's) | Guard step (rules §6 "Bảo vệ") | `GuardProtect` |
| `seerCheck` | `target` | Seer step (rules §6 "Tiên tri") | `SeerCheck` |
| `wolfSeerCheck` | `target` (any living player, wolves and himself too) | Wolf seer step (rules §5 "Sói tiên tri"); not blocked | `WolfSeerCheck`; a dead player → `invalid_target`; second check → `already_acted`; blocked → `blocked`. Answer: `WolfSeerChecked {wolfSeer, target, isSeer}` to the whole pack |
| `detectiveCheck` | `target` (any living player, himself too) | Detective step, once per game (rules §6 "Thám tử"); not blocked | `DetectiveCheck`; a dead player → `invalid_target`; second use → `ability_used`; blocked → `blocked`. Answer: `DetectiveChecked {detective, target, hasWolf}` to the Detective only (which seats count: rules §6 "Thám tử") |
| `cupidPair` | `a`, `b` (two different living players; Cupid himself allowed) | Cupid step, night 1, once (rules §7); not blocked | `CupidPair`; the same player twice or a dead one → `invalid_target`; second choice → `already_acted`. Answer: `LoversPaired {cupid, a, b}` to Cupid only. A missing choice is made by the server when the step ends. The lovers are told at the phone check (`me.phoneCheck`), then the `LoversMeet` step has no command (rules §3, §7) |
| `cursedWolfCurse` | `target` (a living non-wolf, not himself) | Cursed wolf step, nights 1–3, once per game (rules §5 "Sói nguyền"); not blocked | `CursedWolfCurse`; a wolf, himself or a dead player → `invalid_target`; second curse → `already_acted`; blocked → `blocked` |
| `hunterShoot` | `target` (a living player; required) | `HunterShot` phase, the dead Hunter only (rules §6 "Thợ săn") | `HunterShoot`; no target or a dead one → `invalid_target`; none before the deadline = nobody shot |
| `witchHeal` | – | Witch step, heal potion left, someone bitten (rules §6 "Phù thủy") | `WitchHeal`; not blocked |
| `witchPoison` | `target` (any living player, herself too) | Witch step, poison potion left (rules §6 "Phù thủy") | `WitchPoison`; not blocked |

The engine's own commands (setup, Chief, day, `newGame`, `hunterShoot`) are
listed in `transport/ws/commands.go`; every other row is declared by its role
in its own file (`Commands()`, see `docs/architecture.md` "Roles"). The server
builds the full table at start-up and does not start if two commands share a
name. A new role still adds its row here. The server **rejects** any `seed`
field and never accepts `PhaseTimeout` from a client: seeds and timeouts come
from the room.

## Server → client

### `view`: the screen now
Sent on (re)join and whenever this phone's screen changed. It replaces the
previous one. A phone whose screen did not change gets **nothing**: when one
player acts in secret at night, no other phone receives a message (a message
on every phone would tell that the role being called is alive, rules §3).
```json
{"type": "view", "deadline": 1760500000000, "speaker": true, "view": { … }}
```
`speaker` is added by the room, only `true` and only for the phone(s) that read
aloud. Which phones those are follows the sound mode the host sets with
`configureSound` before the deal (rules §1 "Máy làm loa", "Âm thanh"; default
`oneSpeaker`): one phone at a time in `oneSpeaker` mode, every phone in
`allPhones` mode.
`deadline` is added by the room: when the running phase or night step ends.
Every night step lasts its full time (rules §3), so it tells nothing about who
is alive.

The `view` object (`engine.View`); absent parts are omitted:

| field | who gets it | content |
|---|---|---|
| `phase`, `seq`, `night` | all | phase name, change counter, night number |
| `roles` | all | public role list `{role: count}` |
| `roleCards` | all | the same roles in the order of rules §9b, each `{role, name, team, description}` (`team` is the role's own team; Vietnamese text from the server; no card text in the frontend) |
| `players` | all | seats in order: `id`, `nickname`, `seat`, `alive`. Empty in the lobby until the host sends `setSeats` (who is in the room is known from the `PlayerJoined`/`PlayerLeft`/`PlayerKicked` events until then); `role`/`team` only when allowed (own seat, lover partner's role, full view); a dead Hunter's `role` (never `team`) is public; `card` `{name, team, goal, description}` exactly when `role` is set (dead players and the end screen: every seat) |
| `chief`, `succession`, `shooter` | all | Chief id; the dead Chief naming a successor; the dead Hunter choosing a target |
| `votes` | all | `chief {voter: candidate}`, `nominations {voter: target}`, `skips [id]`, `defendant`, `verdicts {voter: kill}`, `spared [id]`, `abortVotes {voter: yes}` |
| `winner`, `draw`, `winners` | all, at the end | winning team, or a draw; `winners`: the ids of the players who win, dead or alive, in seat order (who wins: rules §8) |
| `summary` | all, at the end | the game's own story, one line per night and day plus the final outcome (audit K4), e.g. `["Đêm 1: Sói cắn An, Phù thủy cứu An. → Không ai chết.", "Ngày 1: Không ai bị treo.", "Phe Sói chiến thắng."]`; built server-side from facts that stayed secret during play (Vietnamese text from `i18n`, ready for another language later), so the frontend never has to reconstruct it from raw events |
| `nightCall` | all phones | `{role, duration}`, nothing else: a fake call looks the same |
| `me` | players | `id`, `role`, `team`, `alive`, `cursed` (known only from the phone check on), `canShoot` (only the dead Hunter while he chooses), `phoneCheck`, `turn`, `roleCard`, `statusCards` |
| `me.roleCard` | players, once roles are dealt | `{name, team, goal, description}` of the player's own role: `name`/`description` of the role, `team`/`goal` of the player's current side as their screen knows it (which side: rules §9b; it changes at the phone check). A revealed Hunter's `card` seen by the living is the printed card (role's own side). Every card has these four keys, all set, whatever the role. Show it as rules §9b says (fixed size, hidden until pressed) |
| `me.statusCards` | players | extra cards held now, each `{kind, name, text}`: `cursed` (from the phone check on), `lovers` or `lovers_cross` (from the phone check on; `lovers_cross` stays after the partner died), `chief` (while holding the title) |
| `me.phoneCheck` | every living player, only while the `PhoneCheck` step runs | `{cursed: bool, love: {inLove: bool, partner, partnerRole, colour}}`; one notice combining both. When the step runs: rules §3 "Kiểm tra điện thoại". Show the curse line only if the (public) role list has a Cursed wolf and the love line only if it has Cupid. Text: `cursed` false → "Bạn không bị nguyền", true → "Bạn đã bị nguyền"; `inLove` false → "Bạn không ở trong cặp đôi", true → "Bạn yêu {partner}, vai {partnerRole}". Every notice has exactly the same JSON keys whatever the answer: this message is the one exception to "empty values are omitted" (`inLove: false` comes with empty `partner`, `partnerRole`, `colour`), so the screen never tells who was cursed or in love. |
| `me.turn` | player being called | `step` (`LoversMeet` on night 1 for the two lovers only: they open their eyes and look at each other, no action); `blocked: true` → show "Bạn không thể sử dụng chức năng đêm nay" (never says who blocked) and no action; `doomed: true` → this player is dead by the rules from this step (rules §6 "Hiệp sĩ kiếm gỉ", §7): no action, no vote, they are shown dead only in the morning; `grief: true` (own phone only) → say it is grief: "Người yêu của bạn đã chết, bạn đau buồn mà chết từ lúc này"; `bitten`, `biteShown` (Witch with heal potion: `biteShown` true and `bitten` empty = nobody bitten; no `biteShown` = she is told nothing) |
| `pack` | living wolves (cursed too), full views | `members [id]`, `votes {wolf: target}` live |
| `lover` | the two lovers, from the phone check on (never before; Cupid never) | `partner`, `partnerRole`, `colour`: `white` / `black` / `striped` (rules §7 table) |
| `full`, `secrets` | dead players, everybody after the game | every role and team, tonight's choices, the lovers pair |

### `events`: what this screen may know, in order
```json
{"type": "events", "reset": false, "from": 41, "events": [
  {"type": "DawnAnnounced", "data": {"kind": "deaths", "dead": ["p5", "p8"]},
   "say": "Đêm qua, … đã chết."}
]}
```
- On (re)join, and whenever the visible list is rebuilt (the player dies or the
  game ends: they now see the whole history), `reset: true` and `events` is
  the full list. Otherwise only the new ones, `from` = index of the first.
- **`reset: true` replaces the whole event list; a client must never append it.**
  The new list is longer than the old one: it has the events that were hidden
  from the player until now put back in their places, so indexes shift. A client
  that appends a reset message, or keeps counting "events already shown" by
  index, shows old events again (the repeated lines seen in `make demo-12`
  came from exactly this). Replace the list, then show only what is new: find
  the events already shown inside the new list, in order, and continue after
  them (`cmd/demo/printer.go`, `afterSubsequence`). A message without reset
  always has `from` equal to the length of the list the client holds.
- `say` is present on announcements (events with an i18n key): the speaker
  phone reads it aloud; every phone shows it as text. Since audit [L3]
  (docs/voice-lines.md) this also covers `RoleViewStarted` (right after the
  deal), `ChiefElectionStarted`, `NightStarted`, `DawnStarted` (right after
  `NightEnded`, before the morning announcement), `DayStarted`,
  `TrialStarted` (names the defendant), `VerdictStarted` and `HunterTimedOut`
  (names the Hunter). `NightStepCalled` gets a `say` too, but it names the
  role being called, never a player (e.g. "Bầy sói hãy mở mắt..."); the two
  night steps that belong to no role (the phone check, the lovers looking at
  each other) read their own line instead (`PhoneCheckAnnounced`,
  `LoversMeetAnnounced`) and get none on `NightStepCalled`, so nothing is read
  twice.
- Event types a living player can get: public ones (joins, seating, role
  list, Chief election and votes, `NightStarted`, `NightEnded`,
  `DawnAnnounced`, day votes and trial, `Spared`, `Hanged`, `HunterRevealed`,
  `HunterShotStarted` (`hunter`, `duration` only), `HunterShot`, `HunterTimedOut`, `GameEnded` (its `say` tells a normal win/draw from an abort, see `votes.abortVotes`), `AbortVoteCast`, `NewGameStarted`, `PhoneCheckAnnounced` (has `say`: "Mọi người mở mắt xem điện thoại, rồi nhắm mắt lại."), `LoversMeetAnnounced` (has `say`: "Cặp đôi hãy mở mắt nhìn nhau.")); own
  results (`SeerChecked` for the Seer, `PlayerProtected` for the Guard,
  `PlayerCursed` for the Cursed wolf, `PhoneChecked` for each living player: their own notice, `LoversPaired` for Cupid); `WolfVoted` and `WolfSeerChecked` for the pack, from the moment
  the player is in it.
- Never sent to the living: `PlayerDied` and `LoverGrieving` (its order would tell bitten from
  poisoned; deaths are known through `DawnAnnounced`, `Hanged` and the alive
  flags), `HunterChainNoted` (the chain of deaths behind a Hunter's shot, in the
  order the night killed them), `RolesDealt`, (`NightStepCalled` is public: role and duration, no real/fake flag).

### `error`
```json
{"type": "error", "code": "not_your_turn", "message": "…"}
```
`code` is the snake_case name of the engine error (`ErrNotYourTurn` →
`not_your_turn`; one exception, `room.ErrClosed` → `room_closed`). Each error
carries its code where it is declared (`engine.NewError`), so a role's own
error needs no table elsewhere. A rejected command changes nothing. Room errors use the same
scheme (`not_host`, `forbidden`, `disconnected`, `bad_credentials`,
`pin_locked` with `retryAfter`, `room_not_found`), and so do the transport's own:
`bad_message`, `unknown_type`, `unknown_command`, `message_too_large`,
`rate_limited`, `not_joined`, `already_joined`, `too_many_connections`. An
error the server does not know how to name is `internal` with no detail; so
is a command that hit a bug in the server (it changed nothing, the room goes
on). A connection closed for failing to join/rejoin too many times, or for
not answering a ping, gets no error frame first — it is already not behaving
like a phone that is still there.

A player who does not hold the role of a night command (the pack's vote: who
is not on the wolf team) always gets exactly
`{"code": "not_your_turn", "message": "not your turn"}`, whatever their own
state (blocked, dead by the rules, dead, any other role), so an error never
tells them they were visited. The holder gets the detail (`blocked`,
`not_alive`, `not your turn: Seer is not being called`...).

`DawnAnnounced.kind` is `deaths`, `protected` or `nobodyDied`; phases and
teams are sent by name (`Night`, `Village`).

## What each screen gets (summary)

- **Every phone**: the public table and votes, the current night call (role +
  duration), announcements with their Vietnamese text. The speaker phone reads
  them aloud.
- **Living player**: plus own role and team, whether it
  is their turn, own night results.
- **Wolves**: plus the pack and its live votes, the Wolf seer's results.
- **Witch**: plus the bitten player on her turn while she can heal, unless blocked.
- **Lovers**: plus the partner, the partner's role and the link colour, from the phone check on. The pack is not told that one of its wolves is in love.
- **Dead players, and everybody once the game is over**: everything.
