package roles_test

import (
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/i18n"
)

func TestDay_NominationsClose_WhenEveryLivingPlayerChose(t *testing.T) {
	_, d := day(t)
	for _, name := range all {
		if name != "Giang" && name != "Lan" {
			d.Nominate(name, "Giang")
		}
	}
	d.Nominate("Lan", "Giang")
	d.ExpectPhase(engine.PhaseDay) // Giang has not chosen yet
	d.NominateNobody("Giang")
	d.ExpectPhase(engine.PhaseNominationLock).TimeOut()
	d.ExpectPhase(engine.PhaseTrial).ExpectOnTrial("Giang")
}

func TestDay_NominationCanBeChangedUntilItCloses(t *testing.T) {
	_, d := day(t)
	d.Nominate("An", "Giang").Nominate("An", "Hoa")
	nobodyExcept(d, "An", "Hoa", "Giang")
	d.NominateNobody("Hoa").Nominate("Giang", "Phuc").ExpectPhase(engine.PhaseNominationLock)
	d.TimeOut().ExpectOnTrial("Hoa")
}

func TestDay_EveryoneChoosesNobody_DiscussionContinues(t *testing.T) {
	_, d := day(t)
	nobodyExcept(d)
	d.ExpectPhase(engine.PhaseDay)
	// Changing one's mind later means everybody has chosen: the countdown starts.
	d.Nominate("An", "Giang").ExpectPhase(engine.PhaseNominationLock).TimeOut()
	d.ExpectOnTrial("Giang")
}

func TestDay_MostNominationsGoesOnTrial_EvenIfMostPlayersNominatedNobody(t *testing.T) {
	_, d := day(t)
	d.Nominate("An", "Hoa")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock).TimeOut().ExpectOnTrial("Hoa")
}

func TestDay_CannotNominateYourselfOrADeadPlayerOrBeDeadYourself(t *testing.T) {
	g := chiefTable(t, "Phuc")
	g.Night().Wolf("An").Bites("Lan").Wolf("Binh").Bites("Lan")
	g.Dawn().ExpectDead("Lan")
	d := g.Day()
	d.NominateExpectErr("An", "An", engine.ErrSelfNomination)
	d.NominateExpectErr("An", "Lan", engine.ErrInvalidTarget)
	d.NominateExpectErr("Lan", "An", engine.ErrNotAlive)
	d.NominateExpectErr("Nobody", "An", engine.ErrUnknownPlayer)
	d.NominateExpectErr("An", "Nobody", engine.ErrInvalidTarget)
}

func TestDay_TieInNominations_EarliestStandingVoteGoesOnTrial(t *testing.T) {
	tests := []struct {
		name  string
		votes [][2]string // voter, target ("" = no one), in the order cast
		want  string
	}{
		{"first voted goes first", [][2]string{{"An", "Giang"}, {"Binh", "Phuc"}, {"Chi", "Phuc"}, {"Dung", "Giang"}}, "Giang"},
		{"other side first", [][2]string{{"Binh", "Phuc"}, {"An", "Giang"}, {"Chi", "Giang"}, {"Dung", "Phuc"}}, "Phuc"},
		{"withdrawn vote does not count", [][2]string{{"An", "Giang"}, {"Binh", "Phuc"}, {"An", ""}, {"Chi", "Giang"}, {"Dung", "Phuc"}}, "Phuc"},
		{"a changed vote is a new vote", [][2]string{{"An", "Giang"}, {"Binh", "Phuc"}, {"An", "Hoa"}, {"An", "Giang"}, {"Chi", "Phuc"}, {"Dung", "Giang"}}, "Phuc"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, d := day(t)
			voted := map[string]bool{}
			for _, v := range tc.votes {
				voted[v[0]] = true
				if v[1] == "" {
					d.NominateNobody(v[0])
				} else {
					d.Nominate(v[0], v[1])
				}
			}
			for _, name := range all { // everyone else, last, chooses no one
				if !voted[name] {
					d.NominateNobody(name)
				}
			}
			d.ExpectPhase(engine.PhaseNominationLock).TimeOut().ExpectOnTrial(tc.want)
		})
	}
}

func TestDay_ChiefsNominationCountsDouble(t *testing.T) {
	_, d := chiefDay(t)
	d.Nominate("An", "Hoa").Nominate("Phuc", "Giang") // one head each, but the Chief counts twice
	nobodyExcept(d, "An", "Phuc")
	d.TimeOut().ExpectOnTrial("Giang")
}

