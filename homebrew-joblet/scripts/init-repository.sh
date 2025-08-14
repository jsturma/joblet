#!/bin/bash
set -e

# RNX Homebrew Repository Initialization Script
# This script sets up the ehsaniara/homebrew-joblet repository

REPO_NAME="homebrew-joblet"
GITHUB_ORG="ehsaniara"
REPO_URL="https://github.com/$GITHUB_ORG/$REPO_NAME"

echo "🍺 RNX Homebrew Repository Setup"
echo "================================="

# Check prerequisites
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is required but not installed"
    echo "💡 Install with: brew install gh"
    exit 1
fi

if ! command -v git &> /dev/null; then
    echo "❌ Git is required but not installed"
    exit 1
fi

# Verify GitHub authentication
echo "🔐 Verifying GitHub authentication..."
if ! gh auth status &> /dev/null; then
    echo "❌ Not authenticated with GitHub"
    echo "💡 Run: gh auth login"
    exit 1
fi

echo "✅ GitHub authentication verified"

# Check if repository already exists
echo "🔍 Checking if repository exists..."
if gh repo view "$GITHUB_ORG/$REPO_NAME" &> /dev/null; then
    echo "⚠️  Repository $REPO_URL already exists"
    echo "🤔 Do you want to continue and update the existing repository? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "❌ Aborted by user"
        exit 1
    fi
    REPO_EXISTS=true
else
    echo "✅ Repository does not exist, will create new one"
    REPO_EXISTS=false
fi

# Create temporary directory for setup
TEMP_DIR=$(mktemp -d)
echo "📁 Using temporary directory: $TEMP_DIR"

cd "$TEMP_DIR"

if [ "$REPO_EXISTS" = true ]; then
    # Clone existing repository
    echo "📥 Cloning existing repository..."
    git clone "$REPO_URL.git" "$REPO_NAME"
    cd "$REPO_NAME"
    
    # Create backup of existing files
    if [ -f "Formula/rnx.rb" ]; then
        echo "💾 Backing up existing formula..."
        cp "Formula/rnx.rb" "Formula/rnx.rb.backup.$(date +%Y%m%d_%H%M%S)"
    fi
else
    # Create new directory structure
    echo "📁 Creating repository structure..."
    mkdir -p "$REPO_NAME"
    cd "$REPO_NAME"
    
    # Initialize git repository
    git init
    git remote add origin "$REPO_URL.git"
fi

# Copy files from source directory
echo "📋 Copying homebrew files..."
SOURCE_DIR="$(dirname "$0")/.."

# Create directory structure
mkdir -p Formula .github/workflows docs scripts

# Copy formula
cp "$SOURCE_DIR/Formula/rnx.rb" "Formula/"
echo "✅ Copied Formula/rnx.rb"

# Copy workflows
cp "$SOURCE_DIR/.github/workflows/update.yml" ".github/workflows/"
echo "✅ Copied .github/workflows/update.yml"

# Copy documentation
cp "$SOURCE_DIR/README.md" "."
cp "$SOURCE_DIR/docs/INSTALLATION.md" "docs/"
cp "$SOURCE_DIR/docs/TROUBLESHOOTING.md" "docs/"
echo "✅ Copied documentation files"

# Copy this script for future use
cp "$0" "scripts/"
echo "✅ Copied initialization script"

# Create additional helpful files
echo "📝 Creating additional files..."

# Create .gitignore
cat > .gitignore << 'EOF'
# macOS
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Temporary files
*.tmp
*.temp
*.log
*.backup.*

# Test artifacts
test-*/
*.tar.gz
*.zip

# Editor files
.vscode/
.idea/
*.swp
*.swo
*~

# Local development
.env
.env.local
EOF

# Create CONTRIBUTING.md
cat > CONTRIBUTING.md << 'EOF'
# Contributing to RNX Homebrew Tap

## Formula Updates

