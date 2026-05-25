FROM denoland/deno:2.8.0 AS builder

WORKDIR /workspace

COPY package.json deno.json deno.lock ./
RUN deno install --frozen

COPY . .

RUN deno task postinstall && deno task build

FROM denoland/deno:alpine-2.8.0 AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=8080

COPY --from=builder /workspace/.output/ ./

EXPOSE 8080

CMD ["task", "start"]
