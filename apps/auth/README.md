```shell
uv run uvicorn backend.main:app --reload
```

`.env` example
```dotenv
# App
APP_NAME="Quiz Auth"
DEBUG=False

# Database
DATABASE_URL=postgresql+asyncpg://auth_user:password@localhost:5432/quiz_auth

# JWT
JWT_SECRET_KEY=change_me_in_prod
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=1440 # 1 day (temporarily)
JWT_REFRESH_TOKEN_EXPIRE_DAYS=30
```
