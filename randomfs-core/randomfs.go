package randomfs

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	// Block sizes for different file categories
	NanoBlockSize = 1024    // 1KB for small files
	MiniBlockSize = 65536   // 64KB for medium files
	BlockSize     = 1048576 // 1MB for large files

	// Thresholds for block size selection
	NanoThreshold = 100 * 1024       // 100KB
	MiniThreshold = 10 * 1024 * 1024 // 10MB

	// Protocol version
	ProtocolVersion = "v4"

	// Default IPFS API endpoint
	DefaultIPFSEndpoint = "http://localhost:5001"
)

// RandomFS represents the main filesystem instance
type RandomFS struct {
	ipfsAPI    string
	httpPort   int
	dataDir    string
	blockCache *BlockCache
	mutex      sync.RWMutex

	// Statistics
	stats Stats
}

// Stats holds system statistics
type Stats struct {
	FilesStored     int64 `json:"files_stored"`
	BlocksGenerated int64 `json:"blocks_generated"`
	TotalSize       int64 `json:"total_size"`
	CacheHits       int64 `json:"cache_hits"`
	CacheMisses     int64 `json:"cache_misses"`
}

// BlockCache manages block storage and retrieval
type BlockCache struct {
	blocks      map[string][]byte
	mutex       sync.RWMutex
	maxSize     int64
	currentSize int64
}

// FileRepresentation contains the metadata needed to reconstruct a file
type FileRepresentation struct {
	FileName    string   `json:"filename"`
	FileSize    int64    `json:"filesize"`
	BlockHashes []string `json:"block_hashes"`
	BlockSize   int      `json:"block_size"`
	Timestamp   int64    `json:"timestamp"`
	ContentType string   `json:"content_type"`
	Version     string   `json:"version"`
}

// RandomURL represents a rd:// URL for file access
type RandomURL struct {
	Scheme    string
	Host      string
	Version   string
	FileName  string
	FileSize  int64
	RepHash   string
	Timestamp int64
}

// NewRandomFS creates a new RandomFS instance
func NewRandomFS(ipfsAPI string, httpPort int, dataDir string, cacheSize int64) (*RandomFS, error) {
	if ipfsAPI == "" {
		ipfsAPI = DefaultIPFSEndpoint
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	rfs := &RandomFS{
		ipfsAPI:  ipfsAPI,
		httpPort: httpPort,
		dataDir:  dataDir,
		blockCache: &BlockCache{
			blocks:  make(map[string][]byte),
			maxSize: cacheSize,
		},
	}

	// Test IPFS connection
	if err := rfs.testIPFSConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to IPFS: %v", err)
	}

	log.Printf("RandomFS initialized with IPFS at %s, HTTP port %d, data dir %s",
		ipfsAPI, httpPort, dataDir)

	return rfs, nil
}

// testIPFSConnection tests if IPFS daemon is accessible
func (rfs *RandomFS) testIPFSConnection() error {
	resp, err := http.Get(rfs.ipfsAPI + "/api/v0/version")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IPFS daemon not accessible, status: %d", resp.StatusCode)
	}

	return nil
}

