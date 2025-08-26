#!/bin/bash

# Build script for File Indexer TUI - Complete automated build

echo "🚀 File Indexer TUI - Complete Build Process"
echo "=============================================="

# Set Go environment
export PATH=$PATH:$HOME/go/bin
export GOPATH=$HOME/go

# Change to project directory
cd "$(dirname "$0")"

# Check if Go is available
if ! command -v $HOME/go/bin/go &> /dev/null; then
    echo "❌ Go not found at $HOME/go/bin/go"
    echo "Please install Go from https://golang.org/dl/"
    echo "Or update the PATH in this script"
    exit 1
fi

echo "✓ Go found: $($HOME/go/bin/go version)"

# Download dependencies
echo ""
echo "📦 Downloading Go dependencies..."
$HOME/go/bin/go mod tidy
if [ $? -ne 0 ]; then
    echo "❌ Failed to download dependencies"
    exit 1
fi

# Clean old binaries
echo ""
echo "🧹 Cleaning old binaries..."
rm -f file-indexer-linux file-indexer-windows.exe file-indexer-macos

# Build for Linux (static linking, optimized)
echo ""
echo "🐧 Building for Linux..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $HOME/go/bin/go build -ldflags="-s -w" -o file-indexer-linux main.go
if [ $? -eq 0 ]; then
    echo "✅ Linux build successful ($(ls -lh file-indexer-linux | awk '{print $5}'))"
else
    echo "❌ Linux build failed"
    exit 1
fi

# Build for Windows (static linking, optimized, maximum compatibility)
echo ""
echo "🪟 Building for Windows..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $HOME/go/bin/go build \
    -ldflags="-s -w -X main.version=1.0.0" \
    -trimpath \
    -o file-indexer-windows.exe main.go
if [ $? -eq 0 ]; then
    echo "✅ Windows build successful ($(ls -lh file-indexer-windows.exe | awk '{print $5}'))"
else
    echo "❌ Windows build failed"
    exit 1
fi

# Build for macOS (static linking, optimized)
echo ""
echo "🍎 Building for macOS..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $HOME/go/bin/go build -ldflags="-s -w" -o file-indexer-macos main.go
if [ $? -eq 0 ]; then
    echo "✅ macOS build successful ($(ls -lh file-indexer-macos | awk '{print $5}'))"
else
    echo "❌ macOS build failed"
    exit 1
fi

echo ""
echo "🎉 Build completed successfully!"
echo ""
echo "📦 Binaries created:"
echo "  🐧 file-indexer-linux      ($(ls -lh file-indexer-linux | awk '{print $5}'))"
echo "  🪟 file-indexer-windows.exe ($(ls -lh file-indexer-windows.exe | awk '{print $5}'))"
echo "  🍎 file-indexer-macos       ($(ls -lh file-indexer-macos | awk '{print $5}'))"
echo ""
echo "🔧 Build optimizations:"
echo "  ✓ Static linking (no external dependencies)"
echo "  ✓ Stripped debug info (smaller size)"
echo "  ✓ Trimmed paths (more portable)"
echo "  ✓ Maximum Windows compatibility"
echo ""
echo "📋 Next steps:"
echo "  1. Test the executables"
echo "  2. Run: python3 create_release.py (to package for distribution)"
echo "  3. Upload release packages to GitHub"
echo ""
echo "🪟 Windows users can now just:"
echo "  1. Download the release ZIP"
echo "  2. Install Python 3.6+"
echo "  3. Run: pip install -r requirements.txt"
echo "  4. Double-click file-indexer.exe"