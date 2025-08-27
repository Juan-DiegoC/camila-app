#!/usr/bin/env python3
"""
Python bundling script to create a standalone executable that includes Python runtime.
This will create a single .exe file that doesn't require Python to be installed.
"""

import os
import shutil
import subprocess
import sys
from pathlib import Path


def create_bundled_version():
    """Create a bundled version using PyInstaller."""
    
    print("🐍 Creating bundled Python executable...")
    
    # Check if PyInstaller is installed
    try:
        import PyInstaller
    except ImportError:
        print("Installing PyInstaller...")
        subprocess.check_call([sys.executable, "-m", "pip", "install", "pyinstaller"])
    
    # Create PyInstaller spec file
    spec_content = '''# -*- mode: python ; coding: utf-8 -*-

block_cipher = None

a = Analysis(
    ['file_metadata_extractor.py'],
    pathex=[],
    binaries=[],
    datas=[],
    hiddenimports=[
        'PyPDF2',
        'fitz',
        'openpyxl',
        'mimetypes',
        'csv',
        'logging',
        'traceback'
    ],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=block_cipher,
    noarchive=False,
)

pyz = PYZ(a.pure, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.zipfiles,
    a.datas,
    [],
    name='file_metadata_extractor_standalone',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)
'''
    
    # Write spec file
    with open('bundle.spec', 'w') as f:
        f.write(spec_content)
    
    # Build with PyInstaller
    print("Building standalone executable...")
    subprocess.check_call([
        sys.executable, "-m", "PyInstaller", 
        "--onefile",
        "--console", 
        "--name=file_metadata_extractor_standalone",
        "bundle.spec"
    ])
    
    # Copy the built executable
    if os.path.exists("dist/file_metadata_extractor_standalone.exe"):
        shutil.copy2("dist/file_metadata_extractor_standalone.exe", "file_metadata_extractor_standalone.exe")
        print("✅ Windows standalone executable created: file_metadata_extractor_standalone.exe")
    
    if os.path.exists("dist/file_metadata_extractor_standalone"):
        shutil.copy2("dist/file_metadata_extractor_standalone", "file_metadata_extractor_standalone")
        print("✅ Linux standalone executable created: file_metadata_extractor_standalone")
    
    # Clean up
    if os.path.exists("build"):
        shutil.rmtree("build")
    if os.path.exists("dist"):
        shutil.rmtree("dist")
    if os.path.exists("bundle.spec"):
        os.remove("bundle.spec")
    if os.path.exists("file_metadata_extractor_standalone.spec"):
        os.remove("file_metadata_extractor_standalone.spec")
    
    return True


def create_embedded_python():
    """Create an embedded Python solution for Windows."""
    
    print("🐍 Creating embedded Python solution...")
    
    # This would download Python embeddable package and bundle it
    # For now, let's create the structure
    
    embedded_dir = "file_indexer_embedded"
    os.makedirs(embedded_dir, exist_ok=True)
    
    # Create a batch file that uses embedded Python
    batch_content = '''@echo off
cd /d "%~dp0"

REM Check if embedded Python exists
if not exist "python\\python.exe" (
    echo ERROR: Embedded Python not found!
    echo Please run setup_embedded.bat first
    pause
    exit /b 1
)

REM Run the Python script with embedded interpreter
python\\python.exe file_metadata_extractor.py %*

pause
'''
    
    with open(f"{embedded_dir}/run_indexer.bat", "w") as f:
        f.write(batch_content)
    
    # Create setup script
    setup_content = '''@echo off
echo Setting up File Indexer with embedded Python...

REM This script would download and extract Python embeddable
echo.
echo Manual setup required:
echo 1. Download Python 3.11 embeddable from https://python.org
echo 2. Extract to 'python' folder here
echo 3. Install pip in embedded Python
echo 4. Run: python\\python.exe -m pip install -r requirements.txt
echo.

pause
'''
    
    with open(f"{embedded_dir}/setup_embedded.bat", "w") as f:
        f.write(setup_content)
    
    # Copy Python files
    shutil.copy2("file_metadata_extractor.py", embedded_dir)
    shutil.copy2("requirements.txt", embedded_dir)
    
    print(f"✅ Embedded Python structure created in {embedded_dir}/")
    return embedded_dir


