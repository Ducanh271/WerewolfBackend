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
		Shooter:    shooter(s),
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
			if !full {
				pv.Team = VisibleTeam(p)
			}
		} else if p.Flags[FlagRevealed] != 0 {
			pv.Role = p.Role // a revealed Hunter: the role, not the team
		}
		out.Players = append(out.Players, pv)
	}
	if full {
		out.Secrets = secrets(s)
	}
	out.NightCall = nightCall(s)

	me := s.Players[v.Player]
	if me == nil {
		return out
	}
	out.Me = myView(s, me)
	if full || (me.Alive && VisibleTeam(me) == TeamWolves) {
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
	return p != nil && s.Dealt && !p.Alive
}

func successionFormer(s State) PlayerID {
	if s.Phase != PhaseChiefSuccession {
		return ""
	}
	return s.Succession.Former
}

func shooter(s State) PlayerID {
	if s.Phase != PhaseHunterShot {
		return ""
	}
	return s.Shot.Hunter
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
	pending.Blocked = nil // who blocked whom stays inside the engine
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
	return &NightCallView{Role: step, Duration: NightStepDuration(s, step)}
}

func myView(s State, p *Player) *MyView {
	out := &MyView{ID: p.ID, Role: p.Role, Team: p.Team, Alive: p.Alive, CanShoot: shooter(s) == p.ID}
	out.Cursed = IsCursed(p) && p.Flags[FlagTeamHidden] == 0
	out.Team = VisibleTeam(p)
	if step, ok := CurrentNightRole(s); ok && step == StepPhoneCheck && p.Alive {
		out.PhoneCheck = &PhoneCheckView{Cursed: IsCursed(p)}
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
	out := &TurnView{Step: step, Blocked: Blocked(s, p.ID), Doomed: isPackStep(step) && Doomed(s, p.ID)}
	if out.Blocked || step != p.Role {
		return out
	}
	if w, ok := Registry[step].(BiteWitness); ok && w.SeesBite(p) {
		out.BiteShown = true
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

// packStepReached: tonight's pack step has started, so the poisoned wolf is
// dead by the rules from now on.
func packStepReached(s State) bool {
	if s.Phase != PhaseNight {
		return false
	}
	for i, step := range nightSteps(s, s.Night) {
		if isPackStep(step) {
			return s.NightStep >= i+1
		}
	}
	return false
}

func packView(s State) *PackView {
	out := &PackView{Votes: copyMap(s.Pending.WolfVotes)}
	reached := packStepReached(s)
	for _, id := range s.Seats {
		if reached && Doomed(s, id) {
			out.Doomed = append(out.Doomed, id)
		}
		if p := s.Players[id]; p != nil && VisibleTeam(p) == TeamWolves {
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
	case VisibleTeam(a) == TeamLovers || VisibleTeam(b) == TeamLovers || VisibleTeam(a) != VisibleTeam(b):
		return LinkStriped
	case VisibleTeam(a) == TeamWolves:
		return LinkBlack
	default:
		return LinkWhite
	}
}
