# Protocol (T9 draft, implemented in T11)

JSON over one WebSocket per screen. Implemented by `internal/transport/ws` (T11).
Every message is an object with a `type` field. Field names are camelCase,
durations are milliseconds, times are Unix milliseconds.

## Rooms and the socket (HTTP)

| request | reply |
|---|---|
| `POST /rooms` (body optional: `{"mode": "inPerson"}`) | `201 {"room": "K7QP"}`; `mode: "online"` → `400` with `mode_unsupported`; `503` `too_many_rooms` |
| `GET /join/{code}` | `200 {"room": "K7QP"}`, or `404` `room_not_found`. This is the link the QR code will carry (the UI is built later); case is ignored |
| `GET /ws` | upgrades to the game socket |

- The room code is 4 characters from `ABCDEFGHJKMNPQRSTUVWXYZ23456789` (no O, 0, I,
  1, L). It is not a secret.
- A room nobody has joined, left, dropped from or sent a command to for **2 hours**
  is deleted, and its sockets are closed.
- `/ws` takes **no query string** (a request with one gets `400` `secret_in_url`):
  the secret code and the PIN only ever travel inside messages, because URLs end
  up in logs. The room code goes in the first message, `join` or `rejoin`, which
  must arrive within 10 s.
- **Origin.** The `Origin` header must be one of the configured origins
  (`ALLOWED_ORIGINS`, comma separated, `scheme://host[:port]`). No `Origin` is
  refused too, unless `ALLOW_ANY_ORIGIN=1` (demo and tests), which allows
  everything. Refused → `403` `bad_origin`.
- **Limits per connection.** A message over 4 KB gets `error` `message_too_large`
  (over 64 KB the connection is closed). More than 10 messages per second
  (a burst of 20 is allowed) gets `error` `rate_limited` and the message is
  ignored; staying over the limit without a break for 3 s closes the connection.
- **Strict decoding.** An unknown `type` is `unknown_type`, an unknown command
  `unknown_command`. Unknown fields anywhere (a `seed`, a `by`, a `voter`),
  wrong field types, wrong capitals and trailing data are `bad_message`.
- A dropped connection never removes the player; reconnect with `rejoin`. A
  rejoin from another phone closes the old socket (its next command gets
  `disconnected`).
- `welcome` always comes first, before the first `view`.

The secrecy rule: the server never sends `engine.State`. Every outgoing game
message is built from `engine.ViewFor` (the screen now) and `engine.Visible`
(the events this screen may know about). See `internal/engine/view.go` for who
sees what.

## Identity

- A `PlayerID` is **public**: it appears in every view (seats, votes). It is
  not a credential.
- On join the server returns a secret `token` (the phone stores it) and a
  4-digit `pin` (shown only on that player's screen, with a reminder to
  remember it). Rejoin (rules §2): same phone with the token, other phone with
  nickname + PIN. 5 wrong PINs lock PIN rejoin for that nickname for 1 minute.
  A nickname alone never rejoins.
- The connection decides who acts. Commands never carry `by`/`voter`; the
  server fills it from the connection.

## Client → server

### Joining
```json
{"type": "join",     "room": "K7QP", "nickname": "An"}
{"type": "rejoin",   "room": "K7QP", "token": "…"}
{"type": "rejoin",   "room": "K7QP", "nickname": "An", "pin": "4821"}
{"type": "leave"}
```
Reply: `{"type": "welcome", "player": "p3", "token": "…", "pin": "4821"}` (on
rejoin by token only `player`; rejoin by nickname + PIN returns a **new** `token`
and the old one stops working), then a full `view` and
a full `events` message. A refused rejoin is an `error` with code
`bad_credentials` or `pin_locked` (with `retryAfter` in ms).

`leave` (lobby only) removes the player for good; a closed connection keeps their
place. Creating a room in `Online` mode is refused with `mode_unsupported`.

### Commands
```json
{"type": "command", "name": "nominate", "args": {"target": "p5"}}
```

