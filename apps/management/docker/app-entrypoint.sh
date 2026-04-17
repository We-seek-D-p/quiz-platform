#!/bin/sh
set -eu

cd /srv/management
export PYTHONPATH="/srv/management/src:${PYTHONPATH:-}"

echo "===> Applying database migrations..."
alembic upgrade head

echo "===> Starting API server..."
exec uvicorn quiz_management.main:app --host 0.0.0.0 --port 8000
