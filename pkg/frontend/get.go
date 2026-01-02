package frontend

import (
	"context"
	"fmt"

	"bharani/proto/blockindex"
	"bharani/proto/osd"
	"bharani/proto/replication"
)

// Get retrieves a block from the system
func (f *Frontend) Get(ctx context.Context, hash string) ([]byte, error) {
	getEntryReq := &blockindex.GetEntryRequest{Hash: hash}
	getEntryResp, err := f.blockIndexClient.GetEntry(ctx, getEntryReq)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup block: %w", err)
	}

	if !getEntryResp.Found {
		return nil, fmt.Errorf("block not found: %s", hash)
	}

	listVolumesReq := &replication.ListVolumesRequest{
		CellId: getEntryResp.CellId,
	}
	listVolumesResp, err := f.replicationClient.ListVolumes(ctx, listVolumesReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	for _, volumeID := range listVolumesResp.VolumeIds {
		getVolumeReq := &replication.GetVolumeRequest{VolumeId: volumeID}
		getVolumeResp, err := f.replicationClient.GetVolume(ctx, getVolumeReq)
		if err != nil || !getVolumeResp.Found {
			continue
		}

		for _, osdAddr := range getVolumeResp.OsdAddresses {
			client, err := f.GetOSDClient(osdAddr)
			if err != nil {
				continue
			}

			getBlockReq := &osd.GetBlockRequest{
				Hash:     hash,
				BucketId: getEntryResp.BucketId,
				VolumeId: volumeID,
			}

			getBlockResp, err := client.GetBlock(ctx, getBlockReq)
			if err == nil && getBlockResp.Success {
				return getBlockResp.Data, nil
			}
		}
	}

	return nil, fmt.Errorf("block not found on any OSD")
}

