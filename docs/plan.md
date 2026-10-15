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
- [ ] **T19 Cupid + lovers.** Link colour, team recalculated on curse, third team win. *§7, §8.*
  - Decided: a wolf in a cross-team couple keeps `Player.Team = Wolves`. The couple's third-team status is computed at the couple level (from `State.Lovers`), never by changing `Player.Team`. Seer and Detective rely on this.
- [ ] **T20 Cross-role scenarios.** One test per row of §9 "Tình huống đã chốt"; fix whatever fails.

## Open questions (ask the user, don't guess)
- None open.
