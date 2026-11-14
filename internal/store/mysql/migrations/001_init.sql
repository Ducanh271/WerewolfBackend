-- docs/database.md "Schema". Never edit once pushed: add 002_*.sql instead.

CREATE TABLE rooms (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  log_key    VARCHAR(40) NOT NULL,
  code       VARCHAR(8)  NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY rooms_log_key (log_key),
  KEY rooms_code (code, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE room_events (
  room_id    BIGINT UNSIGNED NOT NULL,
  seq        INT UNSIGNED    NOT NULL,
  type       VARCHAR(64)     NOT NULL,
  data       JSON            NOT NULL,
  created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (room_id, seq),
  CONSTRAINT room_events_room FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE games (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  room_id      BIGINT UNSIGNED NOT NULL,
  game_no      INT UNSIGNED    NOT NULL,
  mode         VARCHAR(16)     NOT NULL,
  winner       VARCHAR(16)     NOT NULL,
  draw         BOOLEAN         NOT NULL,
  aborted      BOOLEAN         NOT NULL,
  nights       INT UNSIGNED    NOT NULL,
  player_count INT UNSIGNED    NOT NULL,
  first_seq    INT UNSIGNED    NOT NULL,
  last_seq     INT UNSIGNED    NOT NULL,
  started_at   DATETIME(3)     NOT NULL,
  ended_at     DATETIME(3)     NOT NULL,
  UNIQUE KEY games_room_no (room_id, game_no),
  KEY games_ended (ended_at),
  CONSTRAINT games_room FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE game_players (
  game_id   BIGINT UNSIGNED NOT NULL,
  player_id VARCHAR(16)     NOT NULL,
  seat      INT UNSIGNED    NOT NULL,
  nickname  VARCHAR(80)     NOT NULL,
  role      VARCHAR(32)     NOT NULL,
  team      VARCHAR(16)     NOT NULL,
  alive     BOOLEAN         NOT NULL,
  won       BOOLEAN         NOT NULL,
  PRIMARY KEY (game_id, player_id),
  CONSTRAINT game_players_game FOREIGN KEY (game_id) REFERENCES games (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
