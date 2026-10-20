# Role view

Phase `RoleView`: right after `dealRoles`, every player looks at their own
role for a countdown (rules §2 step 5); when it runs out the game moves on
by itself to the [Chief election](chief.md).

## What drives the screen

| field | use |
|---|---|
| `deadline` | when the role view ends |
| `view.me.roleCard` | the card to show: `name`, `description` of the role, `team`, `goal` of the player's side (rules §9b) |
| `view.me.role`, `view.me.team` | the same, as ids, for the app's own logic |
| `view.roleCards` | the public role list of the game, for a "roles in this game" screen |
| `view.players` | seats; the player's own seat also carries `role`, `team` and `card` |

Show the card at a **fixed size, covered, revealed only while pressed**
(rules §9b; [secrecy](../secrecy.md#role-cards)). The same card stays
available for the whole game, behind the same press-and-hold.

<!-- example: server/view.roleView.json type deadline view.phase view.me -->
```json
{
  "type": "view",
  "deadline": 1792094430000,
  "view": {
    "me": {
      "alive": true,
      "canShoot": false,
      "cursed": false,
      "id": "p2",
      "role": "Villager",
      "roleCard": {
        "description": "Không có năng lực ban đêm. Vũ khí của bạn là lá phiếu và khả năng suy luận ban ngày.",
        "goal": "Dân thắng khi không còn con Sói nào.",
        "name": "Dân làng",
        "team": "Dân"
      },
      "team": "Village"
    },
    "phase": "RoleView"
  }
}
```
<!-- /example -->

## What the phone may send

No game command. `transferHost` (host) and the limits of
`../README.md` still apply.

## Errors

Any command now is `wrong_phase`.

## Events

`RolesDealt` is never sent to the living (it holds every role).
`RoleViewStarted` has `say` ("Mời mọi người xem vai của mình. Nhớ che màn
hình nhé.").

<!-- example: events/RoleViewStarted.json -->
```json
{
  "type": "RoleViewStarted",
  "data": {
    "duration": 30000
  },
  "say": "Mời mọi người xem vai của mình. Nhớ che màn hình nhé."
}
```
<!-- /example -->
