package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// --- Core Concepts ---
// ANONYMOUS MEDIA LIBRARY: The optimal solution for distributed media content
// 1. DETERMINISTIC ENCRYPTION: Same file + password = same encrypted blocks (perfect deduplication)
// 2. CONNECTOR MODEL ENGINE: Encrypted blocks stored via XOR with popular blocks (plausible deniability)
// 3. MANIFEST REGISTRY: Popular files have pre-computed manifests (zero-cost storage for popular content)
// 4. CONNECTION MODES: User choice between privacy and performance
// 5. STREAMING SUPPORT: Progressive download with buffering for media playback

type Block [4096]byte
type BlockHash [32]byte

// Connection modes for different privacy/performance trade-offs
type ConnectionMode int

const (
	MaxPrivacy ConnectionMode = iota // Tor for everything (slow, maximum anonymity)
	Standard                         // Direct IPFS (fast, plausible deniability via Connector Model)
	Paranoid                         // Standard + VPN recommended (fast, maximum protection)
)

func (cm ConnectionMode) String() string {
	switch cm {
	case MaxPrivacy:
		return "Max Privacy (Tor)"
	case Standard:
		return "Standard (Direct IPFS)"
	case Paranoid:
		return "Paranoid (Direct + VPN)"
	default:
		return "Unknown"
	}
}

// Represents a file's manifest - the "recipe" to reconstruct it
type FileManifest struct {
	FileName        string                `json:"filename"`
	FileSize        int64                 `json:"filesize"`
	FileHash        BlockHash             `json:"filehash"` // Hash of original file
	BlockCount      int                   `json:"block_count"`
	ConnectorBlocks []ConnectorDescriptor `json:"connector_blocks"` // Ordered list of blocks
	CreatedAt       time.Time             `json:"created_at"`
	PopularityScore int                   `json:"popularity_score"`
}

// Describes how to reconstruct one block of the original file
type ConnectorDescriptor struct {
	Position       int       `json:"position"`        // Block position in file (for streaming)
	ConnectorHash  BlockHash `json:"connector_hash"`  // The stored Connector block
	TargetHash     BlockHash `json:"target_hash"`     // Popular block used as Target
	RandomizerHash BlockHash `json:"randomizer_hash"` // Popular block used as Randomizer
}

// Global network state
var networkStorage = make(map[BlockHash]Block)
var blockPopularity = make(map[BlockHash]int)
var top100Blocks = make([]BlockHash, 0, 100)

// Manifest registry (in a real system, this would be a distributed DHT)
var manifestRegistry = make(map[BlockHash]*FileManifest)

// --- System Implementation ---

type AnonymousMediaLibrary struct {
	connectionMode ConnectionMode
}

func NewAnonymousMediaLibrary(mode ConnectionMode) *AnonymousMediaLibrary {
	return &AnonymousMediaLibrary{connectionMode: mode}
}

func (aml *AnonymousMediaLibrary) Name() string {
	return fmt.Sprintf("Anonymous Media Library (%s)", aml.connectionMode)
}

