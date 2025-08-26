package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			MarginLeft(2).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginLeft(2)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000")).
			MarginLeft(2)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			MarginLeft(2)

	pathStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#874BFD")).
			MarginLeft(2)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1).
			MarginLeft(2).
			MarginRight(2)

	activeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#04B575")).
			Padding(1).
			MarginLeft(2).
			MarginRight(2)
)

// App states
type state int

const (
	selectingDirectory state = iota
	selectingOutput
	configuring
	processing
	finished
)

// Directory item for the list
type directoryItem struct {
	name string
	path string
	isDir bool
}

func (i directoryItem) Title() string {
	if i.isDir {
		return fmt.Sprintf("📁 %s", i.name)
	}
	return fmt.Sprintf("📄 %s", i.name)
}

func (i directoryItem) Description() string {
	if i.name == ".." {
		return "Go up to parent directory"
	}
	if i.isDir {
		return "" // Don't show "Directory" - it's obvious from the folder icon
	}
	return "File"
}

func (i directoryItem) FilterValue() string {
	return i.name
}

type model struct {
	state           state
	directoryList   list.Model
	currentPath     string
	outputInput     textinput.Model
	selectedDir     string
	outputPath      string
	exportFormat    string // "excel", "csv", "both"
	debugMode       bool
	processing      bool
	result          string
	error           string
	width           int
	height          int
	autocompleteOptions []string
	autocompleteIndex   int
	showingAutocomplete bool
	// Filter mode
	filtering       bool
	filterInput     textinput.Model
	allDirectories  []directoryItem
	filteredDirs    []directoryItem
	advancedMode    bool // Toggle with Ctrl+D
}

func initialModel() model {
	// Get Downloads directory
	startDir := getDownloadsDirectory()
	
	// Initialize directory list
	items := getDirectoryItems(startDir)
	
	// Create list with nice styling
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#626262"))
		
	directoryList := list.New(items, delegate, 60, 15)
	directoryList.Title = fmt.Sprintf("Navigate: %s", startDir)
	directoryList.SetShowStatusBar(false)
	directoryList.SetShowHelp(false)

	// Initialize output input
	ti := textinput.New()
	ti.Placeholder = "Enter filename (e.g. my_index.xlsx)"
	ti.Focus()
	ti.Width = 50

	// Initialize filter input
	filterInput := textinput.New()
	filterInput.Placeholder = "Type to filter directories..."
	filterInput.Width = 50

	return model{
		state:         selectingDirectory,
		directoryList: directoryList,
		currentPath:   startDir,
		outputInput:   ti,
		exportFormat:  "excel", // Default to Excel only
		debugMode:     false,
		filtering:     false,
		filterInput:   filterInput,
		allDirectories: convertToDirectoryItems(items),
		advancedMode:  false,
	}
}

func getDownloadsDirectory() string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	// Try common download directories in order of preference
	downloadDirs := []string{
		filepath.Join(userDir, "Downloads"),
		filepath.Join(userDir, "Download"),
		filepath.Join(userDir, "Documents"),
		userDir,
		".",
	}

	for _, dir := range downloadDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return "."
}

func getDirectoryItems(dirPath string) []list.Item {
	var items []list.Item

	// Add parent directory option if not at root
	if parent := filepath.Dir(dirPath); parent != dirPath {
		items = append(items, directoryItem{
			name:  "..",
			path:  parent,
			isDir: true,
		})
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return items
	}

	// Only add directories (no files shown in directory selection)
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			// Skip hidden directories and ensure name is not empty
			if name == "" || strings.HasPrefix(name, ".") {
				continue
			}
			// Ensure the name is valid and displayable
			if len(strings.TrimSpace(name)) == 0 {
				continue
			}
			items = append(items, directoryItem{
				name:  name,
				path:  filepath.Join(dirPath, name),
				isDir: true,
			})
		}
	}

	return items
}

func convertToDirectoryItems(items []list.Item) []directoryItem {
	var dirItems []directoryItem
	for _, item := range items {
		if dirItem, ok := item.(directoryItem); ok {
			dirItems = append(dirItems, dirItem)
		}
	}
	return dirItems
}

