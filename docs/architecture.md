# Architecture

## Goals
1. **Rules are testable without a server.** All game logic is a pure library.
2. **Adding a role = adding one file.** No edits to existing roles.
3. **No secret leaks.** Clients only ever receive a per-player view.
4. **Swappable edges.** WebSocket, storage and clock can be replaced without touching rules (offline → online later).

## Package layout

```
cmd/server/            main: wires config, store, room manager, transport
internal/engine/       pure game core: state, commands, events, phases, night pipeline, win check, views
internal/engine/enginetest/  test helpers and scenario builder
internal/roles/        one file per role, each registers into engine.Registry
internal/i18n/         Vietnamese announcement strings
internal/room/         room lifecycle: join/rejoin, timers, command queue per room, broadcasting views
internal/store/        EventStore interface + in-memory implementation
internal/transport/ws/ WebSocket handler, message encoding (see docs/protocol.md)
```

Dependency direction (arrows = "imports"):

```
cmd/server ─► transport/ws ─► room ─► engine ◄─ roles
                               │        ▲
                               └► store ┘ (store only knows engine.Event)
```

`engine` imports nothing from this repo except `i18n`.

## Core model (engine)

Event-sourced: the game is a list of events. State is rebuilt by folding events. This gives replay, rejoin and the "dead players see full history" view for free.

```go
type Command interface{ isCommand() }   // a player's intent, e.g. GuardProtect{By, Target}
type Event   interface{ isEvent() }     // a fact, e.g. PlayerProtected{Guard, Target}

// Decide validates a command against the current state. No side effects.
func Decide(s State, c Command) ([]Event, error)

// Evolve applies one event. Must never fail.
func Evolve(s State, e Event) State
```

Timers are not inside the engine: `room` sends `PhaseTimeout{Phase, Seq}` when a deadline passes. It is the only command that moves a timed phase or a night step forward, and a night step never ends early, so every call lasts its full time. A timeout whose `Phase` or `Seq` no longer matches the state is stale: the engine ignores it (no events, no error).

Host powers end with `DealRoles`. After the deal the engine has no host-only command; the host is a normal player. Lobby commands then fail with `ErrWrongPhase` and a second deal with `ErrAlreadyDealt`.

Randomness (role dealing) is a command input: `DealRoles{Seed int64}`, dealt by a deterministic shuffle so tests are reproducible.

**Seeds never come from a client.** The room layer (T10) generates every seed with `crypto/rand` and puts it in the command. The transport must not accept a seed field from clients, or a player could pick their own role. The same goes for the Chief election tie-break.

### State (sketch)
```go
type State struct {
    Phase      Phase          // Lobby, RoleView, Night, Dawn, Day, NominationLock, Trial, Verdict, Dusk, HunterShot, Ended
    Seq        int            // increments on every phase or night step change
    Night      int            // 1-based night counter
    NightStep  int            // night steps called so far; current step = steps[NightStep-1]
    Seats      []PlayerID     // seating circle, clockwise; "left" = index-1 (wraps)
    Players    map[PlayerID]*Player
    Pending    NightActions   // choices collected this night, resolved at dawn
    Day        DayState       // nominations (ordered), skip votes, trial, verdict votes, clock
    Lovers     *LoversLink
    Settings   Settings       // timers, role list
    Winner     *Team
}

type Player struct {
    ID       PlayerID
    Nickname string
    Role     RoleID
    Team     Team           // can change (curse)
    Alive    bool
    Flags    map[string]int // role-owned counters: potions left, curse used, last guarded, poisoned-until...
}
```

`Flags` lets roles keep their own state without changing `Player`.

`Command` and `Event` are sealed by an unexported marker method, so roles embed
`engine.CommandMarker` / `engine.EventMarker` in their own commands and events.

### Roles
```go
type Role interface {
    ID() RoleID
    BaseTeam() Team
    NightCall() *NightCall          // nil = not called at night. Includes order and which nights (e.g. nights 1–3)
    Decide(s State, c Command) (evs []Event, handled bool, err error) // handles its own commands
    Evolve(s *State, e Event) bool                    // applies its own events; bool=handled
}

var Registry = map[RoleID]Role{}  // roles call engine.Register(...) in init()

// Optional interfaces, checked by type assertion. They keep the rules §2 setup
// limits in each role's own file instead of hardcoding role IDs in the engine.
type MinPlayersRule interface{ MinPlayers() int }           // Detective 9, Cupid 10
type MultiCopyRole  interface{ AllowsMultipleCopies() bool } // only Villager and Werewolf
type CurseSource    interface{ CursesVillager() bool }      // reserves one wolf slot
type LoveSource     interface{ CreatesLovers() bool }       // Cupid: the phone check also runs on the nights it is called
type AlwaysCalled   interface{ AlwaysCalled() bool }        // called every night even if not in the role list: the Werewolf pack step
```

