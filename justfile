set dotenv-load := true
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

# List recipes
[private]
[default]
default:
    @just --list --unsorted

# Run command in auth service
auth command='default':
    just --justfile apps/auth/justfile --working-directory apps/auth {{ command }}

# Run command in management service
management command='default':
    just --justfile apps/management/justfile --working-directory apps/management {{ command }}

# Run command in frontend service
frontend command='default':
    just --justfile apps/frontend/justfile --working-directory apps/frontend {{ command }}

# Install dependencies in all services
install: (auth "install") (management "install") (frontend "install") hooks-install

# Fix issues in all services
fix: (auth "fix") (management "fix") (frontend "fix")

# Format code in all services
fmt: (auth "fmt") (management "fmt") (frontend "fmt")

# Check formatting in all services
fmt-check: (auth "fmt-check") (management "fmt-check") (frontend "fmt-check")

# Lint code in all services
lint: (auth "lint") (management "lint") (frontend "lint")

# Run tests in all services
test: (auth "test") (management "test") (frontend "test")

# Check code in all services
check: (auth "check") (management "check") (frontend "check")

# Run all tasks in all services
all: (auth "all") (management "all") (frontend "all")

# Export OpenAPI specs for Python services
openapi: (auth "openapi") (management "openapi")

# Install git hooks
hooks-install:
    uvx pre-commit install

# Run pre-commit
precommit:
    uvx pre-commit run --all-files
