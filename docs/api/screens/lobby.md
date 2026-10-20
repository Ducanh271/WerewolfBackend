# Join screen and lobby

Phase `Lobby`: from creating the room until the host deals the roles
(rules §2). Every guide in this folder lists what drives the screen, what
the phone may send, which [errors](../errors.md) can come back, and real
messages from the example games (generated: see `../README.md` "Keeping
this folder true").

## Join screen

Before a socket has joined, the app shows a form: room code (from the QR
link `GET /join/{code}` or typed), nickname, and "I was already in this
room" (nickname + PIN). The flow, and when to stop reconnecting:
`../README.md` "Joining and rejoining".

| send | when |
|---|---|
| `join` `{room, nickname}` (+ `hostKey` on the creator's first join) | a new player |
| `rejoin` `{room, token}` | this phone was in the room before (token saved) |
| `rejoin` `{room, nickname, pin}` | another phone, or the token was lost |

After a successful `join`, show the PIN with a reminder to remember it
(rules §2 "Vào lại phòng"); it is never shown again.

Errors: `room_not_found`, `invalid_nickname`, `nickname_taken`,
`room_full`, `wrong_phase` (situation `join`: the game has started),
`bad_credentials` (situation `token` for a token rejoin), `pin_locked`,
`too_many_connections`, `room_closed`. When the socket is closed with
`kicked_by_host` or `replaced_by_other_device`, come back to this screen.

<!-- example: server/welcome.join.json -->
```json
{
  "type": "welcome",
  "pin": "6663",
  "player": "p1",
  "protocolVersion": 1,
  "token": "00903b35b33dc80d9058e53bea6d93ac"
}
```
<!-- /example -->

## What drives the lobby

| field | use |
|---|---|
| `people` (next to `view`) | everybody in the room in join order: `id`, `nickname`, `connected`, `isHost`. The list to draw; it changes on join, leave, kick, drop, return and `transferHost` |
| `speaker` | `true` on the phone that reads announcements aloud (rules §1 "Máy làm loa") |
| `view.me.id` | this player; compare with `people[].isHost` to show the host's controls |
| `view.players` | empty until the host sends `setSeats`, then the seating order (`seat` 1..n, clockwise) |
| `view.roles`, `view.roleCards` | the public role list once the host configured it, with each role's name, team and description (rules §2 step 3, §9b) |

The current timers and sound mode are not in the view: the host's screen
starts from the defaults of rules §2 "Thời gian mặc định" and §1 "Âm thanh"
and follows `TimersConfigured` (only the times changed, `0` = unchanged)
and `SoundConfigured` in `events`.

<!-- example: server/view.lobby.host.json type speaker people[:3] view.phase view.me view.players -->
```json
{
  "type": "view",
  "people": [
    {
      "connected": true,
      "id": "p1",
      "isHost": true,
      "nickname": "An"
    },
    {
      "connected": true,
      "id": "p2",
      "isHost": false,
      "nickname": "Bình"
    },
    {
      "connected": true,
      "id": "p3",
      "isHost": false,
      "nickname": "Chi"
    }
  ],
  "speaker": true,
  "view": {
    "me": {
      "alive": true,
      "canShoot": false,
      "cursed": false,
      "id": "p1"
    },
    "phase": "Lobby",
    "players": [
      {
        "alive": true,
        "id": "p1",
        "nickname": "An",
        "seat": 1
      },
      {
        "alive": true,
        "id": "p2",
        "nickname": "Bình",
        "seat": 2
      },
      {
        "alive": true,
        "id": "p3",
        "nickname": "Chi",
        "seat": 3
      },
      {
        "alive": true,
        "id": "p4",
        "nickname": "Dũng",
        "seat": 4
      },
      {
        "alive": true,
        "id": "p5",
        "nickname": "Giang",
        "seat": 5
      },
      {
        "alive": true,
        "id": "p6",
        "nickname": "Hà",
        "seat": 6
      },
      {
        "alive": true,
        "id": "p7",
        "nickname": "Khoa",
        "seat": 7
      },
      {
        "alive": true,
        "id": "p8",
        "nickname": "Lan",
        "seat": 8
      },
      {
        "alive": true,
        "id": "p9",
        "nickname": "Minh",
        "seat": 9
      },
      {
        "alive": true,
        "id": "p10",
        "nickname": "Nga",
        "seat": 10
      },
      {
        "alive": true,
        "id": "p11",
        "nickname": "Phúc",
        "seat": 11
      },
      {
        "alive": true,
        "id": "p12",
        "nickname": "Quân",
        "seat": 12
      }
    ]
  }
}
```
<!-- /example -->

## What the phone may send

| send | who | notes |
|---|---|---|
| `leave` | anybody | leaves the room for good; the server closes the socket |
| `kick` `{target}` | host | not oneself; the target's phone gets `kicked_by_host` |
| `transferHost` `{target}` | host | to anybody in the room, not oneself; also allowed during the game |
| `setSeats` `{order}` | host | every player exactly once, in seating order |
| `configureRoles` `{roles}` | host | `{role: count}`, limits in rules §2 |
| `configureTimers` `{timers}` | host | only the times to change, in ms; ranges in rules §2 |
| `configureSound` `{mode}` | host | `oneSpeaker` or `allPhones` |
| `dealRoles` | host | starts the game (next: [role view](role-view.md)) |

Once the roles are dealt, everything above except `transferHost` is
refused (`wrong_phase`; `dealRoles` itself `already_dealt`): from then on a
player who closes the app keeps their seat (rules §2).

## Errors

The host's setup errors say what to fix: `roles_not_configured`,
`seats_not_set`, `seats_mismatch`, `too_few_players`,
`role_count_mismatch`, `no_wolves`, `too_many_wolves`,
`role_needs_more_players`, `role_required`, `timer_out_of_range`,
`already_dealt`. After a `leave` or `kick` the role list may no longer fit:
`dealRoles` then answers with one of them until the host fixes it (rules §2
"Nickname, rời phòng và đá người"). Also `not_host`, `unknown_player`,
`rate_limited`. A correct app never gets `forbidden`, `duplicate_seat`,
`unknown_role`, `invalid_role_count`, `role_must_be_unique`.

<!-- example: server/error.too_many_wolves.json -->
```json
{
  "type": "error",
  "code": "too_many_wolves",
  "message": "too many wolves for this number of players: 4 wolves (+0 to be cursed) at 12 players, allowed at most 3"
}
```
<!-- /example -->

## Events

`PlayerJoined`, `PlayerLeft`, `PlayerKicked`, `SeatsSet`, `RolesConfigured`,
`TimersConfigured`, `SoundConfigured`. None has `say`.
