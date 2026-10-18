package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ineedspace/pkg/cleaner"
	"ineedspace/pkg/model"
	"ineedspace/pkg/scanner"
)

type state int

const (
	stateScanning state = iota
	stateBrowsing
	stateConfirm
	stateCleaning
	stateDone
)

type scanDoneMsg struct {
	components []model.Component
}

type cleanProgressMsg struct {
	component model.Component
	index     int
	total     int
}

type cleanDoneMsg struct {
	results []cleaner.CleanResult
}

// Model represents the Bubble Tea TUI state.
type Model struct {
	state          state
	spinner        spinner.Model
	progress       progress.Model
	components     []model.Component
	cursor         int
	selectedCount  int
	selectedBytes  int64
	totalBytes     int64
	dryRun         bool
	cleaningIndex  int
	cleaningTotal  int
	cleaningItem   string
	cleanResults   []cleaner.CleanResult
	filterCategory string
	categories     []string
	categoryIdx    int
	width          int
	height         int
	scanStatus     string
}

// InitialModel initializes the Bubble Tea model.
func InitialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(secondaryColor)

	p := progress.New(progress.WithDefaultGradient())
	p.Width = 40

	return Model{
		state:      stateScanning,
		spinner:    s,
		progress:   p,
		scanStatus: "Scanning AI models, containers, dev tools, and caches...",
		dryRun:     false,
		categories: []string{
			"All",
			string(model.CategoryAI),
			string(model.CategoryContainers),
			string(model.CategoryDevTools),
			string(model.CategorySDKs),
			string(model.CategoryPackageMgr),
			string(model.CategorySystem),
			string(model.CategoryDownloads),
		},
		categoryIdx: 0,
	}
}

// Init triggers initial scanning.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startScanCmd(),
	)
}

func (m Model) startScanCmd() tea.Cmd {
	return func() tea.Msg {
		components := scanner.ScanAll(nil)
		return scanDoneMsg{components: components}
	}
}

func (m Model) startCleanCmd() tea.Cmd {
	return func() tea.Msg {
		var toClean []model.Component
		for _, c := range m.components {
			if c.Selected {
				toClean = append(toClean, c)
			}
		}

		results := cleaner.CleanBatch(toClean, m.dryRun, nil)
		return cleanDoneMsg{results: results}
	}
}

// VisibleComponents returns components matching the active category filter.
func (m Model) VisibleComponents() []int {
	var indices []int
	activeCat := m.categories[m.categoryIdx]
	for i, c := range m.components {
		if activeCat == "All" || string(c.Category) == activeCat {
			indices = append(indices, i)
		}
	}
	return indices
}

// Update handles incoming messages and user input.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 10
		if m.progress.Width > 50 {
			m.progress.Width = 50
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case scanDoneMsg:
		m.components = msg.components
		m.state = stateBrowsing
		m.totalBytes = 0
		for _, c := range m.components {
			m.totalBytes += c.SizeBytes
		}
		// Select safe items by default to save user time
		for i := range m.components {
			if m.components[i].Risk == model.RiskSafe {
				m.components[i].Selected = true
			}
		}
		m.recalculateSelection()
		return m, nil

	case cleanDoneMsg:
		m.cleanResults = msg.results
		m.state = stateDone
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateBrowsing:
			return m.handleBrowsingKeys(msg)
		case stateConfirm:
			return m.handleConfirmKeys(msg)
		case stateDone:
			if msg.String() == "q" || msg.String() == "enter" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m Model) handleBrowsingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visible := m.VisibleComponents()
	visLen := len(visible)

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if visLen > 0 {
			m.cursor = (m.cursor - 1 + visLen) % visLen
		}

	case "down", "j":
		if visLen > 0 {
			m.cursor = (m.cursor + 1) % visLen
		}

	case " ":
		if visLen > 0 && m.cursor < visLen {
			idx := visible[m.cursor]
			m.components[idx].Selected = !m.components[idx].Selected
			m.recalculateSelection()
		}

	case "a": // Toggle All
		allSelected := true
		for _, idx := range visible {
			if !m.components[idx].Selected {
				allSelected = false
				break
			}
		}
		for _, idx := range visible {
			m.components[idx].Selected = !allSelected
		}
		m.recalculateSelection()

	case "s": // Select only Safe items
		for _, idx := range visible {
			m.components[idx].Selected = (m.components[idx].Risk == model.RiskSafe)
		}
		m.recalculateSelection()

	case "tab": // Next category
		m.categoryIdx = (m.categoryIdx + 1) % len(m.categories)
		m.cursor = 0

	case "shift+tab": // Prev category
		m.categoryIdx = (m.categoryIdx - 1 + len(m.categories)) % len(m.categories)
		m.cursor = 0

	case "d": // Toggle dry run
		m.dryRun = !m.dryRun

	case "enter":
		if m.selectedCount > 0 {
			m.state = stateConfirm
		}
	}

	return m, nil
}

