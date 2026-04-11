FROM oven/bun:1.2.5-alpine AS builder

WORKDIR /workspace/apps/frontend

COPY apps/frontend/package.json apps/frontend/bun.lock ./

RUN --mount=type=cache,target=/root/.bun \
    bun install --frozen-lockfile

COPY apps/frontend/ ./

RUN bun run build

FROM nginxinc/nginx-unprivileged:1.29.5-alpine AS runtime

COPY apps/frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder --chown=101:101 /workspace/apps/frontend/dist /usr/share/nginx/html

USER 101
EXPOSE 8080
