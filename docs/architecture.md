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

`engine` imports nothing from this repo except `i18n`. Nothing imports a
specific role: `cmd/server` and `cmd/demo` blank-import `roles` so every role
registers itself, and the other layers only read `engine.Registry`.

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

The engine does not know hosts (the room checks them, below). Setup ends with `DealRoles`: lobby commands then fail with `ErrWrongPhase` and a second deal with `ErrAlreadyDealt`.

Randomness (role dealing) is a command input: `DealRoles{Seed int64}`, dealt by a deterministic shuffle so tests are reproducible.

**Seeds never come from a client.** The room layer (T10) generates every seed with `crypto/rand` and puts it in the command. The transport must not accept a seed field from clients, or a player could pick their own role. The same goes for the Chief election tie-break.

### State (sketch)
```go
type State struct {
    Phase      Phase          // Lobby, RoleView, ChiefElection, Night, Dawn, Day, NominationLock, Trial, Verdict, Dusk, HunterShot, ChiefSuccession, Ended
    Seq        int            // increments on every phase or night step change
    Night      int            // 1-based night counter
    NightStep  int            // night steps called so far; current step = steps[NightStep-1]
    Seats      []PlayerID     // seating circle, clockwise; "left" = the next seat, index+1 (wraps, rules §6)
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
type MinCopiesRole  interface{ MinCopies() int }             // Villager: at least 1 (rules §2)
type CurseSource    interface{ CursesVillager() bool }      // reserves one wolf slot
type LoveSource     interface{ CreatesLovers() bool }       // Cupid: the phone check also runs on the nights it is called
type AlwaysCalled   interface{ AlwaysCalled() bool }        // called every night even if not in the role list: the Werewolf pack step

// What a role brings to the wire and the screens, also from its own file (audit C1).
type CommandRole    interface{ Commands() []CommandSpec }    // its wire commands: engine.WireCommand(name, func(args A, by PlayerID) Command)
type CardRole       interface{ CardOrder() int }            // its rank in the rules §9b table (10, 20, …): orders the public role list
```

A role's command args are a struct with json tags (`roles/wire_args.go` has the
shared shapes); if it implements `ArgsValidator` the transport checks it after
decoding (`wolfVote` needs `target`). A role's own errors are declared in its
file with `engine.NewError(code, message)`: the error carries its wire code.
The card text stays in `i18n/vi.go`, keyed by role ID, with the other
user-facing strings.

