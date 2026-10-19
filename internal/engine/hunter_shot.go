package engine

import (
	"fmt"
	"time"

	"masoi/internal/i18n"
)

// A dead Hunter is revealed at once, then everybody waits while he picks the
// player he takes with him (rules §6 "Thợ săn"). The shot is mandatory, but a
// Hunter who does not choose in time shoots nobody. The flow lives in the
// engine because it ends in the shared aftermath of deaths (afterDeaths).

// FlagRevealed is the Player.Flags key set when a role was revealed by a death.
const FlagRevealed = "revealed"

// DeathShooter is implemented by the role that is revealed when its holder
// dies, from any cause, and then shoots one living player.
type DeathShooter interface {
	ShootsOnDeath() bool
}

// HunterShoot is the dead Hunter's choice of a living player to take with him.
type HunterShoot struct {
	CommandMarker
	By     PlayerID
	Target PlayerID
}

// HunterRevealed makes the dead Hunter's role public.
type HunterRevealed struct {
	EventMarker
	Hunter PlayerID
}

// HunterShotStarted opens the pause in which the Hunter chooses. It is public,
// so it carries nothing but the Hunter and the time: the chain of deaths is in
// HunterChainNoted, which the living never see.
type HunterShotStarted struct {
	EventMarker
	Hunter   PlayerID
	Duration time.Duration
}

// HunterChainNoted comes right before HunterShotStarted and keeps what the
// shooting needs to go on: Queue, the other dead Hunters still to shoot, and
// Dead, everybody who died in this chain. Dead is in the order the pipeline
// killed them, which would tell bitten from poisoned (rules §4), so the event
// has no audience: only full views get it.
type HunterChainNoted struct {
	EventMarker
	Queue []PlayerID
	Dead  []PlayerID
}

// HunterShot is the Hunter's shot. The target dies right after it.
type HunterShot struct {
	EventMarker
	Hunter PlayerID
	Target PlayerID
}

// HunterTimedOut means the Hunter did not choose in time: nobody is shot.
type HunterTimedOut struct {
	EventMarker
	Hunter PlayerID
}

// Shot is the shooting in progress while PhaseHunterShot runs.
type Shot struct {
	Hunter PlayerID
	Queue  []PlayerID
	Dead   []PlayerID
	Resume Phase
}

func (s Shot) clone() Shot {
	s.Queue = append([]PlayerID(nil), s.Queue...)
	s.Dead = append([]PlayerID(nil), s.Dead...)
	return s
}

// AnnouncementKey implementations: all three are read aloud to everybody.
func (HunterRevealed) AnnouncementKey() string { return i18n.HunterRevealed }
func (HunterShot) AnnouncementKey() string     { return i18n.HunterShot }
func (HunterTimedOut) AnnouncementKey() string { return i18n.HunterTimedOut }

func evolveHunter(s State, e Event) (State, bool) {
	switch ev := e.(type) {
	case HunterRevealed:
		out := s.clone()
		if p := out.Players[ev.Hunter]; p != nil {
			SetFlag(p, FlagRevealed, 1)
		}
		return out, true
	case HunterChainNoted:
		out := s.clone()
		out.Shot.Queue = append([]PlayerID(nil), ev.Queue...)
		out.Shot.Dead = append([]PlayerID(nil), ev.Dead...)
		return out, true
	case HunterShotStarted:
		out := s.clone()
		out.Shot.Hunter, out.Shot.Resume = ev.Hunter, s.Phase
		out.Phase = PhaseHunterShot
		out.Seq++
		return out, true
	case HunterShot, HunterTimedOut:
		out := s.clone()
		out.Phase = s.Shot.Resume
		out.Shot = Shot{}
		out.Seq++
		return out, true
	}
	return s, false
}

func decideHunterShoot(s State, cmd HunterShoot) ([]Event, error) {
	if s.Phase != PhaseHunterShot {
		return nil, fmt.Errorf("%w: no Hunter is shooting", ErrWrongPhase)
	}
	if cmd.By != s.Shot.Hunter {
		return nil, fmt.Errorf("%w: only the dead Hunter shoots", ErrNotYourTurn)
	}
	if _, err := LivingTarget(s, cmd.Target); err != nil {
		return nil, err
	}
	return shotResolved(s, HunterShot{Hunter: cmd.By, Target: cmd.Target}, cmd.Target), nil
}

// shotResolved ends the shooting: the target (if any) dies at once, and the
// chain goes on with the next dead Hunter or with the win check and the Chief.
func shotResolved(s State, end Event, target PlayerID) []Event {
	evs := []Event{end}
	dead := append([]PlayerID(nil), s.Shot.Dead...)
	queue := append([]PlayerID(nil), s.Shot.Queue...)
	if target != "" {
		evs = append(evs, PlayerDied{Player: target})
		dead = append(dead, target)
		if isDeathShooter(s.Players[target]) {
			queue = append(queue, target)
		}
	}
	return append(evs, settle(s, evs, queue, dead)...)
}

func isDeathShooter(p *Player) bool {
	if p == nil {
		return false
	}
	d, ok := Registry[p.Role].(DeathShooter)
	return ok && d.ShootsOnDeath()
}
