# Локальный запуск платформы

`docker-compose.yaml` в корне репозитория используется для локального запуска и проверки платформы перед деплоем.

## Первый запуск

```bash
cp .env.example .env
docker compose up --build
```

## Пересоздание с очисткой volume

```bash
docker compose down -v
docker compose up --build
```
