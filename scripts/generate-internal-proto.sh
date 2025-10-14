#!/bin/bash
# Generate protocol buffer code for internal IPC and persist protos
# These are internal to the joblet monorepo

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROTO_DIR="${PROJECT_ROOT}/internal/proto"
GEN_DIR="${PROJECT_ROOT}/internal/proto/gen"

echo "Generating internal protobuf code..."

# Ensure we're in the project root
cd "${PROJECT_ROOT}"

# Check if protoc is available
if ! command -v protoc &> /dev/null; then
    echo "❌ Error: protoc is not installed or not in PATH"
    echo "Please install protobuf compiler: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Verify proto files exist
if [ ! -f "${PROTO_DIR}/ipc.proto" ]; then
    echo "❌ Error: ipc.proto not found at ${PROTO_DIR}/ipc.proto"
    exit 1
fi

# Clean and create output directory
echo "Cleaning output directory..."
rm -rf "${GEN_DIR}"
mkdir -p "${GEN_DIR}/ipc"

# Generate IPC proto only (persist is now external in joblet-proto/v2)
echo "Generating IPC protobuf code..."
protoc \
    --proto_path="${PROTO_DIR}" \
    --go_out="${GEN_DIR}/ipc" \
    --go_opt=paths=source_relative \
    "${PROTO_DIR}/ipc.proto"

# Verify generation succeeded
if [ ! -f "${GEN_DIR}/ipc/ipc.pb.go" ]; then
    echo "❌ Error: Proto generation failed - ipc.pb.go not found"
    exit 1
fi

echo "✅ Internal protocol buffer generation complete"
echo "Generated files:"
ls -la "${GEN_DIR}"/ipc/*.go
