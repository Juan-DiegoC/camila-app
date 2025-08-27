package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	// Nordic-style colors
	nordicAccent = "#88C0D0" // Nordic blue accent
	nordicGreen = "#A3BE8C"  // Nordic green
	nordicRed = "#BF616A"    // Nordic red
	nordicPurple = "#B48EAD" // Nordic purple

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(nordicGreen)).
			MarginLeft(2).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginLeft(2)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(nordicRed)).
			MarginLeft(2)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(nordicGreen)).
			MarginLeft(2)

	pathStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(nordicPurple)).
			MarginLeft(2)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(nordicPurple)).
			Padding(1).
			MarginLeft(2).
			MarginRight(2)

	activeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(nordicAccent)).
			Padding(1).
			MarginLeft(2).
			MarginRight(2)
)

// App states
type state int

const (
	selectingDirectory state = iota
	selectingOutput
	selectingLitigant
	configuring
	processing
	finished
)

// Directory item for the list
type directoryItem struct {
	name string
	path string
	isDir bool
	modTime time.Time
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
	litigantInput   textinput.Model
	selectedDir     string
	outputPath      string
	litigantName    string
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
	isSpanish       bool // Language toggle with Ctrl+E
}

// Language strings
type langStrings struct {
	appTitle                string
	currentPath             string
	step1Title              string
	step2Title              string
	step3Title              string
	step4Title              string
	navigation              string
	browseDirectories       string
	selectFolder            string
	navigateInto            string
	goUp                    string
	selectCurrent           string
	leftRight               string
	filterDirectories       string
	advancedMode            string
	processing              string
	processingDetails       string
	finished                string
	processingFailed        string
	processingSuccess       string
	willSaveAs              string
	fileExtension           string
	fileSavedIn             string
	continueBack            string
	advancedOptions         string
	readyToProcess          string
	changeDirectory         string
	changeFilename          string
	enterToStart            string
	backToGoBack            string
	filesProcessed          string
	savedTo                 string
	readyToUse              string
	tryAgain                string
	processAnother          string
	quitAnytime             string
	filterPrompt            string
	filterControls          string
	indexCreated            string
	goToParent              string
	filterMode              string
	applyFilter             string
	cancelFilter            string
	autoComplete            string
	changeLanguage          string
	reset                   string
	// Step 3 specific strings
	reviewSettings          string
	directory               string
	output                  string
	format                  string
	debug                   string
	advancedModeTitle       string
	toggleFormat            string
	toggleDebug             string
	exitAdvancedMode        string
	toQuit                  string
}

