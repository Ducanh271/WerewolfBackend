# History: the games you played

`POST /history` returns the finished games this phone took part in, newest
first, so the app can show "your games" with who held which role and who won.
It needs no socket and no open room: the rooms may be long gone.

## Asking

The phone sends the rejoin tokens it kept (one per room, from `welcome`; see
[joining](../README.md#joining-and-rejoining)), in the body, never in the URL:

<!-- example: http/history.request.json -->
```json
{
  "tokens": [
    "00903b35b33dc80d9058e53bea6d93ac"
  ]
}
```
<!-- /example -->

- 1 to 20 tokens; each 1 to 128 characters. Keep the token of a room after
  leaving it if the player may want its history.
- A token only reads the games its player sat in. A token the server does not
  know finds nothing; the reply never says which tokens were known.
- After a rejoin by nickname + PIN the old token stops working, for history
  too: keep the newest token of each room (rules §2 "Vào lại phòng"). If the
  database is down at that moment, the change is retried with the room's
  next event.
- A player who left or was kicked from a lobby keeps reading the games they
  sat in before: they played them. A game they were not dealt into is not
  theirs.

## Reply

<!-- example: http/history.reply.json games[:1] -->
```json
{
  "games": [
    {
      "aborted": false,
      "draw": false,
      "endedAt": 1792095740000,
      "gameNo": 1,
      "id": 1,
      "mode": "inPerson",
      "nights": 3,
      "players": [
        {
          "alive": true,
          "nickname": "An",
          "player": "p1",
          "role": "RustyKnight",
          "seat": 0,
          "team": "Village",
          "won": true,
          "you": true
        },
        {
          "alive": false,
          "nickname": "Bình",
          "player": "p2",
          "role": "Villager",
          "seat": 1,
          "team": "Village",
          "won": true,
          "you": false
        },
        {
          "alive": true,
          "nickname": "Chi",
          "player": "p3",
          "role": "Guard",
          "seat": 2,
          "team": "Village",
          "won": true,
          "you": false
        },
        {
          "alive": true,
          "nickname": "Dũng",
          "player": "p4",
          "role": "Visitor",
          "seat": 3,
          "team": "Village",
          "won": true,
          "you": false
        },
        {
          "alive": false,
          "nickname": "Giang",
          "player": "p5",
          "role": "Villager",
          "seat": 4,
          "team": "Village",
          "won": true,
          "you": false
        },
        {
          "alive": true,
          "nickname": "Hà",
          "player": "p6",
          "role": "Witch",
          "seat": 5,
          "team": "Village",
          "won": true,
          "you": false
        },
        {
          "alive": true,
          "nickname": "Khoa",
          "player": "p7",
          "role": "Seer",
          "seat": 6,
          "team": "Village",
          "won": true,
          "you": false
        },
        {
          "alive": false,
          "nickname": "Lan",
          "player": "p8",
          "role": "Hunter",
          "seat": 7,
          "team": "Village",
          "won": true,
          "you": false
        },
        {
          "alive": false,
          "nickname": "Minh",
          "player": "p9",
          "role": "CursedWolf",
          "seat": 8,
          "team": "Wolves",
          "won": false,
          "you": false
        },
        {
          "alive": true,
          "nickname": "Nga",
          "player": "p10",
          "role": "Cupid",
          "seat": 9,
          "team": "Village",
          "won": true,
          "you": false
        },
        {
          "alive": false,
          "nickname": "Phúc",
          "player": "p11",
          "role": "WolfSeer",
          "seat": 10,
          "team": "Wolves",
          "won": false,
          "you": false
        },
        {
          "alive": false,
          "nickname": "Quân",
          "player": "p12",
          "role": "Detective",
          "seat": 11,
          "team": "Wolves",
          "won": false,
          "you": false
        }
      ],
      "room": "K7QP",
      "startedAt": 1792094400000,
      "winner": "Village"
    }
  ]
}
```
<!-- /example -->

| field | use |
|---|---|
| `games[]` | at most 50, newest first; only **finished** games, kept 90 days |
| `id` | stable id of the saved game, for the app's own list keys |
| `room`, `gameNo` | room code and the game's number in that room (1, then +1 per `newGame`) |
| `mode` | `inPerson` or `online` |
| `winner` | `Village`, `Wolves` or `Lovers`; absent on a draw |
| `draw`, `aborted` | no winner; `aborted` = called off by vote (rules §8b), always a draw too |
| `nights`, `startedAt`, `endedAt` | nights played; Unix milliseconds of the deal and of the end |
| `players[]` | seating order: `seat`, `player`, `nickname`, `role`, `team` (own team at the end, after a curse), `alive`, `won`, `you` |

Show the roles with the names of [`GET /guide`](guide.md) (`roles[].title`).
A third-team lover keeps `team` `Village` or `Wolves` while `winner` is
`Lovers`: use `won` to mark the winners, not `team` (rules §7, §8).

A server that keeps no history (no database) answers `200` with `games: []`.

## Errors

- `400` `bad_message`: no tokens, more than 20, an empty or too long one, or a
  malformed body.
- `413` `message_too_large`: body over 8 KB.
- `503` `history_unavailable`: the database did not answer; try again later.

<!-- example: http/history.error.bad_message.json -->
```json
{
  "type": "error",
  "code": "bad_message",
  "message": "tokens: 1 to 20 of them"
}
```
<!-- /example -->

Secrecy: [secrecy.md](../secrecy.md#history).
