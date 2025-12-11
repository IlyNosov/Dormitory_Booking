CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS bookings (
                                        id           TEXT PRIMARY KEY,
                                        start_at     TIMESTAMPTZ NOT NULL,
                                        end_at       TIMESTAMPTZ NOT NULL,
                                        room         INTEGER NOT NULL CHECK (room IN (21,132,256)),
    title        TEXT NOT NULL,
    description  TEXT,
    telegram_id  TEXT NOT NULL,
    is_private   BOOLEAN NOT NULL DEFAULT false
    );

ALTER TABLE bookings
    ADD CONSTRAINT room_time_no_overlap
    EXCLUDE USING gist (
        room WITH =,
        tstzrange(start_at, end_at, '[)') WITH &&
    );

CREATE INDEX IF NOT EXISTS bookings_start_idx ON bookings(start_at);
CREATE INDEX IF NOT EXISTS bookings_end_idx   ON bookings(end_at);