func TestDay_SkipNeedsMoreThanSixtySixPercent(t *testing.T) {
	_, d := day(t)
	for _, name := range all[:6] {
		d.Skip(name)
	}
	d.ExpectPhase(engine.PhaseDay) // 60%
	d.Skip("Giang").ExpectPhase(engine.PhaseDusk)
}

func TestDay_ChiefsSkipCountsOnce(t *testing.T) {
	_, d := chiefDay(t)
	for _, name := range []string{"An", "Binh", "Chi", "Dung", "Em"} {
		d.Skip(name)
	}
	d.Skip("Phuc").ExpectPhase(engine.PhaseDay) // 6 of 10, not 7
	d.Skip("Giang").ExpectPhase(engine.PhaseDusk)
}

func TestDay_SkipCanBeReleased(t *testing.T) {
	_, d := day(t)
	for _, name := range all[:6] {
		d.Skip(name)
	}
	d.Unskip("An").Skip("Giang").ExpectPhase(engine.PhaseDay)
	d.Skip("An").ExpectPhase(engine.PhaseDusk)
}

func TestDay_NominatingSomeoneCancelsYourSkip(t *testing.T) {
	_, d := day(t)
	d.Skip("An").Nominate("An", "Giang")
	for _, name := range all[1:7] {
		d.Skip(name)
	}
	d.ExpectPhase(engine.PhaseDay) // six skips: An's was cancelled
}

func TestDay_ChoosingNobodyKeepsYourSkip(t *testing.T) {
	_, d := day(t)
	d.Skip("An").NominateNobody("An")
	for _, name := range all[1:7] {
		d.Skip(name)
	}
	d.ExpectPhase(engine.PhaseDusk)
}

func TestDay_SkipEndsTheDayEvenWithNominations(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang")
	for _, name := range all[1:8] {
		d.Skip(name)
	}
	d.ExpectPhase(engine.PhaseDusk)
	g.ExpectAlive(all...).Night() // on to night 2
	d.ExpectAnnounced(i18n.DayNobodyHanged)
}

func TestDay_ClockRunsOut_NobodyNominated_GoesStraightToNight(t *testing.T) {
	_, d := day(t)
	nobodyExcept(d)
	d.TimeOut().ExpectPhase(engine.PhaseDusk).ExpectAnnounced(i18n.DayNobodyHanged)
}

func TestDay_ClockRunsOut_TopNomineeStillGetsDefenseAndVerdict(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang").Nominate("Binh", "Giang").Nominate("Chi", "Hoa") // nominations never closed
	d.TimeOut().ExpectPhase(engine.PhaseTrial).ExpectOnTrial("Giang")
	d.TimeOut().ExpectPhase(engine.PhaseVerdict)
	for _, name := range all[:5] {
		d.Kill(name)
	}
	d.TimeOut().ExpectPhase(engine.PhaseDusk).ExpectAnnounced(i18n.DayHanged)
	g.ExpectDead("Giang")
}

func TestDay_ClockRunsOut_SparedAtTheLastVerdict_GoesToNightNotBackToNominations(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang").TimeOut().TimeOut()
	for _, name := range all[:5] {
		d.Spare(name)
	}
	d.TimeOut().ExpectPhase(engine.PhaseDusk).ExpectAnnounced(i18n.DaySpared)
	g.ExpectAlive("Giang")
}

func TestDay_SparedPlayer_CannotBeNominatedAgainTheSameDay(t *testing.T) {
	_, d := day(t)
	toVerdict(d, "An")
	for _, name := range all[1:] {
		d.Spare(name)
	}
	d.ExpectPhase(engine.PhaseDay)
	d.NominateExpectErr("Binh", "An", engine.ErrAlreadySpared)
	d.Nominate("Binh", "Giang") // anyone else is fine
}

