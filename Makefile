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
GOLANGCI_TAG:=1.50.1
GOLANGCI_LINT_VERSION:=v1.50.1

# Check local bin version
ifneq ($(wildcard $(GOLANGCI_BIN)),)
GOLANGCI_BIN_VERSION:=$(shell $(GOLANGCI_BIN) --version)
ifneq ($(GOLANGCI_BIN_VERSION),)
GOLANGCI_BIN_VERSION_SHORT:=$(shell echo "$(GOLANGCI_BIN_VERSION)" | sed -E 's/.* version (.*) built from .* on .*/\1/g')
else
GOLANGCI_BIN_VERSION_SHORT:=0
endif
ifneq "$(GOLANGCI_TAG)" "$(word 1, $(sort $(GOLANGCI_TAG) $(GOLANGCI_BIN_VERSION_SHORT)))"
GOLANGCI_BIN:=
endif
endif

default: help

.PHONY: install-lint
install-lint: ## install golangci-lint binary
ifeq ($(wildcard $(GOLANGCI_BIN)),)
	$(info Downloading golangci-lint v$(GOLANGCI_TAG))
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_TAG)
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
endif

.PHONY: lint
lint: install-lint
	$(info Running lint...)
	$(GOLANGCI_BIN) run --new-from-rev=origin/master --config=.golangci.yaml ./...

.PHONY: lint-full
lint-full: install-lint
	$(GOLANGCI_BIN) run --config=.golangci.yaml ./...

.PHONY: test
test:
	@go test -v ./...

.PHONY: build
build:
	@echo "Building $(APP) $(VERSION) for $(OS)/$(ARCH)"
	@mkdir -p $(LOCAL_BIN)
	@GOOS=$(OS) GOARCH=$(ARCH) go build -o $(LOCAL_BIN)/$(APP) ./cmd/main.go

.PHONY: run
run:
	@$(LOCAL_BIN)/$(APP)

.PHONY: clean
clean:
	@rm -rf $(LOCAL_BIN)/$(APP)

.PHONY: migrate-up
migrate-up:
	@goose -dir "$(MIGRATIONS)" postgres "host=$(PG_HOST) port=$(PG_PORT) user=$(PG_USER) password=$(PG_PASSWORD) dbname=$(PG_DATABASE) sslmode=disable" up

.PHONY: migrate-down
migrate-down:
	@goose -dir "$(MIGRATIONS)" postgres "host=$(PG_HOST) port=$(PG_PORT) user=$(PG_USER) password=$(PG_PASSWORD) dbname=$(PG_DATABASE) sslmode=disable" down

.PHONY: compose-up
compose-up:
	@docker-compose up -d

.PHONY: compose-down
compose-down:
	@docker-compose down

.PHONY: compose-clean
compose-clean:
	@docker-compose down -v --rmi all

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  lint            Run golangci-lint"
	@echo "  test            Run tests"
	@echo "  build           Build the project"
	@echo "  run             Run the project"
	@echo "  clean           Remove build artifacts"
	@echo "  migrate-up      Run goose up"
	@echo "  migrate-down    Run goose down"
	@echo "  compose-up      Run docker-compose up"
	@echo "  compose-down    Run docker-compose down"
	@echo "  compose-clean   Run docker-compose down -v"
	@echo "  help            Show this help message"