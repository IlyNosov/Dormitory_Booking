CREATE TABLE IF NOT EXISTS bot_users (
    telegram_chat_id BIGINT PRIMARY KEY,
    email            TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
