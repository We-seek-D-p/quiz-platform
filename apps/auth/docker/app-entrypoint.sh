#!/usr/bin/env bash
set -euo pipefail

cd /app

export PYTHONPATH="/app/src:${PYTHONPATH:-}"

echo "===> Applying database migrations..."
alembic upgrade head

echo "===> Starting API server..."
exec python -m uvicorn quiz_auth.main:app --host 0.0.0.0 --port 8000
