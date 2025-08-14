# RNX Installation Guide

Comprehensive installation guide for RNX CLI and Admin UI on macOS via Homebrew.

## 🍺 Quick Start

1. **Add the tap**:
   ```bash
   brew tap ehsaniara/joblet
   ```

2. **Install RNX**:
   ```bash
   brew install rnx
   ```

The installer will guide you through the process and auto-detect your system capabilities.

## 📋 Installation Scenarios

### Scenario 1: Node.js Detected + Interactive
```bash
$ brew install rnx
==> Installing rnx from ehsaniara/joblet
✅ Node.js detected: v20.10.0
🤔 Would you like to install the web admin UI? (Y/n): y
🔧 Setting up admin UI...
📁 Installing admin UI files...
📦 Installing admin server dependencies...
✅ Admin UI setup complete!
🍺 rnx was successfully installed!

Usage:
  CLI: rnx --help
  Web UI: rnx admin
```

### Scenario 2: No Node.js + User Wants Admin
```bash
$ brew install rnx
==> Installing rnx from ehsaniara/joblet
❌ Node.js not detected
🤔 Would you like to install Node.js and the web admin UI? (y/N): y
📦 Installing Node.js...
🔧 Setting up admin UI...
✅ Complete installation ready!
```

### Scenario 3: CLI-Only Installation
```bash
$ brew install rnx
❌ Node.js not detected
🤔 Would you like to install Node.js and the web admin UI? (y/N): n
📱 CLI-only installation complete.

💡 To install the web admin UI later, run:
   brew reinstall rnx --with-admin
```

## 🎛️ Non-Interactive Installation

### Force Admin Installation
```bash
brew install rnx --with-admin
```
- Installs Node.js if missing
- Sets up admin UI automatically
- No prompts

### Force CLI-Only Installation  
```bash
brew install rnx --without-admin
```
- Installs only the CLI binary
- Skips Node.js and admin UI
- No prompts

## 🔧 Post-Installation Setup

### 1. Configure RNX Client
```bash
# Create config directory
mkdir -p ~/.rnx

# Copy configuration from your Joblet server
scp your-server:/opt/joblet/config/rnx-config.yml ~/.rnx/
```

### 2. Test CLI Installation
```bash
# Check version
rnx --version

# Test help
rnx --help

# Test connection (requires server config)
rnx list
```

### 3. Test Admin UI (if installed)
```bash
# Start admin UI
rnx admin

# Or use the direct command
rnx-admin
```

## 📊 Installation Validation

### CLI Validation
```bash
# Version check
$ rnx --version
rnx version 1.0.0

# Help output
$ rnx --help
RNX - Remote job execution CLI for Joblet

Usage:
  rnx [command]

Available Commands:
  list        List jobs
  run         Execute a job
  log         Stream job logs
  stop        Stop a job
  admin       Launch admin UI
  
# Basic connectivity (requires config)
$ rnx list
ID          COMMAND         STATUS    CREATED
job-123     python test.py  running   2m ago
```

### Admin UI Validation
```bash
# Admin UI launcher
$ rnx admin
🚀 Starting RNX Admin UI...
📍 Admin server directory: /opt/homebrew/share/rnx/admin/server
🌐 Opening http://localhost:5173 in your browser...
✅ Admin UI is running!
🛑 Press Ctrl+C to stop

# Files should exist
$ ls -la $(brew --prefix)/share/rnx/admin/
drwxr-xr-x  server/
drwxr-xr-x  ui/

$ ls -la $(brew --prefix)/share/rnx/admin/server/
-rw-r--r--  package.json
drwxr-xr-x  node_modules/
-rw-r--r--  server.js

$ ls -la $(brew --prefix)/share/rnx/admin/ui/
drwxr-xr-x  dist/
-rw-r--r--  index.html
```

## 🔄 Managing Installation

### Upgrade to Latest Version
```bash
# Standard upgrade (preserves admin/CLI choice)
brew upgrade rnx

# Check for updates
brew outdated
```

### Switch Between Modes

#### CLI-Only → Full Installation
```bash
brew reinstall rnx --with-admin
```

#### Full Installation → CLI-Only
```bash
brew reinstall rnx --without-admin
```

#### Reset to Interactive Mode
```bash
brew uninstall rnx
brew install rnx
```

