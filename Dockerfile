FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install protoc and dependencies
RUN apk add --no-cache protobuf protobuf-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate proto files
RUN chmod +x scripts/generate-proto.sh && ./scripts/generate-proto.sh

# Build binaries
RUN go build -o bin/osd ./cmd/osd && \
    go build -o bin/blockindex ./cmd/blockindex && \
    go build -o bin/replication ./cmd/replication && \
    go build -o bin/master ./cmd/master && \
    go build -o bin/volumemanager ./cmd/volumemanager && \
    go build -o bin/frontend ./cmd/frontend

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin /app/bin

ENTRYPOINT ["/app/bin/frontend"]


