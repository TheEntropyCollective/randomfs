package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"strings"
)

// --- Core Concepts ---
// 1. ANONYMITY BY OBSCURITY: The primary goal is to make any stored file look
//    like high-entropy, random noise. It does not require the user to pre-encrypt.
// 2. NO GUARANTEED REUSE: The system *tries* to find existing randomizer blocks
//    on the network to reuse, but there is no guarantee. Efficiency is secondary.
// 3. DEAD-END BLOCKS: The final 'Anonymized' block is unique and unlikely to be
//    reused, diluting the network over time.

// Represents a block of data, which could be an Original file block or a Randomizer.
type Block [4096]byte

// Represents the global network of stored blocks.
// In a real system, this would be a distributed network like IPFS.
var networkStorage = make(map[[32]byte]Block)

// --- System Implementation ---

// OriginalOFFSystem implements the storage logic.
type OriginalOFFSystem struct{}

func (s *OriginalOFFSystem) Name() string {
	return "Original OFFSystem (Anonymity by Obscurity)"
}

// Store processes and "stores" a block of a file.
// It returns the number of *new* blocks that had to be created and stored on the network.
func (s *OriginalOFFSystem) Store(originalBlock Block) (newBlocksStored int) {
	log.Printf("--- Storing block using %s ---\n", s.Name())

	// Step 1: Find two randomizer blocks on the network.
	// This model prioritizes high-entropy, available blocks, not necessarily the
	// most popular ones. For this simulation, we'll assume it sometimes fails to find
	// existing blocks, forcing it to create new ones.
	log.Println("Step 1: Searching for two randomizer blocks...")
	randomizerA, foundA := s.findRandomizerBlock()
	randomizerB, foundB := s.findRandomizerBlock()

	if !foundA {
		log.Println(" -> Randomizer A not found. Creating a new one.")
		randomizerA = s.createNewBlock()
		newBlocksStored++
	} else {
		log.Println(" -> Found existing Randomizer A.")
	}

	if !foundB {
		log.Println(" -> Randomizer B not found. Creating a new one.")
		randomizerB = s.createNewBlock()
		newBlocksStored++
	} else {
		log.Println(" -> Found existing Randomizer B.")
	}

	// Step 2: Create the Anonymized block by XORing the original with the randomizers.
	// The resulting 'Anonymized' block is always new and unique.
	log.Println("Step 2: Creating the new 'Anonymized' block via XOR operation.")
	anonymizedBlock := xor(originalBlock, xor(randomizerA, randomizerB))
	newBlocksStored++

	// Step 3: Store the new blocks on the network.
	log.Printf("Step 3: Storing %d new block(s) on the network.\n", newBlocksStored)
	if !foundA {
		networkStorage[hash(randomizerA)] = randomizerA
	}
	if !foundB {
		networkStorage[hash(randomizerB)] = randomizerB
	}
	networkStorage[hash(anonymizedBlock)] = anonymizedBlock

	log.Printf("Store complete. Amortized cost for this operation: %d blocks.\n", newBlocksStored)
	return newBlocksStored
}

// findRandomizerBlock simulates searching the network for a suitable block.
// In this model, the search is unpredictable. We'll simulate a 50% success rate.
func (s *OriginalOFFSystem) findRandomizerBlock() (Block, bool) {
	// Simple probabilistic simulation: 50% chance of finding a reusable block.
	if len(networkStorage) > 0 && randInt(2) == 0 {
		// Pick a random block from the network to reuse.
		for _, block := range networkStorage {
			return block, true // In a real system, we'd loop, but for simulation, one is fine.
		}
	}
	return Block{}, false
}

// createNewBlock generates a new high-entropy block and adds it to the network.
func (s *OriginalOFFSystem) createNewBlock() Block {
	var newBlock Block
	_, err := rand.Read(newBlock[:])
	if err != nil {
		panic("failed to generate random data for block")
	}
	return newBlock
}

// --- Utility Functions ---

func xor(a, b Block) Block {
	var result Block
	for i := 0; i < len(a); i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

func hash(b Block) [32]byte {
	// In a real system, this would be SHA256. For simplicity, we just use the block itself
	// as a key, but we'll represent it as a hash.
	var h [32]byte
	copy(h[:], b[:32]) // Not a real hash, but sufficient for map keys.
	return h
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
	system := &OriginalOFFSystem{}
	fmt.Printf("Simulating storage system: %s\n\n", system.Name())

	// Create a sample file block to store. This one is LOW entropy.
	var textBlock Block
	copy(textBlock[:], "This is a low-entropy plain text file block. It does not look random at all.")
	log.Println("Storing a LOW-ENTROPY text block.")
	system.Store(textBlock)
	fmt.Println(strings.Repeat("-", 40))

	// Create another sample file block to store. This one is HIGH entropy.
	var encryptedBlock Block
	_, _ = rand.Read(encryptedBlock[:])
	log.Println("Storing a HIGH-ENTROPY encrypted block.")
	system.Store(encryptedBlock)
	fmt.Println(strings.Repeat("-", 40))

	log.Printf("Simulation finished. Total blocks on network: %d\n", len(networkStorage))
}