func filterDirectories(dirs []directoryItem, filter string) []directoryItem {
	if filter == "" {
		return dirs
	}
	
	filter = strings.ToLower(filter)
	var filtered []directoryItem
	
	for _, dir := range dirs {
		if strings.Contains(strings.ToLower(dir.name), filter) {
			filtered = append(filtered, dir)
		}
	}
	
	return filtered
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.directoryList.SetWidth(msg.Width - 6)
		m.directoryList.SetHeight(msg.Height - 10)
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case selectingDirectory:
			if m.filtering {
				// Handle filtering mode
				switch msg.String() {
				case "ctrl+c", "q":
					return m, tea.Quit
				case "esc":
					// Exit filter mode
					m.filtering = false
					m.filterInput.SetValue("")
					m.filterInput.Blur()
					// Reset to show all directories
					items := getDirectoryItems(m.currentPath)
					m.directoryList.SetItems(items)
					m.directoryList.Title = fmt.Sprintf("Navigate: %s", m.currentPath)
					m.allDirectories = convertToDirectoryItems(items)
					return m, nil
				case "enter":
					// Apply filter and exit filter mode
					m.filtering = false
					m.filterInput.Blur()
					return m, nil
				case "tab":
					// Auto-complete with first match
					if len(m.filteredDirs) > 0 {
						m.filterInput.SetValue(m.filteredDirs[0].name)
					}
					return m, nil
				default:
					// Update filter input and apply filter
					m.filterInput, cmd = m.filterInput.Update(msg)
					filterText := m.filterInput.Value()
					m.filteredDirs = filterDirectories(m.allDirectories, filterText)
					// Update list with filtered items
					var filteredItems []list.Item
					for _, dir := range m.filteredDirs {
						filteredItems = append(filteredItems, dir)
					}
					m.directoryList.SetItems(filteredItems)
					return m, cmd
				}
			} else {
				// Normal directory navigation mode
				switch msg.String() {
				case "ctrl+c", "q":
					return m, tea.Quit
				case "ctrl+d":
					// Toggle advanced mode
					m.advancedMode = !m.advancedMode
					return m, nil
				case "i":
					// Enter filter mode
					m.filtering = true
					m.filterInput.Focus()
					m.filterInput.SetValue("")
					// Store all directories for filtering
					items := getDirectoryItems(m.currentPath)
					m.directoryList.Title = fmt.Sprintf("Filter in: %s", m.currentPath)
					m.allDirectories = convertToDirectoryItems(items)
					m.filteredDirs = m.allDirectories
					return m, textinput.Blink
			case "enter":
				if selected, ok := m.directoryList.SelectedItem().(directoryItem); ok {
					if selected.isDir {
						if selected.name == ".." {
							// Go to parent directory
							m.currentPath = selected.path
						} else {
							// Go into subdirectory or select current directory
							if selected.name != ".." {
								m.currentPath = selected.path
							}
						}
						// Update directory list and path display
						items := getDirectoryItems(m.currentPath)
						m.directoryList.SetItems(items)
						m.directoryList.Title = fmt.Sprintf("Navigate: %s", m.currentPath)
						m.allDirectories = convertToDirectoryItems(items)
						return m, nil
					}
				}
			case " ", "space":
				// Select current directory
				m.selectedDir = m.currentPath
				m.state = selectingOutput
				// Use simple "index" filename (xlsx will be auto-appended)
				m.outputInput.SetValue("index")
				return m, nil
			case "u", "up":
				// Go up one directory level
				parentDir := filepath.Dir(m.currentPath)
				if parentDir != m.currentPath { // Not at root
					m.currentPath = parentDir
					items := getDirectoryItems(m.currentPath)
					m.directoryList.SetItems(items)
					m.directoryList.Title = fmt.Sprintf("Navigate: %s", m.currentPath)
					m.allDirectories = convertToDirectoryItems(items)
					return m, nil
				}
			}
			}

		case selectingOutput:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "ctrl+d", "\x04":
				// Toggle advanced mode and preview it immediately
				m.advancedMode = !m.advancedMode
				// If we have a valid filename, go to confirmation step to show the change
				if strings.TrimSpace(m.outputInput.Value()) != "" {
					m.outputPath = m.outputInput.Value()
					m.state = configuring
				}
				return m, nil
			case "b", "backspace", "esc":
				m.state = selectingDirectory
				return m, nil
			case "enter":
				if m.outputInput.Value() != "" {
					m.outputPath = m.outputInput.Value()
					m.state = configuring
					return m, nil
				}
			case "tab":
				// Toggle between xlsx and csv preview (but don't force extension in input)
				if m.advancedMode {
					if m.exportFormat == "excel" {
						m.exportFormat = "csv"
					} else {
						m.exportFormat = "excel"
					}
				}
				return m, nil
			}

		case configuring:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "ctrl+d", "\x04":
				// Toggle advanced mode (Ctrl+D can be represented as \x04)
				m.advancedMode = !m.advancedMode
				return m, nil
			case "1":
				// Edit directory
				m.state = selectingDirectory
				return m, nil
			case "2":
				// Edit output filename
				m.state = selectingOutput
				return m, nil
			case "3":
				if m.advancedMode {
					// Toggle format in advanced mode
					if m.exportFormat == "excel" {
						m.exportFormat = "csv"
					} else if m.exportFormat == "csv" {
						m.exportFormat = "both"
					} else {
						m.exportFormat = "excel"
					}
				}
			case "d":
				if m.advancedMode {
					m.debugMode = !m.debugMode
				}
			case "b", "backspace", "esc":
				m.state = selectingOutput
				return m, nil
			case "enter":
				m.state = processing
				return m, m.runPythonScript()
			}

		case processing:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case finished:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "r":
				// Reset and start over
				return initialModel(), nil
			case "b", "backspace":
				// Go back to configuration
				m.state = configuring
				m.error = ""
				m.result = ""
				return m, nil
			case "enter", "esc":
				return m, tea.Quit
			}
		}
	
	case processCompleteMsg:
		m.processing = false
		if msg.error != "" {
			m.error = msg.error
		} else {
			m.result = msg.result
		}
		m.state = finished
		return m, nil
	}

	// Update components based on state
	switch m.state {
	case selectingDirectory:
		m.directoryList, cmd = m.directoryList.Update(msg)
		// Don't update filter input here - it's handled in the key cases above
	case selectingOutput:
		m.outputInput, cmd = m.outputInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("📁 File Metadata Indexer"))
	content.WriteString("\n\n")

	switch m.state {
	case selectingDirectory:
		// Show current path prominently
		content.WriteString(pathStyle.Render(fmt.Sprintf("📍 Current: %s", m.currentPath)))
		content.WriteString("\n\n")
		
		// Show filter input if filtering
		if m.filtering {
			content.WriteString(activeBoxStyle.Render(
				"🔍 Filter Directories:\n\n" +
				m.filterInput.View() + "\n\n" +
				m.directoryList.View() + "\n\n" +
				"⌨️ Controls:\n" +
				"  Type to filter  Tab = Auto-complete  Enter = Apply\n" +
				"  Esc = Cancel filter"))
		} else {
			// Ensure list title is updated with current path
			m.directoryList.Title = fmt.Sprintf("🗺️ %s", m.currentPath)
			content.WriteString(activeBoxStyle.Render(
				"Step 1: Navigate and Select Directory (Directories Only)\n\n" +
				m.directoryList.View() + "\n\n" +
				"📍 Navigation:\n" +
				"  ↑↓ Browse directories    📁 Enter = Go into folder\n" +
				"  Space = Select this directory  U = Go up one level\n" +
				"  I = Filter directories  Ctrl+D = Advanced mode"))
		}

	case selectingOutput:
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("Selected: %s", m.selectedDir)))
		content.WriteString("\n\n")
		// Show where file will be saved
		previewName := m.outputInput.Value()
		if !strings.Contains(previewName, ".") {
			previewName += ".xlsx" // Default extension
		}
		fullPath := filepath.Join(m.selectedDir, previewName)
		
		content.WriteString(activeBoxStyle.Render(
			"Step 2: Output File Name\n\n" +
			m.outputInput.View() + "\n\n" +
			fmt.Sprintf("💾 Will save as: %s\n\n", fullPath) +
			"💡 Tips:\n" +
			"  • .xlsx extension auto-added\n" +
			"  • File saved in selected directory\n" +
			"  • Enter to continue, B/Esc to go back\n" +
			"  • Ctrl+D = Advanced mode"))

	case configuring:
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("📁 Directory: %s", m.selectedDir)))
		content.WriteString("\n")
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("📄 Output: %s", m.outputPath)))
		content.WriteString("\n\n")

		configBox := "Step 3: Confirm Settings\n\n"
		configBox += "✨ Review your settings:\n\n"
		configBox += fmt.Sprintf("📁 Directory: %s\n", filepath.Base(m.selectedDir))
		configBox += fmt.Sprintf("📄 Output: %s\n", m.outputPath)
		
		if m.advancedMode {
			formatName := "Excel (.xlsx)"
			if m.exportFormat == "csv" {
				formatName = "CSV (.csv)"
			} else if m.exportFormat == "both" {
				formatName = "Excel + CSV"
			}
			configBox += fmt.Sprintf("📊 Format: %s\n", formatName)
			if m.debugMode {
				configBox += "🔧 Debug: Enabled\n"
			}
			configBox += "\n🅰️ ADVANCED MODE - More Options Available:\n"
			configBox += "  ❶ = Change directory     ❷ = Change filename\n"
			configBox += "  ❸ = Toggle format        🔧 d = Toggle debug\n"
			configBox += "  🔄 Ctrl+D = Exit advanced mode\n"
		} else {
			configBox += "📊 Format: Excel (.xlsx)\n\n"
			configBox += "⚙️ QUICK MODE - Press keys to edit:\n"
			configBox += "  ❶ = Change directory     ❷ = Change filename\n"
			configBox += "  🔥 Ctrl+D = Advanced mode (CSV, debug)\n"
		}
		configBox += "\n⏩ Enter to start processing  •  B/Esc to go back"

		content.WriteString(activeBoxStyle.Render(configBox))

	case processing:
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("📁 Directory: %s", m.selectedDir)))
		content.WriteString("\n")
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("📄 Output: %s", m.outputPath)))
		content.WriteString("\n\n")
		content.WriteString(activeBoxStyle.Render(
			"⏳ Processing Files...\n\n" +
				"🔄 Scanning directory and extracting metadata\n" +
				"📊 This may take a while for large directories\n\n" +
				"Press Ctrl+C to cancel"))

	case finished:
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("📁 Directory: %s", m.selectedDir)))
		content.WriteString("\n")
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("📄 Output: %s", m.outputPath)))
		content.WriteString("\n\n")

		if m.error != "" {
			content.WriteString(errorStyle.Render("❌ Processing Failed"))
			content.WriteString("\n")
			content.WriteString(boxStyle.Render(m.error))
			content.WriteString("\n")
			content.WriteString(helpStyle.Render("🔄 Press 'r' to try again  •  Enter/Esc to quit"))
		} else {
			content.WriteString(successStyle.Render("✅ Processing Completed Successfully!"))
			content.WriteString("\n")
			content.WriteString(boxStyle.Render(m.result))
			content.WriteString("\n")
			content.WriteString(helpStyle.Render("🔄 Press 'r' to process another directory  •  Enter/Esc to quit"))
		}
	}

	content.WriteString("\n\n")
	content.WriteString(helpStyle.Render("Press 'q' or Ctrl+C to quit anytime"))

	return content.String()
}

