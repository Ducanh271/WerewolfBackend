package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/roles"
	"masoi/internal/room"
)

// The golden tests pin the wire protocol (docs/protocol.md) as it was before
// T30b moved the command table and the error codes into the roles: the command
// names, the args each one takes, and the code and text of every error. A
// refactor must leave them all passing without editing this file.

// goldenCommands is every command a client may send: a payload using every
// args key with a distinct value, and the engine command it becomes when sent
// by "p9".
var goldenCommands = []struct {
	name string
	args string
	want engine.Command
}{
	{"setSeats", `{"order":["p2","p1","p3"]}`, engine.SetSeats{Order: []engine.PlayerID{"p2", "p1", "p3"}}},
	{"configureRoles", `{"roles":{"Werewolf":2,"Seer":1}}`, engine.ConfigureRoles{Roles: map[engine.RoleID]int{"Werewolf": 2, "Seer": 1}}},
	{"configureTimers", `{"timers":{"roleView":1,"chiefElection":2,"nightStep":3,"phoneCheck":4,"loversMeet":5,"announce":6,"day":7,"nominationLock":8,"defense":9,"verdict":10,"hunterShot":11,"chiefSuccession":12}}`,
		engine.ConfigureTimers{Timers: engine.Timers{
			RoleView: 1 * time.Millisecond, ChiefElection: 2 * time.Millisecond, NightStep: 3 * time.Millisecond,
			PhoneCheck: 4 * time.Millisecond, LoversMeet: 5 * time.Millisecond, Announce: 6 * time.Millisecond,
			Day: 7 * time.Millisecond, NominationLock: 8 * time.Millisecond, Defense: 9 * time.Millisecond,
			Verdict: 10 * time.Millisecond, HunterShot: 11 * time.Millisecond, ChiefSuccession: 12 * time.Millisecond,
		}}},
	{"configureSound", `{"mode":"oneSpeaker"}`, engine.ConfigureSound{Mode: engine.OneSpeaker}},
	{"dealRoles", `{}`, engine.DealRoles{}},
	{"voteChief", `{"candidate":"p1"}`, engine.VoteChief{Voter: "p9", Candidate: "p1"}},
	{"nameSuccessor", `{"successor":"p1"}`, engine.NameSuccessor{By: "p9", Successor: "p1"}},
	{"nominate", `{"target":"p1"}`, engine.Nominate{Voter: "p9", Target: "p1"}},
	{"skipDay", `{"on":true}`, engine.SkipDay{Voter: "p9", On: true}},
	{"verdict", `{"kill":true}`, engine.CastVerdict{Voter: "p9", Kill: true}},
	{"abortVote", `{"on":true}`, engine.CastAbortVote{Voter: "p9", Yes: true}},
	{"newGame", `{}`, engine.NewGame{}},
	{"wolfVote", `{"target":"p1"}`, roles.WolfVote{By: "p9", Target: "p1"}},
	{"guardProtect", `{"target":"p1"}`, roles.GuardProtect{By: "p9", Target: "p1"}},
	{"visitorVisit", `{"target":"p1"}`, roles.VisitorVisit{By: "p9", Target: "p1"}},
	{"cupidPair", `{"a":"p1","b":"p2"}`, roles.CupidPair{By: "p9", A: "p1", B: "p2"}},
	{"hunterShoot", `{"target":"p1"}`, engine.HunterShoot{By: "p9", Target: "p1"}},
	{"witchHeal", `{}`, roles.WitchHeal{By: "p9"}},
	{"witchPoison", `{"target":"p1"}`, roles.WitchPoison{By: "p9", Target: "p1"}},
	{"cursedWolfCurse", `{"target":"p1"}`, roles.CursedWolfCurse{By: "p9", Target: "p1"}},
	{"wolfSeerCheck", `{"target":"p1"}`, roles.WolfSeerCheck{By: "p9", Target: "p1"}},
	{"detectiveCheck", `{"target":"p1"}`, roles.DetectiveCheck{By: "p9", Target: "p1"}},
	{"seerCheck", `{"target":"p1"}`, roles.SeerCheck{By: "p9", Target: "p1"}},
}

func TestGolden_CommandNames(t *testing.T) {
	var want []string
	for _, c := range goldenCommands {
		want = append(want, c.name)
	}
	got := wireCommandNames(t)
	sort.Strings(want)
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command names changed:\n got  %v\n want %v", got, want)
	}
}

func TestGolden_CommandPayloads(t *testing.T) {
	for _, c := range goldenCommands {
		t.Run(c.name, func(t *testing.T) {
			build, err := wireCommand(t, c.name, json.RawMessage(c.args))
			if err != nil {
				t.Fatalf("args %s refused: %v", c.args, err)
			}
			if got := build("p9"); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("args %s gave %#v, want %#v", c.args, got, c.want)
			}
			if _, err := wireCommand(t, c.name, json.RawMessage(`{"zz":1}`)); errCode(err) != codeBadMessage {
				t.Fatalf("unknown arg accepted or wrong code: %v", err)
			}
			if _, err := wireCommand(t, c.name, json.RawMessage(`{"By":"p1","by":"p1"}`)); errCode(err) != codeBadMessage {
				t.Fatalf("a sender in the args was accepted or wrong code: %v", err)
			}
		})
	}
}

