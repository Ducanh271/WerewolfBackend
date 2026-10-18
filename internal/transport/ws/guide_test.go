package ws

import (
	"encoding/json"
	"net/http"
	"testing"

	"masoi/internal/engine"
)

// T25 step 3b: GET /guide is public, needs no room, and carries every
// section.
func TestGuide_ReturnsEverySection(t *testing.T) {
	e := newEnv(t, nil)
	resp, err := http.Get(e.srv.URL + "/guide")
	if err != nil {
		t.Fatalf("GET /guide: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /guide: status %d", resp.StatusCode)
	}
	var body guideResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.QuickRules) != 12 {
		t.Errorf("quickRules has %d lines, want 12", len(body.QuickRules))
	}
	if len(body.Roles) != len(engine.Registry) {
		t.Errorf("roles has %d entries, Registry has %d roles", len(body.Roles), len(engine.Registry))
	}
	if len(body.Statuses) != 3 {
		t.Errorf("statuses has %d entries, want 3 (chief, lovers, cursed)", len(body.Statuses))
	}
	if len(body.FirstTime) == 0 {
		t.Error("firstTime is empty")
	}
	if len(body.Tips) == 0 {
		t.Error("tips is empty")
	}
	for _, role := range body.Roles {
		if role.Title == "" || role.WhenCalled == "" || role.HowTo == "" || role.OnTimeout == "" || role.Gotcha == "" {
			t.Errorf("role %q has an incomplete entry: %+v", role.Role, role)
		}
	}
}

// The guide needs no room code and no socket: a bare GET works from the
// lobby, before any room exists.
func TestGuide_NoRoomNeeded(t *testing.T) {
	e := newEnv(t, nil)
	resp, err := http.Get(e.srv.URL + "/guide")
	if err != nil {
		t.Fatalf("GET /guide: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /guide: status %d", resp.StatusCode)
	}
}
