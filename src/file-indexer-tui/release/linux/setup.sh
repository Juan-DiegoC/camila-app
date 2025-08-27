#!/bin/bash
# File Indexer Linux Setup

echo "File Indexer Linux Setup"
echo "========================"

# Make executable
chmod +x file-indexer

# Install Python dependencies
echo "Installing Python dependencies..."
if command -v pip3 &> /dev/null; then
    pip3 install -r requirements.txt
elif command -v pip &> /dev/null; then
    pip install -r requirements.txt  
else
    echo "ERROR: pip not found. Please install pip first:"
    echo "  Ubuntu/Debian: sudo apt install python3-pip"
    echo "  Fedora: sudo dnf install python3-pip" 
    exit 1
fi

echo ""
echo "✅ Setup complete!"
echo "Run: ./file-indexer"