Game-wide rule limits (player count, wolf percentage, the cursed wolf's reserved slot) live in `internal/engine/limits.go`. Limits that belong to one role (minimum table size, number of copies) live in that role's own file.

The night call order (rules §3) comes from each role's `NightCall().Order`. The role list of a game is public (rules §2), so a night calls exactly the roles in `Settings.Roles` that have a `NightCall` for that night; a role nobody holds is never called. The one exception is the wolf pack step (`Werewolf`, implements `AlwaysCalled`): every game has a wolf, and Werewolf, CursedWolf and WolfSeer all bite in it, so it runs every night even when no plain Werewolf is in the list. With all three wolf roles the order is CursedWolf → WolfSeer → pack. A role in the list whose holder is dead is still called (a fake call). The engine itself produces the calls as `NightStepCalled{Role, Duration}` events, public to every phone (role and duration only). The step list depends on the role configuration and the night number, never on who is alive, and the event carries no real/fake flag, so a real call and a fake one are identical and last the same time. `engine.CurrentNightRole(s)` tells a role whether it is its turn.

### Night resolution pipeline (rules §9)
An ordered slice of steps, each a pure function `func(*Resolution)`:

```
1 VisitorBlock → 2 Curse (nothing left to do) → 3 Bite(guard, witch heal) → 4 RustyKnight
→ 5 WitchPoison → 6 DelayedDeaths → 7 GriefWithoutCall + LoverDeaths → 8 Announce → 9 WinCheck
```

Roles hook into named steps via `Resolution` fields instead of editing the pipeline. The hunter shot is a separate phase after the announcement (it needs player input). `afterDeaths` (`after_death.go`) is the one place for what follows any death: each dead Hunter (role implements `DeathShooter`) is revealed (`HunterRevealed`, public, sets `FlagRevealed`) and gets `PhaseHunterShot`; `HunterShoot` or the timeout (`hunter_shot.go`) kills the target and re-enters `settle`, which repeats for a shot Hunter, then runs the win check (step 9) and the Chief succession. `State.Shot` remembers the queue and the phase to return to. By day, `settle` marks the surviving partner of each dead lover (`LoverGrieving`, `FlagGrieving`) before the win check; they die in the next night (below).

**Win check only by day** (rules §8): `settle` is the only caller of `CheckWin`, and it runs after the morning announcement (even when nobody died) and after a hanging, always on the situation of that moment. A team change (the curse) never ends the game by itself; the night keeps running all its steps.

**Phone check** (rules §3): `engine.StepPhoneCheck` is an engine step, not a role. `engineSlots` (used by `nightSteps`) adds it right after the step of any role implementing `CurseSource` on the nights that role is called, and right after the step of a `LoveSource` (Cupid, night 1), followed there by `engine.StepLoversMeet` (the lovers look at each other; `Settings.LoversMeet`, 10 s; voice line `LoversMeetAnnounced`; only the lovers' `TurnView` shows it). With both roles night 1 has two phone checks. The schedule depends on the role list and the night only, so it runs whether or not the curse was used, the Cursed wolf is dead or blocked. Its start emits `PhoneCheckAnnounced` (voice line) and one `PhoneChecked{Player, Cursed, Love}` per living player; all notices have the same shape. `PhoneChecked` also sets `FlagLoveKnown` on the lovers: until then `loverView` shows nothing. A team changed during the night carries `FlagTeamHidden` until its player's `PhoneChecked`; every view uses `VisibleTeam`, so neither the cursed player nor the pack (`packView`, `Audience.Wolves`, lover link colour) shows the change earlier, while `State` and the Seer/win logic use the real team at once. The duration of a step comes from `NightStepDuration` (the phone check has its own `Settings.PhoneCheck`, the lovers' step `Settings.LoversMeet`).

**Winners** (rules §8): `CheckWin` returns `Outcome.Winners`, the players who win in seat order: both lovers for the third team; otherwise every player of the winning team, dead or alive, except the two lovers of a cross-team couple, alive or broken: `PlayerDied` of the first lover sets `LoversLink.ThirdTeam` when the couple was cross-team at that moment, and they stay on the third team for good. It is copied to `State.Winners` and `View.Winners` at the end.

### Views (secrecy boundary)
```go
func ViewFor(s State, viewer Viewer) View              // the screen now; Viewer = PlayerViewer(id) (there is no narrator device; every phone is a player)
func Visible(s State, log []Event, viewer Viewer) []Event // the events this screen may know, judged at the time of each event
```
The rules are summed up in `view.go`; the projection is in `view_project.go`, the event audiences in `view_audience.go`, the role cards (rules §9b) in `view_cards.go`: `View.RoleCards` (public role list with name and description), `PlayerView.Card` (set exactly when `Role` is), `MyView.RoleCard` and `MyView.StatusCards`. Card text is the one place a view carries Vietnamese: it comes from `i18n` (`Card`, `StatusCard`), so the frontend holds no card text. Tests per role (`internal/roles/view_*_test.go`) prove what each screen can and cannot see. Dead players, and everybody once the game is over, get the full view. Every view carries the public part (table, votes, announcements, the current night call); the speaker phone reads it aloud, the others only show text.

- Engine events get their audience in `view_audience.go`. A role event declares its own by implementing `Audienced` (`Audience(s) Audience`: everyone, the living wolves, or listed players). An event without an audience is seen only by full views, so a forgotten audience hides information instead of leaking it.
- `PlayerDied` is never shown to the living: the order of a night's deaths would tell who was bitten and who was poisoned.
- Role-specific secrets derived from state go through optional role interfaces, e.g. `BiteWitness` (the Witch sees the bite on her turn while she can heal, not when visited). `Pending.Blocked` (target → blocking role, engine-internal, stripped from every view) drives the generic "cannot use your ability tonight" notice (`TurnView.Blocked`, never names the blocker) and is enforced by `NightActor` (`ErrBlocked`) and `WolfVoters`, so a future blocking role only records into it. A role that must act when its night step runs out (the Visitor's forced random visit) implements `NightTimeoutRole`; the seed comes from the room's `PhaseTimeout.Seed`; `TeamChanged` (the curse) tells only the player concerned.
- **Rusty knight (rules §6)**: a role with a passive ability implements `BiteAvenger`; pipeline step 4 (`stepKnight`) asks the bitten player's role which wolf to poison (the Knight uses `engine.NearestLivingWolf`, direction +1 = the left, the next seat). `resolveNight` records it as `WolfPoisoned{Wolf, Night}`, which sets `FlagWolfPoisoned` on that wolf and is visible to nobody alive. From the next night (`engine.Doomed`) the wolf is out of `WolfVoters` and step 6 (`stepDelayedDeaths`) adds it to the night's deaths, so it is announced with the others and Hunter/Chief consequences run through `afterDeaths`. The wolf stays `Alive` in state and in every view until then, so nothing leaks during the night. Its own phone gets `TurnView.Doomed` and the pack gets `PackView.Doomed` once the pack step has started.
- Message shapes: `docs/protocol.md`.

## Room layer
- One goroutine per room owns the state; commands arrive through a channel (no locks on game state).
- **Host** = the player who created the room (and plays). Only the host may send `SetSeats`, `ConfigureRoles`, `DealRoles`; the room checks this, the engine does not know hosts. If the host leaves before the deal, the next joined player becomes host. After the deal there is no host power.
- **Speaker policy**: who speaks sits behind `room.SpeakerPolicy`; v1 has only `InPersonSpeaker`. The host's phone reads announcements aloud; other phones stay silent. If the speaker disconnects, the next connected player in seat order (join order before seats are set) takes over. The room stores who the speaker is and sets `speaker: true` in that player's `view` message (an envelope field, not part of `engine.View`).
- **Identity**: the room issues each player a secret token and a 4-digit PIN (both from `crypto/rand`) behind an `Identity` interface. The token (or the PIN) decides who a connection is; commands never carry a player id from the client, the room fills `By`/`Voter`. Rejoin by token (same phone), or nickname + PIN (other phone, which gets a **new token**; the old one stops working). 5 wrong PINs lock PIN rejoin for that nickname for 1 minute (injected clock); token rejoin is never locked. A nickname alone never rejoins.
- **Seeds** (`DealRoles`, Chief election tie-break) come from `crypto/rand` inside the room, never from a client.
- Sets each deadline and sends `PhaseTimeout{Phase, Seq}`; never early. The 7-minute day clock keeps running through the nomination countdown and pauses only during defense and verdict; its timeout is sent with `DayClock: true`.
- Persists every event to `store.EventStore` before broadcasting.
- Every outgoing message for a phone is `ViewFor`/`Visible` for that player. `NightStepCalled` is public and identical for real and fake calls.
- **Leaving**: `LeaveRoom` (lobby only) removes the player and clears seats and role list; a lost connection keeps the player's place, and before the deal passes the host's powers on. After the deal there is no leaving.
- Dawn and dusk wait `Settings.Announce` (10 s, rules §2) so the speaker can read before the game moves on.
- `ConfigureTimers` (host only, lobby only) changes the phase times; a zero field is left alone and every other value must be within its limits (`internal/engine/timers_config.go`).
- Clock is injected (`room.Clock` interface) so timer tests run instantly.

## Transport
JSON over WebSocket. Message shapes in `docs/protocol.md`. Transport only decodes → `Client.Submit(cmd)` and encodes views; no game logic. `room.Manager` creates rooms by code and deletes idle ones. The transport imports `roles` only to build the commands of the command table (one row per command); `internal/transport/ws/wsclient` is the fake phone used by tests and `make demo`.

## Extending later
- Online play: `Settings.Mode` already has `InPerson | Online`, and a room created as `Online` fails with `ErrModeUnsupported` in v1. Adding the mode means a new `SpeakerPolicy` (every phone speaks, instead of one) in `room`, and chat/video in `transport`. The engine does not change.
- Database: implement `store.EventStore` for Postgres.
- New role: new file in `internal/roles/` + tests + row in rules.md.