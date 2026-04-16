set dotenv-load := true
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

# Maybe be useful in the future
# set dotenv-required := true

py_services := "auth management"

# List recipes
[default]
default:
    @just --list --unsorted

# Install Python deps
py-install:
    for service in {{ py_services }}; do \
        if [ -f "apps/$service/pyproject.toml" ]; then \
            uv sync --project apps/$service --all-groups; \
        fi; \
    done

# Fix Python issues
py-fix:
    for service in {{ py_services }}; do \
        if [ -d "apps/$service" ]; then \
            uv run --project apps/$service \
                ruff check apps/$service --config ruff.toml --fix; \
        fi; \
    done

# Format Python code
py-fmt:
    for service in {{ py_services }}; do \
        if [ -d "apps/$service" ]; then \
            uv run --project apps/$service \
                ruff format apps/$service --config ruff.toml; \
        fi; \
    done

# Check Python format
py-fmt-check:
    for service in {{ py_services }}; do \
        if [ -d "apps/$service" ]; then \
            uv run --project apps/$service \
                ruff format apps/$service --config ruff.toml --check; \
        fi; \
    done

# Lint Python code
py-lint:
    for service in {{ py_services }}; do \
        if [ -d "apps/$service" ]; then \
            uv run --project apps/$service \
                ruff check apps/$service --config ruff.toml; \
        fi; \
    done

# Run Python tests
py-test:
    for service in {{ py_services }}; do \
        if [ -f "apps/$service/pyproject.toml" ] && [ -d "apps/$service/tests" ]; then \
            uv run --project apps/$service pytest apps/$service/tests; \
        fi; \
    done

# Check Python code
py-check: py-fmt-check py-lint py-test

# Run all Python tasks
py-all: py-fix py-fmt py-lint py-test

# Fix one Python service
py-fix-one service:
    uv run --project apps/{{ service }} \
        ruff check apps/{{ service }} --config ruff.toml --fix

# Format one Python service
py-fmt-one service:
    uv run --project apps/{{ service }} \
        ruff format apps/{{ service }} --config ruff.toml

# Check format for one Python service
py-fmt-check-one service:
    uv run --project apps/{{ service }} \
        ruff format apps/{{ service }} --config ruff.toml --check

# Lint one Python service
py-lint-one service:
    uv run --project apps/{{ service }} \
        ruff check apps/{{ service }} --config ruff.toml

# Test one Python service
py-test-one service:
    if [ -d "apps/{{ service }}/tests" ]; then \
        uv run --project apps/{{ service }} pytest apps/{{ service }}/tests; \
    fi

# Check one Python service
py-check-one service: (py-fmt-check-one service) (py-lint-one service) (py-test-one service)

# Run all tasks for one Python service
py-all-one service: (py-fix-one service) (py-fmt-one service) (py-lint-one service) (py-test-one service)

# Export auth OpenAPI
auth-openapi:
    uv run --project apps/auth python -m quiz_auth.openapi.export

# Export management OpenAPI
management-openapi:
    uv run --project apps/management python -m quiz_management.openapi.export

# Install frontend deps
[working-directory("apps/frontend")]
front-install:
    bun install

# Fix frontend issues
[working-directory("apps/frontend")]
front-fix:
    bun run lint:fix

# Format frontend code
[working-directory("apps/frontend")]
front-fmt:
    bun run format

# Lint frontend code
[working-directory("apps/frontend")]
front-lint:
    bun run lint

# Run frontend tests
front-test:
    @echo "Frontend tests are not added yet"

# Check frontend code
front-check: front-fmt front-lint front-test

# Run all frontend tasks
front-all: front-fix front-fmt front-lint front-test

# Install Go deps
go-install:
    @echo "Go service is not added yet"

# Fix Go issues
go-fix:
    @echo "Go service is not added yet"

# Format Go code
go-fmt:
    @echo "Go service is not added yet"

# Lint Go code
go-lint:
    @echo "Go service is not added yet"

# Run Go tests
go-test:
    @echo "Go service is not added yet"

# Check Go code
go-check: go-fmt go-lint go-test

# Run all Go tasks
go-all: go-fix go-fmt go-lint go-test

# Install git hooks
hooks-install:
    uvx pre-commit install

# Install all deps
install: py-install front-install go-install hooks-install

# Run all tests
test: py-test front-test go-test

# Run all checks
check: py-check front-check go-check

# Run pre-commit
precommit:
    uvx pre-commit run --all-files
