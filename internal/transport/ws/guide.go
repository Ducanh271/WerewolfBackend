package ws

import (
	"net/http"

	"masoi/internal/engine"
	"masoi/internal/i18n"
	"masoi/internal/room"
)

// guideEntryJSON is the JSON shape of i18n.GuideEntry (docs/protocol.md
// "GET /guide"): the same shape for every role and every status.
type guideEntryJSON struct {
	Role       string `json:"role"`
	Title      string `json:"title"`
	WhenCalled string `json:"whenCalled"`
	HowTo      string `json:"howTo"`
	OnTimeout  string `json:"onTimeout"`
	Gotcha     string `json:"gotcha"`
	Note       string `json:"note,omitempty"`
}

// guideResponse is the body of GET /guide.
type guideResponse struct {
	QuickRules []string         `json:"quickRules"`
	Roles      []guideEntryJSON `json:"roles"`
	Statuses   []guideEntryJSON `json:"statuses"`
	FirstTime  []string         `json:"firstTime"`
	Tips       []string         `json:"tips"`
}

// minPlayersOf reads a role's own MinPlayersRule (rules §2), 0 when it has
// none.
func minPlayersOf(id engine.RoleID) int {
	role, ok := engine.Registry[id]
	if !ok {
		return 0
	}
	rule, ok := role.(engine.MinPlayersRule)
	if !ok {
		return 0
	}
	return rule.MinPlayers()
}

// buildGuide renders the in-app guide (T25) once, from the real default
// settings and constants, so GET /guide is the same for everybody and never
// drifts from the numbers the game actually uses (rules/timers are read
// once at server start; they are not a per-room setting here, since the
// guide is public and used before a room even exists).
func buildGuide() guideResponse {
	set := engine.DefaultSettings()
	timers := i18n.GuideTimers{
		RoleView:            set.RoleView,
		HunterShot:          set.HunterShot,
		LoversMeet:          set.LoversMeet,
		ChiefSuccession:     set.ChiefSuccession,
		PINMaxFailures:      room.MaxPINFailures,
		PINLockTime:         room.PINLockTime,
		CupidMinPlayers:     minPlayersOf("Cupid"),
		DetectiveMinPlayers: minPlayersOf("Detective"),
	}
	return guideResponse{
		QuickRules: i18n.QuickRules(),
		Roles:      toGuideEntries(i18n.RoleGuides(timers)),
		Statuses:   toGuideEntries(i18n.StatusGuides(timers)),
		FirstTime:  i18n.FirstTimeChecklist(timers),
		Tips:       i18n.OptionalTips(),
	}
}

func toGuideEntries(entries []i18n.GuideEntry) []guideEntryJSON {
	out := make([]guideEntryJSON, len(entries))
	for i, e := range entries {
		out[i] = guideEntryJSON{
			Role:       e.ID,
			Title:      e.Title,
			WhenCalled: e.WhenCalled,
			HowTo:      e.HowTo,
			OnTimeout:  e.OnTimeout,
			Gotcha:     e.Gotcha,
			Note:       e.Note,
		}
	}
	return out
}

func (s *Server) serveGuide(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.guide)
}