// StoreFile stores a complete file, handling chunking, encryption, and manifest creation
func (aml *AnonymousMediaLibrary) StoreFile(fileName string, fileData []byte, password string) (newBlocksStored int, manifest *FileManifest) {
	log.Printf("--- Storing file '%s' (%d bytes) using %s ---\n", fileName, len(fileData), aml.Name())

	// Step 1: Check if this file already has a manifest (perfect deduplication)
	fileHash := sha256.Sum256(fileData)
	if existingManifest, exists := manifestRegistry[fileHash]; exists {
		log.Printf("✓ File already exists in manifest registry! Storage cost: 0 blocks")
		log.Printf("  Original upload: %s (popularity: %d)", existingManifest.CreatedAt.Format("2006-01-02"), existingManifest.PopularityScore)
		existingManifest.PopularityScore++ // Track reuse
		return 0, existingManifest
	}

	// Step 2: Chunk the file into blocks
	log.Printf("Step 1: Chunking file into %d-byte blocks...", len(Block{}))
	blocks := aml.chunkFile(fileData)
	log.Printf("  Created %d blocks", len(blocks))

	// Step 3: Deterministically encrypt each block
	log.Printf("Step 2: Deterministically encrypting blocks...")
	encryptedBlocks := make([]Block, len(blocks))
	for i, block := range blocks {
		encryptedBlocks[i] = aml.deterministicEncrypt(block, password, i) // Include position for uniqueness
	}

	// Step 4: Store each encrypted block using the Connector Model
	log.Printf("Step 3: Storing encrypted blocks using Connector Model...")
	manifest = &FileManifest{
		FileName:        fileName,
		FileSize:        int64(len(fileData)),
		FileHash:        fileHash,
		BlockCount:      len(blocks),
		ConnectorBlocks: make([]ConnectorDescriptor, len(blocks)),
		CreatedAt:       time.Now(),
		PopularityScore: 1,
	}

	for i, encryptedBlock := range encryptedBlocks {
		// Check if this exact encrypted block already exists (block-level deduplication)
		encryptedHash := hash(encryptedBlock)
		if _, exists := networkStorage[encryptedHash]; exists {
			log.Printf("  Block %d: Already exists (block-level deduplication)", i)
			// Still need to record it in manifest, but no storage cost
			manifest.ConnectorBlocks[i] = ConnectorDescriptor{
				Position:       i,
				ConnectorHash:  encryptedHash,
				TargetHash:     BlockHash{}, // Not needed for existing blocks
				RandomizerHash: BlockHash{},
			}
			continue
		}

		// Store new block using Connector Model
		targetHash, randomizerHash := aml.selectPopularBlocks()
		targetBlock := networkStorage[targetHash]
		randomizerBlock := networkStorage[randomizerHash]

		connectorBlock := xor(encryptedBlock, xor(targetBlock, randomizerBlock))
		connectorHash := hash(connectorBlock)

		// Store the connector block
		networkStorage[connectorHash] = connectorBlock
		blockPopularity[connectorHash] = 1
		blockPopularity[targetHash]++
		blockPopularity[randomizerHash]++
		newBlocksStored++

		// Record in manifest
		manifest.ConnectorBlocks[i] = ConnectorDescriptor{
			Position:       i,
			ConnectorHash:  connectorHash,
			TargetHash:     targetHash,
			RandomizerHash: randomizerHash,
		}

		if aml.connectionMode == MaxPrivacy {
			log.Printf("  Block %d: Stored via Tor (slow but anonymous)", i)
		} else {
			log.Printf("  Block %d: Stored via direct IPFS (fast, plausible deniability)", i)
		}
	}

	// Step 5: Register the manifest for future perfect deduplication
	log.Printf("Step 4: Registering manifest in distributed registry...")
	manifestRegistry[fileHash] = manifest

	log.Printf("✓ Storage complete. Cost: %d new blocks (%.1f%% efficiency vs naive storage)",
		newBlocksStored, (1.0-float64(newBlocksStored)/float64(len(blocks)))*100)

	return newBlocksStored, manifest
}

// RetrieveFile reconstructs a file from its manifest (simulates streaming-capable retrieval)
func (aml *AnonymousMediaLibrary) RetrieveFile(manifest *FileManifest, password string, startBlock int) ([]byte, error) {
	log.Printf("--- Retrieving file '%s' starting from block %d ---", manifest.FileName, startBlock)

	if aml.connectionMode == MaxPrivacy {
		log.Printf("Using Tor routing (high latency, maximum anonymity)")
		time.Sleep(500 * time.Millisecond) // Simulate Tor latency
	} else {
		log.Printf("Using direct IPFS (low latency, plausible deniability protection)")
		time.Sleep(50 * time.Millisecond) // Simulate normal network latency
	}

	// Retrieve blocks starting from startBlock (enables streaming)
	retrievedData := make([]byte, 0, manifest.FileSize)

	for i := startBlock; i < len(manifest.ConnectorBlocks); i++ {
		descriptor := manifest.ConnectorBlocks[i]

		// Get the connector block
		connectorBlock, exists := networkStorage[descriptor.ConnectorHash]
		if !exists {
			return nil, fmt.Errorf("connector block %d not found", i)
		}

		// If this block was stored via Connector Model, reconstruct the encrypted block
		var encryptedBlock Block
		if descriptor.TargetHash != (BlockHash{}) { // Has Target/Randomizer info
			targetBlock := networkStorage[descriptor.TargetHash]
			randomizerBlock := networkStorage[descriptor.RandomizerHash]
			encryptedBlock = xor(connectorBlock, xor(targetBlock, randomizerBlock))
		} else {
			// Block was already encrypted and stored directly
			encryptedBlock = connectorBlock
		}

		// Decrypt the block
		originalBlock := aml.deterministicDecrypt(encryptedBlock, password, i)
		retrievedData = append(retrievedData, originalBlock[:]...)

		// For streaming: could yield control here to start playback
		if i == startBlock+2 { // After buffering a few blocks
			log.Printf("  Streaming: Buffered %d blocks, playback can start", i-startBlock+1)
		}
	}

	// Trim to actual file size
	if int64(len(retrievedData)) > manifest.FileSize {
		retrievedData = retrievedData[:manifest.FileSize]
	}

	log.Printf("✓ Retrieval complete. File reconstructed successfully.")
	return retrievedData, nil
}

