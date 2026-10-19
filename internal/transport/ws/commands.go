package ws

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"masoi/internal/engine"
)

// A commandSpec decodes the args of one command name.
type commandSpec func(args json.RawMessage) (func(by engine.PlayerID) engine.Command, error)

// spec makes a commandSpec from the shape of the args and a builder. Anything
// not in A (a "seed", a "by", a "voter") is refused by decodeStrict.
func spec[A any](build func(a A, by engine.PlayerID) engine.Command) commandSpec {
	return func(raw json.RawMessage) (func(engine.PlayerID) engine.Command, error) {
		var a A
		if err := decodeStrict(raw, &a); err != nil {
			return nil, err
		}
		return func(by engine.PlayerID) engine.Command { return build(a, by) }, nil
	}
}

type (
	noArgs    struct{}
	orderArgs struct {
		Order []engine.PlayerID `json:"order"`
	}
	rolesArgs struct {
		Roles map[engine.RoleID]int `json:"roles"`
	}
	timersArgs struct {
		Timers json.RawMessage `json:"timers"`
	}
	candidateAr struct {
		Candidate engine.PlayerID `json:"candidate"`
	}
	successorAr struct {
		Successor engine.PlayerID `json:"successor"`
	}
	targetArgs struct {
		Target engine.PlayerID `json:"target"`
	}
	onArgs struct {
		On bool `json:"on"`
	}
	soundModeArgs struct {
		Mode string `json:"mode"`
	}
	killArgs struct {
		Kill bool `json:"kill"`
	}
)

// timerMillis are the phase times in milliseconds, only the ones to change.
type timerMillis struct {
	RoleView        int64 `json:"roleView"`
	ChiefElection   int64 `json:"chiefElection"`
	NightStep       int64 `json:"nightStep"`
	PhoneCheck      int64 `json:"phoneCheck"`
	LoversMeet      int64 `json:"loversMeet"`
	Announce        int64 `json:"announce"`
	Day             int64 `json:"day"`
	NominationLock  int64 `json:"nominationLock"`
	Defense         int64 `json:"defense"`
	Verdict         int64 `json:"verdict"`
	HunterShot      int64 `json:"hunterShot"`
	ChiefSuccession int64 `json:"chiefSuccession"`
}

// maxTimerMillis keeps the conversion to a Duration far from overflowing; the
// engine's own limits are much tighter.
const maxTimerMillis = 24 * 60 * 60 * 1000

func (t timerMillis) timers() (engine.Timers, error) {
	d := func(ms int64) (time.Duration, error) {
		if ms < 0 || ms > maxTimerMillis {
			return 0, fmt.Errorf("%w: %d ms", engine.ErrTimerOutOfRange, ms)
		}
		return time.Duration(ms) * time.Millisecond, nil
	}
	var out engine.Timers
	for _, f := range []struct {
		ms  int64
		dst *time.Duration
	}{
		{t.RoleView, &out.RoleView}, {t.ChiefElection, &out.ChiefElection},
		{t.NightStep, &out.NightStep}, {t.PhoneCheck, &out.PhoneCheck}, {t.LoversMeet, &out.LoversMeet}, {t.Announce, &out.Announce}, {t.Day, &out.Day},
		{t.NominationLock, &out.NominationLock}, {t.Defense, &out.Defense},
		{t.Verdict, &out.Verdict}, {t.HunterShot, &out.HunterShot},
		{t.ChiefSuccession, &out.ChiefSuccession},
	} {
		v, err := d(f.ms)
		if err != nil {
			return engine.Timers{}, err
		}
		*f.dst = v
	}
	return out, nil
}

