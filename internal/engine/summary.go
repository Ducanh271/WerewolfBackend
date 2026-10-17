package engine

import (
	"sort"
	"strconv"
	"strings"

	"masoi/internal/i18n"
)

// summaryFor narrates the game that just ended, one line per night and day,
// plus the final outcome (audit K4). Only called once the game is Ended, when
// everybody already sees every role and the whole history (rules §4), so it
// may freely name a secret actor and target (who the pack bit, who the Witch
// healed) that stayed hidden from living players during play. Every piece of
// text is a full sentence template from i18n; this file only ever joins
// whole rendered sentences with punctuation, never builds one out of pieces.
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
	sec := newSection()
	flush := func() {
		if line, ok := sec.render(); ok {
			lines = append(lines, line)
		}
		sec = newSection()
	}

	at := NewState()
	for i, e := range log {
		if i < start {
			at = Evolve(at, e)
			continue
		}
		switch ev := e.(type) {
		case NightStarted:
			flush()
			sec.label = i18n.Vi(i18n.SummaryNightLabel, strconv.Itoa(ev.Night))
		case DayStarted:
			flush()
			sec.label = i18n.Vi(i18n.SummaryDayLabel, strconv.Itoa(at.Night))
		case NightEnded:
			if target, ok := BiteTarget(at, WolfVoters(at)); ok {
				sec.add(narrateBite, target, i18n.Vi(i18n.SummaryBite, name(target)))
			}
		case DawnAnnounced:
			sec.dawn = dawnClause(ev, name)
		case WolfPoisoned:
			sec.add(narrateKnight, ev.Wolf, i18n.Vi(i18n.SummaryKnightPoison, name(ev.Wolf)))
		case ChiefElected:
			sec.add(narrateChief, "", trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Chief))))
		case ChiefSucceeded:
			sec.add(narrateChief, "", trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.To))))
		case ChiefLost:
			sec.add(narrateChief, "", trimPeriod(i18n.Vi(ev.AnnouncementKey())))
		case Hanged:
			sec.add(narrateVerdict, "", trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Player))))
		case Spared:
			sec.add(narrateVerdict, "", trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Player))))
		case DayEnded:
			sec.add(narrateVerdict, "", trimPeriod(i18n.Vi(ev.AnnouncementKey())))
		case HunterRevealed:
			sec.add(narrateHunter, "", trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Hunter))))
		case HunterShot:
			sec.add(narrateHunter, "", trimPeriod(i18n.Vi(ev.AnnouncementKey(), name(ev.Hunter), name(ev.Target))))
		case GameEnded:
			flush()
			lines = append(lines, finalOutcomeLine(ev))
		default:
			if nar, ok := e.(Narrated); ok {
				if order, text, group, has := nar.Narrate(name); has {
					sec.add(order, group, text)
				}
			}
		}
		at = Evolve(at, e)
	}
	flush()
	return lines
}

// section accumulates the clauses of one night or one day. Clauses that name
// the same player merge into one sentence (audit K4): the pack's bite and the
// Witch's save read as "Sói cắn An, Phù thủy cứu An.", one fact, not two. A
// night's outcome (dawn) is never merged: it is always the line's last
// sentence, read after "→ ".
type section struct {
	label  string
	groups []*sectionGroup
	dawn   string
}

type sectionGroup struct {
	key     PlayerID // "" never merges with another clause
	order   int
	clauses []clause
}

type clause struct {
	order int
	text  string
}

func newSection() *section { return &section{} }

// add places a clause in the group named by key (merging it with any earlier
// clause of the same key in this section), or, for key "", in a new group of
// its own.
func (s *section) add(order int, key PlayerID, text string) {
	if key != "" {
		for _, g := range s.groups {
			if g.key == key {
				g.clauses = append(g.clauses, clause{order, text})
				if order < g.order {
					g.order = order
				}
				return
			}
		}
	}
	s.groups = append(s.groups, &sectionGroup{key: key, order: order, clauses: []clause{{order, text}}})
}

// render turns the section into its line, in the order its groups happened
// (rules §9), false when there is nothing to say (an empty section is simply
// dropped, never shown as an empty "Đêm N:").
func (s *section) render() (string, bool) {
	if len(s.groups) == 0 && s.dawn == "" {
		return "", false
	}
	sort.SliceStable(s.groups, func(i, j int) bool { return s.groups[i].order < s.groups[j].order })
	sentences := make([]string, 0, len(s.groups)+1)
	for _, g := range s.groups {
		sort.SliceStable(g.clauses, func(i, j int) bool { return g.clauses[i].order < g.clauses[j].order })
		parts := make([]string, len(g.clauses))
		for i, c := range g.clauses {
			parts[i] = c.text
		}
		sentences = append(sentences, strings.Join(parts, ", ")+".")
	}
	if s.dawn != "" {
		sentences = append(sentences, "→ "+s.dawn+".")
	}
	body := strings.Join(sentences, " ")
	if s.label != "" {
		return s.label + ": " + body, true
	}
	return body, true
}

// dawnClause is the night's outcome, read after "→ ": who died, never why
// (the clauses before it already say that a Guard or a Witch saved the
// target, so a save and an uneventful night both read as "Không ai chết.").
func dawnClause(ev DawnAnnounced, name func(PlayerID) string) string {
	if ev.Kind != DawnDeaths {
		return i18n.Vi(i18n.SummaryDawnNobodyDied)
	}
	who := make([]string, len(ev.Dead))
	for i, id := range ev.Dead {
		who[i] = name(id)
	}
	return i18n.Vi(i18n.SummaryDawnDeaths, who...)
}

// finalOutcomeLine is the game's last line. An abort (audit K1) gets its own,
// shorter wording (i18n.SummaryAborted), never the live announcement's.
func finalOutcomeLine(ev GameEnded) string {
	if ev.Outcome.Aborted {
		return i18n.Vi(i18n.SummaryAborted)
	}
	return i18n.Vi(ev.AnnouncementKey())
}

func trimPeriod(s string) string { return strings.TrimSuffix(s, ".") }
