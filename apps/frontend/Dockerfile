FROM oven/bun:1.2.5-alpine AS builder

WORKDIR /workspace

COPY package.json bun.lock ./

RUN --mount=type=cache,target=/root/.bun \
    bun install --frozen-lockfile

COPY . .

RUN bun run build

FROM nginxinc/nginx-unprivileged:1.29.5-alpine AS runtime

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder --chown=101:101 /workspace/dist /usr/share/nginx/html

USER 101
EXPOSE 8080
