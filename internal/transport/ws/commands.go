package ws

import (
	"encoding/json"
	"fmt"
	"time"

	"masoi/internal/engine"
	"masoi/internal/roles"
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
	killArgs struct {
		Kill bool `json:"kill"`
	}
)

// timerMillis are the phase times in milliseconds, only the ones to change.
type timerMillis struct {
	RoleView        int64 `json:"roleView"`
	ChiefElection   int64 `json:"chiefElection"`
	NightStep       int64 `json:"nightStep"`
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
		{t.NightStep, &out.NightStep}, {t.Announce, &out.Announce}, {t.Day, &out.Day},
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

// commandTable is the table of docs/protocol.md "Commands". A new role adds
// its rows here.
var commandTable = map[string]commandSpec{
	"setSeats": spec(func(a orderArgs, _ engine.PlayerID) engine.Command {
		return engine.SetSeats{Order: a.Order}
	}),
	"configureRoles": spec(func(a rolesArgs, _ engine.PlayerID) engine.Command {
		return engine.ConfigureRoles{Roles: a.Roles}
	}),
	"configureTimers": configureTimers,
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
	"wolfVote": spec(func(a targetArgs, by engine.PlayerID) engine.Command {
		return roles.WolfVote{By: by, Target: a.Target}
	}),
	"guardProtect": spec(func(a targetArgs, by engine.PlayerID) engine.Command {
		return roles.GuardProtect{By: by, Target: a.Target}
	}),
	"visitorVisit": spec(func(a targetArgs, by engine.PlayerID) engine.Command {
		return roles.VisitorVisit{By: by, Target: a.Target}
	}),
	"hunterShoot": spec(func(a targetArgs, by engine.PlayerID) engine.Command {
		return engine.HunterShoot{By: by, Target: a.Target}
	}),
	"witchHeal": spec(func(_ noArgs, by engine.PlayerID) engine.Command {
		return roles.WitchHeal{By: by}
	}),
	"witchPoison": spec(func(a targetArgs, by engine.PlayerID) engine.Command {
		return roles.WitchPoison{By: by, Target: a.Target}
	}),
	"seerCheck": spec(func(a targetArgs, by engine.PlayerID) engine.Command {
		return roles.SeerCheck{By: by, Target: a.Target}
	}),
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
