# Joblet Platform Installation Guide

This comprehensive installation guide provides detailed procedures for deploying Joblet across diverse operating systems
and hardware architectures. The guide covers both server-side components for Linux systems and client-side tools for
cross-platform environments.

> **Important**: Joblet leverages native Linux kernel capabilities to deliver enterprise-grade performance, security,
> and resource management through namespaces and cgroups v2. Direct installation on Linux hosts ensures optimal
> performance characteristics with kernel-level process isolation.

## System Requirements

### Server Requirements (Linux Exclusive)

- **Operating System**: Linux distributions with kernel version 3.10 or higher
- **Processor Architecture**: x86_64 (amd64) or ARM64
- **Control Groups**: cgroups v2 recommended (v1 compatibility supported)
- **Access Requirements**: Root privileges or sudo access
- **Memory Requirements**: Minimum 512MB RAM (2GB recommended for production)
- **Storage Requirements**: Minimum 1GB available disk space

### Client Requirements (Cross-Platform)

- **Operating Systems**: Linux, macOS, Windows
- **Processor Architecture**: x86_64, ARM64, Apple Silicon (M1/M2/M3)
- **Network Connectivity**: TCP access to Joblet server (default port: 50051)

## Linux Platform Installation

### Ubuntu/Debian Installation (Version 20.04 and Later)

```bash
# Update package list
sudo apt update

# Install dependencies
sudo apt install -y curl tar make gcc

# Download and install
curl -L https://github.com/ehsaniara/joblet/releases/latest/download/joblet-linux-amd64.tar.gz | tar xz
sudo mv joblet /usr/local/bin/
sudo mv rnx /usr/local/bin/
sudo chmod +x /usr/local/bin/joblet /usr/local/bin/rnx

# Create directories
sudo mkdir -p /opt/joblet/{config,state,certs,jobs,volumes}
sudo mkdir -p /var/log/joblet

# Verify installation
joblet --version
rnx --version
```

### Red Hat Enterprise Linux/CentOS/Fedora Installation (Version 8 and Later)

```bash
# Install dependencies
sudo dnf install -y curl tar make gcc

# Download and install
curl -L https://github.com/ehsaniara/joblet/releases/latest/download/joblet-linux-amd64.tar.gz | tar xz
sudo mv joblet /usr/local/bin/
sudo mv rnx /usr/local/bin/
sudo chmod +x /usr/local/bin/joblet /usr/local/bin/rnx

# Create directories
sudo mkdir -p /opt/joblet/{config,state,certs,jobs,volumes}
sudo mkdir -p /var/log/joblet

# Enable cgroups v2 if needed
sudo grubby --update-kernel=ALL --args="systemd.unified_cgroup_hierarchy=1"
# Reboot required after this change
```

### Amazon Linux 2 Installation

```bash
# Install dependencies
sudo yum install -y curl tar make gcc

# Download and install
curl -L https://github.com/ehsaniara/joblet/releases/latest/download/joblet-linux-amd64.tar.gz | tar xz
sudo mv joblet /usr/local/bin/
sudo mv rnx /usr/local/bin/
sudo chmod +x /usr/local/bin/joblet /usr/local/bin/rnx

# Create directories
sudo mkdir -p /opt/joblet/{config,state,certs,jobs,volumes}
sudo mkdir -p /var/log/joblet
```

### Arch Linux Installation

```bash
# Install from AUR (if available)
yay -S joblet

# Or manual installation
sudo pacman -S curl tar make gcc
curl -L https://github.com/ehsaniara/joblet/releases/latest/download/joblet-linux-amd64.tar.gz | tar xz
sudo mv joblet /usr/local/bin/
sudo mv rnx /usr/local/bin/
```

### ARM64 Architecture Systems (Raspberry Pi, AWS Graviton)

