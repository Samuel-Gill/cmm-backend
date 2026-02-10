-- Indexes focused on common read paths: auth lookup, discovery queries,
-- like->match transitions, and active match retrieval.

CREATE INDEX IF NOT EXISTS idx_auth_users_email ON auth_users (email);
CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_user_id ON auth_refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_expires_at ON auth_refresh_tokens (expires_at);

CREATE INDEX IF NOT EXISTS idx_profiles_discoverable ON matchmaking_profiles (is_discoverable);
CREATE INDEX IF NOT EXISTS idx_profiles_country_discoverable ON matchmaking_profiles (country_code, is_discoverable);
CREATE INDEX IF NOT EXISTS idx_profiles_birth_date ON matchmaking_profiles (birth_date);

CREATE INDEX IF NOT EXISTS idx_profile_interests_interest ON profile_interests (interest);

CREATE INDEX IF NOT EXISTS idx_profile_likes_liked_status ON profile_likes (liked_user_id, status);
CREATE INDEX IF NOT EXISTS idx_profile_likes_liker_status ON profile_likes (liker_user_id, status);

CREATE INDEX IF NOT EXISTS idx_matches_user_low_status ON matches (user_low_id, status);
CREATE INDEX IF NOT EXISTS idx_matches_user_high_status ON matches (user_high_id, status);
CREATE INDEX IF NOT EXISTS idx_matches_matched_at ON matches (matched_at DESC);
