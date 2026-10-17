package engine

import (
	"fmt"
	"sort"
	"strings"

	"masoi/internal/i18n"
)

// summaryFor narrates the game that just ended, one line per night and day,
// plus the final outcome (audit K4). Only called once the game is Ended, when
// everybody already sees every role and the whole history (rules §4), so it
// may freely name a secret actor and target (who the pack bit, who the Witch
// healed) that stayed hidden from living players during play.
//
// log is the room's whole event log, which may hold an earlier game in the
// same room (audit K3, NewGame): only the events from the latest RolesDealt
// on are narrated.
func summaryFor(log []Event) []string {
	start := 0
	for i, e := range log {
		if _, ok := e.(RolesDealt); ok {
			start = i
		}
	}

	names := map[PlayerID]string{}
	for _, e := range log {
		if ev, ok := e.(PlayerJoined); ok {
			names[ev.Player] = ev.Nickname
		}
	}
	name := func(id PlayerID) string {
		if n, ok := names[id]; ok {
			return n
		}
		return string(id)
	}

	var lines []string
	var clauses []clause
	label := ""
	flush := func() {
		if len(clauses) == 0 {
			return
		}
		sort.SliceStable(clauses, func(i, j int) bool { return clauses[i].order < clauses[j].order })
		text := make([]string, len(clauses))
		for i, c := range clauses {
			text[i] = c.text
		}
		line := strings.Join(text, ", ") + "."
		if label != "" {
			line = label + ": " + line
		}
		lines = append(lines, line)
		clauses = nil
	}
	add := func(order int, text string) { clauses = append(clauses, clause{order, text}) }

	at := NewState()
	for i, e := range log {
		if i < start {
			at = Evolve(at, e)
			continue
		}
		switch ev := e.(type) {
		case NightStarted:
			flush()
			label = fmt.Sprintf("Đêm %d", ev.Night)
		case DayStarted:
			flush()
			label = fmt.Sprintf("Ngày %d", at.Night)
		case NightEnded:
			if target, ok := BiteTarget(at, WolfVoters(at)); ok {
				add(narrateBite, i18n.Vi(i18n.SummaryBite, name(target)))
			}
		case DawnAnnounced:
			add(narrateDawn, dawnClause(ev, name))
		case WolfPoisoned:
			add(narrateKnight, i18n.Vi(i18n.SummaryKnightPoison, name(ev.Wolf)))
		case ChiefElected:
			add(narrateChief, trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Chief))))
		case ChiefSucceeded:
			add(narrateChief, trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.To))))
		case ChiefLost:
			add(narrateChief, trimPeriod(i18n.Vi(ev.AnnouncementKey())))
		case Hanged:
			add(narrateVerdict, trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Player))))
		case Spared:
			add(narrateVerdict, trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Player))))
		case DayEnded:
			add(narrateVerdict, trimPeriod(i18n.Vi(ev.AnnouncementKey())))
		case HunterRevealed:
			add(narrateHunter, trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Hunter))))
		case HunterShot:
			add(narrateHunter, trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Hunter), name(ev.Target))))
		case GameEnded:
			flush()
			lines = append(lines, i18n.Vi(ev.AnnouncementKey()))
		default:
			if nar, ok := e.(Narrated); ok {
				if order, text, has := nar.Narrate(name); has {
					add(order, text)
				}
			}
		}
		at = Evolve(at, e)
	}
	flush()
	return lines
}

// clause is one fact placed in its reading order within a night or day line:
// the night's deaths are announced after every other step, but read better
// right after the bite that caused them, so the order here is the
// narration's own, not the log's (see the Narrate* constants).
type clause struct {
	order int
	text  string
}

func dawnClause(ev DawnAnnounced, name func(PlayerID) string) string {
	switch ev.Kind {
	case DawnDeaths:
		who := make([]string, len(ev.Dead))
		for i, id := range ev.Dead {
			who[i] = name(id)
		}
		return trimPeriod(i18n.Vi(i18n.DawnSomeoneDied, who...))
	case DawnProtected:
		return trimPeriod(i18n.Vi(i18n.DawnProtected))
	default:
		return trimPeriod(i18n.Vi(i18n.DawnNobodyDied))
	}
}

func trimPeriod(s string) string { return strings.TrimSuffix(s, ".") }
