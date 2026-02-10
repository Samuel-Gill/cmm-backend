# internal/profiles

## Responsibility
Owns user profile lifecycle and profile read models: create, update, get-by-id, and paginated listing, including privacy-flag persistence.

## Why this module is separated
Profile data has its own schema evolution path and validation rules that are distinct from auth/session concerns and matching orchestration. Separation keeps domain invariants focused and prevents transport/persistence leakage across unrelated modules.

## Depends on
- Auth context (actor user id) for write authorization decisions
- Profile repository abstraction backed by PostgreSQL
- HTTP handlers for payload decoding and response serialization

## Depended on by
- Router for `/profiles/*` endpoint registration
- Matchmaking module for profile attributes used in browse/filter queries
- API documentation and integration tests targeting profile contracts
