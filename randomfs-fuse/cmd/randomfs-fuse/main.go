package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/TheEntropyCollective/randomfs-core/pkg/randomfs"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var (
	mountPoint = flag.String("mount", "", "Mount point directory")
	dataDir    = flag.String("data", "./data", "Data directory for RandomFS")
	ipfsAPI    = flag.String("ipfs", "http://localhost:5001", "IPFS API endpoint (empty to disable)")
	noIPFS     = flag.Bool("no-ipfs", false, "Disable IPFS (for testing without IPFS)")
	password   = flag.String("password", "fuse-default-password", "Default password for file operations")
)

type RandomFSNode struct {
	fs.Inode
	rfs      *randomfs.RandomFS
	password string
}

func (n *RandomFSNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755 | syscall.S_IFDIR
	out.Size = 0
	return 0
}

func (n *RandomFSNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if name == ".randomfs-info" {
		child := n.Inode.NewInode(ctx, &RandomFSInfoNode{rfs: n.rfs}, fs.StableAttr{Mode: syscall.S_IFREG})
		return child, 0
	}
	return nil, syscall.ENOENT
}

func (n *RandomFSNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entries := []fuse.DirEntry{
		{Name: ".randomfs-info", Mode: syscall.S_IFREG},
	}
	return fs.NewListDirStream(entries), 0
}

type RandomFSInfoNode struct {
	fs.Inode
	rfs *randomfs.RandomFS
}

func (n *RandomFSInfoNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	stats := n.rfs.GetStats()
	info := fmt.Sprintf("RandomFS Info\nFiles Stored: %d\nBlocks Generated: %d\nTotal Size: %d bytes\nCache Hits: %d\nCache Misses: %d\n",
		stats.FilesStored, stats.BlocksGenerated, stats.TotalSize, stats.CacheHits, stats.CacheMisses)
	
	out.Size = uint64(len(info))
	out.Mode = 0644 | syscall.S_IFREG
	return 0
}

func (n *RandomFSInfoNode) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	stats := n.rfs.GetStats()
	info := fmt.Sprintf("RandomFS Info\nFiles Stored: %d\nBlocks Generated: %d\nTotal Size: %d bytes\nCache Hits: %d\nCache Misses: %d\n",
		stats.FilesStored, stats.BlocksGenerated, stats.TotalSize, stats.CacheHits, stats.CacheMisses)
	
	if off >= int64(len(info)) {
		return fuse.ReadResultData(nil), 0
	}
	
	end := off + int64(len(dest))
	if end > int64(len(info)) {
		end = int64(len(info))
	}
	
	return fuse.ReadResultData([]byte(info[off:end])), 0
}

func main() {
	flag.Parse()

	if *mountPoint == "" {
		log.Fatal("Please specify a mount point with -mount")
	}

	if err := os.MkdirAll(*mountPoint, 0755); err != nil {
		log.Fatalf("Failed to create mount point: %v", err)
	}

	var rfs *randomfs.RandomFS
	var err error

	if *noIPFS || *ipfsAPI == "" {
		rfs, err = randomfs.NewRandomFSWithoutIPFS(*dataDir, 100*1024*1024)
	} else {
		rfs, err = randomfs.NewRandomFS(*ipfsAPI, *dataDir, 100*1024*1024)
	}

	if err != nil {
		log.Fatalf("Failed to initialize RandomFS: %v", err)
	}

	root := &RandomFSNode{
		rfs:      rfs,
		password: *password,
	}

	server, err := fs.Mount(*mountPoint, root, &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug: true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to mount filesystem: %v", err)
	}

	log.Printf("RandomFS mounted at %s", *mountPoint)
	log.Printf("Data directory: %s", *dataDir)
	if *noIPFS || *ipfsAPI == "" {
		log.Printf("IPFS: disabled")
	} else {
		log.Printf("IPFS API: %s", *ipfsAPI)
	}
	log.Printf("Default password: %s", *password)
	log.Printf("Press Ctrl+C to unmount")

	server.Wait()
}
