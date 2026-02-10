# internal/storage

## Responsibility
Owns persistence artifacts and data-access boundaries: SQL migrations, schema bootstrap, and PostgreSQL repository implementations used by domain services.

## Why this module is separated
Data schema/versioning and SQL execution plans must evolve with strict backward-compatibility and operational controls. Isolating storage concerns allows schema changes, index tuning, and transactional guarantees to be managed independently from transport and business orchestration.

## Depends on
- PostgreSQL driver/runtime connectivity
- Domain model definitions for scan/bind contracts
- Migration execution tooling/scripts used during deploy/startup workflows

## Depended on by
- Auth, profiles, matches, and notifications services via repository interfaces
- Deployment scripts (migration execution)
- Integration tests that apply migrations and validate persisted behavior
