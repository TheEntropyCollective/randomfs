# RandomFS - Owner Free File System

## Overview

RandomFS is a distributed file system that provides privacy-preserving storage through cryptographic techniques. The system uses XOR-based storage with randomizer blocks to achieve plausible deniability, ensuring that stored data cannot be linked to its original content.

## Project Structure

### 🚀 **Production Components**
- **`randomfs-core/`** - Core library for programmatic access
- **`randomfs-cli/`** - Command-line interface with Cobra
- **`randomfs-http/`** - HTTP server with REST API
- **`randomfs-fuse/`** - FUSE filesystem integration
- **`randomfs-web/`** - Web interface and monitoring dashboard

### 🔬 **Research & Development**
- **`research/models/`** - Theoretical model implementations
  - `original-offsystem/` - Basic OFFSystem implementation
  - `connector-privacy/` - Connector Model with Differential Privacy
  - `anonymous-library/` - Optimal Anonymous Media Library
- **`research/simulations/`** - Cross-model efficiency comparisons
- **`research/papers/`** - Technical documentation and specifications

### 🛠️ **Prototypes**
- **`prototypes/`** - Experimental integrations and enhancements
  - `connector-enhanced/` - RandomFS + Connector Model integration

## Key Features

### Privacy Protection
- **Plausible Deniability**: XOR-based storage creates mathematical uncertainty
- **User Anonymity**: Network-level protection options (Tor/VPN)
- **Content Encryption**: All data encrypted before storage

### Storage Efficiency
- **Block-Level Deduplication**: Shared blocks across files
- **Network Densification**: Popular blocks become more valuable
- **Predictable Performance**: Consistent efficiency characteristics

### Performance
- **Streaming Support**: Progressive download with seeking
- **Connection Modes**: Privacy vs performance trade-offs
- **Intelligent Caching**: Popular content cached automatically

## Research Models

### Original OFFSystem
- **Privacy**: ⭐⭐⭐⭐⭐ (Maximum)
- **Efficiency**: ⭐⭐ (Unpredictable)
- **Use Case**: Maximum privacy scenarios

### Connector Model with Differential Privacy
- **Privacy**: ⭐⭐⭐⭐ (Balanced)
- **Efficiency**: ⭐⭐⭐⭐ (Predictable)
- **Use Case**: General purpose with configurable privacy

### Anonymous Media Library
- **Privacy**: ⭐⭐ (Accepts confirmation attacks)
- **Efficiency**: ⭐⭐⭐⭐⭐ (Perfect deduplication)
- **Use Case**: Media distribution, streaming services

## Quick Start

### Building the Project
```bash
# Build all components
make build

# Build specific component
make build-cli
make build-server

# Run tests
make test
```

### Using the CLI
```bash
# Store a file
./build/randomfs-cli store example.txt

# Retrieve a file
./build/randomfs-cli retrieve <file-id>

# List stored files
./build/randomfs-cli list
```

### Running the Server
```bash
# Start HTTP server
./build/randomfs-server -port 8080

# Access web interface
open http://localhost:8080
```

## Research & Development

### Running Model Comparisons
```bash
# Compare all research models
cd research/simulations/efficiency-comparison
go run main.go

# Test specific model
cd research/models/anonymous-library
go run main.go
```

### Understanding the Models
1. **Start with Original OFFSystem** (`research/models/original-offsystem/`) for basic concepts
2. **Explore Connector Model** (`research/models/connector-privacy/`) for balanced approach
3. **Study Anonymous Library** (`research/models/anonymous-library/`) for optimal solution

### Integration Paths
- **Phase 1**: Research validation and model comparison
- **Phase 2**: Prototype development with RandomFS integration
- **Phase 3**: Production deployment with backward compatibility

## Architecture

### Core Components
```
RandomFS Core
├── Block Storage      # XOR-based storage with randomizers
├── Block Selection    # Multi-factor scoring algorithm
├── Network Layer      # IPFS integration
└── Crypto Layer       # Encryption and key management
```

### Research Extensions
```
Research Models
├── Connector Model    # Enhanced efficiency with plausible deniability
├── Manifest Registry  # Popular content deduplication
├── Streaming Support  # Progressive download capabilities
└── Connection Modes   # Privacy vs performance options
```

## Performance Characteristics

### Storage Efficiency
- **Small Networks (1K users)**: 20-30% storage savings
- **Medium Networks (10K users)**: 30-60% storage savings
- **Large Networks (100K+ users)**: 60-90% storage savings
- **Popular Content**: Approaches 95% efficiency

### Retrieval Performance
- **Cached Blocks**: <100ms latency
- **Network Blocks**: 100-500ms depending on connection mode
- **Streaming**: Progressive download with 2-block buffer

## Security Model

### Threat Model
- **Confirmation Attacks**: Prevented by plausible deniability
- **Traffic Analysis**: Mitigated by connection mode choice
- **Content Inference**: Prevented by encryption

### Protection Mechanisms
- **Plausible Deniability**: 10,000+ possible explanations for any block
- **User Anonymity**: Network-level protection (Tor/VPN)
- **Content Encryption**: All data encrypted before storage

## Use Cases

### Privacy-Focused Storage
- **Sensitive Documents**: Legal, medical, financial data
- **Whistleblower Platforms**: Anonymous document sharing
- **Journalistic Sources**: Secure source communication

### Content Distribution
- **Media Streaming**: Netflix-like experience with privacy
- **Software Distribution**: Efficient package distribution
- **Academic Sharing**: Research paper distribution

### Enterprise Applications
- **Secure File Sharing**: Internal document distribution
- **Backup Systems**: Encrypted backup storage
- **Compliance**: Regulatory compliance with privacy requirements

## Development

### Prerequisites
- Go 1.21 or later
- Git
- Make (for build system)

### Development Setup
```bash
# Clone the repository
git clone https://github.com/TheEntropyCollective/randomfs.git
cd randomfs

# Initialize submodules
git submodule update --init --recursive

# Install dependencies
make deps

# Build all components
make build

# Run tests
make test
```

### Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

### Code Organization
- **Production Code**: `randomfs-*/` directories
- **Research Code**: `research/` directory
- **Prototypes**: `prototypes/` directory
- **Documentation**: README files in each directory

## Roadmap

### Short Term (3-6 months)
- [ ] Integrate Connector Model into RandomFS core
- [ ] Add streaming support for large files
- [ ] Implement manifest registry system
- [ ] Create connection mode selection

### Medium Term (6-12 months)
- [ ] Performance optimization and benchmarking
- [ ] Security auditing and penetration testing
- [ ] Large-scale deployment testing
- [ ] Community adoption and feedback

### Long Term (12+ months)
- [ ] Advanced features (homomorphic operations, zero-knowledge proofs)
- [ ] Quantum-resistant encryption
- [ ] Cross-platform mobile support
- [ ] Enterprise features and integrations

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **OFFSystem**: Original concept and inspiration
- **IPFS**: Distributed storage infrastructure
- **Research Community**: Academic contributions and feedback

## Support

- **Documentation**: See individual component README files
- **Issues**: Report bugs and feature requests on GitHub
- **Discussions**: Join community discussions on GitHub
- **Research**: Explore `research/` directory for advanced concepts

---

**RandomFS**: Privacy-preserving distributed storage for the modern web. 🔒🚀 