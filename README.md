# RandomFS - Owner Free File System

A modern implementation of the Owner Free File System (OFFS) concept using IPFS as the backing store. Files are split into randomized blocks that appear as noise, providing deniability while maintaining the ability to reconstruct original files using rfs:// URLs.

## Project Structure

This repository contains four independent components that can be used separately or together:

### 🧠 [randomfs-core](randomfs-core/)
**Core Library** - Pure Go library for programmatic access to RandomFS functionality.

- Multi-tier block sizing (1KB, 64KB, 1MB)
- XOR-based block randomization
- IPFS HTTP API integration
- LRU caching system
- rfs:// URL scheme implementation

### 💻 [randomfs-cli](randomfs-cli/)
**Command Line Interface** - Full-featured CLI tool built with Cobra.

- File storage and retrieval
- rfs:// URL parsing and generation
- System statistics
- Verbose output and debugging
- Shell completion support

### 🌐 [randomfs-http](randomfs-http/)
**HTTP Server** - Production-ready REST API server with web interface.

- REST API for programmatic access
- Modern web interface
- CORS support
- Health monitoring
- Configurable deployment

### 🎨 [randomfs-web](randomfs-web/)
**Web Interface** - Standalone web application for browser-based file management.

- Drag-and-drop file upload
- Real-time statistics
- Responsive design
- Cross-browser compatibility
- No framework dependencies

## Quick Start

### Option 1: Use Individual Components

```bash
# Core library only
cd randomfs-core
go build

# CLI tool
cd randomfs-cli
go build
./randomfs-cli store example.txt

# HTTP server
cd randomfs-http
go build
./randomfs-http -port 8080

# HTTP server without IPFS (for testing)
./randomfs-http -port 8080 -no-ipfs

# Web interface (standalone)
cd randomfs-web
python3 -m http.server 8000
```

### Option 2: Use HTTP Server with Web Interface

```bash
# Start HTTP server with web interface
cd randomfs-http
go build
./randomfs-http -web ../randomfs-web

# Start HTTP server without IPFS (for testing)
./randomfs-http -web ../randomfs-web -no-ipfs

# Open browser to http://localhost:8080
```

### Option 3: Use CLI for File Operations

```bash
# Store a file
cd randomfs-cli
go build
./randomfs-cli store example.txt

# Download using rfs:// URL
./randomfs-cli download rfs://QmX...abc/text/plain/example.txt
```

## Features

### 🔐 Owner Free File System
- **Deniability**: Individual blocks appear as random data
- **Reconstruction**: Original files can be perfectly reconstructed
- **Decentralized**: Uses IPFS for distributed storage
- **Privacy**: No metadata linking blocks to original files

### 📊 Multi-tier Block Sizing
- **Small files (< 1MB)**: 1KB blocks for efficiency
- **Medium files (1MB - 64MB)**: 64KB blocks for balance
- **Large files (> 64MB)**: 1MB blocks for performance

### 🔗 rfs:// URL Scheme
Files are accessed using the custom rfs:// URL format:
```
rfs://<representation-hash>/<content-type>/<original-filename>
```

### 🚀 Performance Optimizations
- **LRU Caching**: Configurable block caching
- **HTTP API**: Direct IPFS integration without complex dependencies
- **Efficient Storage**: Optimized block size selection
- **Parallel Processing**: Concurrent block operations

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   randomfs-web  │    │  randomfs-http  │    │  randomfs-cli   │
│   (Frontend)    │    │   (API Server)  │    │  (CLI Tool)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  randomfs-core  │
                    │  (Core Library) │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │      IPFS       │
                    │  (Backend Store)│
                    └─────────────────┘
```

## Dependencies

- **Go 1.21+** - For all Go components
- **IPFS Node (Kubo)** - With HTTP API enabled
- **Modern Web Browser** - For web interface

## Development

Each component is designed to be developed independently:

```bash
# Core library development
cd randomfs-core
go test -v
go build

# CLI development
cd randomfs-cli
go test -v
go build

# HTTP server development
cd randomfs-http
go test -v
go build

# Web interface development
cd randomfs-web
# Edit HTML/CSS/JS files
python3 -m http.server 8000
```

## Deployment

### Production Deployment
Each component can be deployed independently:

- **Core Library**: Import as Go module
- **CLI Tool**: Install binary on target systems
- **HTTP Server**: Deploy as systemd service with nginx reverse proxy
- **Web Interface**: Deploy to static hosting (GitHub Pages, Netlify, etc.)

### Docker Deployment
Each component includes Docker support for containerized deployment.

## Contributing

1. Choose the component you want to contribute to
2. Fork the specific repository
3. Create a feature branch
4. Make your changes
5. Add tests
6. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Related Projects

- **IPFS** - InterPlanetary File System
- **js-offs** - Original Owner Free File System concept
- **Kubo** - IPFS implementation

## Community

- **GitHub**: [TheEntropyCollective](https://github.com/TheEntropyCollective)
- **Issues**: Report bugs and request features
- **Discussions**: Join community discussions
- **Wiki**: Documentation and guides

---

**RandomFS** - Making file storage truly decentralized and owner-free. 🌌 