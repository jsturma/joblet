# RNX Homebrew Maintainer Guide

Comprehensive guide for maintaining the RNX Homebrew tap and formula.

## 🏗️ Repository Setup

### Initial Setup

1. **Create the homebrew repository**:
   ```bash
   cd /path/to/joblet/homebrew-joblet
   ./scripts/init-repository.sh
   ```

2. **Configure GitHub secrets**:
   - Go to repository settings → Secrets and variables → Actions
   - Add `HOMEBREW_UPDATE_TOKEN`: GitHub token with `repo` scope for both repositories

3. **Configure main repository**:
   - Ensure `HOMEBREW_UPDATE_TOKEN` is set in ehsaniara/joblet repository secrets
   - The quick-release.yml workflow will trigger homebrew updates automatically

### Directory Structure

```
homebrew-joblet/
├── Formula/
│   └── rnx.rb                     # Main homebrew formula
├── .github/
│   ├── workflows/
│   │   └── update.yml             # Auto-update workflow
│   └── ISSUE_TEMPLATE/            # Issue templates
├── docs/
│   ├── INSTALLATION.md            # User installation guide
│   └── TROUBLESHOOTING.md         # Troubleshooting guide
├── scripts/
│   ├── init-repository.sh         # Repository setup script
│   ├── test-formula.sh           # Formula testing script
│   └── verify-installation.sh    # User verification script
├── README.md                      # Main documentation
├── CONTRIBUTING.md                # Contribution guidelines
└── MAINTAINER_GUIDE.md           # This file
```

## 🔄 Release Process

### Automatic Updates

The homebrew formula updates automatically when:

1. **New release is published** in ehsaniara/joblet repository
2. **Release workflow completes** successfully
3. **Admin UI is built** and packaged with macOS archives
4. **Homebrew repository dispatch** is triggered

### Manual Updates

If automatic updates fail or you need to make manual changes:

1. **Update formula manually**:
   ```bash
   cd homebrew-joblet
   
   # Edit Formula/rnx.rb
   # Update version, URLs, and SHA256 checksums
   
   # Calculate new checksums
   curl -sL "URL_TO_AMD64_ARCHIVE" | shasum -a 256
   curl -sL "URL_TO_ARM64_ARCHIVE" | shasum -a 256
   ```

2. **Test the changes**:
   ```bash
   ./scripts/test-formula.sh
   ```

3. **Commit and push**:
   ```bash
   git add Formula/rnx.rb
   git commit -m "Update rnx formula to vX.Y.Z"
   git push origin main
   ```

## 🧪 Testing

### Formula Testing

Use the comprehensive test suite:

```bash
# Run all tests
./scripts/test-formula.sh

# Run cleanup only
./scripts/test-formula.sh --cleanup-only

# Get help
./scripts/test-formula.sh --help
```

The test suite covers:
- Formula syntax validation
- Content validation
- URL accessibility
- Archive structure
- Installation testing (CLI-only and with admin UI)
- Post-installation verification

### Manual Testing

1. **Test installation from tap**:
   ```bash
   # Add tap locally
   brew tap ehsaniara/joblet /path/to/homebrew-joblet
   
   # Test different installation modes
   brew install rnx --without-admin
   brew uninstall rnx
   
   brew install rnx --with-admin
   brew uninstall rnx
   
   # Interactive mode (requires actual prompts)
   brew install rnx
   ```

2. **Test functionality**:
   ```bash
   rnx --version
   rnx --help
   rnx admin  # If admin UI installed
   ```

3. **Clean up test installation**:
   ```bash
   brew uninstall rnx
   brew untap ehsaniara/joblet
   ```

## 📋 Formula Maintenance

### Formula Structure

The `Formula/rnx.rb` contains:

