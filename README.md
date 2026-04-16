# Quiz Platform monorepo

## Документация для разработки

- [Правила разработки](docs/development-workflow.md)
- [Локальный запуск платформы](docs/local-stack.md)

## Justfile

Получить список глобальных команд (из корня репозитория):

```shell
just
```

Получить список локальных команд сервиса (из папки сервиса):

```shell
cd apps/auth && just
cd apps/management && just
cd apps/frontend && just
```

Команды сервисов можно запускать из корня следующим образом:

```shell
just <service> <command>
```
