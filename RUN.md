# How to Run Bharani

## Quick Start (3 Steps)

### Step 1: Generate Proto Files
```bash
make proto
```

This generates the gRPC code from `.proto` files. You only need to do this once (or after changing proto files).

### Step 2: Build All Services
```bash
make build
```

This compiles all services into the `bin/` directory.

### Step 3: Run the System

**Option A: Docker Compose (Recommended - Easiest)**
```bash
docker-compose up --build
```

This starts all services automatically. Wait for all containers to be ready.

**Option B: Manual (5 Terminals)**

Open 5 terminal windows and run:

```bash
# Terminal 1: BlockIndex
./bin/blockindex -port 9091 -db ./data/blockindex.db

# Terminal 2: ReplicationTable  
./bin/replication -port 9092 -db ./data/replication.db

# Terminal 3: Master
./bin/master -port 9093 -cell cell1 -replication localhost:9092

# Terminal 4: OSD 1
./bin/osd -port 9090 -address localhost:9090 -cell cell1 -data-dir ./data/osd1

# Terminal 5: OSD 2 (open another terminal)
./bin/osd -port 9095 -address localhost:9095 -cell cell1 -data-dir ./data/osd2

# Terminal 6: OSD 3 (open another terminal)  
./bin/osd -port 9096 -address localhost:9096 -cell cell1 -data-dir ./data/osd3

# Terminal 7: Frontend (open another terminal)
./bin/frontend -port 8080 -blockindex localhost:9091 -replication localhost:9092 -master localhost:9093
```

## Test It

Once all services are running, test with:

```bash
go run ./examples/client.go
```

Or use `grpcurl`:
```bash
# Install grpcurl: brew install grpcurl (macOS) or go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Put a block
echo -n "Hello, World!" | grpcurl -plaintext -d @ localhost:8080 frontend.FrontendService/Put

# Get a block (use hash from Put response)
grpcurl -plaintext -d '{"hash": "YOUR_HASH"}' localhost:8080 frontend.FrontendService/Get
```

## Clean Up

```bash
make clean 
```