```ruby
class Rnx < Formula
  # Metadata
  desc "Cross-platform CLI for Joblet distributed job execution system"
  homepage "https://github.com/ehsaniara/joblet"
  license "MIT"
  version_scheme 1

  # Multi-architecture URLs and checksums
  if Hardware::CPU.intel?
    url "https://github.com/ehsaniara/joblet/releases/download/vX.Y.Z/rnx-vX.Y.Z-darwin-amd64.tar.gz"
    sha256 "amd64_sha256_checksum"
  else
    url "https://github.com/ehsaniara/joblet/releases/download/vX.Y.Z/rnx-vX.Y.Z-darwin-arm64.tar.gz"
    sha256 "arm64_sha256_checksum"
  end

  # Dependencies and options
  depends_on "node" => :optional
  option "with-admin", "Install with web admin UI"
  option "without-admin", "Install CLI only"

  # Installation logic
  def install
    # Install base binary
    bin.install "rnx"
    
    # Determine admin UI installation
    install_admin = determine_admin_installation
    
    if install_admin
      setup_admin_ui
    end
    
    # Generate completions
    generate_completions_from_executable(bin/"rnx", "completion", "bash")
    generate_completions_from_executable(bin/"rnx", "completion", "zsh")
    generate_completions_from_executable(bin/"rnx", "completion", "fish")
  end

  # Test functionality
  def test
    assert_match version.to_s, shell_output("#{bin}/rnx --version")
    # Additional admin UI tests if installed
  end
  
  # Private helper methods
  private
  
  def determine_admin_installation
    # Interactive installation logic
  end
  
  def setup_admin_ui
    # Admin UI installation logic
  end
  
  def create_admin_launcher
    # Admin launcher creation
  end
  
  def caveats
    # Post-installation messages
  end
end
```

### Key Components

1. **Multi-architecture support**: Different URLs for Intel and Apple Silicon
2. **Interactive installation**: Prompts user based on Node.js availability
3. **Admin UI setup**: Installs Node.js dependencies and creates launcher
4. **Error handling**: Graceful degradation to CLI-only on failures
5. **Post-installation guidance**: Helpful messages and tips

### Common Maintenance Tasks

#### Update Version and URLs

When a new version is released:

1. **Update URLs**:
   ```ruby
   url "https://github.com/ehsaniara/joblet/releases/download/v1.2.3/rnx-v1.2.3-darwin-amd64.tar.gz"
   ```

2. **Calculate and update SHA256 checksums**:
   ```bash
   curl -sL "https://github.com/ehsaniara/joblet/releases/download/v1.2.3/rnx-v1.2.3-darwin-amd64.tar.gz" | shasum -a 256
   curl -sL "https://github.com/ehsaniara/joblet/releases/download/v1.2.3/rnx-v1.2.3-darwin-arm64.tar.gz" | shasum -a 256
   ```

3. **Update SHA256 values in formula**:
   ```ruby
   sha256 "new_amd64_checksum_here"
   # and
   sha256 "new_arm64_checksum_here"
   ```

#### Troubleshoot Installation Issues

Common issues and solutions:

1. **Checksum mismatch**:
   - Download the archive manually and recalculate checksum
   - Verify the archive is not corrupted
   - Check if the release was updated after initial publication

2. **Node.js installation failures**:
   - Check if Node.js formula is available in Homebrew
   - Test Node.js installation independently
   - Review error logs from user reports

3. **Admin UI setup failures**:
   - Verify archive contains `admin/` directory
   - Check that `admin/ui/dist/` contains built React app
   - Ensure `admin/server/package.json` has correct dependencies

#### Archive Structure Validation

The macOS archives must contain:

```
rnx-v1.2.3-darwin-amd64.tar.gz:
├── rnx                           # Main CLI binary
└── admin/
    ├── server/
    │   ├── package.json          # Node.js dependencies
    │   ├── server.js             # Admin server
    │   └── node_modules/         # Installed dependencies (production only)
    └── ui/
        ├── dist/                 # Built React application
        │   ├── index.html
        │   └── assets/
        └── index.html            # Dev server entry point
```

## 🔧 Automation Workflows

### Auto-Update Workflow

The `.github/workflows/update.yml` handles automatic formula updates:

#### Triggered by:
- Repository dispatch from main joblet repository
- Manual workflow dispatch with parameters

#### Process:
1. **Download release archives**
2. **Verify archive integrity**
3. **Calculate SHA256 checksums**
4. **Update formula with new URLs and checksums**
5. **Test formula syntax and installation**
6. **Commit and push changes**

#### Configuration:

Required secrets:
- `GITHUB_TOKEN`: Automatically provided by GitHub Actions
- Access to homebrew repository for commits

#### Troubleshooting Auto-Update:

1. **Check workflow logs** in GitHub Actions
2. **Verify release archives** are accessible and valid
3. **Test formula syntax** manually if workflow fails
4. **Check repository permissions** and secrets

### Manual Workflow Dispatch

You can trigger updates manually:

1. **Go to Actions tab** in homebrew repository
2. **Select "Update Formula" workflow**
3. **Click "Run workflow"**
4. **Provide parameters**:
   - Version tag (e.g., v1.2.3)
   - Clean version (e.g., 1.2.3)
   - AMD64 URL
   - ARM64 URL

## 🐛 Debugging and Troubleshooting

### Common Issues

