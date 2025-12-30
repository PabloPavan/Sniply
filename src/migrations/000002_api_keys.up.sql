DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'api_key_scope') THEN
    CREATE TYPE api_key_scope AS ENUM ('read', 'write', 'read_write');
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS api_keys (
  id            TEXT PRIMARY KEY,
  user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name          TEXT NOT NULL DEFAULT '',
  scope         api_key_scope NOT NULL DEFAULT 'read_write',
  token_hash    TEXT NOT NULL UNIQUE,
  token_prefix  TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_created
  ON api_keys (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_api_keys_token_hash
  ON api_keys (token_hash);
