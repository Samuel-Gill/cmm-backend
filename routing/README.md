# routing module

## Purpose
Defines HTTP/gRPC route registration, middleware chaining, and transport-level request handling.

## Technical Reason
Transport logic should be decoupled from domain logic to keep handlers thin and maintain replaceable interfaces/protocols.

## Scope
- endpoint registration and version grouping
- middleware composition (auth, logging, recovery, rate limiting)
- request validation and response shaping
- error-to-status mapping strategy

## Interfaces
Describe handler contracts between transport layer and domain services.
