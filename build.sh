#!/bin/bash

echo "PrivacyCheck Go Build Script"
echo "============================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

echo "Go version:"
go version

# Clean previous builds
echo ""
echo "Cleaning previous builds..."
rm -rf dist
mkdir -p dist

# Download dependencies
echo ""
echo "Downloading dependencies..."
go mod tidy
if [ $? -ne 0 ]; then
    echo "Error: Failed to download dependencies"
    exit 1
fi

# Get build info
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date '+%Y-%m-%d %H:%M:%S')

LDFLAGS="-s -w -X privacycheck/core.BuildDate='$BUILD_DATE' -X privacycheck/core.GitCommit=$GIT_COMMIT"

# Build for Windows x64
echo ""
echo "Building for Windows x64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "dist/privacycheck-windows-x64.exe" .
if [ $? -ne 0 ]; then
    echo "Error: Failed to build for Windows x64"
    exit 1
fi

# Build for Linux x64
echo ""
echo "Building for Linux x64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "dist/privacycheck-linux-x64" .
if [ $? -ne 0 ]; then
    echo "Error: Failed to build for Linux x64"
    exit 1
fi

# Make Linux binary executable
chmod +x "dist/privacycheck-linux-x64"

# Show build results
echo ""
echo "Build completed successfully!"
echo ""
echo "Built files:"
ls -la dist/

echo ""
echo "File sizes:"
du -h dist/*

echo ""
echo "Build script completed."
