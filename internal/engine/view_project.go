package engine

// ViewFor projects the state for one viewer. It is the only way state leaves
// the engine towards a screen.
func ViewFor(s State, v Viewer) View {
	full := seesAll(s, v)
	out := View{
		Phase:      s.Phase,
		Seq:        s.Seq,
		Night:      s.Night,
		Roles:      copyRoles(s.Settings.Roles),
		Chief:      s.Chief,
		Succession: successionFormer(s),
		Votes:      publicVotes(s),
		Full:       full,
	}
	if s.Phase == PhaseEnded {
		out.Winner, out.Draw = copyWinner(s.Winner), s.Winner == nil
	}
	for i, id := range s.Seats {
		p := s.Players[id]
		if p == nil {
			continue
		}
		pv := PlayerView{ID: id, Nickname: p.Nickname, Seat: i + 1, Alive: p.Alive}
		if full || id == v.Player {
			pv.Role, pv.Team = p.Role, p.Team
		}
		out.Players = append(out.Players, pv)
	}
	if full {
		out.Secrets = secrets(s)
	}
	if v.Narrator || full {
		out.NightCall = nightCall(s)
	}

	me := s.Players[v.Player]
	if v.Narrator || me == nil {
		return out
	}
	out.Me = myView(s, me)
	if full || (me.Alive && me.Team == TeamWolves) {
		out.Pack = packView(s)
	}
	out.Lover = loverView(s, me)
	if out.Lover != nil {
		for i := range out.Players {
			if out.Players[i].ID == out.Lover.Partner {
				out.Players[i].Role = out.Lover.PartnerRole
			}
		}
	}
	return out
}

// seesAll: dead players see everything (rules §4), and so does everybody once
// the game is over.
func seesAll(s State, v Viewer) bool {
	if s.Phase == PhaseEnded {
		return true
	}
	p := s.Players[v.Player]
	return !v.Narrator && p != nil && s.Dealt && !p.Alive
}

func successionFormer(s State) PlayerID {
	if s.Phase != PhaseChiefSuccession {
		return ""
	}
	return s.Succession.Former
}

func copyWinner(w *Team) *Team {
	if w == nil {
		return nil
	}
	out := *w
	return &out
}

func publicVotes(s State) PublicVotes {
	out := PublicVotes{
		Chief:     copyMap(s.ChiefVotes),
		Verdicts:  copyMap(s.Day.Verdicts),
		Defendant: s.Day.Defendant,
	}
	if len(s.Day.Nominations) > 0 {
		out.Nominations = make(map[PlayerID]PlayerID, len(s.Day.Nominations))
		for voter, n := range s.Day.Nominations {
			out.Nominations[voter] = n.Target
		}
	}
	for _, id := range s.Seats {
		if s.Day.Skips[id] {
			out.Skips = append(out.Skips, id)
		}
		if s.Day.Spared[id] {
			out.Spared = append(out.Spared, id)
		}
	}
	return out
}

func secrets(s State) *Secrets {
	pending := s.Pending
	pending.WolfVotes = copyMap(s.Pending.WolfVotes)
	out := &Secrets{Pending: pending}
	if s.Lovers != nil {
		lovers := *s.Lovers
		out.Lovers = &lovers
	}
	return out
}

func nightCall(s State) *NightCallView {
	step, ok := CurrentNightRole(s)
	if !ok {
		return nil
	}
	return &NightCallView{Role: step, Duration: s.Settings.NightStep}
}

func myView(s State, p *Player) *MyView {
	out := &MyView{ID: p.ID, Role: p.Role, Team: p.Team, Alive: p.Alive}
	if role := Registry[p.Role]; role != nil {
		out.Cursed = p.Team == TeamWolves && role.BaseTeam() != TeamWolves
	}
	out.Turn = turnView(s, p)
	return out
}

// turnView is set while the viewer is called: their role's step, or the pack
// step for anybody on the wolf team.
func turnView(s State, p *Player) *TurnView {
	step, ok := CurrentNightRole(s)
	if !ok || !p.Alive {
		return nil
	}
	if step != p.Role && !(isPackStep(step) && p.Team == TeamWolves) {
		return nil
	}
	out := &TurnView{Step: step, Visited: s.Pending.Visited == p.ID}
	if out.Visited || step != p.Role {
		return out
	}
	if w, ok := Registry[step].(BiteWitness); ok && w.SeesBite(p) {
		if target, bitten := BiteTarget(s, WolfVoters(s)); bitten {
			out.Bitten = target
		}
	}
	return out
}

func isPackStep(step RoleID) bool {
	always, ok := Registry[step].(AlwaysCalled)
	return ok && always.AlwaysCalled()
}

func packView(s State) *PackView {
	out := &PackView{Votes: copyMap(s.Pending.WolfVotes)}
	for _, id := range s.Seats {
		if p := s.Players[id]; p != nil && p.Team == TeamWolves {
			out.Members = append(out.Members, id)
		}
	}
	return out
}

// loverView: from the moment the pair exists, each lover sees the partner, the
// partner's role and the colour of the pair's current team (rules §7). Cupid
// gets nothing.
func loverView(s State, me *Player) *LoverView {
	if s.Lovers == nil {
		return nil
	}
	var partnerID PlayerID
	switch me.ID {
	case s.Lovers.A:
		partnerID = s.Lovers.B
	case s.Lovers.B:
		partnerID = s.Lovers.A
	default:
		return nil
	}
	partner := s.Players[partnerID]
	if partner == nil {
		return nil
	}
	return &LoverView{Partner: partnerID, PartnerRole: partner.Role, Colour: linkColour(me, partner)}
}

func linkColour(a, b *Player) LinkColour {
	switch {
	case a.Team == TeamLovers || b.Team == TeamLovers || a.Team != b.Team:
		return LinkStriped
	case a.Team == TeamWolves:
		return LinkBlack
	default:
		return LinkWhite
	}
}
