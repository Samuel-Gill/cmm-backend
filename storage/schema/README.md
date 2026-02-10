# storage/schema

## Purpose
Provides consolidated SQL schema entrypoints for PostgreSQL-based environments.

## Technical Reason
Some environments bootstrap from a single SQL file while others use migration tooling. This folder keeps a canonical schema view that references migration SQL so both approaches stay consistent.
