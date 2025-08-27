# Build Guide - File Indexer TUI

## Quick Build (Get Windows .exe)

### Prerequisites
- Go 1.21+ installed
- Python 3.6+ with pip

### Build Steps

1. **Clone/Download the repository**
2. **Navigate to the project folder**
3. **Run the build script**:
   ```bash
   ./build.sh
   ```
   
   Or on Windows:
   ```cmd
   go build -o file-indexer-windows.exe main.go
   ```

4. **Create release packages**:
   ```bash
   python3 create_release.py
   ```

### What You Get

After building:
```
📁 Your folder/
├── file-indexer-windows.exe     ← Windows TUI app
├── file-indexer-linux           ← Linux TUI app  
├── file-indexer-macos           ← macOS TUI app
├── file_metadata_extractor.py   ← Python backend
├── requirements.txt             ← Python deps
└── release/                     ← Distribution packages
    ├── file-indexer-windows-v1.0.0.zip
    ├── file-indexer-linux-v1.0.0.tar.gz
    └── file-indexer-macos-v1.0.0.tar.gz
```

## For End Users (No Building Needed)

### Windows
1. **Download**: `file-indexer-windows-v1.0.0.zip` from releases
2. **Extract** to any folder
3. **Install Python** 3.6+ from python.org
4. **Run**: `pip install -r requirements.txt`  
5. **Launch**: Double-click `file-indexer.exe`

### Linux/macOS
1. **Download**: `file-indexer-linux-v1.0.0.tar.gz` or macOS version
2. **Extract**: `tar -xzf file-indexer-linux-v1.0.0.tar.gz`
3. **Run setup**: `./setup.sh`
4. **Launch**: `./file-indexer`

## Build Requirements

### Go Dependencies
- github.com/charmbracelet/bubbletea v0.25.0
- github.com/charmbracelet/bubbles v0.17.1
- github.com/charmbracelet/lipgloss v0.9.1

### Python Dependencies  
- PyPDF2==3.0.1 (PDF processing)
- PyMuPDF==1.23.14 (PDF fallback)
- openpyxl==3.1.2 (Excel creation)

## Troubleshooting Build Issues

### "Go not found"
Install Go from https://golang.org/dl/

### "Python not found"  
Install Python from https://python.org

### "Build failed"
```bash
# Clean and retry
rm -f file-indexer-*
go clean
go mod tidy
./build.sh
```

### "Release packages empty"
Make sure binaries exist before running `create_release.py`:
```bash
ls -la file-indexer-*
# Should show the executables
```

## Cross-Platform Building

### Build for Windows (from Linux/macOS)
```bash
GOOS=windows GOARCH=amd64 go build -o file-indexer-windows.exe main.go
```

### Build for Linux (from Windows/macOS)  
```bash
GOOS=linux GOARCH=amd64 go build -o file-indexer-linux main.go
```

### Build for macOS (from Windows/Linux)
```bash
GOOS=darwin GOARCH=amd64 go build -o file-indexer-macos main.go
```

## File Structure

```
project/
├── main.go                      ← Go TUI source
├── go.mod                       ← Go dependencies  
├── build.sh                     ← Build script
├── file_metadata_extractor.py   ← Python backend
├── requirements.txt             ← Python deps
├── create_release.py            ← Release packager
└── README.md                    ← Documentation
```

The Go TUI calls the Python script as a subprocess, so both files are needed for the complete application.