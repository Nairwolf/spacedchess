CREATE TABLE users (
  id            SERIAL PRIMARY KEY,
  username      TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE card_type_enum AS ENUM ('tactic', 'blunder', 'strategy');

CREATE TABLE cards (
  id            SERIAL PRIMARY KEY,
  user_id       INTEGER NOT NULL REFERENCES users (id),
  card_type     card_type_enum NOT NULL,
  fen           TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ,
  notes         TEXT,
  card_details  JSONB
);

CREATE INDEX idx_cards_user_id ON cards (user_id);

-- id is the cookie value, so it must be an unguessable random token
-- generated with crypto/rand — never a predictable sequence.
 CREATE TABLE sessions (
  id         TEXT PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL
);
