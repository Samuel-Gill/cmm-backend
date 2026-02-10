# cmd/matchmaking-service

## Purpose
Contains the service entrypoint composition for bootstrapping modules and starting servers.

## Technical Reason
Keeping startup wiring in `cmd/` isolates runtime composition from reusable package logic.
