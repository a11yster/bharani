package osd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Storage handles disk storage for blocks
type Storage struct {
	dataDir string
	mu      sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Storage{
		dataDir: dataDir,
	}, nil
}

// StoreBlock stores a block on disk
func (s *Storage) StoreBlock(cellID, bucketID, hash string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	blockPath := s.getBlockPath(cellID, bucketID, hash)
	blockDir := filepath.Dir(blockPath)

	if err := os.MkdirAll(blockDir, 0755); err != nil {
		return fmt.Errorf("failed to create block directory: %w", err)
	}

	file, err := os.Create(blockPath)
	if err != nil {
		return fmt.Errorf("failed to create block file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write block data: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync block file: %w", err)
	}

	return nil
}

// GetBlock retrieves a block from disk
func (s *Storage) GetBlock(cellID, bucketID, hash string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blockPath := s.getBlockPath(cellID, bucketID, hash)
	data, err := os.ReadFile(blockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("block not found: %s", hash)
		}
		return nil, fmt.Errorf("failed to read block: %w", err)
	}

	return data, nil
}

// HasBlock checks if a block exists
func (s *Storage) HasBlock(cellID, bucketID, hash string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blockPath := s.getBlockPath(cellID, bucketID, hash)
	_, err := os.Stat(blockPath)
	return err == nil
}

// getBlockPath returns the file path for a block
func (s *Storage) getBlockPath(cellID, bucketID, hash string) string {
	return filepath.Join(s.dataDir, cellID, bucketID, hash)
}

// GetAvailableSpace returns the available space in bytes
func (s *Storage) GetAvailableSpace() (int64, error) {
	return 100 * 1024 * 1024 * 1024 * 1024, nil
}
