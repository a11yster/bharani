#!/bin/bash

# Generate Go code from proto files

set -e

PROTO_DIR="./proto"
OUT_DIR="./proto"

# Ensure protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed. Please install Protocol Buffers compiler."
    echo "On macOS: brew install protobuf"
    exit 1
fi

# Generate code for each proto file
for proto_file in "$PROTO_DIR"/*.proto; do
    if [ -f "$proto_file" ]; then
        echo "Generating code for $proto_file..."
        # Extract package name from go_package option
        pkg_path=$(grep "go_package" "$proto_file" | sed 's/.*go_package = "bharani\/proto\/\([^"]*\)".*/\1/')
        
        protoc --go_out="$OUT_DIR" \
               --go_opt=paths=source_relative \
               --go-grpc_out="$OUT_DIR" \
               --go-grpc_opt=paths=source_relative \
               --go-grpc_opt=require_unimplemented_servers=false \
               "$proto_file"
    fi
done

# Move files from proto/proto/ to correct subdirectories
if [ -d "$OUT_DIR/proto" ]; then
    # Process all .go files in proto/proto/
    for go_file in "$OUT_DIR/proto"/*.go; do
        if [ -f "$go_file" ]; then
            # Extract package name from filename
            filename=$(basename "$go_file")
            if [[ "$filename" == *_grpc.pb.go ]]; then
                pkg_name=$(echo "$filename" | sed 's/_grpc\.pb\.go//')
            else
                pkg_name=$(echo "$filename" | sed 's/\.pb\.go//')
            fi
            
            # Find the corresponding proto file to get the correct package path
            proto_file="$PROTO_DIR/${pkg_name}.proto"
            if [ -f "$proto_file" ]; then
                pkg_path=$(grep "go_package" "$proto_file" | sed 's/.*go_package = "bharani\/proto\/\([^"]*\)".*/\1/')
                if [ -n "$pkg_path" ]; then
                    # Create directory if it doesn't exist
                    mkdir -p "$OUT_DIR/$pkg_path"
                    # Move the file
                    mv "$go_file" "$OUT_DIR/$pkg_path/" 2>/dev/null || true
                fi
            fi
        fi
    done
    
    # Clean up empty proto/proto/ directory
    rmdir "$OUT_DIR/proto" 2>/dev/null || rm -rf "$OUT_DIR/proto" 2>/dev/null || true
fi

echo "Proto code generation complete!"

