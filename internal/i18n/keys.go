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
