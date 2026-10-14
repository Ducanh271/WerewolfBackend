# CLAUDE.md — Ma Sói backend

Backend for a Werewolf (Ma Sói) game-master app for in-person play. A narrator device reads instructions aloud; each player uses their phone for secret roles, night actions and day votes. The server is the single source of truth.

## Read before working
- `docs/plan.md` — task list. Work on ONE task per session. Tick it when done.
- `docs/architecture.md` — packages, dependency rules, core types. Follow it; propose changes before deviating.
- `docs/rules.md` — game rules (Vietnamese). It is the spec. Read only the sections the task names.
- If a rule is ambiguous or missing: STOP and ask. Never invent game rules.

## Stack
- Go 1.23+, standard library first. Allowed deps: `github.com/coder/websocket` (transport only), `github.com/stretchr/testify` (tests only). Ask before adding anything else.
- No database in v1: in-memory store behind an interface.

## Commands
- `make test` — run all tests (`go test ./... -race`)
- `make lint` — `go vet` + `golangci-lint run`
- `make fmt` — `gofmt` + `goimports`
- `make run` — start the server
Run `make fmt test lint` before declaring a task done. All must pass.

## Architecture rules (hard)
1. `internal/engine` is pure: no I/O, no goroutines, no `time.Now()`, no global `rand`, no logging. Time and randomness come in through commands/arguments.
2. Dependencies point inward only: `transport → room → engine`, `roles → engine`. The engine never imports room/transport/store.
3. Each role lives in its own file in `internal/roles/` and registers itself. Adding a role must not require editing other roles.
4. The server never sends full game state to a client. Every outgoing message goes through a view projection (`engine.ViewFor`). Secret info leaks are the #1 bug class.
5. State changes only through events: `Decide(state, cmd) → ([]Event, error)` then `Evolve(state, event) → state`.
6. Night resolution follows the ordered pipeline in `docs/rules.md` §9. Do not reorder steps.

## Conventions
- Code, identifiers, comments: English. User-facing strings (announcements): Vietnamese, kept in `internal/i18n/vi.go`.
- Errors: return typed errors from `engine/errors.go` (e.g. `ErrNotYourTurn`); never panic on player input.
- Tests: table-driven. Rule tests are scenario tests named after the rule, e.g. `TestNight_GuardAndWitchSaveSameTarget_AnnouncesProtected`. Every row of the table in `docs/rules.md` §9 must have a test.
- Use the test helper `enginetest.NewGame(t, roles...)` and builder-style actions; keep scenarios readable by a non-programmer.
- Small files (< 300 lines). One exported concept per file where reasonable.

## Glossary (rules.md term → code name)
| Vietnamese | Code |
|---|---|
| Sói | `Werewolf` |
| Sói nguyền | `CursedWolf` (action: `Curse`) |
| Sói tiên tri | `WolfSeer` |
| Dân làng | `Villager` |
| Tiên tri | `Seer` |
| Bảo vệ | `Guard` |
| Phù thủy | `Witch` (`HealPotion`, `PoisonPotion`) |
| Thợ săn | `Hunter` |
| Nguyệt Nữ | `Visitor` (blocks target's night ability) |
| Thám tử | `Detective` |
| Hiệp sĩ kiếm gỉ | `RustyKnight` |
| Cupid / cặp đôi | `Cupid` / `Lovers` |
| Phe Dân / Sói / thứ ba | `TeamVillage` / `TeamWolves` / `TeamLovers` |
| Quản trò | `Narrator` |
| Đề cử / giàn / bào chữa / giết-tha | `Nominate` / `Trial` / `Defense` / `Verdict` |

## Don't
- Don't write frontend code. Don't touch files outside the task's scope.
- Don't add features from `docs/rules.md` §10 (out of scope).
- Don't read the whole repo; open only files the task needs.
- Don't commit secrets or generated binaries.
