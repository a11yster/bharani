package master

import (
	"context"
	"fmt"

	"bharani/proto/replication"
)

// RepairPlan represents a plan for repairing volumes after OSD failure
type RepairPlan struct {
	FailedOSD    string
	Volumes      []VolumeRepair
	SourceOSDs   []string
	TargetOSDs   []string
}

// VolumeRepair represents repair information for a single volume
type VolumeRepair struct {
	VolumeID   string
	SourceOSDs []string
	TargetOSD  string
}

// BuildRepairPlan builds a repair plan for a failed OSD
func (m *Master) BuildRepairPlan(ctx context.Context, failedOSD string) (*RepairPlan, error) {
	replicationClient := replication.NewReplicationTableServiceClient(m.replicationConn)

	listReq := &replication.ListVolumesRequest{
		CellId: m.cellID,
	}

	listResp, err := replicationClient.ListVolumes(ctx, listReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	volumesToRepair := make([]VolumeRepair, 0)
	healthyOSDs := m.GetHealthyOSDs()

	if len(healthyOSDs) == 0 {
		return nil, fmt.Errorf("no healthy OSDs available for repair")
	}

	for _, volumeID := range listResp.VolumeIds {
		getReq := &replication.GetVolumeRequest{VolumeId: volumeID}
		getResp, err := replicationClient.GetVolume(ctx, getReq)
		if err != nil || !getResp.Found {
			continue
		}

		onFailedOSD := false
		for _, osdAddr := range getResp.OsdAddresses {
			if osdAddr == failedOSD {
				onFailedOSD = true
				break
			}
		}

		if onFailedOSD {
			sourceOSDs := make([]string, 0)
			for _, osdAddr := range getResp.OsdAddresses {
				if osdAddr != failedOSD {
					for _, healthyOSD := range healthyOSDs {
						if osdAddr == healthyOSD {
							sourceOSDs = append(sourceOSDs, osdAddr)
							break
						}
					}
				}
			}

			if len(sourceOSDs) > 0 {
				targetOSD := healthyOSDs[0]

				volumesToRepair = append(volumesToRepair, VolumeRepair{
					VolumeID:   volumeID,
					SourceOSDs: sourceOSDs,
					TargetOSD:  targetOSD,
				})
			}
		}
	}

	return &RepairPlan{
		FailedOSD:  failedOSD,
		Volumes:    volumesToRepair,
		SourceOSDs: healthyOSDs,
		TargetOSDs: healthyOSDs,
	}, nil
}

