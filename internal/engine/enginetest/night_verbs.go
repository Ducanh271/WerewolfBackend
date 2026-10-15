package enginetest

import (
	"masoi/internal/engine"
	"masoi/internal/roles"
)

// advanceTo lets the night run until it is the given role's turn, timing out
// the steps before it (nobody acted there). The night is written in call order
// (rules §3), so a role whose turn has passed, or who is not in the game, fails
// the scenario.
func (n *Night) advanceTo(role engine.RoleID) {
	n.g.t.Helper()
	for {
		cur, ok := engine.CurrentNightRole(n.g.state)
		if !ok {
			n.g.t.Fatalf("%s cannot act: the night is over, %s is not in the game, or you wrote the night out of call order (Guard, then Wolves, then Seer)", role, role)
		}
		if cur == role {
			return
		}
		n.g.Do(engine.PhaseTimeout{Phase: n.g.state.Phase, Seq: n.g.state.Seq})
	}
}

// Wolf is a wolf of the pack: Wolf("Binh").Bites("An"). Every wolf role votes
// here; a second call changes the vote.
func (n *Night) Wolf(name string) *WolfActor { return &WolfActor{n: n, id: n.Player(name)} }

// WolfActor is one wolf voting.
type WolfActor struct {
	n  *Night
	id engine.PlayerID
}

// Bites votes for the player to be bitten.
func (w *WolfActor) Bites(target string) *Night {
	w.n.g.t.Helper()
	w.n.advanceTo(roles.Werewolf)
	return w.n.Do(roles.WolfVote{By: w.id, Target: w.n.Player(target)})
}

// CannotBite votes and expects the engine to refuse with want.
func (w *WolfActor) CannotBite(target string, want error) *Night {
	w.n.g.t.Helper()
	w.n.advanceTo(roles.Werewolf)
	return w.n.DoExpectErr(roles.WolfVote{By: w.id, Target: w.n.Player(target)}, want)
}

// Guard is the guard: Guard("An").Protects("Binh").
func (n *Night) Guard(name string) *GuardActor { return &GuardActor{n: n, id: n.Player(name)} }

// GuardActor is the guard acting.
type GuardActor struct {
	n  *Night
	id engine.PlayerID
}

// Protects shields the player from the bite tonight.
func (a *GuardActor) Protects(target string) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.Guard)
	return a.n.Do(roles.GuardProtect{By: a.id, Target: a.n.Player(target)})
}

// CannotProtect shields and expects the engine to refuse with want.
func (a *GuardActor) CannotProtect(target string, want error) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.Guard)
	return a.n.DoExpectErr(roles.GuardProtect{By: a.id, Target: a.n.Player(target)}, want)
}

// Seer is the seer: Seer("An").Checks("Binh").
func (n *Night) Seer(name string) *SeerActor { return &SeerActor{n: n, id: n.Player(name)} }

// SeerActor is the seer acting.
type SeerActor struct {
	n  *Night
	id engine.PlayerID
}

// Checks looks at one player; read the answer with ExpectSeerSaw.
func (a *SeerActor) Checks(target string) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.Seer)
	return a.n.Do(roles.SeerCheck{By: a.id, Target: a.n.Player(target)})
}

// CannotCheck looks and expects the engine to refuse with want.
func (a *SeerActor) CannotCheck(target string, want error) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.Seer)
	return a.n.DoExpectErr(roles.SeerCheck{By: a.id, Target: a.n.Player(target)}, want)
}

// ExpectSeerSaw checks the Seer's private answer for a player: a wolf or not.
func (n *Night) ExpectSeerSaw(seer, target string, wolf bool) *Night {
	n.g.t.Helper()
	for _, e := range n.g.events {
		r, ok := e.(roles.SeerChecked)
		if ok && r.Seer == n.Player(seer) && r.Target == n.Player(target) {
			if r.IsWolf != wolf {
				n.g.t.Fatalf("expected %s to see %s as wolf=%v, but saw wolf=%v", seer, target, wolf, r.IsWolf)
			}
			return n
		}
	}
	n.g.t.Fatalf("expected %s to have checked %s, but no such check happened", seer, target)
	return n
}
