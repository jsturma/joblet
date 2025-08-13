#!/bin/bash
# Generate README.md for release package

PLATFORM="$1"
GOOS="$2"
GOARCH="$3"
OUTPUT_FILE="$4"

cat > "$OUTPUT_FILE" << EOF
# Joblet RNX CLI - $PLATFORM

## Installation

### Linux
\`\`\`bash
tar -xzf rnx-$PLATFORM.tar.gz
cd rnx-$PLATFORM
./install.sh [install-directory]

# Install admin UI dependencies (required for rnx admin command)
cd /opt/joblet/admin/server  # or your install directory
npm install
\`\`\`

### macOS
\`\`\`bash
tar -xzf rnx-$PLATFORM.tar.gz
cd rnx-$PLATFORM
./install.sh [install-directory]

# For admin UI support, clone the repository and install dependencies
git clone https://github.com/ehsaniara/joblet.git
cd joblet/admin/server
npm install
\`\`\`

### Windows
\`\`\`cmd
tar -xzf rnx-$PLATFORM.tar.gz
cd rnx-$PLATFORM
install.bat [install-directory]

:: For admin UI support, clone the repository and install dependencies
git clone https://github.com/ehsaniara/joblet.git
cd joblet\\admin\\server
npm install
\`\`\`

## Important Notes

- **Node.js Required**: The admin UI requires Node.js to be installed on your system
- **npm install Required**: You MUST run \`npm install\` in the admin/server directory before using \`rnx admin\`
- **Error Prevention**: Without installing dependencies, you will encounter \`ERR_MODULE_NOT_FOUND\` errors when running \`rnx admin\`

## Platform Support

- **RNX CLI**: Works on all platforms
- **Joblet Server**: Linux only
- **Admin UI**: Requires Node.js and npm dependencies installed

## Architecture: $GOARCH
## OS: $GOOS

For full server installation, use the Linux release.
EOF