#!/bin/sh
set -eu

cd /srv/auth
export PYTHONPATH="/srv/auth/src:${PYTHONPATH:-}"

echo "===> Applying database migrations..."
alembic upgrade head

echo "===> Starting API server..."
exec uvicorn quiz_auth.main:app --host 0.0.0.0 --port 8000
