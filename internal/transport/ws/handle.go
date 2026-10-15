package ws

import (
	"errors"

	"github.com/coder/websocket"

	"masoi/internal/room"
)

// handle processes one client message; false ends the connection. Logs name
// the room, the player and the kind of message, never its content: the
// messages carry secret codes and PINs.
func (s *session) handle(data []byte) bool {
	ok, flood := s.lim.allow(s.srv.cfg.Now())
	if !ok {
		if flood {
			s.srv.log.Warn("flood, closing", "room", s.roomCode(), "player", s.player)
			s.shut(websocket.StatusPolicyViolation, "too many messages")
			return false
		}
		s.reply(errorFrame(&protoError{codeRateLimited, "too many messages, slow down"}))
		return true
	}
	if len(data) > s.srv.cfg.MaxMessageBytes {
		s.reply(errorFrame(&protoError{codeTooLarge, "message too large"}))
		return true
	}
	typ, err := messageType(data)
	if err == nil {
		err = s.dispatch(typ, data)
	}
	if errors.Is(err, errLeft) {
		s.shut(websocket.StatusNormalClosure, "left")
		return false
	}
	if errors.Is(err, room.ErrDisconnected) {
		// Another phone took this player over; this one is out.
		s.reply(errorFrame(err))
		s.shut(websocket.StatusPolicyViolation, "replaced by another phone")
		return false
	}
	if err != nil {
		if isInternal(err) {
			s.srv.log.Error("internal error", "room", s.roomCode(), "player", s.player, "type", typ, "err", err)
		}
		s.reply(errorFrame(err))
	}
	return true
}

func (s *session) roomCode() string {
	if s.rm == nil {
		return ""
	}
	return s.rm.Code()
}

var errLeft = errors.New("left the room")

func (s *session) dispatch(typ string, data []byte) error {
	switch typ {
	case typeJoin:
		var m joinMsg
		if err := decodeStrict(data, &m); err != nil {
			return err
		}
		return s.join(m)
	case typeRejoin:
		var m rejoinMsg
		if err := decodeStrict(data, &m); err != nil {
			return err
		}
		return s.rejoin(m)
	case typeLeave:
		var m leaveMsg
		if err := decodeStrict(data, &m); err != nil {
			return err
		}
		if s.client == nil {
			return &protoError{codeNotJoined, "join a room first"}
		}
		if err := s.client.Leave(); err != nil {
			return err
		}
		return errLeft
	case typeCommand:
		var m commandMsg
		if err := decodeStrict(data, &m); err != nil {
			return err
		}
		if s.client == nil {
			return &protoError{codeNotJoined, "join a room first"}
		}
		build, err := buildCommand(m)
		if err != nil {
			return err
		}
		return s.client.Submit(room.PlayerCommand(build))
	}
	return unknownType(typ)
}

func (s *session) join(m joinMsg) error {
	if s.client != nil {
		return &protoError{codeAlreadyJoined, "this connection already joined"}
	}
	rm, err := s.srv.cfg.Rooms.Get(m.Room)
	if err != nil {
		return err
	}
	cl, creds, err := rm.Join(m.Nickname, s)
	if err != nil {
		return err
	}
	s.enter(rm, cl)
	s.open(welcome(cl.Player(), &creds))
	return nil
}

func (s *session) rejoin(m rejoinMsg) error {
	if s.client != nil {
		return &protoError{codeAlreadyJoined, "this connection already joined"}
	}
	byToken := m.Token != "" && m.Nickname == "" && m.PIN == ""
	byPIN := m.Token == "" && m.Nickname != "" && m.PIN != ""
	if !byToken && !byPIN {
		return badMessage("rejoin needs a token, or a nickname and a PIN")
	}
	rm, err := s.srv.cfg.Rooms.Get(m.Room)
	if err != nil {
		return err
	}
	var (
		cl    *room.Client
		creds *room.Credentials
	)
	if byToken {
		cl, err = rm.RejoinToken(m.Token, s)
	} else {
		var c room.Credentials
		cl, c, err = rm.RejoinPIN(m.Nickname, m.PIN, s)
		creds = &c
	}
	if err != nil {
		return err
	}
	s.enter(rm, cl)
	s.open(welcome(cl.Player(), creds))
	return nil
}

func (s *session) enter(rm *room.Room, cl *room.Client) {
	s.rm, s.client, s.player = rm, cl, cl.Player()
	go s.watchRoom(rm.Done())
	s.srv.log.Info("connected", "room", rm.Code(), "player", s.player)
}