func (m Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.state = stateCleaning
		return m, tea.Batch(m.spinner.Tick, m.startCleanCmd())
	case "d":
		m.dryRun = !m.dryRun
	case "n", "esc", "q":
		m.state = stateBrowsing
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) recalculateSelection() {
	m.selectedCount = 0
	m.selectedBytes = 0
	for _, c := range m.components {
		if c.Selected {
			m.selectedCount++
			m.selectedBytes += c.SizeBytes
		}
	}
}

// View renders the TUI interface.
func (m Model) View() string {
	switch m.state {
	case stateScanning:
		return m.viewScanning()
	case stateBrowsing:
		return m.viewBrowsing()
	case stateConfirm:
		return m.viewConfirm()
	case stateCleaning:
		return m.viewCleaning()
	case stateDone:
		return m.viewDone()
	default:
		return "Unknown state"
	}
}

func (m Model) viewScanning() string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("  INEEDSPACE — DISK CLEANER FOR DEVS & ADMINS  "))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  %s %s\n\n", m.spinner.View(), m.scanStatus))
	b.WriteString(helpStyle.Render("  Scanning ~/Library/Caches, Podman, Ollama, dev tools, and ~/Downloads..."))
	b.WriteString("\n")
	return appStyle.Render(b.String())
}

func (m Model) viewBrowsing() string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("  INEEDSPACE  "))
	b.WriteString("  ")
	b.WriteString(subtitleStyle.Render("Disk Cleaner for Devs & Sysadmins"))
	b.WriteString("  |  Total Found: ")
	b.WriteString(sizeStyle.Render(model.FormatBytes(m.totalBytes)))
	b.WriteString("\n\n")

	// Category Tabs
	var tabs []string
	for i, cat := range m.categories {
		if i == m.categoryIdx {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(primaryColor).
				Bold(true).
				Padding(0, 1).
				Render(cat))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(mutedColor).
				Padding(0, 1).
				Render(cat))
		}
	}
	b.WriteString(strings.Join(tabs, " ") + "\n\n")

	// Component List
	visible := m.VisibleComponents()
	if len(visible) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("  No items found in this category.\n\n"))
	} else {
		maxVisible := 12
		start := 0
		if m.cursor >= maxVisible {
			start = m.cursor - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(visible) {
			end = len(visible)
		}

		for i := start; i < end; i++ {
			cIdx := visible[i]
			c := m.components[cIdx]

			cursor := "  "
			isCursor := (i == m.cursor)
			if isCursor {
				cursor = cursorStyle.Render("▶ ")
			}

			checkbox := checkboxUncheckedStyle.Render("[ ]")
			if c.Selected {
				checkbox = checkboxCheckedStyle.Render("[✓]")
			}

			// Risk Badge
			var riskBadge string
			switch c.Risk {
			case model.RiskSafe:
				riskBadge = riskSafeStyle.Render("SAFE")
			case model.RiskNotice:
				riskBadge = riskNoticeStyle.Render("NOTICE")
			case model.RiskCaution:
				riskBadge = riskCautionStyle.Render("CAUTION")
			}

			nameStr := model.TruncateString(c.Name, 36)
			sizeStr := sizeStyle.Render(fmt.Sprintf("%9s", model.FormatBytes(c.SizeBytes)))

			line := fmt.Sprintf("%s%s %-36s %s %s", cursor, checkbox, nameStr, sizeStr, riskBadge)
			if isCursor {
				b.WriteString(itemSelectedStyle.Render(line) + "\n")
			} else {
				b.WriteString(itemNormalStyle.Render(line) + "\n")
			}
		}

		if len(visible) > maxVisible {
			b.WriteString(helpStyle.Render(fmt.Sprintf("  ... showing %d of %d items (use ↑/↓ to scroll)\n", maxVisible, len(visible))))
		}
	}

	// Details box for currently focused item
	if len(visible) > 0 && m.cursor < len(visible) {
		current := m.components[visible[m.cursor]]
		var detail strings.Builder
		detail.WriteString(fmt.Sprintf("  Path:   %s\n", lipgloss.NewStyle().Foreground(secondaryColor).Render(model.TruncateString(current.Path, 65))))
		detail.WriteString(fmt.Sprintf("  Action: %s\n", current.CleanAction))
		detail.WriteString(fmt.Sprintf("  Impact: %s", lipgloss.NewStyle().Foreground(mutedColor).Render(current.Impact)))
		b.WriteString(detailBoxStyle.Render(detail.String()) + "\n")
	}

	// Selection & Reclaim Summary Bar
	dryRunTag := ""
	if m.dryRun {
		dryRunTag = lipgloss.NewStyle().Foreground(warningColor).Bold(true).Render(" [DRY-RUN ON]")
	}
	summaryLine := fmt.Sprintf("  Selected: %d components | Reclaimable Space: %s%s",
		m.selectedCount,
		summaryBadgeStyle.Render(model.FormatBytes(m.selectedBytes)),
		dryRunTag,
	)
	b.WriteString("\n" + summaryLine + "\n")

	// Help Bar
	b.WriteString(helpStyle.Render(
		"  [Space] Toggle  •  [a] Toggle All  •  [s] Select Safe  •  [Tab] Category  •  [d] Dry Run  •  [Enter] Clean  •  [q] Quit\n",
	))

	return appStyle.Render(b.String())
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(accentColor).
		Padding(0, 1).
		Render("  CONFIRM CLEANUP  "))
	b.WriteString("\n\n")

	modeText := lipgloss.NewStyle().Foreground(dangerColor).Bold(true).Render("ACTUAL DELETION")
	if m.dryRun {
		modeText = lipgloss.NewStyle().Foreground(warningColor).Bold(true).Render("DRY RUN (simulation only)")
	}

	b.WriteString(fmt.Sprintf("  Mode: %s\n", modeText))
	b.WriteString(fmt.Sprintf("  You have selected %s components.\n", lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", m.selectedCount))))
	b.WriteString(fmt.Sprintf("  Total space to be reclaimed: %s\n\n", summaryBadgeStyle.Render(model.FormatBytes(m.selectedBytes))))

	b.WriteString("  Components to be cleaned:\n")
	count := 0
	for _, c := range m.components {
		if c.Selected {
			count++
			if count <= 8 {
				b.WriteString(fmt.Sprintf("   • %-35s (%s) -> %s\n",
					c.Name,
					model.FormatBytes(c.SizeBytes),
					lipgloss.NewStyle().Foreground(mutedColor).Render(c.CleanAction)))
			}
		}
	}
	if count > 8 {
		b.WriteString(fmt.Sprintf("   ... and %d more items\n", count-8))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("  Are you sure you want to proceed?\n\n"))
	b.WriteString(fmt.Sprintf("  %s Confirm & Clean    %s Toggle Dry-Run    %s Cancel / Back\n",
		keyStyle.Render("[y/Enter]"),
		keyStyle.Render("[d]"),
		keyStyle.Render("[n/Esc]"),
	))

	return appStyle.Render(b.String())
}

func (m Model) viewCleaning() string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("  CLEANING IN PROGRESS  "))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  %s Cleaning selected components...\n\n", m.spinner.View()))
	b.WriteString(helpStyle.Render("  Executing safe removal commands and purging caches... Please wait."))
	b.WriteString("\n")
	return appStyle.Render(b.String())
}

func (m Model) viewDone() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(successColor).
		Padding(0, 1).
		Render("  CLEANUP COMPLETE  "))
	b.WriteString("\n\n")

	var totalFreed int64
	var errCount int

	for _, r := range m.cleanResults {
		if r.Error != nil {
			errCount++
			b.WriteString(fmt.Sprintf("  ❌ %s: %v\n", r.Component.Name, r.Error))
		} else {
			totalFreed += r.Freed
			b.WriteString(fmt.Sprintf("  ✓ Cleaned %s (%s)\n",
				r.Component.Name,
				sizeStyle.Render(model.FormatBytes(r.Freed))))
		}
	}

	b.WriteString("\n")
	dryRunNote := ""
	if m.dryRun {
		dryRunNote = " (Simulated — Dry Run)"
	}
	b.WriteString(fmt.Sprintf("  Total Space Reclaimed: %s%s\n",
		summaryBadgeStyle.Render(model.FormatBytes(totalFreed)),
		dryRunNote,
	))

	if errCount > 0 {
		b.WriteString(fmt.Sprintf("  Issues encountered on %d components.\n", errCount))
	}

	b.WriteString("\n" + helpStyle.Render("  Press [q] or [Enter] to exit.\n"))
	return appStyle.Render(b.String())
}
