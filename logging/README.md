# logging module

## Purpose
Provides structured JSON request/error logging middleware for the service.

## Technical Reason
Central middleware-based logging ensures every request has consistent fields (request id, method/path/status, duration, authenticated user id when present) and panic/error traces for diagnostics.

## Middleware
- `RequestID()`: injects/propagates `X-Request-ID` and context request id.
- `RequestLogger()`: logs request start/finish and error events.
- `Recoverer()`: captures panics and logs stack traces.

## Router Integration
Attach middleware early in router setup:
- request id
- recoverer
- request logger
