# Windows Terminal Settings Configuration

This directory contains a pre-configured Windows Terminal `settings.json` file that provides a consistent development environment across different machines.

## Installation Instructions

1. **Locate your Windows Terminal settings file:**
   - Press `Win + R`, type `%LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState`, and press Enter
   - Or navigate to: `C:\Users\[YourUsername]\AppData\Local\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\`

2. **Backup your existing settings (recommended):**
   - Copy your current `settings.json` to `settings.json.backup`

3. **Install the new settings:**
   - Copy the `windows-terminal-settings.json` from this repository to the LocalState folder
   - Rename it to `settings.json`

## Configuration Details

### Default Settings Applied:
- **Default Profile**: Windows PowerShell (instead of Ubuntu/WSL)
- **Font**: Consolas 12pt (standard Windows font, available on all systems)
- **Color Scheme**: "One Half Dark" (modern, easy on the eyes)
- **Starting Directory**: `%USERPROFILE%` (user home directory)

### Available Profiles:
1. **Windows PowerShell** (default)
2. **Command Prompt** 
3. **Ubuntu** (if WSL is installed)

### Color Schemes Included:
- **Campbell** - Classic Windows Terminal colors
- **Campbell Powershell** - PowerShell variant
- **One Half Dark** (default) - Modern dark theme
- **One Half Light** - Light theme variant

### Key Bindings:
- `Ctrl+C` - Copy
- `Ctrl+V` - Paste  
- `Ctrl+Shift+F` - Find
- `Alt+Shift+D` - Split pane with duplicate

## Customization Notes

- **Color Scheme**: To change the color scheme, modify the `"colorScheme"` value in the `defaults` section
- **Font**: The font is set to "Consolas" which is available on all Windows systems. You can change this to any installed font
- **Default Profile**: The default profile is set to PowerShell. The GUID `{61c54bbd-c2c6-5271-96e7-009a87ff44bf}` corresponds to Windows PowerShell

## Troubleshooting

- **Settings not applying**: Make sure Windows Terminal is completely closed before replacing the settings file
- **Missing profiles**: Some profiles (like Ubuntu) will only appear if the corresponding software is installed
- **Font issues**: If Consolas doesn't display correctly, try changing the font face to "Courier New" or any monospace font

## Reverting Changes

To revert to your original settings:
1. Delete the current `settings.json`
2. Rename `settings.json.backup` to `settings.json`
3. Restart Windows Terminal