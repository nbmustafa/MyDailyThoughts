BINARY        := k8sdiag
BUILD_DIR     := ./bin
CMD_PATH      := ./cmd/k8sdiag
VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS       := -ldflags "-s -w -X main.version=$(VERSION)"
PLATFORMS     := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
INSTALL_DIR   ?= /usr/local/bin

.PHONY: all build install clean lint test cross-build release help

## Default
all: build

## ── Build ─────────────────────────────────────────────────────────────────────

build: ## Build binary for host platform
	@echo "→ Building $(BINARY) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)
	@echo "✓ Binary written to $(BUILD_DIR)/$(BINARY)"

install: build ## Install binary to INSTALL_DIR (default: /usr/local/bin)
	@echo "→ Installing to $(INSTALL_DIR)/$(BINARY)..."
	@install -m 755 $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "✓ Installed. Run: $(BINARY) --help"

cross-build: ## Cross-compile for all target platforms
	@mkdir -p $(BUILD_DIR)
	@$(foreach PLATFORM,$(PLATFORMS), \
		GOOS=$(word 1,$(subst /, ,$(PLATFORM))) \
		GOARCH=$(word 2,$(subst /, ,$(PLATFORM))) \
		go build $(LDFLAGS) \
			-o $(BUILD_DIR)/$(BINARY)-$(word 1,$(subst /, ,$(PLATFORM)))-$(word 2,$(subst /, ,$(PLATFORM))) \
			$(CMD_PATH) && \
		echo "✓ Built $(PLATFORM)"; \
	)

## ── Quality ───────────────────────────────────────────────────────────────────

test: ## Run unit tests
	go test ./... -v -timeout 30s

lint: ## Run golangci-lint (install separately)
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go.mod
	go mod tidy

## ── Cleanup ───────────────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

## ── Docker ────────────────────────────────────────────────────────────────────

docker-build: ## Build Docker image
	docker build -t $(BINARY):$(VERSION) .

## ── Help ──────────────────────────────────────────────────────────────────────

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
