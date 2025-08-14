# Homebrew Joblet

Official Homebrew tap for [Joblet](https://github.com/ehsaniara/joblet) - the distributed job execution system.

## 🍺 Quick Installation

### Interactive Installation (Recommended)

```bash
brew tap ehsaniara/joblet
brew install rnx
```

The installer will auto-detect Node.js and ask if you want to install the web admin UI.

### Installation Options

#### 🖥️ CLI Only

```bash
brew install ehsaniara/joblet/rnx --without-admin
```

#### 🌐 Full Installation (CLI + Admin UI)

```bash
brew install ehsaniara/joblet/rnx --with-admin
```

## 🚀 What You Get

### RNX CLI

- **Cross-platform job execution** on remote Joblet servers
- **Resource management** (CPU, memory, disk quotas)
- **Real-time log streaming** from remote jobs
- **File upload/download** capabilities
- **Job lifecycle management** (create, monitor, stop)

### Admin UI (Optional)

- **Web dashboard** at `http://localhost:5173`
- **Visual job management** interface
- **Real-time metrics** and monitoring
- **Interactive job builder** with templates
- **WebSocket log streaming** with auto-scroll
- **File management** through web interface

## 🔧 Usage

### CLI Commands

```bash
# Connect and manage jobs
rnx list                           # List all jobs
rnx run "python script.py"         # Execute job on remote server
rnx log <job-id>                   # Stream logs from job
rnx stop <job-id>                  # Stop running job

# Resource limits
rnx run --max-cpu=50 --max-memory=512 "intensive-task.py"
rnx run --max-disk=1GB --volume=data "data-processing.py"

# File operations
rnx upload --file=script.py        # Upload file to remote workspace
rnx run --upload=data.csv "process-data.py"  # Upload and run
```

### Admin UI

```bash
rnx admin     # Launch web interface
```

## 🔧 Configuration

### Initial Setup

1. **Install on macOS** (this tap):
   ```bash
   brew install ehsaniara/joblet/rnx
   ```

2. **Setup Joblet server** on your Linux server:
   ```bash
   # On your Linux server
   wget https://github.com/ehsaniara/joblet/releases/latest/download/joblet_1.0.0_amd64.deb
   sudo dpkg -i joblet_1.0.0_amd64.deb
   sudo systemctl start joblet
   ```

3. **Copy client configuration**:
   ```bash
   scp your-server:/opt/joblet/config/rnx-config.yml ~/.rnx/
   ```

### Configuration File Location

- **Config file**: `~/.rnx/rnx-config.yml`
- **Logs**: `~/.rnx/logs/`

## 🏗️ Architecture

```
┌─────────────────────────────────────┐    gRPC/mTLS     ┌─────────────────────────────────────┐
│           macOS Client              │ ──────────────── │         Linux Server               │
│                                     │                  │                                     │
│  ┌─────────────────────────────────┐│                  │  ┌─────────────────────────────────┐│
│  │         rnx CLI                 ││                  │  │        joblet server            ││
│  │      (Go binary)                ││                  │  │       (Go binary)               ││
│  │   - Job management              ││                  │  │   - Job execution               ││
│  │   - Log streaming               ││                  │  │   - Resource isolation          ││
│  │   - File uploads                ││                  │  │   - Network management          ││
│  └─────────────────────────────────┘│                  │  └─────────────────────────────────┘│
│                                     │                  │                                     │
│  ┌─────────────────────────────────┐│                  │                                     │
│  │       rnx admin                 ││                  │                                     │
│  │    (Node.js + React)            ││                  │                                     │
│  │   - Local web server            ││                  │                                     │
│  │   - React dashboard             ││                  │                                     │
│  │   - WebSocket log streaming     ││                  │                                     │
│  │   - Spawns local rnx commands   ││                  │                                     │
│  └─────────────────────────────────┘│                  │                                     │
└─────────────────────────────────────┘                  └─────────────────────────────────────┘
         ↑
    User Browser
    (localhost:5173)
```

## 📊 Installation Decision Matrix

| Node.js Present? | User Choice | Installation Result    |
|------------------|-------------|------------------------|
| ✅ Yes            | Y (default) | Full Installation      |
| ✅ Yes            | N           | CLI Only               |
| ❌ No             | Y           | Install Node.js + Full |
| ❌ No             | N (default) | CLI Only               |

## 🔄 Upgrading

### Update to Latest Version

```bash
brew upgrade rnx
```

### Switch Installation Types

```bash
# Switch from CLI-only to full
brew reinstall rnx --with-admin

# Switch from full to CLI-only  
brew reinstall rnx --without-admin
```

## 🛠️ Troubleshooting

### Formula Issues

```bash
# Reinstall with verbose output
brew reinstall --verbose rnx

# Force rebuild from source
brew install --build-from-source rnx

# Check formula syntax
brew audit rnx
```

### Admin UI Issues

```bash
# Verify Node.js installation
node --version
npm --version

# Check admin UI files
ls -la $(brew --prefix)/share/rnx/admin/

# Manually start admin UI
cd $(brew --prefix)/share/rnx/admin/server
node server.js
```

### Connection Issues

```bash
# Test connectivity to joblet server
rnx list -v  # Verbose output

# Check configuration
cat ~/.rnx/rnx-config.yml

# View logs
tail -f ~/.rnx/logs/rnx.log
```

## 📋 System Requirements

### macOS

- **macOS**: 10.15 (Catalina) or later
- **Architecture**: Intel (x86_64) or Apple Silicon (ARM64)

### Optional (for Admin UI)

- **Node.js**: 18.x or later (auto-installed if chosen)
- **Browser**: Modern browser for web interface

### Remote Server

- **Linux server** with Joblet installed
- **Network**: Accessible from macOS client
- **Certificates**: mTLS certificates configured

## 🎯 Use Cases

### Individual Developers

```bash
# Quick job execution
rnx run "python train-model.py"

# Monitor long-running jobs
rnx log job-123 --follow
```

### Data Scientists

```bash
# Resource-intensive workloads
rnx run --max-cpu=80 --max-memory=8192 "process-large-dataset.py"

# Admin UI for visual monitoring
rnx admin
```

### DevOps Teams

```bash
# Automated deployments
rnx run --upload=app.tar.gz "deploy-app.sh"

# Infrastructure monitoring
rnx admin  # Visual dashboard
```

## 🔗 Links

- **Main Repository**: https://github.com/ehsaniara/joblet
- **Documentation**: https://github.com/ehsaniara/joblet/tree/main/docs
- **Issues**: https://github.com/ehsaniara/joblet/issues
- **Releases**: https://github.com/ehsaniara/joblet/releases

## 📄 License

MIT License - see the [LICENSE](https://github.com/ehsaniara/joblet/blob/main/LICENSE) file for details.