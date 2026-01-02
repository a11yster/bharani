package storage

import (
	"fmt"
	"sync"
)

// VolumeState represents the state of a volume
type VolumeState string

const (
	VolumeStateOpen   VolumeState = "open"   // Accepting new writes
	VolumeStateClosed VolumeState = "closed" // No longer accepting writes
)

// Volume represents one or more buckets replicated across OSDs
type Volume struct {
	ID         string             // Unique volume identifier
	Buckets    map[string]*Bucket // Map of bucket ID -> bucket
	OSDs       []string           // List of OSD addresses storing this volume
	State      VolumeState        // Current state (open/closed)
	Generation int64              // Generation number for consistency
	CellID     string             // Cell where this volume is stored
	mu         sync.RWMutex       // Mutex for thread-safe access
}

// NewVolume creates a new volume
func NewVolume(id, cellID string) *Volume {
	return &Volume{
		ID:         id,
		Buckets:    make(map[string]*Bucket),
		OSDs:       make([]string, 0),
		State:      VolumeStateOpen,
		Generation: 1,
		CellID:     cellID,
	}
}

// AddBucket adds a bucket to the volume
func (v *Volume) AddBucket(bucket *Bucket) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.State == VolumeStateClosed {
		return fmt.Errorf("volume %s is closed, cannot add bucket", v.ID)
	}

	if _, exists := v.Buckets[bucket.ID]; exists {
		return fmt.Errorf("bucket %s already exists in volume", bucket.ID)
	}

	v.Buckets[bucket.ID] = bucket
	return nil
}

// GetBucket retrieves a bucket by ID
func (v *Volume) GetBucket(bucketID string) (*Bucket, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	bucket, exists := v.Buckets[bucketID]
	return bucket, exists
}

// Close marks the volume as closed
func (v *Volume) Close() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.State = VolumeStateClosed
}

// IsOpen checks if the volume is open
func (v *Volume) IsOpen() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.State == VolumeStateOpen
}

// IncrementGeneration increments the generation number
func (v *Volume) IncrementGeneration() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Generation++
}

// SetOSDs sets the list of OSDs storing this volume
func (v *Volume) SetOSDs(osds []string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.OSDs = make([]string, len(osds))
	copy(v.OSDs, osds)
}

// GetOSDs returns the list of OSDs
func (v *Volume) GetOSDs() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	osds := make([]string, len(v.OSDs))
	copy(osds, v.OSDs)
	return osds
}