// Commands without args accept no args at all, null, or {}.
func TestGolden_CommandsWithoutArgs(t *testing.T) {
	for _, name := range []string{"dealRoles", "newGame", "witchHeal"} {
		for _, raw := range []string{"", "null", "{}"} {
			if _, err := wireCommand(t, name, json.RawMessage(raw)); err != nil {
				t.Errorf("%s with args %q refused: %v", name, raw, err)
			}
		}
	}
}

// The refusals the transport makes itself while decoding a command.
func TestGolden_CommandRefusals(t *testing.T) {
	tests := []struct {
		name, args, code, message string
	}{
		{"nope", `{}`, "unknown_command", `unknown command "nope"`},
		{"wolfVote", `{}`, "bad_message", `wolfVote needs "target" (use "" for no bite)`},
		{"wolfVote", `{"target":null}`, "bad_message", `wolfVote needs "target" (use "" for no bite)`},
		{"configureSound", `{"mode":"megaphone"}`, "bad_message", `unknown sound mode "megaphone"`},
		{"configureTimers", `{"timers":{"day":-1}}`, "timer_out_of_range", "timer out of range: -1 ms"},
		{"configureTimers", `{"timers":{"week":1}}`, "bad_message", `unknown field "week"`},
		{"nominate", `[]`, "bad_message", "not a JSON object"},
		{"nominate", `{"Target":"p1"}`, "bad_message", `unknown field "Target"`},
		{"nominate", `{"target":1}`, "bad_message", `field "target" has wrong type`},
		{"guardProtect", `{"target":1}`, "bad_message", `field "target" has wrong type`},
		{"cupidPair", `{"a":"p1","b":2}`, "bad_message", `field "b" has wrong type`},
		{"wolfVote", `{"target":1}`, "bad_message", `field "target" has wrong type`},
		{"verdict", `{"kill":"yes"}`, "bad_message", `field "kill" has wrong type`},
		{"configureTimers", `{"timers":{"day":"x"}}`, "bad_message", `field "day" has wrong type`},
	}
	for _, tc := range tests {
		_, err := wireCommand(t, tc.name, json.RawMessage(tc.args))
		m := toErrorMsg(err)
		if err == nil || m.Code != tc.code || m.Message != tc.message {
			t.Errorf("%s %s: got %+v (err %v), want code %q message %q", tc.name, tc.args, m, err, tc.code, tc.message)
		}
	}
	// wolfVote with "" is the no-bite vote, not a refusal.
	if build, err := wireCommand(t, "wolfVote", json.RawMessage(`{"target":""}`)); err != nil ||
		!reflect.DeepEqual(build("p9"), roles.WolfVote{By: "p9", Target: engine.NoBite}) {
		t.Errorf("wolfVote with an empty target: %v", err)
	}
}

func errCode(err error) string {
	if err == nil {
		return ""
	}
	return toErrorMsg(err).Code
}

