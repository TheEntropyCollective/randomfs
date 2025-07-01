.PHONY: build run test clean install deps check-ipfs help

# Variables
BINARY_NAME := randomfs
BUILD_DIR := ./build
MAIN_PATH := ./cmd/randomfs
DATA_DIR := ./data
WEB_DIR := ./web

# Default target
.DEFAULT_GOAL := help

# Build the application
build: ## Build the RandomFS binary
	@echo "Building RandomFS..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run: build check-ipfs ## Build and run RandomFS
	@echo "Starting RandomFS..."
	@mkdir -p $(DATA_DIR)
	@$(BUILD_DIR)/$(BINARY_NAME) -data $(DATA_DIR)

# Run with custom parameters
run-dev: build check-ipfs ## Run RandomFS in development mode
	@echo "Starting RandomFS in development mode..."
	@mkdir -p $(DATA_DIR)
	@$(BUILD_DIR)/$(BINARY_NAME) -port 8080 -data $(DATA_DIR) -cache 268435456

# Install dependencies
deps: ## Download and install Go dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run tests
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint: ## Run golangci-lint
	@echo "Linting code..."
	@golangci-lint run ./...

# Security scan
security: ## Run security scan with gosec
	@echo "Running security scan..."
	@gosec ./...

# Clean build artifacts
clean: ## Clean build artifacts and data
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DATA_DIR)
	@rm -f coverage.out coverage.html

# Install the binary to GOPATH/bin
install: build ## Install RandomFS to GOPATH/bin
	@echo "Installing RandomFS..."
	@go install $(MAIN_PATH)

# Check if IPFS is running
check-ipfs: ## Check if IPFS daemon is running
	@echo "Checking IPFS daemon..."
	@curl -s http://localhost:5001/api/v0/id > /dev/null || \
		(echo "❌ IPFS daemon not running! Please start it with 'ipfs daemon'" && exit 1)
	@echo "✅ IPFS daemon is running"

# Start IPFS daemon (if not running)
start-ipfs: ## Start IPFS daemon
	@echo "Starting IPFS daemon..."
	@ipfs daemon &

# Initialize IPFS (first time setup)
init-ipfs: ## Initialize IPFS node
	@echo "Initializing IPFS..."
	@ipfs init

# Docker build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t randomfs .

# Docker run
docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 -v $(PWD)/data:/app/data randomfs

# Development setup
setup-dev: deps init-ipfs ## Setup development environment
	@echo "Setting up development environment..."
	@mkdir -p $(DATA_DIR)
	@mkdir -p $(WEB_DIR)
	@echo "✅ Development environment ready!"
	@echo "Next steps:"
	@echo "  1. Start IPFS: make start-ipfs"
	@echo "  2. Run RandomFS: make run-dev"

# Generate self-signed certificates for HTTPS
gen-certs: ## Generate self-signed certificates
	@echo "Generating self-signed certificates..."
	@mkdir -p certs
	@openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem \
		-days 365 -nodes -subj "/C=US/ST=State/L=City/O=RandomFS/CN=localhost"
	@echo "Certificates generated in ./certs/"

# Create release archive
release: build ## Create release archive
	@echo "Creating release archive..."
	@mkdir -p release
	@tar -czf release/randomfs-$(shell uname -s)-$(shell uname -m).tar.gz \
		-C $(BUILD_DIR) $(BINARY_NAME)
	@echo "Release archive created: release/randomfs-$(shell uname -s)-$(shell uname -m).tar.gz"

# Performance benchmark
bench: ## Run performance benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Profile CPU usage
profile-cpu: build ## Profile CPU usage
	@echo "Profiling CPU usage..."
	@$(BUILD_DIR)/$(BINARY_NAME) -cpuprofile=cpu.prof &
	@sleep 30
	@pkill $(BINARY_NAME)
	@go tool pprof cpu.prof

# Profile memory usage
profile-mem: build ## Profile memory usage
	@echo "Profiling memory usage..."
	@$(BUILD_DIR)/$(BINARY_NAME) -memprofile=mem.prof &
	@sleep 30
	@pkill $(BINARY_NAME)
	@go tool pprof mem.prof

# Show help
help: ## Show this help message
	@echo "RandomFS - Modern Owner Free File System"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make setup-dev    # Setup development environment"
	@echo "  make run-dev      # Run in development mode"
	@echo "  make test         # Run all tests"
	@echo "  make clean        # Clean build artifacts" 