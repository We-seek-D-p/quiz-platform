set dotenv-load := true

default:
    @just --list

# Install
install-py:
    uv sync --all-packages

install-front:
    cd apps/frontend && bun install

install-hooks:
    uv run pre-commit install

install:
    just install-py
    just install-front
    just install-hooks

# Python
fix-py service="":
    if [ -n "{{service}}" ]; then uv run ruff check apps/{{service}} --fix; else uv run ruff check apps/auth apps/management --fix; fi

fmt-py service="":
    if [ -n "{{service}}" ]; then uv run ruff format apps/{{service}}; else uv run ruff format apps/auth apps/management; fi

lint-py service="":
    if [ -n "{{service}}" ]; then uv run ruff check apps/{{service}}; else uv run ruff check apps/auth apps/management; fi

# test-py service="":
#     if [ -n "{{service}}" ]; then uv run pytest apps/{{service}}/tests; else uv run pytest apps/auth/tests apps/management/tests; fi

check-py service="":
    just fix-py {{service}}
    just fmt-py {{service}}
    just lint-py {{service}}

openapi-auth:
    uv run --package quiz-auth python -m quiz_auth.openapi.export

#openapi-management:
#    uv run --package quiz-management python -m quiz_management.openapi.export
#
#openapi:
#    just openapi-auth
#    just openapi-management

# Frontend
fmt-front:
    cd apps/frontend && bun run format

lint-front:
    cd apps/frontend && bun run lint

fix-front:
    cd apps/frontend && bun run lint:fix

test-front:
    cd apps/frontend && bun run test

check-front:
    just fmt-front
    just lint-front
    just test-front

# Go
fmt-go:
    @echo "Go service is not added yet"

lint-go:
    @echo "Go service is not added yet"

test-go:
    @echo "Go service is not added yet"

# Combined
fmt:
    just fmt-py
    just fmt-front

lint:
    just lint-py
    just lint-front

check:
    just check-py
    just check-front

pc:
    uv run pre-commit run --all-files