func getStrings(isSpanish bool) langStrings {
	if isSpanish {
		return langStrings{
			appTitle:          "📁 Indexador de Metadatos de Archivos",
			currentPath:       "📍 Actual",
			step1Title:        "Paso 1: Navegar y Seleccionar Directorio (Solo Directorios)",
			step2Title:        "Paso 2: Nombre del Archivo de Salida",
			step3Title:        "Paso 3: Nombre del Litigante",
			step4Title:        "Paso 4: Confirmar Configuración",
			navigation:        "📍 Navegación:",
			browseDirectories: "↑↓ Explorar directorios",
			selectFolder:      "✅ Enter = SELECCIONAR carpeta resaltada",
			navigateInto:      "→ = Entrar a carpeta",
			goUp:              "← = Subir",
			selectCurrent:     "Espacio = Seleccionar actual",
			leftRight:         "←→ = Navegar directorios",
			filterDirectories: "I = Filtrar directorios",
			advancedMode:      "Ctrl+D = Modo avanzado",
			processing:        "⏳ Procesando Archivos...",
			processingDetails: "🔄 Escaneando directorio y extrayendo metadatos\n📊 Esto puede tomar tiempo para directorios grandes\n\nPresiona Ctrl+C para cancelar",
			finished:          "Finalizado",
			processingFailed:  "❌ Procesamiento Fallido",
			processingSuccess: "✅ ¡Procesamiento Completado Exitosamente!",
			willSaveAs:        "💾 Se guardará como:",
			fileExtension:     "• Formato Excel (.xlsx)",
			fileSavedIn:       "• Archivo guardado en directorio seleccionado",
			continueBack:      "• Enter para continuar, B/Esc para regresar",
			advancedOptions:   "• Ctrl+D = Opciones avanzadas (CSV, debug)",
			readyToProcess:    "✨ Listo para procesar:",
			changeDirectory:   "❶ = Cambiar directorio",
			changeFilename:    "❷ = Cambiar nombre",
			enterToStart:      "⏩ Enter para empezar procesamiento",
			backToGoBack:      "B/Esc para regresar",
			filesProcessed:    "📋 Archivos procesados:",
			savedTo:           "💾 Guardado en:",
			readyToUse:        "✅ ¡Tu índice de archivos está listo para usar!",
			tryAgain:          "🔄 Presiona 'r' para intentar de nuevo",
			processAnother:    "🔄 Presiona 'r' para procesar otro directorio",
			quitAnytime:       "Presiona 'q' o Ctrl+C para salir",
			filterPrompt:      "🔍 Filtrar Directorios:",
			filterControls:    "⌨️ Controles:\n  Escribe para filtrar  Tab = Autocompletar  Enter = Aplicar\n  Esc = Cancelar filtro",
			indexCreated:      "🎉 ¡Índice creado exitosamente!",
			goToParent:        "Ir al directorio padre",
			changeLanguage:    "Ctrl+E = Cambiar idioma",
			reset:             "R = Reiniciar",
			// Step 3 specific strings
			reviewSettings:    "Revisar configuración:",
			directory:         "Directorio:",
			output:            "Salida:",
			format:            "Formato:",
			debug:             "Debug: Habilitado",
			advancedModeTitle: "MODO AVANZADO - Más Opciones Disponibles:",
			toggleFormat:      "❸ = Alternar formato",
			toggleDebug:       "🔧 d = Alternar debug",
			exitAdvancedMode:  "🔄 Ctrl+D = Salir modo avanzado",
			toQuit:            "para salir",
		}
	}
	
	// English strings
	return langStrings{
		appTitle:          "📁 File Metadata Indexer",
		currentPath:       "📍 Current",
		step1Title:        "Step 1: Navigate and Select Directory (Directories Only)",
		step2Title:        "Step 2: Output File Name",
		step3Title:        "Step 3: Litigant Name",
		step4Title:        "Step 4: Confirm Settings",
		navigation:        "📍 Navigation:",
		browseDirectories: "↑↓ Browse directories",
		selectFolder:      "✅ Enter = SELECT highlighted folder",
		navigateInto:      "→ = Enter folder",
		goUp:              "← = Go up",
		selectCurrent:     "Space = Select current",
		leftRight:         "←→ = Navigate directories",
		filterDirectories: "I = Filter directories",
		advancedMode:      "Ctrl+D = Advanced mode",
		processing:        "⏳ Processing Files...",
		processingDetails: "🔄 Scanning directory and extracting metadata\n📊 This may take a while for large directories\n\nPress Ctrl+C to cancel",
		finished:          "Finished",
		processingFailed:  "❌ Processing Failed",
		processingSuccess: "✅ Processing Completed Successfully!",
		willSaveAs:        "💾 Will save as:",
		fileExtension:     "• Excel (.xlsx) format",
		fileSavedIn:       "• File saved in selected directory",
		continueBack:      "• Enter to continue, B/Esc to go back",
		advancedOptions:   "• Ctrl+D = Advanced options (CSV, debug)",
		readyToProcess:    "✨ Ready to process:",
		changeDirectory:   "❶ = Change directory",
		changeFilename:    "❷ = Change filename",
		enterToStart:      "⏩ Enter to start processing",
		backToGoBack:      "B/Esc to go back",
		filesProcessed:    "📋 Files processed:",
		savedTo:           "💾 Saved to:",
		readyToUse:        "✅ Your file index is ready to use!",
		tryAgain:          "🔄 Press 'r' to try again",
		processAnother:    "🔄 Press 'r' to process another directory",
		quitAnytime:       "Press 'q' or Ctrl+C to quit anytime",
		filterPrompt:      "🔍 Filter Directories:",
		filterControls:    "⌨️ Controls:\n  Type to filter  Tab = Auto-complete  Enter = Apply\n  Esc = Cancel filter",
		indexCreated:      "🎉 Index created successfully!",
		goToParent:        "Go up to parent directory",
		changeLanguage:    "Ctrl+E = Change language",
		reset:             "R = Reset",
		// Step 3 specific strings
		reviewSettings:    "Review your settings:",
		directory:         "Directory:",
		output:            "Output:",
		format:            "Format:",
		debug:             "Debug: Enabled",
		advancedModeTitle: "ADVANCED MODE - More Options Available:",
		toggleFormat:      "❸ = Toggle format",
		toggleDebug:       "🔧 d = Toggle debug",
		exitAdvancedMode:  "🔄 Ctrl+D = Exit advanced mode",
		toQuit:            "to quit",
	}
}

