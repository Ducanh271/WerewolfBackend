package ws

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"masoi/internal/i18n"
)

// Groups of docs/api/errors.md. The first three are errors a player can hit
// at the table with a client that follows the docs, each with its own text;
// groupBug is the rest, which share i18n.ErrorGeneric.
const (
	groupJoin  = "Joining and rejoining"
	groupLobby = "Host setting up the game (lobby)"
	groupGame  = "During the game"
	groupBug   = "Only a buggy client or a server bug"
)

var errorGroups = []string{groupJoin, groupLobby, groupGame, groupBug}

// errorRow is one line of the tables: a wire code, the situation that picks
// its second text (i18n.Situation*), and when the server sends it.
type errorRow struct {
	code, situation, group, when string
}

// errorRows is every code the server sends. TestAPIDocs_ErrorTable checks it
// against the source and the i18n texts.
var errorRows = []errorRow{
	{"room_not_found", "", groupJoin, "`join`, `rejoin` or `GET /join/{code}` with a code that does not exist or has expired (HTTP 404)"},
	{"room_closed", "", groupJoin, "the room is shutting down while the phone joins or acts; `POST /rooms` while the server shuts down (HTTP 503)"},
	{"invalid_nickname", "", groupJoin, "`join` with an empty name, over 20 characters, or with symbols or emoji (rules §2 \"Nickname, rời phòng và đá người\")"},
	{"nickname_taken", "", groupJoin, "`join` with a name already in the room (case and accents normalised)"},
	{"room_full", "", groupJoin, "`join` when the room already has 30 players"},
	{"wrong_phase", i18n.SituationJoin, groupJoin, "`join` once the roles are dealt: a new player cannot enter a running or ended game"},
	{"bad_credentials", "", groupJoin, "`rejoin` with nickname + PIN: wrong PIN or unknown nickname"},
	{"bad_credentials", i18n.SituationToken, groupJoin, "`rejoin` with the saved token, which no longer works: another phone took the player over with the PIN, or the host kicked them. Show the join screen"},
	{"pin_locked", "", groupJoin, "5 wrong PINs for that nickname (rules §2 \"Vào lại phòng\"). `%s` is the seconds left: `retryAfter` / 1000, rounded up"},
	{"kicked_by_host", "", groupJoin, "the host kicked this player from the lobby; the socket closes. Do not reconnect"},
	{"replaced_by_other_device", "", groupJoin, "the player joined from another phone or tab (by token or by PIN); the socket closes. Do not reconnect"},
	{"too_many_rooms", "", groupJoin, "`POST /rooms` when the server holds its maximum number of rooms (HTTP 503)"},
	{"too_many_connections", "", groupJoin, "`GET /ws` refused: the server or this network has too many open sockets (HTTP 503)"},
	{"rate_limited", "", groupJoin, "over 10 messages a second; the message was ignored. Staying over the limit for 3 seconds closes the socket"},

	{"not_host", "", groupLobby, "a host-only message from a player who is not, or no longer, the host (e.g. right after `transferHost`)"},
	{"unknown_player", "", groupLobby, "`kick` or `transferHost` naming someone who just left the room"},
	{"seats_mismatch", "", groupLobby, "`setSeats` whose list no longer matches the room: someone joined or left meanwhile"},
	{"seats_not_set", "", groupLobby, "`dealRoles` while someone has no seat (joined after `setSeats`)"},
	{"roles_not_configured", "", groupLobby, "`dealRoles` before any `configureRoles`"},
	{"timer_out_of_range", "", groupLobby, "`configureTimers` with a time outside its range (rules §2 \"Thời gian mặc định\")"},
	{"too_few_players", "", groupLobby, "`configureRoles` or `dealRoles` with fewer than 5 players"},
	{"role_count_mismatch", "", groupLobby, "`configureRoles` or `dealRoles` when the number of roles is not the number of players (also after a `leave` or `kick`)"},
	{"no_wolves", "", groupLobby, "the role list has no wolf role"},
	{"too_many_wolves", "", groupLobby, "wolves over 30% of the players, the one to be cursed included (rules §2 \"Giới hạn số sói\")"},
	{"role_needs_more_players", "", groupLobby, "Detective with 8 players or fewer, Cupid with fewer than 10 (rules §2 \"Vai theo số người\")"},
	{"role_required", "", groupLobby, "the role list has no Villager (rules §2 \"Số lượng mỗi vai\")"},
	{"already_dealt", "", groupLobby, "`dealRoles` a second time (double tap)"},

	{"wrong_phase", "", groupGame, "a command the current phase does not take, usually a tap that lands just after the phase or step ended"},
	{"not_your_turn", "", groupGame, "a night command outside the sender's own step, usually a tap just after it ended. Every player who does not hold the step gets exactly this, whatever their role or state ([secrecy](secrecy.md))"},
	{"blocked", "", groupGame, "a night command while `me.turn.blocked` (rules §6 \"Nguyệt Nữ\"); the text is `me.turn.notice`'s and never says who"},
	{"already_acted", "", groupGame, "a second choice on the same step: every night choice is final except the pack's vote (double tap)"},
	{"ability_used", "", groupGame, "`detectiveCheck` a second time in the game (rules §6 \"Thám tử\")"},
	{"potion_used", "", groupGame, "`witchHeal` or `witchPoison` with that potion already used (rules §6 \"Phù thủy\")"},
	{"nothing_to_heal", "", groupGame, "`witchHeal` when nobody is bitten tonight"},
	{"same_guard_target", "", groupGame, "`guardProtect` on last night's player (rules §6 \"Bảo vệ\")"},
	{"same_visitor_target", "", groupGame, "`visitorVisit` on last night's player (rules §6 \"Nguyệt Nữ\")"},
	{"invalid_target", "", groupGame, "a target the role may not pick: a dead player, a wolf for the curse, oneself where not allowed, the same player twice for Cupid"},

	{"bad_message", "", groupBug, "malformed JSON, an unknown or wrongly typed field, wrong capitals, a `seed`"},
	{"unknown_type", "", groupBug, "a message `type` the server does not know"},
	{"unknown_command", "", groupBug, "a command `name` the server does not know"},
	{"message_too_large", "", groupBug, "a message over 4 KB"},
	{"not_joined", "", groupBug, "any message before `join` or `rejoin` succeeded"},
	{"already_joined", "", groupBug, "`join` or `rejoin` a second time on the same socket"},
	{"disconnected", "", groupBug, "a message that still reaches the room from a socket another phone replaced (it was told `replaced_by_other_device` first)"},
	{"secret_in_url", "", groupBug, "`GET /ws` with a query string (HTTP 400)"},
	{"bad_origin", "", groupBug, "the page's origin is not configured on the server (HTTP 403)"},
	{"forbidden", "", groupBug, "`kick` or `transferHost` naming oneself"},
	{"invalid_player_id", "", groupBug, "never from the wire: the server makes player ids"},
	{"duplicate_seat", "", groupBug, "`setSeats` listing a player twice"},
	{"mode_unsupported", "", groupBug, "`POST /rooms` with `mode: \"online\"` (HTTP 400)"},
	{"invalid_sound_mode", "", groupBug, "never from the wire: an unknown sound mode is `bad_message` first"},
	{"unknown_role", "", groupBug, "`configureRoles` with a role that does not exist"},
	{"invalid_role_count", "", groupBug, "`configureRoles` with a count below 1"},
	{"role_must_be_unique", "", groupBug, "`configureRoles` with more than one of a role allowed once (rules §2 \"Số lượng mỗi vai\"); cap it in the picker"},
	{"too_many_players", "", groupBug, "never in practice: a room is full at 30"},
	{"self_nomination", "", groupBug, "`nominate` oneself"},
	{"already_spared", "", groupBug, "`nominate` someone in `votes.spared`"},
	{"on_trial", "", groupBug, "`verdict` from `votes.defendant`"},
	{"not_alive", "", groupBug, "a vote or night command from a dead player"},
	{"internal", "", groupBug, "a bug in the server; the message changed nothing"},
}

