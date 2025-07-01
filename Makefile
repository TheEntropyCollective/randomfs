# RandomFS - Owner Free File System
# Makefile for building all components

.PHONY: all build build-cli build-server clean test install deps help status validate demo release

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directories and output binaries
BUILD_DIR=./build
CORE_DIR=./randomfs-core
CLI_DIR=./cmd/randomfs-cli
SERVER_DIR=./cmd/randomfs-server

# Binary names
CLI_BIN=randomfs-cli
SERVER_BIN=randomfs-server

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Default target
all: deps
	@$(MAKE) build

# Help target
help:
	@echo "RandomFS - Owner Free File System Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  all           - Build all components (default)"
	@echo "  build         - Build all binaries"
	@echo "  build-cli     - Build CLI tool"
	@echo "  build-server  - Build HTTP server"
	@echo "  test          - Run all tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  install       - Install binaries to GOPATH/bin"
	@echo "  help          - Show this help"
	@echo ""
	@echo "Components:"
	@echo "  • randomfs-core:   Core library for programmatic access"
	@echo "  • randomfs-cli:    Command-line interface with Cobra"
	@echo "  • randomfs-server: HTTP server with REST API"
	@echo "  • randomfs-web:    Static web interface files"

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	cd $(CORE_DIR) && $(GOMOD) download
	cd $(CLI_DIR) && $(GOMOD) download
	cd $(SERVER_DIR) && $(GOMOD) download

# Build all components
build: build-cli build-server
	@echo "✅ All components built successfully!"

# Build CLI tool
build-cli: $(BUILD_DIR)
	@echo "💻 Building RandomFS CLI..."
	cd $(CLI_DIR) && $(GOBUILD) $(BUILD_FLAGS) -o ../../$(BUILD_DIR)/$(CLI_BIN) .
	@echo "🎯 CLI built: $(BUILD_DIR)/$(CLI_BIN)"

# Build HTTP server
build-server: $(BUILD_DIR)
	@echo "🌐 Building RandomFS HTTP Server..."
	cd $(SERVER_DIR) && $(GOBUILD) $(BUILD_FLAGS) -o ../../$(BUILD_DIR)/$(SERVER_BIN) .
	@echo "🚀 Server built: $(BUILD_DIR)/$(SERVER_BIN)"

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...
	cd $(CORE_DIR) && $(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	cd $(CORE_DIR) && $(GOCLEAN)
	cd $(CLI_DIR) && $(GOCLEAN)
	cd $(SERVER_DIR) && $(GOCLEAN)

# Install binaries to GOPATH/bin
install: build
	@echo "📥 Installing binaries..."
	cp $(BUILD_DIR)/$(CLI_BIN) $(GOPATH)/bin/
	cp $(BUILD_DIR)/$(SERVER_BIN) $(GOPATH)/bin/
	@echo "✅ Binaries installed to $(GOPATH)/bin/"

# Development targets
dev-cli: build-cli
	@echo "🔧 Starting CLI in development mode..."
	./$(BUILD_DIR)/$(CLI_BIN) --help

dev-server: build-server
	@echo "🔧 Starting server in development mode..."
	./$(BUILD_DIR)/$(SERVER_BIN) -port 8080 -web ./web-interface/web

# Example usage
demo: build
	@echo "🎬 RandomFS Demo:"
	@echo "CLI:    ./$(BUILD_DIR)/$(CLI_BIN) store example.txt"
	@echo "Server: ./$(BUILD_DIR)/$(SERVER_BIN) -port 8080"

# Project structure validation
validate:
	@echo "🔍 Validating project structure..."
	@test -f $(CORE_DIR)/go.mod || (echo "❌ Core module missing" && exit 1)
	@test -f $(CLI_DIR)/go.mod || (echo "❌ CLI module missing" && exit 1)
	@test -f $(SERVER_DIR)/go.mod || (echo "❌ Server module missing" && exit 1)
	@test -f web-interface/web/index.html || (echo "❌ Web interface missing" && exit 1)
	@echo "✅ Project structure is valid"

# Release build
release: clean deps build
	@echo "🚢 Building release artifacts..."
	mkdir -p release
	cp $(BUILD_DIR)/* release/
	cp -r web-interface/web release/
	cp README.md LICENSE release/
	tar -czf randomfs-release.tar.gz release/
	@echo "📦 Release package: randomfs-release.tar.gz"

# Show component status
status:
	@echo "📊 RandomFS Component Status:"
	@echo "================================"
	@echo "Core Library:     $(if $(wildcard $(CORE_DIR)/go.mod),✅ Ready,❌ Missing)"
	@echo "CLI Tool:         $(if $(wildcard $(CLI_DIR)/go.mod),✅ Ready,❌ Missing)"
	@echo "HTTP Server:      $(if $(wildcard $(SERVER_DIR)/go.mod),✅ Ready,❌ Missing)"
	@echo "Web Interface:    $(if $(wildcard web-interface/web/index.html),✅ Ready,❌ Missing)"
	@echo ""
	@echo "Build Artifacts:"
	@echo "CLI Binary:       $(if $(wildcard $(BUILD_DIR)/$(CLI_BIN)),✅ Built,❌ Not built)"
	@echo "Server Binary:    $(if $(wildcard $(BUILD_DIR)/$(SERVER_BIN)),✅ Built,❌ Not built)" 