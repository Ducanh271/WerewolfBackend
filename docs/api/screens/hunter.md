# Hunter's shot

Phase `HunterShot`: a Hunter died, whatever the cause. His role is made
public (only the role, never the team: rules §1, §6 "Thợ săn") and he must
shoot a living player before the deadline (15 s by default); if he does not,
nobody is shot. A Hunter shot dead by another Hunter shoots in turn. It
comes after the morning announcement or after a hanging, before the win
check (rules §8).

## What drives the screen

| field | who | use |
|---|---|---|
| `view.shooter` | every phone | the dead Hunter choosing |
| `view.me.canShoot` | the dead Hunter only | show the target picker on this phone |
| `deadline` | every phone | end of the choice |
| `view.players[]` | every phone | the Hunter's seat now carries `role` `Hunter` and his printed `card` (never `team`) |

The Hunter is dead, so his phone has the full view
([secrecy](../secrecy.md#dead-players)). Every other phone waits in
silence (rules §6).

<!-- example: server/view.hunterShot.json type deadline view.phase view.shooter -->
```json
{
  "type": "view",
  "deadline": 1792095450000,
  "view": {
    "phase": "HunterShot",
    "shooter": "p8"
  }
}
```
<!-- /example -->

<!-- example: server/view.hunterShot.hunter.json view.phase view.shooter view.me.canShoot view.full -->
```json
{
  "view": {
    "full": true,
    "me": {
      "canShoot": true
    },
    "phase": "HunterShot",
    "shooter": "p8"
  }
}
```
<!-- /example -->

## What the phone may send

| send | who | notes |
|---|---|---|
| `hunterShoot` `{target}` | the dead Hunter | a living player; no "nobody" choice |

Errors: `not_your_turn` (anybody else), `invalid_target`, `wrong_phase`
after the deadline.

## Events and announcements

All with `say`: `HunterRevealed` ("X là Thợ săn."), then `HunterShot`
("X bắn Y.") or `HunterTimedOut`. `HunterShotStarted` (`hunter`,
`duration`) has none.

<!-- example: events/HunterShot.json -->
```json
{
  "type": "HunterShot",
  "data": {
    "hunter": "p8",
    "target": "p11"
  },
  "say": "Lan bắn Phúc."
}
```
<!-- /example -->
