-- Enables helpful PostgreSQL extensions used by this service.
-- pgcrypto gives gen_random_uuid() for UUID primary keys.
-- citext gives case-insensitive text for email uniqueness.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
