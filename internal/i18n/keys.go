package i18n

// Keys of the announcements the narrator reads aloud. Scenarios and the
// engine use the keys; the Vietnamese wording lives in vi.go (task T6).
const (
	// DawnSomeoneDied announces who died in the night, without their role.
	DawnSomeoneDied = "dawn.someone_died"
	// DawnProtected: the Guard shielded the player the pack bit.
	DawnProtected = "dawn.protected"
	// DawnNobodyDied: nobody died for any other reason.
	DawnNobodyDied = "dawn.nobody_died"
)

// Chief announcements (rules §7b), read to everybody.
const (
	// ChiefElected: the election result; the name is the new Chief.
	ChiefElected = "chief.elected"
	// ChiefSucceeded: the dead Chief named the new Chief.
	ChiefSucceeded = "chief.succeeded"
	// ChiefLost: the dead Chief named nobody in time.
	ChiefLost = "chief.lost"
)

// Day announcements (rules §4), read to everybody. No role is ever named.
const (
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
)

// Night notices, shown as text on one player's phone (rules §3).
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
	// NoticeNotCursed and NoticeCursed are the two phone check notices. The
	// view carries one boolean; the client picks the text with it.
	NoticeNotCursed = "notice.not_cursed"
	NoticeCursed    = "notice.cursed"
)
