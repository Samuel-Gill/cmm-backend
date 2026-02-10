-- Extends matchmaking_profiles to support detailed profile management fields.
-- These columns are added to back profile CRUD APIs and privacy controls.

ALTER TABLE matchmaking_profiles
    ADD COLUMN IF NOT EXISTS age SMALLINT,
    ADD COLUMN IF NOT EXISTS profession TEXT,
    ADD COLUMN IF NOT EXISTS education TEXT,
    ADD COLUMN IF NOT EXISTS income INTEGER,
    ADD COLUMN IF NOT EXISTS residency_status TEXT,
    ADD COLUMN IF NOT EXISTS location TEXT,
    ADD COLUMN IF NOT EXISTS marital_status TEXT,
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS hide_contact_info BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS hide_address BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS hide_income BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS hide_visa_status BOOLEAN NOT NULL DEFAULT FALSE;

-- Age and income should remain realistic and non-negative.
ALTER TABLE matchmaking_profiles
    DROP CONSTRAINT IF EXISTS chk_matchmaking_profiles_age,
    ADD CONSTRAINT chk_matchmaking_profiles_age CHECK (age IS NULL OR age BETWEEN 18 AND 100),
    DROP CONSTRAINT IF EXISTS chk_matchmaking_profiles_income,
    ADD CONSTRAINT chk_matchmaking_profiles_income CHECK (income IS NULL OR income >= 0);

CREATE INDEX IF NOT EXISTS idx_matchmaking_profiles_location ON matchmaking_profiles (location);
CREATE INDEX IF NOT EXISTS idx_matchmaking_profiles_marital_status ON matchmaking_profiles (marital_status);
