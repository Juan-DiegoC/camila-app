# Windows Setup Guide

## Quick Setup (5 minutes)

### Step 1: Download Files
Download these 3 files and put them in the same folder:
- `file-indexer-windows.exe` 
- `file_metadata_extractor.py`
- `requirements.txt`

### Step 2: Install Python
If you don't have Python installed:
1. Go to https://python.org/downloads
2. Download Python 3.11 or newer
3. **IMPORTANT**: Check "Add Python to PATH" during installation

### Step 3: Install Python Dependencies
Open Command Prompt in the folder with your files and run:
```cmd
pip install -r requirements.txt
```

### Step 4: Run the Application
Double-click `file-indexer-windows.exe` or run:
```cmd
file-indexer-windows.exe
```

## Troubleshooting

### "Python not found"
- Reinstall Python with "Add to PATH" checked
- Or try: `py -3 -m pip install -r requirements.txt`

### "Permission denied when saving Excel"
- Close Excel if you have the output file open
- Try running as Administrator
- Use CSV export instead

### Terminal/Display Issues  
- Use Windows Terminal (recommended) instead of old Command Prompt
- Right-click title bar → Properties → Use Unicode UTF-8

## Folder Structure
```
📁 FileIndexer/
  ├── file-indexer-windows.exe    ← Main application
  ├── file_metadata_extractor.py  ← Python script  
  ├── requirements.txt            ← Dependencies
  └── README.md                   ← Instructions
```

## What This Tool Does
1. **Select a folder** - Navigate to any folder on your computer
2. **Choose output name** - Name your index file (Excel/CSV)  
3. **Pick format** - Excel, CSV, or both
4. **Process files** - Automatically extracts metadata from all files
5. **Get results** - Organized spreadsheet with file information

Perfect for organizing document folders, analyzing file collections, or creating inventories!