#!/bin/bash

# Create output directory if it doesn't exist
mkdir -p pkg/proto

# Generate gRPC code
protoc --go_out=pkg/proto --go_opt=paths=source_relative \
    --go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative \
    pkg/proto/auth.proto 