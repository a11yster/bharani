package osd

import (
	"context"

	"bharani/proto/osd"
)

// OSDService implements the gRPC OSD service
type OSDService struct {
	osd *OSD
	osd.UnimplementedOSDServiceServer
}

// NewOSDService creates a new OSD service
func NewOSDService(osdInstance *OSD) *OSDService {
	return &OSDService{
		osd: osdInstance,
	}
}

// PutBlock handles PutBlock requests
func (s *OSDService) PutBlock(ctx context.Context, req *osd.PutBlockRequest) (*osd.PutBlockResponse, error) {
	err := s.osd.PutBlock(ctx, req.Hash, req.BucketId, req.VolumeId, req.Data)
	if err != nil {
		return &osd.PutBlockResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &osd.PutBlockResponse{
		Success: true,
	}, nil
}

// GetBlock handles GetBlock requests
func (s *OSDService) GetBlock(ctx context.Context, req *osd.GetBlockRequest) (*osd.GetBlockResponse, error) {
	data, err := s.osd.GetBlock(ctx, req.Hash, req.BucketId, req.VolumeId)
	if err != nil {
		return &osd.GetBlockResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &osd.GetBlockResponse{
		Success: true,
		Data:    data,
	}, nil
}

// HealthCheck handles health check requests
func (s *OSDService) HealthCheck(ctx context.Context, req *osd.HealthCheckRequest) (*osd.HealthCheckResponse, error) {
	healthy := s.osd.HealthCheck()
	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	return &osd.HealthCheckResponse{
		Healthy: healthy,
		Status:  status,
	}, nil
}

