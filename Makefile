# =============================================================================
# Makefile for echoip
# =============================================================================

# Project configuration
PROJECTNAME := echoip
PROJECTORG := apimgr
REGISTRY := ghcr.io

# Version management
VERSION := $(shell cat release.txt 2>/dev/null || echo "0.0.1")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build configuration
GOBIN := $(shell go env GOPATH)/bin
BINARY_DIR := binaries
RELEASE_DIR := releases
SRC_DIR := src

# Build flags
LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE) -w -s

# Platform targets
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64 \
	freebsd/amd64 \
	freebsd/arm64

.PHONY: all
all: clean lint test build

# =============================================================================
# Build targets
# =============================================================================

.PHONY: build
build: clean-binaries
	@echo "Building $(PROJECTNAME) $(VERSION)..."
	@mkdir -p $(BINARY_DIR) $(RELEASE_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		OUTPUT_NAME=$(BINARY_DIR)/$(PROJECTNAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then OUTPUT_NAME=$$OUTPUT_NAME.exe; fi; \
		echo "  Building $$GOOS/$$GOARCH..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build \
			-ldflags "$(LDFLAGS)" \
			-a -installsuffix cgo \
			-o $$OUTPUT_NAME \
			./$(SRC_DIR) || exit 1; \
		cp $$OUTPUT_NAME $(RELEASE_DIR)/ || exit 1; \
	done
	@# Create host platform binary
	@echo "  Creating host binary..."
	@CGO_ENABLED=0 go build \
		-ldflags "$(LDFLAGS)" \
		-a -installsuffix cgo \
		-o $(BINARY_DIR)/$(PROJECTNAME) \
		./$(SRC_DIR)
	@echo "✓ Build complete: $(VERSION)"

# =============================================================================
# Testing targets
# =============================================================================

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v -race -timeout 5m ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

.PHONY: check-fmt
check-fmt:
	@echo "Checking formatting..."
	@bash -c "diff --line-format='%L' <(echo -n) <(gofmt -d -s .)"

.PHONY: lint
lint: check-fmt vet
	@echo "✓ Lint checks passed"

# =============================================================================
# Docker targets
# =============================================================================

.PHONY: docker
docker:
	@echo "Building multi-platform Docker images..."
	@docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(REGISTRY)/$(PROJECTORG)/$(PROJECTNAME):latest \
		-t $(REGISTRY)/$(PROJECTORG)/$(PROJECTNAME):$(VERSION) \
		--push \
		.
	@echo "✓ Docker images pushed to $(REGISTRY)/$(PROJECTORG)/$(PROJECTNAME):$(VERSION)"

.PHONY: docker-dev
docker-dev:
	@echo "Building development Docker image..."
	@docker build \
		--build-arg VERSION=$(VERSION)-dev \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(PROJECTNAME):dev \
		.
	@echo "✓ Docker development image built: $(PROJECTNAME):dev"

.PHONY: docker-test
docker-test: docker-dev
	@echo "Testing Docker image..."
	@docker-compose -f docker-compose.test.yml up -d
	@echo "Waiting for service..."
	@timeout 30 bash -c 'until curl -sf http://localhost:64181/health; do sleep 1; done' || (docker-compose -f docker-compose.test.yml logs && exit 1)
	@echo "✓ Service is running"
	@docker-compose -f docker-compose.test.yml down
	@rm -rf /tmp/$(PROJECTNAME)/rootfs
	@echo "✓ Docker test complete"

# =============================================================================
# Release targets
# =============================================================================

.PHONY: release
release: build
	@echo "Creating GitHub release $(VERSION)..."
	@# Create source archives
	@git archive --format=tar.gz --prefix=$(PROJECTNAME)-$(VERSION)/ HEAD > $(RELEASE_DIR)/$(PROJECTNAME)-$(VERSION)-src.tar.gz
	@git archive --format=zip --prefix=$(PROJECTNAME)-$(VERSION)/ HEAD > $(RELEASE_DIR)/$(PROJECTNAME)-$(VERSION)-src.zip
	@# Delete existing release if it exists
	@gh release delete $(VERSION) -y 2>/dev/null || true
	@# Create new release
	@gh release create $(VERSION) \
		--title "$(VERSION)" \
		--notes "Release $(VERSION)" \
		$(RELEASE_DIR)/*
	@echo "✓ Release $(VERSION) created successfully"
	@# Auto-increment version after successful release
	@$(MAKE) version-bump

.PHONY: version-bump
version-bump:
	@echo "Auto-incrementing version..."
	@CURRENT_VERSION=$$(cat release.txt); \
	MAJOR=$$(echo $$CURRENT_VERSION | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT_VERSION | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT_VERSION | cut -d. -f3); \
	NEW_PATCH=$$((PATCH + 1)); \
	NEW_VERSION="$$MAJOR.$$MINOR.$$NEW_PATCH"; \
	echo "$$NEW_VERSION" > release.txt; \
	echo "✓ Version bumped: $$CURRENT_VERSION → $$NEW_VERSION"

# =============================================================================
# GeoIP database download (sapics/ip-location-db via jsdelivr CDN)
# =============================================================================

.PHONY: geoip-download
geoip-download:
	@echo "Downloading GeoIP databases from sapics/ip-location-db (4 databases)..."
	@mkdir -p data/geoip
	@echo "  Downloading geolite2-city-ipv4.mmdb (~50MB)..."
	@curl -fsSL -o data/geoip/geolite2-city-ipv4.mmdb \
		"https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv4.mmdb"
	@echo "  Downloading geolite2-city-ipv6.mmdb (~40MB)..."
	@curl -fsSL -o data/geoip/geolite2-city-ipv6.mmdb \
		"https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv6.mmdb"
	@echo "  Downloading geo-whois-asn-country.mmdb (~8MB)..."
	@curl -fsSL -o data/geoip/geo-whois-asn-country.mmdb \
		"https://cdn.jsdelivr.net/npm/@ip-location-db/geo-whois-asn-country-mmdb/geo-whois-asn-country.mmdb"
	@echo "  Downloading asn.mmdb (~5MB)..."
	@curl -fsSL -o data/geoip/asn.mmdb \
		"https://cdn.jsdelivr.net/npm/@ip-location-db/asn-mmdb/asn.mmdb"
	@echo "✓ All 4 GeoIP databases downloaded (~103MB total)"
	@ls -lh data/geoip/*.mmdb

# =============================================================================
# Installation targets
# =============================================================================

.PHONY: install
install: build
	@echo "Installing $(PROJECTNAME)..."
	@install -m 755 $(BINARY_DIR)/$(PROJECTNAME) $(GOBIN)/$(PROJECTNAME)
	@echo "✓ Installed to $(GOBIN)/$(PROJECTNAME)"

# =============================================================================
# Development targets
# =============================================================================

.PHONY: run
run:
	@echo "Running $(PROJECTNAME) in development mode..."
	@go run ./$(SRC_DIR) -l :8080

.PHONY: run-full
run-full: geoip-download
	@echo "Running $(PROJECTNAME) with all features..."
	@go run ./$(SRC_DIR) \
		-a data/asn.mmdb \
		-c data/city.mmdb \
		-f data/country.mmdb \
		-H x-forwarded-for \
		-r \
		-s \
		-p \
		-l :8080

# =============================================================================
# Cleanup targets
# =============================================================================

.PHONY: clean-binaries
clean-binaries:
	@echo "Cleaning binaries..."
	@rm -rf $(BINARY_DIR) $(RELEASE_DIR)

.PHONY: clean-docker
clean-docker:
	@echo "Cleaning Docker artifacts..."
	@docker-compose -f docker-compose.test.yml down 2>/dev/null || true
	@docker rmi $(PROJECTNAME):dev 2>/dev/null || true
	@rm -rf /tmp/$(PROJECTNAME)/rootfs

.PHONY: clean
clean: clean-binaries clean-docker
	@echo "Cleaning all build artifacts..."
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

# =============================================================================
# Help target
# =============================================================================

.PHONY: help
help:
	@echo "$(PROJECTNAME) - Makefile targets:"
	@echo ""
	@echo "Build:"
	@echo "  make build          - Build binaries for all platforms"
	@echo "  make install        - Install binary to \$$GOPATH/bin"
	@echo ""
	@echo "Test:"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linters"
	@echo ""
	@echo "Docker:"
	@echo "  make docker         - Build and push multi-arch Docker images"
	@echo "  make docker-dev     - Build development Docker image"
	@echo "  make docker-test    - Test Docker image locally"
	@echo ""
	@echo "Release:"
	@echo "  make release        - Create GitHub release with binaries"
	@echo "  make version-bump   - Increment patch version"
	@echo ""
	@echo "Development:"
	@echo "  make run            - Run in development mode"
	@echo "  make run-full       - Run with all features (requires GeoIP data)"
	@echo "  make geoip-download - Download GeoIP databases"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean          - Clean all build artifacts"
	@echo "  make clean-binaries - Clean binary files only"
	@echo "  make clean-docker   - Clean Docker artifacts"
	@echo ""
	@echo "Current version: $(VERSION)"
