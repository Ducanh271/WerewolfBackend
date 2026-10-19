package engine

// Table limits of rules §2. Everything you may want to tune when the rules
// change is in this file; validateRoleConfig (config.go) only applies it.
//
// Other setup limits live with their role instead of here: the minimum table
// size of Detective and Cupid is the role's MinPlayers(), and which roles may
// have several copies is the role's AllowsMultipleCopies() (see registry.go).
const (
	// MinPlayers and MaxPlayers bound a game ("Một ván có 5–30 người").
	MinPlayers = 5
	MaxPlayers = 30

	// MaxWolfPercent is the share of players that may be wolves, rounded down
	// (rules §2 "Giới hạn số sói"). Every wolf role counts.
	MaxWolfPercent = 30

	// CurseWolfReserve is how many extra wolf slots a CursedWolf takes from the
	// limit, for the villager it will curse later ("tính cả người sẽ bị nguyền").
	CurseWolfReserve = 1
)

// MaxWolves returns the highest number of wolves allowed at a table of
// playerCount, before the curse reserve is added to the wolves picked.
//
// The percentage floor reproduces every row of the rules §2 table and also
// covers the table's gaps: 5-6 -> 1, 7-9 -> 2, 10-13 -> 3, 14-16 -> 4,
// 20 -> 6, 30 -> 9. With a CursedWolf the room for other wolves is one less,
// so at 10-13 players a CursedWolf and a WolfSeer fit without a plain Werewolf.
func MaxWolves(playerCount int) int {
	return playerCount * MaxWolfPercent / 100
}
