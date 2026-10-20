# Ma Sói client API (protocol version 1)

This folder is the contract between the server and the phone app (web now,
mobile later). A client can be built from it alone, without reading Go code.
The game rules are in [`../rules.md`](../rules.md) (Vietnamese); this folder
cites their section numbers.

| path | what | how it is made |
|---|---|---|
| `schema/server.schema.json` | every WebSocket message the server sends | generated from the Go types |
| `schema/client.schema.json` | every WebSocket message a client may send | generated from the Go types |
| `schema/http.schema.json` | the bodies of the HTTP endpoints | generated from the Go types |
| `examples/` | real messages from two example games | captured from the real server |
| this README | connecting, the message envelope, versioning | written by hand |
| [`screens/`](screens/lobby.md) | one guide per screen: the view fields that drive it, the commands it may send, its errors, real messages | by hand; the JSON blocks generated from `examples/` |
| [`errors.md`](errors.md) | every error code, when it comes, the sentence to show | by hand; its tables generated from the server's codes and `internal/i18n` |
| [`secrecy.md`](secrecy.md) | what the app must never show, sound, vibrate or animate differently | by hand |

The schemas are standard JSON Schema **draft 2020-12**, so the usual tools
can generate client types from them (`json-schema-to-typescript`, `quicktype`,
...). Each schema is a `oneOf` over its messages, told apart by `type`; the
named shapes are in `$defs` (`View`, `MyView`, `DawnAnnouncedEvent`, ...).
Field descriptions come from the server's own comments.

The examples come from two games played in a test through the real server:
room `K7QP`, players An, Bình, Chi, Dũng, Giang, Hà, Khoa, Lan, Minh, Nga,
Phúc, Quân (plus Tú, who leaves, and Vy, who is kicked), every role in play.
Game 1 runs to a village win; game 2 (a new game in the same room) ends with
the table voting to call it off.

- `examples/client/`: one file per client message and per command name
  (`command.<name>.json`).
- `examples/server/`: `welcome.*`, `view.<screen>.json` (one per screen and
  role, e.g. `view.night.witch.json`), `events.reset.json` and
  `events.batch.json`, `error.<code>.json`.
- `examples/events/<Type>.json`: one item of an `events` list per event type.
- `examples/http/`: request and reply bodies.

