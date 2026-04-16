FROM ghcr.io/astral-sh/uv:0.11.2 AS uv

FROM python:3.14-slim AS build
COPY --from=uv /uv /uvx /bin/

ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    UV_PYTHON_DOWNLOADS=never

WORKDIR /workspace

COPY pyproject.toml uv.lock ./
COPY src ./src
COPY docker ./docker
COPY alembic ./alembic
COPY alembic.ini ./
COPY README.md ./

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev --no-editable

FROM python:3.14-slim AS runtime

ENV PATH="/workspace/.venv/bin:$PATH" \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

RUN groupadd -g 10001 -r svc \
    && useradd -u 10001 -r -g svc -s /usr/sbin/nologin svc

COPY --from=build --chown=svc:svc /workspace/.venv /workspace/.venv
COPY --from=build --chown=svc:svc /workspace /srv/management

WORKDIR /srv/management

RUN chmod +x docker/app-entrypoint.sh

USER svc

EXPOSE 8000
ENTRYPOINT ["./docker/app-entrypoint.sh"]
