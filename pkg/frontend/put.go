package frontend

import (
	"context"
	"fmt"

	"bharani/pkg/storage"
	"bharani/proto/blockindex"
	"bharani/proto/master"
	"bharani/proto/osd"
	"bharani/proto/replication"

	"github.com/google/uuid"
)

// Put stores a block in the system
func (f *Frontend) Put(ctx context.Context, data []byte) (string, error) {
	block, err := storage.NewBlock(data)
	if err != nil {
		return "", fmt.Errorf("failed to create block: %w", err)
	}

	existsReq := &blockindex.ExistsRequest{Hash: block.Hash}
	existsResp, err := f.blockIndexClient.Exists(ctx, existsReq)
	if err == nil && existsResp.Exists {
		return block.Hash, nil
	}

	openVolumesReq := &master.GetOpenVolumesRequest{
		CellId: f.config.CellID,
	}
	openVolumesResp, err := f.masterClient.GetOpenVolumes(ctx, openVolumesReq)
	if err != nil {
		return "", fmt.Errorf("failed to get open volumes: %w", err)
	}

	var volumeID string
	if len(openVolumesResp.VolumeIds) > 0 {
		volumeID = openVolumesResp.VolumeIds[0]
	} else {
		volumeID = uuid.New().String()
		healthyOSDs := f.getHealthyOSDs(ctx)
		if len(healthyOSDs) < f.config.ReplicationFactor {
			return "", fmt.Errorf("not enough healthy OSDs: need %d, have %d",
				f.config.ReplicationFactor, len(healthyOSDs))
		}

		selectedOSDs := healthyOSDs[:f.config.ReplicationFactor]

		createReq := &replication.CreateVolumeRequest{
			VolumeId:     volumeID,
			OsdAddresses: selectedOSDs,
			CellId:       f.config.CellID,
		}
		_, err = f.replicationClient.CreateVolume(ctx, createReq)
		if err != nil {
			return "", fmt.Errorf("failed to create volume: %w", err)
		}
	}

	getVolumeReq := &replication.GetVolumeRequest{VolumeId: volumeID}
	getVolumeResp, err := f.replicationClient.GetVolume(ctx, getVolumeReq)
	if err != nil || !getVolumeResp.Found {
		return "", fmt.Errorf("failed to get volume info: %w", err)
	}

	bucketID := uuid.New().String()

	successCount := 0
	for _, osdAddr := range getVolumeResp.OsdAddresses {
		client, err := f.GetOSDClient(osdAddr)
		if err != nil {
			continue
		}

		putReq := &osd.PutBlockRequest{
			Hash:     block.Hash,
			Data:     block.Data,
			BucketId: bucketID,
			VolumeId: volumeID,
		}

		putResp, err := client.PutBlock(ctx, putReq)
		if err == nil && putResp.Success {
			successCount++
		}
	}

	if successCount < f.config.ReplicationFactor {
		return "", fmt.Errorf("failed to replicate block: only %d/%d writes succeeded",
			successCount, f.config.ReplicationFactor)
	}

	putEntryReq := &blockindex.PutEntryRequest{
		Hash:     block.Hash,
		CellId:   f.config.CellID,
		BucketId: bucketID,
		Checksum: block.Hash,
	}

	_, err = f.blockIndexClient.PutEntry(ctx, putEntryReq)
	if err != nil {
		return "", fmt.Errorf("block stored but index update failed: %w", err)
	}

	return block.Hash, nil
}

// getHealthyOSDs gets list of healthy OSDs
func (f *Frontend) getHealthyOSDs(ctx context.Context) []string {
	return []string{"localhost:9090", "localhost:9095", "localhost:9096"}
}