Game-wide rule limits (player count, wolf percentage, the cursed wolf's reserved slot) live in `internal/engine/limits.go`. Limits that belong to one role (minimum table size, number of copies) live in that role's own file.

Who is called at night, in what order and on which nights is rules §3. The order comes from each role's `NightCall().Order`; a night calls the roles in `Settings.Roles` that have a `NightCall` for that night, so the step list depends on the role list and the night number only, never on who is alive. The wolf pack step (`Werewolf`) implements `AlwaysCalled`, so it runs even when no plain Werewolf is in the list. The engine itself produces the calls as `NightStepCalled{Role, Duration}` events, public to every phone (role and duration only, no real/fake flag: fake calls are rules §3). `engine.CurrentNightRole(s)` tells a role whether it is its turn.

### Night resolution pipeline (rules §9)
An ordered slice of steps, each a pure function `func(*Resolution)`:

```
1 VisitorBlock → 2 Curse (nothing left to do) → 3 Bite(guard, witch heal) → 4 RustyKnight
→ 5 WitchPoison → 6 DelayedDeaths → 7 GriefWithoutCall + LoverDeaths → 8 Announce → 9 WinCheck
```

Roles hook into named steps via `Resolution` fields instead of editing the pipeline. The hunter shot is a separate phase after the announcement (it needs player input). `afterDeaths` (`after_death.go`) is the one place for what follows any death: each dead Hunter (role implements `DeathShooter`) is revealed (`HunterRevealed`, public, sets `FlagRevealed`) and gets `PhaseHunterShot`; `HunterShoot` or the timeout (`hunter_shot.go`) kills the target and re-enters `settle`, which repeats for a shot Hunter, then runs the win check (step 9) and the Chief succession. `State.Shot` remembers the queue and the phase to return to. The queue and the chain of deaths come in `HunterChainNoted` (no audience, full views only: its order is the pipeline's), right before the public `HunterShotStarted{Hunter, Duration}`. By day, `settle` marks the surviving partner of each dead lover (`LoverGrieving`, `FlagGrieving`) before the win check; they die in the next night (below).

**Win check** (when: rules §8): `settle` is the only caller of `CheckWin`. It runs after the morning announcement, after a hanging and after a Hunter's shot, never inside the night pipeline.

**Phone check** (schedule: rules §3): `engine.StepPhoneCheck` is an engine step, not a role. `engineSlots` (used by `nightSteps`) adds it right after the step of any role implementing `CurseSource` on the nights that role is called, and right after the step of a `LoveSource` (Cupid), followed there by `engine.StepLoversMeet` (`Settings.LoversMeet`; voice line `LoversMeetAnnounced`; only the lovers' `TurnView` shows it). Its start emits `PhoneCheckAnnounced` (voice line) and one `PhoneChecked{Player, Cursed, Love}` per living player; all notices have the same shape. `PhoneChecked` also sets `FlagLoveKnown` on the lovers: until then `loverView` shows nothing. A team changed during the night carries `FlagTeamHidden` until its player's `PhoneChecked`; every view uses `VisibleTeam`, so neither the cursed player nor the pack (`packView`, `Audience.Wolves`, lover link colour) shows the change earlier, while `State` and the Seer/win logic use the real team at once. The duration of a step comes from `NightStepDuration` (the phone check has its own `Settings.PhoneCheck`, the lovers' step `Settings.LoversMeet`).

**Winners** (who wins: rules §8): `CheckWin` returns `Outcome.Winners` in seat order. A broken cross-team couple stays on the third team: `PlayerDied` of the first lover sets `LoversLink.ThirdTeam` when the couple was cross-team at that moment. It is copied to `State.Winners` and `View.Winners` at the end.

### Views (secrecy boundary)
```go
func ViewFor(s State, log []Event, viewer Viewer) View  // the screen now; Viewer = PlayerViewer(id) (there is no narrator device; every phone is a player); log is only read once the game has ended, to build the summary below
func Visible(s State, log []Event, viewer Viewer) []Event // the events this screen may know, judged at the time of each event
```
The rules are summed up in `view.go`; the projection is in `view_project.go`, the event audiences in `view_audience.go`, the role cards (rules §9b) in `view_cards.go`: `View.RoleCards` (public role list with name and description), `PlayerView.Card` (set exactly when `Role` is; `fillCards` puts the player's current side on it: real side for full views, the side the viewer's screen knows for their own card and their partner's, the printed side for a revealed Hunter), `MyView.RoleCard` and `MyView.StatusCards`. Card text is the one place a view carries Vietnamese: it comes from `i18n` (`Card`, `StatusCard`), so the frontend holds no card text. Tests per role (`internal/roles/view_*_test.go`) prove what each screen can and cannot see. Dead players, and everybody once the game is over, get the full view. Every view carries the public part (table, votes, announcements, the current night call); the speaker phone reads it aloud, the others only show text.

- Engine events get their audience in `view_audience.go`. A role event declares its own by implementing `Audienced` (`Audience(s) Audience`: everyone, the living wolves, or listed players). An event without an audience is seen only by full views, so a forgotten audience hides information instead of leaking it.

**Summary** (audit K4): `View.Summary` is the game's own story, one line per night and day plus the final outcome, set once the game has ended (`view_project.go` calls `summaryFor(log)` then, `summary.go`). Every piece of text is a full sentence template from `i18n`, substituted there, never assembled from fragments in code (so adding a language later is one new `i18n` file, not a code change); `summary.go` only joins whole rendered sentences with punctuation (a night's clauses, a final "→ outcome", a day's clauses). It freely names a secret actor and target (who the pack bit, who the Witch healed) that stayed hidden from living players during play, because by then `seesAll` already gives everyone the whole raw log too (rules §4) — the summary only turns that log into a line a non-programmer can read, instead of leaving it to the frontend. A role event that should appear in it implements `Narrated` (`Narrate(name) (order int, text string, group PlayerID, ok bool)`, in `narrate.go`), the same extension point as `Audienced`, since `engine` cannot switch on a role's concrete event type. `order` places the clause within its night/day line (the `Narrate*`/`narrate*` constants, matching the call order of rules §9); `group` merges clauses about the same player into one sentence ("Sói cắn An, Phù thủy cứu An.") instead of two, `""` never merging. The engine's own clauses (the pack's bite, the dawn outcome, a hanging, a Hunter's shot, the Chief) are placed the same way without asking a role. A `NewGame` (audit K3) starts the story over: only the events from the latest `RolesDealt` on are narrated.
- `PlayerDied` is never shown to the living: the order of a night's deaths would tell who was bitten and who was poisoned.
- Role-specific secrets derived from state go through optional role interfaces, e.g. `BiteWitness` (the Witch sees the bite on her turn while she can heal, not when visited). `Pending.Blocked` (target → blocking role, engine-internal, stripped from every view) drives the generic "cannot use your ability tonight" notice (`TurnView.Blocked`, never names the blocker) and is enforced by `NightActor` (`ErrBlocked`; it first checks that the actor holds the step's role, a wolf for the pack step, and answers anybody else with the bare `ErrNotYourTurn`, so no error tells a player their own state) and `WolfVoters`, so a future blocking role only records into it. A role that must act when its night step runs out (the Visitor's forced random visit) implements `NightTimeoutRole`; the seed comes from the room's `PhaseTimeout.Seed`; `TeamChanged` (the curse) tells only the player concerned.
- **Rusty knight (rules §6)**: a role with a passive ability implements `BiteAvenger`; pipeline step 4 (`stepKnight`) asks the bitten player's role which wolf to poison (the Knight uses `engine.NearestLivingWolf`, direction +1 = the left, the next seat). `resolveNight` records it as `WolfPoisoned{Wolf, Night}`, which sets `FlagWolfPoisoned` on that wolf and is visible to nobody alive. From the next night (`engine.Doomed`) the wolf is out of `WolfVoters` and step 6 (`stepDelayedDeaths`) adds it to the night's deaths, so it is announced with the others and Hunter/Chief consequences run through `afterDeaths`. The wolf stays `Alive` in state and in every view until then, so nothing leaks during the night. Its own phone gets `TurnView.Doomed` and the pack gets `PackView.Doomed` once the pack step has started.
- Message shapes: `docs/protocol.md`.

## Room layer
- One goroutine per room owns the state; commands arrive through a channel (no locks on game state). A panic in a job (a bug in the engine or a role) is recovered and logged with the room code: a command gets `ErrInternal` and changes nothing (`apply` folds the new state before keeping any event), the room and its timers go on, other rooms are untouched.
- **Host** = the player who created the room (and plays; rules §2). The room checks host commands, the engine does not know hosts: `SetSeats`, `ConfigureRoles`, `ConfigureTimers`, `ConfigureSound`, `DealRoles` and `KickPlayer` only before the deal (`hostOnly`, `room/submit.go`), `NewGame` once the game has ended, `TransferHost` at any time. Losing the title by `Leave` or by a drop longer than `hostDisconnectGrace` (`room/join.go`) only happens before the deal.
- **Speaker policy**: who speaks sits behind `room.SpeakerPolicy`. `InPersonSpeaker` (the default) is rules §1 "Máy làm loa": one phone at a time, remembered in `Room.speaker`. `AllPhones` has every phone speak; the host picks between them with `ConfigureSound` before the deal (`engine.Settings.Sound`, rules §1 "Âm thanh"), and it is what the online mode will use. The room sets `speaker: true` in a player's `view` message (an envelope field, not part of `engine.View`) for whichever phones the policy names.
- **Identity**: the room issues each player a secret token and a 4-digit PIN (both from `crypto/rand`) behind an `Identity` interface. The token (or the PIN) decides who a connection is; commands never carry a player id from the client, the room fills `By`/`Voter`. Rejoin and the PIN lock follow rules §2 "Vào lại phòng" (`room/identity.go`, injected clock).
- **Seeds** (`DealRoles`, Chief election tie-break) come from `crypto/rand` inside the room, never from a client.
- Sets each deadline and sends `PhaseTimeout{Phase, Seq}`; never early. The day clock (`room/dayclock.go`) runs and pauses as rules §4 says; its timeout is sent with `DayClock: true`.
- Persists every event to `store.EventStore` before broadcasting.
- Every outgoing message for a phone is `ViewFor`/`Visible` for that player. A phone whose view is unchanged and has no new event is sent nothing (`send` keeps the last view per connection), so a secret night action reaches no other phone. `NightStepCalled` is public and identical for real and fake calls.
- **Leaving**: `LeaveRoom` (lobby only) removes the player; only their seat is freed, the others keep their order and the role list stays (`evolve.go`). A lost connection keeps the player's place, and before the deal passes the host's powers on after the grace above. After the deal there is no leaving.
- Dawn and dusk wait `Settings.Announce` (rules §2) so the speaker can read before the game moves on.
- `ConfigureTimers` (host only, lobby only) changes the phase times; a zero field is left alone and every other value must be within its limits (`internal/engine/timers_config.go`).
- Clock is injected (`room.Clock` interface) so timer tests run instantly.

## Transport
JSON over WebSocket. Message shapes in `docs/protocol.md`. Transport only decodes → `Client.Submit(cmd)` and encodes views; no game logic. `room.Manager` creates rooms by code and deletes idle ones. The command table is built once by `ws.New` from the engine's core commands (`commands.go`) and every registered role's `Commands()`; `ws.New` returns an error, and the server does not start, when two commands share a name. Error codes come from the errors themselves (`engine.ErrorCode`); the transport keeps only its own codes (`bad_message`...) and the special cases (`pin_locked` with `retryAfter`, `bad_credentials` without detail). The transport never imports `roles`; `internal/transport/ws/wsclient` is the fake phone used by tests and `make demo`.

## Extending later
- Online play: `Settings.Mode` already has `InPerson | Online`, and a room created as `Online` fails with `ErrModeUnsupported` in v1. Adding the mode means a new `SpeakerPolicy` (every phone speaks, instead of one) in `room`, and chat/video in `transport`. The engine does not change.
- Database: implement `store.EventStore` for Postgres.
- New role: new file in `internal/roles/` (ID, team, night call, `Commands()`, `CardOrder()`, its own errors) + its card in `i18n/vi.go` and guide entry in `i18n/guide.go` (user-facing text) + tests (scenarios can use `Night.Act` before the role has verbs of its own) + a row in rules.md and in protocol.md. `transport/ws/newrole_test.go` shows a throwaway role doing exactly this.