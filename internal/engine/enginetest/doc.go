// Package enginetest is the scenario kit for engine tests. A scenario reads
// like a game told at the table:
//
//	g := NewGame(t,
//		Seats("An", "Binh", "Chi", "Dung", "Em"),
//		Cast("An", Seer, "Binh", Werewolf, "Chi", Villager, "Dung", Villager, "Em", Villager),
//	)
//	g.Night().Seer("An").Checks("Binh")
//	g.Dawn().ExpectNobodyDied()
//	g.Day().NobodyHanged()
//	g.Night()
//
// The kit enforces the shape of a game: the table must obey rules §2, and
// nights and days must alternate (Night, Dawn, Day, Night, ...). Role verbs such
// as Seer("An").Checks(...) are added by each role's own task in its own file;
// they only need Night.Do and Night.Player.
package enginetest
