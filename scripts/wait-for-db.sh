#!/usr/bin/env bash
# wait-for-db.sh — blocks until PostgreSQL is accepting connections.
# Usage: ./scripts/wait-for-db.sh [max_attempts]

set -euo pipefail

MAX_ATTEMPTS="${1:-30}"
ATTEMPT=0

echo "Waiting for PostgreSQL at ${DB_HOST:-localhost}:${DB_PORT:-5432}..."

until pg_isready -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-nodus_protocol}" > /dev/null 2>&1; do
  ATTEMPT=$(( ATTEMPT + 1 ))
  if [ "$ATTEMPT" -ge "$MAX_ATTEMPTS" ]; then
    echo "ERROR: PostgreSQL not ready after ${MAX_ATTEMPTS} attempts. Aborting."
    exit 1
  fi
  echo "  attempt ${ATTEMPT}/${MAX_ATTEMPTS} — retrying in 2s..."
  sleep 2
done

echo "PostgreSQL is ready."
