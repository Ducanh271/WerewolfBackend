package roles_test

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
)

const hello = "Tôi nghi An"

// toTrial puts target on trial, with the defense running.
func toTrial(d *Day, target string) {
	for _, name := range all {
		if name == target {
			d.NominateNobody(name)
		} else {
			d.Nominate(name, target)
		}
	}
	d.TimeOut().ExpectOnTrial(target).ExpectPhase(engine.PhaseTrial)
}

// aborted is a game called off by vote, with every player still alive.
func aborted(t *testing.T) *Game {
	g, d := day(t)
	for _, name := range all[:7] {
		g.Do(engine.CastAbortVote{Voter: g.Player(name), Yes: true})
	}
	d.ExpectPhase(engine.PhaseEnded)
	return g
}

// Rules §4 "Nhắn tin", played through a real game. Hoa is a living bystander.
func TestChat_InARealGame_OpenOrClosedByPhase(t *testing.T) {
	tests := []struct {
		name  string
		phase engine.Phase
		setup func(t *testing.T) *Game
		open  bool
	}{
		{"role view", engine.PhaseRoleView, func(t *testing.T) *Game { return newTable(t) }, false},
		{"Chief election", engine.PhaseChiefElection, func(t *testing.T) *Game {
			g := newTable(t)
			g.Election()
			return g
		}, true},
		{"night", engine.PhaseNight, func(t *testing.T) *Game {
			g := chiefTable(t, "Lan")
			g.Night()
			return g
		}, false},
		{"dawn", engine.PhaseDawn, func(t *testing.T) *Game {
			g := chiefTable(t, "Lan")
			g.Night()
			g.Dawn()
			return g
		}, false},
		{"discussion", engine.PhaseDay, func(t *testing.T) *Game {
			g, _ := day(t)
			return g
		}, true},
		{"nomination countdown", engine.PhaseNominationLock, func(t *testing.T) *Game {
			g, d := day(t)
			d.Nominate("An", "Giang")
			nobodyExcept(d, "An")
			return g
		}, true},
		{"kill or spare", engine.PhaseVerdict, func(t *testing.T) *Game {
			g, d := day(t)
			toVerdict(d, "Giang")
			return g
		}, false},
		{"waiting for the night", engine.PhaseDusk, func(t *testing.T) *Game {
			g, d := day(t)
			d.Hanged("Giang")
			return g
		}, false},
		{"the Hunter's shot", engine.PhaseHunterShot, func(t *testing.T) *Game {
			g := hunterTable(t)
			bitHunter(g)
			g.Dawn()
			return g
		}, false},
		{"the dead Chief naming a successor", engine.PhaseChiefSuccession, chiefBitten, false},
		{"after the game", engine.PhaseEnded, aborted, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := tc.setup(t)
			g.ExpectPhase(tc.phase)
			if tc.open {
				g.Chat("Hoa", hello).As("Khoa").ChatReads(engine.ChatSent{By: "Hoa", Text: hello})
				return
			}
			g.ChatExpectErr("Hoa", hello, engine.ErrChatClosed)
		})
	}
}

func TestChat_InTheLobby_NobodyWrites(t *testing.T) {
	g := NewGame(t, Seats("An", "Binh", "Chi", "Dung", "Em"))
	g.ChatExpectErr("An", hello, engine.ErrChatClosed)
}

func TestChat_DuringTheDefense_OnlyThePlayerOnTrialWrites(t *testing.T) {
	g, d := day(t)
	toTrial(d, "Giang")
	g.Chat("Giang", "Tôi là dân")
	for _, name := range all {
		if name != "Giang" {
			g.ChatExpectErr(name, hello, engine.ErrChatClosed)
		}
	}
	g.As("Hoa").ChatReads(engine.ChatSent{By: "Giang", Text: "Tôi là dân"})
}

// bittenHoaDay is day 1 with Hoa dead.
func bittenHoaDay(t *testing.T) (*Game, *Day) {
	g := chiefTable(t, "Lan")
	g.Night().Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Hoa")
	g.Dawn().ExpectDead("Hoa")
	return g, g.Day()
}

