# Plan

One task per agent session. Start each session with:
> Read CLAUDE.md and docs/plan.md. Do task Tn. Plan first, wait for my OK, then implement.

A task is done when `make fmt test lint` passes and its acceptance tests exist. Tick `[x]` and commit.

## Milestone 1 — Engine core

- [x] **T0 Project skeleton.** `go mod init`, folders from architecture.md, Makefile targets working, empty `main`, `.golangci.yml`, one passing placeholder test.
- [x] **T1 Core types.** `State`, `Player`, `Team`, `Phase`, `Command`, `Event`, `Decide`/`Evolve`, typed errors, role `Registry`. Lobby commands: `JoinRoom`, `SetSeats`, `ConfigureRoles`. *Rules: §2.* Accept: joining, seating, invalid configs rejected (fewer than 5 or more than 30 players, too many wolves vs the 30% table, Detective ≤ 8 players, Cupid < 10).
- [x] **T2 Dealing roles.** `DealRoles{Seed}` with deterministic shuffle; Cursed-wolf vs plain-wolves option. *Rules: §2.* Accept: same seed → same deal; wolf count matches table for 5, 7, 10, 14, 20, 30.
- [x] **T3 Test kit.** `enginetest` scenario builder: `g := NewGame(t, Seats("An","Binh",...), Roles(...))`, `g.Night().Guard("An").Protects("Binh")...`, `g.Dawn().ExpectDead(...)`, `ExpectAnnouncement(...)`. Readable by a non-programmer.

## Milestone 2 — Basic loop (Werewolf, Villager, Seer, Guard)

- [x] **T4 Night phase machine.** After dealing: `RoleView` phase with countdown (`Settings.RoleView`, default 30s); on `PhaseTimeout` it moves on (hook for the Chief election of T6b, for now straight to night 1). No host commands after dealing. Night call order from `NightCall()`; every call, real or fake, runs its full time and ends only by `PhaseTimeout` (never early). The narrator event carries only the role and the duration, never a fake flag. *Rules: §2, §3.* **Also:** write the real `enginetest.Driver` night half (`StartNight`, `EndNight`) and make it the default in `enginetest` (replacing `unbuiltDriver`), so `g.Night()` / `g.Dawn()` work.
- [x] **T5 Wolves + Guard + Seer.** Wolf live voting, majority > half at timeout, guard rules (self ok, not same target twice in a row), seer result. Pipeline step 3 only; the dawn announcement event with its three variants is emitted (wording in T6), win check is T8. *Rules: §5 Sói, §6 Tiên tri/Bảo vệ, §9.*
- [x] **T6 Dawn announcements.** Three announcement variants. *Rules: §4 Thông báo buổi sáng.*
- [x] **T6b Chief election + succession.** `Chief` title (not a role, original role unchanged, outside the 30% limit). Public 1-minute vote after role viewing, before night 1; self vote allowed; tie → random pick from `Seed` in the command (room generates it with crypto/rand). On Chief death (any cause) 15s to name a living successor, announced publicly; timeout → title lost; Hunter shoots first if both die together. *Rules: §7b.*
- [x] **T7 Day phase.** Shared 7-min day clock (paused in trial), skip button > 66% (nominating cancels own skip), public nominations, tie → earliest nominated, defense, kill vs spare (kill > spare), votes are weighted (Chief x2 in nominations and in both kill and spare; a Chief on trial cannot vote), re-nominate after spare, max 1 hanging. *Rules: §4.* **Also:** write the day half of the real `enginetest.Driver` (`NobodyHanged`, `Hang`) and make it part of the default, so `g.Day().NobodyHanged()` / `g.Day().Hanged(...)` work.
- [x] **T8 Win check.** After every death and team change. Lovers excluded from wolf count. *Rules: §8.*

## Milestone 3 — Secrecy and room

