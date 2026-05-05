FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /bin/session-service \
    ./cmd/session-service

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

USER app

COPY --from=builder /bin/session-service /bin/session-service

EXPOSE 8000

ENTRYPOINT ["/bin/session-service"]
