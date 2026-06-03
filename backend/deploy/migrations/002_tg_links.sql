CREATE TABLE IF NOT EXISTS telegram_links (
    session_id   TEXT PRIMARY KEY,
    telegram_id  TEXT NOT NULL DEFAULT '',
    token        TEXT NOT NULL DEFAULT '',
    confirmed    BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS tg_links_token_idx ON telegram_links(token);
CREATE INDEX IF NOT EXISTS tg_links_tg_idx    ON telegram_links(telegram_id);
