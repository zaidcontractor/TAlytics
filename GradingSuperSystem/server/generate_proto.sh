#!/bin/bash

# Install protoc and required plugins if not already installed
echo "Setting up protobuf generation..."

# Create the proto output directory
mkdir -p proto

# Generate Go protobuf files
echo "Generating Go protobuf files..."
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
       proto/talytics.proto

echo "Protobuf files generated successfully!"