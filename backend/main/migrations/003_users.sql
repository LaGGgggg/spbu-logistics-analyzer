CREATE TABLE IF NOT EXISTS users (
    id          text PRIMARY KEY,
    email       text NOT NULL UNIQUE,
    name        text NOT NULL,
    password_hash text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    id          text PRIMARY KEY,
    user_id     text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions(user_id);

CREATE TABLE IF NOT EXISTS auth_audit (
    id          bigserial PRIMARY KEY,
    user_id     text,
    email       text NOT NULL,
    action      text NOT NULL CHECK (action IN ('register', 'login', 'login_failed', 'logout')),
    ip          text NOT NULL,
    at          timestamptz NOT NULL DEFAULT now()
);
