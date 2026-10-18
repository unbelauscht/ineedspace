package model

import (
	"fmt"
	"strings"
)

// RiskLevel defines how safe an action is.
type RiskLevel int

const (
	RiskSafe RiskLevel = iota // Safe caches that auto-rebuild or re-download (e.g. go-build, uv, npm)
	RiskNotice                 // Downloads or models you can re-fetch, but take time (e.g. Ollama, LM Studio)
	RiskCaution                // Containers, VMs, or user files that require explicit intent (e.g. Downloads zips, Podman VM)
)

func (r RiskLevel) String() string {
	switch r {
	case RiskSafe:
		return "SAFE"
	case RiskNotice:
		return "NOTICE"
	case RiskCaution:
		return "CAUTION"
	default:
		return "UNKNOWN"
	}
}

// Category represents a grouping of cleanable items.
type Category string

const (
	CategoryAI         Category = "AI & LLM Models"
	CategoryContainers Category = "Containers & VMs"
	CategoryDevTools   Category = "Dev & Build Caches"
	CategorySDKs       Category = "Mobile & Embedded SDKs"
	CategoryPackageMgr Category = "Package Managers"
	CategorySystem     Category = "App & System Caches"
	CategoryDownloads  Category = "Downloads & Archives"
)

// CleanFunc is a function that performs the cleanup. Returns bytes freed and any error.
type CleanFunc func(dryRun bool) (int64, error)

// Component represents a detected component on the disk.
type Component struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    Category  `json:"category"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	SizeBytes   int64     `json:"size_bytes"`
	ItemCount   int       `json:"item_count"`
	Risk        RiskLevel `json:"risk"`
	Impact      string    `json:"impact"`
	CleanAction string    `json:"clean_action"`
	Cleaner     CleanFunc `json:"-"`
	Selected    bool      `json:"selected"`
}

// FormatBytes formats bytes into human readable binary units.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), units[exp])
}

// TruncateString truncates a string with ellipsis if it exceeds maxLength.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return strings.TrimSpace(s[:maxLen-3]) + "..."
}
