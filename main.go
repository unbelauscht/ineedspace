package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ineedspace/pkg/cleaner"
	"ineedspace/pkg/model"
	"ineedspace/pkg/scanner"
	"ineedspace/pkg/tui"
)

var (
	scanOnlyFlag  bool
	cleanSafeFlag bool
	dryRunFlag    bool
	categoryFlag  string
	helpFlag      bool
)

func init() {
	flag.BoolVar(&scanOnlyFlag, "scan", false, "Scan and print disk breakdown without launching interactive TUI")
	flag.BoolVar(&cleanSafeFlag, "clean-safe", false, "Clean all safe caches immediately without interactive TUI")
	flag.BoolVar(&dryRunFlag, "dry-run", false, "Simulate cleaning without deleting anything")
	flag.StringVar(&categoryFlag, "category", "", "Filter scan by category (ai, containers, dev, sdks, system, downloads)")
	flag.BoolVar(&helpFlag, "help", false, "Show help and command usage")
}

func main() {
	flag.Parse()

	// Handle subcommands like "ineedspace scan"
	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "scan":
			scanOnlyFlag = true
		case "help":
			printHelp()
			return
		}
	}

	if helpFlag {
		printHelp()
		return
	}

	// 1. Non-interactive Clean Safe mode
	if cleanSafeFlag {
		runNonInteractiveCleanSafe(dryRunFlag)
		return
	}

	// 2. Scan-only mode (reports summary to stdout)
	if scanOnlyFlag {
		runScanOnly(categoryFlag)
		return
	}

	// 3. Interactive TUI (default)
	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting ineedspace: %v\n", err)
		os.Exit(1)
	}
}

func runScanOnly(categoryFilter string) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("Scanning disk components..."))
	components := scanner.ScanAll(nil)

	if len(components) == 0 {
		fmt.Println("No large cleanable components found.")
		return
	}

	var totalBytes int64
	var safeBytes int64

	fmt.Println()
	fmt.Printf("%-38s %-22s %-10s %-8s\n", "COMPONENT", "CATEGORY", "SIZE", "RISK")
	fmt.Println(strings.Repeat("─", 82))

	for _, c := range components {
		if categoryFilter != "" && !strings.Contains(strings.ToLower(string(c.Category)), strings.ToLower(categoryFilter)) {
			continue
		}

		totalBytes += c.SizeBytes
		if c.Risk == model.RiskSafe {
			safeBytes += c.SizeBytes
		}

		riskLabel := c.Risk.String()
		switch c.Risk {
		case model.RiskSafe:
			riskLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#5AF78E")).Render(riskLabel)
		case model.RiskNotice:
			riskLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#F3F99D")).Render(riskLabel)
		case model.RiskCaution:
			riskLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5C57")).Render(riskLabel)
		}

		fmt.Printf("%-38s %-22s %-10s %-8s\n",
			model.TruncateString(c.Name, 36),
			model.TruncateString(string(c.Category), 20),
			model.FormatBytes(c.SizeBytes),
			riskLabel,
		)
	}

	fmt.Println(strings.Repeat("─", 82))
	fmt.Printf("Total Reclaimable: %s  (Safe to wipe immediately: %s)\n\n",
		lipgloss.NewStyle().Bold(true).Render(model.FormatBytes(totalBytes)),
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5AF78E")).Render(model.FormatBytes(safeBytes)),
	)
}

func runNonInteractiveCleanSafe(dryRun bool) {
	modeStr := "CLEANING"
	if dryRun {
		modeStr = "SIMULATING (DRY RUN)"
	}
	fmt.Printf("[%s] Scanning for safe caches...\n", modeStr)
	components := scanner.ScanAll(nil)

	var toClean []model.Component
	for _, c := range components {
		if c.Risk == model.RiskSafe {
			toClean = append(toClean, c)
		}
	}

	if len(toClean) == 0 {
		fmt.Println("No safe caches found to clean.")
		return
	}

	results := cleaner.CleanBatch(toClean, dryRun, func(c model.Component, idx, total int) {
		fmt.Printf("[%d/%d] Cleaning %s (%s)...\n", idx, total, c.Name, model.FormatBytes(c.SizeBytes))
	})

	var totalFreed int64
	for _, r := range results {
		if r.Error != nil {
			fmt.Printf("  ❌ %s: %v\n", r.Component.Name, r.Error)
		} else {
			totalFreed += r.Freed
		}
	}

	fmt.Printf("\nDone! Total Space Reclaimed: %s\n", model.FormatBytes(totalFreed))
}

func printHelp() {
	help := `
ineedspace — Interactive CLI Disk Cleaner for Developers & System Administrators

USAGE:
  ineedspace [flags]
  ineedspace scan

FLAGS:
  --scan             Scan and print disk breakdown table to stdout without TUI
  --clean-safe       Automatically clean all safe caches without prompting
  --dry-run          Simulate cleanup actions without deleting any files
  --category <name>  Filter scan by category (ai, containers, dev, sdks, system, downloads)
  --help             Show this help screen

KEYBINDINGS IN TUI:
  [↑/k] / [↓/j]      Move up / down the list
  [Space]            Select / deselect highlighted component
  [a]                Toggle select all components in current category
  [s]                Select only SAFE items (rebuildable caches)
  [Tab] / [Shift+Tab]Cycle through categories
  [d]                Toggle Dry-Run mode on / off
  [Enter]            Proceed to confirmation and clean selected
  [q] / [Ctrl+C]     Quit

EXAMPLES:
  # Launch interactive Terminal UI
  ineedspace

  # Quick terminal scan
  ineedspace scan

  # Purge all safe developer & build caches directly
  ineedspace --clean-safe
`
	fmt.Println(help)
}
