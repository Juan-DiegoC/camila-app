#!/usr/bin/env python3
"""
Create GitHub release with all necessary files
"""

import os
import shutil
import zipfile
from pathlib import Path

def build_executables():
    """Build all executables before packaging."""
    
    print("🔨 Building executables first...")
    
    # Run the build script
    import subprocess
    try:
        result = subprocess.run(["./build.sh"], check=True, capture_output=True, text=True)
        print("✅ Build completed successfully!")
        print(result.stdout[-500:])  # Show last 500 chars of output
        return True
    except subprocess.CalledProcessError as e:
        print(f"❌ Build failed: {e}")
        print("STDOUT:", e.stdout)
        print("STDERR:", e.stderr)
        return False
    except FileNotFoundError:
        print("❌ build.sh not found or not executable")
        print("Please run: chmod +x build.sh")
        return False

def create_release_packages():
    """Create release packages for different platforms."""
    
    # Build executables first
    if not build_executables():
        print("❌ Cannot create packages without successful build")
        return None
    
    print("\n📦 Creating release packages...")
    
    # Create release directory
    release_dir = "release"
    if os.path.exists(release_dir):
        shutil.rmtree(release_dir)
    os.makedirs(release_dir)
    
    # Common files needed for all packages
    common_files = [
        "file_metadata_extractor.py",
        "requirements.txt", 
        "README.md",
        "RELEASE_NOTES.md"
    ]
    
    # Windows package
    print("Creating Windows package...")
    windows_dir = f"{release_dir}/windows"
    os.makedirs(windows_dir)
    
    # Copy Windows binary
    if os.path.exists("file-indexer-windows.exe"):
        shutil.copy2("file-indexer-windows.exe", f"{windows_dir}/file-indexer.exe")
        print("✓ Copied file-indexer-windows.exe")
    else:
        print("❌ file-indexer-windows.exe not found - run ./build.sh first")
    
    # Copy common files
    for file in common_files:
        if os.path.exists(file):
            shutil.copy2(file, windows_dir)
    
    # Create Windows-specific README
    windows_readme = """# File Indexer for Windows

## Quick Setup (2 minutes)

1. **Install Python** (if not already installed)
   - Download from https://python.org
   - IMPORTANT: Check "Add Python to PATH" during installation

2. **Install Dependencies**
   - Open Command Prompt in this folder
   - Run: `pip install -r requirements.txt`

3. **Run the Application**
   - Double-click `file-indexer.exe`
   - Or run: `file-indexer.exe` from Command Prompt

## Features
- 🎯 Super easy navigation starting in Downloads folder
- 📊 Export to Excel, CSV, or both formats  
- 🔄 Tab completion and smart file extension handling
- 🐛 Debug mode for troubleshooting

## Usage
1. Navigate with arrow keys, press Space to select directory
2. Enter output filename, Tab toggles .xlsx ↔ .csv
3. Choose export format and press Enter to process

See README.md for full documentation.
"""
    
    with open(f"{windows_dir}/WINDOWS_README.txt", "w") as f:
        f.write(windows_readme)
    
    # Create Windows ZIP
    with zipfile.ZipFile(f"{release_dir}/file-indexer-windows-v1.0.0.zip", "w") as zipf:
        for root, dirs, files in os.walk(windows_dir):
            for file in files:
                file_path = os.path.join(root, file)
                arcname = os.path.relpath(file_path, windows_dir)
                zipf.write(file_path, arcname)
    
    print("✅ Windows package: file-indexer-windows-v1.0.0.zip")
    
    # Linux package
    print("Creating Linux package...")
    linux_dir = f"{release_dir}/linux"
    os.makedirs(linux_dir)
    
    if os.path.exists("file-indexer-linux"):
        shutil.copy2("file-indexer-linux", f"{linux_dir}/file-indexer")
        os.chmod(f"{linux_dir}/file-indexer", 0o755)
    
    for file in common_files:
        if os.path.exists(file):
            shutil.copy2(file, linux_dir)
    
    # Create Linux setup script
    linux_setup = """#!/bin/bash
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
"""
    
    with open(f"{linux_dir}/setup.sh", "w") as f:
        f.write(linux_setup)
    os.chmod(f"{linux_dir}/setup.sh", 0o755)
    
    # Create Linux tar.gz
    shutil.make_archive(f"{release_dir}/file-indexer-linux-v1.0.0", "gztar", linux_dir)
    print("✅ Linux package: file-indexer-linux-v1.0.0.tar.gz")
    
    # macOS package  
    print("Creating macOS package...")
    macos_dir = f"{release_dir}/macos"
    os.makedirs(macos_dir)
    
    if os.path.exists("file-indexer-macos"):
        shutil.copy2("file-indexer-macos", f"{macos_dir}/file-indexer")
        os.chmod(f"{macos_dir}/file-indexer", 0o755)
    
    for file in common_files:
        if os.path.exists(file):
            shutil.copy2(file, macos_dir)
    
    # macOS setup is same as Linux
    shutil.copy2(f"{linux_dir}/setup.sh", macos_dir)
    
    shutil.make_archive(f"{release_dir}/file-indexer-macos-v1.0.0", "gztar", macos_dir)
    print("✅ macOS package: file-indexer-macos-v1.0.0.tar.gz")
    
    # Source code package
    print("Creating source code package...")
    source_files = [
        "main.go",
        "go.mod", 
        "build.sh",
        "file_metadata_extractor.py",
        "requirements.txt",
        "README.md",
        "RELEASE_NOTES.md"
    ]
    
    source_dir = f"{release_dir}/source"
    os.makedirs(source_dir)
    
    for file in source_files:
        if os.path.exists(file):
            shutil.copy2(file, source_dir)
    
    shutil.make_archive(f"{release_dir}/file-indexer-source-v1.0.0", "zip", source_dir)
    print("✅ Source package: file-indexer-source-v1.0.0.zip")
    
    # Create release summary
    summary = f"""# File Indexer v1.0.0 Release

## Download Links

### Windows (Recommended)
- **file-indexer-windows-v1.0.0.zip** - Complete Windows package
  - Includes: TUI executable + Python script + dependencies list
  - Requires: Python 3.6+ installation

### Linux  
- **file-indexer-linux-v1.0.0.tar.gz** - Complete Linux package
  - Includes: TUI executable + Python script + setup script
  - Run: `./setup.sh` then `./file-indexer`

### macOS
- **file-indexer-macos-v1.0.0.tar.gz** - Complete macOS package  
  - Includes: TUI executable + Python script + setup script
  - Run: `./setup.sh` then `./file-indexer`

### Source Code
- **file-indexer-source-v1.0.0.zip** - Full source code
  - Build with: `./build.sh` (requires Go 1.21+)

## Quick Start

1. **Download** the package for your platform
2. **Extract** to any folder
3. **Install Python** 3.6+ (Windows/macOS from python.org, Linux: use package manager)  
4. **Run setup** (Linux/macOS: `./setup.sh`, Windows: `pip install -r requirements.txt`)
5. **Launch**: `./file-indexer` (Linux/macOS) or double-click `file-indexer.exe` (Windows)

## What's New in v1.0.0

✅ **Fixed Navigation** - Clear path display, proper Downloads startup
✅ **Better UX** - Space to select, Tab for smart autocomplete  
✅ **Visual Feedback** - Icons, checkmarks, helpful error messages
✅ **Cross-platform** - Windows, Linux, macOS binaries

See RELEASE_NOTES.md for full details.

## File Size Info
- Windows package: ~2MB (requires Python install)
- Linux/macOS packages: ~1.5MB each  
- Total download: ~5MB for all platforms

## Next: Fully Bundled Version (v1.1.0)
Working on a single .exe with embedded Python runtime - no Python installation needed!
"""
    
    with open(f"{release_dir}/RELEASE_SUMMARY.md", "w") as f:
        f.write(summary)
    
    # List all created files
    print("\n📦 Release packages created:")
    for file in os.listdir(release_dir):
        if file.endswith(('.zip', '.tar.gz')):
            size = os.path.getsize(f"{release_dir}/{file}") / 1024 / 1024
            print(f"  📄 {file} ({size:.1f}MB)")
    
    print(f"\n✅ All release files in: {release_dir}/")
    return release_dir

