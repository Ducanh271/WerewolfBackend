# Chief: election and succession

The Chief is a public title, not a role (rules §7b). Two phases:

- `ChiefElection`: after the role view, before night 1. Everybody votes in
  public for one living player (oneself allowed) until the deadline.
- `ChiefSuccession`: the Chief died; the dead Chief names a living
  successor before the deadline, or the title is lost. It comes after a
  dead Hunter's shot and after the win check (rules §7b, §8), so it never
  runs once the game has ended.

## Election

| field | use |
|---|---|
| `deadline` | end of the vote |
| `view.votes.chief` | `{voter: candidate}`, live, public |
| `view.players` | candidates (the living) |

| send | who | notes |
|---|---|---|
| `voteChief` `{candidate}` | every player | a living player, oneself allowed; sending again changes the vote |

Errors: `wrong_phase` after the deadline; `invalid_target` for a dead or
unknown candidate (a correct app offers only the living).

When the vote ends: `ChiefElected` (with `say`), then `view.chief` holds the
Chief and the Chief's own `me.statusCards` has a `chief` card. Ties and an
empty vote are drawn by the server (rules §7b).

<!-- example: server/view.chiefElection.json type deadline view.phase view.votes -->
```json
{
  "type": "view",
  "deadline": 1792094490000,
  "view": {
    "phase": "ChiefElection",
    "votes": {
      "chief": {
        "p1": "p8",
        "p10": "p8",
        "p11": "p8",
        "p12": "p8",
        "p2": "p2",
        "p3": "p8",
        "p4": "p8",
        "p5": "p8",
        "p6": "p8",
        "p7": "p8",
        "p8": "p8",
        "p9": "p8"
      }
    }
  }
}
```
<!-- /example -->

## Succession

| field | use |
|---|---|
| `deadline` | end of the choice |
| `view.succession` | the dead Chief choosing |
| `view.me.id == view.succession` | show the successor picker on this phone only |

| send | who | notes |
|---|---|---|
| `nameSuccessor` `{successor}` | the dead Chief | a living player |

Errors: `not_your_turn` from anybody else; `invalid_target`; `wrong_phase`
after the deadline.

Events: `ChiefSuccessionStarted` (`former`, `duration`), then
`ChiefSucceeded` (with `say`) or `ChiefLost` (with `say`) when time runs
out.

The dead Chief's phone has the full view (rules §4: the dead see
everything), so the picker sits on a screen that also shows every role:
[secrecy](../secrecy.md#dead-players).

<!-- example: server/view.chiefSuccession.json type deadline view.phase view.succession -->
```json
{
  "type": "view",
  "deadline": 1792095450000,
  "view": {
    "phase": "ChiefSuccession",
    "succession": "p8"
  }
}
```
<!-- /example -->

<!-- example: server/view.chiefSuccession.chief.json view.me view.succession view.full -->
```json
{
  "view": {
    "full": true,
    "me": {
      "alive": false,
      "canShoot": false,
      "cursed": false,
      "id": "p8",
      "role": "Hunter",
      "roleCard": {
        "description": "Khi chết vì bất kỳ lý do gì, bạn lộ vai và bắt buộc bắn một người còn sống.",
        "goal": "Dân thắng khi không còn con Sói nào.",
        "name": "Thợ săn",
        "team": "Dân"
      },
      "team": "Village"
    },
    "succession": "p8"
  }
}
```
<!-- /example -->