func checkmark(selected bool) string {
	if selected {
		return "✅"
	}
	return "⬜"
}

// Message type for process completion
type processCompleteMsg struct {
	result string
	error  string
}

func (m model) runPythonScript() tea.Cmd {
	return func() tea.Msg {
		// Find the Python script
		scriptPath := findPythonScript()
		if scriptPath == "" {
			return processCompleteMsg{
				error: "❌ Could not find file_metadata_extractor.py\n\nPlease ensure the Python script is in the same directory as this executable.\n\nRequired files:\n• file_metadata_extractor.py\n• requirements.txt",
			}
		}

		// Build command arguments
		args := []string{scriptPath}
		
		// Add directory argument
		args = append(args, "--directory", m.selectedDir)
		
		// Build full output path: selected_directory/filename.xlsx
		filename := m.outputPath
		// Auto-append .xlsx if no extension provided
		if !strings.Contains(filename, ".") {
			if m.exportFormat == "csv" {
				filename += ".csv"
			} else {
				filename += ".xlsx"
			}
		}
		// Create full path in the selected directory
		fullOutputPath := filepath.Join(m.selectedDir, filename)
		args = append(args, "--output", fullOutputPath)
		
		// Add format-specific arguments
		switch m.exportFormat {
		case "csv":
			args = append(args, "--csv-only")
		case "both":
			args = append(args, "--csv")
		}
		
		// Add debug if enabled
		if m.debugMode {
			args = append(args, "--debug")
		}

		// Execute Python script
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("python", args...)
		} else {
			cmd = exec.Command("python3", args...)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			errorMsg := fmt.Sprintf("❌ Python execution failed: %v\n\n", err)
			if strings.Contains(err.Error(), "executable file not found") {
				errorMsg += "🐍 Python is not installed or not in PATH\n\n"
				errorMsg += "Solutions:\n"
				errorMsg += "1. Install Python from https://python.org\n"
				errorMsg += "2. Make sure 'Add Python to PATH' is checked during installation\n"
				errorMsg += "3. Restart this application after installing Python\n\n"
			}
			errorMsg += "📋 Output:\n" + string(output)
			
			return processCompleteMsg{
				error: errorMsg,
			}
		}

		return processCompleteMsg{
			result: fmt.Sprintf("🎉 Files successfully processed!\n\n📊 Results:\n%s", string(output)),
		}
	}
}

func findPythonScript() string {
	// Get the directory where the executable is located
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	execDir := filepath.Dir(execPath)

	// Look for the Python script in several locations
	searchPaths := []string{
		filepath.Join(execDir, "file_metadata_extractor.py"),
		"file_metadata_extractor.py",
		filepath.Join("..", "file_metadata_extractor.py"),
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}