# groot-trigger — build, test, release scaffolding (GNU Make).
# Feature code lands via GSD against docs/SPECIFICATIONS.md.

APP_NAME := groot-trigger
BIN_DIR  := bin
MODULE   := github.com/hrodrig/groot-trigger

VERSION_RAW ?= $(shell cat VERSION 2>/dev/null | tr -d '\n\r')
VERSION     := $(patsubst v%,%,$(VERSION_RAW))
TAG         := v$(VERSION)
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BRANCH      := $(shell git symbolic-ref --short HEAD 2>/dev/null || echo unknown)
BUILDDATE   := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT)' \
	-X 'main.branch=$(BRANCH)' \
	-X 'main.buildDate=$(BUILDDATE)'

GOLANGCI_LINT_VERSION ?= v2.5.0
# Stub has no meaningful coverage yet; raise to 80 after GSD tests land.
COVER_MIN ?= 0
GRYPE_FAIL_ON ?= high
GRYPE_DIR_EXCLUDES := --exclude './bin/**' --exclude './dist/**'
STRICT_RELEASE ?= 0
IMAGE     ?= $(APP_NAME):local
IMAGE_AMD64 ?= $(IMAGE)-amd64
PLATFORMS ?= linux/amd64,linux/arm64

check-docker = @docker info >/dev/null 2>&1 || { echo "Error: Docker is not running. Start Docker and try again."; exit 1; }

DOCKER_BUILD_ARGS := \
	--build-arg APP_VERSION=$(VERSION) \
	--build-arg GIT_COMMIT=$(COMMIT) \
	--build-arg GIT_BRANCH=$(BRANCH) \
	--build-arg BUILD_DATE=$(BUILDDATE)

