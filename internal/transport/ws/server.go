package ws

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/coder/websocket"

	"masoi/internal/engine"
	"masoi/internal/room"
)

// Server is the HTTP side of the game: creating rooms and the WebSocket.
type Server struct {
	cfg   Config
	log   *slog.Logger
	mux   *http.ServeMux
	conns *connGuard
}

// New returns the handler serving:
//
//	POST /rooms        create a room, reply {"room": "K7QP"}
//	GET  /join/{code}  does the room exist (the link behind the QR code)
//	GET  /ws           the game socket; everything, secrets included, goes in messages
func New(cfg Config) http.Handler {
	cfg.setDefaults()
	s := &Server{
		cfg: cfg, log: cfg.Logger, mux: http.NewServeMux(),
		conns: newConnGuard(cfg.MaxConnections, cfg.MaxConnectionsPerIP, cfg.NewConnPerMinute, cfg.NewConnBurst, cfg.Now),
	}
	s.mux.HandleFunc("POST /rooms", s.createRoom)
	s.mux.HandleFunc("GET /join/{code}", s.checkRoom)
	s.mux.HandleFunc("GET /ws", s.serveWS)
	return s.mux
}

func (s *Server) createRoom(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Mode string `json:"mode"`
	}
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<10))
	if err != nil {
		writeJSON(w, http.StatusRequestEntityTooLarge, errorFrameMap(&protoError{codeTooLarge, "body too large"}))
		return
	}
	mode := engine.InPerson
	if len(strings.TrimSpace(string(raw))) > 0 {
		if err := decodeStrict(raw, &body); err != nil {
			writeJSON(w, http.StatusBadRequest, errorFrameMap(err))
			return
		}
		switch body.Mode {
		case "", "inPerson":
		case "online":
			mode = engine.Online
		default:
			writeJSON(w, http.StatusBadRequest, errorFrameMap(badMessage("unknown mode")))
			return
		}
	}
	rm, err := s.cfg.Rooms.Create(mode)
	if err != nil {
		writeJSON(w, statusOf(err), errorFrameMap(err))
		return
	}
	s.log.Info("room created", "room", rm.Code())
	writeJSON(w, http.StatusCreated, map[string]any{"room": rm.Code()})
}

func (s *Server) checkRoom(w http.ResponseWriter, r *http.Request) {
	rm, err := s.cfg.Rooms.Get(r.PathValue("code"))
	if err != nil {
		writeJSON(w, statusOf(err), errorFrameMap(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"room": rm.Code()})
}

func (s *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		// Secrets travel in messages, never in the URL, which ends up in logs.
		writeJSON(w, http.StatusBadRequest, errorFrameMap(&protoError{codeSecretInURL, "send everything in messages, not in the URL"}))
		return
	}
	if !s.originAllowed(r.Header.Get("Origin")) {
		writeJSON(w, http.StatusForbidden, errorFrameMap(&protoError{codeBadOrigin, "origin not allowed"}))
		return
	}
	ip := clientIP(r)
	if !s.conns.admit(ip) {
		writeJSON(w, http.StatusServiceUnavailable, errorFrameMap(&protoError{codeTooManyConnections, "too many connections, try again later"}))
		return
	}
	// The origin was checked above, against the configured list.
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		s.conns.release(ip)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	sess := newSession(s, c, ctx, cancel)
	sess.startHeartbeat()
	go sess.writeLoop()
	sess.run()
	s.conns.release(ip)
}

// clientIP is the host part of the request's remote address. No reverse
// proxy header is read yet (D7): behind one, every connection would look
// like it came from the proxy.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) originAllowed(origin string) bool {
	if s.cfg.AllowAnyOrigin {
		return true
	}
	if origin == "" {
		return false
	}
	for _, o := range s.cfg.AllowedOrigins {
		if strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}

func statusOf(err error) int {
	switch {
	case errors.Is(err, room.ErrRoomNotFound):
		return http.StatusNotFound
	case errors.Is(err, room.ErrTooManyRooms), errors.Is(err, room.ErrClosed):
		return http.StatusServiceUnavailable
	case errors.Is(err, engine.ErrModeUnsupported):
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func errorFrameMap(err error) map[string]any { return toErrorMsg(err).json() }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
