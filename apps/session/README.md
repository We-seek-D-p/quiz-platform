# Session Service

Микросервис runtime-части игровых сессий платформы Quiz. Он отвечает за проведение игры в реальном времени:
лобби, WebSocket-подключения хоста и игроков, переходы между фазами, таймеры вопросов, приём ответов, подсчёт очков,
leaderboard и reconnect игроков.

## Как сервис участвует в игре

Host создаёт игровую сессию через Management Service. После этого Management инициализирует runtime в Session Service.
Session получает bootstrap-данные по квизу, создаёт код комнаты, сохраняет runtime-состояние в Redis и начинает
принимать WebSocket-подключения.

Игроки подключаются к комнате по коду и nickname. После входа игрок получает token для reconnect, чтобы восстановить
соединение без повторного создания участника. Хост управляет стартом и завершением игры, а переходы между вопросом,
reveal-фазой и leaderboard-фазой выполняются сервером по таймерам.

## Основные части сервиса

- **Transport**. В [internal/transport/http](internal/transport/http) находится HTTP-контур: health checks, internal
  endpoints и общие middleware.
  В [internal/transport/ws](internal/transport/ws) находится WebSocket-контур для хоста и игроков, обработка входящих
  сообщений, рассылка событий и локальный timer loop
- **Service**. В [internal/service/session](internal/service/session) сосредоточена основная orchestration-логика
  runtime-сессии:
  инициализация комнаты, подключение участников, старт игры, отправка ответов, переходы state-machine, завершение игры и
  подготовка событий для транспорта
- **Domain**. В [internal/domain](internal/domain) находятся доменные типы и правила: runtime-статусы, snapshot сессии,
  участники, ответы, leaderboard, расчёт очков и доменные ошибки
- **Repositories**. В [internal/repository/redis](internal/repository/redis) находится работа с Redis: runtime-состояние
  активных сессий, коды комнат, участники, ответы и leaderboard
- **Clients**. В [internal/client/management](internal/client/management) находится клиент для внутреннего
  взаимодействия с Management Service:
  получение bootstrap-данных, обновление статуса и отправка финальных результатов

[cmd/session-service/main.go](cmd/session-service/main.go) является точкой входа приложения.  
Сборка зависимостей и запуск HTTP-сервера находятся в [internal/app](internal/app).
