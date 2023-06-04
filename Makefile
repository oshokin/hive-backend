export GO111MODULE=on

APP:=hive-backend
OS:=$(shell go env GOOS)
ARCH:=$(shell go env GOARCH)
PG_HOST:=localhost
PG_PORT:=5432
PG_USER:=admin
PG_PASSWORD:=hard-password
PG_DATABASE:=hive
LOCAL_BIN:=$(CURDIR)/bin
MIGRATIONS:=migrations
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
GOLANGCI_TAG:=1.52.2
GOLANGCI_CONFIG:=.golangci.yaml
GOLANGCI_STRICT_CONFIG:=.golangci-strict.yaml
GOOSE_BIN:=$(LOCAL_BIN)/goose
GOOSE_TAG:=3.10.0
K6_BIN:=$(LOCAL_BIN)/k6
K6_TAG:=0.43.1
USER_ID:=$(shell id -u)
USER_GROUP_ID:=$(shell id -g)
DB_MASTER_CONTAINER:=hive-backend-db-master
DB_SYNC_CONTAINER:=hive-backend-db-sync
DB_ASYNC_CONTAINER:=hive-backend-db-async
DB_SYNC_DATA:=$(CURDIR)/data/db-sync
DB_ASYNC_DATA:=$(CURDIR)/data/db-async
DB_TEMP_DATA:=$(CURDIR)/data/temp
DB_MASTER_BACKUP_DIR:=/backup

ifneq ($(wildcard $(GOLANGCI_BIN)),)
GOLANGCI_BIN_VERSION:=$(shell $(GOLANGCI_BIN) --version)
ifneq ($(GOLANGCI_BIN_VERSION),)
GOLANGCI_BIN_VERSION_SHORT:=$(shell echo "$(GOLANGCI_BIN_VERSION)" | sed -E 's/.* version (.*) built .* from .*/\1/g')
else
GOLANGCI_BIN_VERSION_SHORT:=0
endif
ifneq "$(GOLANGCI_TAG)" "$(word 1, $(sort $(GOLANGCI_TAG) $(GOLANGCI_BIN_VERSION_SHORT)))"
GOLANGCI_BIN:=
endif
endif

ifneq ($(wildcard $(GOOSE_BIN)),)
GOOSE_BIN_VERSION:=$(shell $(GOOSE_BIN) --version)
ifneq ($(GOOSE_BIN_VERSION),)
GOOSE_BIN_VERSION_SHORT:=$(shell echo "$(GOOSE_BIN_VERSION)" | sed -E 's/goose version:v(.*)/\1/g')
else
GOOSE_BIN_VERSION_SHORT:=0
endif
ifneq "$(GOOSE_TAG)" "$(word 1, $(sort $(GOOSE_TAG) $(GOOSE_BIN_VERSION_SHORT)))"
GOOSE_BIN:=
endif
endif

ifneq ($(wildcard $(K6_BIN)),)
K6_BIN_VERSION:=$(shell $(K6_BIN) version)
ifneq ($(K6_BIN_VERSION),)
K6_BIN_VERSION_SHORT:=$(shell echo "$(K6_BIN_VERSION)" | sed -E 's/^k6 v([0-9.]*).*/\1/')
else
K6_BIN_VERSION_SHORT:=0
endif
ifneq "$(K6_TAG)" "$(word 1, $(sort $(K6_TAG) $(K6_BIN_VERSION_SHORT)))"
K6_BIN:=
endif
endif

default: help

.PHONY: install-lint
install-lint:
ifeq ($(wildcard $(GOLANGCI_BIN)),)
	$(info Downloading golangci-lint v$(GOLANGCI_TAG))
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_TAG)
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
endif

.PHONY: lint
lint: install-lint
ifeq ($(filter strict,$(MAKECMDGOALS)),strict)
	$(info Running lint in strict mode...)
	$(GOLANGCI_BIN) run --new-from-rev=origin/master --config=$(GOLANGCI_STRICT_CONFIG) ./...
else
	$(info Running lint in normal mode...)
	$(GOLANGCI_BIN) run --new-from-rev=origin/master --config=$(GOLANGCI_CONFIG) ./...
endif

.PHONY: lint-full
lint-full: install-lint
ifeq ($(filter strict,$(MAKECMDGOALS)),strict)
	$(info Running lint-full in strict mode...)
	$(GOLANGCI_BIN) run --config=$(GOLANGCI_STRICT_CONFIG) ./...
else
	$(info Running lint-full in normal mode...)
	$(GOLANGCI_BIN) run --config=$(GOLANGCI_CONFIG) ./...
endif

.PHONY: test
test:
	@go test -v ./...

.PHONY: build
build:
	$(info Building $(APP) for $(OS)/$(ARCH))
	@mkdir -p $(LOCAL_BIN)
	@GOOS=$(OS) GOARCH=$(ARCH) go build -o $(LOCAL_BIN)/$(APP) ./cmd/main.go

.PHONY: run
run:
	@mkdir -p $(LOCAL_BIN)
	@$(LOCAL_BIN)/$(APP)

.PHONY: clean
clean:
	@mkdir -p $(LOCAL_BIN)
	@rm -rf $(LOCAL_BIN)/$(APP)

.PHONY: install-goose
install-goose:
ifeq ($(wildcard $(GOOSE_BIN)),)
	$(info Downloading goose v$(GOOSE_TAG))
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) $ go install github.com/pressly/goose/v3/cmd/goose@v$(GOOSE_TAG)
GOOSE_BIN:=$(LOCAL_BIN)/goose
endif

