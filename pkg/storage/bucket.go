package storage

import (
	"fmt"
	"sync"
)

// Bucket represents a logical container for blocks (1GB)
type Bucket struct {
	ID      string            // Unique bucket identifier
	Blocks  map[string]*Block // Map of hash -> block
	Size    int64             // Current size in bytes
	MaxSize int64             // Maximum size (1GB)
	mu      sync.RWMutex      // Mutex for thread-safe access
}

// NewBucket creates a new bucket with the given ID
func NewBucket(id string, maxSize int64) *Bucket {
	return &Bucket{
		ID:      id,
		Blocks:  make(map[string]*Block),
		Size:    0,
		MaxSize: maxSize,
	}
}

// AddBlock adds a block to the bucket
func (b *Bucket) AddBlock(block *Block) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.Size+block.Size() > b.MaxSize {
		return fmt.Errorf("bucket %s is full (size: %d, max: %d)", b.ID, b.Size, b.MaxSize)
	}

	if _, exists := b.Blocks[block.Hash]; exists {
		return fmt.Errorf("block %s already exists in bucket", block.Hash)
	}

	b.Blocks[block.Hash] = block
	b.Size += block.Size()
	return nil
}

// GetBlock retrieves a block by hash
func (b *Bucket) GetBlock(hash string) (*Block, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	block, exists := b.Blocks[hash]
	return block, exists
}

// HasBlock checks if a block exists in the bucket
func (b *Bucket) HasBlock(hash string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	_, exists := b.Blocks[hash]
	return exists
}

// IsFull checks if the bucket is full
func (b *Bucket) IsFull() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.Size >= b.MaxSize
}

// GetBlocks returns all blocks in the bucket
func (b *Bucket) GetBlocks() map[string]*Block {
	b.mu.RLock()
	defer b.mu.RUnlock()

	blocks := make(map[string]*Block)
	for hash, block := range b.Blocks {
		blocks[hash] = block
	}
	return blocks
}

