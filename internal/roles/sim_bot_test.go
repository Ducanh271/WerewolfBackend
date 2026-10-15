package roles_test

import (
	"errors"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// nightBots is what each role's bot sends on its night step. A role that is
// registered but has no entry here (a new one) just times out, which the
// "many games" test must survive too; add its line here when its task lands.
var nightBots = map[engine.RoleID]func(g *simGame, by engine.PlayerID) []engine.Command{
	roles.Guard: func(g *simGame, by engine.PlayerID) []engine.Command {
		return []engine.Command{roles.GuardProtect{By: by, Target: g.anyLiving()}}
	},
	roles.Seer: func(g *simGame, by engine.PlayerID) []engine.Command {
		return []engine.Command{roles.SeerCheck{By: by, Target: g.anyLiving()}}
	},
	roles.Detective: func(g *simGame, by engine.PlayerID) []engine.Command {
		return []engine.Command{roles.DetectiveCheck{By: by, Target: g.anyLiving()}}
	},
	roles.WolfSeer: func(g *simGame, by engine.PlayerID) []engine.Command {
		return []engine.Command{roles.WolfSeerCheck{By: by, Target: g.anyLiving()}}
	},
	roles.CursedWolf: func(g *simGame, by engine.PlayerID) []engine.Command {
		return []engine.Command{roles.CursedWolfCurse{By: by, Target: g.anyLiving()}}
	},
	roles.Visitor: func(g *simGame, by engine.PlayerID) []engine.Command {
		return []engine.Command{roles.VisitorVisit{By: by, Target: g.anyLiving()}}
	},
	roles.Witch: func(g *simGame, by engine.PlayerID) []engine.Command {
		return []engine.Command{roles.WitchHeal{By: by}, roles.WitchPoison{By: by, Target: g.anyLiving()}}
	},
}

// playStep lets the bots act in the current phase, once. It does not end the
// phase: the caller times it out afterwards.
func (g *simGame) playStep() {
	switch g.s.Phase {
	case engine.PhaseChiefElection:
		for _, id := range g.living() {
			if g.rng.IntN(10) < 7 {
				_ = g.do(engine.VoteChief{Voter: id, Candidate: g.anyLiving()})
			}
		}
	case engine.PhaseNight:
		g.playNight()
	case engine.PhaseDay:
		g.playDay()
	case engine.PhaseVerdict:
		for _, id := range g.living() {
			_ = g.do(engine.CastVerdict{Voter: id, Kill: g.rng.IntN(10) < 7})
		}
	case engine.PhaseHunterShot:
		if g.rng.IntN(10) < 7 {
			_ = g.do(engine.HunterShoot{By: g.s.Shot.Hunter, Target: g.anyLiving()})
		}
	case engine.PhaseChiefSuccession:
		if g.rng.IntN(10) < 7 {
			_ = g.do(engine.NameSuccessor{By: g.s.Succession.Former, Successor: g.anyLiving()})
		}
	}
	g.garbage()
}

func (g *simGame) playNight() {
	step, ok := engine.CurrentNightRole(g.s)
	if !ok {
		return
	}
	if step == roles.Werewolf { // the pack step: every wolf bites, mostly the same player
		target := g.anyLiving()
		if g.rng.IntN(6) == 0 {
			target = engine.NoBite // the whole pack chooses "Không cắn ai"
		}
		for _, id := range g.living() {
			if p := g.s.Players[id]; p.Team == engine.TeamWolves {
				t := target
				if g.rng.IntN(4) == 0 {
					t = g.anyLiving()
				}
				if g.rng.IntN(12) == 0 {
					t = engine.NoBite
				}
				_ = g.do(roles.WolfVote{By: id, Target: t})
			}
		}
		return
	}
	bot, known := nightBots[step]
	if !known || g.rng.IntN(5) == 0 { // sometimes the holder does not choose in time
		return
	}
	for _, id := range g.ids() {
		if g.s.Players[id].Role == step {
			for _, c := range bot(g, id) {
				_ = g.do(c)
			}
		}
	}
}

func (g *simGame) playDay() {
	for _, id := range g.living() {
		switch g.rng.IntN(10) {
		case 0:
			_ = g.do(engine.SkipDay{Voter: id, On: true})
		case 1, 2, 3, 4, 5, 6:
			_ = g.do(engine.Nominate{Voter: id, Target: g.anyLiving()})
		}
	}
}

// garbage sends one random command from a random player, living or dead, in a
// phase where it may not belong. The engine must refuse it with a typed error
// or accept it; never panic.
func (g *simGame) garbage() {
	ids := g.ids()
	by, target := ids[g.rng.IntN(len(ids))], ids[g.rng.IntN(len(ids))]
	cmds := []engine.Command{
		engine.VoteChief{Voter: by, Candidate: target},
		engine.Nominate{Voter: by, Target: target},
		engine.CastVerdict{Voter: by, Kill: true},
		engine.HunterShoot{By: by, Target: target},
		engine.NameSuccessor{By: by, Successor: target},
		roles.WolfVote{By: by, Target: target},
		roles.GuardProtect{By: by, Target: target},
		roles.SeerCheck{By: by, Target: target},
		roles.WolfSeerCheck{By: by, Target: target},
		roles.CursedWolfCurse{By: by, Target: target},
		roles.VisitorVisit{By: by, Target: target},
		roles.WitchHeal{By: by},
		roles.WitchPoison{By: by, Target: target},
		roles.WolfSeerCheck{By: by, Target: "nobody"},
	}
	_ = g.do(cmds[g.rng.IntN(len(cmds))])
}

func (g *simGame) ids() []engine.PlayerID { return g.s.Seats }

func (g *simGame) living() []engine.PlayerID {
	var out []engine.PlayerID
	for _, id := range g.s.Seats {
		if g.s.Players[id].Alive {
			out = append(out, id)
		}
	}
	return out
}

// anyLiving is a random living player (sometimes a dead one on purpose: the
// engine must refuse it).
func (g *simGame) anyLiving() engine.PlayerID {
	if g.rng.IntN(25) == 0 {
		ids := g.ids()
		return ids[g.rng.IntN(len(ids))]
	}
	l := g.living()
	if len(l) == 0 {
		return ""
	}
	return l[g.rng.IntN(len(l))]
}

// knownErrors are the typed errors of the engine. An error that wraps none of
// them is an internal error. A new typed error in engine/errors.go goes here.
var knownErrors = []error{
	engine.ErrUnknownCommand, engine.ErrWrongPhase, engine.ErrInvalidPlayerID, engine.ErrInvalidNickname,
	engine.ErrNicknameTaken, engine.ErrAlreadyJoined, engine.ErrRoomFull, engine.ErrUnknownPlayer,
	engine.ErrSeatsMismatch, engine.ErrDuplicateSeat, engine.ErrTimerOutOfRange, engine.ErrModeUnsupported,
	engine.ErrRolesNotConfigured, engine.ErrSeatsNotSet, engine.ErrAlreadyDealt, engine.ErrTooFewPlayers,
	engine.ErrTooManyPlayers, engine.ErrUnknownRole, engine.ErrInvalidRoleCount, engine.ErrRoleMustBeUnique,
	engine.ErrRoleCountMismatch, engine.ErrNoWolves, engine.ErrTooManyWolves, engine.ErrRoleNeedsMorePlayers,
	engine.ErrNotYourTurn, engine.ErrNotAlive, engine.ErrInvalidTarget, engine.ErrAlreadyActed, engine.ErrBlocked,
	engine.ErrPotionUsed, engine.ErrAbilityUsed, engine.ErrNothingToHeal, engine.ErrSameGuardTarget, engine.ErrSameVisitorTarget,
	engine.ErrSelfNomination, engine.ErrAlreadySpared, engine.ErrOnTrial,
}

func isTypedError(err error) bool {
	for _, known := range knownErrors {
		if errors.Is(err, known) {
			return true
		}
	}
	return false
}
