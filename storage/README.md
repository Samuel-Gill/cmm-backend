# storage module

## Purpose
Owns persistence concerns for profiles snapshots, active queues, matches, and operational metadata.

## Technical Reason
Encapsulating data access behind storage abstractions keeps business modules datastore-agnostic and enables safer schema evolution.

## Scope
- repository interfaces and implementations
- transaction boundaries and consistency guarantees
- migration strategy and schema ownership
- caching and read/write performance patterns

## Interfaces
Define repository contracts consumed by `matches/` and `profiles/` modules.
