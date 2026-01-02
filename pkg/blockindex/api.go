package blockindex

import (
	"context"

	"bharani/proto/blockindex"
)

// BlockIndexService implements the gRPC BlockIndex service
type BlockIndexService struct {
	index *Index
	blockindex.UnimplementedBlockIndexServiceServer
}

// NewBlockIndexService creates a new BlockIndex service
func NewBlockIndexService(index *Index) *BlockIndexService {
	return &BlockIndexService{
		index: index,
	}
}

// PutEntry handles PutEntry requests
func (s *BlockIndexService) PutEntry(ctx context.Context, req *blockindex.PutEntryRequest) (*blockindex.PutEntryResponse, error) {
	entry := &Entry{
		Hash:     req.Hash,
		CellID:   req.CellId,
		BucketID: req.BucketId,
		Checksum: req.Checksum,
	}

	err := s.index.PutEntry(entry)
	if err != nil {
		return &blockindex.PutEntryResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &blockindex.PutEntryResponse{
		Success: true,
	}, nil
}

// GetEntry handles GetEntry requests
func (s *BlockIndexService) GetEntry(ctx context.Context, req *blockindex.GetEntryRequest) (*blockindex.GetEntryResponse, error) {
	entry, err := s.index.GetEntry(req.Hash)
	if err != nil {
		return &blockindex.GetEntryResponse{
			Found: false,
			Error: err.Error(),
		}, nil
	}

	if entry == nil {
		return &blockindex.GetEntryResponse{
			Found: false,
		}, nil
	}

	return &blockindex.GetEntryResponse{
		Found:    true,
		CellId:   entry.CellID,
		BucketId: entry.BucketID,
		Checksum: entry.Checksum,
	}, nil
}

// Exists handles Exists requests
func (s *BlockIndexService) Exists(ctx context.Context, req *blockindex.ExistsRequest) (*blockindex.ExistsResponse, error) {
	exists, err := s.index.Exists(req.Hash)
	if err != nil {
		return &blockindex.ExistsResponse{
			Exists: false,
		}, nil
	}

	return &blockindex.ExistsResponse{
		Exists: exists,
	}, nil
}

