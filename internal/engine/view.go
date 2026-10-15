package engine

import "time"

// The secrecy boundary (architecture "Views"). A client never receives State:
// it receives ViewFor (what its screen shows now) and Visible (the events it
// may know about). What each viewer may see:
//
//   - Every player: the public part (read aloud by the speaker phone, shown as
//     text on the others; at night only the role being called and the
//     duration, never whether the call is real), own role and team, the public role list, seating, who is
//     alive, the Chief, every day vote, and whether it is their turn.
//   - Living wolves (a cursed player too): the pack and its live bite votes.
//   - A role implementing BiteWitness (the Witch): the bitten player, only on
//     her own turn, only while SeesBite holds, and not when visited.
//   - The Visitor's target: the "visited" notice on their turn.
//   - Lovers: the partner, the partner's role and the link colour.
//   - Dead players, and everybody once the game is over (rules §4): all.
//
// Nobody alive ever sees another player's role (lovers excepted), a cause of
// death, or who was bitten, protected or poisoned.

// Viewer is who a view is built for: one player's phone.
type Viewer struct {
	Player PlayerID
}

// PlayerViewer is the phone of the given player.
func PlayerViewer(id PlayerID) Viewer { return Viewer{Player: id} }

// View is one screen's projection of the game.
type View struct {
	Phase Phase
	Seq   int
	Night int

	Roles      map[RoleID]int // the public role list of the game
	Players    []PlayerView   // in seating order
	Chief      PlayerID
	Succession PlayerID // the dead Chief naming a successor, if any
	Votes      PublicVotes
	Winner     *Team // set when the game ended with a winner
	Draw       bool  // the game ended and nobody won

	// Full is true when the viewer sees everything: a dead player, or anybody
	// once the game is over. Then every PlayerView carries Role and Team and
	// Secrets is set.
	Full    bool
	Secrets *Secrets

	NightCall *NightCallView // every phone: the call running now
	Me        *MyView        // player viewers only
	Pack      *PackView      // living wolves, and full views
	Lover     *LoverView     // lovers only
}

// PlayerView is one seat as the viewer may see it. Role and Team are empty
// unless the viewer may know them.
type PlayerView struct {
	ID       PlayerID
	Nickname string
	Seat     int // 1-based, clockwise
	Alive    bool
	Role     RoleID
	Team     Team
}

// PublicVotes are the votes everybody sees (rules §4, §7b).
type PublicVotes struct {
	Chief       map[PlayerID]PlayerID // voter -> candidate, in the election
	Nominations map[PlayerID]PlayerID // voter -> nominee; "" is "no one"
	Skips       []PlayerID            // in seating order
	Defendant   PlayerID
	Verdicts    map[PlayerID]bool // voter -> true for kill
	Spared      []PlayerID        // spared today, in seating order
}

// NightCallView is the cue shown on every phone. It has no real/fake flag on purpose.
type NightCallView struct {
	Role     RoleID
	Duration time.Duration
}

// MyView is the viewer's own card.
type MyView struct {
	ID     PlayerID
	Role   RoleID
	Team   Team
	Alive  bool
	Cursed bool      // on the wolf team though the role is not a wolf role
	Turn   *TurnView // set while the viewer is being called at night
}

// TurnView is what the phone shows during the viewer's own night step.
type TurnView struct {
	Step    RoleID   // the step being called (the pack step for a wolf)
	Visited bool     // the Visitor blocked the viewer tonight: no action
	Bitten  PlayerID // BiteWitness only: the pack's current target
	// BiteShown is set when the viewer is shown the bite information at all, so
	// an empty Bitten then means "nobody is bitten".
	BiteShown bool
}

// PackView is the wolf pack as its members see it.
type PackView struct {
	Members []PlayerID            // wolf team, in seating order, dead ones too
	Votes   map[PlayerID]PlayerID // wolf -> bite target, live during the night
}

// LinkColour is the colour of the lovers' link (rules §7).
type LinkColour string

// Link colours.
const (
	LinkWhite   LinkColour = "white"   // both villagers
	LinkBlack   LinkColour = "black"   // both wolves
	LinkStriped LinkColour = "striped" // one of each: the third team
)

// LoverView is what a lover's phone shows about the pair.
type LoverView struct {
	Partner     PlayerID
	PartnerRole RoleID
	Colour      LinkColour
}

// Secrets is the hidden part of the state, for full views only.
type Secrets struct {
	Pending NightActions
	Lovers  *LoversLink
}
