package ws

import (
	"time"
)

// firstGame plays a whole game that touches every role. The deal of the
// first seed (pinned by the checks below) is: An Rusty knight, Bình and
// Giang villagers, Chi Guard, Dũng Visitor, Hà Witch, Khoa Seer, Lan Hunter,
// Minh Cursed wolf, Nga Cupid, Phúc Wolf seer, Quân Detective.
func (g *exGame) firstGame() {
	an, binh, chi, dung := g.phone("An"), g.phone("Bình"), g.phone("Chi"), g.phone("Dũng")
	giang, ha, khoa, lan := g.phone("Giang"), g.phone("Hà"), g.phone("Khoa"), g.phone("Lan")
	minh, nga, phuc, quan := g.phone("Minh"), g.phone("Nga"), g.phone("Phúc"), g.phone("Quân")

	g.ok(an, "dealRoles", nil)
	g.sync()
	g.checkRoles(map[string]string{
		"An": "RustyKnight", "Bình": "Villager", "Chi": "Guard", "Dũng": "Visitor", "Giang": "Villager", "Hà": "Witch",
		"Khoa": "Seer", "Lan": "Hunter", "Minh": "CursedWolf", "Nga": "Cupid", "Phúc": "WolfSeer", "Quân": "Detective",
	})
	g.expectPhase("RoleView")
	g.shot(binh, "roleView")

	// The Chief election: Lan, the Hunter, wins.
	g.toDeadline()
	g.expectPhase("ChiefElection")
	g.fails(binh, "wrong_phase", "nominate", map[string]any{"target": chi.id})
	for _, p := range g.phones {
		target := lan
		if p == binh {
			target = binh
		}
		g.ok(p, "voteChief", map[string]any{"candidate": target.id})
	}
	g.sync()
	g.shot(binh, "chiefElection")
	g.toDeadline()

	// Night 1.
	g.expectPhase("Night")
	g.expectStep("Cupid")
	g.shot(chi, "night.asleep")
	g.shot(nga, "night.cupid")
	g.fails(binh, "not_your_turn", "cupidPair", map[string]any{"a": binh.id, "b": giang.id})
	g.fails(nga, "invalid_target", "cupidPair", map[string]any{"a": binh.id, "b": binh.id})
	g.ok(nga, "cupidPair", map[string]any{"a": binh.id, "b": giang.id})
	g.toStep("PhoneCheck")
	g.shot(binh, "night.phoneCheck.inLove")
	g.shot(chi, "night.phoneCheck.notInLove")
	g.toStep("LoversMeet")
	g.shot(binh, "night.loversMeet")
	g.toStep("Visitor")
	g.shot(dung, "night.visitor")
	g.ok(dung, "visitorVisit", map[string]any{"target": khoa.id})
	g.toStep("Guard")
	g.shot(chi, "night.guard")
	g.ok(chi, "guardProtect", map[string]any{"target": ha.id})
	g.toStep("CursedWolf")
	g.shot(minh, "night.cursedWolf")
	g.fails(minh, "invalid_target", "cursedWolfCurse", map[string]any{"target": phuc.id})
	g.ok(minh, "cursedWolfCurse", map[string]any{"target": quan.id})
	g.sync()
	g.fails(minh, "already_acted", "cursedWolfCurse", map[string]any{"target": khoa.id})
	g.toStep("PhoneCheck")
	g.shot(quan, "night.phoneCheck.cursed")
	g.toStep("WolfSeer")
	g.ok(phuc, "wolfSeerCheck", map[string]any{"target": khoa.id})
	g.sync()
	g.shot(phuc, "night.wolfSeer")
	g.toStep("Werewolf")
	for _, p := range []*exPhone{minh, phuc, quan} {
		g.ok(p, "wolfVote", map[string]any{"target": an.id})
	}
	g.ok(phuc, "wolfVote", map[string]any{"target": ""})
	g.ok(phuc, "wolfVote", map[string]any{"target": an.id})
	g.sync()
	g.shot(minh, "night.pack")
	g.toStep("Witch")
	g.shot(ha, "night.witch")
	g.ok(ha, "witchHeal", nil)
	g.toStep("Seer")
	g.shot(khoa, "night.seer.blocked")
	g.fails(khoa, "blocked", "seerCheck", map[string]any{"target": minh.id})
	g.toStep("Detective")
	g.ok(quan, "detectiveCheck", map[string]any{"target": ha.id})
	g.sync()
	g.shot(quan, "night.detective")
	g.toDeadline()

	// Morning 1: nobody died (the Witch healed An, but Minh is poisoned).
	g.expectPhase("Dawn")
	g.shot(binh, "dawn")
	g.toDeadline()

	// Day 1: a countdown starts and stops, Phúc is put on trial and spared,
	// then the day clock runs out with Chi nominated.
	g.expectPhase("Day")
	g.fails(binh, "self_nomination", "nominate", map[string]any{"target": binh.id})
	g.ok(binh, "skipDay", map[string]any{"on": true})
	g.ok(binh, "abortVote", map[string]any{"on": true})
	g.ok(binh, "abortVote", map[string]any{"on": false})
	g.nominateAll(func(p *exPhone) *exPhone {
		if p == chi {
			return phuc
		}
		return nil
	})
	g.expectPhase("NominationLock")
	g.ok(chi, "nominate", map[string]any{"target": ""})
	g.expectPhase("Day")
	g.nominateAll(func(p *exPhone) *exPhone {
		switch p {
		case phuc:
			return chi
		case khoa, dung:
			return nil
		}
		return phuc
	})
	g.expectPhase("NominationLock")
	g.shot(binh, "nominationLock")
	g.toDeadline()
	g.expectPhase("Trial")
	g.shot(binh, "trial")
	g.toDeadline()
	g.expectPhase("Verdict")
	g.fails(phuc, "on_trial", "verdict", map[string]any{"kill": false})
	// The vote closes once everybody voted: the last one, Quân, comes after
	// the screenshot.
	for i, p := range g.phones {
		if p != phuc && p != quan {
			g.ok(p, "verdict", map[string]any{"kill": i%3 == 0})
		}
	}
	g.shot(binh, "verdict")
	g.ok(quan, "verdict", map[string]any{"kill": false})
	g.expectPhase("Day")
	g.fails(binh, "already_spared", "nominate", map[string]any{"target": phuc.id})
	g.ok(binh, "nominate", map[string]any{"target": chi.id})
	g.toDeadline()
	g.expectPhase("Trial")
	g.untilPhase("Night")

	// Night 2: the poison kills Minh at the pack step; Lan is bitten.
	g.expectPhase("Night")
	g.toStep("Visitor")
	g.ok(dung, "visitorVisit", map[string]any{"target": phuc.id})
	g.toStep("Guard")
	g.fails(chi, "same_guard_target", "guardProtect", map[string]any{"target": ha.id})
	g.ok(chi, "guardProtect", map[string]any{"target": khoa.id})
	g.toStep("Werewolf")
	g.shot(minh, "night.pack.doomed")
	g.ok(quan, "wolfVote", map[string]any{"target": lan.id})
	g.toStep("Witch")
	g.fails(ha, "potion_used", "witchHeal", nil)
	g.toStep("Seer")
	g.ok(khoa, "seerCheck", map[string]any{"target": quan.id})
	g.sync()
	g.shot(khoa, "night.seer")
	g.toStep("Detective")
	g.fails(quan, "ability_used", "detectiveCheck", map[string]any{"target": an.id})
	g.toDeadline()

	// Morning 2: Lan and Minh died. Lan, the Hunter, shoots Phúc, then names
	// Chi Chief.
	g.expectPhase("HunterShot")
	g.shot(lan, "hunterShot.hunter")
	g.shot(binh, "hunterShot")
	g.ok(lan, "hunterShoot", map[string]any{"target": phuc.id})
	g.sync()
	g.expectPhase("ChiefSuccession")
	g.shot(lan, "chiefSuccession.chief")
	g.shot(binh, "chiefSuccession")
	g.ok(lan, "nameSuccessor", map[string]any{"successor": chi.id})
	g.sync()
	g.toDeadline()

	// Day 2: the village hangs Bình; Giang, his lover, will die of grief.
	g.expectPhase("Day")
	g.nominateAll(func(p *exPhone) *exPhone {
		if p == binh {
			return quan
		}
		return binh
	})
	g.toDeadline()
	g.toDeadline()
	g.expectPhase("Verdict")
	for _, p := range g.livingPhones() {
		if p != binh {
			g.ok(p, "verdict", map[string]any{"kill": true})
		}
	}
	g.sync()
	g.toDeadline()

	// Night 3: the Witch poisons Quân, the last wolf.
	g.expectPhase("Night")
	g.shot(lan, "night.dead")
	g.toStep("Witch")
	g.ok(ha, "witchPoison", map[string]any{"target": quan.id})
	g.untilPhase("Ended")
	g.shot(binh, "ended")
}

