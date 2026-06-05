CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    email         TEXT UNIQUE NOT NULL,
    telegram_id   TEXT NOT NULL DEFAULT '',
    is_admin      BOOLEAN NOT NULL DEFAULT false,
    is_super_admin BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS otp_codes (
    id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    email      TEXT NOT NULL,
    code       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS otp_email_created_idx ON otp_codes(email, created_at DESC);
CREATE INDEX IF NOT EXISTS sessions_expires_idx   ON sessions(expires_at);
