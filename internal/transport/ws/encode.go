package ws

import (
	"encoding/json"
	"reflect"

	"masoi/internal/engine"
	"masoi/internal/i18n"
	"masoi/internal/room"
)

// announcement is implemented by events with a Vietnamese line to read aloud.
type announcement interface{ AnnouncementKey() string }

// frames turns what the room sends one phone into the JSON messages of
// docs/protocol.md: `view`, then `events` if there are any.
func frames(m room.Message) ([][]byte, error) {
	var out [][]byte
	var names func(engine.PlayerID) string
	if m.View != nil {
		v := m.View
		names = nicknames(v.View)
		msg := map[string]any{"type": "view", "view": toTree(v.View)}
		if !v.Deadline.IsZero() {
			msg["deadline"] = v.Deadline.UnixMilli()
		}
		if v.Speaker {
			msg["speaker"] = true
		}
		b, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if m.Events != nil {
		b, err := json.Marshal(eventsMsg(m.Events, names))
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

func eventsMsg(e *room.EventsMsg, names func(engine.PlayerID) string) map[string]any {
	list := make([]any, 0, len(e.Events))
	for _, ev := range e.Events {
		list = append(list, eventTree(ev, names))
	}
	return map[string]any{"type": "events", "reset": e.Reset, "from": e.From, "events": list}
}

// eventTree is {"type", "data", "say"}: the say line only for announcements.
func eventTree(ev engine.Event, names func(engine.PlayerID) string) map[string]any {
	data, _ := toTree(ev).(map[string]any)
	if data == nil {
		data = map[string]any{}
	}
	out := map[string]any{"type": reflect.TypeOf(ev).Name(), "data": data}
	if a, ok := ev.(announcement); ok {
		var who []string
		for _, id := range announcedPlayers(ev) {
			if names != nil {
				who = append(who, names(id))
			} else {
				who = append(who, string(id))
			}
		}
		out["say"] = i18n.Vi(a.AnnouncementKey(), who...)
	}
	return out
}

// announcedPlayers lists, in order, the players an announcement names.
func announcedPlayers(ev engine.Event) []engine.PlayerID {
	switch e := ev.(type) {
	case engine.DawnAnnounced:
		return e.Dead
	case engine.ChiefElected:
		return []engine.PlayerID{e.Chief}
	case engine.ChiefSucceeded:
		return []engine.PlayerID{e.To}
	case engine.Hanged:
		return []engine.PlayerID{e.Player}
	case engine.Spared:
		return []engine.PlayerID{e.Player}
	}
	return nil
}

// nicknames looks a seat up in the view the phone got with the events.
func nicknames(v engine.View) func(engine.PlayerID) string {
	byID := map[engine.PlayerID]string{}
	for _, p := range v.Players {
		byID[p.ID] = p.Nickname
	}
	return func(id engine.PlayerID) string {
		if n, ok := byID[id]; ok {
			return n
		}
		return string(id)
	}
}

func welcome(player engine.PlayerID, creds *room.Credentials) []byte {
	m := map[string]any{"type": "welcome", "player": string(player)}
	if creds != nil && creds.Token != "" {
		m["token"] = creds.Token
	}
	if creds != nil && creds.PIN != "" {
		m["pin"] = creds.PIN
	}
	b, _ := json.Marshal(m)
	return b
}

func errorFrame(err error) []byte {
	b, _ := json.Marshal(toErrorMsg(err).json())
	return b
}
