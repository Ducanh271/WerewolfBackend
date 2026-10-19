# CLAUDE.md — Ma Sói backend

Backend for a Werewolf (Ma Sói) game-master app for in-person play. Each player uses their own phone for secret roles, night actions and day votes; the host's phone is the speaker and reads the announcements aloud, the other phones only show text. The server is the single source of truth.

## Read before working
- `docs/plan.md` — task list. Work on ONE task per session. Tick it when done.
- `docs/architecture.md` — packages, dependency rules, core types. Follow it; propose changes before deviating.
- `docs/rules.md` — game rules (Vietnamese). It is the spec and the only place in the repo where rules are written: other docs and code comments cite its section numbers (`rules §6 "Thám tử"`) instead of re-explaining. Read only the sections the task names.
- If a rule is ambiguous or missing: STOP and ask. Never invent game rules.

## Stack
- Go 1.24+, standard library first. Allowed deps: `github.com/coder/websocket` (transport only), `github.com/stretchr/testify` (tests only), `golang.org/x/text` (nickname NFC, pinned to a version that builds with Go 1.24). Ask before adding anything else.
- No database in v1: in-memory store behind an interface.

## Commands
- `make test` — full check: all tests, race detector, simulation at its default `SIM_GAMES=300` (`go test ./... -race`)
- `make test-fast` — fast loop: no race detector, skips the simulation and other slow tests (`go test ./... -short`)
- `make lint` — `go vet` + `golangci-lint run`
- `make fmt` — `gofmt` + `goimports`
- `make run` — start the server
While iterating, run only the affected package (`go test ./internal/<pkg> -run <Name>`) or `make test-fast`. Run `make fmt test lint` ONCE before committing — all must pass. Run a bigger simulation (`SIM_GAMES=1000+`) only when the task asks for it.

## Architecture rules (hard)
1. `internal/engine` is pure: no I/O, no goroutines, no `time.Now()`, no global `rand`, no logging. Time and randomness come in through commands/arguments.
2. Dependencies point inward only: `transport → room → engine`, `roles → engine`. The engine never imports room/transport/store.
3. Each role lives in its own file in `internal/roles/` and registers itself. Adding a role must not require editing other roles.
4. The server never sends full game state to a client. Every outgoing message goes through a view projection (`engine.ViewFor`). Secret info leaks are the #1 bug class.
5. State changes only through events: `Decide(state, cmd) → ([]Event, error)` then `Evolve(state, event) → state`.
6. Night resolution follows the ordered pipeline in `docs/rules.md` §9. Do not reorder steps.

## Conventions
- Code, identifiers, comments: English. User-facing strings (announcements): Vietnamese, kept in `internal/i18n/vi.go`.
- Mọi chữ người dùng thấy nằm trong i18n dưới dạng mẫu câu đầy đủ có placeholder; không ghép mảnh câu trong code (để sau thêm ngôn ngữ khác chỉ cần thêm 1 file).
- Errors: return typed errors from `engine/errors.go` (e.g. `ErrNotYourTurn`); never panic on player input.
- Tests: table-driven. Rule tests are scenario tests named after the rule, e.g. `TestNight_GuardAndWitchSaveSameTarget_AnnouncesProtected`. Every row of the table in `docs/rules.md` §9 must have a test.
- Use the test helper `enginetest.NewGame(t, roles...)` and builder-style actions; keep scenarios readable by a non-programmer.
- Small files (< 300 lines). One exported concept per file where reasonable.
- Git: work on main only. Commit and push directly to main after make fmt test lint pass. Do not create branches.

## Talking to me
- Reply to me in Vietnamese. Keep it short: what changed, test results, questions.
- When you ask me a question, explain it in game terms (players, roles, what happens at the table), not code terms, and give your recommendation.
- Code, identifiers, commit messages and docs stay in English.

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
| Nguyệt Nữ | `Visitor` |
| Thám tử | `Detective` |
| Hiệp sĩ kiếm gỉ | `RustyKnight` |
| Cupid / cặp đôi | `Cupid` / `Lovers` |
| Phe Dân / Sói / thứ ba | `TeamVillage` / `TeamWolves` / `TeamLovers` |
| Lời dẫn của quản trò | voice lines (announcements) played by the speaker phone |
| Chủ phòng | `Host` |
| Máy làm loa | `Speaker` |
| Đề cử / giàn / bào chữa / giết-tha | `Nominate` / `Trial` / `Defense` / `Verdict` |

- When the architecture changes (layers, dependency rules, request flow, config), update `docs/huong-dan-du-an.md` (Vietnamese project guide) in the same change.

## Don't
- Don't write frontend code. Don't touch files outside the task's scope.
- Don't add features from `docs/rules.md` §10 (out of scope).
- Don't read the whole repo; open only files the task needs.
- Don't commit secrets or generated binaries.
