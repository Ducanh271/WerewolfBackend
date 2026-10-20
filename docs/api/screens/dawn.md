# Dawn

Phase `Dawn`: the night is over and the speaker says who died (rules §4
"Thông báo buổi sáng"). Then, in this order and only when they apply: a
dead [Hunter](hunter.md) shoots, the win check runs (rules §8; the game may
[end](end.md) here), a dead Chief names a [successor](chief.md), and the
[day](day.md) starts.

## What drives the screen

| field | use |
|---|---|
| `events` | `DawnStarted` (`say` "Trời đã sáng, mọi người mở mắt."), then `DawnAnnounced` with `say`: show the sentence |
| `DawnAnnounced.data.kind` | `deaths` (with `dead`, in seat order), `protected` or `nobodyDied` |
| `view.players[].alive` | the dead of the night are now dead on every screen |
| `deadline` | end of the announcement |

Deaths are named in seat order, never with cause or role (rules §4). A
dead player now gets the full view and an `events` message with
`reset: true`: replace the list, never append it (`../README.md`).

<!-- example: server/view.dawn.json type deadline view.phase view.night view.chief -->
```json
{
  "type": "view",
  "deadline": 1792094790000,
  "view": {
    "chief": "p8",
    "night": 1,
    "phase": "Dawn"
  }
}
```
<!-- /example -->

<!-- example: events/DawnAnnounced.json -->
```json
{
  "type": "DawnAnnounced",
  "data": {
    "kind": "nobodyDied"
  },
  "say": "Không có ai chết."
}
```
<!-- /example -->

## What the phone may send

No game command. Errors: `wrong_phase`.