.PHONY: migrate-up
migrate-up: install-goose
	$(GOOSE_BIN) -dir "$(MIGRATIONS)" postgres "host=$(PG_HOST) port=$(PG_PORT) user=$(PG_USER) password=$(PG_PASSWORD) dbname=$(PG_DATABASE) sslmode=disable" up

.PHONY: migrate-down
migrate-down: install-goose
	$(GOOSE_BIN) -dir "$(MIGRATIONS)" postgres "host=$(PG_HOST) port=$(PG_PORT) user=$(PG_USER) password=$(PG_PASSWORD) dbname=$(PG_DATABASE) sslmode=disable" down

.PHONY: compose-up
compose-up:
	@docker compose up -d

.PHONY: compose-down
compose-down:
	@docker compose down

.PHONY: compose-clean
compose-clean:
	@docker compose down -v --rmi all

.PHONY: install-k6
install-k6:
ifeq ($(wildcard $(K6_BIN)),)
	cd $(LOCAL_BIN) && \
	GOBIN=$(LOCAL_BIN) $ go install go.k6.io/xk6/cmd/xk6@latest && \
	$(LOCAL_BIN)/xk6 build --with github.com/grafana/xk6-sql && \
	rm -rf $(LOCAL_BIN)/xk6
K6_BIN:=$(LOCAL_BIN)/k6
endif

.PHONY: sync-replicas
sync-replicas:
	@if [ $$(docker ps -q -f name=$(DB_SYNC_CONTAINER)) ]; then docker stop $(DB_SYNC_CONTAINER); fi
	@if [ $$(docker ps -q -f name=$(DB_ASYNC_CONTAINER)) ]; then docker stop $(DB_ASYNC_CONTAINER); fi
	@mkdir -p $(DB_SYNC_DATA) $(DB_ASYNC_DATA)
	@if docker inspect --format '{{.State.Health.Status}}' $(DB_MASTER_CONTAINER) | grep -q healthy; then \
	    docker exec $(DB_MASTER_CONTAINER) bash -c "rm -rf $(DB_MASTER_BACKUP_DIR) && pg_basebackup -h localhost -D $(DB_MASTER_BACKUP_DIR) -U replicator -v -P --wal-method=stream" && \
		docker cp $(DB_MASTER_CONTAINER):$(DB_MASTER_BACKUP_DIR) $(DB_TEMP_DATA) && \
		docker exec $(DB_MASTER_CONTAINER) bash -c "rm -rf $(DB_MASTER_BACKUP_DIR)" && \
		sudo chown -R $(USER_ID):$(USER_GROUP_ID) $(DB_TEMP_DATA) && \
		sudo rm -rf $(DB_SYNC_DATA) && \
		sudo rm -rf $(DB_ASYNC_DATA) && \
		cp -r $(DB_TEMP_DATA)/. $(DB_SYNC_DATA) && \
		cp -r $(DB_TEMP_DATA)/. $(DB_ASYNC_DATA) && \
		rm -rf $(DB_TEMP_DATA) && \
		docker compose up -d $(DB_SYNC_CONTAINER) $(DB_ASYNC_CONTAINER); \
	else \
		@echo "$(DB_MASTER_CONTAINER) is not running or not healthy."; \
		@exit 1; \
	fi

.PHONY: stop-replicas
stop-replicas:
	@if [ $$(docker ps -q -f name=$(DB_SYNC_CONTAINER)) ]; then docker stop $(DB_SYNC_CONTAINER); fi
	@if [ $$(docker ps -q -f name=$(DB_ASYNC_CONTAINER)) ]; then docker stop $(DB_ASYNC_CONTAINER); fi
	sudo chown -R $(USER_ID):$(USER_GROUP_ID) $(DB_SYNC_DATA) && \
	sudo chown -R $(USER_ID):$(USER_GROUP_ID) $(DB_ASYNC_DATA)

.PHONY: clean-replicas
clean-replicas:
	@if [ $$(docker ps -q -f name=$(DB_SYNC_CONTAINER)) ]; then docker clean $(DB_SYNC_CONTAINER); fi
	@if [ $$(docker ps -q -f name=$(DB_ASYNC_CONTAINER)) ]; then docker clean $(DB_ASYNC_CONTAINER); fi
	sudo rm -rf $(DB_SYNC_DATA) && \
	sudo rm -rf $(DB_ASYNC_DATA)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  help                    Show this help message"
	@echo "  install-lint            Download and install golangci-lint to $(LOCAL_BIN) directory if it's not already installed"
	@echo "  lint                    Run golangci-lint with normal checks and compare changes against master branch."
	@echo "  lint strict             Same as 'lint', but with more strict checks."
	@echo "  lint-full               Run golangci-lint with normal checks for all files in the repository."
	@echo "  lint-full strict        Same as 'lint-full', but with more strict checks."
	@echo "  test                    Run unit tests"
	@echo "  build                   Build the $(APP) binary for $(OS)/$(ARCH)"
	@echo "  run                     Run the $(APP) binary"
	@echo "  clean                   Remove the $(APP) binary"
	@echo "  install-goose           Download and install goose to $(LOCAL_BIN) directory if it's not already installed"
	@echo "  migrate-up              Run goose up"
	@echo "  migrate-down            Run goose down"
	@echo "  compose-up              Run docker-compose up"
	@echo "  compose-down            Run docker-compose down"
	@echo "  compose-clean           Run docker-compose down -v"
	@echo "  install-k6              Download and install k6 to $(LOCAL_BIN) directory if it's not already installed."
	@echo "  sync-replicas           Stop, backup and sync the primary database to the synchronous and asynchronous replicas"
	@echo "  stop-replicas           Stop the synchronous and asynchronous replicas and change ownership of their data directories"
	@echo "  clean-replicas          Stop and clean the synchronous and asynchronous replicas and remove their data directories"