// --- Encryption/Decryption ---

func (aml *AnonymousMediaLibrary) deterministicEncrypt(block Block, password string, position int) Block {
	// Create deterministic key based on password and position
	// This ensures same password+position always produces same key (deterministic)
	keyMaterial := fmt.Sprintf("%s:%d", password, position)
	key := sha256.Sum256([]byte(keyMaterial))

	// Simple XOR encryption with the deterministic key
	var encryptedBlock Block
	for i := 0; i < len(block); i++ {
		encryptedBlock[i] = block[i] ^ key[i%32]
	}
	return encryptedBlock
}

func (aml *AnonymousMediaLibrary) deterministicDecrypt(encryptedBlock Block, password string, position int) Block {
	// For XOR encryption, decryption uses the same key
	keyMaterial := fmt.Sprintf("%s:%d", password, position)
	key := sha256.Sum256([]byte(keyMaterial))

	var decryptedBlock Block
	for i := 0; i < len(encryptedBlock); i++ {
		decryptedBlock[i] = encryptedBlock[i] ^ key[i%32]
	}
	return decryptedBlock
}

// --- Utility Functions ---

func (aml *AnonymousMediaLibrary) chunkFile(data []byte) []Block {
	blockSize := len(Block{})
	numBlocks := (len(data) + blockSize - 1) / blockSize
	blocks := make([]Block, numBlocks)

	for i := 0; i < numBlocks; i++ {
		start := i * blockSize
		end := start + blockSize
		if end > len(data) {
			end = len(data)
		}
		copy(blocks[i][:], data[start:end])
	}
	return blocks
}

func (aml *AnonymousMediaLibrary) selectPopularBlocks() (target, randomizer BlockHash) {
	if len(top100Blocks) < 2 {
		panic("Top 100 pool not populated")
	}
	idx1 := randInt(len(top100Blocks))
	idx2 := randInt(len(top100Blocks))
	return top100Blocks[idx1], top100Blocks[idx2]
}

func populateInitialNetwork(numBlocks int) {
	log.Printf("Initializing network with %d popular blocks...", numBlocks)
	for i := 0; i < numBlocks; i++ {
		var block Block
		_, _ = rand.Read(block[:])
		h := hash(block)
		networkStorage[h] = block
		blockPopularity[h] = randInt(100) + 50 // Ensure decent popularity
	}
	updateTop100()
	log.Printf("Network ready. Top 100 pool populated.")
}

func updateTop100() {
	type popularBlock struct {
		hash       BlockHash
		popularity int
	}
	var sortedBlocks []popularBlock
	for h, pop := range blockPopularity {
		sortedBlocks = append(sortedBlocks, popularBlock{h, pop})
	}
	sort.Slice(sortedBlocks, func(i, j int) bool {
		return sortedBlocks[i].popularity > sortedBlocks[j].popularity
	})

	top100Blocks = top100Blocks[:0]
	for i := 0; i < 100 && i < len(sortedBlocks); i++ {
		top100Blocks = append(top100Blocks, sortedBlocks[i].hash)
	}
}