## 🛠️ Troubleshooting

### Installation Failures

#### Node.js Installation Failed
```bash
# Error message
❌ Failed to install admin server dependencies
⚠️  Admin UI setup failed: npm install failed
📱 Continuing with CLI-only installation...

# Solutions
1. Install Node.js manually: brew install node
2. Reinstall with admin: brew reinstall rnx --with-admin
3. Check npm registry: npm config get registry
```

#### Admin UI Build Missing
```bash
# Error message  
❌ Admin UI build not found
⚠️  Admin UI setup failed: missing dist directory
📱 Continuing with CLI-only installation...

# Solutions
1. Check archive contents: tar -tzf downloaded-archive.tar.gz
2. Report issue to: https://github.com/ehsaniara/joblet/issues  
3. Use CLI-only: brew reinstall rnx --without-admin
```

#### Permission Issues
```bash
# Error message
❌ Permission denied: cannot create directory

# Solutions
1. Fix Homebrew permissions: sudo chown -R $(whoami) $(brew --prefix)/*
2. Reinstall Homebrew: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
3. Use sudo (not recommended): sudo brew install rnx
```

### Runtime Issues

#### Command Not Found
```bash
# Error
$ rnx
command not found: rnx

# Solutions  
1. Reload shell: source ~/.zshrc  # or ~/.bashrc
2. Check PATH: echo $PATH
3. Reinstall: brew reinstall rnx
4. Link manually: brew link rnx
```

#### Admin UI Won't Start
```bash
# Error
$ rnx admin
❌ Admin UI not installed. Install with: brew reinstall rnx --with-admin

# Solution
brew reinstall rnx --with-admin
```

#### Node.js Version Issues
```bash
# Check Node.js version
$ node --version
v18.17.0  # Should be 18+ for admin UI

# Upgrade Node.js if needed
brew upgrade node
```

### Configuration Issues

#### Missing Config File
```bash
# Error
❌ Configuration file not found: ~/.rnx/rnx-config.yml

# Solutions
1. Copy from server: scp server:/opt/joblet/config/rnx-config.yml ~/.rnx/
2. Generate template: rnx config init
3. Check docs: https://github.com/ehsaniara/joblet/docs/CONFIGURATION.md
```

#### Connection Refused
```bash
# Error  
❌ Failed to connect to server: connection refused

# Solutions
1. Check server status: ssh server "sudo systemctl status joblet"
2. Verify config: cat ~/.rnx/rnx-config.yml
3. Test network: ping your-joblet-server
4. Check certificates: rnx config verify
```

## 📋 System Requirements

### Minimum Requirements
- **macOS**: 10.15 (Catalina) or later
- **Architecture**: Intel (x86_64) or Apple Silicon (ARM64)  
- **Disk space**: 50MB (CLI only), 200MB (with admin UI)
- **Memory**: 50MB (CLI), 150MB (admin UI)

### Recommended Requirements
- **macOS**: 12.0 (Monterey) or later
- **Node.js**: 18.x or later (for admin UI)
- **Browser**: Chrome, Firefox, Safari, Edge (for admin UI)
- **Network**: Stable connection to Joblet server

## 🎯 Installation Best Practices

### 1. Start with Interactive Mode
```bash
brew install rnx
```
Let the installer guide you based on your system setup.

### 2. Test CLI First
```bash
rnx --version
rnx --help
```
Ensure basic functionality before using admin UI.

### 3. Configure Before Use
```bash
# Setup config directory
mkdir -p ~/.rnx

# Copy or create config file
scp server:/opt/joblet/config/rnx-config.yml ~/.rnx/

# Test connectivity  
rnx list
```

### 4. Use Admin UI for Complex Tasks
- Job monitoring and management
- File uploads via web interface
- Real-time log viewing
- Visual job building

### 5. Keep Updated
```bash
# Check for updates regularly
brew update && brew upgrade rnx
```

## 🔗 Next Steps

1. **Server Setup**: Install Joblet on your Linux server
2. **Configuration**: Setup mTLS certificates and client config
3. **First Job**: Run your first job with `rnx run "echo hello"`
4. **Explore Features**: Try file uploads, resource limits, and monitoring

For detailed server setup, see: [Joblet Documentation](https://github.com/ehsaniara/joblet/docs)