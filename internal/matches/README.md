# internal/matches

## Responsibility
Owns discovery and relationship workflows: filtered profile browse, like actions, reciprocal-like detection, and mutual match state materialization.

## Why this module is separated
Matching behavior combines query-heavy selection logic and transactional relationship updates. Keeping it isolated enables targeted performance tuning (indexes/SQL paths) without introducing side effects into profile/auth modules.

## Depends on
- Authenticated user context from middleware
- Matchmaking repository implementation using PostgreSQL
- Notification persistence hooks triggered during like/match transitions

## Depended on by
- Router for `/matches/*` endpoint surface
- Notifications module (indirectly via repository-side notification writes)
- Client-facing relationship APIs consumed by frontend/mobile clients