def create_github_release_script():
    """Create a script for GitHub release creation."""
    
    script = '''#!/bin/bash
# GitHub Release Creation Script
# Run this after pushing to GitHub

VERSION="v1.0.0"
REPO="your-username/file-indexer-tui"  # Update with your repo

echo "Creating GitHub release $VERSION..."

# Create release
gh release create $VERSION \\
    --title "File Indexer TUI $VERSION" \\
    --notes-file RELEASE_NOTES.md \\
    release/file-indexer-windows-v1.0.0.zip \\
    release/file-indexer-linux-v1.0.0.tar.gz \\
    release/file-indexer-macos-v1.0.0.tar.gz \\
    release/file-indexer-source-v1.0.0.zip

echo "✅ Release created!"
echo "View at: https://github.com/$REPO/releases/tag/$VERSION"
'''
    
    with open("create_github_release.sh", "w") as f:
        f.write(script)
    os.chmod("create_github_release.sh", 0o755)
    
    print("📝 GitHub release script created: create_github_release.sh")
    print("Edit the REPO variable and run: ./create_github_release.sh")

if __name__ == "__main__":
    print("🚀 File Indexer Release Creator")
    print("=" * 50)
    
    # Create release packages
    release_dir = create_release_packages()
    
    # Create GitHub release script
    create_github_release_script()
    
    print("\n🎉 Release creation complete!")
    print("\nNext steps:")
    print("1. Test the packages on different platforms")
    print("2. Push code to GitHub repository") 
    print("3. Edit create_github_release.sh with your repo name")
    print("4. Run: ./create_github_release.sh")
    print("\n📁 All files ready in:", release_dir)