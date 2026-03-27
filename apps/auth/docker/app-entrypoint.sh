#!/usr/bin/env bash
set -euo pipefail

cd /app

echo "===> Applying database migrations..."
cd backend
alembic upgrade head
cd ..

echo "===> Starting API server..."
exec python -m uvicorn backend.main:app --host 0.0.0.0 --port 8000
