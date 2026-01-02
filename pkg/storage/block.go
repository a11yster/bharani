package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Block represents an immutable block of data
type Block struct {
	Hash string // SHA-256 hash of the block data
	Data []byte // The actual block data (up to 4MB)
}

// NewBlock creates a new block from data and computes its hash
func NewBlock(data []byte) (*Block, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("block data cannot be empty")
	}

	hash := ComputeHash(data)
	return &Block{
		Hash: hash,
		Data: data,
	}, nil
}

// ComputeHash computes the SHA-256 hash of the data
func ComputeHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Validate checks if the block's hash matches its data
func (b *Block) Validate() bool {
	expectedHash := ComputeHash(b.Data)
	return b.Hash == expectedHash
}

// Size returns the size of the block in bytes
func (b *Block) Size() int64 {
	return int64(len(b.Data))
}