func TestDay_AfterASpare_NominationsAndSkipsStartOver_ButTheSparedListStays(t *testing.T) {
	_, d := day(t)
	d.Skip("Hoa").Skip("Khoa")
	toVerdict(d, "An")
	for _, name := range all[1:] {
		d.Spare(name)
	}
	// Nominations are empty again: it takes everybody to close them, and the
	// earlier skips are gone: five more skips would have made seven.
	for _, name := range all[2:7] {
		d.Skip(name)
	}
	d.ExpectPhase(engine.PhaseDay)
	d.Nominate("Binh", "Giang").ExpectPhase(engine.PhaseDay) // the earlier nominations are gone
	for _, name := range all {
		if name == "Giang" {
			d.NominateNobody(name)
		} else {
			d.Nominate(name, "Giang")
		}
	}
	d.TimeOut().ExpectOnTrial("Giang").TimeOut()
	for _, name := range all {
		if name != "Giang" {
			d.Spare(name)
		}
	}
	// Both An and Giang were spared today and stay protected.
	d.NominateExpectErr("Binh", "An", engine.ErrAlreadySpared)
	d.NominateExpectErr("Binh", "Giang", engine.ErrAlreadySpared)
}

func TestDay_SparedPlayerCanBeNominatedAgainTheNextDay(t *testing.T) {
	g, d := day(t)
	toVerdict(d, "An")
	for _, name := range all[1:] {
		d.Spare(name)
	}
	d.Skip("Binh")
	for _, name := range all[2:8] {
		d.Skip(name)
	}
	d.ExpectPhase(engine.PhaseDusk)
	g.Night()
	g.Dawn()
	d2 := g.Day()
	d2.Nominate("Binh", "An")
}

func TestDay_DeadPlayerCannotVote(t *testing.T) {
	g := chiefTable(t, "Phuc")
	g.Night().Wolf("An").Bites("Lan").Wolf("Binh").Bites("Lan")
	g.Dawn()
	d := g.Day()
	for _, name := range all[:9] {
		if name == "Giang" {
			d.NominateNobody(name)
		} else {
			d.Nominate(name, "Giang")
		}
	}
	d.TimeOut().TimeOut().ExpectPhase(engine.PhaseVerdict)
	d.VerdictExpectErr("Lan", true, engine.ErrNotAlive)
}

func TestDay_Hanged_PlayerDiesAnnouncedWithoutRole_NightFollows(t *testing.T) {
	g, d := day(t)
	d.Hanged("Giang")
	g.ExpectDead("Giang").ExpectPhase(engine.PhaseDusk)
	d.ExpectAnnounced(i18n.DayHanged)
	g.Night()
	if g.State().Night != 2 {
		t.Fatalf("night = %d, want 2", g.State().Night)
	}
}

func TestDay_ChiefHanged_NamesASuccessorThenTheNightFollows(t *testing.T) {
	g, d := chiefDay(t)
	d.Hanged("Phuc")
	g.ExpectDead("Phuc").ExpectPhase(engine.PhaseChiefSuccession).ExpectNoChief()
	g.NameSuccessor("Phuc", "Em")
	g.ExpectChief("Em").ExpectPhase(engine.PhaseDusk)
	g.Night()
}

func TestDay_ChiefHanged_NoSuccessorInTime_TitleIsLost(t *testing.T) {
	g, d := chiefDay(t)
	d.Hanged("Phuc")
	g.TimeOut()
	g.ExpectNoChief().ExpectPhase(engine.PhaseDusk)
}

func TestDay_NominationCountdown_ChangingAChoiceRestartsIt(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock)
	stale := engine.PhaseTimeout{Phase: g.State().Phase, Seq: g.State().Seq}

	d.Nominate("Binh", "Hoa").Nominate("Chi", "Hoa") // two changes: Hoa now leads
	d.ExpectPhase(engine.PhaseNominationLock)
	g.Do(stale) // the first countdown's timeout no longer matches
	d.ExpectPhase(engine.PhaseNominationLock)

	d.TimeOut().ExpectPhase(engine.PhaseTrial).ExpectOnTrial("Hoa")
}

func TestDay_NominationCountdown_IsTenSecondsByDefault(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	if got := g.State().Settings.NominationLock; got != 10*time.Second {
		t.Fatalf("countdown = %v, want 10s", got)
	}
}

func TestDay_NominationCountdown_NobodyOnTrialUntilItRunsOut(t *testing.T) {
	_, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock) // still nobody on trial
	d.TimeOut().ExpectOnTrial("Giang")
}

func TestDay_NominationCountdown_EveryoneBacksOut_DiscussionContinues(t *testing.T) {
	_, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock)
	d.NominateNobody("An").ExpectPhase(engine.PhaseDay) // nobody is nominated any more
}

func TestDay_NominationCountdown_PlayersCanStillSkip(t *testing.T) {
	_, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	for _, name := range all[1:8] {
		d.Skip(name)
	}
	d.ExpectPhase(engine.PhaseDusk)
}
