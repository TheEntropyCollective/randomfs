module superlinear-growth-demo

go 1.23.0

toolchain go1.24.4

require github.com/TheEntropyCollective/randomfs-core v0.0.0

require (
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/TheEntropyCollective/randomfs-core => ../../randomfs-core