The formula is automatically updated when new releases are published in the main [joblet repository](https://github.com/ehsaniara/joblet).

### Manual Updates

If you need to manually update the formula:

1. **Update version and URLs** in `Formula/rnx.rb`
2. **Calculate new checksums**:
   ```bash
   curl -sL "URL_TO_ARCHIVE" | shasum -a 256
   ```
3. **Test the formula**:
   ```bash
   brew audit --strict Formula/rnx.rb
   brew install --build-from-source ./Formula/rnx.rb
   ```
4. **Submit pull request**

### Testing Changes

Before submitting changes:

```bash
# Syntax check
brew audit --strict Formula/rnx.rb

# Test installation
brew install --build-from-source ./Formula/rnx.rb --with-admin
brew install --build-from-source ./Formula/rnx.rb --without-admin

# Test functionality
rnx --version
rnx --help
```

### Reporting Issues

- **Formula issues**: Open issue in this repository
- **RNX functionality**: Open issue in [main repository](https://github.com/ehsaniara/joblet/issues)
EOF

# Create GitHub issue templates
mkdir -p .github/ISSUE_TEMPLATE

cat > .github/ISSUE_TEMPLATE/installation-issue.yml << 'EOF'
name: Installation Issue
description: Report problems with homebrew installation
title: "[INSTALL] "
labels: ["installation", "bug"]
body:
  - type: markdown
    attributes:
      value: |
        Thanks for reporting an installation issue! Please provide details below.
        
  - type: input
    id: macos-version
    attributes:
      label: macOS Version
      description: Output of `sw_vers`
      placeholder: "macOS 13.5 (22G74)"
    validations:
      required: true
      
  - type: dropdown
    id: architecture
    attributes:
      label: Architecture
      description: What type of Mac are you using?
      options:
        - Intel (x86_64)
        - Apple Silicon (ARM64)
    validations:
      required: true
      
  - type: input
    id: homebrew-version
    attributes:
      label: Homebrew Version
      description: Output of `brew --version`
      placeholder: "Homebrew 4.1.0"
    validations:
      required: true
      
  - type: dropdown
    id: installation-type
    attributes:
      label: Installation Type Attempted
      description: Which installation method did you use?
      options:
        - Interactive (brew install rnx)
        - With Admin UI (--with-admin)
        - CLI Only (--without-admin)
    validations:
      required: true
      
  - type: textarea
    id: error-message
    attributes:
      label: Error Message
      description: Complete error message from the installation
      placeholder: Paste the full error message here...
    validations:
      required: true
      
  - type: textarea
    id: command-output
    attributes:
      label: Command Output
      description: Full output of the installation command
      placeholder: Paste the complete output here...
    validations:
      required: false
EOF

cat > .github/ISSUE_TEMPLATE/formula-issue.yml << 'EOF'
name: Formula Issue
description: Report problems with the homebrew formula itself
title: "[FORMULA] "
labels: ["formula", "bug"]
body:
  - type: markdown
    attributes:
      value: |
        Report issues specific to the homebrew formula (not RNX functionality).
        
  - type: dropdown
    id: issue-type
    attributes:
      label: Issue Type
      description: What kind of formula issue is this?
      options:
        - Formula syntax error
        - Checksum mismatch
        - Download failure
        - Dependency issue
        - Installation logic error
    validations:
      required: true
      
  - type: textarea
    id: description
    attributes:
      label: Problem Description
      description: Describe the issue in detail
    validations:
      required: true
      
  - type: textarea
    id: expected-behavior
    attributes:
      label: Expected Behavior
      description: What should have happened?
    validations:
      required: true
      
  - type: textarea
    id: formula-audit
    attributes:
      label: Formula Audit Output
      description: Output of `brew audit --strict Formula/rnx.rb`
      placeholder: Paste audit output here...
    validations:
      required: false
EOF

# Set up git configuration
echo "⚙️  Configuring git..."
git config user.name "RNX Homebrew Bot"
git config user.email "noreply@github.com"

# Add all files
git add .

# Create initial commit
if [ "$REPO_EXISTS" = true ]; then
    COMMIT_MSG="Update homebrew tap with latest formula and documentation"
else
    COMMIT_MSG="Initial homebrew tap setup for RNX CLI

- Interactive homebrew formula with admin UI support
- Auto-update GitHub Actions workflow
- Comprehensive documentation and troubleshooting guides
- Support for both Intel and Apple Silicon Macs
- Node.js detection and optional admin UI installation"
fi

git commit -m "$COMMIT_MSG"

if [ "$REPO_EXISTS" = false ]; then
    # Create repository
    echo "🏗️  Creating GitHub repository..."
    gh repo create "$GITHUB_ORG/$REPO_NAME" \
        --public \
        --description "Homebrew tap for RNX - Remote job execution CLI for Joblet" \
        --homepage "https://github.com/ehsaniara/joblet"
fi

# Push changes
echo "📤 Pushing to GitHub..."
git push -u origin main

echo ""
echo "🎉 Repository setup complete!"
echo ""
echo "📋 Next Steps:"
echo "1. Set up repository secrets:"
echo "   - Go to: $REPO_URL/settings/secrets/actions"
echo "   - Add secret: HOMEBREW_UPDATE_TOKEN (GitHub token with repo access)"
echo ""
echo "2. Test the formula:"
echo "   brew tap $GITHUB_ORG/joblet"
echo "   brew install rnx --dry-run"
echo ""
echo "3. Test auto-update workflow:"
echo "   - Create a test release in ehsaniara/joblet repository"
echo "   - Check if homebrew formula updates automatically"
echo ""
echo "4. Update main repository:"
echo "   - Ensure HOMEBREW_UPDATE_TOKEN secret is set in main repo"
echo "   - Test release workflow triggers homebrew update"
echo ""
echo "🔗 Repository URL: $REPO_URL"

# Cleanup
cd "$HOME"
rm -rf "$TEMP_DIR"

echo "✅ Setup complete! Temporary files cleaned up."