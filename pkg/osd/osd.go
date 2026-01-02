package osd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"bharani/pkg/config"
)

// OSD represents an Object Storage Daemon
type OSD struct {
	config        *config.Config
	storage       *Storage
	address       string
	cellID        string
	healthy       bool
	mu            sync.RWMutex
	lastHeartbeat time.Time
}

// NewOSD creates a new OSD instance
func NewOSD(cfg *config.Config, address, cellID string) (*OSD, error) {
	storage, err := NewStorage(cfg.OSDDataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	return &OSD{
		config:        cfg,
		storage:       storage,
		address:       address,
		cellID:        cellID,
		healthy:       true,
		lastHeartbeat: time.Now(),
	}, nil
}

// PutBlock stores a block on this OSD
func (o *OSD) PutBlock(ctx context.Context, hash, bucketID, volumeID string, data []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.healthy {
		return fmt.Errorf("OSD is not healthy")
	}

	return o.storage.StoreBlock(o.cellID, bucketID, hash, data)
}

// GetBlock retrieves a block from this OSD
func (o *OSD) GetBlock(ctx context.Context, hash, bucketID, volumeID string) ([]byte, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if !o.healthy {
		return nil, fmt.Errorf("OSD is not healthy")
	}

	return o.storage.GetBlock(o.cellID, bucketID, hash)
}

// HealthCheck returns the health status
func (o *OSD) HealthCheck() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if time.Since(o.lastHeartbeat) > 2*time.Minute {
		o.healthy = false
	}

	return o.healthy
}

// SetHealthy sets the health status
func (o *OSD) SetHealthy(healthy bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.healthy = healthy
	o.lastHeartbeat = time.Now()
}

// GetAddress returns the OSD address
func (o *OSD) GetAddress() string {
	return o.address
}

// GetCellID returns the cell ID
func (o *OSD) GetCellID() string {
	return o.cellID
}

// GetAvailableSpace returns available storage space
func (o *OSD) GetAvailableSpace() (int64, error) {
	return o.storage.GetAvailableSpace()
}

