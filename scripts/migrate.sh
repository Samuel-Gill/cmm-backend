#!/usr/bin/env sh
set -eu

docker compose exec -T postgres psql \
  -U "${POSTGRES_USER:-postgres}" \
  -d "${POSTGRES_DB:-matchmaking}" \
  -f /storage/schema/schema.sql
