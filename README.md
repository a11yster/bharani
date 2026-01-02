# Bharani - Magic Pocket Clone

A distributed immutable block storage system inspired by Dropbox's Magic Pocket, built in Go.

## Architecture

Bharani implements a multi-component distributed storage system:

- **Frontend**: Client-facing API for Put/Get operations
- **Block Index**: Maps block hashes to storage locations (cell, bucket)
- **Replication Table**: Maps volumes to OSD addresses
- **Master**: Coordinates repairs, volume management, and OSD monitoring
- **OSD (Object Storage Daemon)**: Stores blocks on disk
- **Volume Manager**: Handles data transfer and erasure coding

## Features

- Immutable block storage (up to 4MB per block)
- Replication for durability (configurable replication factor)
- Reed-Solomon erasure coding (10 data + 4 parity shards)
- Distributed architecture with multiple OSDs
- Automatic repair on OSD failure
- SQLite-based metadata storage

## Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`)
  - macOS: `brew install protobuf`
  - Linux: `apt-get install protobuf-compiler`
  - Or download from https://grpc.io/docs/protoc-installation/

## Building

**Important**: You must have `protoc` (Protocol Buffers compiler) installed before building.

1. Install dependencies:

```bash
make deps
```

2. Generate proto files (requires `protoc`):

```bash
make proto
```

3. Build all services:

```bash
make build
```

This will create binaries in the `bin/` directory.

**Note**: If you see import errors for `bharani/proto/*` packages, you need to run `make proto` first to generate the Go code from the `.proto` files.

## Running Locally

### Option 1: Using Makefile (separate terminals)

Terminal 1 - BlockIndex:

```bash
make run-blockindex
```

Terminal 2 - ReplicationTable:

```bash
make run-replication
```

Terminal 3 - Master:

```bash
make run-master
```

Terminal 4 - OSD (run multiple instances):

```bash
make run-osd
# Or manually:
go run ./cmd/osd -port 9090 -address localhost:9090 -cell cell1 -data-dir ./data/osd1
go run ./cmd/osd -port 9095 -address localhost:9095 -cell cell1 -data-dir ./data/osd2
go run ./cmd/osd -port 9096 -address localhost:9096 -cell cell1 -data-dir ./data/osd3
```

Terminal 5 - Frontend:

```bash
make run-frontend
```

### Option 2: Using Docker Compose

```bash
docker-compose up --build
```

This will start all services in containers.

## Usage

### Using gRPC Client

The Frontend service exposes a gRPC API on port 8080. You can use any gRPC client to interact with it.

Example using `grpcurl`:

```bash
# Put a block
echo -n "Hello, World!" | grpcurl -plaintext -d @ localhost:8080 frontend.FrontendService/Put

# Get a block (replace HASH with actual hash from Put response)
grpcurl -plaintext -d '{"hash": "HASH"}' localhost:8080 frontend.FrontendService/Get
```

## Configuration

Configuration can be set via environment variables or modified in `pkg/config/config.go`:

- `CELL_ID`: Cell identifier (default: "cell1")
- `ZONE_ID`: Zone identifier (default: "zone1")
- `MAX_BLOCK_SIZE`: Maximum block size in bytes (default: 4MB)
- `BUCKET_SIZE`: Bucket size in bytes (default: 1GB)
- `DATA_SHARDS`: Number of data shards for erasure coding (default: 10)
- `PARITY_SHARDS`: Number of parity shards (default: 4)
- `REPLICATION_FACTOR`: Number of replicas (default: 3)

## Testing

Run tests:

```bash
make test
```

## License

MIT
