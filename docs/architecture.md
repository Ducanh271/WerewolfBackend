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

Commands carry `At time.Time` when timing matters. Timers are not inside the engine: `room` sends `PhaseTimeout{Phase, Seq}` commands when a deadline passes; the engine ignores stale timeouts by `Seq`.

Randomness (role dealing) is a command input: `DealRoles{Seed int64}`, dealt by a deterministic shuffle so tests are reproducible.

### State (sketch)
```go
type State struct {
    Phase      Phase          // Lobby, Night, Dawn, Day, Trial, Verdict, HunterShot, Ended
    Seq        int            // increments on every phase change
    Night      int            // 1-based night counter
    NightStep  int            // index into the night call order
    Seats      []PlayerID     // seating circle, clockwise; "left" = index-1 (wraps)
    Players    map[PlayerID]*Player
    Pending    NightActions   // choices collected this night, resolved at dawn
    Day        DayState       // nominations (ordered), skip votes, trial, verdict votes, clock
    Lovers     *LoversLink
    Settings   Settings       // timers, role list
    Winner     *Team
}

type Player struct {
    ID      PlayerID
    Role    RoleID
    Team    Team            // can change (curse)
    Alive   bool
    Flags   map[string]int  // role-owned counters: potions left, curse used, last guarded, poisoned-until...
}
```

`Flags` lets roles keep their own state without changing `Player`.

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
```

The night call order (rules §3) comes from each role's `NightCall().Order`. Fake calls for dead/absent roles are produced by the room layer using the same order so the narrator timing never leaks information.

### Night resolution pipeline (rules §9)
An ordered slice of steps, each a pure function `func(*Resolution)`:

```
1 VisitorBlock → 2 CurseRecorded → 3 Bite(guard, witch heal) → 4 RustyKnight
→ 5 WitchPoison → 6 DelayedDeaths → 7 LoverDeaths → 8 Announce → 9 WinCheck
```

Roles hook into named steps via `Resolution` fields instead of editing the pipeline. The hunter shot is a separate phase after the announcement (it needs player input), and re-enters steps 7 and 9.

### Views (secrecy boundary)
```go
func ViewFor(s State, viewer Viewer) View   // Viewer = Player(id) | Narrator
```
Rules for the projection live in one file (`view.go`) with a test per role proving what that role can and cannot see. Dead players get the full view. The narrator view contains only public info plus what must be read aloud.

## Room layer
- One goroutine per room owns the state; commands arrive through a channel (no locks on game state).
- Persists every event to `store.EventStore` before broadcasting.
- On rejoin (room code + nickname), rebuilds the player's view from state.
- Clock is injected (`room.Clock` interface) so timer tests run instantly.

## Transport
JSON over WebSocket. Message shapes in `docs/protocol.md` (written in task T9). Transport only decodes → `room.Submit(cmd)` and encodes views; no game logic.

## Extending later
- Online play: add chat/voice in transport, reuse engine untouched.
- Database: implement `store.EventStore` for Postgres.
- New role: new file in `internal/roles/` + tests + row in rules.md.
