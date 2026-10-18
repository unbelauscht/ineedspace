package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Theme colors
	primaryColor   = lipgloss.Color("#7D56F4") // Purple
	secondaryColor = lipgloss.Color("#00D7D7") // Cyan
	accentColor    = lipgloss.Color("#FF5F87") // Coral/Pink
	successColor   = lipgloss.Color("#5AF78E") // Neon Green
	warningColor   = lipgloss.Color("#F3F99D") // Yellow
	dangerColor    = lipgloss.Color("#FF5C57") // Red
	mutedColor     = lipgloss.Color("#6272A4") // Muted blue/gray
	bgDarkColor    = lipgloss.Color("#1E1E2E")
	borderGray     = lipgloss.Color("#44475A")

	// Base Styles
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1).
			MarginBottom(1)

	// List Item Styles
	itemNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	itemSelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#383A59")).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	checkboxCheckedStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	checkboxUncheckedStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	sizeStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	categoryBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#44475A")).
				Padding(0, 1).
				MarginRight(1)

	// Risk Badges
	riskSafeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(successColor).
			Bold(true).
			Padding(0, 1)

	riskNoticeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(warningColor).
			Bold(true).
			Padding(0, 1)

	riskCautionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(dangerColor).
				Bold(true).
				Padding(0, 1)

	// Footer / Info Box
	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderGray).
			Padding(0, 1).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	keyStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	summaryBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(accentColor).
				Bold(true).
				Padding(0, 1)
)
