# Auth Service

Микросервис аутентификации платформы Quiz.

Функционал:

- Регистрация и вход пользователей
- Выпуск access/refresh токенов
- Валидация access-токена через GET `/validate` (возвращает `X-User-Id` и `X-User-Role` в заголовках)

## Локальный запуск

Команду нужно выполнять из корня репозитория:

```shell
uv run --package quiz-auth uvicorn quiz_auth.main:app --reload
```

## Пример `.env`

```dotenv
# Приложение
AUTH_APP_NAME="Quiz Auth"
AUTH_DEBUG=false

# База данных
AUTH_DATABASE_URL=postgresql+asyncpg://quiz_auth:auth_password@db:5432/quiz

# JWT
AUTH_JWT_SECRET_KEY=replace-with-at-least-32-characters
AUTH_JWT_ALGORITHM=HS256
AUTH_ACCESS_TOKEN_TTL_MINUTES=10
AUTH_REFRESH_TOKEN_TTL_DAYS=30
```