Screens, in the order of a game: [join and lobby](screens/lobby.md),
[role view](screens/role-view.md), [Chief](screens/chief.md),
[night](screens/night.md), [phone check](screens/phone-check.md),
[dawn](screens/dawn.md), [day](screens/day.md),
[Hunter's shot](screens/hunter.md), [end](screens/end.md), and the
[rules guide](screens/guide.md). Read [secrecy.md](secrecy.md) before
designing any of them.

## Connecting

All JSON. Field names are camelCase, durations are milliseconds, times are
Unix milliseconds.

| request | reply |
|---|---|
| `POST /rooms`, body optional `{"mode": "inPerson"}` | `201` `{"room", "hostKey"}` (`http/create-room.reply.json`). Keep `hostKey` in memory: it makes this phone the host on its first `join`, and nobody else ever sees it. |
| `GET /join/{code}` | `200` `{"room"}` when the room exists (case ignored); `404` `room_not_found`. This is the link a QR code carries. |
| `GET /guide` | `200` the rules guide and the "how to play your role" texts, in Vietnamese (`http/guide.reply.json`). Public, needs no room. |
| `GET /ws` | the game WebSocket. **No query string** (`400` `secret_in_url`): secrets only travel inside messages. |

An HTTP error has the same body as a WebSocket `error` message
(`HTTPError` in `http.schema.json`). The browser must send an `Origin` the
server is configured to accept (`403` `bad_origin` otherwise); too many
sockets get `503` `too_many_connections`.

## Joining and rejoining

The first message on a new socket must be `join` or `rejoin`, within 10
seconds of opening it.

1. **First time in a room**: `join` with `room` and `nickname` (the creator
   adds `hostKey`). Reply: `welcome` with `player` (your public id), `token`
   and `pin`, then a `view` and an `events` message. Store the token with the
   room code (e.g. `localStorage`); show the PIN to the player and ask them
   to remember it (rules §2 "Vào lại phòng"). Refusals: `invalid_nickname`,
   `nickname_taken`, `room_full`, `room_not_found`.
2. **Same phone again** (reload, lost signal, app reopened): `rejoin` with
   `room` and `token`. Reply: `welcome` with `player` only, then `view` and
   `events`.
3. **Another phone**: `rejoin` with `room`, `nickname` and `pin`. Reply:
   `welcome` with a **new** `token` (store it; the old one stops working),
   then `view` and `events`. Wrong PIN: `bad_credentials`; too many wrong
   PINs: `pin_locked` with `retryAfter` (ms).

`welcome` always comes first. A dropped connection never removes the player:
reconnect and `rejoin` with the token, with a growing delay between tries.
Stop reconnecting and show the join screen when:

- the socket got `error` `replaced_by_other_device` before closing: another
  phone took this player over
  (`examples/server/error.replaced_by_other_device.json`);
- the socket got `error` `kicked_by_host`: the host removed them from the
  lobby (`error.kicked_by_host.json`);
- a token `rejoin` gets `bad_credentials` (the token was replaced or
  revoked) or `room_not_found` (the room expired).

`leave` (lobby only) removes the player for good and the server closes the
socket. A socket that fails to join or rejoin 8 times in a row is closed.
The server pings every socket about every 12 seconds; browsers answer on
their own, and a socket that does not answer within 10 seconds is closed.

## Messages

### Client → server

`join`, `rejoin`, `leave`, `kick`, `transferHost`, and `command`:

```json
{"type": "command", "name": "nominate", "args": {"target": "p5"}}
```

The server knows who sends a message from the socket: no message ever names
its sender. Unknown keys, wrong types and wrong capitals are refused with
`bad_message`; an unknown `type` is `unknown_type`, an unknown command
`unknown_command`. Commands without args (`dealRoles`, `newGame`,
`witchHeal`) may omit `args`. Every command and its args:
`client.schema.json` (`<Name>Command`, `<Name>Args`) and
`examples/client/command.<name>.json`.

A refused command changes nothing and gets an `error` back on that socket
only. An accepted command gets no reply of its own: its effect arrives as
the next `view`/`events`.

### Server → client

- **`welcome`**: first message after a successful `join`/`rejoin`.
- **`view`**: the whole screen now. Each one **replaces** the previous one;
  build the screen from the latest `view` alone. Next to the `view` object:
  `deadline` (when the running phase or night step ends), `speaker: true` on
  the phone that reads the announcements aloud (rules §1 "Máy làm loa"), and
  `people` (everybody in the room, host, who is connected) in the lobby and
  once the game has ended. A phone whose screen did not change gets
  **nothing**: a message on every phone while one player acts at night
  would tell the table that this role is alive (rules §3).
- **`events`**: what happened, in order, as this screen may know it.
  `reset: true` means `events` is the **whole list and replaces** the one
  the phone holds (on join, when the player dies, when the game ends: hidden
  events are put back in their places, so indexes shift). Never append a
  reset. Without reset, `from` is the length of the list the phone already
  holds and `events` continues it. An event with `say` is an announcement:
  the `speaker` phone reads it aloud, every phone shows it as text.
- **`error`**: `code` is stable snake_case, `message` is an English detail
  for logs, never for players. `retryAfter` (ms) only with `pin_locked`.
  Handle codes you do not know. Every code and the sentence to show:
  [`errors.md`](errors.md).

## Limits

A message over 4 KB is refused (`message_too_large`; over 64 KB the socket
is closed). More than 10 messages a second (bursts of 20 allowed) are
refused with `rate_limited`; staying over the limit for 3 seconds closes the
socket. Details and the reasons behind them: [`../protocol.md`](../protocol.md).

## Versioning

`welcome.protocolVersion` is **1**. It goes up only on a change an older
client cannot ignore (a field removed or renamed, a meaning changed). An app
compares it with the version it was built for and, when the server's is
higher, tells the player to update the app instead of playing on.

Within one version the server may add, without notice: fields, event types,
commands, error codes and enum values (a phase, a role). Clients must ignore
fields and events they do not know and treat an unknown error code like any
other refusal. The schemas describe the current server exactly
(`additionalProperties: false`), so generate types from them, but do not
use them to reject the server's messages at run time.

## Changes

Changes to the wire within version 1, newest last. The ones made before
any client existed did not raise `protocolVersion`.

- `SoundConfigured.data.mode` is the mode's name, `"oneSpeaker"` or
  `"allPhones"` (the strings `configureSound` takes); it was a number.

## Keeping this folder true

`make api-docs` rewrites `schema/` and `examples/` from the server, and the
marked regions of the Markdown files here (`<!-- example: ... -->`,
`<!-- generated: ... -->`): edit around them, never inside. The same
test runs in `make test` and `make test-fast` without rewriting anything and
fails when a file here is stale, so a server change cannot leave this folder
behind. It also checks that the schemas are valid draft 2020-12, that every
message of the example games matches them, and that the games send every
client message type, every command and every event type.
