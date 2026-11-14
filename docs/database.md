# Database (MySQL)

The server keeps every room's event log and a record of every finished game in
MySQL. Without `MASOI_MYSQL_DSN` it runs on the in-memory store, as before.

## What is stored, and why this shape

The game is its event log (architecture rule 5: `Decide` → events → `Evolve`).
So the database stores **the events, unchanged**, and everything else is derived:

- `room_events`: every event of every room, in order. Folding a room's events
  through `engine.Fold` gives back its exact `State`: the game can be replayed,
  audited, or (later, T33c) restored after a restart.
- `games` + `game_players`: one row per finished game and one per player in it,
  written when `GameEnded` is applied. They are a summary for queries (history,
  statistics) and can always be rebuilt from `room_events`.

There are **no player accounts** (rules §10: Google login is out of scope). A
player exists only inside a room, as `PlayerID` + nickname. The same nickname in
two games is not known to be the same person, so the database keeps no
cross-game player table.

Not stored: rejoin tokens and PINs (they live in the room's memory only), and
nothing that is not already an event.

## Schema

Migrations live in `internal/store/mysql/migrations/NNN_name.sql`, are embedded
in the binary and applied in order at start (`schema_version` remembers which
ran). A migration is never edited once pushed; a change is a new file.

```sql
CREATE TABLE rooms (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  log_key    VARCHAR(40) NOT NULL UNIQUE, -- room code + '-' + random: codes are reused
  code       CHAR(4)     NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY rooms_code (code, created_at)
);

CREATE TABLE room_events (
  room_id    BIGINT UNSIGNED NOT NULL,
  seq        INT UNSIGNED    NOT NULL,   -- 0-based position in the room's log
  type       VARCHAR(64)     NOT NULL,   -- Go type name, e.g. "RolesDealt"
  data       JSON            NOT NULL,   -- the event's fields
  created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (room_id, seq),
  FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE
);

CREATE TABLE games (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  room_id      BIGINT UNSIGNED NOT NULL,
  game_no      INT UNSIGNED    NOT NULL, -- 1, then +1 for each newGame in the room
  mode         VARCHAR(16)     NOT NULL, -- inPerson | online
  winner       VARCHAR(16)     NOT NULL, -- Village | Wolves | Lovers; '' on a draw
  draw         BOOLEAN         NOT NULL, -- no winner: everybody died at once, or aborted
  aborted      BOOLEAN         NOT NULL, -- ended by an abort vote (rules §8b); always a draw too
  nights       INT UNSIGNED    NOT NULL,
  player_count INT UNSIGNED    NOT NULL,
  first_seq    INT UNSIGNED    NOT NULL, -- this game's events are room_events
  last_seq     INT UNSIGNED    NOT NULL, --   first_seq..last_seq of the room
  started_at   DATETIME(3)     NOT NULL, -- roles dealt
  ended_at     DATETIME(3)     NOT NULL,
  UNIQUE KEY games_room_no (room_id, game_no),
  KEY games_ended (ended_at),
  FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE
);

CREATE TABLE game_players (
  game_id   BIGINT UNSIGNED NOT NULL,
  player_id VARCHAR(16)     NOT NULL,
  seat      INT UNSIGNED    NOT NULL, -- 0-based, seating order
  nickname  VARCHAR(80)     NOT NULL, -- engine.MaxNicknameLen runes
  role      VARCHAR(32)     NOT NULL, -- RoleID
  team      VARCHAR(16)     NOT NULL, -- own team at the end (after a curse); a third-team lover
                                      -- keeps Village or Wolves (rules §7), see games.winner and won
  alive     BOOLEAN         NOT NULL,
  won       BOOLEAN         NOT NULL,
  PRIMARY KEY (game_id, player_id),
  FOREIGN KEY (game_id) REFERENCES games (id) ON DELETE CASCADE
);
```

Every table is `utf8mb4` with the binary collation `utf8mb4_bin`: nicknames are
stored exactly as the engine accepted them (NFC), and "An" and "Ân" stay two
different names.

## Event encoding (`internal/store/codec.go`)

- `type` is the event's Go type name, `data` its JSON (`encoding/json`, field
  names as in Go). Every event type is registered with `engine.RegisterEvents`:
  the engine's own in `engine/event_types.go`, each role's in its own file, so a
  new role does not edit any other file. Decoding an unknown `type` is an error.
- The stored format is a contract with every log already written:
  - a field or event is **renamed or removed only with a migration** that rewrites old rows;
  - enum values (`Team`, `Phase`, `Mode`, …) are stored as numbers, so they are **only ever appended**, never reordered.
  - `store/testdata/events.golden.json` pins the encoding, and its test fails on any such change.
- `TestCodec_RoundTripsEverySimulatedGame` encodes and decodes the full logs of
  simulated games and checks that folding the decoded log gives the same `State`.

## Write path

1. `room.apply`: `Decide` → `store.Append(logKey, events)` → fold → broadcast,
   as today (architecture: "persists every event before broadcasting"). The
   MySQL store appends with one multi-row INSERT (atomic); on error the command
   fails and nothing changes. A player's command gets an internal error; a
   timeout is logged and tried again after 1 s (`room/timers.go`), since a
   phase only ends by its timeout.
2. When the events contain `GameEnded`, the room builds a `store.GameRecord`
   from its state and calls `GameArchive.SaveGame`. A failure is logged and does
   not fail the command: the events are already safe and the record can be
   rebuilt from them.

`room.Manager` gives every room the shared store and a unique `log_key`. The
`rooms` row is created by the first `Append`.

## Running it

- Server: `MASOI_MYSQL_DSN='masoi:secret@tcp(127.0.0.1:3306)/masoi?parseTime=true' make run`.
  Migrations run at start; the server does not start if they fail.
- Tests that need MySQL read `MASOI_MYSQL_DSN` (a database they may empty) and
  are skipped without it and under `-short`. `make mysql-test-up` starts a
  throwaway MySQL 8.4 in Docker on port 3307; `make test-mysql` runs the store
  tests against it. CI runs them against a MySQL service container.

## Later (see `docs/plan.md` T33)

- T33b History API: finished games over HTTP for the app, with a contract in
  `docs/api/` first. Who may read which game is a secrecy decision for the
  project owner (`docs/api/secrecy.md`).
- T33c Restore live rooms after a restart (needs the identity layer to survive
  too; goes with T29e graceful shutdown).
- T33d App screen for the history, after T33b.
