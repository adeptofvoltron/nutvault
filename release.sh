#!/bin/bash

# release.sh - Build binaries for Linux, macOS, and Windows
# Usage: ./release.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project name
PROJECT_NAME="nutvault"

# Create dist directory
DIST_DIR="dist"
echo -e "${GREEN}Creating ${DIST_DIR} directory...${NC}"
rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

# Build flags for smaller binaries
LDFLAGS="-s -w"

# Build function
build_binary() {
    local os=$1
    local arch=$2
    local output_name="${PROJECT_NAME}-${os}-${arch}"
    
    # Add .exe extension for Windows
    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo -e "${YELLOW}Building ${output_name}...${NC}"
    
    GOOS="${os}" GOARCH="${arch}" go build \
        -ldflags "${LDFLAGS}" \
        -o "${DIST_DIR}/${output_name}" \
        ./cmd/nutvault
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built ${output_name}${NC}"
        
        # Generate SHA256 checksum
        if command -v sha256sum &> /dev/null; then
            sha256sum "${DIST_DIR}/${output_name}" > "${DIST_DIR}/${output_name}.sha256"
        elif command -v shasum &> /dev/null; then
            shasum -a 256 "${DIST_DIR}/${output_name}" > "${DIST_DIR}/${output_name}.sha256"
        else
            echo -e "${YELLOW}Warning: sha256sum or shasum not found, skipping checksum generation${NC}"
        fi
    else
        echo -e "${RED}✗ Failed to build ${output_name}${NC}"
        exit 1
    fi
}

# Build for different platforms
echo -e "${GREEN}Starting build process...${NC}"
echo ""

# Linux
build_binary "linux" "amd64"
build_binary "linux" "arm64"

# macOS
build_binary "darwin" "amd64"
build_binary "darwin" "arm64"

# Windows
build_binary "windows" "amd64"
build_binary "windows" "arm64"

echo ""
echo -e "${GREEN}Build complete! Binaries are in ${DIST_DIR}/${NC}"
echo ""
echo "Files created:"
ls -lh "${DIST_DIR}" | tail -n +2