func initialModel() model {
	// Get Downloads directory
	startDir := getDownloadsDirectory()
	
	// Initialize directory list
	items := getDirectoryItems(startDir)
	
	// Create list with nice styling and double width
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#626262"))
		
	directoryList := list.New(items, delegate, 120, 15) // Double width
	directoryList.Title = fmt.Sprintf("Navigate: %s", startDir)
	directoryList.SetShowStatusBar(false)
	directoryList.SetShowHelp(false)

	// Initialize output input with double width
	ti := textinput.New()
	ti.Placeholder = "Enter filename (e.g. my_index.xlsx)"
	ti.Focus()
	ti.Width = 100 // Double width

	// Initialize filter input with double width
	filterInput := textinput.New()
	filterInput.Placeholder = "Type to filter directories..."
	filterInput.Width = 100 // Double width

	// Initialize litigant input
	litigantInput := textinput.New()
	litigantInput.Placeholder = "Enter litigant name (e.g. Juan Pérez)"
	litigantInput.Width = 100

	return model{
		state:         selectingDirectory,
		directoryList: directoryList,
		currentPath:   startDir,
		outputInput:   ti,
		litigantInput: litigantInput,
		exportFormat:  "excel", // Default to Excel only
		debugMode:     false,
		filtering:     false,
		filterInput:   filterInput,
		allDirectories: convertToDirectoryItems(items),
		advancedMode:  false,
		isSpanish:     true, // Default to Spanish for Spanish users
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
	var dirItems []directoryItem

	// Add parent directory option if not at root
	if parent := filepath.Dir(dirPath); parent != dirPath {
		items = append(items, directoryItem{
			name:   "..",
			path:   parent,
			isDir:  true,
			modTime: time.Time{}, // Parent gets zero time to always appear first
		})
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return items
	}

	// Collect directories with modification times
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
			
			// Get modification time
			fullPath := filepath.Join(dirPath, name)
			info, err := entry.Info()
			var modTime time.Time
			if err == nil {
				modTime = info.ModTime()
			}
			
			// Check if directory is empty (has files)
			if !isDirEmpty(fullPath) {
				dirItems = append(dirItems, directoryItem{
					name:    name,
					path:    fullPath,
					isDir:   true,
					modTime: modTime,
				})
			}
		}
	}

	// Sort by modification time (most recent first)
	sort.Slice(dirItems, func(i, j int) bool {
		return dirItems[i].modTime.After(dirItems[j].modTime)
	})

	// Convert to list items
	for _, dirItem := range dirItems {
		items = append(items, dirItem)
	}

	return items
}

