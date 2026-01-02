package frontend

import (
	"bharani/pkg/config"
	"bharani/proto/blockindex"
	"bharani/proto/master"
	"bharani/proto/osd"
	"bharani/proto/replication"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Frontend coordinates Put/Get operations
type Frontend struct {
	config            *config.Config
	blockIndexClient  blockindex.BlockIndexServiceClient
	replicationClient replication.ReplicationTableServiceClient
	masterClient      master.MasterServiceClient
	osdClients        map[string]osd.OSDServiceClient
}

// NewFrontend creates a new Frontend instance
func NewFrontend(cfg *config.Config, blockIndexAddr, replicationAddr, masterAddr string) (*Frontend, error) {
	blockIndexConn, err := grpc.NewClient(blockIndexAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to block index: %w", err)
	}

	replicationConn, err := grpc.NewClient(replicationAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to replication table: %w", err)
	}

	masterConn, err := grpc.NewClient(masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master: %w", err)
	}

	return &Frontend{
		config:            cfg,
		blockIndexClient:  blockindex.NewBlockIndexServiceClient(blockIndexConn),
		replicationClient: replication.NewReplicationTableServiceClient(replicationConn),
		masterClient:      master.NewMasterServiceClient(masterConn),
		osdClients:        make(map[string]osd.OSDServiceClient),
	}, nil
}

// GetOSDClient gets or creates a gRPC client for an OSD
func (f *Frontend) GetOSDClient(osdAddress string) (osd.OSDServiceClient, error) {
	if client, exists := f.osdClients[osdAddress]; exists {
		return client, nil
	}

	conn, err := grpc.NewClient(osdAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to OSD %s: %w", osdAddress, err)
	}

	client := osd.NewOSDServiceClient(conn)
	f.osdClients[osdAddress] = client
	return client, nil
}
