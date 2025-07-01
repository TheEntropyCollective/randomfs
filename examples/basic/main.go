package main

import (
	"fmt"
	"log"
	"os"

	"github.com/TheEntropyCollective/randomfs-core/randomfs-core"
)

func main() {
	// Create a new RandomFS instance
	rfs, err := randomfs.NewRandomFS("localhost:5001", 8080, "./data", 100*1024*1024)
	if err != nil {
		log.Fatalf("Failed to create RandomFS: %v", err)
	}

	// Example file data
	testData := []byte("Hello, RandomFS! This is a test file demonstrating the Owner Free File System.")
	filename := "test.txt"
	contentType := "text/plain"

	// Store the file
	fmt.Printf("Storing file: %s\n", filename)
	randomURL, err := rfs.StoreFile(filename, testData, contentType)
	if err != nil {
		log.Fatalf("Failed to store file: %v", err)
	}

	fmt.Printf("✅ File stored successfully!\n")
	fmt.Printf("rd:// URL: %s\n", randomURL.String())
	fmt.Printf("IPFS Hash: %s\n", randomURL.RepHash)

	// Retrieve the file
	fmt.Printf("\nRetrieving file from hash: %s\n", randomURL.RepHash)
	retrievedData, rep, err := rfs.RetrieveFile(randomURL.RepHash)
	if err != nil {
		log.Fatalf("Failed to retrieve file: %v", err)
	}

	fmt.Printf("✅ File retrieved successfully!\n")
	fmt.Printf("Original filename: %s\n", rep.FileName)
	fmt.Printf("File size: %d bytes\n", rep.FileSize)
	fmt.Printf("Content type: %s\n", rep.ContentType)
	fmt.Printf("Block size used: %d bytes\n", rep.BlockSize)
	fmt.Printf("Number of blocks: %d\n", len(rep.BlockHashes))
	fmt.Printf("Retrieved content: %s\n", string(retrievedData))

	// Verify data integrity
	if string(retrievedData) == string(testData) {
		fmt.Printf("✅ Data integrity verified!\n")
	} else {
		fmt.Printf("❌ Data integrity check failed!\n")
		os.Exit(1)
	}

	// Demonstrate rd:// URL parsing
	fmt.Printf("\nTesting rd:// URL parsing:\n")
	parsedURL, err := randomfs.ParseRandomURL(randomURL.String())
	if err != nil {
		log.Fatalf("Failed to parse rd:// URL: %v", err)
	}

	fmt.Printf("Parsed URL components:\n")
	fmt.Printf("  Scheme: %s\n", parsedURL.Scheme)
	fmt.Printf("  Host: %s\n", parsedURL.Host)
	fmt.Printf("  Version: %s\n", parsedURL.Version)
	fmt.Printf("  Filename: %s\n", parsedURL.FileName)
	fmt.Printf("  File size: %d\n", parsedURL.FileSize)
	fmt.Printf("  Timestamp: %d\n", parsedURL.Timestamp)
	fmt.Printf("  Rep hash: %s\n", parsedURL.RepHash)

	fmt.Printf("\n🎉 RandomFS example completed successfully!\n")
}
