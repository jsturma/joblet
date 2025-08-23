#!/bin/bash
# OpenJDK 21 Runtime Setup - Platform Detection and Delegation
set -e

echo "Starting OpenJDK 21 runtime setup..."
echo "Build ID: ${BUILD_ID:-unknown}"
echo "Runtime Spec: ${RUNTIME_SPEC:-unknown}"  
echo "Chroot: ${JOBLET_CHROOT:-false}"

# Detect platform and architecture
detect_platform() {
    if [ -f /etc/os-release ]; then
        source /etc/os-release
        case "$ID" in
            ubuntu|debian)
                echo "ubuntu"
                ;;
            centos|rhel|rocky|almalinux)
                echo "rhel"
                ;;
            amzn)
                echo "amzn"
                ;;
            *)
                echo "ubuntu"  # Default fallback
                ;;
        esac
    else
        echo "ubuntu"  # Default fallback
    fi
}

detect_architecture() {
    case "$(uname -m)" in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            echo "amd64"  # Default fallback
            ;;
    esac
}

PLATFORM=$(detect_platform)
ARCHITECTURE=$(detect_architecture)

echo "Detected platform: $PLATFORM"
echo "Detected architecture: $ARCHITECTURE"

# Delegate to platform-specific setup script
PLATFORM_SCRIPT="setup-${PLATFORM}-${ARCHITECTURE}.sh"

if [ -f "$PLATFORM_SCRIPT" ]; then
    echo "Delegating to platform-specific script: $PLATFORM_SCRIPT"
    chmod +x "$PLATFORM_SCRIPT"
    exec "./$PLATFORM_SCRIPT"
else
    echo "ERROR: Platform-specific script $PLATFORM_SCRIPT not found"
    echo "Available platform combinations:"
    echo "  - ubuntu-amd64, ubuntu-arm64"
    echo "  - rhel-amd64, rhel-arm64" 
    echo "  - amzn-amd64, amzn-arm64"
    echo "Detected: $PLATFORM-$ARCHITECTURE"
    exit 1
fi