func TestChat_TheDeadReadButNeverWrite(t *testing.T) {
	g, _ := bittenHoaDay(t)
	g.Chat("Em", hello)
	before := len(g.Events())

	evs, err := engine.Decide(g.State(), engine.SendChat{By: "Hoa", Text: hello})
	if err == nil || len(evs) != 0 {
		t.Fatalf("the dead wrote: %v, %v", evs, err)
	}
	g.ChatExpectErr("Hoa", hello, engine.ErrNotAlive)
	if len(g.Events()) != before {
		t.Fatal("a refused message reached the log")
	}
	g.As("Hoa").ChatReads(engine.ChatSent{By: "Em", Text: hello})
}

// At night nobody can tell anything from the answer: the same refusal for a
// wolf, the Guard, a villager and the dead (rules §3).
func TestChat_AtNight_EverybodyGetsTheSameAnswer(t *testing.T) {
	g, d := bittenHoaDay(t)
	d.NobodyHanged()
	g.Night()
	g.ExpectPhase(engine.PhaseNight)
	for _, name := range all {
		g.ChatExpectErr(name, hello, engine.ErrChatClosed)
	}
}

// Every phone gets the same message: living wolf, villager, dead, the writer.
func TestChat_ReachesEveryViewerWithTheSameFields(t *testing.T) {
	g, _ := bittenHoaDay(t)
	g.Chat("Em", "  Tôi nghi\nAn ").Chat("An", "Không phải tôi")
	want := []engine.ChatSent{{By: "Em", Text: "Tôi nghi An"}, {By: "An", Text: "Không phải tôi"}}
	for _, name := range all {
		g.As(name).ChatReads(want...)
	}
}

func TestChat_ChangesNothingInTheGame(t *testing.T) {
	tests := []struct {
		name   string
		writer string
		setup  func(t *testing.T) *Game
	}{
		{"discussion, with votes cast", "Hoa", func(t *testing.T) *Game {
			g, d := day(t)
			d.Nominate("An", "Giang").Skip("Binh")
			return g
		}},
		{"nomination countdown", "Hoa", func(t *testing.T) *Game {
			g, d := day(t)
			d.Nominate("An", "Giang")
			nobodyExcept(d, "An")
			return g
		}},
		{"defense", "Giang", func(t *testing.T) *Game {
			g, d := day(t)
			toTrial(d, "Giang")
			return g
		}},
		{"Chief election", "Hoa", func(t *testing.T) *Game {
			g := newTable(t)
			g.Election().Vote("An", "Em")
			return g
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := tc.setup(t)
			before := g.State()
			g.Chat(tc.writer, hello).Chat(tc.writer, hello)
			if !reflect.DeepEqual(g.State(), before) {
				t.Fatalf("the state changed:\n before %+v\n after  %+v", before, g.State())
			}
		})
	}
}

// A message is not a change of choice: the running countdown stays the live one.
func TestChat_DoesNotRestartTheNominationCountdown(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock)
	first := engine.PhaseTimeout{Phase: g.State().Phase, Seq: g.State().Seq}

	g.Chat("Hoa", hello).Chat("Giang", "Tôi là dân")

	g.Do(first)
	d.ExpectPhase(engine.PhaseTrial).ExpectOnTrial("Giang")
}

func TestChat_AfterTheGame_EverybodyReadsItAndTheSummaryIsTheSame(t *testing.T) {
	g, d := day(t)
	g.Chat("Em", hello)
	for _, name := range all[:7] {
		g.Do(engine.CastAbortVote{Voter: g.Player(name), Yes: true})
	}
	d.ExpectPhase(engine.PhaseEnded)
	g.ChatExpectErr("Em", hello, engine.ErrChatClosed)

	var withoutChat []engine.Event
	for _, e := range g.Events() {
		if _, ok := e.(engine.ChatSent); !ok {
			withoutChat = append(withoutChat, e)
		}
	}
	for _, name := range all {
		g.As(name).ChatReads(engine.ChatSent{By: "Em", Text: hello})
		got := engine.ViewFor(g.State(), g.Events(), engine.PlayerViewer(g.Player(name))).Summary
		want := engine.ViewFor(g.State(), withoutChat, engine.PlayerViewer(g.Player(name))).Summary
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("the chat changed the summary:\n got  %v\n want %v", got, want)
		}
	}
}
