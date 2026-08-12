# groot-trigger — build, test, release scaffolding (GNU Make).
# Feature code lands via GSD against docs/SPECIFICATIONS.md.

APP_NAME := groot-trigger
BIN_DIR  := bin
MODULE   := github.com/hrodrig/groot-trigger

VERSION_RAW ?= $(shell cat VERSION 2>/dev/null | tr -d '\n\r')
VERSION     := $(patsubst v%,%,$(VERSION_RAW))
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BRANCH      := $(shell git symbolic-ref --short HEAD 2>/dev/null || echo unknown)
BUILDDATE   := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT)' \
	-X 'main.branch=$(BRANCH)' \
	-X 'main.buildDate=$(BUILDDATE)'

GOLANGCI_LINT_VERSION ?= v2.5.0
COVER_MIN ?= 80
IMAGE     ?= $(APP_NAME):local

DOCKER_BUILD_ARGS := \
	--build-arg APP_VERSION=$(VERSION) \
	--build-arg GIT_COMMIT=$(COMMIT) \
	--build-arg GIT_BRANCH=$(BRANCH) \
	--build-arg BUILD_DATE=$(BUILDDATE)

GREEN  := \033[0;32m
YELLOW := \033[0;33m
RESET  := \033[0m
ifneq ($(NO_COLOR),)
  GREEN  :=
  YELLOW :=
  RESET  :=
endif

.DEFAULT_GOAL := help

.PHONY: help all build test cover fmt fmt-check lint-fix lint vet run clean install \
	docker-build goreleaser-check release-check ci

help:
	@echo "$(GREEN)groot-trigger$(RESET) — on-demand HTTP → groot Job"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "$(YELLOW)Build:$(RESET)"
	@echo "  $(GREEN)build$(RESET)              Build $(BIN_DIR)/$(APP_NAME)"
	@echo "  $(GREEN)install$(RESET)            Install to GOPATH/bin"
	@echo "  $(GREEN)clean$(RESET)              Remove $(BIN_DIR)/"
	@echo ""
	@echo "$(YELLOW)Test / quality:$(RESET)"
	@echo "  $(GREEN)test$(RESET)               go test ./... -race -count=1"
	@echo "  $(GREEN)cover$(RESET)              Coverage with COVER_MIN=$(COVER_MIN)"
	@echo "  $(GREEN)fmt-check$(RESET)          Fail if gofmt -s would change files"
	@echo "  $(GREEN)lint$(RESET)               golangci-lint ($(GOLANGCI_LINT_VERSION))"
	@echo "  $(GREEN)vet$(RESET)                go vet ./..."
	@echo "  $(GREEN)ci$(RESET)                 fmt-check + vet + test"
	@echo ""
	@echo "$(YELLOW)Docker / release:$(RESET)"
	@echo "  $(GREEN)docker-build$(RESET)       Local image ($(IMAGE))"
	@echo "  $(GREEN)goreleaser-check$(RESET)   goreleaser check"
	@echo "  $(GREEN)release-check$(RESET)      goreleaser-check + fmt-check + lint + cover"
	@echo ""
	@echo "Contract: docs/SPECIFICATIONS.md  |  VERSION=$(VERSION)"

all: build

build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

install: build
	install -m 755 $(BIN_DIR)/$(APP_NAME) $$(go env GOPATH)/bin/$(APP_NAME)

clean:
	rm -rf $(BIN_DIR)/$(APP_NAME) dist coverage.out

run: build
	$(BIN_DIR)/$(APP_NAME)

test:
	go test ./... -race -count=1

cover:
	go test ./... -race -count=1 -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | tail -1
	@pct=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$NF}' | tr -d '%'); \
	awk -v p="$$pct" -v m="$(COVER_MIN)" 'BEGIN { if ((p+0) < (m+0)) { printf "coverage %.1f%% < COVER_MIN %s%%\n", p, m; exit 1 } }'

fmt:
	gofmt -w .

fmt-check:
	@out=$$(gofmt -s -l .); if [ -n "$$out" ]; then echo "$$out"; exit 1; fi

lint-fix:
	gofmt -s -w .

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found; install v$(GOLANGCI_LINT_VERSION:v%=%)"; exit 1; }
	golangci-lint run

vet:
	go vet ./...

ci: fmt-check vet test

docker-build:
	docker info >/dev/null 2>&1 || { echo "Error: Docker is not running"; exit 1; }
	docker build $(DOCKER_BUILD_ARGS) -t $(IMAGE) .

goreleaser-check:
	goreleaser check

release-check: goreleaser-check fmt-check lint cover
	@echo "$(GREEN)release-check OK$(RESET)"
