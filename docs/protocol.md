# Protocol (draft for review, T9)

JSON over one WebSocket per screen. Implemented by `internal/transport/ws` (T11).
Every message is an object with a `type` field. Field names are camelCase,
durations are milliseconds, times are Unix milliseconds.

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
| `configureTimers` | `timers: {roleView, chiefElection, nightStep, announce, day, nominationLock, defense, verdict, hunterShot, chiefSuccession}` in ms, only the ones to change | lobby, host | `ConfigureTimers`; out of range → `timer_out_of_range` |
| `dealRoles` | – | lobby, host | `DealRoles` (seed from the room) |
| `voteChief` | `candidate` | Chief election | `VoteChief` |
| `nameSuccessor` | `successor` | dead Chief, 15 s | `NameSuccessor` |
| `nominate` | `target` (`""` = no one) | day | `Nominate` |
| `skipDay` | `on: bool` | day | `SkipDay` |
| `verdict` | `kill: bool` | verdict | `CastVerdict` |
| `wolfVote` | `target` | pack step | `WolfVote` |
| `guardProtect` | `target` | Guard step | `GuardProtect` |
| `seerCheck` | `target` | Seer step | `SeerCheck` |

Later roles add their own rows (Witch, Visitor, Cursed wolf, Wolf seer,
Detective, Cupid, Hunter). The server **rejects** any `seed` field and never
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
| `players` | all | seats in order: `id`, `nickname`, `seat`, `alive`; `role`/`team` only when allowed (own seat, lover partner's role, full view) |
| `chief`, `succession` | all | Chief id; the dead Chief naming a successor |
| `votes` | all | `chief {voter: candidate}`, `nominations {voter: target}`, `skips [id]`, `defendant`, `verdicts {voter: kill}`, `spared [id]` |
| `winner`, `draw` | all, at the end | winning team, or a draw |
| `nightCall` | all phones | `{role, duration}`, nothing else: a fake call looks the same |
| `me` | players | `id`, `role`, `team`, `alive`, `cursed`, `turn` |
| `me.turn` | player being called | `step`; `visited: true` → show "Được Nguyệt Nữ ghé chơi, không thể sử dụng chức năng" and no action; `bitten` (Witch with heal potion) |
| `pack` | living wolves (cursed too), full views | `members [id]`, `votes {wolf: target}` live |
| `lover` | the two lovers | `partner`, `partnerRole`, `colour`: `white` / `black` / `striped` |
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
- `say` is present on announcements (events with an i18n key): the speaker
  phone reads it aloud; every phone shows it as text.
- Event types a living player can get: public ones (joins, seating, role
  list, Chief election and votes, `NightStarted`, `NightEnded`,
  `DawnAnnounced`, day votes and trial, `Spared`, `Hanged`, `GameEnded`); own
  results (`SeerChecked` for the Seer, `PlayerProtected` for the Guard,
  `TeamChanged` for a cursed player); `WolfVoted` for the pack, from the moment
  the player is in it.
- Never sent to the living: `PlayerDied` (its order would tell bitten from
  poisoned; deaths are known through `DawnAnnounced`, `Hanged` and the alive
  flags), `RolesDealt`, (`NightStepCalled` is public: role and duration, no real/fake flag).

### `error`
```json
{"type": "error", "code": "not_your_turn", "message": "…"}
```
`code` is the snake_case name of the engine error (`ErrNotYourTurn` →
`not_your_turn`). A rejected command changes nothing.

## What each screen gets (summary)

- **Every phone**: the public table and votes, the current night call (role +
  duration), announcements with their Vietnamese text. The speaker phone reads
  them aloud.
- **Living player**: plus own role and team, whether it
  is their turn, own night results.
- **Wolves**: plus the pack and its live votes, the Wolf seer's results.
- **Witch**: plus the bitten player on her turn while she can heal, unless visited.
- **Lovers**: plus the partner, the partner's role and the link colour.
- **Dead players, and everybody once the game is over**: everything.
