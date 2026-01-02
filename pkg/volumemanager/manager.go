package volumemanager

import (
	"context"
	"fmt"

	"bharani/pkg/config"
	"bharani/pkg/erasure"
	"bharani/proto/osd"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Manager handles volume operations like transfer and erasure coding
type Manager struct {
	config     *config.Config
	encoder    *erasure.Encoder
	osdClients map[string]osd.OSDServiceClient
}

// NewManager creates a new Volume Manager
func NewManager(cfg *config.Config) (*Manager, error) {
	encoder, err := erasure.NewEncoder(cfg.DataShards, cfg.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create erasure encoder: %w", err)
	}

	return &Manager{
		config:     cfg,
		encoder:    encoder,
		osdClients: make(map[string]osd.OSDServiceClient),
	}, nil
}

// GetOSDClient gets or creates a gRPC client for an OSD
func (m *Manager) GetOSDClient(osdAddress string) (osd.OSDServiceClient, error) {
	if client, exists := m.osdClients[osdAddress]; exists {
		return client, nil
	}

	conn, err := grpc.NewClient(osdAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to OSD %s: %w", osdAddress, err)
	}

	client := osd.NewOSDServiceClient(conn)
	m.osdClients[osdAddress] = client
	return client, nil
}

// CopyVolume copies a volume from source OSDs to target OSD
func (m *Manager) CopyVolume(ctx context.Context, volumeID, bucketID string, sourceOSDs []string, targetOSD string) error {
	if len(sourceOSDs) == 0 {
		return fmt.Errorf("no source OSDs provided")
	}

	_, err := m.GetOSDClient(sourceOSDs[0])
	if err != nil {
		return err
	}

	_, err = m.GetOSDClient(targetOSD)
	if err != nil {
		return err
	}

	return fmt.Errorf("volume copy not fully implemented - requires block enumeration")
}

// ErasureCodeVolume performs erasure coding on a volume
func (m *Manager) ErasureCodeVolume(ctx context.Context, volumeID, bucketID string, sourceOSDs []string, targetOSDs []string) error {
	if len(targetOSDs) < m.encoder.GetShardCount() {
		return fmt.Errorf("not enough target OSDs for erasure coding: need %d, have %d",
			m.encoder.GetShardCount(), len(targetOSDs))
	}

	return fmt.Errorf("erasure coding not fully implemented - requires block reading/writing")
}

// ReconstructBlock reconstructs a block from erasure-coded shards
func (m *Manager) ReconstructBlock(ctx context.Context, hash, bucketID, volumeID string, osdAddresses []string) ([]byte, error) {
	shards := make([][]byte, m.encoder.GetShardCount())
	shardsRead := 0

	for i, osdAddr := range osdAddresses {
		if i >= m.encoder.GetShardCount() {
			break
		}

		client, err := m.GetOSDClient(osdAddr)
		if err != nil {
			continue
		}

		resp, err := client.GetBlock(ctx, &osd.GetBlockRequest{
			Hash:     hash,
			BucketId: bucketID,
			VolumeId: volumeID,
		})

		if err == nil && resp.Success {
			shards[i] = resp.Data
			shardsRead++
		}
	}

	if shardsRead < m.encoder.GetDataShardCount() {
		return nil, fmt.Errorf("not enough shards to reconstruct: have %d, need %d",
			shardsRead, m.encoder.GetDataShardCount())
	}

	data, err := m.encoder.Decode(shards)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return data, nil
}