func isDirEmpty(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return true // Consider it empty if we can't read it
	}
	
	// Check if directory has any files (not just subdirectories)
	for _, entry := range entries {
		if !entry.IsDir() {
			return false // Found a file
		}
	}
	
	// If only directories, check if any subdirectory has files
	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(dirPath, entry.Name())
			if !isDirEmpty(subPath) {
				return false // Found files in subdirectory
			}
		}
	}
	
	return true // No files found anywhere
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
		// Global shortcuts available in all states
		switch msg.String() {
		case "ctrl+e":
			// Toggle language
			m.isSpanish = !m.isSpanish
			return m, nil
		case "r":
			// Reset to downloads page (except when processing)
			if m.state != processing {
				return initialModel(), nil
			}
		}
		
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
					m.directoryList.Select(0) // Reset cursor to first item
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
					m.directoryList.Select(0) // Reset cursor to first item when filtering
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
					if selected.name == ".." {
						// Go to parent directory
						m.currentPath = selected.path
						// Update directory list and path display
						items := getDirectoryItems(m.currentPath)
						m.directoryList.SetItems(items)
						m.directoryList.Select(0) // Reset cursor to first item
						m.directoryList.Title = fmt.Sprintf("Navigate: %s", m.currentPath)
						m.allDirectories = convertToDirectoryItems(items)
						return m, nil
					} else {
						// Select the highlighted directory
						m.selectedDir = selected.path
						m.state = selectingOutput
						// Use simple "index" filename (xlsx will be auto-appended)
						m.outputInput.SetValue("index")
						return m, nil
					}
				}
			case "n":
				// Navigate into directory (old enter behavior)
				if selected, ok := m.directoryList.SelectedItem().(directoryItem); ok {
					if selected.isDir && selected.name != ".." {
						m.currentPath = selected.path
						// Update directory list and path display
						items := getDirectoryItems(m.currentPath)
						m.directoryList.SetItems(items)
						m.directoryList.Select(0) // Reset cursor to first item
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
			case "left":
				// Go up one directory level
				parentDir := filepath.Dir(m.currentPath)
				if parentDir != m.currentPath { // Not at root
					m.currentPath = parentDir
					items := getDirectoryItems(m.currentPath)
					m.directoryList.SetItems(items)
					m.directoryList.Select(0) // Reset cursor to first item
					m.directoryList.Title = fmt.Sprintf("Navigate: %s", m.currentPath)
					m.allDirectories = convertToDirectoryItems(items)
					return m, nil
				}
			case "right":
				// Navigate into highlighted directory
				if selected, ok := m.directoryList.SelectedItem().(directoryItem); ok {
					if selected.isDir && selected.name != ".." {
						m.currentPath = selected.path
						// Update directory list and path display
						items := getDirectoryItems(m.currentPath)
						m.directoryList.SetItems(items)
						m.directoryList.Select(0) // Reset cursor to first item
						m.directoryList.Title = fmt.Sprintf("Navigate: %s", m.currentPath)
						m.allDirectories = convertToDirectoryItems(items)
						return m, nil
					}
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
				// If we have a valid filename, go to litigant step to show the change
				if strings.TrimSpace(m.outputInput.Value()) != "" {
					m.outputPath = m.outputInput.Value()
					m.state = selectingLitigant
					m.litigantInput.Focus()
				}
				return m, nil
			case "b", "backspace", "esc":
				m.state = selectingDirectory
				return m, nil
			case "enter":
				if m.outputInput.Value() != "" {
					m.outputPath = m.outputInput.Value()
					m.state = selectingLitigant
					m.litigantInput.Focus()
					return m, nil
				}
			}

		case selectingLitigant:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "b", "backspace", "esc":
				m.state = selectingOutput
				m.outputInput.Focus()
				return m, nil
			case "enter":
				m.litigantName = m.litigantInput.Value()
				m.state = configuring
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
				// Edit litigant name
				m.state = selectingLitigant
				m.litigantInput.Focus()
				return m, nil
			case "4":
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
		// Only pass certain keys to the directoryList, not navigation keys
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "up", "down", "j", "k":
				// Pass these keys for list navigation (up/down selection)
				m.directoryList, cmd = m.directoryList.Update(msg)
			case "left", "right":
				// Don't pass left/right to list - we handle them ourselves
				// No action needed, already handled above
			default:
				// Pass other keys normally
				if !m.filtering { // Don't update list when filtering
					m.directoryList, cmd = m.directoryList.Update(msg)
				}
			}
		}
		// Don't update filter input here - it's handled in the key cases above
	case selectingOutput:
		m.outputInput, cmd = m.outputInput.Update(msg)
	case selectingLitigant:
		m.litigantInput, cmd = m.litigantInput.Update(msg)
	}

	return m, cmd
}