```bash
# Download ARM64 version
curl -L https://github.com/ehsaniara/joblet/releases/latest/download/joblet-linux-arm64.tar.gz | tar xz
sudo mv joblet /usr/local/bin/
sudo mv rnx /usr/local/bin/
sudo chmod +x /usr/local/bin/joblet /usr/local/bin/rnx
```

## macOS Client Installation

### Installation via Homebrew Package Manager (Recommended)

```bash
# Add Joblet tap
brew tap ehsaniara/joblet

# Install RNX client
brew install rnx
```

### Manual Binary Installation

```bash
# Intel Macs
curl -L https://github.com/ehsaniara/joblet/releases/latest/download/rnx-darwin-amd64.tar.gz | tar xz

# Apple Silicon (M1/M2)
curl -L https://github.com/ehsaniara/joblet/releases/latest/download/rnx-darwin-arm64.tar.gz | tar xz

# Install
sudo mv rnx /usr/local/bin/
sudo chmod +x /usr/local/bin/rnx

# Create config directory
mkdir -p ~/.rnx
```

## Windows Client Installation

### Installation via Scoop Package Manager

```powershell
# Add Joblet bucket
scoop bucket add joblet https://github.com/ehsaniara/scoop-joblet

# Install RNX
scoop install rnx
```

### Manual Binary Installation

1. Download the Windows binary:
    - [rnx-windows-amd64.zip](https://github.com/ehsaniara/joblet/releases/latest/download/rnx-windows-amd64.zip)

2. Extract to a directory (e.g., `C:\Program Files\Joblet`)

3. Add to PATH:
   ```powershell
   [Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\Joblet", "User")
   ```

4. Create config directory:
   ```powershell
   mkdir $env:USERPROFILE\.rnx
   ```

## 🔨 Building from Source

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional but recommended)
- GCC (for CGO dependencies)

### Build Steps

```bash
# Clone repository
git clone https://github.com/ehsaniara/joblet.git
cd joblet

# Build all binaries
make build

# Or build manually
go build -o joblet ./cmd/joblet
go build -o rnx ./cmd/rnx

# Run tests
make test

# Install binaries
sudo make install
```

### Verify Installation

After installation, verify both client and server versions:

```bash
# Check RNX client version
rnx --version

# Output should show both client and server versions:
# RNX Client:
# rnx version v4.3.3 (abc1234)
# Built: 2025-09-14T05:17:17Z
# ...
#
# Joblet Server (default):
# joblet version v4.3.3 (abc1234)
# Built: 2025-09-14T05:18:24Z
# ...

# If server is not reachable, you'll see:
# Joblet Server: failed to connect to server: <error>

# Test basic functionality
rnx job list  # Should connect to server and list jobs
```

### Cross-compilation

```bash
# Build for Linux AMD64
GOOS=linux GOARCH=amd64 go build -o joblet-linux-amd64 ./cmd/joblet

# Build for macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o rnx-darwin-arm64 ./cmd/rnx

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o rnx.exe ./cmd/rnx
```

## 🔐 Certificate Generation

### Automatic Generation

```bash
# Set server address (REQUIRED)
export JOBLET_SERVER_ADDRESS='192.168.1.100'  # Use your server's IP

# Generate certificates with embedded configuration
sudo /usr/local/bin/certs_gen_embedded.sh
```

This creates:

- `/opt/joblet/config/joblet-config.yml` - Server config with embedded certificates
- `/opt/joblet/config/rnx-config.yml` - Client config with embedded certificates

### Manual Certificate Generation

```bash
# Create CA
openssl genrsa -out ca-key.pem 4096
openssl req -new -x509 -key ca-key.pem -out ca-cert.pem -days 3650 \
  -subj "/CN=Joblet CA"

# Create server certificate
openssl genrsa -out server-key.pem 4096
openssl req -new -key server-key.pem -out server.csr \
  -subj "/CN=joblet"
openssl x509 -req -in server.csr -CA ca-cert.pem -CAkey ca-key.pem \
  -out server-cert.pem -days 365 -CAcreateserial \
  -extensions v3_req -extfile <(echo "[v3_req]
subjectAltName = DNS:localhost,DNS:joblet,IP:127.0.0.1,IP:${JOBLET_SERVER_ADDRESS}")

# Create client certificate
openssl genrsa -out client-key.pem 4096
openssl req -new -key client-key.pem -out client.csr \
  -subj "/CN=rnx-client/OU=admin"
openssl x509 -req -in client.csr -CA ca-cert.pem -CAkey ca-key.pem \
  -out client-cert.pem -days 365 -CAcreateserial
```

