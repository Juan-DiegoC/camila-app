# File Indexer TUI

A user-friendly Terminal User Interface (TUI) for the File Metadata Extractor. This tool provides an intuitive way to index files and extract metadata without needing to remember command-line arguments.

## Features

- 🎯 **Super Easy to Use** - Step-by-step guided interface
- 🗂️ **Smart Directory Navigation** - Starts in Downloads folder, easy browsing
- ⌨️ **Tab Completion** - Auto-complete file paths and filenames
- 📊 **Multiple Export Formats** - Excel (.xlsx), CSV, or both
- 🔧 **Debug Mode** - Optional detailed logging
- 💻 **Cross-Platform** - Runs on Windows, macOS, and Linux
- 📦 **Single Binary** - No dependencies except Python

## Quick Start

### Windows
1. Download `file-indexer-windows.exe` 
2. Copy `file_metadata_extractor.py` to the same folder
3. Double-click `file-indexer-windows.exe` or run from Command Prompt
4. Ensure Python 3 is installed on your system

### Linux/macOS
1. Download `file-indexer-linux` (Linux) or `file-indexer-macos` (macOS)
2. Copy `file_metadata_extractor.py` to the same folder  
3. Make executable: `chmod +x file-indexer-linux`
4. Run: `./file-indexer-linux`

## Usage Flow

### Step 1: Select Directory
- Use arrow keys to navigate folders
- Press **Enter** to select the current directory
- Press **Tab** to go up one level
- Starts in your Downloads folder for convenience

### Step 2: Output File Name
- Enter desired output filename (e.g., `my_index.xlsx`)
- Press **Tab** for auto-completion of existing files
- Supports both `.xlsx` and `.csv` extensions

### Step 3: Configuration
- **1** - Excel only (.xlsx format)
- **2** - CSV only (.csv format)  
- **3** - Both Excel and CSV (recommended)
- **d** - Toggle debug mode (shows detailed processing info)
- Press **Enter** to start processing

### Step 4: Processing
- Wait while files are being indexed
- Progress shown in real-time
- Press **Ctrl+C** to cancel if needed

### Step 5: Results
- Shows success message and file locations
- Press **r** to process another directory
- Press **Enter** or **Esc** to quit

## Navigation Keys

| Key | Action |
|-----|--------|
| ↑↓ Arrow Keys | Navigate files/directories |
| **Enter** | Select directory or confirm action |
| **Tab** | Auto-complete or go up directory |
| **Esc** | Go back to previous step |
| **q** or **Ctrl+C** | Quit application |

## Export Formats

### Excel (.xlsx)
- Formatted with Spanish headers
- Automatic column sizing  
- Formula calculations for page ranges
- Compatible with Microsoft Excel, LibreOffice

### CSV (.csv)
- Plain text format
- Command-line viewable with `head`, `cat`
- Universal compatibility
- Easy to import into any spreadsheet program

### Both Formats
- Creates both .xlsx and .csv files
- Best of both worlds - formatted Excel + universal CSV
- Automatic fallback if Excel creation fails

## File Requirements

The TUI needs these files in the same directory:
```
file-indexer-windows.exe     # Windows executable
file_metadata_extractor.py   # Python script (required)
requirements.txt            # Python dependencies
```

## System Requirements

- **Python 3.6+** installed and available as `python` or `python3`
- **Windows**: No additional requirements
- **Linux/macOS**: Terminal emulator

## Python Dependencies

The Python script automatically handles these dependencies:
- PyPDF2 (PDF processing)
- PyMuPDF (fallback PDF processing) 
- openpyxl (Excel file creation)

Install with: `pip install -r requirements.txt`

## Troubleshooting

### "Permission Denied" Error
- Close Excel if the output file is open
- Try a different output directory
- Use CSV-only mode as fallback

### "Python not found" Error  
- Install Python 3 from python.org
- Ensure Python is in your system PATH
- Try running `python --version` in Command Prompt

### TUI Display Issues
- Use a modern terminal (Windows Terminal, iTerm2, etc.)
- Ensure terminal supports color and Unicode
- Try resizing the terminal window

### File Processing Errors
- Enable debug mode (press 'd' in configuration)
- Check the log output for detailed error information
- Ensure you have read permissions for the target directory

## Building from Source

If you want to build the binaries yourself:

```bash
# Clone or download the source
cd file-indexer-tui

# Install Go 1.21+
# Run build script
./build.sh
```

This creates binaries for Windows, Linux, and macOS.

## Examples

### Indexing Downloads Folder
1. Run TUI → automatically starts in Downloads
2. Press Enter to select Downloads
3. Enter "downloads_index.xlsx"
4. Select "Both Excel+CSV" 
5. Process and get both formats

### Processing Large Directory with Debug
1. Navigate to target directory
2. Enter output filename
3. Press 'd' to enable debug mode
4. Select export format and process
5. Review detailed logs for any issues

### Quick CSV Export
1. Select directory
2. Enter filename with .csv extension  
3. Choose "CSV only" format
4. View results immediately with command line tools

## Support

For issues with the TUI interface, check that:
- Python script is in the same directory
- Python 3 is properly installed
- You have write permissions for output directory
- Terminal supports the required features

The TUI is designed to be extremely user-friendly - if you're having trouble, the interface should guide you through each step with clear instructions.