# Команды монорепозитория

В репозитории используется root `justfile` как единая точка входа для локальных команд.

## Установка

### Python

```bash
just install-py
````

Устанавливает Python-зависимости для workspace:

* `apps/auth`
* `apps/management`

### Frontend

```bash
just install-front
```

Устанавливает frontend-зависимости в `apps/frontend`.

### Git hooks

```bash
just install-hooks
```

Устанавливает `pre-commit` hooks.

### Полная установка

```bash
just install
```

Выполняет:

* `just install-py`
* `just install-front`
* `just install-hooks`

---

## Python

### Автоисправление Ruff

```bash
just fix-py
```

Запуск для одного сервиса:

```bash
just fix-py auth
just fix-py management
```

### Форматирование Ruff

```bash
just fmt-py
```

Запуск для одного сервиса:

```bash
just fmt-py auth
just fmt-py management
```

### Линтинг Ruff

```bash
just lint-py
```

Запуск для одного сервиса:

```bash
just lint-py auth
just lint-py management
```

---

## Frontend

### Форматирование

```bash
just fmt-front
```

### Линтинг

```bash
just lint-front
```

### Тесты

```bash
just test-front
```

---

## Go

На текущем этапе Go-сервис ещё не добавлен, но интерфейс под него зарезервирован.

Команды:

```bash
just fmt-go
just lint-go
just test-go
```

После добавления Go-сервиса эти команды будут привязаны к соответствующему приложению.

---

## Комбинированные команды

### Форматирование всего

```bash
just fmt
```

Сейчас выполняет:

* `just fmt-py`
* `just fmt-front`

### Линтинг всего

```bash
just lint
```

Сейчас выполняет:

* `just lint-py`
* `just lint-front`

---

## Pre-commit

### Прогон всех hook'ов вручную

```bash
just pc
```

---

### Примечания

* Python-команды запускаются из root workspace через `uv`
* frontend-команды запускаются внутри `apps/frontend`
