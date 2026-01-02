package storage

import (
	"testing"
)

func TestNewBlock(t *testing.T) {
	data := []byte("Hello, World!")
	block, err := NewBlock(data)
	if err != nil {
		t.Fatalf("Failed to create block: %v", err)
	}

	if block.Hash == "" {
		t.Error("Block hash should not be empty")
	}

	if len(block.Data) != len(data) {
		t.Errorf("Block data length mismatch: got %d, want %d", len(block.Data), len(data))
	}
}

func TestBlockValidate(t *testing.T) {
	data := []byte("Test data")
	block, err := NewBlock(data)
	if err != nil {
		t.Fatalf("Failed to create block: %v", err)
	}

	if !block.Validate() {
		t.Error("Block validation should pass for valid block")
	}

	// Corrupt the data
	block.Data[0] = 'X'
	if block.Validate() {
		t.Error("Block validation should fail for corrupted data")
	}
}

func TestComputeHash(t *testing.T) {
	data1 := []byte("Hello")
	data2 := []byte("Hello")
	data3 := []byte("World")

	hash1 := ComputeHash(data1)
	hash2 := ComputeHash(data2)
	hash3 := ComputeHash(data3)

	if hash1 != hash2 {
		t.Error("Same data should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("Different data should produce different hashes")
	}
}