// nominateAll has every living player nominate who(p), nil for no one.
func (g *exGame) nominateAll(who func(p *exPhone) *exPhone) {
	g.t.Helper()
	for _, p := range g.livingPhones() {
		target := ""
		if q := who(p); q != nil {
			target = q.id
		}
		g.ok(p, "nominate", map[string]any{"target": target})
	}
}

func (g *exGame) checkRoles(want map[string]string) {
	g.t.Helper()
	for _, p := range g.phones {
		p.role = p.view().Str("view", "me", "role")
		if want[p.nick] != p.role {
			g.t.Fatalf("%s got %s, the script expects %s: the deal changed, fix the script", p.nick, p.role, want[p.nick])
		}
	}
}

func (g *exGame) host() *exPhone { return g.phones[0] }

func (g *exGame) phase() string { return g.host().view().Str("view", "phase") }

func (g *exGame) expectPhase(phase string) {
	g.t.Helper()
	if got := g.phase(); got != phase {
		g.t.Fatalf("phase %s, want %s", got, phase)
	}
}

func (g *exGame) step() string { return g.host().view().Str("view", "nightCall", "role") }

func (g *exGame) expectStep(role string) {
	g.t.Helper()
	if got := g.step(); got != role {
		g.t.Fatalf("night step %s, want %s", got, role)
	}
}

// toDeadline lets the running phase or night step run out.
func (g *exGame) toDeadline() {
	g.t.Helper()
	ms, ok := g.host().view()["deadline"].(float64)
	if !ok {
		g.t.Fatalf("phase %s has no deadline", g.phase())
	}
	g.advance(time.UnixMilli(int64(ms)).Sub(g.e.clk.Now()))
}

// toStep lets night steps run out until role is called.
func (g *exGame) toStep(role string) {
	g.t.Helper()
	for i := 0; g.step() != role; i++ {
		if i > 15 || g.phase() != "Night" {
			g.t.Fatalf("night step %s never came (phase %s, step %s)", role, g.phase(), g.step())
		}
		g.toDeadline()
	}
}

// untilPhase lets phases and steps run out until phase starts.
func (g *exGame) untilPhase(phase string) {
	g.t.Helper()
	for i := 0; g.phase() != phase; i++ {
		if i > 20 {
			g.t.Fatalf("phase %s never came (now %s)", phase, g.phase())
		}
		g.toDeadline()
	}
}

func (g *exGame) livingPhones() []*exPhone {
	var out []*exPhone
	for _, p := range g.phones {
		if p.view().Get("view", "me", "alive") == true {
			out = append(out, p)
		}
	}
	return out
}
