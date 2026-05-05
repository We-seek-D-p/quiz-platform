FROM oven/bun:1.2.5-alpine AS builder

WORKDIR /workspace

COPY package.json bun.lock ./

RUN --mount=type=cache,target=/root/.bun \
    bun install --frozen-lockfile

COPY . .

RUN bun run build

FROM oven/bun:1.2.5-alpine AS runtime

WORKDIR /app

COPY --from=builder /workspace/.output ./.output

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV NITRO_HOST=0.0.0.0
ENV PORT=8080
ENV NITRO_PORT=8080

EXPOSE 8080

CMD ["bun", ".output/server/index.mjs"]
