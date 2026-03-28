# Pinned uv image used only to copy uv/uvx binaries.
FROM ghcr.io/astral-sh/uv:0.11.2 AS uv_bin

# Build stage: install dependencies and the auth package.
FROM python:3.14-slim AS builder
COPY --from=uv_bin /uv /uvx /bin/

ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    UV_PYTHON_DOWNLOADS=never

WORKDIR /workspace

# System packages required to build Python dependencies.
RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy workspace metadata first to maximize Docker cache reuse.
COPY pyproject.toml uv.lock ./
COPY apps/auth/pyproject.toml apps/auth/pyproject.toml
COPY apps/management/pyproject.toml apps/management/pyproject.toml

# Install workspace dependencies without installing workspace packages themselves.
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-install-workspace --no-dev

# Copy only the auth service source code.
COPY apps/auth apps/auth

# Install the auth package into the virtual environment.
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --package quiz-auth --no-dev --no-editable

FROM python:3.14-slim AS runtime

ENV PATH="/workspace/.venv/bin:$PATH" \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

RUN groupadd -r svc && useradd -r -g svc svc

COPY --from=builder --chown=svc:svc /workspace/.venv /workspace/.venv
COPY --from=builder --chown=svc:svc /workspace/apps/auth /srv/auth

RUN chmod +x /srv/auth/docker/app-entrypoint.sh

USER svc
WORKDIR /srv/auth

EXPOSE 8000
ENTRYPOINT ["/srv/auth/docker/app-entrypoint.sh"]
