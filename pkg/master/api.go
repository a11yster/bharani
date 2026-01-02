package master

import (
	"context"

	"bharani/proto/master"
)

// MasterService implements the gRPC Master service
type MasterService struct {
	master *Master
	master.UnimplementedMasterServiceServer
}

// NewMasterService creates a new Master service
func NewMasterService(masterInstance *Master) *MasterService {
	return &MasterService{
		master: masterInstance,
	}
}

// RegisterOSD handles RegisterOSD requests
func (s *MasterService) RegisterOSD(ctx context.Context, req *master.RegisterOSDRequest) (*master.RegisterOSDResponse, error) {
	return s.master.RegisterOSD(ctx, req)
}

// Heartbeat handles Heartbeat requests
func (s *MasterService) Heartbeat(ctx context.Context, req *master.HeartbeatRequest) (*master.HeartbeatResponse, error) {
	return s.master.Heartbeat(ctx, req)
}

// GetOpenVolumes handles GetOpenVolumes requests
func (s *MasterService) GetOpenVolumes(ctx context.Context, req *master.GetOpenVolumesRequest) (*master.GetOpenVolumesResponse, error) {
	return s.master.GetOpenVolumes(ctx, req)
}

// CloseVolume handles CloseVolume requests
func (s *MasterService) CloseVolume(ctx context.Context, req *master.CloseVolumeRequest) (*master.CloseVolumeResponse, error) {
	return s.master.CloseVolume(ctx, req)
}

// TriggerRepair handles TriggerRepair requests
func (s *MasterService) TriggerRepair(ctx context.Context, req *master.TriggerRepairRequest) (*master.TriggerRepairResponse, error) {
	return s.master.TriggerRepair(ctx, req)
}


