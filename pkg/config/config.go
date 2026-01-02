package config

import (
	"os"
)

// Config holds the configuration for the storage system
type Config struct {
	MaxBlockSize       int64
	BucketSize         int64
	DataShards         int
	ParityShards       int
	ReplicationFactor  int
	FrontendPort       string
	OSDPort            string
	BlockIndexPort     string
	ReplicationPort    string
	MasterPort         string
	VolumeManagerPort  string
	OSDDataDir         string
	CellID             string
	ZoneID             string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxBlockSize:      4 * 1024 * 1024,
		BucketSize:        1 * 1024 * 1024 * 1024,
		DataShards:        10,
		ParityShards:      4,
		ReplicationFactor: 3,
		FrontendPort:      "8080",
		OSDPort:           "9090",
		BlockIndexPort:    "9091",
		ReplicationPort:   "9092",
		MasterPort:        "9093",
		VolumeManagerPort: "9094",
		OSDDataDir:        "./data",
		CellID:            getEnvOrDefault("CELL_ID", "cell1"),
		ZoneID:            getEnvOrDefault("ZONE_ID", "zone1"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

