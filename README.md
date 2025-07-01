# RandomFS - Modern Owner Free File System

RandomFS is a modern Go implementation of the Owner Free File System (OFF System) concept, inspired by [js-offs](https://github.com/Prometheus-SCN/js-offs) but built with IPFS as the backing storage layer and enhanced with modern improvements.

## Overview

RandomFS implements the innovative concept of "owner-free" file storage where files are split into randomized, multi-use data blocks. Each block appears as digital noise individually but can be mathematically combined to reconstruct the original files. This approach provides:

- **Privacy**: Individual blocks contain no identifiable file data
- **Redundancy**: Blocks are shared across multiple files
- **Permanence**: Files persist as long as the network exists
- **Censorship Resistance**: No central authority controls the data

## Key Features

### 🔄 **Randomized Block Storage**
- Files are split into randomized blocks using XOR operations
- Each block can be part of multiple files simultaneously
- Three optimized block sizes (1KB, 64KB, 1MB) based on file size

### 🌐 **IPFS Integration**
- Uses [Kubo (go-ipfs)](https://github.com/ipfs/kubo) as the decentralized storage layer
- Automatic block distribution across the IPFS network
- Content-addressed storage ensures data integrity

### 🔗 **rd:// URL Scheme**
- Custom URL scheme for accessing stored files
- Format: `rd://host/version/filesize/filename/timestamp/hash`
- Enables direct browser access to files via HTTP gateway

### ⚡ **Performance Optimizations**
- Local block caching for faster retrieval
- Intelligent block size selection
- Concurrent operations for better throughput

### 🎯 **Modern Web Interface**
- Drag-and-drop file upload
- Real-time progress tracking
- System statistics dashboard
- Mobile-responsive design

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   RandomFS API  │    │   IPFS Network  │
│                 │    │                 │    │                 │
│ • File Upload   │◄──►│ • Block Gen.    │◄──►│ • Distributed   │
│ • rd:// URLs    │    │ • Caching       │    │   Storage       │
│ • Statistics    │    │ • HTTP Gateway  │    │ • Content Hash  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Installation

### Prerequisites

1. **Go 1.21+**
2. **IPFS Node** (Kubo)
   ```bash
   # Install IPFS
   wget https://dist.ipfs.tech/kubo/v0.24.0/kubo_v0.24.0_linux-amd64.tar.gz
   tar -xzf kubo_v0.24.0_linux-amd64.tar.gz
   sudo ./kubo/install.sh
   
   # Initialize and start IPFS
   ipfs init
   ipfs daemon
   ```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/TheEntropyCollective/randomfs-core.git
cd randomfs-core

# Build the application
go build -o randomfs ./cmd/randomfs

# Run RandomFS
./randomfs -port 8080 -ipfs localhost:5001
```

### Using Go Install

```bash
go install github.com/TheEntropyCollective/randomfs-core/cmd/randomfs@latest
randomfs -port 8080
```

## Usage

### Starting the Server

```bash
# Default configuration
./randomfs

# Custom configuration
./randomfs -port 8080 -ipfs localhost:5001 -data ./data -cache 1073741824
```

### Command Line Options

- `-port`: HTTP server port (default: 8080)
- `-ipfs`: IPFS API endpoint (default: localhost:5001)
- `-data`: Data directory (default: ./data)
- `-cache`: Cache size in bytes (default: 500MB)

### Web Interface

1. Open `http://localhost:8080` in your browser
2. Drag and drop files or click to browse
3. Get rd:// URLs for uploaded files
4. Share URLs for decentralized access

### API Endpoints

#### Store File
```bash
curl -X POST -F "file=@example.txt" http://localhost:8080/api/v1/store
```

Response:
```json
{
  "url": "rd://randomfs/v4/1024/example.txt/1640995200/QmHash123...",
  "hash": "QmHash123..."
}
```

#### Retrieve File
```bash
curl http://localhost:8080/api/v1/retrieve/QmHash123...
```

#### Access via rd:// URL
```bash
# Base64 encode the rd:// URL and access via HTTP
curl http://localhost:8080/rd/cmQ6Ly9yYW5kb21mcy92NC8xMDI0L2V4YW1wbGUudHh0LzE2NDA5OTUyMDAvUW1IYXNoMTIz
```

#### System Statistics
```bash
curl http://localhost:8080/api/v1/stats
```

## rd:// URL Format

The rd:// URL scheme provides a standardized way to reference files in the RandomFS network:

```
rd://host/version/filesize/filename/timestamp/representation_hash
```

- **host**: Network identifier (e.g., "randomfs")
- **version**: Protocol version (e.g., "v4")
- **filesize**: Original file size in bytes
- **filename**: Original filename
- **timestamp**: Unix timestamp of storage
- **representation_hash**: IPFS hash of file representation metadata

### Example
```
rd://randomfs/v4/2048/document.pdf/1640995200/QmYwAPJzv5CZsnA8VQF4EjA6JBGn2vqb9kj9xhQ3xNK8F1
```

## Block Storage Algorithm

RandomFS uses a sophisticated multi-use block generation algorithm:

1. **File Analysis**: Determine optimal block size based on file size
2. **Block Generation**: Split file into chunks and XOR with random data
3. **IPFS Storage**: Store randomized blocks in IPFS network
4. **Metadata Creation**: Generate file representation with block references
5. **URL Generation**: Create rd:// URL for file access

### Block Size Selection
- **Nano blocks (1KB)**: Files ≤ 100KB
- **Mini blocks (64KB)**: Files ≤ 10MB  
- **Standard blocks (1MB)**: Files > 10MB

## Security Considerations

- **Block Randomization**: Individual blocks appear as random noise
- **No Direct Mapping**: No one-to-one relationship between blocks and files
- **Distributed Storage**: Blocks spread across IPFS network
- **Content Addressing**: IPFS ensures data integrity via cryptographic hashing

## Performance

RandomFS is optimized for performance with:

- **Local Caching**: Frequently accessed blocks cached in memory
- **Concurrent Operations**: Parallel block processing
- **Smart Prefetching**: Predictive block loading
- **Connection Pooling**: Efficient IPFS API usage

## Comparison with Original js-offs

| Feature | js-offs | RandomFS |
|---------|---------|----------|
| Language | JavaScript/Node.js | Go |
| Storage Backend | Custom P2P | IPFS |
| Block Sizes | Fixed size | 3 adaptive sizes |
| Web Interface | Vue.js | Vanilla JS |
| Caching | Basic | Advanced LRU |
| URL Scheme | /offsystem/v3/... | rd://... |
| Performance | Good | Optimized |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original OFF System concept by [The Big Hack](http://offsystem.sourceforge.net/)
- [js-offs](https://github.com/Prometheus-SCN/js-offs) implementation by Prometheus SCN
- [IPFS](https://ipfs.tech/) for providing the distributed storage foundation
- [Kubo](https://github.com/ipfs/kubo) for the IPFS implementation

## Links

- [Original OFF System](http://offsystem.sourceforge.net/)
- [js-offs Implementation](https://github.com/Prometheus-SCN/js-offs)
- [IPFS Documentation](https://docs.ipfs.tech/)
- [OFF System Website](https://www.off.systems/)

---

**RandomFS**: Own nothing. Access everything. 🌐 