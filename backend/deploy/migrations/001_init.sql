CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Dynamic rooms management
CREATE TABLE IF NOT EXISTS rooms (
    number          INTEGER PRIMARY KEY,
    name            TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    weekday_open    INTEGER NOT NULL DEFAULT 6,
    weekday_close   INTEGER NOT NULL DEFAULT 23,
    frisat_open     INTEGER NOT NULL DEFAULT 6,
    frisat_close    INTEGER NOT NULL DEFAULT 25,
    sun_open        INTEGER NOT NULL DEFAULT 6,
    sun_close       INTEGER NOT NULL DEFAULT 23,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO rooms (number, name, weekday_open, weekday_close, frisat_open, frisat_close, sun_open, sun_close) VALUES
    (21,   'Комната 21 (1 корпус)',   6, 23, 6, 25, 6, 23),
    (256,  'Комната 256 (1 корпус)',  6, 23, 6, 25, 6, 23),
    (132,  'Комната 132 (2 корпус)',  6, 22, 6, 23, 6, 22),
    (2812, 'Комната 812 (2 корпус)',  6, 23, 6, 25, 6, 23),
    (3812, 'Комната 812 (3 корпус)',  6, 23, 6, 25, 6, 23)
ON CONFLICT DO NOTHING;

-- Email-based users
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    telegram_id TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- OTP codes for email auth
CREATE TABLE IF NOT EXISTS otp_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT NOT NULL,
    code       TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS otp_codes_lookup ON otp_codes(email, expires_at) WHERE NOT used;

-- Web sessions
CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS sessions_expires ON sessions(expires_at);

-- Admin email whitelist
CREATE TABLE IF NOT EXISTS admins (
    email      TEXT PRIMARY KEY,
    added_by   TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO admins (email, added_by) VALUES ('irnosov@edu.hse.ru', 'system') ON CONFLICT DO NOTHING;

-- Bookings (room is FK to rooms table)
CREATE TABLE IF NOT EXISTS bookings (
    id          TEXT PRIMARY KEY,
    start_at    TIMESTAMPTZ NOT NULL,
    end_at      TIMESTAMPTZ NOT NULL,
    room        INTEGER NOT NULL REFERENCES rooms(number),
    title       TEXT NOT NULL,
    description TEXT,
    user_email  TEXT NOT NULL DEFAULT '',
    telegram_id TEXT NOT NULL DEFAULT '',
    is_private  BOOLEAN NOT NULL DEFAULT false
);

DO $$ BEGIN
    ALTER TABLE bookings ADD CONSTRAINT room_time_no_overlap
        EXCLUDE USING gist (
            room WITH =,
            tstzrange(start_at, end_at, '[)') WITH &&
        );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS bookings_start_idx ON bookings(start_at);
CREATE INDEX IF NOT EXISTS bookings_end_idx   ON bookings(end_at);
CREATE INDEX IF NOT EXISTS bookings_room_idx  ON bookings(room);
