package erasure

import (
	"fmt"

	"github.com/klauspost/reedsolomon"
)

// Encoder handles Reed-Solomon encoding and decoding
type Encoder struct {
	dataShards   int
	parityShards int
	encoder      reedsolomon.Encoder
}

// NewEncoder creates a new Reed-Solomon encoder
func NewEncoder(dataShards, parityShards int) (*Encoder, error) {
	encoder, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reed-solomon encoder: %w", err)
	}

	return &Encoder{
		dataShards:   dataShards,
		parityShards: parityShards,
		encoder:      encoder,
	}, nil
}

// Encode splits data into shards and computes parity shards
func (e *Encoder) Encode(data []byte) ([][]byte, error) {
	shards, err := e.encoder.Split(data)
	if err != nil {
		return nil, fmt.Errorf("failed to split data: %w", err)
	}

	shardSize := len(data) / e.dataShards
	if len(data)%e.dataShards != 0 {
		shardSize++
	}

	for i := 0; i < e.dataShards+e.parityShards; i++ {
		if len(shards[i]) < shardSize {
			padded := make([]byte, shardSize)
			copy(padded, shards[i])
			shards[i] = padded
		}
	}

	err = e.encoder.Encode(shards)
	if err != nil {
		return nil, fmt.Errorf("failed to encode: %w", err)
	}

	return shards, nil
}

// Decode reconstructs the original data from shards
func (e *Encoder) Decode(shards [][]byte) ([]byte, error) {
	validShards := 0
	for _, shard := range shards {
		if shard != nil {
			validShards++
		}
	}

	if validShards < e.dataShards {
		return nil, fmt.Errorf("not enough shards to reconstruct: have %d, need %d", validShards, e.dataShards)
	}

	err := e.encoder.Reconstruct(shards)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct: %w", err)
	}

	ok, err := e.encoder.Verify(shards)
	if err != nil {
		return nil, fmt.Errorf("failed to verify: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("data verification failed")
	}

	totalSize := 0
	for i := 0; i < e.dataShards; i++ {
		if shards[i] != nil {
			totalSize += len(shards[i])
		}
	}

	result := make([]byte, 0, totalSize)
	for i := 0; i < e.dataShards; i++ {
		if shards[i] != nil {
			result = append(result, shards[i]...)
		}
	}

	return result, nil
}

// GetShardCount returns the total number of shards (data + parity)
func (e *Encoder) GetShardCount() int {
	return e.dataShards + e.parityShards
}

// GetDataShardCount returns the number of data shards
func (e *Encoder) GetDataShardCount() int {
	return e.dataShards
}

// GetParityShardCount returns the number of parity shards
func (e *Encoder) GetParityShardCount() int {
	return e.parityShards
}

