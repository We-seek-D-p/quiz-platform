# Auth Service

Микросервис аутентификации и операций с пользователем платформы Quiz. Всё API микросервиса можно найти в документации
OpenAPI по пути [auth/openapi/openapi.json](openapi/openapi.json)
В [src/quiz_auth/models](src/quiz_auth/models) находятся модели для работы с пользователем и токенами.
В [auth/alembic](alembic) находится настройка миграций через Alembic.
В [src/quiz_auth/core](src/quiz_auth/core) находится общий функционал модуля.
В [src/quiz_auth/utils/security.py](src/quiz_auth/utils/security.py) содержится функционал с хешированием пароля и
созданием и декодированием токенов.
Основной код разбит на 3 слоя:

+ __API__. В [api/auth.py](src/quiz_auth/api/auth.py) содержится функционал регистрации, входа, обновление токена и куки
  и выход из аккаунта. Validate
  из [api/internal.py](src/quiz_auth/api/internal.py) нужен для установки заголовков `X-User-Id` и `X-User-Role`.
  В [api/users/py](src/quiz_auth/api/users.py) сосредоточена основная работа с пользователем, не касающаяся
  аутентификации, - смена пароля, обновление информации об аккаунте и удаление аккаунта. Через DI API использует
  `get_session` и `get_current_user` из [core/dependencies.py](src/quiz_auth/core/dependencies.py) для получения сессии
  БД и текущего пользователя с проверкой токена, соответственно.
+ __Services__. В [services/auth_service.py](src/quiz_auth/services/auth_service.py)
  и [services/user_service.py](src/quiz_auth/services/user_service.py) сосредоточена основная бизнес-логика, которую для
  чистоты кода вызывает API слой.
+ __Repositories__. [repositories/refresh_token_repository.py](src/quiz_auth/repositories/refresh_token_repository.py)
  отвечает за создание, получение, отзыв токена пользователя на уровне
  БД, [repositories/role_repository.py](src/quiz_auth/repositories/role_repository.py) отвечает за поиск роли в БД по её
  текстовому идентификатору (slug) и проверка наличия и автоматическое создание обязательной роли host при её
  отсутствии, [repositories/user_repository.py](src/quiz_auth/repositories/user_repository.py) содержит необходимый
  функционал на уровне БД для CRUD пользователя, инкремента счётчика токена и обновления последнего выхода из аккаунта.

[main.py](src/quiz_auth/main.py) подключает все роутеры в один общий `app`.

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
AUTH_REFRESH_COOKIE_SECURE=true
```

`AUTH_REFRESH_COOKIE_SECURE=false` можно использовать только для локальной разработки по HTTP.
