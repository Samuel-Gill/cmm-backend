-- Canonical schema bootstrap for environments that apply schema as one file.
-- Migration-first deployments should continue using files in storage/migrations.

\i ../migrations/V1__create_extensions.sql
\i ../migrations/V2__create_matchmaking_tables.sql
\i ../migrations/V3__create_indexes.sql
\i ../migrations/V4__create_password_reset_table.sql
\i ../migrations/V5__extend_profiles_for_management.sql
\i ../migrations/V6__matchmaking_filter_indexes.sql
\i ../migrations/V7__create_notifications.sql
