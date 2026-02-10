# internal/auth

## Responsibility
Owns authentication and credential lifecycle operations: signup, login, logout/revocation, password change, and password reset token flow.

## Why this module is separated
Authentication is a security-critical boundary with independent change velocity (token semantics, hashing policy, credential workflows). Isolating it reduces coupling with profile/matching rules and allows stricter review and testing around auth-specific logic.

## Depends on
- HTTP routing/handler layer for request binding and status mapping
- Token management utilities for JWT and opaque token generation
- Password hashing implementation (`bcrypt`/`argon2`)
- Storage/repository interfaces for users, refresh tokens, and reset tokens

## Depended on by
- Router composition that mounts `/auth/*` endpoints
- Auth middleware that validates bearer tokens and resolves user identity
- Other domain modules indirectly, through authenticated user context
