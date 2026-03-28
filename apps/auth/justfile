# Setup environment and hooks
install:
    @echo "===> INSTALLING DEPENDENCIES (UV SYNC)..."
    uv sync
    @echo "===> SETTING UP PRE-COMMIT HOOKS..."
    uv run pre-commit install
    @echo "===> SUCCESS: ENVIRONMENT READY"

# Run Ruff: Fix errors and format code
ruff:
    @echo "===> RUNNING RUFF (LINT & FORMAT)..."
    uv run ruff check . --fix
    uv run ruff format .

# Run Mypy: Type checking only
types:
    @echo "===> RUNNING MYPY (TYPE CHECKING)..."
    uv run mypy .

# Run Pytest: Unit tests with coverage
test:
    @echo "===> RUNNING PYTEST (UNIT TESTS)..."
    -uv run pytest

# Static analysis and formatting
lint: ruff types

# Full verification: Lint + Types + Tests
check: lint test
    @echo "===> SUCCESS: ALL CHECKS PASSED"

# Start application
run:
    @echo "===> STARTING APPLICATION..."
    uv run python -m src.main

# Run pre-commit hooks manually
pc:
    @echo "===> RUNNING PRE-COMMIT HOOKS..."
    uv run pre-commit run --all-files

# Clean project caches and bytecode
clean:
    @echo "===> CLEANING CACHES AND PYCACHE..."
    rm -rf .pytest_cache .ruff_cache .mypy_cache .coverage htmlcov
    find . -type d -name "__pycache__" -exec rm -rf {} +
