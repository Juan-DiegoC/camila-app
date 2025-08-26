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
		return "Directory"
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
	directoryList.Title = "Select Directory to Index"
	directoryList.SetShowStatusBar(false)
	directoryList.SetShowHelp(false)

	// Initialize output input
	ti := textinput.New()
	ti.Placeholder = "Enter filename (e.g. my_index.xlsx)"
	ti.Focus()
	ti.Width = 50

	return model{
		state:         selectingDirectory,
		directoryList: directoryList,
		currentPath:   startDir,
		outputInput:   ti,
		exportFormat:  "both",
		debugMode:     false,
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

	// Sort: directories first, then files
	var dirs, files []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	// Add directories
	for _, entry := range dirs {
		items = append(items, directoryItem{
			name:  entry.Name(),
			path:  filepath.Join(dirPath, entry.Name()),
			isDir: true,
		})
	}

	// Add files (limited to first 20 for performance)
	for i, entry := range files {
		if i >= 20 {
			break
		}
		items = append(items, directoryItem{
			name:  entry.Name(),
			path:  filepath.Join(dirPath, entry.Name()),
			isDir: false,
		})
	}

	return items
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
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
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
						// Update directory list
						items := getDirectoryItems(m.currentPath)
						m.directoryList.SetItems(items)
						return m, nil
					}
				}
			case " ", "space":
				// Select current directory
				m.selectedDir = m.currentPath
				m.state = selectingOutput
				// Pre-fill output with directory name
				dirName := filepath.Base(m.selectedDir)
				m.outputInput.SetValue(fmt.Sprintf("%s_index.xlsx", dirName))
				return m, nil
			case "tab":
				// Quick navigation to common directories
				commonDirs := []string{
					getDownloadsDirectory(),
					filepath.Join(getDownloadsDirectory(), ".."), // Documents usually
					os.Getenv("USERPROFILE"), // Windows home
					os.Getenv("HOME"),        // Unix home
				}
				
				// Find next valid directory
				for _, dir := range commonDirs {
					if dir != "" && dir != m.currentPath {
						if info, err := os.Stat(dir); err == nil && info.IsDir() {
							m.currentPath = dir
							items := getDirectoryItems(m.currentPath)
							m.directoryList.SetItems(items)
							return m, nil
						}
					}
				}
			}

		case selectingOutput:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.state = selectingDirectory
				return m, nil
			case "enter":
				if m.outputInput.Value() != "" {
					m.outputPath = m.outputInput.Value()
					m.state = configuring
					return m, nil
				}
			case "tab":
				// Auto-complete file extension
				current := m.outputInput.Value()
				if !strings.Contains(current, ".") {
					if m.exportFormat == "csv" {
						m.outputInput.SetValue(current + ".csv")
					} else {
						m.outputInput.SetValue(current + ".xlsx")
					}
				} else if strings.HasSuffix(current, ".csv") {
					m.outputInput.SetValue(strings.TrimSuffix(current, ".csv") + ".xlsx")
				} else if strings.HasSuffix(current, ".xlsx") {
					m.outputInput.SetValue(strings.TrimSuffix(current, ".xlsx") + ".csv")
				}
				return m, nil
			}

		case configuring:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.state = selectingOutput
				return m, nil
			case "1":
				m.exportFormat = "excel"
			case "2":
				m.exportFormat = "csv"
			case "3":
				m.exportFormat = "both"
			case "d":
				m.debugMode = !m.debugMode
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
		
		content.WriteString(activeBoxStyle.Render(
			"Step 1: Navigate and Select Directory\n\n" +
				m.directoryList.View() + "\n\n" +
				"📍 Navigation:\n" +
				"  ↑↓ Browse files/folders    📁 Enter = Go into folder\n" +
				"  Space = Select this directory  Tab = Jump to common folders"))

	case selectingOutput:
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("Selected: %s", m.selectedDir)))
		content.WriteString("\n\n")
		content.WriteString(activeBoxStyle.Render(
			"Step 2: Output File Name\n\n" +
				m.outputInput.View() + "\n\n" +
				"💡 Tips:\n" +
				"  • Tab toggles between .xlsx ↔ .csv extension\n" +
				"  • Enter to continue, Esc to go back"))

	case configuring:
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("📁 Directory: %s", m.selectedDir)))
		content.WriteString("\n")
		content.WriteString(boxStyle.Render(
			fmt.Sprintf("📄 Output: %s", m.outputPath)))
		content.WriteString("\n\n")

		configBox := "Step 3: Export Configuration\n\n"
		configBox += "📊 Export Format:\n"
		configBox += fmt.Sprintf("  1) Excel only (.xlsx)     %s\n", checkmark(m.exportFormat == "excel"))
		configBox += fmt.Sprintf("  2) CSV only (.csv)        %s\n", checkmark(m.exportFormat == "csv"))
		configBox += fmt.Sprintf("  3) Both Excel + CSV       %s\n", checkmark(m.exportFormat == "both"))
		configBox += "\n"
		configBox += fmt.Sprintf("🔧 d) Debug mode          %s\n", checkmark(m.debugMode))
		configBox += "\n💫 Press 1-3 to select format, 'd' for debug\n"
		configBox += "⏩ Enter to start processing, Esc to go back"

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
		
		// Add output argument
		args = append(args, "--output", m.outputPath)
		
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