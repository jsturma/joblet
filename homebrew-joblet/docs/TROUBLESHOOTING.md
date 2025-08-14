# RNX Homebrew Troubleshooting Guide

Comprehensive troubleshooting guide for RNX Homebrew installation and runtime issues.

## 🚨 Installation Issues

### Formula Installation Failures

#### 1. Formula Not Found
```bash
# Error
Error: No available formula or cask with name "rnx"

# Solutions
brew tap ehsaniara/joblet           # Add the tap first
brew update                         # Update tap definitions
brew search rnx                     # Verify formula exists
```

#### 2. Checksum Mismatch
```bash
# Error  
Error: SHA256 mismatch
Expected: abc123...
Actual: def456...

# Solutions
brew cleanup --prune=0             # Clear cache
brew untap ehsaniara/joblet        # Remove tap
brew tap ehsaniara/joblet          # Re-add tap  
brew install rnx                   # Retry installation
```

#### 3. Download Failures
```bash
# Error
Error: Failed to download resource "rnx"

# Solutions
curl -I https://github.com/ehsaniara/joblet/releases/latest  # Check GitHub status
brew install --verbose rnx         # See detailed error
brew install --force-bottle rnx    # Force bottle installation
```

#### 4. Build from Source Failures
```bash
# Error
Error: Failed executing: go build

# Solutions  
brew install go                     # Ensure Go is installed
export GOPROXY=direct              # Bypass proxy issues
brew install --build-from-source --verbose rnx
```

### Node.js Related Issues

#### 1. Node.js Installation Failed During Setup
```bash
# Error
❌ Failed to install Node.js
⚠️  Admin UI setup failed

# Solutions
brew install node                   # Install Node.js manually
node --version                     # Verify installation
brew reinstall rnx --with-admin   # Retry with admin UI
```

#### 2. npm Install Failures
```bash
# Error
❌ Failed to install admin server dependencies
npm ERR! code EACCES

# Solutions
sudo chown -R $(whoami) ~/.npm     # Fix npm permissions
npm config set registry https://registry.npmjs.org/  # Reset registry
brew reinstall rnx --with-admin   # Retry installation
```

#### 3. Node.js Version Conflicts
```bash
# Error
⚠️  Warning: Node.js version 16.x detected. Admin UI requires 18+

# Solutions
brew upgrade node                   # Upgrade Node.js
nvm install 18                     # Use nvm if available
brew install node@18              # Install specific version
```

### Permission Issues

#### 1. Homebrew Permission Errors
```bash
# Error
Error: Permission denied @ rb_sysopen

# Solutions
sudo chown -R $(whoami) $(brew --prefix)/*  # Fix ownership
sudo chmod -R 755 $(brew --prefix)         # Fix permissions
brew doctor                                 # Diagnose issues
```

#### 2. Admin Directory Creation Failed
```bash
# Error  
❌ Permission denied: cannot create admin directory

# Solutions
ls -la $(brew --prefix)/share/             # Check permissions
sudo chown -R $(whoami) $(brew --prefix)/share/
brew reinstall rnx --with-admin
```

## 🔧 Runtime Issues

### Command Not Found

#### 1. rnx Command Missing
```bash
# Error
$ rnx
zsh: command not found: rnx

# Diagnosis
which rnx                          # Should show path
echo $PATH | grep $(brew --prefix)/bin  # Check PATH
ls -la $(brew --prefix)/bin/rnx   # Verify binary exists

# Solutions
source ~/.zshrc                    # Reload shell config
brew link rnx                     # Link binary
brew reinstall rnx               # Reinstall completely
```

#### 2. rnx-admin Command Missing
```bash
# Error
$ rnx-admin
command not found: rnx-admin

# Diagnosis  
ls -la $(brew --prefix)/bin/ | grep rnx    # Check binaries
brew list rnx                              # Show installed files

# Solutions
brew reinstall rnx --with-admin   # Ensure admin installation
brew link rnx                     # Re-link binaries
```

