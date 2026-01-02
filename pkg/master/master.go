package master

import (
	"context"
	"fmt"
	"sync"
	"time"

	"bharani/pkg/config"
	"bharani/proto/master"
	"bharani/proto/osd"
	"bharani/proto/replication"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OSDInfo tracks OSD health and metadata
type OSDInfo struct {
	Address        string
	CellID         string
	AvailableSpace int64
	LastHeartbeat  time.Time
	Healthy        bool
}

// Master coordinates repairs, volume management, and OSD monitoring
type Master struct {
	config          *config.Config
	cellID          string
	osds            map[string]*OSDInfo // OSD address -> info
	openVolumes     map[string]bool     // Volume ID -> is open
	replicationConn *grpc.ClientConn
	osdClients      map[string]osd.OSDServiceClient
	mu              sync.RWMutex
}

// NewMaster creates a new Master instance
func NewMaster(cfg *config.Config, cellID, replicationAddr string) (*Master, error) {
	replicationConn, err := grpc.NewClient(replicationAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to replication table: %w", err)
	}

	return &Master{
		config:          cfg,
		cellID:          cellID,
		osds:            make(map[string]*OSDInfo),
		openVolumes:     make(map[string]bool),
		replicationConn: replicationConn,
		osdClients:      make(map[string]osd.OSDServiceClient),
	}, nil
}

// RegisterOSD registers a new OSD
func (m *Master) RegisterOSD(ctx context.Context, req *master.RegisterOSDRequest) (*master.RegisterOSDResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	osdInfo := &OSDInfo{
		Address:        req.OsdAddress,
		CellID:         req.CellId,
		AvailableSpace: req.AvailableSpace,
		LastHeartbeat:  time.Now(),
		Healthy:        true,
	}

	m.osds[req.OsdAddress] = osdInfo

	// Create gRPC client for this OSD
	conn, err := grpc.NewClient(req.OsdAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &master.RegisterOSDResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to connect to OSD: %v", err),
		}, nil
	}
	m.osdClients[req.OsdAddress] = osd.NewOSDServiceClient(conn)

	return &master.RegisterOSDResponse{
		Success: true,
	}, nil
}

// Heartbeat processes OSD heartbeats
func (m *Master) Heartbeat(ctx context.Context, req *master.HeartbeatRequest) (*master.HeartbeatResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	osdInfo, exists := m.osds[req.OsdAddress]
	if !exists {
		// Auto-register if not exists
		osdInfo = &OSDInfo{
			Address:       req.OsdAddress,
			LastHeartbeat: time.Now(),
			Healthy:       req.Healthy,
		}
		m.osds[req.OsdAddress] = osdInfo
	} else {
		osdInfo.LastHeartbeat = time.Now()
		osdInfo.Healthy = req.Healthy
		osdInfo.AvailableSpace = req.AvailableSpace
	}

	return &master.HeartbeatResponse{
		Success: true,
	}, nil
}

// GetOpenVolumes returns list of open volumes
func (m *Master) GetOpenVolumes(ctx context.Context, req *master.GetOpenVolumesRequest) (*master.GetOpenVolumesResponse, error) {
	replicationClient := replication.NewReplicationTableServiceClient(m.replicationConn)

	listReq := &replication.ListVolumesRequest{
		CellId: req.CellId,
	}

	listResp, err := replicationClient.ListVolumes(ctx, listReq)
	if err != nil {
		return &master.GetOpenVolumesResponse{}, err
	}

	openVolumes := make([]string, 0)
	for _, volumeID := range listResp.VolumeIds {
		getReq := &replication.GetVolumeRequest{VolumeId: volumeID}
		getResp, err := replicationClient.GetVolume(ctx, getReq)
		if err != nil || !getResp.Found {
			continue
		}
		if getResp.State == "open" {
			openVolumes = append(openVolumes, volumeID)
		}
	}

	return &master.GetOpenVolumesResponse{
		VolumeIds: openVolumes,
	}, nil
}

// CloseVolume closes a volume
func (m *Master) CloseVolume(ctx context.Context, req *master.CloseVolumeRequest) (*master.CloseVolumeResponse, error) {
	replicationClient := replication.NewReplicationTableServiceClient(m.replicationConn)

	getReq := &replication.GetVolumeRequest{VolumeId: req.VolumeId}
	getResp, err := replicationClient.GetVolume(ctx, getReq)
	if err != nil {
		return &master.CloseVolumeResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	if !getResp.Found {
		return &master.CloseVolumeResponse{
			Success: false,
			Error:   "volume not found",
		}, nil
	}

	updateReq := &replication.UpdateVolumeRequest{
		VolumeId:     req.VolumeId,
		OsdAddresses: getResp.OsdAddresses,
		Generation:   getResp.Generation,
		State:        "closed",
	}

	_, err = replicationClient.UpdateVolume(ctx, updateReq)
	if err != nil {
		return &master.CloseVolumeResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	m.mu.Lock()
	delete(m.openVolumes, req.VolumeId)
	m.mu.Unlock()

	return &master.CloseVolumeResponse{
		Success: true,
	}, nil
}

// TriggerRepair triggers a repair operation for a failed OSD
func (m *Master) TriggerRepair(ctx context.Context, req *master.TriggerRepairRequest) (*master.TriggerRepairResponse, error) {
	return &master.TriggerRepairResponse{
		Success: true,
	}, nil
}

// MonitorOSDs periodically checks OSD health
func (m *Master) MonitorOSDs(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkOSDHealth(ctx)
		}
	}
}

// checkOSDHealth checks health of all registered OSDs
func (m *Master) checkOSDHealth(ctx context.Context) {
	m.mu.RLock()
	osdAddresses := make([]string, 0, len(m.osds))
	for addr := range m.osds {
		osdAddresses = append(osdAddresses, addr)
	}
	m.mu.RUnlock()

	for _, addr := range osdAddresses {
		m.mu.RLock()
		osdInfo := m.osds[addr]
		m.mu.RUnlock()

		if osdInfo == nil {
			continue
		}

		if time.Since(osdInfo.LastHeartbeat) > 2*time.Minute {
			osdInfo.Healthy = false
			go m.triggerRepairForOSD(ctx, addr)
		}
	}
}

// triggerRepairForOSD triggers repair for a specific OSD
func (m *Master) triggerRepairForOSD(ctx context.Context, osdAddress string) {
	fmt.Printf("Triggering repair for OSD: %s\n", osdAddress)
}

// GetHealthyOSDs returns list of healthy OSD addresses
func (m *Master) GetHealthyOSDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthy := make([]string, 0)
	for addr, info := range m.osds {
		if info.Healthy {
			healthy = append(healthy, addr)
		}
	}
	return healthy
}

// Close closes connections
func (m *Master) Close() error {
	return m.replicationConn.Close()
}