// coreCommands are the commands of the engine itself (docs/protocol.md
// "Commands"). The roles' own commands come from engine.Registry: see
// newCommandTable.
var coreCommands = map[string]commandSpec{
	"setSeats": spec(func(a orderArgs, _ engine.PlayerID) engine.Command {
		return engine.SetSeats{Order: a.Order}
	}),
	"configureRoles": spec(func(a rolesArgs, _ engine.PlayerID) engine.Command {
		return engine.ConfigureRoles{Roles: a.Roles}
	}),
	"configureTimers": configureTimers,
	"configureSound":  configureSound,
	"dealRoles": spec(func(_ noArgs, _ engine.PlayerID) engine.Command {
		return engine.DealRoles{} // the room draws the seed
	}),
	"voteChief": spec(func(a candidateAr, by engine.PlayerID) engine.Command {
		return engine.VoteChief{Voter: by, Candidate: a.Candidate}
	}),
	"nameSuccessor": spec(func(a successorAr, by engine.PlayerID) engine.Command {
		return engine.NameSuccessor{By: by, Successor: a.Successor}
	}),
	"nominate": spec(func(a targetArgs, by engine.PlayerID) engine.Command {
		return engine.Nominate{Voter: by, Target: a.Target}
	}),
	"skipDay": spec(func(a onArgs, by engine.PlayerID) engine.Command {
		return engine.SkipDay{Voter: by, On: a.On}
	}),
	"verdict": spec(func(a killArgs, by engine.PlayerID) engine.Command {
		return engine.CastVerdict{Voter: by, Kill: a.Kill}
	}),
	"abortVote": spec(func(a onArgs, by engine.PlayerID) engine.Command {
		return engine.CastAbortVote{Voter: by, Yes: a.On}
	}),
	"newGame": spec(func(_ noArgs, _ engine.PlayerID) engine.Command {
		return engine.NewGame{}
	}),
	"hunterShoot": spec(func(a targetArgs, by engine.PlayerID) engine.Command {
		return engine.HunterShoot{By: by, Target: a.Target}
	}),
}

// commandTable maps every command name a client may send to its decoder.
type commandTable map[string]commandSpec

// newCommandTable joins the core commands and the commands each registered
// role declares (engine.CommandRole). Two commands with one name are an
// error, so a server with a clash does not start.
func newCommandTable(registry map[engine.RoleID]engine.Role) (commandTable, error) {
	t := commandTable{}
	owner := map[string]string{}
	for name, s := range coreCommands {
		t[name] = s
		owner[name] = "the engine"
	}
	ids := make([]engine.RoleID, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		cr, ok := registry[id].(engine.CommandRole)
		if !ok {
			continue
		}
		for _, c := range cr.Commands() {
			if c.Name == "" || c.NewArgs == nil || c.Build == nil {
				return nil, fmt.Errorf("role %s declares an incomplete command %q", id, c.Name)
			}
			if prev, dup := owner[c.Name]; dup {
				return nil, fmt.Errorf("command %q is declared by both %s and role %s", c.Name, prev, id)
			}
			t[c.Name] = roleSpec(c)
			owner[c.Name] = "role " + string(id)
		}
	}
	return t, nil
}

// roleSpec makes a commandSpec from a role's declaration: the same strict
// decoding as the core commands, then the args' own check if they have one.
func roleSpec(c engine.CommandSpec) commandSpec {
	return func(raw json.RawMessage) (func(engine.PlayerID) engine.Command, error) {
		a := c.NewArgs()
		if err := decodeStrict(raw, a); err != nil {
			return nil, err
		}
		if v, ok := a.(engine.ArgsValidator); ok {
			if err := v.Validate(); err != nil {
				return nil, badMessage(err.Error())
			}
		}
		return func(by engine.PlayerID) engine.Command { return c.Build(a, by) }, nil
	}
}

// build turns a `command` message into a function that builds the engine
// command for the connection's player. The player never comes from the
// message.
func (t commandTable) build(m commandMsg) (func(by engine.PlayerID) engine.Command, error) {
	spec, ok := t[m.Name]
	if !ok {
		return nil, &protoError{codeUnknownCommand, fmt.Sprintf("unknown command %q", m.Name)}
	}
	args := m.Args
	if len(args) == 0 || string(args) == "null" {
		args = json.RawMessage("{}")
	}
	return spec(args)
}

// soundModes maps the protocol's string to the engine's SoundMode.
var soundModes = map[string]engine.SoundMode{
	"oneSpeaker": engine.OneSpeaker,
	"allPhones":  engine.AllPhonesSpeak,
}

func configureSound(raw json.RawMessage) (func(engine.PlayerID) engine.Command, error) {
	var a soundModeArgs
	if err := decodeStrict(raw, &a); err != nil {
		return nil, err
	}
	mode, ok := soundModes[a.Mode]
	if !ok {
		return nil, badMessage(fmt.Sprintf("unknown sound mode %q", a.Mode))
	}
	return func(engine.PlayerID) engine.Command { return engine.ConfigureSound{Mode: mode} }, nil
}

func configureTimers(raw json.RawMessage) (func(engine.PlayerID) engine.Command, error) {
	var a timersArgs
	if err := decodeStrict(raw, &a); err != nil {
		return nil, err
	}
	var ms timerMillis
	if len(a.Timers) > 0 {
		if err := decodeStrict(a.Timers, &ms); err != nil {
			return nil, err
		}
	}
	t, err := ms.timers()
	if err != nil {
		return nil, err
	}
	return func(engine.PlayerID) engine.Command { return engine.ConfigureTimers{Timers: t} }, nil
}
