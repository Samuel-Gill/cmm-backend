# internal/notifications

## Responsibility
Owns notification retrieval and acknowledgment semantics: unread listing and mark-as-read operations for authenticated users.

## Why this module is separated
Notification state has independent retention/read policies and query patterns. Isolating it keeps event consumption/read APIs decoupled from the domains that emit events (matches/auth), simplifying future channel expansion.

## Depends on
- Auth context for user-scoped access control
- Notification repository backed by PostgreSQL
- HTTP handler layer for endpoint input/output contracts

## Depended on by
- Router for `/notifications/*` endpoints
- Matchmaking flows that create notification records on like/mutual match events
- Clients polling unread notifications and updating read state