- [x] **T9 Views.** `ViewFor(player|narrator)`; tests per role proving no leaks; dead players see everything. Write `docs/protocol.md` (message shapes) for review.
- [x] **T10 Room + timers.** Room goroutine, injected clock, timeouts → commands, event store in memory, rejoin per rules §2 "Vào lại phòng": on join the room issues a secret token (kept by the phone) and a 4-digit PIN (shown only on that player's screen); rejoin with the token (same phone) or nickname + PIN (other phone); 5 wrong PINs lock PIN rejoin for that nickname for 1 minute (injected clock; token rejoin unaffected). Token and PIN come from `crypto/rand`; identity checks sit behind an interface so Google login can be added later (rules §10). The room generates every seed (`DealRoles`, Chief election tie-break) with `crypto/rand`; seeds never come from a client.
- [x] **T11 WebSocket transport.** Implements `docs/protocol.md`. One end-to-end test with fake clients playing a 5-player game.

## Milestone 4 — Remaining roles (one task each)

- [x] **T12 Witch.** *§6 Phù thủy, §9.*
- [x] **T13 Hunter** (shot phase, any cause of death, reveals role). *§6 Thợ săn.*
- [x] **T14 Visitor (Nguyệt Nữ).** Blocks target; per-role effect of being blocked. *§6, §5.*
- [x] **T15 Cursed wolf.** Immediate effect, nights 1–3, blocked on night 3 = lost, "phone check" step (rules §3) with the same notice for everybody; win check only by day (rules §8). *§5 Sói nguyền.*
- [x] **T16 Wolf seer.** *§5 Sói tiên tri.*
- [x] **T16b Demo and simulation.** `make demo PLAYERS=.. ROLES=..`, `make demo-12`; 10–13 player tables; 300-game bot simulation (`SIM_GAMES`, `SIM_SEED` to replay).
- [x] **T17 Detective.** Uses seating; one use; not lost when blocked; called every night (fake after use). *§6 Thám tử.*
- [x] **T18 Rusty knight.** Bitten Knight (not Guarded) poisons the nearest living wolf to his left (next seat); the wolf dies when the pack step starts next night, announced at dawn. *§6, §9 steps 4 and 6.*
  - Decided: "left" = next seat (seats are numbered clockwise). The poisoned wolf stays `Alive` in views until the morning announcement.
- [x] **T19 Cupid + lovers.** Link colour, team recalculated on curse, third team win. *§7, §8.*
  - Decided: a wolf in a cross-team couple keeps `Player.Team = Wolves`. The couple's third-team status is computed at the couple level (from `State.Lovers`), never by changing `Player.Team`. Seer and Detective rely on this.
  - Decided: lovers learn of the pair at the phone check (one notice combining curse and love; the step runs night 1 with Cupid or the Cursed wolf). A lover who dies by day does not take the partner at once: the partner dies in the next night, at their first night call, or at §9 step 7 if nothing calls them. Shared with the Rusty knight's poison (`delayed_death.go`). A lover who dies in the night takes the partner the same night.
- [x] **T20 Cross-role scenarios.** One test per row of §9 "Tình huống đã chốt"; fix whatever fails.
  - Decided: a player dead by the rules but not yet announced (Knight-poisoned wolf, grieving lover) counts as alive for the Detective (neighbours and wolf count). The Knight's poison still skips such a wolf (internal, invisible).
  - Decided: the Visitor is called first every night, after Cupid and the lovers on night 1.
  - Decided: night 1 with Cupid: Cupid → phone check (love) → `LoversMeet` (10 s, the lovers look at each other) → Visitor and the rest. With the Cursed wolf too, its own phone check stays after its step (two checks on night 1). Lovers never look at each other during a phone check.
  - Decided: once a couple is cross-team (third team), both lovers stay on the third team for good, also after one of them died (`LoversLink.ThirdTeam`, frozen at the first lover's death). They win only as the couple: never with the wolves or the village (`Outcome.Winners`).
  - Decided: 0 plain wolves allowed when the Cursed wolf or the Wolf seer is in the game (wording of rules §2).
- [x] **T20b Role cards.** Card texts of rules §9b in `i18n/vi.go`; `me.roleCard`, `me.statusCards`, public `roleCards`, every seat's card for dead players and at the end; same shape for every role.
  - Decided: a card's team and goal are the player's current side (cursed → Sói; cross-team couple → "Phe thứ ba"); the public role list carries each role's own team.

## Milestone 5 — Frontend

- [ ] **T21 Frontend guide (docs/frontend) — do LAST, after the backend backlog is done.**
  - Role cards have a fixed size, are hidden by default, and are revealed only by press-and-hold.
  - A dead player's phone keeps receiving live night updates, so at night it must stay dark/blank until tapped, otherwise neighbours can see the screen change when someone acts.

## Phase 2 — fix the audit, reach "self-use" (play in person with friends)
Goal: a group of 8–13 friends plays 3 games in a row on their own phones (link/QR, no install, not on the same wifi), first-timers understand their role from the app, nothing leaks, a dead phone can rejoin from another phone, an end-of-game summary is shown.
IDs in brackets refer to docs/audit.md.

- [x] T22 Audit → docs/audit.md (Opus)
- [ ] P0 Throwaway prototype: one human + bots, browser speech (Sonnet). Any time.
- [x] T23 Fix leaks: [S1] [S2] [S3], leak tests [T1], panic recovery [R4] (Opus)
- [x] T24 Rules vs code: [L1] [L2] [L4] [L5] [R3] + missing tests [T3] [T5] (Sonnet)
- [x] T25 Voice lines for the speaker phone [L3] + in-app rules guide and per-role how-to (merged from T28b); all Vietnamese text approved by me first (Sonnet)
- [x] T26 Connection robustness: [R1] [R2] [S6] [S7] + tests [T4] (Sonnet)
- [x] T27a Lobby: [S8] host = room creator, [S4] nickname rules, [K2] lobby kick (Sonnet)
  - Decided: leaving or being kicked only removes that one player from the seat order (the rest keep their relative order, no reshuffle) and never touches the chosen role list. If the role list stops matching the new player count (wolf ratio, Detective/Cupid headcount), the lobby blocks "Deal" and tells the host why, instead of silently clearing seats/roles for everyone.
  - Decided: nickname is 1–20 characters after trimming, NFC-normalized, and restricted to Vietnamese/Latin letters, digits and single spaces (no other scripts, no control or invisible characters) — closes both the length-spam and the look-alike-character tricks from audit S4.
  - Decided: a `TransferHost` command lets the host hand the title to anyone currently in the room, no consent needed, usable at any time including mid-game (so the table can line up who will run the next game, T27b, without waiting for this one to end).
  - Decided: the creator's `hostKey` (S8) only ever grants host once — the first time it is presented. Once the title has moved on (by `TransferHost` or by the 60 s auto-transfer below), the creator does not get it back by presenting the key again.
  - Decided: the existing 60 s auto-transfer when the host leaves or disconnects (to the next-joined player still connected, rules §2) is unchanged.
- [x] T27b Game lifecycle: [K3] new game in the same room, [K1] abort by vote (Sonnet)
  - Decided (K1, replaces the T24 "over 70% of everyone who ever joined" rule — see `docs/rules.md` §8b): only **living** players may vote to abort, and only while it is **day** (`Day`, `NominationLock`, `Trial`, `Verdict`, `Dusk` — never at night, during role viewing, or during the Chief election, so voting never reveals who is awake). The vote passes once yes votes reach **2/3 of the living players**, computed with integers as `yes*3 >= alive*2` (3 alive → needs 2, 4 → 3, 5 → 4, 6 → 4, 9 → 6), re-checked after every vote and after every death that happens the same day (a hanging, a Hunter's shot) — a death alone can push a pending vote over the line, with no new vote cast. On a pass the game ends exactly like a normal win: `Ended`, no winner, **everyone sees every role and the full history**. The table only goes back to the lobby when the host runs `NewGame` (K3 below), same as after a normal win. Abort votes are cleared when the next night starts (a vote does not carry over to the next day).
  - Decided (K3): once a game is `Ended` (by winning or by this abort vote), the host can start a new game in the same room with a `NewGame` command. It keeps the players, their seats and the configured role list, resets the match (no roles dealt, nobody dead, no night/day state), and the host re-deals with a fresh seed as usual.
- [x] T28a End-of-game summary [K4] (Sonnet)
- [ ] T29 Deploy: [D1]–[D9], [S5] IP limits, [R7] room expiry (Sonnet)
- [ ] T30 Code health: [C1] roles self-register commands, [C6] rules single source, [C2]–[C5] cleanup, [R5] [T2] faster broadcast (Opus) — split into sub-tasks, one session each, done in this order: T30b → T30c → T30d → T30a. No game-rule or protocol (wire format) changes in any of them; every sub-task must leave `make fmt test lint` green.
  - [x] T30b Roles self-register commands, cards and errors [C1] (Opus). Riskiest, so first while the tests are fresh: touches transport parsing, i18n and error codes for every role.
    - Changes: each role file declares its wire command name + payload type, card order, and its own errors; `transport/ws/commands.go` `commandTable` is built from the `roles` registry instead of a hand-written list; `vi.go` `roleCardOrder` derived from the role's order field (card text stays in `vi.go`, keyed by role); `transport/ws/errors.go` `codeTable` generated from the error variable names (or a registry) so a new error needs no table edit; `enginetest` verbs available generically for registered actions. Dependency direction unchanged (`transport → room → engine`, `roles → engine`; transport reads the registry, never imports a specific role).
    - Packages: `internal/roles`, `internal/engine` (registry.go, errors.go), `internal/transport/ws` (commands.go, errors.go), `internal/i18n` (vi.go), `internal/engine/enginetest`; `docs/protocol.md`, `docs/architecture.md`, `docs/huong-dan-du-an.md` (architecture change).
    - Proof: (1) golden test captured BEFORE the refactor: the full list of wire command names, payload shapes and error codes — must be identical after (no protocol change); (2) a test with a throwaway fake role registered only in the test file proves it gets a command, card slot and error code without editing any other file; (3) all existing role, transport and simulation tests pass unchanged.
  - [x] T30c Cleanup: shared helpers [C5], dead code [C2], split long test file [C4] (Sonnet).
    - Changes: engine helpers `SetFlag`, `HolderOf`, the "not the same target two nights in a row" rule, merge `livingCount`/`anyoneAlive`; use them in `guard.go`, `visitor.go`, `cupid.go` and the ~10 `if p.Flags == nil` sites. Delete or use `engine.BlockedBy`, `DealRoles.DayClock` (+ wrong comment), `store.Load`, the unused server i18n keys; fix or document the `secrets()` / `Pending.Blocked` mismatch. Split `roles/visitor_test.go` into block tests and random-pick tests (< 300 lines each).
    - Packages: `internal/engine`, `internal/roles`, `internal/store`, `internal/i18n`.
    - Proof: pure refactor — all existing tests pass unchanged (no test edited except moving visitor tests between files; test count for visitor stays the same); small table tests for each new helper; simulation at default size.
  - [x] T30d Docs single source [C6] + doc drift [C3] (Sonnet).
    - Changes: `docs/rules.md` is the only place rules are written; `protocol.md`, `architecture.md` and code comments link to rule section numbers instead of re-explaining. Fix C3 drift: seat "left = next seat (+1)" in `command.go` `SetSeats` and `architecture.md` `State` sketch, `protocol.md` title (drop "T9 draft"), full `Phase` list in `architecture.md`, Go version in CLAUDE.md vs `go.mod` (1.24), `players` empty in lobby until seats are set. Card text in `vi.go` stays (players see it) but gets a test.
    - Packages: `docs/`, `CLAUDE.md`, `internal/i18n` (test only), comment-only edits in `internal/engine`.
    - Proof: new test that each role card in `vi.go` matches the role's row in `docs/rules.md` §9b (fails if one changes without the other); existing tests pass; no non-test Go code changes besides comments.
  - [ ] T30a Measure first: broadcast benchmark [R5], every-night leak check [T2] (Opus).
    - Step 1 (measure only): write a broadcast benchmark — one 13-player full game and one 30-player full game (scripted or sim-driven to the end, many day vote changes), measuring ms per broadcast (one valid command fanned out to every connection). Report the numbers. No change to the view path in this step.
    - Decision: if a broadcast at 13 players takes < 5 ms, STOP — move [R5] to the "Later" list with the measured numbers and do not rewrite the view path. Only if it is really slow, do the incremental view (per-connection cursor in `engine`, used by `room/broadcast.go`, full replay when the viewer dies / game ends / `NewGame` / reconnect), proven by a differential test: incremental view == `ViewFor` full replay for every viewer after every event, also checked inside the simulation; existing leak tests [T1] unchanged.
    - [T2], independent of [R5]: if checking every night is slow only because of the simulation, `checkEvents` runs on every night under `make test` (full) and keeps one night in three under `make test-fast` (`testing.Short()`).
    - Packages: `internal/room` (benchmark), `internal/roles` (sim_test.go); `internal/engine` only if the incremental view is needed.
    - Proof: benchmark numbers in the commit message; `make test` with every-night checks still passes at default `SIM_GAMES=300`.
- [ ] T30e Server sends the text of the notices that today are only flags (`me.turn` blocked / doomed / grief, `me.phoneCheck` cursed / love, the pack's doomed-wolf notice) from i18n, like `roleCard`, so no frontend rewrites them. New fields only, old fields unchanged; update `docs/protocol.md` + golden test. The keys exist (`NightBlocked`, `NightWolfDoomed`, `NightGrief`, `NightPackDoomed`, `Notice*` in `i18n/keys.go`).
  - Also: view ở sảnh có danh sách người trong phòng (id, nickname, đang kết nối hay không, ai là chủ phòng), để màn hình tự đủ thông tin khi vào lại phòng — không bắt frontend gom từ sự kiện PlayerJoined. (Today `players` is empty in the lobby until `setSeats`, see `docs/protocol.md`.)
- [ ] T21 Frontend guide docs/frontend (Opus) — after T23–T30
- [ ] F1… Phone UI (planned separately), following the secrecy checklist
- [ ] P1 Playtest with friends, collect feedback, rebalance

Later (Low in audit): [S9] [S10] [R6] [R8].

Later (not designed yet): room visibility — public (listed, anyone can join) vs
private (invite-only, needs a friends feature).

Later (not designed yet): Đa ngôn ngữ: thêm `i18n/en.go` cùng khoá, người chơi
chọn ngôn ngữ khi vào phòng, `ViewFor` soạn câu theo ngôn ngữ người xem.
