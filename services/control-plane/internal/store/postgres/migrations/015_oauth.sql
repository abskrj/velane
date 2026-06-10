-- Social login (Google / GitHub). Users may now exist without a password
-- (OAuth-only accounts), so password_hash becomes nullable.
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- External identities linked to a Velane user. A single user can link multiple
-- providers; (provider, subject) is globally unique.
CREATE TABLE IF NOT EXISTS oauth_identities (
    id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider   TEXT NOT NULL,              -- 'google' | 'github'
    subject    TEXT NOT NULL,              -- stable provider-issued user id
    email      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, subject)
);

CREATE INDEX IF NOT EXISTS idx_oauth_identities_user_id ON oauth_identities(user_id);
