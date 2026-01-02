package erasure

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	encoder, err := NewEncoder(10, 4)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	originalData := []byte("This is test data for erasure coding. It should be encoded and decoded correctly.")
	
	// Encode
	shards, err := encoder.Encode(originalData)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	if len(shards) != 14 {
		t.Errorf("Expected 14 shards, got %d", len(shards))
	}

	// Decode with all shards
	decodedData, err := encoder.Decode(shards)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if !bytes.Equal(decodedData[:len(originalData)], originalData) {
		t.Error("Decoded data does not match original")
	}
}

func TestDecodeWithMissingShards(t *testing.T) {
	encoder, err := NewEncoder(10, 4)
	if err != nil {
		t.Fatalf("Failed to create encoder: %v", err)
	}

	originalData := []byte("Test data for missing shard recovery")
	
	// Encode
	shards, err := encoder.Encode(originalData)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	// Remove some shards (simulate failures)
	shards[0] = nil
	shards[5] = nil
	shards[12] = nil

	// Decode with missing shards
	decodedData, err := encoder.Decode(shards)
	if err != nil {
		t.Fatalf("Failed to decode with missing shards: %v", err)
	}

	if !bytes.Equal(decodedData[:len(originalData)], originalData) {
		t.Error("Decoded data does not match original after shard recovery")
	}
}


