# Quiz Management Service

Микросервис управления квизами, вопросами и игровыми сессиями платформы Quiz. Всё API микросервиса можно найти в документации OpenAPI по пути [management/openapi/openapi.json](openapi/openapi.json)

В [src/quiz_management/models](src/quiz_management/models) находятся модели для работы с квизами, вопросами, ответами, сессиями и ошибками.
В [management/alembic](alembic) находится настройка миграций через Alembic.
В [src/quiz_management/core](src/quiz_management/core) находится общий функционал модуля.

Основной код разбит на 3 слоя:

+ __API__. В [api/quiz.py](src/quiz_management/api/quiz.py), [api/question.py](src/quiz_management/api/question.py) и [api/session.py](src/quiz_management/api/session.py) содержится функционал эндпоинтов для базового CRUD квизов, вопросов (совместно с вариантами ответов) и игровых сессий. [api/internal.py](src/quiz_management/api/internal.py) отвечает за внутренние общение с Session микросервисом.
+ __Services__. В [services/quiz.py](src/quiz_management/services/quiz.py), [services/question.py](src/quiz_management/services/question.py) и [services/session.py](src/quiz_management/services/session.py) сосредоточена основная бизнес-логика управления сущностями, которую для чистоты кода вызывает API слой. Класс `SessionServiceClient` из [services/session_client.py](src/quiz_management/services/session_client.py) отвечает за внутреннее HTTP-взаимодействие (инициализация, получение и удаление игровых сессий) с микросервисом Session Service.
+ __Repositories__. Слой взаимодействия с базой данных. [repositories/quiz_repository.py](src/quiz_management/repositories/quiz_repository.py) отвечает за CRUD операции квизов, [repositories/question_repository.py](src/quiz_management/repositories/question_repository.py) осуществляет работу с вопросами и их вариантами ответов в БД, а [repositories/session_repositories.py](src/quiz_management/repositories/session_repositories.py) управляет хранением и обновлением игровых сессий.

[main.py](src/quiz_management/main.py) подключает все роутеры в один общий `app`.
