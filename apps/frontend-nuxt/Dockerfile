FROM oven/bun:1.2.13-alpine AS builder

WORKDIR /workspace

COPY package.json bun.lock ./

RUN --mount=type=cache,target=/root/.bun \
    bun install --frozen-lockfile

COPY . .

RUN bun run generate

FROM nginxinc/nginx-unprivileged:1.29.8-alpine AS runtime

WORKDIR /app

COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder --chown=101:101 /workspace/.output/public /usr/share/nginx/html

USER 101

EXPOSE 8080