def create_portable_solution():
    """Create the most user-friendly solution."""
    
    print("📦 Creating portable solution...")
    
    # Create a hybrid approach: Go TUI + Bundled Python
    portable_dir = "file_indexer_portable"
    os.makedirs(portable_dir, exist_ok=True)
    
    # Copy Go binaries
    if os.path.exists("file-indexer-windows.exe"):
        shutil.copy2("file-indexer-windows.exe", f"{portable_dir}/file-indexer.exe")
    
    # Try to create bundled Python version
    try:
        create_bundled_version()
        if os.path.exists("file_metadata_extractor_standalone.exe"):
            shutil.copy2("file_metadata_extractor_standalone.exe", f"{portable_dir}/file_metadata_extractor_bundled.exe")
    except Exception as e:
        print(f"Bundled Python creation failed: {e}")
        # Fall back to regular Python files
        shutil.copy2("file_metadata_extractor.py", portable_dir)
        shutil.copy2("requirements.txt", portable_dir)
    
    # Create unified executable that tries bundled first, then regular Python
    unified_go_content = '''package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	execDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	
	// Try bundled Python version first
	bundledPath := filepath.Join(execDir, "file_metadata_extractor_bundled.exe")
	if _, err := os.Stat(bundledPath); err == nil {
		fmt.Println("Using bundled Python version...")
		cmd := exec.Command(bundledPath, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
		return
	}
	
	// Fall back to regular Python
	scriptPath := filepath.Join(execDir, "file_metadata_extractor.py")
	if _, err := os.Stat(scriptPath); err == nil {
		fmt.Println("Using system Python...")
		var cmd *exec.Cmd
		args := append([]string{scriptPath}, os.Args[1:]...)
		
		if runtime.GOOS == "windows" {
			cmd = exec.Command("python", args...)
		} else {
			cmd = exec.Command("python3", args...)
		}
		
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
		return
	}
	
	fmt.Println("ERROR: No Python executable found!")
	fmt.Println("Please ensure either:")
	fmt.Println("1. file_metadata_extractor_bundled.exe is present (no Python needed)")
	fmt.Println("2. file_metadata_extractor.py is present and Python is installed")
}
'''
    
    with open("unified_launcher.go", "w") as f:
        f.write(unified_go_content)
    
    # Build unified launcher
    if os.system("go build -o file_indexer_portable/file_metadata_extractor_unified.exe unified_launcher.go") == 0:
        print("✅ Unified launcher created")
    
    # Clean up
    if os.path.exists("unified_launcher.go"):
        os.remove("unified_launcher.go")
    
    # Create README for portable version
    readme_content = '''# File Indexer - Portable Version

This is a completely portable version that works with or without Python installed.

## What's Included

- `file-indexer.exe` - Main TUI application
- `file_metadata_extractor_bundled.exe` - Standalone Python processor (no Python needed)
- `file_metadata_extractor.py` - Backup Python script (requires Python)
- `requirements.txt` - Python dependencies (if using backup)

## How to Use

### Option 1: Zero Setup (Recommended)
Just run `file-indexer.exe` - it will automatically use the bundled Python version.

### Option 2: With System Python
If the bundled version doesn't work, install Python 3.11+ and run:
```
pip install -r requirements.txt
```
Then run `file-indexer.exe`

## No Installation Required
This entire folder is portable - copy it anywhere and run!
'''
    
    with open(f"{portable_dir}/README.txt", "w") as f:
        f.write(readme_content)
    
    print(f"✅ Portable solution created in {portable_dir}/")
    return portable_dir


if __name__ == "__main__":
    print("🚀 File Indexer Bundling Tool")
    print("=" * 50)
    
    # Install dependencies first
    print("Installing Python dependencies...")
    subprocess.check_call([sys.executable, "-m", "pip", "install", "-r", "requirements.txt"])
    
    try:
        # Create bundled version
        create_bundled_version()
        
        # Create portable solution
        portable_dir = create_portable_solution()
        
        print("\n✅ All bundling complete!")
        print(f"📦 Portable version: {portable_dir}/")
        print("\nFor end users:")
        print("1. Download the portable folder")
        print("2. Run file-indexer.exe")
        print("3. No Python installation needed!")
        
    except Exception as e:
        print(f"\n❌ Bundling failed: {e}")
        print("\nFalling back to standard Python version...")
        print("Users will need to install Python and dependencies manually.")