// goldenErrors is every error that may reach a client, with what it is told.
var goldenErrors = []struct {
	err           error
	code, message string
}{
	{engine.ErrUnknownCommand, "unknown_command", "unknown command"},
	{engine.ErrWrongPhase, "wrong_phase", "command not allowed in this phase"},
	{engine.ErrInvalidPlayerID, "invalid_player_id", "invalid player id"},
	{engine.ErrInvalidNickname, "invalid_nickname", "invalid nickname"},
	{engine.ErrNicknameTaken, "nickname_taken", "nickname already taken"},
	{engine.ErrAlreadyJoined, "already_joined", "player already joined"},
	{engine.ErrRoomFull, "room_full", "room is full"},
	{engine.ErrUnknownPlayer, "unknown_player", "unknown player"},
	{engine.ErrSeatsMismatch, "seats_mismatch", "seating must list every player exactly once"},
	{engine.ErrDuplicateSeat, "duplicate_seat", "player seated twice"},
	{engine.ErrTimerOutOfRange, "timer_out_of_range", "timer out of range"},
	{engine.ErrModeUnsupported, "mode_unsupported", "game mode not supported yet"},
	{engine.ErrInvalidSoundMode, "invalid_sound_mode", "invalid sound mode"},
	{engine.ErrRolesNotConfigured, "roles_not_configured", "roles are not configured"},
	{engine.ErrSeatsNotSet, "seats_not_set", "seats must be set for every player before dealing"},
	{engine.ErrAlreadyDealt, "already_dealt", "roles were already dealt"},
	{engine.ErrTooFewPlayers, "too_few_players", "too few players"},
	{engine.ErrTooManyPlayers, "too_many_players", "too many players"},
	{engine.ErrUnknownRole, "unknown_role", "unknown role"},
	{engine.ErrInvalidRoleCount, "invalid_role_count", "invalid role count"},
	{engine.ErrRoleMustBeUnique, "role_must_be_unique", "role can only be dealt to one player"},
	{engine.ErrRoleCountMismatch, "role_count_mismatch", "role slots must match the number of players"},
	{engine.ErrNoWolves, "no_wolves", "the game needs at least one wolf"},
	{engine.ErrTooManyWolves, "too_many_wolves", "too many wolves for this number of players"},
	{engine.ErrRoleNeedsMorePlayers, "role_needs_more_players", "role needs more players"},
	{engine.ErrRoleRequired, "role_required", "role list is missing a required role"},
	{engine.ErrNotYourTurn, "not_your_turn", "not your turn"},
	{engine.ErrNotAlive, "not_alive", "player is dead"},
	{engine.ErrInvalidTarget, "invalid_target", "invalid target"},
	{engine.ErrAlreadyActed, "already_acted", "already acted this night"},
	{engine.ErrBlocked, "blocked", "blocked tonight: no night ability"},
	{engine.ErrPotionUsed, "potion_used", "this potion is already used up"},
	{engine.ErrAbilityUsed, "ability_used", "this once-a-game ability is already used"},
	{engine.ErrNothingToHeal, "nothing_to_heal", "nobody is bitten tonight"},
	{engine.ErrSameGuardTarget, "same_guard_target", "the guard cannot protect the same player two nights in a row"},
	{engine.ErrSameVisitorTarget, "same_visitor_target", "the Visitor cannot visit the same player two nights in a row"},
	{engine.ErrSelfNomination, "self_nomination", "a player cannot nominate themselves"},
	{engine.ErrAlreadySpared, "already_spared", "this player was already spared today"},
	{engine.ErrOnTrial, "on_trial", "the player on trial cannot vote"},

	{room.ErrNotHost, "not_host", "only the host can do this"},
	{room.ErrForbidden, "forbidden", "command not allowed from a player"},
	{room.ErrDisconnected, "disconnected", "this connection is no longer the player's"},
	{room.ErrClosed, "room_closed", "room is closed"},
	{room.ErrBadCredentials, "bad_credentials", "bad credentials"},
	{room.ErrPINLocked, "pin_locked", "pin rejoin is locked"},
	{room.ErrRoomNotFound, "room_not_found", "room not found"},
	{room.ErrTooManyRooms, "too_many_rooms", "too many rooms"},
	{room.ErrInternal, "internal", "internal error"},
	{errors.New("something else"), "internal", "internal error"},
}

func TestGolden_ErrorCodes(t *testing.T) {
	for _, tc := range goldenErrors {
		want := map[string]any{"type": "error", "code": tc.code, "message": tc.message}
		if got := toErrorMsg(tc.err).json(); !reflect.DeepEqual(got, want) {
			t.Errorf("%v: got %v, want %v", tc.err, got, want)
		}
	}
}

// A wrapped error keeps its code and carries the detail in the message, except
// bad credentials, which never say more.
func TestGolden_WrappedErrors(t *testing.T) {
	tests := []struct {
		err  error
		want map[string]any
	}{
		{fmt.Errorf("%w: Villager needs at least 1", engine.ErrRoleRequired),
			map[string]any{"type": "error", "code": "role_required", "message": "role list is missing a required role: Villager needs at least 1"}},
		{fmt.Errorf("%w: heal", engine.ErrPotionUsed),
			map[string]any{"type": "error", "code": "potion_used", "message": "this potion is already used up: heal"}},
		{fmt.Errorf("%w: p3", engine.ErrSameGuardTarget),
			map[string]any{"type": "error", "code": "same_guard_target", "message": "the guard cannot protect the same player two nights in a row: p3"}},
		{fmt.Errorf("%w: who", room.ErrBadCredentials),
			map[string]any{"type": "error", "code": "bad_credentials", "message": "bad credentials"}},
		{room.PINLockedError{RetryAfter: 3 * time.Second},
			map[string]any{"type": "error", "code": "pin_locked", "message": "pin rejoin is locked", "retryAfter": int64(3000)}},
	}
	for _, tc := range tests {
		if got := toErrorMsg(tc.err).json(); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%v: got %v, want %v", tc.err, got, tc.want)
		}
	}
}

// The codes the transport raises before the room hears of a message.
func TestGolden_TransportCodes(t *testing.T) {
	got := []string{codeBadMessage, codeUnknownType, codeUnknownCommand, codeTooLarge, codeRateLimited,
		codeNotJoined, codeAlreadyJoined, codeInternal, codeSecretInURL, codeBadOrigin,
		codeTimerOutOfRange, codeTooManyConnections}
	want := []string{"bad_message", "unknown_type", "unknown_command", "message_too_large", "rate_limited",
		"not_joined", "already_joined", "internal", "secret_in_url", "bad_origin",
		"timer_out_of_range", "too_many_connections"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("transport codes %v, want %v", got, want)
	}
}
