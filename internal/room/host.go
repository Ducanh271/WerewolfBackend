package room

import (
	"time"

	"masoi/internal/engine"
)

// hostDisconnectGrace is how long the host may stay disconnected, while the
// title can lapse (hostMayLapse), without losing the host's powers (rules §2
// "Chủ phòng"): a few seconds of lost wifi changes nothing. Only pressing
// leave (Client.Leave) transfers them sooner.
const hostDisconnectGrace = 60 * time.Second

// hostMayLapse reports whether losing the connection can cost the host the
// title now: before the deal, and again once the game has ended (rules §2
// "Chủ phòng"). In between, a dead phone keeps it.
func (r *Room) hostMayLapse() bool {
	return !r.state.Dealt || r.state.Phase == engine.PhaseEnded
}

// watchHost keeps the grace timer in step with the room. While the host is
// disconnected and the title may lapse, it is armed for what is left of
// hostDisconnectGrace counted from when the host's phone dropped, so a host
// lost mid-game for longer than that loses the title as soon as the game
// ends. Otherwise it is stopped.
func (r *Room) watchHost() {
	if r.host == "" || r.connected(r.host) || !r.hostMayLapse() {
		r.stopHostGrace()
		return
	}
	if r.hostGraceT != nil {
		return
	}
	id := r.host
	left := max(hostDisconnectGrace-r.clock.Now().Sub(r.offSince[id]), 0)
	r.hostGraceT = r.clock.AfterFunc(left, func() {
		r.post(func() {
			r.hostGraceT = nil
			if r.host == id && !r.connected(id) && r.hostMayLapse() {
				r.passHost(id)
				r.changed()
			}
		})
	})
}

func (r *Room) stopHostGrace() {
	if r.hostGraceT != nil {
		r.hostGraceT.Stop()
		r.hostGraceT = nil
	}
}

// passHost gives the host's powers to the next player (in join order) who is
// connected, or to nobody: then the first player to reconnect takes them.
func (r *Room) passHost(from engine.PlayerID) {
	r.host = ""
	for _, p := range r.joined {
		if p != from && r.connected(p) {
			r.host = p
			return
		}
	}
}