### Admin UI Issues

#### 1. Admin UI Won't Start
```bash
# Error
$ rnx admin
❌ Admin UI not installed. Install with: brew reinstall rnx --with-admin

# Diagnosis
ls -la $(brew --prefix)/share/rnx/admin/   # Check admin files
brew list rnx | grep admin                 # Verify admin files

# Solutions
brew reinstall rnx --with-admin   # Install admin UI
```

#### 2. Node.js Server Crashes
```bash
# Error
🚀 Starting RNX Admin UI...
Error: Cannot find module 'express'

# Diagnosis  
cd $(brew --prefix)/share/rnx/admin/server
ls -la node_modules/               # Check dependencies
cat package.json                   # Verify dependencies

# Solutions
cd $(brew --prefix)/share/rnx/admin/server
npm install                        # Reinstall dependencies
brew reinstall rnx --with-admin   # Full reinstall
```

#### 3. Port Already in Use
```bash
# Error
❌ Error: listen EADDRINUSE: address already in use :::5173

# Solutions
lsof -ti:5173                      # Find process using port
kill -9 $(lsof -ti:5173)         # Kill process
rnx admin                         # Retry
```

#### 4. Browser Won't Open  
```bash
# Error
⚠️  Could not open browser automatically

# Solutions
open http://localhost:5173         # Open manually
export BROWSER=firefox            # Set browser preference  
rnx admin                         # Retry
```

### Configuration Issues

#### 1. Missing Configuration File
```bash
# Error
❌ Configuration file not found: ~/.rnx/rnx-config.yml

# Solutions
mkdir -p ~/.rnx                              # Create config dir
scp server:/opt/joblet/config/rnx-config.yml ~/.rnx/  # Copy from server
rnx config init                             # Generate template (if available)
```

#### 2. Invalid Configuration Format
```bash
# Error
❌ Failed to parse configuration file: yaml: unmarshal errors

# Solutions
cat ~/.rnx/rnx-config.yml          # Check file contents
yamllint ~/.rnx/rnx-config.yml    # Validate YAML syntax  
cp ~/.rnx/rnx-config.yml{,.backup} # Backup current
# Download fresh config from server
```

#### 3. Connection Refused
```bash
# Error
❌ Failed to connect to server: dial tcp: connection refused

# Diagnosis
ping your-joblet-server            # Test network connectivity
telnet your-joblet-server 8443    # Test port access
ssh server "sudo systemctl status joblet"  # Check server status

# Solutions
# On server:
sudo systemctl start joblet        # Start server
sudo systemctl status joblet       # Check status
sudo journalctl -u joblet -f      # View logs

# On client:
cat ~/.rnx/rnx-config.yml         # Verify server address
rnx list -v                       # Verbose connection attempt
```

### Certificate Issues

#### 1. TLS Certificate Errors
```bash
# Error
❌ x509: certificate signed by unknown authority

# Solutions  
# Check certificate files
ls -la ~/.rnx/certs/

# Copy certificates from server
scp -r server:/opt/joblet/certs/ ~/.rnx/

# Verify certificate validity
openssl x509 -in ~/.rnx/certs/client.crt -text -noout
```

#### 2. Certificate Expired
```bash
# Error
❌ x509: certificate has expired

# Solutions
# Check expiration date
openssl x509 -in ~/.rnx/certs/client.crt -noout -dates

# Regenerate certificates on server
ssh server "sudo /opt/joblet/scripts/regenerate-certs.sh"

# Copy new certificates
scp -r server:/opt/joblet/certs/ ~/.rnx/
```

## 🔍 Diagnostic Commands

### System Information
```bash
# macOS version
sw_vers

# Architecture
uname -m                           # arm64 or x86_64

# Homebrew version
brew --version

# Homebrew prefix
brew --prefix

# Shell information
echo $SHELL
echo $PATH
```

