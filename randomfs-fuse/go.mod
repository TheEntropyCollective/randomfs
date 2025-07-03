module github.com/TheEntropyCollective/randomfs-fuse

go 1.24.4

replace github.com/TheEntropyCollective/randomfs-core => ../randomfs-core

require (
	github.com/TheEntropyCollective/randomfs-core v0.1.5
	github.com/hanwen/go-fuse/v2 v2.8.0
)

require (
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)
