package room

import (
	"fmt"
	"time"

	"masoi/internal/engine"
)

// hostDisconnectGrace is how long the host may be disconnected, before the
// deal, without losing the host's powers (rules §2): a few seconds of lost
// wifi changes nothing. Only pressing leave (Client.Leave) transfers them
// sooner.
const hostDisconnectGrace = 60 * time.Second

// Client is one phone's handle on the room. Everything it does is done as the
// player it was admitted as: the room, not the caller, says who that is.
type Client struct {
	room   *Room
	player engine.PlayerID
	conn   *connection
}

// Player is who this phone is.
func (c *Client) Player() engine.PlayerID { return c.player }

// live reports whether this handle is still the player's connection. A phone
// that was replaced by a rejoin, or disconnected, is not.
func (c *Client) live() bool { return c.room.conns[c.player] == c.conn }

// Join adds a player to the lobby. The first player is the host (the one who
// created the room). The result carries the secret token and the PIN.
func (r *Room) Join(nickname string, sink Sink) (*Client, Credentials, error) {
	var cl *Client
	var creds Credentials
	err := r.do(func() error {
		id := engine.PlayerID(fmt.Sprintf("p%d", r.nextID+1))
		if err := r.apply(engine.JoinRoom{Player: id, Nickname: nickname}); err != nil {
			return err
		}
		r.nextID++
		r.joined = append(r.joined, id)
		creds = r.ident.Issue(id, nickname)
		if r.host == "" {
			r.host = id
		}
		cl = r.attach(id, sink)
		r.changed()
		return nil
	})
	return cl, creds, err
}

// RejoinToken is the same phone coming back with its secret code.
func (r *Room) RejoinToken(token string, sink Sink) (*Client, error) {
	var cl *Client
	err := r.do(func() error {
		id, err := r.ident.ByToken(token)
		if err != nil {
			return err
		}
		cl = r.reattach(id, sink, ReasonAnotherDevice)
		return nil
	})
	return cl, err
}

// RejoinPIN is another phone coming in with nickname + PIN. It gets a new
// token; the old one stops working.
func (r *Room) RejoinPIN(nickname, pin string, sink Sink) (*Client, Credentials, error) {
	var cl *Client
	var creds Credentials
	err := r.do(func() error {
		id, token, err := r.ident.ByPIN(nickname, pin)
		if err != nil {
			return err
		}
		creds = Credentials{Token: token}
		cl = r.reattach(id, sink, ReasonPINTakenOver)
		return nil
	})
	return cl, creds, err
}

func (r *Room) attach(id engine.PlayerID, sink Sink) *Client {
	conn := &connection{player: id, sink: sink, fresh: true}
	r.conns[id] = conn
	return &Client{room: r, player: id, conn: conn}
}

// reattach replaces any earlier connection of the player: one phone at a
// time. A connection still open at this player's old seat is told why and
// closed (rules §2); a connection already dropped (Disconnect ran first) has
// nothing left to evict.
func (r *Room) reattach(id engine.PlayerID, sink Sink, reason error) *Client {
	if old := r.conns[id]; old != nil {
		old.sink.Evict(reason)
	}
	cl := r.attach(id, sink)
	if id == r.host {
		r.stopHostGrace() // the host is back: no transfer was due anyway
	}
	if r.host == "" && !r.state.Dealt {
		r.host = id
	}
	r.changed()
	return cl
}

// Disconnect drops the phone. The player keeps their seat and role. Before
// the deal, if the host drops, their powers only pass once they have stayed
// disconnected past hostDisconnectGrace (rules §2); reconnecting before then
// keeps them.
func (c *Client) Disconnect() {
	r := c.room
	_ = r.do(func() error {
		if !c.live() {
			return nil
		}
		delete(r.conns, c.player)
		if c.player == r.host && !r.state.Dealt {
			r.armHostGrace(c.player)
		}
		r.changed()
		return nil
	})
}

// armHostGrace schedules the host's powers to pass to the next connected
// player if id, the host, stays disconnected for hostDisconnectGrace.
func (r *Room) armHostGrace(id engine.PlayerID) {
	r.stopHostGrace()
	r.hostGraceT = r.clock.AfterFunc(hostDisconnectGrace, func() {
		r.post(func() {
			if r.host == id && !r.connected(id) && !r.state.Dealt {
				r.passHost(id)
				r.changed()
			}
			r.hostGraceT = nil
		})
	})
}

func (r *Room) stopHostGrace() {
	if r.hostGraceT != nil {
		r.hostGraceT.Stop()
		r.hostGraceT = nil
	}
}

// Leave takes the player out of the room for good. Only before the deal.
func (c *Client) Leave() error {
	r := c.room
	return r.do(func() error {
		if !c.live() {
			return ErrDisconnected
		}
		if err := r.apply(engine.LeaveRoom{Player: c.player}); err != nil {
			return err
		}
		delete(r.conns, c.player)
		r.ident.Revoke(c.player)
		for i, p := range r.joined {
			if p == c.player {
				r.joined = append(r.joined[:i], r.joined[i+1:]...)
				break
			}
		}
		if c.player == r.host {
			r.passHost(c.player)
		}
		r.changed()
		return nil
	})
}

// passHost gives the host's powers to the next player (in join order) who is
// connected.
func (r *Room) passHost(from engine.PlayerID) {
	r.host = ""
	for _, p := range r.joined {
		if p != from && r.connected(p) {
			r.host = p
			return
		}
	}
}
