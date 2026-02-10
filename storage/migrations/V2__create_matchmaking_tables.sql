-- ============================================================================
-- V2: Core schema for auth, profile, search filters, likes, and matches.
-- This migration intentionally keeps all constraints in SQL so correctness does
-- not depend on application-layer checks.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- auth_users
--
-- Purpose:
--   Canonical account identity table used by matchmaking.
-- Why these fields/constraints:
--   - UUID PK for distributed-safe ID creation.
--   - CITEXT email with unique constraint for case-insensitive login identity.
--   - account_state constrained enum-style text to keep state transitions clear.
--   - password_hash required to store only hashed credentials, never plaintext.
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS auth_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    account_state TEXT NOT NULL DEFAULT 'active'
        CHECK (account_state IN ('active', 'suspended', 'deleted')),
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);

-- ----------------------------------------------------------------------------
-- auth_refresh_tokens
--
-- Purpose:
--   Stores refresh-token metadata for session rotation and revocation.
-- Why these fields/constraints:
--   - token_hash is unique; only hashed token material is persisted.
--   - revoked_at supports explicit logout/revocation workflows.
--   - expires_at allows database-level expiry filtering and cleanup jobs.
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS auth_refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    user_agent TEXT,
    ip_address INET
);

-- ----------------------------------------------------------------------------
-- matchmaking_profiles
--
-- Purpose:
--   Stores discoverable profile attributes used by match ranking/filtering.
-- Why these fields/constraints:
--   - One-to-one with auth_users via user_id PK.
--   - birth_date allows age computation without persisting redundant age value.
--   - is_discoverable supports opt-out from search without deleting account.
--   - latitude/longitude checks keep geo data valid at write-time.
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS matchmaking_profiles (
    user_id UUID PRIMARY KEY REFERENCES auth_users(id) ON DELETE CASCADE,
    display_name VARCHAR(80) NOT NULL,
    bio TEXT,
    gender TEXT NOT NULL,
    birth_date DATE NOT NULL,
    locale VARCHAR(16),
    country_code CHAR(2),
    latitude NUMERIC(8,5),
    longitude NUMERIC(8,5),
    is_discoverable BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (latitude IS NULL OR (latitude BETWEEN -90 AND 90)),
    CHECK (longitude IS NULL OR (longitude BETWEEN -180 AND 180))
);

-- ----------------------------------------------------------------------------
-- profile_interests
--
-- Purpose:
--   Normalized many-to-many style tag list for profile interests.
-- Why these fields/constraints:
--   - Composite PK prevents duplicate interests per user.
--   - Lowercased/trimmed interest labels are enforced by a check to improve
--     filter quality and avoid near-duplicate tags.
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS profile_interests (
    user_id UUID NOT NULL REFERENCES matchmaking_profiles(user_id) ON DELETE CASCADE,
    interest VARCHAR(40) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, interest),
    CHECK (interest = LOWER(BTRIM(interest)))
);

-- ----------------------------------------------------------------------------
-- search_filters
--
-- Purpose:
--   Stores per-user matchmaking preference filters.
-- Why these fields/constraints:
--   - One row per user simplifies reads on search path.
--   - min/max age and distance constraints validated in DB to prevent bad state.
--   - preferred_genders as TEXT[] supports multi-select preference in one row.
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS search_filters (
    user_id UUID PRIMARY KEY REFERENCES matchmaking_profiles(user_id) ON DELETE CASCADE,
    min_age SMALLINT NOT NULL DEFAULT 18,
    max_age SMALLINT NOT NULL DEFAULT 99,
    max_distance_km INTEGER NOT NULL DEFAULT 50,
    preferred_genders TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    require_verified_profiles BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (min_age BETWEEN 18 AND 99),
    CHECK (max_age BETWEEN 18 AND 99),
    CHECK (min_age <= max_age),
    CHECK (max_distance_km BETWEEN 1 AND 10000)
);

-- ----------------------------------------------------------------------------
-- profile_filter_exclusions
--
-- Purpose:
--   Fine-grained blocklist of profile IDs the user never wants returned.
-- Why these fields/constraints:
--   - Composite PK enforces one exclusion per pair.
--   - CHECK blocks self-exclusion noise.
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS profile_filter_exclusions (
    user_id UUID NOT NULL REFERENCES matchmaking_profiles(user_id) ON DELETE CASCADE,
    excluded_user_id UUID NOT NULL REFERENCES matchmaking_profiles(user_id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, excluded_user_id),
    CHECK (user_id <> excluded_user_id)
);

-- ----------------------------------------------------------------------------
-- profile_likes
--
-- Purpose:
--   Directed like action from one user to another.
-- Why these fields/constraints:
--   - Composite PK guarantees one active like record per directed pair.
--   - status supports state transitions (liked/withdrawn/passed) without deletes.
--   - CHECK prevents self-like.
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS profile_likes (
    liker_user_id UUID NOT NULL REFERENCES matchmaking_profiles(user_id) ON DELETE CASCADE,
    liked_user_id UUID NOT NULL REFERENCES matchmaking_profiles(user_id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'liked'
        CHECK (status IN ('liked', 'withdrawn', 'passed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (liker_user_id, liked_user_id),
    CHECK (liker_user_id <> liked_user_id)
);

-- ----------------------------------------------------------------------------
-- matches
--
-- Purpose:
--   Represents a mutual pairing between two users.
-- Why these fields/constraints:
--   - user_low_id/user_high_id stores canonical sorted pair to guarantee
--     uniqueness regardless of request order.
--   - pair uniqueness prevents duplicate match rows.
--   - status tracks lifecycle without dropping historical context.
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_low_id UUID NOT NULL REFERENCES matchmaking_profiles(user_id) ON DELETE CASCADE,
    user_high_id UUID NOT NULL REFERENCES matchmaking_profiles(user_id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'unmatched', 'blocked')),
    matched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (user_low_id <> user_high_id),
    CHECK (user_low_id::TEXT < user_high_id::TEXT),
    UNIQUE (user_low_id, user_high_id)
);