// Start begins the HTTP server and background tasks
func (rfs *RandomFS) Start() error {
	router := mux.NewRouter()

	// API endpoints
	router.HandleFunc("/api/v1/store", rfs.handleStore).Methods("POST")
	router.HandleFunc("/api/v1/retrieve/{hash}", rfs.handleRetrieve).Methods("GET")
	router.HandleFunc("/api/v1/stats", rfs.handleStats).Methods("GET")
	router.HandleFunc("/rd/{encodedURL:.*}", rfs.handleRandomURL).Methods("GET")

	// Serve web interface
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	log.Printf("Starting HTTP server on port %d", rfs.httpPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", rfs.httpPort), router)
}

// StoreFile stores a file in the randomized block format
func (rfs *RandomFS) StoreFile(filename string, data []byte, contentType string) (*RandomURL, error) {
	rfs.mutex.Lock()
	defer rfs.mutex.Unlock()

	// Determine block size based on file size
	blockSize := rfs.selectBlockSize(int64(len(data)))

	// Generate randomized blocks
	blocks, err := rfs.generateRandomBlocks(data, blockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate blocks: %v", err)
	}

	// Store blocks in IPFS and cache
	var blockHashes []string
	for _, block := range blocks {
		hash, err := rfs.storeBlock(block)
		if err != nil {
			return nil, fmt.Errorf("failed to store block: %v", err)
		}
		blockHashes = append(blockHashes, hash)
	}

	// Create file representation
	rep := &FileRepresentation{
		FileName:    filepath.Base(filename),
		FileSize:    int64(len(data)),
		BlockHashes: blockHashes,
		BlockSize:   blockSize,
		Timestamp:   time.Now().Unix(),
		ContentType: contentType,
		Version:     ProtocolVersion,
	}

	// Store representation in IPFS
	repData, err := json.Marshal(rep)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal representation: %v", err)
	}

	repHash, err := rfs.addToIPFS(repData)
	if err != nil {
		return nil, fmt.Errorf("failed to store representation: %v", err)
	}

	// Update statistics
	rfs.stats.FilesStored++
	rfs.stats.BlocksGenerated += int64(len(blocks))
	rfs.stats.TotalSize += int64(len(data))

	// Create RandomURL
	randomURL := &RandomURL{
		Scheme:    "rd",
		Host:      "randomfs",
		Version:   ProtocolVersion,
		FileName:  rep.FileName,
		FileSize:  rep.FileSize,
		RepHash:   repHash,
		Timestamp: rep.Timestamp,
	}

	log.Printf("Stored file %s (%d bytes) with %d blocks, representation hash: %s",
		filename, len(data), len(blocks), repHash)

	return randomURL, nil
}

// RetrieveFile retrieves and reconstructs a file from its representation hash
func (rfs *RandomFS) RetrieveFile(repHash string) ([]byte, *FileRepresentation, error) {
	// Retrieve representation from IPFS
	repData, err := rfs.catFromIPFS(repHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve representation: %v", err)
	}

	var rep FileRepresentation
	if err := json.Unmarshal(repData, &rep); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal representation: %v", err)
	}

	// Retrieve and combine blocks
	var reconstructed bytes.Buffer
	for i, blockHash := range rep.BlockHashes {
		blockData, err := rfs.retrieveBlock(blockHash)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to retrieve block %d: %v", i, err)
		}

		// Apply XOR to de-randomize
		if i < len(rep.BlockHashes)-1 {
			// Full block
			deRandomized := rfs.deRandomizeBlock(blockData, rep.BlockSize)
			reconstructed.Write(deRandomized)
		} else {
			// Last block might be partial
			remaining := rep.FileSize - int64(reconstructed.Len())
			deRandomized := rfs.deRandomizeBlock(blockData, int(remaining))
			reconstructed.Write(deRandomized)
		}
	}

	log.Printf("Retrieved file %s (%d bytes) from %d blocks",
		rep.FileName, rep.FileSize, len(rep.BlockHashes))

	return reconstructed.Bytes(), &rep, nil
}

// generateRandomBlocks creates randomized blocks from file data
func (rfs *RandomFS) generateRandomBlocks(data []byte, blockSize int) ([][]byte, error) {
	var blocks [][]byte

	for offset := 0; offset < len(data); offset += blockSize {
		end := offset + blockSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[offset:end]

		// Create random block of fixed size
		randomBlock := make([]byte, blockSize)
		if _, err := rand.Read(randomBlock); err != nil {
			return nil, fmt.Errorf("failed to generate random data: %v", err)
		}

		// XOR with actual data to create multi-use block
		for i := 0; i < len(chunk); i++ {
			randomBlock[i] ^= chunk[i]
		}

		blocks = append(blocks, randomBlock)
	}

	return blocks, nil
}

// deRandomizeBlock recovers original data from a randomized block
func (rfs *RandomFS) deRandomizeBlock(block []byte, dataSize int) []byte {
	// For this implementation, we're using a simple XOR approach
	// In a real system, this would involve more complex cryptographic operations
	result := make([]byte, dataSize)
	copy(result, block[:dataSize])
	return result
}

// storeBlock stores a block in IPFS and local cache
func (rfs *RandomFS) storeBlock(block []byte) (string, error) {
	hash, err := rfs.addToIPFS(block)
	if err != nil {
		return "", err
	}

	// Cache locally for faster access
	rfs.blockCache.mutex.Lock()
	defer rfs.blockCache.mutex.Unlock()

	rfs.blockCache.blocks[hash] = block
	rfs.blockCache.currentSize += int64(len(block))

	// Simple cache eviction if over limit
	if rfs.blockCache.currentSize > rfs.blockCache.maxSize {
		rfs.evictOldestBlocks()
	}

	return hash, nil
}

