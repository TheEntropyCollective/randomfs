#!/bin/bash

# RandomFS Development Environment Manager
# Usage: ./randomfs-dev.sh [start|stop|restart|build|status] [--no-ipfs]

set -e

# Parse command line arguments
NO_IPFS=false
COMMAND=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-ipfs)
            NO_IPFS=true
            shift
            ;;
        start|stop|restart|build|status)
            COMMAND="$1"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Configuration
RANDOMFS_ROOT="/Users/jconnuck/TheEntropyCollective/randomfs"
IPFS_API="http://localhost:5001"
HTTP_PORT=8081
WEB_PORT=3000
DATA_DIR="$RANDOMFS_ROOT/data"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if IPFS is running
check_ipfs() {
    if curl -s "$IPFS_API/api/v0/version" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Check if a port is in use
check_port() {
    local port=$1
    if lsof -i :$port > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Build all RandomFS components
build_all() {
    log_info "Building all RandomFS components..."
    
    cd "$RANDOMFS_ROOT"
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    # Build core library
    log_info "Building randomfs-core..."
    cd randomfs-core
    go build ./pkg/randomfs
    cd ..
    
    # Build CLI
    log_info "Building randomfs-cli..."
    cd randomfs-cli
    go build -o ../bin/randomfs-cli ./cmd/randomfs-cli
    cd ..
    
    # Build HTTP server
    log_info "Building randomfs-http..."
    cd randomfs-http
    go build -o ../bin/randomfs-http ./cmd/randomfs-http
    cd ..
    
    # Build web server
    log_info "Building randomfs-web..."
    cd randomfs-web
    go build -o ../bin/randomfs-web ./cmd/randomfs-web
    cd ..
    
    # Copy web assets
    log_info "Copying web assets..."
    mkdir -p bin/web
    cp -r randomfs-web/assets/web/* bin/web/
    
    log_success "All components built successfully!"
}

# Start IPFS daemon
start_ipfs() {
    if check_ipfs; then
        log_warning "IPFS is already running"
        return
    fi
    
    log_info "Starting IPFS daemon..."
    ipfs daemon > /dev/null 2>&1 &
    local ipfs_pid=$!
    
    # Wait for IPFS to be ready
    log_info "Waiting for IPFS to be ready..."
    for i in {1..30}; do
        if check_ipfs; then
            log_success "IPFS started successfully (PID: $ipfs_pid)"
            return
        fi
        sleep 1
    done
    
    log_error "IPFS failed to start within 30 seconds"
    exit 1
}

# Start HTTP server
start_http() {
    if check_port $HTTP_PORT; then
        log_warning "HTTP server is already running on port $HTTP_PORT"
        return
    fi
    
    log_info "Starting RandomFS HTTP server on port $HTTP_PORT..."
    cd "$RANDOMFS_ROOT"
    
    if [ "$NO_IPFS" = true ]; then
        log_info "Starting HTTP server without IPFS..."
        ./bin/randomfs-http -port $HTTP_PORT -data "$DATA_DIR/http" --no-ipfs > /dev/null 2>&1 &
    else
        ./bin/randomfs-http -port $HTTP_PORT -data "$DATA_DIR/http" -ipfs "$IPFS_API" > /dev/null 2>&1 &
    fi
    
    local http_pid=$!
    
    # Wait for HTTP server to be ready
    log_info "Waiting for HTTP server to be ready..."
    for i in {1..10}; do
        if check_port $HTTP_PORT; then
            log_success "HTTP server started successfully (PID: $http_pid)"
            return
        fi
        sleep 1
    done
    
    log_error "HTTP server failed to start within 10 seconds"
    exit 1
}

# Start web server
start_web() {
    if check_port $WEB_PORT; then
        log_warning "Web server is already running on port $WEB_PORT"
        return
    fi
    
    log_info "Starting RandomFS web server on port $WEB_PORT..."
    cd "$RANDOMFS_ROOT"
    
    # Set environment variables for the web server
    export RANDOMFS_PORT=$WEB_PORT
    export RANDOMFS_DATA_DIR="$DATA_DIR/web"
    
    if [ "$NO_IPFS" = true ]; then
        log_info "Starting web server without IPFS..."
        # Don't set IPFS API environment variable
        ./bin/randomfs-web --no-ipfs > /dev/null 2>&1 &
    else
        export RANDOMFS_IPFS_API="$IPFS_API"
        ./bin/randomfs-web > /dev/null 2>&1 &
    fi
    local web_pid=$!
    
    # Wait for web server to be ready
    log_info "Waiting for web server to be ready..."
    for i in {1..10}; do
        if check_port $WEB_PORT; then
            log_success "Web server started successfully (PID: $web_pid)"
            return
        fi
        sleep 1
    done
    
    log_error "Web server failed to start within 10 seconds"
    exit 1
}

# Start everything
start_all() {
    if [ "$NO_IPFS" = true ]; then
        log_info "Starting RandomFS development environment (IPFS disabled)"
    else
        log_info "Starting RandomFS development environment..."
    fi
    
    # Create data directories
    mkdir -p "$DATA_DIR/http" "$DATA_DIR/web"
    
    # Start IPFS (unless --no-ipfs flag is used)
    if [ "$NO_IPFS" = true ]; then
        log_info "Skipping IPFS startup (--no-ipfs flag used)"
    else
        start_ipfs
    fi
    
    # Start servers
    start_http
    start_web
    
    log_success "RandomFS development environment started!"
    echo
    log_info "Services running:"
    if [ "$NO_IPFS" = true ]; then
        echo "  - IPFS API: DISABLED (--no-ipfs flag used)"
    else
        echo "  - IPFS API: $IPFS_API"
    fi
    echo "  - HTTP Server: http://localhost:$HTTP_PORT"
    echo "  - Web Interface: http://localhost:$WEB_PORT"
    echo
    log_info "Use './scripts/randomfs-dev.sh stop' to shut everything down"
}

# Stop everything
stop_all() {
    log_info "Stopping RandomFS development environment..."
    
    # Stop RandomFS processes
    pkill -f "randomfs-http" 2>/dev/null || true
    pkill -f "randomfs-web" 2>/dev/null || true
    
    # Stop IPFS (unless --no-ipfs flag is used)
    if [ "$NO_IPFS" = true ]; then
        log_info "Skipping IPFS shutdown (--no-ipfs flag used)"
    else
        pkill -f "ipfs daemon" 2>/dev/null || true
    fi
    
    # Wait for processes to fully stop and ports to be released
    log_info "Waiting for processes to stop and ports to be released..."
    for i in {1..15}; do
        if ! pgrep -f "randomfs" > /dev/null 2>&1 && ! check_port $HTTP_PORT && ! check_port $WEB_PORT; then
            break
        fi
        if [ "$NO_IPFS" = false ] && check_ipfs; then
            continue
        fi
        sleep 1
    done
    
    # Force kill any remaining processes
    pkill -9 -f "randomfs" 2>/dev/null || true
    if [ "$NO_IPFS" = false ]; then
        pkill -9 -f "ipfs daemon" 2>/dev/null || true
    fi
    
    # Final check
    if pgrep -f "randomfs" > /dev/null 2>&1; then
        log_warning "Some RandomFS processes may still be running"
    fi
    
    if [ "$NO_IPFS" = false ] && check_ipfs; then
        log_warning "IPFS may still be running"
    fi
    
    log_success "RandomFS development environment stopped!"
}

# Show status
show_status() {
    echo "RandomFS Development Environment Status"
    echo "======================================"
    echo
    
    # Check IPFS
    if check_ipfs; then
        echo -e "IPFS API (port 5001): ${GREEN}RUNNING${NC}"
    else
        echo -e "IPFS API (port 5001): ${RED}STOPPED${NC}"
    fi
    
    # Check HTTP server
    if check_port $HTTP_PORT; then
        echo -e "HTTP Server (port $HTTP_PORT): ${GREEN}RUNNING${NC}"
    else
        echo -e "HTTP Server (port $HTTP_PORT): ${RED}STOPPED${NC}"
    fi
    
    # Check web server
    if check_port $WEB_PORT; then
        echo -e "Web Server (port $WEB_PORT): ${GREEN}RUNNING${NC}"
    else
        echo -e "Web Server (port $WEB_PORT): ${RED}STOPPED${NC}"
    fi
    
    # Check binaries
    echo
    echo "Binaries:"
    if [ -f "$RANDOMFS_ROOT/bin/randomfs-cli" ]; then
        echo -e "  randomfs-cli: ${GREEN}EXISTS${NC}"
    else
        echo -e "  randomfs-cli: ${RED}MISSING${NC}"
    fi
    
    if [ -f "$RANDOMFS_ROOT/bin/randomfs-http" ]; then
        echo -e "  randomfs-http: ${GREEN}EXISTS${NC}"
    else
        echo -e "  randomfs-http: ${RED}MISSING${NC}"
    fi
    
    if [ -f "$RANDOMFS_ROOT/bin/randomfs-web" ]; then
        echo -e "  randomfs-web: ${GREEN}EXISTS${NC}"
    else
        echo -e "  randomfs-web: ${RED}MISSING${NC}"
    fi
}

# Main script logic
case "$COMMAND" in
    "start")
        start_all
        ;;
    "stop")
        stop_all
        ;;
    "restart")
        stop_all
        log_info "Waiting additional time for system to stabilize..."
        sleep 3
        build_all
        start_all
        ;;
    "build")
        build_all
        ;;
    "status")
        show_status
        ;;
    "")
        echo "RandomFS Development Environment Manager"
        echo "========================================"
        echo
        echo "Usage: $0 [start|stop|restart|build|status] [--no-ipfs]"
        echo
        echo "Commands:"
        echo "  start   - Build all components and start the development environment"
        echo "  stop    - Stop all RandomFS services and IPFS"
        echo "  restart - Build all components and restart the entire development environment"
        echo "  build   - Build all RandomFS components"
        echo "  status  - Show status of all services"
        echo
        echo "Options:"
        echo "  --no-ipfs - Run without IPFS (useful for development/testing)"
        echo
        echo "Services:"
        echo "  - IPFS API: $IPFS_API (disabled with --no-ipfs)"
        echo "  - HTTP Server: http://localhost:$HTTP_PORT"
        echo "  - Web Interface: http://localhost:$WEB_PORT"
        echo
        exit 1
        ;;
esac 