## 🚀 Systemd Service Setup

### Create Service File

```bash
sudo tee /etc/systemd/system/joblet.service > /dev/null <<EOF
[Unit]
Description=Joblet Job Execution Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/joblet
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=joblet
Environment="JOBLET_CONFIG_PATH=/opt/joblet/config/joblet-config.yml"

# Security settings
NoNewPrivileges=false
PrivateTmp=false
ProtectSystem=false
ProtectHome=false

[Install]
WantedBy=multi-user.target
EOF
```

### Enable and Start Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable joblet

# Start service
sudo systemctl start joblet

# Check status
sudo systemctl status joblet

# View logs
sudo journalctl -u joblet -f
```

## 🖥️ Development Environment Setup

### Local Development

Joblet provides superior isolation, performance, and resource control through Linux namespaces and cgroups v2.

```bash
# Set up development environment on Linux
# Requires Linux host (VM, WSL2, or native Linux)

# Install development dependencies
sudo apt update
sudo apt install -y build-essential git protobuf-compiler

# Clone and build
git clone https://github.com/ehsaniara/joblet.git
cd joblet
make all

# Run tests
make test

# Install locally for development
sudo make install
```

### Native Process Isolation

Joblet provides native Linux process isolation with:

- **Better Performance**: Direct Linux namespace execution (no container overhead)
- **Superior Resource Control**: cgroups v2 with precise CPU, memory, and I/O limits
- **Enhanced Security**: Process isolation without container escape vulnerabilities
- **Simplified Deployment**: Single binary installation vs container orchestration complexity
- **Instant Startup**: 2-3 second job execution vs container pull/start overhead

**Joblet Commands:**

- `rnx job run` - Execute isolated processes
- `rnx job run --workflow=workflow.yaml` - Run complex workflows
- `rnx runtime install` - Install pre-built runtime environments

## ✅ Post-Installation Verification

### Server Health Check

```bash
# Check if server is running
sudo systemctl is-active joblet

# Test server locally
sudo joblet --version

# Check listening port
sudo ss -tlnp | grep 50051
```

### Client Connectivity Test

```bash
# Copy client config from server
scp server:/opt/joblet/config/rnx-config.yml ~/.rnx/

# Test connection
rnx job list

# Run test job
rnx job run echo "Installation successful!"
```

## 🔧 Troubleshooting Installation

### Common Issues

1. **Permission Denied**
   ```bash
   sudo chmod +x /usr/local/bin/joblet /usr/local/bin/rnx
   ```

2. **Cgroups v2 Not Available**
   ```bash
   # Check cgroups version
   mount | grep cgroup
   
   # Enable cgroups v2 (requires reboot)
   sudo grubby --update-kernel=ALL --args="systemd.unified_cgroup_hierarchy=1"
   ```

3. **Port Already in Use**
   ```bash
   # Find process using port
   sudo lsof -i :50051
   
   # Change port in config
   # Edit /opt/joblet/config/joblet-config.yml
   ```

4. **Certificate Issues**
   ```bash
   # Regenerate certificates
   sudo rm -rf /opt/joblet/config/*.yml
   sudo /usr/local/bin/certs_gen_embedded.sh
   ```

## 📚 Next Steps

- [Configuration Guide](./CONFIGURATION.md) - Customize your setup
- [Quick Start Guide](./QUICKSTART.md) - Start using Joblet
- [Security Guide](./SECURITY.md) - Secure your installation