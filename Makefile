# PGSD Build System Makefile

# Build configuration
GO = go
GOFLAGS = -v
BINDIR = bin
ARTIFACTSDIR = artifacts
ISODIR = iso
WORKDIR = work

# Version information
VERSION = 0.1.0
GIT_COMMIT != git rev-parse --short HEAD 2>/dev/null || echo "unknown"
BUILD_DATE != date -u +"%Y-%m-%dT%H:%M:%SZ"
LDFLAGS = -X main.Version=$(VERSION) \
          -X main.GitCommit=$(GIT_COMMIT) \
          -X main.BuildDate=$(BUILD_DATE)

# Binary names
PGSDBUILD = $(BINDIR)/pgsdbuild
PGSDINST = $(BINDIR)/pgsd-inst

# Phony targets
.PHONY: all build clean install test help \
        build-pgsdbuild build-installer \
        list-images list-variants \
        build-image build-iso

# Build all binaries
all: build

# Build both tools
build: build-pgsdbuild build-installer
	@echo "Build complete!"

# Build pgsdbuild
build-pgsdbuild:
	@echo "Building pgsdbuild $(VERSION) (commit: $(GIT_COMMIT))..."
	@mkdir -p $(BINDIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(PGSDBUILD) ./cmd/pgsdbuild

# Build pgsd-inst installer
build-installer:
	@echo "Building pgsd-inst..."
	@mkdir -p $(BINDIR)
	$(GO) build $(GOFLAGS) -o $(PGSDINST) ./installer/pgsd-inst

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINDIR) $(ARTIFACTSDIR) $(ISODIR) $(WORKDIR) pgsdbuild pgsd-inst
	@echo "Clean complete!"

# Install binaries to /usr/local/bin (requires root)
install: build
	@echo "Installing binaries to /usr/local/bin..."
	install -m 0755 $(PGSDBUILD) /usr/local/bin/
	install -m 0755 $(PGSDINST) /usr/local/bin/
	@echo "Installation complete!"

# Uninstall binaries from /usr/local/bin (requires root)
uninstall:
	@echo "Uninstalling binaries..."
	rm -f /usr/local/bin/pgsdbuild
	rm -f /usr/local/bin/pgsd-inst
	@echo "Uninstall complete!"

# List available images
list-images: build-pgsdbuild
	@$(PGSDBUILD) list-images

# List available variants
list-variants: build-pgsdbuild
	@$(PGSDBUILD) list-variants

# Build a specific image (usage: make build-image IMAGE=pgsd-desktop)
build-image: build-pgsdbuild
	@if [ -z "$(IMAGE)" ]; then \
		echo "Error: IMAGE not specified. Usage: make build-image IMAGE=pgsd-desktop"; \
		exit 1; \
	fi
	@echo "Building image: $(IMAGE)"
	@$(PGSDBUILD) image $(IMAGE)

# Build a specific ISO (usage: make build-iso VARIANT=pgsd-bootenv-arcan)
build-iso: build-pgsdbuild
	@if [ -z "$(VARIANT)" ]; then \
		echo "Error: VARIANT not specified. Usage: make build-iso VARIANT=pgsd-bootenv-arcan"; \
		exit 1; \
	fi
	@echo "Building ISO: $(VARIANT)"
	@$(PGSDBUILD) iso $(VARIANT)

# Build all images
build-all-images: build-pgsdbuild
	@echo "Building all images..."
	@for img in images/*.lua; do \
		name=$$(basename $$img .lua); \
		echo "Building $$name..."; \
		$(PGSDBUILD) image $$name || exit 1; \
	done
	@echo "All images built!"

# Build all ISOs
build-all-isos: build-pgsdbuild build-all-images
	@echo "Building all ISOs..."
	@for var in variants/*.lua; do \
		name=$$(basename $$var .lua); \
		echo "Building $$name..."; \
		$(PGSDBUILD) iso $$name || exit 1; \
	done
	@echo "All ISOs built!"

# Format Go code
fmt:
	@echo "Formatting Go code..."
	$(GO) fmt ./...

# Run Go linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Update Go dependencies
deps:
	@echo "Updating dependencies..."
	$(GO) mod tidy
	$(GO) mod download

# Show help
help:
	@echo "PGSD Build System Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  all                 - Build all binaries (default)"
	@echo "  build               - Build pgsdbuild and pgsd-inst"
	@echo "  build-pgsdbuild     - Build only pgsdbuild"
	@echo "  build-installer     - Build only pgsd-inst"
	@echo "  clean               - Remove all build artifacts"
	@echo "  install             - Install binaries to /usr/local/bin (requires root)"
	@echo "  uninstall           - Uninstall binaries from /usr/local/bin (requires root)"
	@echo "  test                - Run Go tests"
	@echo "  fmt                 - Format Go code"
	@echo "  lint                - Run Go linter (requires golangci-lint)"
	@echo "  deps                - Update Go dependencies"
	@echo ""
	@echo "Image/ISO building:"
	@echo "  list-images         - List available images"
	@echo "  list-variants       - List available variants"
	@echo "  build-image         - Build specific image (IMAGE=name)"
	@echo "  build-iso           - Build specific ISO (VARIANT=name)"
	@echo "  build-all-images    - Build all available images"
	@echo "  build-all-isos      - Build all available ISOs"
	@echo ""
	@echo "Examples:"
	@echo "  make build-image IMAGE=pgsd-desktop"
	@echo "  make build-iso VARIANT=pgsd-bootenv-arcan"
	@echo "  make build-all-images"
	@echo "  make build-all-isos"