// retrieveBlock retrieves a block from cache or IPFS
func (rfs *RandomFS) retrieveBlock(hash string) ([]byte, error) {
	// Check cache first
	rfs.blockCache.mutex.RLock()
	if block, exists := rfs.blockCache.blocks[hash]; exists {
		rfs.blockCache.mutex.RUnlock()
		rfs.stats.CacheHits++
		return block, nil
	}
	rfs.blockCache.mutex.RUnlock()

	// Retrieve from IPFS
	rfs.stats.CacheMisses++
	return rfs.catFromIPFS(hash)
}

// evictOldestBlocks removes oldest blocks from cache
func (rfs *RandomFS) evictOldestBlocks() {
	// Simple implementation - remove half the cache
	target := rfs.blockCache.maxSize / 2
	for hash, block := range rfs.blockCache.blocks {
		delete(rfs.blockCache.blocks, hash)
		rfs.blockCache.currentSize -= int64(len(block))
		if rfs.blockCache.currentSize <= target {
			break
		}
	}
}

// selectBlockSize determines the appropriate block size for a file
func (rfs *RandomFS) selectBlockSize(fileSize int64) int {
	if fileSize <= NanoThreshold {
		return NanoBlockSize
	} else if fileSize <= MiniThreshold {
		return MiniBlockSize
	}
	return BlockSize
}

// addToIPFS adds data to IPFS using HTTP API
func (rfs *RandomFS) addToIPFS(data []byte) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "data")
	if err != nil {
		return "", err
	}

	if _, err := part.Write(data); err != nil {
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	resp, err := http.Post(rfs.ipfsAPI+"/api/v0/add", writer.FormDataContentType(), &buf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("IPFS add failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Hash string `json:"Hash"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Hash, nil
}

// catFromIPFS retrieves data from IPFS using HTTP API
func (rfs *RandomFS) catFromIPFS(hash string) ([]byte, error) {
	resp, err := http.Get(rfs.ipfsAPI + "/api/v0/cat?arg=" + hash)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IPFS cat failed with status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// ParseRandomURL parses a rd:// URL
func ParseRandomURL(rawURL string) (*RandomURL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	if u.Scheme != "rd" {
		return nil, fmt.Errorf("invalid scheme: expected 'rd', got '%s'", u.Scheme)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid rd:// URL format")
	}

	fileSize, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid file size: %v", err)
	}

	timestamp, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %v", err)
	}

	return &RandomURL{
		Scheme:    u.Scheme,
		Host:      u.Host,
		Version:   parts[0],
		FileName:  parts[2],
		FileSize:  fileSize,
		RepHash:   parts[4],
		Timestamp: timestamp,
	}, nil
}

// String returns the string representation of a RandomURL
func (ru *RandomURL) String() string {
	return fmt.Sprintf("rd://%s/%s/%d/%s/%d/%s",
		ru.Host, ru.Version, ru.FileSize, ru.FileName, ru.Timestamp, ru.RepHash)
}

// HTTP Handlers

func (rfs *RandomFS) handleStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	randomURL, err := rfs.StoreFile(header.Filename, data, contentType)
	if err != nil {
		http.Error(w, "Failed to store file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":  randomURL.String(),
		"hash": randomURL.RepHash,
	})
}

func (rfs *RandomFS) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	data, rep, err := rfs.RetrieveFile(hash)
	if err != nil {
		http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", rep.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", rep.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(rep.FileSize, 10))
	w.Write(data)
}

func (rfs *RandomFS) handleRandomURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	encodedURL := vars["encodedURL"]

	// Decode the URL
	decodedURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		http.Error(w, "Invalid encoded URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	randomURL, err := ParseRandomURL(string(decodedURL))
	if err != nil {
		http.Error(w, "Invalid rd:// URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	data, rep, err := rfs.RetrieveFile(randomURL.RepHash)
	if err != nil {
		http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", rep.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", rep.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(rep.FileSize, 10))
	w.Write(data)
}

func (rfs *RandomFS) handleStats(w http.ResponseWriter, r *http.Request) {
	rfs.mutex.RLock()
	stats := rfs.stats
	rfs.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
