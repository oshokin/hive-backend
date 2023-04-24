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
K6_BIN:=$(LOCAL_BIN)/k6
K6_TAG:=0.43.1

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

ifneq ($(wildcard $(K6_BIN)),)
K6_BIN_VERSION:=$(shell $(K6_BIN) version)
ifneq ($(K6_BIN_VERSION),)
K6_BIN_VERSION_SHORT:=$(shell echo "$(K6_BIN_VERSION)" | sed -E 's/k6 v(.*) \(.*\)/\1/g')
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

.PHONY: migrate-up
migrate-up:
	@goose -dir "$(MIGRATIONS)" postgres "host=$(PG_HOST) port=$(PG_PORT) user=$(PG_USER) password=$(PG_PASSWORD) dbname=$(PG_DATABASE) sslmode=disable" up

.PHONY: migrate-down
migrate-down:
	@goose -dir "$(MIGRATIONS)" postgres "host=$(PG_HOST) port=$(PG_PORT) user=$(PG_USER) password=$(PG_PASSWORD) dbname=$(PG_DATABASE) sslmode=disable" down

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
	$(info Downloading k6 v$(K6_TAG))
	@curl -OL https://github.com/loadimpact/k6/releases/download/v$(K6_TAG)/k6-v$(K6_TAG)-$(OS)-$(ARCH).tar.gz
	@tar -xzf k6-v$(K6_TAG)-$(OS)-$(ARCH).tar.gz
	@mkdir -p $(LOCAL_BIN)
	@mv k6-v$(K6_TAG)-$(OS)-$(ARCH)/k6 $(LOCAL_BIN)
	@rm -rf k6-v$(K6_TAG)-$(OS)-$(ARCH) k6-v$(K6_TAG)-$(OS)-$(ARCH).tar.gz
K6_BIN:=$(LOCAL_BIN)/k6
endif

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  install-lint            Download and install golangci-lint to $(LOCAL_BIN) directory if it's not already installed"
	@echo "  lint                    Run golangci-lint with normal checks and compare changes against master branch."
	@echo "  lint strict             Same as 'lint', but with more strict checks."
	@echo "  lint-full               Run golangci-lint with normal checks for all files in the repository."
	@echo "  lint-full strict        Same as 'lint-full', but with more strict checks."
	@echo "  test                    Run unit tests"
	@echo "  install-k6              Download and install k6 to $(LOCAL_BIN) directory if it's not already installed."
	@echo "  build                   Build the $(APP) binary for $(OS)/$(ARCH)"
	@echo "  run                     Run the $(APP) binary"
	@echo "  clean                   Remove the $(APP) binary"
	@echo "  migrate-up              Run goose up"
	@echo "  migrate-down            Run goose down"
	@echo "  compose-up              Run docker-compose up"
	@echo "  compose-down            Run docker-compose down"
	@echo "  compose-clean           Run docker-compose down -v"
	@echo "  help                    Show this help message"