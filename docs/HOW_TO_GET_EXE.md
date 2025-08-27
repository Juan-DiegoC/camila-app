# How to Get the Windows .exe File

## Option 1: Download Pre-built Release (Easiest)

1. **Go to Releases**: Visit the [GitHub Releases page](https://github.com/YOUR-USERNAME/file-indexer-tui/releases)
2. **Download**: Click `file-indexer-windows-v1.0.0.zip` 
3. **Extract**: Unzip to any folder
4. **You get**:
   - `file-indexer.exe` ← The TUI application
   - `file_metadata_extractor.py` ← The Python backend  
   - `requirements.txt` ← Python dependencies
   - `WINDOWS_README.txt` ← Setup guide

## Option 2: Build It Yourself

### Prerequisites
- Go 1.21+ installed from [golang.org](https://golang.org/dl/)
- Git (optional, for cloning)

### Build Steps
```bash
# 1. Get the source code
git clone https://github.com/YOUR-USERNAME/file-indexer-tui.git
cd file-indexer-tui

# 2. Build the Windows executable
./build.sh

# 3. Create release package  
python3 create_release.py

# 4. Your .exe is ready!
# Look in: release/windows/file-indexer.exe
```

### What the Build Script Does
```bash
# Downloads Go dependencies
go mod tidy

# Builds for Windows
GOOS=windows GOARCH=amd64 go build -o file-indexer-windows.exe main.go

# Builds for Linux
GOOS=linux GOARCH=amd64 go build -o file-indexer-linux main.go

# Builds for macOS  
GOOS=darwin GOARCH=amd64 go build -o file-indexer-macos main.go
```

## What You Need to Run the .exe

The Windows executable needs these files in the same folder:
```
📁 Your folder/
├── file-indexer.exe              ← Main TUI app
├── file_metadata_extractor.py    ← Python script  
├── requirements.txt              ← Dependencies list
└── WINDOWS_README.txt            ← Setup instructions
```

## One-Time Setup for Windows Users

1. **Install Python**: Download from [python.org](https://python.org)
   - ✅ **IMPORTANT**: Check "Add Python to PATH" during installation
2. **Install dependencies**: Open Command Prompt in your folder, run:
   ```cmd
   pip install -r requirements.txt
   ```
3. **Run the app**: Double-click `file-indexer.exe`

## Troubleshooting

### "Go not found" during build
- Install Go from https://golang.org/dl/
- Make sure Go is in your PATH

### "Python not found" when running .exe  
- Install Python from https://python.org
- Reinstall with "Add Python to PATH" checked

### Build succeeds but .exe doesn't work
- Make sure `file_metadata_extractor.py` is in the same folder
- Run `pip install -r requirements.txt` first

## File Sizes

- **file-indexer.exe**: ~4MB (Go TUI application)  
- **file_metadata_extractor.py**: ~15KB (Python backend)
- **Complete Windows package**: ~4.2MB
- **With Python dependencies installed**: ~50MB total disk space

## Why Not Bundle Python?

We could create a single 50MB .exe with Python embedded, but the current approach is better because:

✅ **Smaller download** (4MB vs 50MB)  
✅ **Faster startup** (uses system Python)
✅ **Easy to update** (just replace Python file)  
✅ **Works with any Python version** 3.6+
✅ **Professional software approach** (like most dev tools)

The one-time Python setup is worth it for the benefits!