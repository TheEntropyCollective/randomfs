package main

import (
	crypto_rand "crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// --- Core Concepts ---
// 1. EFFICIENCY BY ENGINEERED COLLISION: The goal is to engineer a collision with a popular
//    block ('Target') to maximize reuse, rather than hoping to find randomizers.
// 2. USER-SUPPLIED OPACITY: The system requires the user to provide a high-entropy
//    (encrypted or compressed) 'Original' block for its deniability to hold.
// 3. DIFFERENTIAL PRIVACY (DP): Block selection is not purely random or purely deterministic.
//    It is deliberately noisy, providing strong statistical privacy guarantees.
//    The amount of noise is controlled by a global epsilon (ε).

// Represents a block of data.
type Block [4096]byte

// Represents the global network of stored blocks and their popularity.
var networkStorage = make(map[[32]byte]Block)
var blockPopularity = make(map[[32]byte]int)
var top100Blocks = make([][32]byte, 0, 100)

// --- System Implementation ---

// ConnectorModelDP implements the storage logic.
type ConnectorModelDP struct {
	// Epsilon (ε) is the privacy budget. Lower ε = more noise, more privacy.
	// A value of ~1.0 is often considered a good balance.
	epsilon float64
}

func (s *ConnectorModelDP) Name() string {
	return "Connector Model with Differential Privacy"
}

// Store processes a high-entropy block from the user.
// It returns the number of *new* blocks created (always 1 in this model).
func (s *ConnectorModelDP) Store(highEntropyOriginal Block) (newBlocksStored int) {
	log.Printf("--- Storing block using %s (ε=%.1f) ---\n", s.Name(), s.epsilon)

	// Step 0: Prerequisite check.
	// In a real system, we'd check if the block has high entropy. We simulate that here.
	log.Println("Step 0: Verifying that user-supplied block has high entropy.")

	// Step 1: Select a Target and a Randomizer from the Top 100 pool using DP.
	log.Println("Step 1: Selecting Target and Randomizer from Top 100 using Differential Privacy.")
	if len(top100Blocks) < 2 {
		log.Println(" -> Top 100 pool is empty. Cannot proceed. (Run simulation longer to populate).")
		return 0
	}
	targetHash := s.selectBlockWithDP(highEntropyOriginal, "target")
	randomizerHash := s.selectBlockWithDP(highEntropyOriginal, "randomizer")
	log.Printf(" -> Selected Target hash: %x...\n", targetHash[:4])
	log.Printf(" -> Selected Randomizer hash: %x...\n", randomizerHash[:4])

	targetBlock := networkStorage[targetHash]
	randomizerBlock := networkStorage[randomizerHash]

	// Step 2: Create the new Connector block.
	// This block is always new and stored on the network.
	log.Println("Step 2: Creating the new 'Connector' block via XOR operation.")
	connectorBlock := xor(highEntropyOriginal, xor(targetBlock, randomizerBlock))
	newBlocksStored = 1 // This model always stores exactly one new block.

	// Step 3: Store the new Connector block and update popularity of reused blocks.
	log.Printf("Step 3: Storing 1 new block and updating popularity.\n")
	connectorHash := hash(connectorBlock)
	networkStorage[connectorHash] = connectorBlock
	blockPopularity[connectorHash]++ // The new block has a popularity of 1.
	blockPopularity[targetHash]++
	blockPopularity[randomizerHash]++

	log.Printf("Store complete. Amortized cost for this operation: %d block.\n", newBlocksStored)
	return newBlocksStored
}

// selectBlockWithDP simulates a differentially private selection mechanism.
func (s *ConnectorModelDP) selectBlockWithDP(original Block, salt string) [32]byte {
	// 1. Deterministic Base Choice: Create a starting point from the file content.
	h := sha256.Sum256(append(original[:], []byte(salt)...))
	baseChoice := int(h[0]) % len(top100Blocks) // A number from 0-99

	// 2. Add Noise: Add calibrated noise for privacy.
	// We use the Geometric distribution, a discrete version of the Laplace mechanism for DP.
	noise := geometric(math.Exp(-s.epsilon))
	if randInt(2) == 0 { // 50% chance to be positive or negative
		noise = -noise
	}

	// 3. Final Choice: Apply noise and wrap around the list.
	finalChoice := (baseChoice + noise) % len(top100Blocks)
	if finalChoice < 0 {
		finalChoice += len(top100Blocks)
	}

	return top100Blocks[finalChoice]
}

// --- Utility and Simulation Functions ---

// populateInitialNetwork creates a set of popular blocks to simulate a live network.
func populateInitialNetwork(numBlocks int) {
	log.Printf("Populating network with %d initial blocks to create a Top 100 pool...\n", numBlocks)
	for i := 0; i < numBlocks; i++ {
		var block Block
		_, _ = rand.Read(block[:])
		h := hash(block)
		networkStorage[h] = block
		// Give blocks varying popularity so a Top 100 can emerge.
		blockPopularity[h] = randInt(100)
	}
	updateTop100()
	log.Printf("Network populated. Top 100 pool has %d blocks.\n", len(top100Blocks))
}

// updateTop100 recalculates the most popular blocks.
func updateTop100() {
	type popularBlock struct {
		hash       [32]byte
		popularity int
	}
	var sortedBlocks []popularBlock
	for h, pop := range blockPopularity {
		sortedBlocks = append(sortedBlocks, popularBlock{h, pop})
	}
	sort.Slice(sortedBlocks, func(i, j int) bool {
		return sortedBlocks[i].popularity > sortedBlocks[j].popularity
	})

	top100Blocks = top100Blocks[:0] // Clear the slice
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

func hash(b Block) [32]byte {
	return sha256.Sum256(b[:])
}

// geometric simulates drawing from a geometric distribution for DP.
func geometric(p float64) int {
	return int(math.Log(rand.Float64()) / math.Log(1.0-p))
}

func randInt(max int) int {
	b := make([]byte, 1)
	_, err := crypto_rand.Read(b)
	if err != nil {
		panic(err)
	}
	return int(b[0]) % max
}

// --- Simulation ---

func main() {
	// Seed the math/rand package for non-deterministic results.
	rand.Seed(time.Now().UnixNano())

	// Pre-populate the network to simulate an active system with popular blocks.
	populateInitialNetwork(200)
	fmt.Println(strings.Repeat("-", 40))

	// Instantiate the system with a balanced epsilon.
	system := &ConnectorModelDP{epsilon: 1.0}
	fmt.Printf("Simulating storage system: %s\n\n", system.Name())

	// Create a sample HIGH ENTROPY file block to store.
	// This model REQUIRES high entropy input for its security to hold.
	var encryptedBlock Block
	_, _ = rand.Read(encryptedBlock[:])
	log.Println("Storing a HIGH-ENTROPY (user pre-encrypted) block.")
	system.Store(encryptedBlock)

	// Update the Top 100 list after the store operation and show the new state.
	updateTop100()
	fmt.Println(strings.Repeat("-", 40))
	log.Printf("Simulation finished. Total blocks on network: %d\n", len(networkStorage))
}
