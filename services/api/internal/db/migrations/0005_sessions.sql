-- Session lifecycle (T-112): which client opened it, when it was last used (sliding expiry).
ALTER TABLE sessions
    ADD COLUMN client       text NOT NULL DEFAULT 'web' CHECK (client IN ('web', 'mobile')),
    ADD COLUMN last_seen_at timestamptz;
