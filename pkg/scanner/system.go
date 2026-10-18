package scanner

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"ineedspace/pkg/model"
)

// ScanSystem scans for system caches, package manager caches, and large app caches.
func ScanSystem() []model.Component {
	var components []model.Component

	// 1. Homebrew Download Cache
	brewCache := ExpandPath("~/Library/Caches/Homebrew")
	if DirExists(brewCache) {
		sz, count, _ := DirSize(brewCache)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "sys-brew-cache",
				Name:        "Homebrew Download Cache",
				Category:    model.CategoryPackageMgr,
				Description: "Cached formula bottles and git clones in ~/Library/Caches/Homebrew",
				Path:        brewCache,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskSafe,
				Impact:      "Cached bottle archives will be deleted. Will re-download on next 'brew install'.",
				CleanAction: "brew cleanup -s --prune=all / delete cache",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					if CommandExists("brew") {
						_, _ = RunCmdWithTimeout(20*time.Second, "brew", "cleanup", "-s", "--prune=all")
					}
					return sz, os.RemoveAll(brewCache)
				},
			})
		}
	}

	// 2. Xcode DerivedData
	derivedData := ExpandPath("~/Library/Developer/Xcode/DerivedData")
	if DirExists(derivedData) {
		sz, count, _ := DirSize(derivedData)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "sys-xcode-deriveddata",
				Name:        "Xcode DerivedData",
				Category:    model.CategoryDevTools,
				Description: "Xcode intermediate build products, indexes and logs",
				Path:        derivedData,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskSafe,
				Impact:      "Deletes intermediate build artifacts. Xcode will re-index on next project open.",
				CleanAction: "Delete ~/Library/Developer/Xcode/DerivedData",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(derivedData)
				},
			})
		}
	}

	// 3. Xcode iOS Device Support
	deviceSupport := ExpandPath("~/Library/Developer/Xcode/iOS DeviceSupport")
	if DirExists(deviceSupport) {
		sz, count, _ := DirSize(deviceSupport)
		if sz > 100*1024*1024 {
			components = append(components, model.Component{
				ID:          "sys-xcode-devicesupport",
				Name:        "Xcode iOS Device Support Symbols",
				Category:    model.CategoryDevTools,
				Description: "Debug symbols for previously connected physical iOS devices",
				Path:        deviceSupport,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskNotice,
				Impact:      "Debug symbols for old devices will be deleted. Re-created automatically when debugging.",
				CleanAction: "Delete ~/Library/Developer/Xcode/iOS DeviceSupport",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(deviceSupport)
				},
			})
		}
	}

	// 4. Xcode CoreSimulator Caches & Unavailable Simulators
	simCaches := ExpandPath("~/Library/Developer/CoreSimulator/Caches")
	var simTotal int64
	var simCount int
	if DirExists(simCaches) {
		s, c, _ := DirSize(simCaches)
		simTotal += s
		simCount += c
	}
	if simTotal > 10*1024*1024 {
		components = append(components, model.Component{
			ID:          "sys-xcode-coresimulator",
			Name:        "Xcode Simulator Caches & Unavailable Devices",
			Category:    model.CategoryDevTools,
			Description: "CoreSimulator caches in ~/Library/Developer/CoreSimulator/Caches and defunct simulator runtimes",
			Path:        simCaches,
			SizeBytes:   simTotal,
			ItemCount:   simCount,
			Risk:        model.RiskSafe,
			Impact:      "Deletes simulator caches and removes unusable old simulator instances. Current simulators are preserved.",
			CleanAction: "xcrun simctl delete unavailable / delete simulator caches",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return simTotal, nil
				}
				if CommandExists("xcrun") {
					_, _ = RunCmdWithTimeout(20*time.Second, "xcrun", "simctl", "delete", "unavailable")
				}
				_ = os.RemoveAll(simCaches)
				return simTotal, nil
			},
		})
	}

	// 5. Trash
	trashPath := ExpandPath("~/.Trash")
	if DirExists(trashPath) {
		sz, count := scanTrashPath(trashPath)
		if sz > 0 {
			components = append(components, model.Component{
				ID:          "sys-trash",
				Name:        "macOS Trash Bin",
				Category:    model.CategorySystem,
				Description: "Files currently in your Trash (~/.Trash)",
				Path:        trashPath,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskCaution,
				Impact:      "Permanently empties your Trash bin.",
				CleanAction: "Empty Trash",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return cleanTrashPath(trashPath, sz)
				},
			})
		}
	}

	// 6. Heavy App Caches in ~/Library/Caches (> 300MB)
	userCaches := ExpandPath("~/Library/Caches")
	if DirExists(userCaches) {
		entries, err := os.ReadDir(userCaches)
		if err == nil {
			// Skip caches handled in dedicated components
			skipCaches := map[string]bool{
				"Homebrew":          true,
				"go-build":          true,
				"pip":               true,
				"pipx":              true,
				"pypoetry":          true,
				"pipenv":            true,
				"Yarn":              true,
				"CocoaPods":         true,
				"org.swift.swiftpm": true,
			}
			for _, entry := range entries {
				if entry.IsDir() {
					name := entry.Name()
					if skipCaches[name] {
						continue
					}
					target := filepath.Join(userCaches, name)
					sz, count, _ := DirSize(target)
					if sz > 300*1024*1024 { // Flag any individual app cache larger than 300MB
						components = append(components, model.Component{
							ID:          "cache-" + name,
							Name:        fmt.Sprintf("App Cache: %s", name),
							Category:    model.CategorySystem,
							Description: fmt.Sprintf("Cached files for %s in ~/Library/Caches", name),
							Path:        target,
							SizeBytes:   sz,
							ItemCount:   count,
							Risk:        model.RiskSafe,
							Impact:      fmt.Sprintf("Temporary cache for %s will be wiped. App will re-cache as needed.", name),
							CleanAction: fmt.Sprintf("Delete ~/Library/Caches/%s", name),
							Cleaner: func(dryRun bool) (int64, error) {
								if dryRun {
									return sz, nil
								}
								return sz, os.RemoveAll(target)
							},
						})
					}
				}
			}
		}
	}

	return components
}

