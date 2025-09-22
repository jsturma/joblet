#!/bin/bash
# Standalone proto generation script
# This script generates proto files from the external joblet-proto repository

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROTO_REPO="$PROJECT_ROOT/../joblet-proto"
PROTO_GEN_DIR="$PROJECT_ROOT/api/gen"

echo "📦 Generating proto files from joblet-proto repository..."

# Check if joblet-proto repository exists, clone if not
if [ ! -d "$PROTO_REPO" ]; then
    echo "🔄 Cloning joblet-proto repository..."
    git clone https://github.com/ehsaniara/joblet-proto.git "$PROTO_REPO"
fi

# Ensure we're on main branch with latest changes
echo "🔄 Ensuring we're on main branch with latest changes..."
cd "$PROTO_REPO" && git fetch && git checkout main && git reset --hard origin/main

# Check if generate.sh exists and is executable
if [ ! -f "$PROTO_REPO/generate.sh" ]; then
    echo "❌ Error: generate.sh not found in $PROTO_REPO"
    echo "This might indicate the repository is in an unexpected state."
    ls -la "$PROTO_REPO/"
    exit 1
fi

if [ ! -x "$PROTO_REPO/generate.sh" ]; then
    echo "Making generate.sh executable..."
    chmod +x "$PROTO_REPO/generate.sh"
fi

# Generate Go proto files
echo "Generating Go proto files..."
cd "$PROTO_REPO"
./generate.sh go

# Check if generated files exist
if [ ! -d "$PROTO_REPO/gen" ] || [ -z "$(ls -A "$PROTO_REPO/gen" 2>/dev/null)" ]; then
    echo "❌ Error: No generated files found in $PROTO_REPO/gen"
    exit 1
fi

# Copy generated files to project
echo "📋 Copying generated proto files to project..."
mkdir -p "$PROTO_GEN_DIR"
cp "$PROTO_REPO"/gen/*.pb.go "$PROTO_GEN_DIR/"

echo "✅ Proto files generated and copied successfully to $PROTO_GEN_DIR/"

# Proto version will be read directly from git during build

# List generated files
echo "Generated files:"
ls -la "$PROTO_GEN_DIR"/*.pb.go