package frontend

import (
	"context"

	"bharani/proto/frontend"
)

// FrontendService implements the gRPC Frontend service
type FrontendService struct {
	frontend *Frontend
	frontend.UnimplementedFrontendServiceServer
}

// NewFrontendService creates a new Frontend service
func NewFrontendService(frontendInstance *Frontend) *FrontendService {
	return &FrontendService{
		frontend: frontendInstance,
	}
}

// Put handles Put requests
func (s *FrontendService) Put(ctx context.Context, req *frontend.PutRequest) (*frontend.PutResponse, error) {
	hash, err := s.frontend.Put(ctx, req.Data)
	if err != nil {
		return &frontend.PutResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &frontend.PutResponse{
		Success: true,
		Hash:    hash,
	}, nil
}

// Get handles Get requests
func (s *FrontendService) Get(ctx context.Context, req *frontend.GetRequest) (*frontend.GetResponse, error) {
	data, err := s.frontend.Get(ctx, req.Hash)
	if err != nil {
		return &frontend.GetResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &frontend.GetResponse{
		Success: true,
		Data:    data,
	}, nil
}


