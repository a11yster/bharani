.PHONY: proto build test clean run-osd run-blockindex run-replication run-master run-frontend

# Generate proto files
proto:
	@chmod +x scripts/generate-proto.sh
	@./scripts/generate-proto.sh

# Build all binaries
build: proto
	@echo "Building all services..."
	@go build -o bin/osd ./cmd/osd
	@go build -o bin/blockindex ./cmd/blockindex
	@go build -o bin/replication ./cmd/replication
	@go build -o bin/master ./cmd/master
	@go build -o bin/volumemanager ./cmd/volumemanager
	@go build -o bin/frontend ./cmd/frontend
	@echo "Build complete!"

# Run tests
test:
	@go test ./...

# Clean build artifacts
clean:
	@rm -rf bin/
	@rm -rf data/
	@echo "Clean complete!"

# Run services (for development)
run-osd:
	@go run ./cmd/osd -port 9090 -address localhost:9090 -cell cell1 -data-dir ./data/osd1

run-blockindex:
	@go run ./cmd/blockindex -port 9091 -db ./data/blockindex.db

run-replication:
	@go run ./cmd/replication -port 9092 -db ./data/replication.db

run-master:
	@go run ./cmd/master -port 9093 -cell cell1 -replication localhost:9092

run-frontend:
	@go run ./cmd/frontend -port 8080 -blockindex localhost:9091 -replication localhost:9092 -master localhost:9093

# Install dependencies
deps:
	@go mod download
	@go mod tidy