GREEN  := \033[0;32m
YELLOW := \033[0;33m
CYAN   := \033[0;36m
RESET  := \033[0m
ifneq ($(NO_COLOR),)
  GREEN  :=
  YELLOW :=
  CYAN   :=
  RESET  :=
endif

.DEFAULT_GOAL := help

.PHONY: help all build test cover fmt fmt-check lint-fix lint vet run clean install \
	docker-build docker-build-amd64 docker-scan goreleaser-check release-check ci \
	gocyclo govulncheck vulncheck grype security

help:
	@echo "$(GREEN)groot-trigger$(RESET) — on-demand HTTP → groot Job"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "$(YELLOW)Build:$(RESET)"
	@echo "  $(GREEN)build$(RESET)              Build $(BIN_DIR)/$(APP_NAME)"
	@echo "  $(GREEN)install$(RESET)            Install to GOPATH/bin"
	@echo "  $(GREEN)clean$(RESET)              Remove $(BIN_DIR)/$(APP_NAME) and dist/"
	@echo ""
	@echo "$(YELLOW)Test / quality:$(RESET)"
	@echo "  $(GREEN)test$(RESET)               go test ./... -race -count=1"
	@echo "  $(GREEN)cover$(RESET)              Coverage; gate COVER_MIN=$(COVER_MIN)"
	@echo "  $(GREEN)fmt-check$(RESET)          Fail if gofmt -s would change files"
	@echo "  $(GREEN)lint$(RESET)               golangci-lint ($(GOLANGCI_LINT_VERSION))"
	@echo "  $(GREEN)vet$(RESET)                go vet ./..."
	@echo "  $(GREEN)gocyclo$(RESET)            Fail if complexity > 14"
	@echo "  $(GREEN)ci$(RESET)                 fmt-check + lint + gocyclo + test"
	@echo ""
	@echo "$(YELLOW)Security:$(RESET)"
	@echo "  $(GREEN)govulncheck$(RESET)        golang.org/x/vuln"
	@echo "  $(GREEN)grype$(RESET)              Directory scan (fail-on $(GRYPE_FAIL_ON))"
	@echo "  $(GREEN)security$(RESET)           govulncheck + gocyclo + grype"
	@echo ""
	@echo "$(YELLOW)Docker / release:$(RESET)"
	@echo "  $(GREEN)docker-build$(RESET)       Local image ($(IMAGE))"
	@echo "  $(GREEN)docker-scan$(RESET)        docker-build + Grype image scan"
	@echo "  $(GREEN)goreleaser-check$(RESET)   goreleaser check"
	@echo "  $(GREEN)release-check$(RESET)      VERSION + goreleaser + fmt-check + lint + cover + security"
	@echo ""
	@echo "$(CYAN)Contract:$(RESET) docs/SPECIFICATIONS.md  |  VERSION=$(VERSION)  |  COVER_MIN=$(COVER_MIN)"

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
	go test ./... -race -count=1 -covermode=atomic -coverpkg=./... -coverprofile=coverage.out
	@P=$$(go tool cover -func=coverage.out | tail -1 | sed 's/^.*[[:space:]]\([0-9.]*\)%.*/\1/'); \
		echo "total (merged) statement coverage: $$P% (minimum $(COVER_MIN)%)"; \
		if [ "$(COVER_MIN)" -gt 0 ]; then \
			command -v bc >/dev/null 2>&1 || { echo "COVER_MIN>0 requires bc"; exit 1; }; \
			if [ "$$(echo "$$P < $(COVER_MIN)" | bc)" -eq 1 ]; then \
				echo "coverage below $(COVER_MIN)%"; exit 1; \
			fi; \
		fi

fmt:
	gofmt -w .

fmt-check:
	@out=$$(gofmt -s -l .); if [ -n "$$out" ]; then echo "Run: make lint-fix"; echo "$$out"; exit 1; fi

lint-fix:
	gofmt -s -w .

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...; \
	fi

vet:
	go vet ./...

gocyclo:
	go run github.com/fzipp/gocyclo/cmd/gocyclo@latest -over 14 .

govulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

vulncheck: govulncheck

grype:
	@if command -v grype >/dev/null 2>&1; then \
		grype dir:. $(GRYPE_DIR_EXCLUDES) --fail-on $(GRYPE_FAIL_ON); \
	else \
		echo "grype not found locally, using container image..."; \
		$(check-docker); \
		docker run --rm --pull=always -v "$(CURDIR):/workspace" anchore/grype:latest \
			dir:/workspace $(GRYPE_DIR_EXCLUDES) --fail-on $(GRYPE_FAIL_ON); \
	fi

security: govulncheck gocyclo grype
	@echo "$(GREEN)OK: security (govulncheck, gocyclo, grype)$(RESET)"

ci: fmt-check lint gocyclo test
	@echo "$(GREEN)OK: ci (fmt-check, lint, gocyclo, test)$(RESET)"

docker-build:
	$(check-docker)
	docker build $(DOCKER_BUILD_ARGS) -t $(IMAGE) .

docker-build-amd64:
	$(check-docker)
	docker buildx build $(DOCKER_BUILD_ARGS) --platform linux/amd64 --load -t $(IMAGE_AMD64) .

docker-scan: docker-build
	@command -v grype >/dev/null 2>&1 || { echo "grype required: https://github.com/anchore/grype"; exit 1; }
	grype docker:$(IMAGE) --fail-on $(GRYPE_FAIL_ON)

goreleaser-check:
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser required: https://goreleaser.com/install/"; exit 1; }
	@if git remote get-url origin >/dev/null 2>&1; then \
		goreleaser check; \
	else \
		echo "$(YELLOW)goreleaser check skipped: no git remote origin (local-first)$(RESET)"; \
	fi

# Local-friendly: does not require git remote origin (remote deferred).
release-check:
	@test -f VERSION || { echo "VERSION file is required"; exit 1; }
	@echo "Release version: $(VERSION) (tag: $(TAG))"
	@echo "$(VERSION)" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$$' || { echo "VERSION must be semver (e.g. 0.1.0)"; exit 1; }
	@$(MAKE) goreleaser-check
	@$(MAKE) fmt-check
	@$(MAKE) lint
	@$(MAKE) cover
	@$(MAKE) security
	@if [ "$(STRICT_RELEASE)" = "1" ]; then $(MAKE) docker-scan; fi
	@echo "$(GREEN)release-check OK$(RESET)"
