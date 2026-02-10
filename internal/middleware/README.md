# internal/middleware

## Responsibility
Owns cross-cutting HTTP middleware concerns: request identity propagation, auth protection, structured request/error logging, and panic recovery.

## Why this module is separated
Middleware behavior is orthogonal to domain logic and is shared across all routes. Centralization guarantees consistent request context, error handling, and observability semantics across modules.

## Depends on
- Auth service for bearer token validation in protected routes
- Logging utilities for structured output and stack trace capture
- Router integration points for ordered middleware chaining

## Depended on by
- All HTTP route groups that require consistent context/logging/security behavior
- Domain handlers that consume middleware-injected context values (user id, request id)
