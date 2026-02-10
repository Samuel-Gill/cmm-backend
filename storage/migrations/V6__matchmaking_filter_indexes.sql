-- Adds indexes to support efficient browse/filter and relationship queries with pagination.

CREATE INDEX IF NOT EXISTS idx_profiles_filter_age_gender ON matchmaking_profiles (age, gender);
CREATE INDEX IF NOT EXISTS idx_profiles_filter_income ON matchmaking_profiles (income);
CREATE INDEX IF NOT EXISTS idx_profiles_filter_location_residency ON matchmaking_profiles (location, residency_status);
CREATE INDEX IF NOT EXISTS idx_profiles_filter_discoverable_created ON matchmaking_profiles (is_discoverable, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_profile_likes_liked_created ON profile_likes (liked_user_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_profile_likes_liker_created ON profile_likes (liker_user_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_matches_active_pair_time ON matches (status, matched_at DESC, user_low_id, user_high_id);
