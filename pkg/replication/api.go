package replication

import (
	"context"

	"bharani/proto/replication"
)

// ReplicationTableService implements the gRPC ReplicationTable service
type ReplicationTableService struct {
	table *Table
	replication.UnimplementedReplicationTableServiceServer
}

// NewReplicationTableService creates a new ReplicationTable service
func NewReplicationTableService(table *Table) *ReplicationTableService {
	return &ReplicationTableService{
		table: table,
	}
}

// CreateVolume handles CreateVolume requests
func (s *ReplicationTableService) CreateVolume(ctx context.Context, req *replication.CreateVolumeRequest) (*replication.CreateVolumeResponse, error) {
	err := s.table.CreateVolume(req.VolumeId, req.OsdAddresses, req.CellId)
	if err != nil {
		return &replication.CreateVolumeResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &replication.CreateVolumeResponse{
		Success: true,
	}, nil
}

// GetVolume handles GetVolume requests
func (s *ReplicationTableService) GetVolume(ctx context.Context, req *replication.GetVolumeRequest) (*replication.GetVolumeResponse, error) {
	info, err := s.table.GetVolume(req.VolumeId)
	if err != nil {
		return &replication.GetVolumeResponse{
			Found: false,
			Error: err.Error(),
		}, nil
	}

	if info == nil {
		return &replication.GetVolumeResponse{
			Found: false,
		}, nil
	}

	return &replication.GetVolumeResponse{
		Found:        true,
		VolumeId:     info.VolumeID,
		OsdAddresses: info.OSDAddresses,
		Generation:   info.Generation,
		CellId:       info.CellID,
		State:        info.State,
	}, nil
}

// UpdateVolume handles UpdateVolume requests
func (s *ReplicationTableService) UpdateVolume(ctx context.Context, req *replication.UpdateVolumeRequest) (*replication.UpdateVolumeResponse, error) {
	err := s.table.UpdateVolume(req.VolumeId, req.OsdAddresses, req.Generation, req.State)
	if err != nil {
		return &replication.UpdateVolumeResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &replication.UpdateVolumeResponse{
		Success: true,
	}, nil
}

// ListVolumes handles ListVolumes requests
func (s *ReplicationTableService) ListVolumes(ctx context.Context, req *replication.ListVolumesRequest) (*replication.ListVolumesResponse, error) {
	volumeIDs, err := s.table.ListVolumes(req.CellId)
	if err != nil {
		return &replication.ListVolumesResponse{}, nil
	}

	return &replication.ListVolumesResponse{
		VolumeIds: volumeIDs,
	}, nil
}


