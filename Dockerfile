FROM golang:1.20-alpine AS builder

ARG GOOSE_VERSION=3.10.0

RUN OS=$(uname -s | tr '[:upper:]' '[:lower:]') && \
    ARCH=$(uname -m) && \
    case $ARCH in \
    arm64) GOOSE_ARCH=arm64 ;; \
    x86_64) GOOSE_ARCH=x86_64 ;; \
    *) echo "unsupported architecture $ARCH. exiting."; exit 1 ;; \
    esac && \
    wget -O /bin/goose "https://github.com/pressly/goose/releases/download/v${GOOSE_VERSION}/goose_${OS}_${GOOSE_ARCH}"

WORKDIR /app

COPY go.mod go.sum ./
COPY ./cmd ./cmd
COPY ./internal ./internal

RUN go mod download

RUN go build -o hive-backend ./cmd 

FROM alpine:latest as runner

RUN apk update && \
    apk upgrade && \
    rm -rf /var/cache/apk/*

COPY ./migrations ./migrations
COPY --from=builder /bin/goose /bin/goose
COPY --from=builder /app/hive-backend .

ENV HIVE_BACKEND_LOG_LEVEL=INFO \
    HIVE_BACKEND_SERVER_PORT=8080 \
    HIVE_BACKEND_REQUEST_TIMEOUT=30s \
    HIVE_BACKEND_JWT_SECRET_KEY=lock-code-ends-with-42 \
    HIVE_BACKEND_FAKE_USER_PASSWORD=fixture-person \
    HIVE_BACKEND_DB_HOST=hive-backend-db \
    HIVE_BACKEND_DB_PORT=5432 \
    HIVE_BACKEND_DB_NAME=hive \
    HIVE_BACKEND_DB_USER=admin \
    HIVE_BACKEND_DB_PASSWORD=hard-password \
    HIVE_BACKEND_DB_MAX_CONNECTIONS=100 \
    HIVE_BACKEND_DB_CONNECTION_LIFETIME=1m

EXPOSE $HIVE_BACKEND_SERVER_PORT

CMD chmod +x /bin/goose && \
    chmod +x ./hive-backend && \
    goose -dir=./migrations postgres "postgres://${HIVE_BACKEND_DB_USER}:${HIVE_BACKEND_DB_PASSWORD}@${HIVE_BACKEND_DB_HOST}:${HIVE_BACKEND_DB_PORT}/${HIVE_BACKEND_DB_NAME}?sslmode=disable" up && ./hive-backend
