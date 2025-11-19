# Simple Joblet Makefile
REMOTE_HOST ?= 192.168.1.161
REMOTE_USER ?= joblet

# Fix GOPATH if IntelliJ IDEA has set it incorrectly
export GOPATH := $(HOME)/go

# Architecture support: amd64, arm64, or both (default: both)
# Usage: ARCH=amd64 make all    # Build only for amd64
#        ARCH=arm64 make all    # Build only for arm64
#        make all               # Build for both (default)
ARCH ?= amd64 arm64

# Version information
# Priority: 1. VERSION env var, 2. VERSION file, 3. git tags, 4. fallback to "dev"
#
# Usage:
#   make version                    # Show current version (from git tags)
#   VERSION=v1.2.3 make all         # Build with custom version (CI/CD)
#   echo "v1.2.3" > VERSION && make # Use VERSION file (no git required)
#
# Note: End users do NOT need git - version is embedded in binary at build time
VERSION ?= $(shell [ -f VERSION ] && cat VERSION || git describe --tags --exact-match 2>/dev/null || git describe --tags --abbrev=0 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Ldflags for version injection
LDFLAGS := -s -w \
	-X github.com/ehsaniara/joblet/pkg/version.Version=$(VERSION) \
	-X github.com/ehsaniara/joblet/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/ehsaniara/joblet/pkg/version.GitTag=$(GIT_TAG) \
	-X github.com/ehsaniara/joblet/pkg/version.BuildDate=$(BUILD_DATE)

.PHONY: all clean deploy test proto help joblet rnx persist state version packages

all: proto joblet rnx persist state
	@echo "✅ Build complete - all binaries ready"

packages: all
	@echo "📦 Building packages (DEB and RPM) for all architectures..."
	@mkdir -p packages/dist
	@echo "📦 Building Debian packages..."
	@./scripts/build-deb.sh amd64 $(VERSION) || (echo "❌ DEB build for amd64 failed" && exit 1)
	@./scripts/build-deb.sh arm64 $(VERSION) || (echo "❌ DEB build for arm64 failed" && exit 1)
	@echo "📦 Building RPM packages..."
	@if command -v rpmbuild >/dev/null 2>&1; then \
		./scripts/build-rpm.sh amd64 $(VERSION) || (echo "❌ RPM build for amd64 failed" && exit 1); \
		./scripts/build-rpm.sh arm64 $(VERSION) || (echo "❌ RPM build for arm64 failed" && exit 1); \
	else \
		echo "❌ rpmbuild not found - RPM packages cannot be built"; \
		echo "   Install rpmbuild: brew install rpm (macOS) or yum/dnf install rpm-build (Linux)"; \
		exit 1; \
	fi
	@echo "✅ Package build complete!"
	@echo "📦 Packages in ./packages/dist/:"
	@DEB_COUNT=0; RPM_COUNT=0; \
	for pkg in packages/dist/*.deb; do \
		if [ -f "$$pkg" ]; then \
			ls -lh "$$pkg" | awk '{print "  " $$9 " (" $$5 ")"}'; \
			DEB_COUNT=$$((DEB_COUNT + 1)); \
		fi; \
	done; \
	for pkg in packages/dist/*.rpm; do \
		if [ -f "$$pkg" ]; then \
			ls -lh "$$pkg" | awk '{print "  " $$9 " (" $$5 ")"}'; \
			RPM_COUNT=$$((RPM_COUNT + 1)); \
		fi; \
	done; \
	echo ""; \
	echo "📊 Summary: $$DEB_COUNT DEB package(s), $$RPM_COUNT RPM package(s)"; \
	if [ $$DEB_COUNT -lt 2 ] || [ $$RPM_COUNT -lt 2 ]; then \
		echo "⚠️  Warning: Expected 2 DEB and 2 RPM packages (one per architecture)"; \
		exit 1; \
	fi

joblet:
	@for arch in $(ARCH); do \
		echo "Building joblet daemon for linux/$$arch..."; \
		mkdir -p bin/linux-$$arch; \
		GOOS=linux GOARCH=$$arch go build -ldflags="$(LDFLAGS) -X github.com/ehsaniara/joblet/pkg/version.Component=joblet" -o bin/linux-$$arch/joblet ./cmd/joblet; \
		echo "✅ joblet built for linux/$$arch (version: $(VERSION))"; \
	done

rnx:
	@for arch in $(ARCH); do \
		echo "Building rnx CLI for linux/$$arch..."; \
		mkdir -p bin/linux-$$arch; \
		GOOS=linux GOARCH=$$arch go build -ldflags="$(LDFLAGS) -X github.com/ehsaniara/joblet/pkg/version.Component=rnx" -o bin/linux-$$arch/rnx ./cmd/rnx; \
		echo "✅ rnx built for linux/$$arch (version: $(VERSION))"; \
	done

persist:
	@for arch in $(ARCH); do \
		echo "Building persist for linux/$$arch..."; \
		mkdir -p bin/linux-$$arch; \
		cd persist && GOOS=linux GOARCH=$$arch go build -ldflags="$(LDFLAGS) -X github.com/ehsaniara/joblet/pkg/version.Component=persist" -o ../bin/linux-$$arch/persist ./cmd/persist && cd ..; \
		echo "✅ persist built for linux/$$arch (version: $(VERSION))"; \
	done

state:
	@for arch in $(ARCH); do \
		echo "Building state for linux/$$arch..."; \
		mkdir -p bin/linux-$$arch; \
		cd state && GOOS=linux GOARCH=$$arch go build -ldflags="$(LDFLAGS) -X github.com/ehsaniara/joblet/pkg/version.Component=state" -o ../bin/linux-$$arch/state ./cmd/state && cd ..; \
		echo "✅ state built for linux/$$arch (version: $(VERSION))"; \
	done

proto:
	@echo "Generating proto files..."
	@./scripts/generate-proto.sh
	@go generate ./internal/proto
	@echo "Proto generation complete"

version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Tag: $(GIT_TAG)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Architectures: $(ARCH)"

clean:
	rm -rf bin/ dist/ api/gen/ internal/proto/gen/ packages/

deploy: all
	@echo "Deploying to $(REMOTE_USER)@$(REMOTE_HOST)..."
	@echo "Note: Deploy will use amd64 binaries by default"
	@ssh $(REMOTE_USER)@$(REMOTE_HOST) "mkdir -p /tmp/joblet/build"
	@echo "Copying binaries..."
	@if [ -d bin/linux-amd64 ]; then \
		scp bin/linux-amd64/joblet bin/linux-amd64/rnx bin/linux-amd64/persist bin/linux-amd64/state $(REMOTE_USER)@$(REMOTE_HOST):/tmp/joblet/build/; \
	else \
		scp bin/joblet bin/rnx bin/persist bin/state $(REMOTE_USER)@$(REMOTE_HOST):/tmp/joblet/build/; \
	fi
	@echo "Stopping services..."
	@ssh $(REMOTE_USER)@$(REMOTE_HOST) 'sudo systemctl stop joblet.service || true'
	@echo "Installing binaries..."
	@ssh $(REMOTE_USER)@$(REMOTE_HOST) 'sudo cp /tmp/joblet/build/* /opt/joblet/bin/ && sudo chmod +x /opt/joblet/bin/*'
	@echo "Starting services..."
	@ssh $(REMOTE_USER)@$(REMOTE_HOST) 'sudo systemctl start joblet.service'
	@echo "✅ Deployment complete (persist and state run as subprocesses)"

test:
	@echo "Running tests..."
	@echo "Testing main module..."
	@JOBLET_TEST_MODE=true go test ./...
	@echo "Testing persist module..."
	@cd persist && JOBLET_TEST_MODE=true go test ./...
	@echo "Testing state module..."
	@cd state && JOBLET_TEST_MODE=true go test ./...
	@echo "✅ All tests complete"

help:
	@echo "Joblet Monorepo Build System"
	@echo ""
	@echo "Targets:"
	@echo "  make all       - Build all binaries (joblet, rnx, persist, state) for all architectures"
	@echo "  make packages - Build all binaries and create DEB + RPM packages for all architectures"
	@echo "                   Output: ./packages/dist/ (2 DEB + 2 RPM packages)"
	@echo "  make joblet    - Build joblet daemon only"
	@echo "  make rnx       - Build rnx CLI only"
	@echo "  make persist   - Build persist only"
	@echo "  make state     - Build state only"
	@echo "  make proto     - Generate proto files"
	@echo "  make version   - Show version information"
	@echo "  make clean     - Remove build artifacts (including ./packages/)"
	@echo "  make test      - Run all tests (all modules)"
	@echo "  make deploy    - Deploy to remote server"
	@echo ""
	@echo "Architecture Support:"
	@echo "  ARCH=amd64 make all     - Build only for amd64"
	@echo "  ARCH=arm64 make all     - Build only for arm64"
	@echo "  make all                - Build for both amd64 and arm64 (default)"
	@echo ""
	@echo "Package Building:"
	@echo "  make packages           - Creates packages in ./packages/dist/:"
	@echo "                            - joblet_*_amd64.deb (Debian/Ubuntu amd64)"
	@echo "                            - joblet_*_arm64.deb (Debian/Ubuntu arm64)"
	@echo "                            - joblet-*-1.x86_64.rpm (RHEL/CentOS/Fedora x86_64)"
	@echo "                            - joblet-*-1.aarch64.rpm (RHEL/CentOS/Fedora aarch64)"
	@echo "  Requirements:           - dpkg-deb (for DEB packages)"
	@echo "                            - rpmbuild (for RPM packages, install via: brew install rpm)"
	@echo ""
	@echo "Version Information:"
	@echo "  Version:    $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Date: $(BUILD_DATE)"
	@echo ""
	@echo "Modules:"
	@echo "  Main:    github.com/ehsaniara/joblet"
	@echo "  Persist: github.com/ehsaniara/joblet/persist"
	@echo ""
	@echo "Proto Version:"
	@echo "  $(shell go list -m github.com/ehsaniara/joblet-proto 2>/dev/null | awk '{print $$2}' || echo 'not found')"
	@echo ""
	@echo "Deployment:"
	@echo "  REMOTE_HOST=$(REMOTE_HOST)"
	@echo "  REMOTE_USER=$(REMOTE_USER)"
