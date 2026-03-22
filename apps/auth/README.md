```shell
uv run uvicorn backend.main:app --reload
```

`.env` example
```dotenv
APP_NAME='Quiz App'
DEBUG=True

DATABASE_URL=postgresql+asyncpg://kostamak:seekdeep10@localhost:5432/postgres

JWT_SECRET_KEY=we_very_suck_deep_seek
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=30
JWT_REFRESH_TOKEN_EXPIRE_DAYS=7
```