// scanTrashPath calculates the size and count of items in Trash.
// It uses direct filesystem inspection when available, falling back to macOS Finder AppleScript
// when TCC permissions block listing the ~/.Trash directory directly.
func scanTrashPath(trashPath string) (int64, int) {
	// 1. Direct inspection (works in test environments or when Full Disk Access is granted)
	entries, err := os.ReadDir(trashPath)
	if err == nil {
		var sz int64
		var count int
		for _, e := range entries {
			s, c, _ := DirSize(filepath.Join(trashPath, e.Name()))
			sz += s
			count += c
		}
		return sz, count
	}

	// 2. On macOS when direct listing is blocked by TCC permissions, query Finder via AppleScript
	if runtime.GOOS == "darwin" && CommandExists("osascript") {
		script := `tell application "Finder"
			set names to name of items of trash
			set AppleScript's text item delimiters to linefeed
			return names as text
		end tell`
		out, err := RunCmdWithTimeout(5*time.Second, "osascript", "-e", script)
		if err == nil {
			trimmed := strings.TrimSpace(out)
			if trimmed == "" {
				return 0, 0
			}
			var sz int64
			var count int
			lines := strings.Split(trimmed, "\n")
			for _, name := range lines {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				s, c, _ := DirSize(filepath.Join(trashPath, name))
				sz += s
				count += c
			}
			// If DirSize on individual items returned 0 (e.g. external volume items or permission),
			// fall back to asking Finder for item sizes
			if sz == 0 {
				sizeScript := `tell application "Finder" to get size of items of trash`
				sizeOut, sErr := RunCmdWithTimeout(5*time.Second, "osascript", "-e", sizeScript)
				if sErr == nil {
					parts := strings.Split(sizeOut, ",")
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if val, convErr := strconv.ParseInt(p, 10, 64); convErr == nil {
							sz += val
						}
					}
				}
			}
			if count == 0 && len(lines) > 0 {
				count = len(lines)
			}
			return sz, count
		}
	}

	return 0, 0
}

// cleanTrashPath empties the Trash.
// For the real macOS user trash, it uses Finder's native empty command.
// For test environments or other platforms, it directly deletes entries inside trashPath.
func cleanTrashPath(trashPath string, sz int64) (int64, error) {
	isRealUserTrash := false
	if u, err := user.Current(); err == nil {
		isRealUserTrash = (trashPath == filepath.Join(u.HomeDir, ".Trash"))
	}

	// On macOS for the real user Trash, use Finder AppleScript
	if isRealUserTrash && runtime.GOOS == "darwin" && CommandExists("osascript") {
		_, cmdErr := RunCmdWithTimeout(30*time.Second, "osascript", "-e", `tell application "Finder" to empty trash`)
		if cmdErr == nil {
			return sz, nil
		}
	}

	// Direct removal (for test environments, Full Disk Access, or non-macOS)
	entries, err := os.ReadDir(trashPath)
	if err != nil {
		// If direct ReadDir fails and we can run osascript
		if runtime.GOOS == "darwin" && CommandExists("osascript") {
			_, cmdErr := RunCmdWithTimeout(30*time.Second, "osascript", "-e", `tell application "Finder" to empty trash`)
			if cmdErr != nil {
				return 0, fmt.Errorf("failed to empty trash: %w", cmdErr)
			}
			return sz, nil
		}
		return 0, err
	}

	var firstErr error
	for _, e := range entries {
		if rmErr := os.RemoveAll(filepath.Join(trashPath, e.Name())); rmErr != nil && firstErr == nil {
			firstErr = rmErr
		}
	}

	// Fallback to Finder if direct removal had errors on macOS
	if firstErr != nil && runtime.GOOS == "darwin" && CommandExists("osascript") {
		_, cmdErr := RunCmdWithTimeout(30*time.Second, "osascript", "-e", `tell application "Finder" to empty trash`)
		if cmdErr == nil {
			return sz, nil
		}
	}

	return sz, firstErr
}