#### 1. Formula Syntax Errors

```bash
# Check syntax
brew audit --strict Formula/rnx.rb

# Common issues:
# - Missing end statements
# - Incorrect indentation
# - Invalid Ruby syntax
```

#### 2. Installation Failures

```bash
# Test installation locally
HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source ./Formula/rnx.rb --verbose

# Check logs
brew install --verbose rnx 2>&1 | tee install.log
```

#### 3. Archive Issues

```bash
# Verify archive structure
curl -sL "ARCHIVE_URL" | tar -tzf -

# Test extraction
curl -sL "ARCHIVE_URL" | tar -xzf - -C /tmp/test-extraction
ls -la /tmp/test-extraction
```

#### 4. Node.js Dependency Issues

```bash
# Test Node.js installation
brew install node
node --version
npm --version

# Test admin UI dependencies
cd /tmp/test-extraction/admin/server
npm install --production
```

### Debugging Tools

#### Enable Debug Mode

```bash
# Formula debugging
export HOMEBREW_VERBOSE=1
export HOMEBREW_DEBUG=1

# Install with full output
brew install --verbose --debug rnx
```

#### Log Locations

- **Homebrew logs**: `~/Library/Logs/Homebrew/`
- **User verification**: Run `./scripts/verify-installation.sh --report`
- **GitHub Actions**: Repository Actions tab

### User Support

#### Issue Templates

The repository includes GitHub issue templates for:
- Installation issues
- Formula problems
- Feature requests

#### Support Process

1. **Triage issues** by type (installation vs functionality)
2. **Formula issues**: Handle in homebrew repository
3. **RNX functionality issues**: Forward to main repository
4. **Provide debugging steps** from troubleshooting guide
5. **Test fixes** with the testing scripts

## 📚 Documentation Maintenance

### Files to Keep Updated

1. **README.md**: Main user-facing documentation
2. **docs/INSTALLATION.md**: Detailed installation guide
3. **docs/TROUBLESHOOTING.md**: Comprehensive troubleshooting
4. **CONTRIBUTING.md**: Guidelines for contributors
5. **This file**: Maintainer procedures

### Documentation Updates

When making formula changes:

1. **Update installation examples** if process changes
2. **Add new troubleshooting scenarios** based on user reports
3. **Update version numbers** in examples
4. **Review and update links** to ensure they're current

### Version Management

- **Keep examples current** with latest stable version
- **Document breaking changes** in installation process
- **Maintain backward compatibility** notes where applicable

## 🚀 Release Checklist

### Pre-Release

- [ ] Verify main repository release is complete
- [ ] Check that admin UI archives are properly built
- [ ] Test archives manually if this is a major release

### Post-Release (Automatic)

- [ ] Verify auto-update workflow completed successfully
- [ ] Test installation from tap: `brew install ehsaniara/joblet/rnx`
- [ ] Check GitHub Issues for user reports
- [ ] Update documentation if needed

### Manual Release Process

If auto-update fails:

1. **Calculate checksums**:
   ```bash
   AMD64_URL="https://github.com/ehsaniara/joblet/releases/download/v1.2.3/rnx-v1.2.3-darwin-amd64.tar.gz"
   ARM64_URL="https://github.com/ehsaniara/joblet/releases/download/v1.2.3/rnx-v1.2.3-darwin-arm64.tar.gz"
   
   curl -sL "$AMD64_URL" | shasum -a 256
   curl -sL "$ARM64_URL" | shasum -a 256
   ```

2. **Update formula**:
   ```bash
   # Edit Formula/rnx.rb with new URLs and checksums
   ./scripts/test-formula.sh
   ```

3. **Commit and push**:
   ```bash
   git add Formula/rnx.rb
   git commit -m "Update rnx formula to v1.2.3"
   git push origin main
   ```

## 📞 Contact and Resources

### Maintainer Resources

- **Main repository**: https://github.com/ehsaniara/joblet
- **Homebrew documentation**: https://docs.brew.sh/Formula-Cookbook
- **Ruby documentation**: https://www.ruby-lang.org/en/documentation/

### Getting Help

- **Homebrew community**: https://github.com/Homebrew/discussions
- **Ruby help**: For formula syntax and logic
- **GitHub Actions**: For workflow debugging

### Emergency Contacts

If critical issues arise:
1. **Disable auto-updates** by removing workflow file temporarily
2. **Revert formula** to last known working version
3. **Communicate with users** via repository README or issues
4. **Coordinate with main repository** maintainers

---

*This guide should be updated as the project evolves and new maintenance patterns emerge.*