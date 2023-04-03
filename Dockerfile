FROM golang:1.20-alpine

WORKDIR /hive-backend

COPY go.mod go.sum ./
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./migrations ./migrations

ENV HIVE_BACKEND_LOG_LEVEL=INFO \
    HIVE_BACKEND_SERVER_PORT=8080 \
    HIVE_BACKEND_JWT_SECRET_KEY=lock-code-ends-with-42 \
    HIVE_BACKEND_FAKE_USER_PASSWORD=fixture-person \
    HIVE_BACKEND_DB_HOST=hive-backend-db \
    HIVE_BACKEND_DB_PORT=5432 \
    HIVE_BACKEND_DB_NAME=hive \
    HIVE_BACKEND_DB_USER=admin \
    HIVE_BACKEND_DB_PASSWORD=hard-password \
    HIVE_BACKEND_DB_MAX_CONNECTIONS=10 \
    HIVE_BACKEND_DB_CONNECTION_LIFETIME=1m

EXPOSE $HIVE_BACKEND_SERVER_PORT

RUN go build -o hive-backend ./cmd && \
    go install github.com/pressly/goose/v3/cmd/goose@latest

CMD goose -dir=./migrations postgres "postgres://${HIVE_BACKEND_DB_USER}:${HIVE_BACKEND_DB_PASSWORD}@${HIVE_BACKEND_DB_HOST}:${HIVE_BACKEND_DB_PORT}/${HIVE_BACKEND_DB_NAME}?sslmode=disable" up && \
    ./hive-backend