| name | args | when | engine command |
|---|---|---|---|
| `setSeats` | `order: [id]` | lobby, host (the room checks) | `SetSeats` |
| `configureRoles` | `roles: {role: n}` | lobby, host | `ConfigureRoles` |
| `configureTimers` | `timers: {roleView, chiefElection, nightStep, phoneCheck, loversMeet, announce, day, nominationLock, defense, verdict, hunterShot, chiefSuccession}` in ms, only the ones to change | lobby, host | `ConfigureTimers`; out of range → `timer_out_of_range` |
| `dealRoles` | – | lobby, host | `DealRoles` (seed from the room) |
| `voteChief` | `candidate` | Chief election | `VoteChief` |
| `nameSuccessor` | `successor` | dead Chief, 15 s | `NameSuccessor` |
| `nominate` | `target` (`""` = no one) | day | `Nominate` |
| `skipDay` | `on: bool` | day | `SkipDay` |
| `verdict` | `kill: bool` | verdict | `CastVerdict` |
| `wolfVote` | `target` (required: `""` = "Không cắn ai", no bite; a missing or `null` target → `bad_message`) | pack step | `WolfVote`; `""` counts like a target: more than half of the voters for it and nobody is bitten, announced as the usual "Không có ai chết". The pack sees it in `view.pack.votes` as `""` for that wolf |
| `visitorVisit` | `target` (not herself, not last night's) | Visitor step; a missing choice is made at random by the server when the step ends | `VisitorVisit` |
| `guardProtect` | `target` | Guard step | `GuardProtect` |
| `seerCheck` | `target` | Seer step | `SeerCheck` |
| `wolfSeerCheck` | `target` (any living player, wolves and himself too) | Wolf seer step, every night; not blocked | `WolfSeerCheck`; a dead player → `invalid_target`; second check → `already_acted`; blocked → `blocked`. Answer: `WolfSeerChecked {wolfSeer, target, isSeer}` to the whole pack (a role check: a cursed Seer is still `isSeer: true`) |
| `detectiveCheck` | `target` (any living player, himself too) | Detective step, every night (after use the call is fake); not blocked | `DetectiveCheck`; a dead player → `invalid_target`; second use → `ability_used`; blocked → `blocked`. Answer: `DetectiveChecked {detective, target, hasWolf}` to the Detective only: whether the target or the nearest living neighbour on either side is a wolf (a cursed player counts); no count, no names |
| `cupidPair` | `a`, `b` (two different living players; Cupid himself allowed) | Cupid step, night 1 only (order 0), once; not blocked | `CupidPair`; the same player twice or a dead one → `invalid_target`; second choice → `already_acted`. Answer: `LoversPaired {cupid, a, b}` to Cupid only. A missing choice is made at random by the server when the step ends (two random living players). Nobody else is told: the lovers learn at the phone check right after Cupid's step, then the `LoversMeet` step (no command) lets them look at each other |
| `cursedWolfCurse` | `target` (a living non-wolf, not himself) | Cursed wolf step, nights 1–3, once per game; not blocked | `CursedWolfCurse`; a wolf, himself or a dead player → `invalid_target`; second curse → `already_acted`; blocked → `blocked` |
| `hunterShoot` | `target` (a living player; required) | `HunterShot` phase, the dead Hunter only | `HunterShoot`; no target or a dead one → `invalid_target`; 15 s without it = nobody shot |
| `witchHeal` | – | Witch step, heal potion left, someone bitten | `WitchHeal`; not blocked |
| `witchPoison` | `target` (any living player, herself too) | Witch step, poison potion left | `WitchPoison`; not blocked |

Later roles add their own rows. The server **rejects** any `seed` field and never
accepts `PhaseTimeout` from a client: seeds and timeouts come from the room.

## Server → client

### `view`: the screen now
Sent after every change, and on (re)join. It replaces the previous one.
```json
{"type": "view", "deadline": 1760500000000, "speaker": true, "view": { … }}
```
`speaker` is added by the room, only `true` and only for the phone that reads aloud
(the host, or the next connected player in seat order if the host is gone).
`deadline` is added by the room: when the running phase or night step ends.
Every night step lasts the same time, so it tells nothing about who is alive.

The `view` object (`engine.View`); absent parts are omitted:

| field | who gets it | content |
|---|---|---|
| `phase`, `seq`, `night` | all | phase name, change counter, night number |
| `roles` | all | public role list `{role: count}` |
| `roleCards` | all | the same roles in the order of rules §9b, each `{role, name, description}` (Vietnamese text from the server; no card text in the frontend) |
| `players` | all | seats in order: `id`, `nickname`, `seat`, `alive`; `role`/`team` only when allowed (own seat, lover partner's role, full view); a dead Hunter's `role` (never `team`) is public; `card` `{name, team, goal, description}` exactly when `role` is set (dead players and the end screen: every seat) |
| `chief`, `succession`, `shooter` | all | Chief id; the dead Chief naming a successor; the dead Hunter choosing a target |
| `votes` | all | `chief {voter: candidate}`, `nominations {voter: target}`, `skips [id]`, `defendant`, `verdicts {voter: kill}`, `spared [id]` |
| `winner`, `draw`, `winners` | all, at the end | winning team, or a draw; `winners`: the ids of the players who win, dead or alive, in seat order (the lovers of a cross-team couple, alive or broken, are never in it unless the couple wins) |
| `nightCall` | all phones | `{role, duration}`, nothing else: a fake call looks the same |
| `me` | players | `id`, `role`, `team`, `alive`, `cursed` (known only from the phone check on), `canShoot` (only the dead Hunter while he chooses), `phoneCheck`, `turn`, `roleCard`, `statusCards` |
| `me.roleCard` | players, once roles are dealt | `{name, team, goal, description}` of the player's own role (rules §9b). Every card has these four keys, all set, whatever the role. Show it at a fixed size, hidden by default, revealed only while pressed |
| `me.statusCards` | players | extra cards held now, each `{kind, name, text}`: `cursed` (from the phone check on), `lovers` or `lovers_cross` (from the phone check on; `lovers_cross` stays after the partner died), `chief` (while holding the title) |
| `me.phoneCheck` | every living player, only while the `PhoneCheck` step runs | `{cursed: bool, love: {inLove: bool, partner, partnerRole, colour}}`; one notice combining both. The step runs right after Cupid's step on night 1 when the role list has Cupid, and right after the Cursed wolf's step on nights 1–3 when it has the Cursed wolf; with both, night 1 has two checks (the love part is repeated at the curse checks, so a lover sees the new link colour after a curse). Show the curse line only if the (public) role list has a Cursed wolf and the love line only if it has Cupid. Text: `cursed` false → "Bạn không bị nguyền", true → "Bạn đã bị nguyền"; `inLove` false → "Bạn không ở trong cặp đôi", true → "Bạn yêu {partner}, vai {partnerRole}". Every notice has exactly the same JSON keys whatever the answer: this message is the one exception to "empty values are omitted" (`inLove: false` comes with empty `partner`, `partnerRole`, `colour`), so the screen never tells who was cursed or in love. |
| `me.turn` | player being called | `step` (`LoversMeet` on night 1 for the two lovers only: they open their eyes and look at each other, no action); `blocked: true` → show "Bạn không thể sử dụng chức năng đêm nay" (never says who blocked) and no action; `doomed: true` → this player is dead by the rules from this step (wolf poisoned by the Rusty knight; lover whose partner died by day): no action, no vote, they are shown dead only in the morning; `grief: true` (own phone only) → say it is grief: "Người yêu của bạn đã chết, bạn đau buồn mà chết từ lúc này"; `bitten`, `biteShown` (Witch with heal potion: `biteShown` true and `bitten` empty = nobody bitten; no `biteShown` = she is told nothing) |
| `pack` | living wolves (cursed too), full views | `members [id]`, `votes {wolf: target}` live |
| `lover` | the two lovers, from the phone check on (never before; Cupid never) | `partner`, `partnerRole`, `colour`: `white` / `black` / `striped` |
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
  phone reads it aloud; every phone shows it as text.
- Event types a living player can get: public ones (joins, seating, role
  list, Chief election and votes, `NightStarted`, `NightEnded`,
  `DawnAnnounced`, day votes and trial, `Spared`, `Hanged`, `HunterRevealed`,
  `HunterShotStarted`, `HunterShot`, `HunterTimedOut`, `GameEnded`, `PhoneCheckAnnounced` (has `say`: "Mọi người mở mắt xem điện thoại, rồi nhắm mắt lại."), `LoversMeetAnnounced` (has `say`: "Cặp đôi hãy mở mắt nhìn nhau.")); own
  results (`SeerChecked` for the Seer, `PlayerProtected` for the Guard,
  `PlayerCursed` for the Cursed wolf, `PhoneChecked` for each living player: their own notice, `LoversPaired` for Cupid); `WolfVoted` and `WolfSeerChecked` for the pack, from the moment
  the player is in it.
- Never sent to the living: `PlayerDied` and `LoverGrieving` (its order would tell bitten from
  poisoned; deaths are known through `DawnAnnounced`, `Hanged` and the alive
  flags), `RolesDealt`, (`NightStepCalled` is public: role and duration, no real/fake flag).

### `error`
```json
{"type": "error", "code": "not_your_turn", "message": "…"}
```
`code` is the snake_case name of the engine error (`ErrNotYourTurn` →
`not_your_turn`). A rejected command changes nothing. Room errors use the same
scheme (`not_host`, `forbidden`, `disconnected`, `bad_credentials`,
`pin_locked` with `retryAfter`, `room_not_found`), and so do the transport's own:
`bad_message`, `unknown_type`, `unknown_command`, `message_too_large`,
`rate_limited`, `not_joined`, `already_joined`. An error the server does not
know how to name is `internal` with no detail.

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
