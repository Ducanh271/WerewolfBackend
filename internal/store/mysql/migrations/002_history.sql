-- docs/database.md "History" (T33b): who may read a finished game, and the
-- 90-day purge. Never edit once pushed: add 003_*.sql instead.

ALTER TABLE game_players
  ADD COLUMN token_hash CHAR(64) NOT NULL DEFAULT '' AFTER won,
  ADD KEY game_players_token (token_hash);

ALTER TABLE rooms
  ADD KEY rooms_created (created_at);
