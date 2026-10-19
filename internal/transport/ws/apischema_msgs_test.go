package ws

import (
	"fmt"
	"reflect"
	"sort"

	"masoi/internal/engine"
	"masoi/internal/roles"
	"masoi/internal/room"
)

const schemaDraft = "https://json-schema.org/draft/2020-12/schema"

// eventTypes is every event a client may be sent (dead players and the end
// screen see them all). TestAPIDocs_EveryEventInTheSource checks it against
// the source, so a new event cannot miss its schema.
var eventTypes = []engine.Event{
	engine.AbortVoteCast{}, engine.ChiefElected{}, engine.ChiefElectionStarted{}, engine.ChiefLost{},
	engine.ChiefSucceeded{}, engine.ChiefSuccessionStarted{}, engine.ChiefVoteCast{}, engine.DawnAnnounced{},
	engine.DawnStarted{}, engine.DayClockExpired{}, engine.DayEnded{}, engine.DayStarted{}, engine.GameEnded{},
	engine.Hanged{}, engine.HunterChainNoted{}, engine.HunterRevealed{}, engine.HunterShot{},
	engine.HunterShotStarted{}, engine.HunterTimedOut{}, engine.LoverGrieving{}, engine.LoversMeetAnnounced{},
	engine.NewGameStarted{}, engine.NightEnded{}, engine.NightStarted{}, engine.NightStepCalled{},
	engine.NominationCast{}, engine.NominationsLocking{}, engine.NominationsReopened{},
	engine.PhoneCheckAnnounced{}, engine.PhoneChecked{}, engine.PlayerDied{}, engine.PlayerJoined{},
	engine.PlayerKicked{}, engine.PlayerLeft{}, engine.RolesConfigured{}, engine.RolesDealt{},
	engine.RoleViewStarted{}, engine.SeatsSet{}, engine.SkipSet{}, engine.SoundConfigured{}, engine.Spared{},
	engine.TeamChanged{}, engine.TimersConfigured{}, engine.TrialStarted{}, engine.VerdictCast{},
	engine.VerdictStarted{}, engine.WolfPoisoned{},
	roles.DetectiveChecked{}, roles.LoversPaired{}, roles.PlayerCursed{}, roles.PlayerProtected{},
	roles.PlayerVisited{}, roles.SeerChecked{}, roles.WitchHealed{}, roles.WitchPoisoned{},
	roles.WolfSeerChecked{}, roles.WolfVoted{},
}

// coreArgs is the args shape of each core command (commands.go). The test
// checks it names every core command.
var coreArgs = map[string]any{
	"setSeats": orderArgs{}, "configureRoles": rolesArgs{}, "configureTimers": timersArgs{},
	"configureSound": soundModeArgs{}, "dealRoles": noArgs{}, "voteChief": candidateAr{},
	"nameSuccessor": successorAr{}, "nominate": targetArgs{}, "skipDay": onArgs{}, "verdict": killArgs{},
	"abortVote": onArgs{}, "newGame": noArgs{}, "hunterShoot": targetArgs{},
}

// commandArgs maps every wire command name to its args type.
func commandArgs() map[string]reflect.Type {
	out := map[string]reflect.Type{}
	for name, a := range coreArgs {
		out[name] = reflect.TypeOf(a)
	}
	for _, r := range engine.Registry {
		if cr, ok := r.(engine.CommandRole); ok {
			for _, c := range cr.Commands() {
				out[c.Name] = reflect.TypeOf(c.NewArgs()).Elem()
			}
		}
	}
	return out
}

func schemaDoc(title, desc string, root map[string]any, g *schemaGen) map[string]any {
	doc := map[string]any{"$schema": schemaDraft, "title": title, "description": desc, "$defs": g.defs}
	for k, v := range root {
		doc[k] = v
	}
	return doc
}

func oneOfRefs(names []string) map[string]any {
	var list []any
	for _, n := range names {
		list = append(list, ref(n))
	}
	return map[string]any{"oneOf": list}
}

func object(props map[string]any, required ...string) map[string]any {
	if required == nil {
		required = []string{}
	}
	return map[string]any{"type": "object", "properties": props, "required": required, "additionalProperties": false}
}

func constant(v any) map[string]any { return map[string]any{"const": v} }

func str() map[string]any { return map[string]any{"type": "string"} }

func roomCodeSchema() map[string]any {
	return map[string]any{"type": "string", "pattern": fmt.Sprintf("^[%s]{%d}$", room.CodeAlphabet, room.CodeLength),
		"description": "Room code. Not a secret."}
}