### RNX Installation Status
```bash
# Check installation
brew list | grep rnx

# Show installed files
brew list rnx

# Check binary location
which rnx

# Test basic functionality
rnx --version
rnx --help
```

### Admin UI Status
```bash
# Check admin files
ls -la $(brew --prefix)/share/rnx/admin/

# Check Node.js
node --version
npm --version  

# Check admin dependencies
cd $(brew --prefix)/share/rnx/admin/server && npm list

# Test admin UI files
cd $(brew --prefix)/share/rnx/admin/ui && ls -la dist/
```

### Network Connectivity
```bash
# Test server connectivity
ping your-joblet-server

# Test specific port
telnet your-joblet-server 8443

# Test TLS connection
openssl s_client -connect your-joblet-server:8443 -cert ~/.rnx/certs/client.crt -key ~/.rnx/certs/client.key

# DNS resolution
nslookup your-joblet-server
```

## 🧹 Clean Installation

### Complete Removal
```bash
# Remove RNX
brew uninstall rnx

# Remove tap
brew untap ehsaniara/joblet

# Clean cache
brew cleanup --prune=0

# Remove config (optional)
rm -rf ~/.rnx/

# Remove Node.js if only used for RNX
brew uninstall node
```

### Fresh Installation
```bash
# Update Homebrew
brew update

# Add tap
brew tap ehsaniara/joblet

# Install with preferred option
brew install rnx                  # Interactive
brew install rnx --with-admin     # Force admin
brew install rnx --without-admin  # Force CLI-only
```

## 🐛 Reporting Issues

### Before Reporting
1. **Try diagnostic commands** above
2. **Check logs**: `~/.rnx/logs/`
3. **Test with verbose output**: `brew install --verbose rnx`
4. **Search existing issues**: https://github.com/ehsaniara/joblet/issues

### Information to Include
```bash
# System info
sw_vers && uname -m

# Homebrew info
brew --version && brew --prefix

# RNX info (if installed)
rnx --version 2>/dev/null || echo "RNX not installed"

# Node.js info (if admin UI issue)
node --version 2>/dev/null || echo "Node.js not installed"

# Installation attempt with verbose output
brew install --verbose rnx 2>&1 | tee install-log.txt
```

### Issue Templates

#### Installation Issue
```
**System**: macOS [version] on [Intel/Apple Silicon]
**Homebrew**: [version]
**Error**: [exact error message]
**Command**: [exact command that failed]
**Expected**: [what should have happened]
**Logs**: [relevant log excerpts]
```

#### Runtime Issue  
```
**Installation**: CLI-only / With Admin UI
**Command**: [command that failed]
**Error**: [exact error message]
**Config**: [relevant config file contents]
**Server**: [server status and version]
**Network**: [connectivity test results]
```

## 📞 Getting Help

- **GitHub Issues**: https://github.com/ehsaniara/joblet/issues
- **Documentation**: https://github.com/ehsaniara/joblet/docs
- **Formula Repository**: https://github.com/ehsaniara/homebrew-joblet

## 🔧 Advanced Debugging

### Enable Debug Logging
```bash
# For RNX CLI
export RNX_LOG_LEVEL=debug
rnx list -v

# For admin UI
cd $(brew --prefix)/share/rnx/admin/server
DEBUG=* node server.js
```

### Manual Formula Testing
```bash
# Test formula syntax
brew audit --strict Formula/rnx.rb

# Test installation locally
brew install --build-from-source ./Formula/rnx.rb

# Test with different options
brew install --build-from-source ./Formula/rnx.rb --with-admin
```

### Network Debugging
```bash
# Trace network calls
sudo tcpdump -i any host your-joblet-server

# Test with curl
curl -v --cert ~/.rnx/certs/client.crt --key ~/.rnx/certs/client.key https://your-joblet-server:8443/health
```