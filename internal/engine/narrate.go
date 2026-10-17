package engine

// Narrated is implemented by a role event that has something to say in the
// end-of-game summary (audit K4): who did what to whom, through the i18n
// package (a full sentence template, never a fragment built in code). The
// engine cannot switch on a role's concrete event type (roles → engine is
// one-way), so this interface is the extension point, the same way Audienced
// and AnnouncementKey already are.
//
// order places the clause within its night or day line (see the Narrate*
// constants). group merges this clause with others about the same player
// into one sentence instead of its own (e.g. the pack's bite and the Witch's
// save read as "Sói cắn An, Phù thủy cứu An.", one sentence, because both
// name the same player); "" never merges. ok is false for a clause that
// turned out to have nothing to say (e.g. an unused choice).
type Narrated interface {
	Narrate(name func(PlayerID) string) (order int, text string, group PlayerID, ok bool)
}

// Narrate* is where a role's clause falls within its night, matching the call
// order of rules §9 so a night's clauses read in the order they happened. The
// engine's own clauses (the pack's bite, the dawn announcement, a hanging, a
// Hunter's shot, the Chief) are placed between them without asking a role.
const (
	NarrateCupid = iota
	NarrateVisitor
	NarrateGuard
	NarrateCurse
	NarrateWolfSeer
	narrateBite   // the engine's own, computed from the night's votes
	narrateKnight // the engine's own (delayed_death.go, the Rusty knight)
	NarrateWitch
	NarrateSeer
	NarrateDetective
	narrateDawn
	narrateVerdict
	narrateHunter
	narrateChief
)