// serverSchema is docs/api/schema/server.schema.json: every WebSocket message
// the server sends.
func serverSchema(docs goDocs) (map[string]any, []error) {
	g := newSchemaGen(docs)
	g.defs["Welcome"] = object(map[string]any{
		"type":            constant("welcome"),
		"protocolVersion": withDescription(constant(ProtocolVersion), "Goes up only on a change an older client cannot ignore."),
		"player":          withDescription(g.treeSchema(playerIDType, false), "This phone's player."),
		"token":           withDescription(str(), "Secret: kept by the phone to rejoin. Sent on join and on rejoin by nickname + PIN."),
		"pin":             map[string]any{"type": "string", "pattern": "^[0-9]{4}$", "description": "Secret: shown only on this player's screen. Sent on join."},
	}, "type", "protocolVersion", "player")
	g.defs["ViewMessage"] = object(map[string]any{
		"type":     constant("view"),
		"deadline": map[string]any{"type": "integer", "description": "Unix ms when the running phase or night step ends."},
		"speaker":  map[string]any{"const": true, "description": "Only on the phone(s) that read the announcements aloud."},
		"people":   map[string]any{"type": "array", "items": g.treeSchema(reflect.TypeOf(room.Person{}), true), "description": "Everybody in the room, in join order. Only in the lobby and once the game has ended."},
		"view":     g.treeSchema(reflect.TypeOf(engine.View{}), false),
	}, "type", "view")
	var events []string
	for _, ev := range eventTypes {
		t := reflect.TypeOf(ev)
		props := map[string]any{"type": constant(t.Name()), "data": g.treeSchema(t, true)}
		required := []string{"type", "data"}
		if _, ok := ev.(announcement); ok {
			props["say"] = withDescription(str(), "Vietnamese line: the speaker phone reads it aloud, every phone shows it.")
			required = append(required, "say")
		} else if t == reflect.TypeOf(engine.NightStepCalled{}) {
			props["say"] = withDescription(str(), "The role being called, read aloud. Absent for the phone check and the lovers' step, which read their own line.")
		}
		name := t.Name() + "Event"
		g.defs[name] = object(props, required...)
		events = append(events, name)
	}
	sort.Strings(events)
	g.defs["Event"] = oneOfRefs(events)
	g.defs["EventsMessage"] = object(map[string]any{
		"type":   constant("events"),
		"reset":  withDescription(map[string]any{"type": "boolean"}, "True: events is the whole list and replaces the one the phone holds (never append it)."),
		"from":   withDescription(map[string]any{"type": "integer", "minimum": 0}, "Index of the first event in the phone's list."),
		"events": map[string]any{"type": "array", "items": ref("Event")},
	}, "type", "reset", "from", "events")
	g.defs["ErrorMessage"] = errorSchema()
	doc := schemaDoc("Ma Sói: server to client", "Every message the server sends on the game WebSocket (/ws).",
		oneOfRefs([]string{"Welcome", "ViewMessage", "EventsMessage", "ErrorMessage"}), g)
	return doc, g.errs
}

func errorSchema() map[string]any {
	return object(map[string]any{
		"type":       constant("error"),
		"code":       withDescription(str(), "Stable snake_case code; see docs/api/errors.md. Unknown codes must be handled."),
		"message":    withDescription(str(), "English detail for logs. Never show it to players."),
		"retryAfter": map[string]any{"type": "integer", "minimum": 0, "description": "Milliseconds; only with pin_locked."},
	}, "type", "code", "message")
}

// httpSchema is docs/api/schema/http.schema.json: the bodies of the HTTP
// endpoints.
func httpSchema(docs goDocs) (map[string]any, []error) {
	g := newSchemaGen(docs)
	g.defs["CreateRoomRequest"] = object(map[string]any{
		"mode": map[string]any{"type": "string", "enum": []string{"inPerson", "online"}, "description": "Optional, default inPerson. online is refused (mode_unsupported)."},
	})
	g.defs["CreateRoomReply"] = object(map[string]any{
		"room":    roomCodeSchema(),
		"hostKey": withDescription(str(), "Secret: only this phone ever sees it; send it in its own first join."),
	}, "room", "hostKey")
	g.defs["JoinRoomReply"] = object(map[string]any{"room": roomCodeSchema()}, "room")
	g.defs["GuideReply"] = g.tagSchema(reflect.TypeOf(guideResponse{}))
	g.defs["HTTPError"] = errorSchema()
	doc := schemaDoc("Ma Sói: HTTP bodies", "Request and reply bodies of POST /rooms, GET /join/{code} and GET /guide.",
		map[string]any{"anyOf": []any{ref("CreateRoomRequest"), ref("CreateRoomReply"), ref("JoinRoomReply"), ref("GuideReply"), ref("HTTPError")}}, g)
	return doc, g.errs
}
