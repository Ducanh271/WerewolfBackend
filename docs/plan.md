# Plan

One task per agent session. Start each session with:
> Read CLAUDE.md and docs/plan.md. Do task Tn. Plan first, wait for my OK, then implement.

A task is done when `make fmt test lint` passes and its acceptance tests exist. Tick `[x]` and commit.

## Milestone 1 — Engine core

- [x] **T0 Project skeleton.** `go mod init`, folders from architecture.md, Makefile targets working, empty `main`, `.golangci.yml`, one passing placeholder test.
- [ ] **T1 Core types.** `State`, `Player`, `Team`, `Phase`, `Command`, `Event`, `Decide`/`Evolve`, typed errors, role `Registry`. Lobby commands: `JoinRoom`, `SetSeats`, `ConfigureRoles`. *Rules: §2.* Accept: joining, seating, invalid configs rejected (fewer than 5 or more than 30 players, too many wolves vs the 30% table, Detective ≤ 8 players, Cupid < 10).
- [ ] **T2 Dealing roles.** `DealRoles{Seed}` with deterministic shuffle; Cursed-wolf vs plain-wolves option. *Rules: §2.* Accept: same seed → same deal; wolf count matches table for 5, 7, 10, 14, 20, 30.
- [ ] **T3 Test kit.** `enginetest` scenario builder: `g := NewGame(t, Seats("An","Binh",...), Roles(...))`, `g.Night().Guard("An").Protects("Binh")...`, `g.Dawn().ExpectDead(...)`, `ExpectAnnouncement(...)`. Readable by a non-programmer.

## Milestone 2 — Basic loop (Werewolf, Villager, Seer, Guard)

- [ ] **T4 Night phase machine.** Call order, per-step timeout → no action, fake calls for dead/absent roles (exposed as events for the narrator). *Rules: §3.*
- [ ] **T5 Wolves + Guard + Seer.** Wolf live voting, majority > half at timeout, guard rules (self ok, not same target twice in a row), seer result. Pipeline steps 3, 8, 9. *Rules: §5 Sói, §6 Tiên tri/Bảo vệ, §9.*
- [ ] **T6 Dawn announcements.** Three announcement variants. *Rules: §4 Thông báo buổi sáng.*
- [ ] **T7 Day phase.** Shared 7-min day clock (paused in trial), skip button > 66% (nominating cancels own skip), public nominations, tie → earliest nominated, defense, kill vs spare (kill > spare), re-nominate after spare, max 1 hanging. *Rules: §4.*
- [ ] **T8 Win check.** After every death and team change. Lovers excluded from wolf count. *Rules: §8.*

## Milestone 3 — Secrecy and room

- [ ] **T9 Views.** `ViewFor(player|narrator)`; tests per role proving no leaks; dead players see everything. Write `docs/protocol.md` (message shapes) for review.
- [ ] **T10 Room + timers.** Room goroutine, injected clock, timeouts → commands, event store in memory, rejoin by room code + nickname.
- [ ] **T11 WebSocket transport.** Implements `docs/protocol.md`. One end-to-end test with fake clients playing a 5-player game.

## Milestone 4 — Remaining roles (one task each)

- [ ] **T12 Witch.** *§6 Phù thủy, §9.*
- [ ] **T13 Hunter** (shot phase, any cause of death, reveals role). *§6 Thợ săn.*
- [ ] **T14 Visitor (Nguyệt Nữ).** Blocks target; per-role effect of being blocked. *§6, §5.*
- [ ] **T15 Cursed wolf.** Immediate effect, nights 1–3, blocked on night 3 = lost, all phones vibrate event (in-person). *§5 Sói nguyền.*
- [ ] **T16 Wolf seer.** *§5 Sói tiên tri.*
- [ ] **T17 Detective.** Uses seating; one use; not lost when blocked. *§6 Thám tử.*
- [ ] **T18 Rusty knight.** Nearest living wolf to the left dies next night. *§6, §9 step 4.*
- [ ] **T19 Cupid + lovers.** Link colour, team recalculated on curse, third team win. *§7, §8.*
- [ ] **T20 Cross-role scenarios.** One test per row of §9 "Tình huống đã chốt"; fix whatever fails.

## Open questions (ask the user, don't guess)
- None yet. Add here when a rule is unclear.
