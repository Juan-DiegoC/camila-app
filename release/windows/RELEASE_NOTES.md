# File Indexer TUI v1.0.0

A user-friendly Terminal User Interface (TUI) for extracting file metadata and creating Excel/CSV indexes.

## 🎯 **Major Improvements in v1.0.0**

### **✅ Fixed Navigation Issues**
- **Clear path display** - Always shows current directory prominently  
- **Proper Downloads folder startup** - Automatically starts in user's Downloads folder
- **Improved arrow key navigation** - Smooth browsing through directories and files
- **Visual file browser** - Icons for folders 📁 and files 📄
- **Smart autocomplete** - Tab for extensions (.xlsx ↔ .csv toggle) in Step 2

### **🚀 Enhanced User Experience**
- **Space bar selection** - Press Space to select current directory (more intuitive)
- **Tab navigation** - Jump between common folders (Downloads, Documents, Home)
- **Better visual feedback** - Clear indicators for selected options ✅⬜
- **Improved error messages** - Helpful troubleshooting with emojis and clear steps
- **Real-time path updates** - See exactly where you are at all times

### **📊 Features**
- **Multiple export formats** - Excel (.xlsx), CSV (.csv), or both
- **Debug mode** - Toggle detailed logging for troubleshooting
- **Cross-platform** - Windows, Linux, and macOS binaries
- **Zero dependencies** - Single executable file

## 📥 **Download Options**

### **Windows Users (Recommended)**
Download: `file-indexer-windows.exe` + `file_metadata_extractor.py` + `requirements.txt`

**Setup:**
1. Put all 3 files in the same folder
2. Install Python 3.11+ from python.org (check "Add to PATH")
3. Run: `pip install -r requirements.txt`
4. Double-click `file-indexer-windows.exe`

### **Linux/macOS Users**
Download: `file-indexer-linux` or `file-indexer-macos` + Python files

**Setup:**
```bash
chmod +x file-indexer-linux
pip3 install -r requirements.txt
./file-indexer-linux
```

## 🎮 **How to Use**

### **Step 1: Navigate & Select Directory** 
```
📍 Current: C:\Users\YourName\Downloads

┌─────────────────────────────────────┐
│ Select Directory to Index           │
│                                     │
│ 📁 ..                              │
│ 📁 Project Files                   │
│ 📁 Documents                       │
│ 📄 report.pdf                      │
│ 📄 data.xlsx                       │
└─────────────────────────────────────┘

📍 Navigation:
  ↑↓ Browse files/folders    📁 Enter = Go into folder
  Space = Select this directory  Tab = Jump to common folders
```

- **↑↓** arrows to browse
- **Enter** to go into folders  
- **Space** to select current directory
- **Tab** to jump to common locations

### **Step 2: Choose Output Name**
```
Step 2: Output File Name

my_index.xlsx
_

💡 Tips:
  • Tab toggles between .xlsx ↔ .csv extension
  • Enter to continue, Esc to go back
```

- Type your desired filename
- **Tab** toggles between .xlsx and .csv extensions
- **Enter** to proceed

### **Step 3: Configure Export**
```
Step 3: Export Configuration

📊 Export Format:
  1) Excel only (.xlsx)     ✅
  2) CSV only (.csv)        ⬜
  3) Both Excel + CSV       ⬜

🔧 d) Debug mode          ⬜

💫 Press 1-3 to select format, 'd' for debug
⏩ Enter to start processing, Esc to go back
```

- **1-3** to choose export format
- **d** to toggle debug mode
- **Enter** to start processing

## 🔧 **Technical Details**

### **Navigation Improvements**
- **Current path display** - `📍 Current: /path/to/directory` always visible
- **Downloads folder detection** - Tries Downloads → Documents → Home
- **Parent directory (..)** - Always shown when not at root
- **File limiting** - Shows first 20 files for performance
- **Directory-first sorting** - Folders listed before files

### **Autocomplete Behavior**
- **Step 1 Tab** - Cycles through common directories
- **Step 2 Tab** - Toggles file extensions (.xlsx ↔ .csv)
- **Smart extension handling** - Adds extension if missing
- **Format awareness** - Suggests .csv if CSV mode selected

### **Error Handling**
- **Python detection** - Clear instructions if Python not found
- **Permission issues** - Specific troubleshooting steps
- **File access errors** - Helpful error messages with solutions
- **Debug mode** - Detailed logging for problem diagnosis

## 🐛 **Troubleshooting**

### **"Python not found" Error**
```
❌ Python execution failed: file not found

🐍 Python is not installed or not in PATH

Solutions:
1. Install Python from https://python.org
2. Make sure 'Add Python to PATH' is checked during installation  
3. Restart this application after installing Python
```

### **Navigation Issues**
- **Can't see current directory?** - Look for `📍 Current:` line at top
- **Arrow keys not working?** - Try different terminal (Windows Terminal recommended)
- **Can't select directory?** - Press **Space** instead of Enter
- **Lost in filesystem?** - Press **Tab** to jump to common folders

### **Export Problems**
- **Permission denied** - Close Excel if file is open
- **File exists** - Choose a different name or delete existing file
- **Large directories** - Enable debug mode to see progress

## 🎁 **Bundle Options**

### **Current Setup (Requires Python)**
- Download 3 files: .exe + .py + requirements.txt
- Install Python once
- Works with any Python version 3.6+

### **Future: Fully Bundled Version**
We're working on a single .exe that includes Python runtime:
- No Python installation required
- Single file download
- Larger file size (~50MB vs 5MB)
- Coming in v1.1.0

## 🚀 **What's Next**

### **v1.1.0 (Planned)**
- **Bundled Python version** - No Python installation needed
- **Drag & drop support** - Drop folders onto the executable  
- **Recent directories** - Quick access to previously indexed folders
- **Batch processing** - Process multiple directories at once

### **GitHub Repository**
- Source code available for building from scratch
- Issue tracking for bug reports and feature requests
- Automated releases with GitHub Actions