func (r errorRow) key() string {
	if r.situation == "" {
		return r.code
	}
	return r.code + "." + r.situation
}

// errorTables renders the generated part of errors.md.
func errorTables() string {
	var b strings.Builder
	for _, g := range errorGroups {
		b.WriteString("### " + g + "\n\n")
		if g == groupBug {
			text, _ := i18n.ErrorTemplate("", "")
			b.WriteString("All of these, and any code the app does not know, show:\n\n> " + text + "\n\n")
			b.WriteString("| code | when |\n|---|---|\n")
		} else {
			b.WriteString("| code | situation | when | text |\n|---|---|---|---|\n")
		}
		for _, r := range errorRows {
			if r.group != g {
				continue
			}
			if g == groupBug {
				b.WriteString("| `" + r.code + "` | " + r.when + " |\n")
				continue
			}
			text, _ := i18n.ErrorTemplate(r.code, r.situation)
			situation := ""
			if r.situation != "" {
				situation = "`" + r.situation + "`"
			}
			b.WriteString("| `" + r.code + "` | " + situation + " | " + r.when + " | " + text + " |\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// The error tables name every code the server can send, once per situation;
// a player's error has its own text and a bug's has none.
func TestAPIDocs_ErrorTable(t *testing.T) {
	source, err := serverErrorCodes()
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	for _, k := range i18n.ErrorKeys() {
		keys[k] = true
	}
	var listed []string
	seen := map[string]bool{}
	for _, r := range errorRows {
		if seen[r.key()] {
			t.Errorf("%s listed twice", r.key())
		}
		seen[r.key()] = true
		listed = append(listed, r.code)
		own := keys[r.key()]
		switch {
		case r.group == groupBug && own:
			t.Errorf("%s is a bug's error but has a text of its own", r.key())
		case r.group != groupBug && !own:
			t.Errorf("%s has no text in internal/i18n", r.key())
		}
		delete(keys, r.key())
	}
	for k := range keys {
		t.Errorf("i18n has a text for %s, which no row of errorRows uses", k)
	}
	listed = uniqueSorted(listed)
	if !reflect.DeepEqual(listed, source) {
		t.Errorf("errorRows codes differ from the server's:\n rows   %v\n source %v", listed, source)
	}
}

func uniqueSorted(s []string) []string {
	sort.Strings(s)
	out := s[:0]
	for i, v := range s {
		if i == 0 || v != s[i-1] {
			out = append(out, v)
		}
	}
	return out
}

// serverErrorCodes reads every code in the non-test source: engine.NewError
// calls, the transport's code* constants and the room's eviction reasons.
func serverErrorCodes() ([]string, error) {
	dirs := []string{"."}
	for _, d := range sourceDirs {
		dirs = append(dirs, d)
	}
	var codes []string
	for _, dir := range dirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			return nil, err
		}
		for _, path := range files {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			f, err := parser.ParseFile(token.NewFileSet(), path, src, 0)
			if err != nil {
				return nil, err
			}
			codes = append(codes, codesIn(f)...)
		}
	}
	return uniqueSorted(codes), nil
}

func codesIn(f *ast.File) []string {
	var codes []string
	str := func(e ast.Expr) {
		if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if s, err := strconv.Unquote(lit.Value); err == nil {
				codes = append(codes, s)
			}
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if name := callName(x.Fun); name == "NewError" && len(x.Args) > 0 {
				str(x.Args[0])
			}
		case *ast.CompositeLit:
			if id, ok := x.Type.(*ast.Ident); ok && id.Name == "disconnectReason" && len(x.Elts) > 0 {
				str(x.Elts[0])
			}
		case *ast.ValueSpec:
			for i, name := range x.Names {
				if strings.HasPrefix(name.Name, "code") && i < len(x.Values) {
					str(x.Values[i])
				}
			}
		}
		return true
	})
	return codes
}

func callName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return x.Sel.Name
	}
	return ""
}
