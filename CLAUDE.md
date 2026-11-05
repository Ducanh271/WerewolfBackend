# CLAUDE.md — Ma Sói backend

Backend for a Werewolf (Ma Sói) game-master app for in-person play. Each player uses their own phone for secret roles, night actions and day votes; the host's phone is the speaker and reads the announcements aloud, the other phones only show text. The server is the single source of truth.

## Read before working
- `docs/plan.md` — task list. Work on ONE task per session. Tick it when done.
- `docs/architecture.md` — packages, dependency rules, core types. Follow it; propose changes before deviating.
- `docs/rules.md` — game rules (Vietnamese). It is the spec and the only place in the repo where rules are written: other docs and code comments cite its section numbers (`rules §6 "Thám tử"`) instead of re-explaining. Read only the sections the task names.
- If a rule is ambiguous or missing: STOP and ask. Never invent game rules.

## Stack
- Go 1.24+, standard library first. Allowed deps: `github.com/coder/websocket` (transport only), `github.com/stretchr/testify` (tests only), `github.com/santhosh-tekuri/jsonschema/v6` (tests only: validates `docs/api` schemas and examples), `golang.org/x/text` (nickname NFC, pinned to a version that builds with Go 1.24). Ask before adding anything else.
- No database in v1: in-memory store behind an interface.

## Commands
- `make test` — full check: all tests, race detector, simulation at its default `SIM_GAMES=300` (`go test ./... -race`)
- `make test-fast` — fast loop: no race detector, skips the simulation and other slow tests (`go test ./... -short`)
- `make lint` — `go vet` + `golangci-lint run`
- `make fmt` — `gofmt` + `goimports`
- `make run` — start the server
- `make api-docs` — regenerate `docs/api/schema` and `docs/api/examples` after a wire change (the tests fail while they are stale)
While iterating, run only the affected package (`go test ./internal/<pkg> -run <Name>`) or `make test-fast`. Before committing run `make fmt`, `make test-fast` and `make lint` locally (NOT the full `make test` — CI runs it on every push). After pushing, wait with `gh run watch`; if CI is red, fix it before anything else. Docs-only commits: no local tests. Run a bigger simulation (`SIM_GAMES=1000+`) only when the task asks for it.

## Architecture rules (hard)
1. `internal/engine` is pure: no I/O, no goroutines, no `time.Now()`, no global `rand`, no logging. Time and randomness come in through commands/arguments.
2. Dependencies point inward only: `transport → room → engine`, `bot → room → engine`, `roles → engine`. The engine never imports room/transport/store.
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
| Bot | `Bot` (`internal/bot`) |
| Máy làm loa | `Speaker` |
| Đề cử / giàn / bào chữa / giết-tha | `Nominate` / `Trial` / `Defense` / `Verdict` |

- When the architecture changes (layers, dependency rules, request flow, config), update `docs/huong-dan-du-an.md` (Vietnamese project guide) in the same change.

## Don't
- Don't write frontend code. Don't touch files outside the task's scope.
- Don't add features from `docs/rules.md` §10 (out of scope).
- Don't read the whole repo; open only files the task needs.
- Don't commit secrets or generated binaries.

## Frontend (`web/`) và quy tắc giữ backend–frontend–giao diện đồng nhất

Repo này chứa cả server Go và app web (`web/`). Thứ tự nguồn sự thật, **cao thắng thấp**:

1. `docs/rules.md` — luật chơi.
2. `docs/api/` — hợp đồng server ↔ app (sinh từ code Go bằng `make api-docs`, kèm phần viết tay).
3. `docs/frontend/giao-dien-3d.md` — giao diện trông thế nào, pha → cảnh, hiệu ứng, âm thanh.
4. `web/prototype/` — bản chạy thử, chỉ để tham khảo hình ảnh; **không** là nguồn về dữ liệu hay luồng.

Tài liệu đi kèm: `docs/frontend/ke-hoach.md` (các bước), `docs/frontend/doi-chieu-backend.md` (chỗ bản chạy thử lệch backend + quy trình thêm tính năng), `docs/frontend/ban-giao-frontend.md` (màu, chữ, nút, tay), `docs/api/secrecy.md` (bắt buộc).

### Luật làm việc
- **Hợp đồng trước, code sau.** Tính năng cần dữ liệu/lệnh mới: sửa `docs/rules.md` (nếu đổi luật) → code Go + test → `make api-docs` → sửa `docs/api/screens/*.md` → `docs/frontend/giao-dien-3d.md` → `cd web && npm run gen:types` → code `web/`. Không làm ngược.
- **Frontend không bịa**: không dùng trường, lệnh, mã lỗi, pha nào không có trong `docs/api/schema/`. Thiếu thì dừng lại hỏi.
- **Không sửa tay file sinh ra**: `docs/api/schema/**`, `docs/api/examples/**`, các vùng `<!-- example -->` / `<!-- generated -->`.
- **Không tự đổi luật hay thiết kế** để code cho dễ. Thấy tài liệu mâu thuẫn hoặc thiếu: dừng, nêu rõ chỗ mâu thuẫn (file:dòng), đề xuất cách sửa, chờ trả lời.
- **Mỗi lần chỉ một việc** (một bước trong `ke-hoach.md` hoặc một tính năng). Không "tiện tay" sửa chỗ khác; ghi vào phần "Việc còn lại" trong báo cáo.
- **Lập kế hoạch trước khi sửa**: liệt kê file sẽ sửa, trường API dùng (trích file:dòng), cách kiểm tra; với thay đổi chạm cả Go lẫn `web/` thì chờ duyệt rồi mới làm.
- **Xong nghĩa là kiểm tra đã qua**: Go: `make test`; web: `cd web && npm run check`. Báo cáo kết quả lệnh thật, không viết "chắc là chạy được".
- Chữ hiển thị là tiếng Việt, lấy nguyên văn từ server khi server gửi (`say`, `notice`, `summary`, guide, câu lỗi trong `docs/api/errors.md`).
- Bí mật (`docs/api/secrecy.md`) là lỗi nặng ngang lộ dữ liệu: hình, tiếng, thời gian không được khác nhau giữa các máy theo vai/hành động khi chơi trực tiếp.

### Báo cáo cuối mỗi việc (luôn theo mẫu)
1. Đã làm (file chính).
2. Kiểm tra đã chạy + kết quả.
3. Chỗ còn lệch so với `docs/frontend/giao-dien-3d.md` hoặc `docs/api/`.
4. Việc còn lại / câu hỏi cần quyết.