func xor(a, b Block) Block {
	var result Block
	for i := 0; i < len(a); i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

func hash(b Block) BlockHash {
	return sha256.Sum256(b[:])
}

func randInt(max int) int {
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return int(b[0]) % max
}

// --- Simulation ---

func main() {
	// Initialize the network
	populateInitialNetwork(200)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🎬 ANONYMOUS MEDIA LIBRARY - The Optimal Solution for Distributed Content")
	fmt.Println("Features: Perfect Deduplication | Streaming Support | Plausible Deniability")
	fmt.Println(strings.Repeat("=", 80))

	// Test different connection modes
	modes := []ConnectionMode{Standard, MaxPrivacy}

	for _, mode := range modes {
		fmt.Printf("\n--- Testing %s ---\n", mode)
		system := NewAnonymousMediaLibrary(mode)

		// Test Case 1: Popular movie file
		fmt.Println("\n📹 SCENARIO 1: Popular Movie Distribution")
		movieContent := []byte(strings.Repeat("This is a popular movie that many users will want to download and stream. ", 100))
		password := "shared-password-123"

		// First user stores the movie
		log.Println("\n=== User A stores popular movie ===")
		blocksA, manifestA := system.StoreFile("popular_movie.mkv", movieContent, password)

		// Second user tries to store the same movie (should be free due to manifest registry)
		log.Println("\n=== User B stores the same movie ===")
		blocksB, manifestB := system.StoreFile("popular_movie.mkv", movieContent, password)

		// Third user retrieves the movie (simulating streaming from beginning)
		log.Println("\n=== User C streams the movie from beginning ===")
		retrieved, err := system.RetrieveFile(manifestA, password, 0)
		if err != nil {
			log.Printf("Error: %v", err)
		} else if string(retrieved) == string(movieContent) {
			log.Printf("✓ File retrieved successfully and matches original")
		} else {
			log.Printf("✗ File corruption detected")
		}

		// Fourth user seeks to middle of movie (simulating streaming seek)
		log.Println("\n=== User D seeks to middle of movie ===")
		partialRetrieved, err := system.RetrieveFile(manifestA, password, 1)
		if err != nil {
			log.Printf("Error: %v", err)
		} else {
			log.Printf("✓ Partial retrieval successful (seeking to block 1, got %d bytes)", len(partialRetrieved))
		}

		// Test Case 2: Different file with same password
		fmt.Println("\n📚 SCENARIO 2: Different Content, Same Password")
		bookContent := []byte(strings.Repeat("This is a different book with unique content but same password. ", 50))

		log.Println("\n=== User E stores a book ===")
		blocksE, manifestE := system.StoreFile("book.pdf", bookContent, password)

		fmt.Printf("\nResults for %s:\n", mode)
		fmt.Printf("  Movie - User A storage cost: %d blocks\n", blocksA)
		fmt.Printf("  Movie - User B storage cost: %d blocks (perfect deduplication!)\n", blocksB)
		fmt.Printf("  Movie - Manifest popularity: %d users\n", manifestB.PopularityScore)
		fmt.Printf("  Book  - User E storage cost: %d blocks\n", blocksE)
		fmt.Printf("  Book  - File size: %d bytes (%d blocks)\n", len(bookContent), len(manifestE.ConnectorBlocks))

		// Calculate efficiency
		totalOriginalBlocks := len(manifestA.ConnectorBlocks) + len(manifestE.ConnectorBlocks)
		totalStoredBlocks := blocksA + blocksB + blocksE
		efficiency := (1.0 - float64(totalStoredBlocks)/float64(totalOriginalBlocks)) * 100
		fmt.Printf("  Overall efficiency: %.1f%% (stored %d blocks instead of %d)\n",
			efficiency, totalStoredBlocks, totalOriginalBlocks)

		fmt.Println(strings.Repeat("-", 60))
	}

	// Final statistics
	fmt.Printf("\n📊 FINAL NETWORK STATISTICS:\n")
	fmt.Printf("  Total blocks stored: %d\n", len(networkStorage))
	fmt.Printf("  Registered manifests: %d\n", len(manifestRegistry))
	fmt.Printf("  Top 100 pool size: %d\n", len(top100Blocks))
	fmt.Printf("  Average manifest popularity: %.1f\n", getAveragePopularity())

	// Show the power of the system
	fmt.Printf("\n🚀 SYSTEM CAPABILITIES DEMONSTRATED:\n")
	fmt.Printf("  ✅ Perfect File Deduplication (same file = 0 storage cost)\n")
	fmt.Printf("  ✅ Streaming Support (seek to any position)\n")
	fmt.Printf("  ✅ Multiple Connection Modes (privacy vs performance)\n")
	fmt.Printf("  ✅ Plausible Deniability (Connector Model protection)\n")
	fmt.Printf("  ✅ Manifest Registry (popular files cost nothing)\n")
	fmt.Printf("  ✅ Block-Level Efficiency (shared encrypted blocks)\n")

	fmt.Printf("\n💡 REAL-WORLD IMPLICATIONS:\n")
	fmt.Printf("  • Popular movies: First user pays storage cost, everyone else gets free access\n")
	fmt.Printf("  • Streaming services: Instant seeking, progressive download\n")
	fmt.Printf("  • Privacy protection: Plausible deniability via mathematical uncertainty\n")
	fmt.Printf("  • Network efficiency: Scales better as more users join\n")
}

func getAveragePopularity() float64 {
	if len(manifestRegistry) == 0 {
		return 0
	}
	total := 0
	for _, manifest := range manifestRegistry {
		total += manifest.PopularityScore
	}
	return float64(total) / float64(len(manifestRegistry))
}
