## Local run

```shell
uv run uvicorn backend.main:app --reload
```

### `.env` example

```dotenv
# App
APP_NAME="Quiz Auth"
DEBUG=False

# Database
DATABASE_URL=postgresql+asyncpg://auth_user:password@db:5432/quiz_auth

# JWT
JWT_SECRET_KEY=change_me_in_prod
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=10
JWT_REFRESH_TOKEN_EXPIRE_DAYS=30
```

## Docker Compose

The repo ships with `docker-compose.yml` that brings up Postgres 18 and the auth service.

```shell
docker compose up --build
```

- Compose loads every variable from `.env`. Add `DATABASE_URL_DOCKER` (already present in the sample) if you need a different DSN for containers—the compose file passes it as `DATABASE_URL` while leaving your local `DATABASE_URL` untouched.
- The image entrypoint (`docker/app-entrypoint.sh`) runs `alembic upgrade head` inside the container before booting Uvicorn, so migrations apply automatically on the first run.
- To rerun migrations manually:

  ```shell
  docker compose run --rm app alembic upgrade head
  ```

Postgres data is stored in the `pg_data` Docker volume. Remove it (`docker volume rm quiz-auth_pg_data`) to reset the cluster.
