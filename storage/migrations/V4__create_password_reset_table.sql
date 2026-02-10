-- Stores password reset tokens as hashes for one-time reset flow.
-- token_hash is unique to prevent ambiguous token usage.
-- used_at allows single-use semantics without hard deletion.
CREATE TABLE IF NOT EXISTS auth_password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_password_resets_user_id ON auth_password_resets (user_id);
CREATE INDEX IF NOT EXISTS idx_auth_password_resets_expires_at ON auth_password_resets (expires_at);