func getBoxStyle(backgroundOn bool) lipgloss.Style {
	return boxStyle // No background changes
}

func getActiveBoxStyle(backgroundOn bool) lipgloss.Style {
	return activeBoxStyle // No background changes
}

func (m model) View() string {
	var content strings.Builder
	str := getStrings(m.isSpanish)
	
	// Get dynamic styles (no background changes needed)
	dynamicBoxStyle := getBoxStyle(false)
	dynamicActiveBoxStyle := getActiveBoxStyle(false)

	// Title
	content.WriteString(titleStyle.Render(str.appTitle))
	content.WriteString("\n\n")

	switch m.state {
	case selectingDirectory:
		
		// Show filter input if filtering
		if m.filtering {
			content.WriteString(dynamicActiveBoxStyle.Render(
				str.filterPrompt + "\n\n" +
				m.filterInput.View() + "\n\n" +
				m.directoryList.View() + "\n\n" +
				str.filterControls))
		} else {
			// Don't override the title here - it's already set in navigation logic
			content.WriteString(dynamicActiveBoxStyle.Render(
				str.step1Title + "\n\n" +
				m.directoryList.View() + "\n\n" +
				str.navigation + "\n" +
				"  " + str.browseDirectories + "    " + str.selectFolder + "\n" +
				"  " + str.leftRight + "  " + str.selectCurrent + "\n" +
				"  " + str.filterDirectories + "  " + str.advancedMode))
		}

	case selectingOutput:
		content.WriteString(dynamicBoxStyle.Render(
			fmt.Sprintf("Selected: %s", m.selectedDir)))
		content.WriteString("\n\n")
		// Show where file will be saved
		previewName := m.outputInput.Value()
		if !strings.Contains(previewName, ".") {
			previewName += ".xlsx" // Default extension
		}
		fullPath := filepath.Join(m.selectedDir, previewName)
		
		content.WriteString(dynamicActiveBoxStyle.Render(
			"Step 2: Output File Name\n\n" +
			m.outputInput.View() + "\n\n" +
			fmt.Sprintf("💾 Will save as: %s\n\n", fullPath) +
			"💡 Simple Mode:\n" +
			"  • Excel (.xlsx) format\n" +
			"  • File saved in selected directory\n" +
			"  • Enter to continue, B/Esc to go back\n" +
			"  • Ctrl+D = Advanced options (CSV, debug)"))

	case selectingLitigant:
		content.WriteString(dynamicBoxStyle.Render(
			fmt.Sprintf("Selected: %s", m.selectedDir)))
		content.WriteString("\n\n")
		content.WriteString(dynamicBoxStyle.Render(
			fmt.Sprintf("Output: %s", m.outputPath)))
		content.WriteString("\n\n")
		
		content.WriteString(dynamicActiveBoxStyle.Render(
			"Step 3: Litigant Name\n\n" +
			m.litigantInput.View() + "\n\n" +
			"💼 Enter the name of the person litigating\n" +
			"   (e.g., Juan Pérez, María González)\n\n" +
			"📝 This will be included in the document header\n\n" +
			"Enter to continue, B/Esc to go back"))

	case configuring:
		content.WriteString(dynamicBoxStyle.Render(
			fmt.Sprintf("📁 Directory: %s", m.selectedDir)))
		content.WriteString("\n")
		content.WriteString(dynamicBoxStyle.Render(
			fmt.Sprintf("📄 Output: %s", m.outputPath)))
		content.WriteString("\n\n")

		configBox := str.step4Title + "\n\n"
		configBox += "✨ " + str.reviewSettings + "\n\n"
		configBox += fmt.Sprintf("📁 " + str.directory + " %s\n", filepath.Base(m.selectedDir))
		configBox += fmt.Sprintf("📄 " + str.output + " %s\n", m.outputPath)
	configBox += fmt.Sprintf("🏛️ Litigante: %s\n", m.litigantName)
		
		if m.advancedMode {
			formatName := "Excel (.xlsx)"
			if m.exportFormat == "csv" {
				formatName = "CSV (.csv)"
			} else if m.exportFormat == "both" {
				formatName = "Excel + CSV"
			}
			configBox += fmt.Sprintf("📊 " + str.format + " %s\n", formatName)
			if m.debugMode {
				configBox += "🔧 " + str.debug + "\n"
			}
			configBox += "\n🅰️ " + str.advancedModeTitle + "\n"
			configBox += "  " + str.changeDirectory + "     " + str.changeFilename + "\n"
			configBox += "  3 = Editar litigante\n"
			configBox += "  " + str.toggleFormat + "        " + str.toggleDebug + "\n"
			configBox += "  " + str.exitAdvancedMode + "\n"
		} else {
			configBox += "\n" + str.readyToProcess + "\n"
			configBox += "  " + str.changeDirectory + "     " + str.changeFilename + "\n"
			configBox += "  3 = Editar litigante\n"
			configBox += "  🔥 " + str.advancedOptions + "\n"
		}
		configBox += "\n⏩ " + str.enterToStart + "  •  " + str.backToGoBack

		content.WriteString(dynamicActiveBoxStyle.Render(configBox))

	case processing:
		content.WriteString(dynamicBoxStyle.Render(
			fmt.Sprintf("📁 Directory: %s", m.selectedDir)))
		content.WriteString("\n")
		content.WriteString(dynamicBoxStyle.Render(
			fmt.Sprintf("📄 Output: %s", m.outputPath)))
		content.WriteString("\n\n")
		content.WriteString(dynamicActiveBoxStyle.Render(
			str.processing + "\n\n" +
			str.processingDetails))

	case finished:
		content.WriteString(dynamicBoxStyle.Render(
			fmt.Sprintf("📁 Directory: %s", m.selectedDir)))
		content.WriteString("\n")
		content.WriteString(dynamicBoxStyle.Render(
			fmt.Sprintf("📄 Output: %s", m.outputPath)))
		content.WriteString("\n\n")

		if m.error != "" {
			content.WriteString(errorStyle.Render(str.processingFailed))
			content.WriteString("\n")
			content.WriteString(dynamicBoxStyle.Render(m.error))
			content.WriteString("\n")
			content.WriteString(helpStyle.Render(str.tryAgain + "  •  Enter/Esc " + str.toQuit))
		} else {
			content.WriteString(successStyle.Render(str.processingSuccess))
			content.WriteString("\n")
			content.WriteString(dynamicBoxStyle.Render(m.result))
			content.WriteString("\n")
			content.WriteString(helpStyle.Render(str.processAnother + "  •  Enter/Esc " + str.toQuit)) 
		}
	}

	content.WriteString("\n\n")
	// Bottom status bar
	statusBar := str.quitAnytime
	if !m.advancedMode {
		statusBar += "  •  " + str.advancedMode + "  •  " + str.changeLanguage + "  •  " + str.reset
	} else {
		statusBar += "  •  " + str.changeLanguage + "  •  " + str.reset
	}
	content.WriteString(helpStyle.Render(statusBar))

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
		
		// Add litigant name if provided
		if m.litigantName != "" {
			args = append(args, "--litigant", m.litigantName)
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

		// Parse the output to extract key information
		outputStr := string(output)
		
		// Extract the saved file path from output
		savedPath := fullOutputPath
		if strings.Contains(outputStr, "saved to:") {
			// Try to extract actual path if mentioned in output
			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				if strings.Contains(line, "saved to:") {
					parts := strings.Split(line, "saved to:")
					if len(parts) > 1 {
						savedPath = strings.TrimSpace(parts[1])
					}
					break
				}
			}
		}
		
		// Count files processed (look for common indicators)
		fileCount := "multiple"
		if strings.Contains(outputStr, "processed") {
			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				if strings.Contains(line, "processed") && strings.Contains(line, "file") {
					// Extract number if present
					words := strings.Fields(line)
					for i, word := range words {
						if strings.Contains(word, "file") && i > 0 {
							fileCount = words[i-1]
							break
						}
					}
					break
				}
			}
		}
		
		return processCompleteMsg{
			result: fmt.Sprintf("🎉 Index created successfully!\n\n📋 Files processed: %s\n💾 Saved to: %s\n\n✅ Your file index is ready to use!", fileCount, savedPath),
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