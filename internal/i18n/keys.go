package i18n

// Keys of the announcements the narrator reads aloud. Scenarios and the
// engine use the keys; the Vietnamese wording lives in vi.go (task T6).
const (
	// DawnStarted opens the morning, before the narrator says who died (audit
	// [L3]).
	DawnStarted = "dawn.started"
	// DawnSomeoneDied announces who died in the night, without their role.
	DawnSomeoneDied = "dawn.someone_died"
	// DawnProtected: the Guard shielded the player the pack bit.
	DawnProtected = "dawn.protected"
	// DawnNobodyDied: nobody died for any other reason.
	DawnNobodyDied = "dawn.nobody_died"
)

// RoleViewStarted opens the role-view countdown right after the deal (rules
// §2 bước 5). The rule itself asks for no voice line here; this opener was
// chosen on top of it, not a reading of the rule.
const RoleViewStarted = "roleview.started"

// Night announcements (rules §3), read to everybody. NightStepCalled never
// says whether the role being called is alive, dead or absent (gọi giả).
const (
	// NightStarted opens the night (audit [L3]).
	NightStarted = "night.started"
	// NightStepCalled: "<role>, open your eyes..."; the role's calling name
	// (never a player) is the one placeholder. engineSlots's own steps (the
	// phone check, the lovers looking at each other) read their own line
	// instead and never use this key (see i18n.NightCallName).
	NightStepCalled = "night.step_called"
)

// Chief announcements (rules §7b), read to everybody.
const (
	// ChiefElectionStarted opens the public vote (audit [L3]).
	ChiefElectionStarted = "chief.election_started"
	// ChiefElected: the election result; the name is the new Chief.
	ChiefElected = "chief.elected"
	// ChiefSucceeded: the dead Chief named the new Chief.
	ChiefSucceeded = "chief.succeeded"
	// ChiefLost: the dead Chief named nobody in time.
	ChiefLost = "chief.lost"
)

// Day announcements (rules §4), read to everybody. No role is ever named.
const (
	// DayStarted opens discussion and nominations (audit [L3]).
	DayStarted = "day.started"
	// TrialStarted: the top nominee goes up for their defense (audit [L3]).
	TrialStarted = "day.trial_started"
	// VerdictStarted: defense is over, the kill-or-spare vote opens (audit
	// [L3]).
	VerdictStarted = "day.verdict_started"
	// DayHanged: the village killed the player on trial.
	DayHanged = "day.hanged"
	// DaySpared: the village spared the player on trial.
	DaySpared = "day.spared"
	// DayNobodyHanged: the day ended without an execution.
	DayNobodyHanged = "day.nobody_hanged"
)

// Hunter announcements (rules §6), read to everybody.
const (
	// HunterRevealed: the dead Hunter's name; the role is public from now on.
	HunterRevealed = "hunter.revealed"
	// HunterShot: "<Hunter> bắn <target>".
	HunterShot = "hunter.shot"
	// HunterTimedOut: the 15 seconds ran out, nobody is shot (audit [L3]).
	HunterTimedOut = "hunter.timed_out"
)

// Game over announcements (rules §8), read to everybody.
const (
	// WinVillage: no wolf is left.
	WinVillage = "win.village"
	// WinWolves: the wolves reached the villagers' number.
	WinWolves = "win.wolves"
	// WinLovers: only the cross-team couple is left.
	WinLovers = "win.lovers"
	// WinDraw: everybody died at the same moment.
	WinDraw = "win.draw"
	// GameAborted: the table voted to call the game off (audit K1, rules §8b).
	GameAborted = "win.aborted"
)

// Summary clauses (audit K4): the end-of-game narration, one line per night
// and day. Unlike the announcements above, these name the secret actor and
// target of a night step (who the pack bit, who the Witch healed...); they
// are only ever shown once the game has ended, when everybody already sees
// everything, and never read aloud during play.
const (
	SummaryBite            = "summary.bite"
	SummaryHeal            = "summary.heal"
	SummaryWitchPoison     = "summary.witch_poison"
	SummaryGuardProtect    = "summary.guard_protect"
	SummarySeerWolf        = "summary.seer_wolf"
	SummarySeerVillager    = "summary.seer_villager"
	SummaryWolfSeerIsSeer  = "summary.wolfseer_is_seer"
	SummaryWolfSeerNotSeer = "summary.wolfseer_not_seer"
	SummaryDetectiveWolf   = "summary.detective_wolf"
	SummaryDetectiveClear  = "summary.detective_clear"
	SummaryVisitorBlock    = "summary.visitor_block"
	SummaryCurse           = "summary.curse"
	SummaryKnightPoison    = "summary.knight_poison"
	SummaryCupidPair       = "summary.cupid_pair"
	// SummaryNightLabel and SummaryDayLabel are the "Đêm N" / "Ngày N" line
	// headers; the number is passed as a string (i18n.Vi only substitutes
	// "%s").
	SummaryNightLabel = "summary.night_label"
	SummaryDayLabel   = "summary.day_label"
	// SummaryDawnDeaths and SummaryDawnNobodyDied are the night's outcome,
	// read after "→ ". They say only who died, never why (that is already in
	// the clauses before the arrow): a Guard's or Witch's save and an
	// uneventful night both read as "Không ai chết."
	SummaryDawnDeaths     = "summary.dawn_deaths"
	SummaryDawnNobodyDied = "summary.dawn_nobody_died"
	// SummaryAborted is the final line of a game the table voted to call off
	// (audit K1): shorter than the live announcement (i18n.GameAborted),
	// which the summary does not reuse.
	SummaryAborted = "summary.aborted"
)

// Night notices, shown as text on one player's phone (rules §3). The view
// sends each text next to the flag it spells out (engine/view_notices.go;
// me.turn.notice, me.phoneCheck.cursedText/loveText, pack.doomedNotices in
// docs/protocol.md), so no client writes its own.
const (
	// NightBlocked: the player cannot use their night ability. It never names
	// who blocked them.
	NightBlocked = "night.blocked"
	// NightWolfDoomed: the wolf poisoned by the Rusty knight is dead from now
	// on and cannot vote. Shown on its own phone at the pack step.
	NightWolfDoomed = "night.wolf_doomed"
	// NightGrief: the lover whose partner died by day is dead from their first
	// call of the night; shown on their own phone (rules §7).
	NightGrief = "night.grief"
	// NoticeNoLove and NoticeInLove are the love part of the phone check notice;
	// the second takes the partner's name and role.
	NoticeNoLove = "notice.no_love"
	NoticeInLove = "notice.in_love"
	// NightPackDoomed: the pack is told a wolf can no longer vote, without
	// saying why. The name is the wolf's.
	NightPackDoomed = "night.pack_doomed"
	// NightPhoneCheck is read aloud at the phone check step (rules §3).
	NightPhoneCheck = "night.phone_check"
	// NightLoversMeet is read aloud at the lovers' step of night 1 (rules §3).
	NightLoversMeet = "night.lovers_meet"
	// NoticeNotCursed and NoticeCursed are the two curse lines of the phone
	// check notice.
	NoticeNotCursed = "notice.not_cursed"
	NoticeCursed    = "notice.cursed"
)
